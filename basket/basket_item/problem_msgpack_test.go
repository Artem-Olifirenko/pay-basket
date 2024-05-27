package basket_item

import (
	"bytes"
	"github.com/stretchr/testify/suite"
	"gopkg.in/vmihailenco/msgpack.v2"
	"testing"
)

func TestProblemMsgpackSuite(t *testing.T) {
	suite.Run(t, &ProblemMsgpackSuite{})
}

type ProblemMsgpackSuite struct {
	suite.Suite

	reader  *bytes.Reader
	decoder *msgpack.Decoder
}

func (s *ProblemMsgpackSuite) SetupTest() {}

func (s *ProblemMsgpackSuite) SetupSubTest() {}

func (s *ProblemMsgpackSuite) TestProblem_DecodeMsgpack() {
	tests := []struct {
		name     string
		args     interface{}
		err      string
		expected *Problem
	}{
		{
			name: "base type substitution negative",
			args: []byte{},
			err:  "can't decode msgpack field `Problem array len`[0]: msgpack: invalid code c4 decoding array length",
		},
		{
			name: "empty",
			args: []interface{}{},
			err:  "can't decode msgpack field `(basket_item.Problem) len doesn't match`[0]: (basket_item.Problem) len doesn't match: 0",
		},
		{
			name: "id is not int negative",
			args: []interface{}{"", "message", []interface{}{[]interface{}{[]interface{}{}}}},
			err:  "can't decode msgpack field `Problem id`[1]: msgpack: invalid code a0 decoding int64",
		},
		{
			name: "message is not string negative",
			args: []interface{}{1, 0, []interface{}{[]interface{}{[]interface{}{}}}},
			err:  "can't decode msgpack field `Problem message`[2]: msgpack: invalid code 0 decoding bytes length",
		},
		{
			name: "additions is not struct negative",
			args: []interface{}{1, "message", []interface{}{[]interface{}{[]byte{}}}},
			err:  "can't decode msgpack field `Problem additions`[3]: can't decode msgpack field `ProblemAdditions ConfigurationProblemAdditions`[1]: can't decode msgpack field `ConfigurationProblemAdditions NotAvailableProductItemIds`[1]: msgpack: invalid code c4 decoding array length",
		},
		{
			name: "positive",
			args: []interface{}{1, "message", []interface{}{[]interface{}{[]interface{}{}}}},
			expected: &Problem{
				id:      1,
				message: "message",
				additions: ProblemAdditions{
					ConfigurationProblemAdditions: ConfigurationProblemAdditions{
						NotAvailableProductItemIds: []ItemId{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, _ := msgpack.Marshal(tt.args)
			problem := &Problem{}
			s.reader = bytes.NewReader(buf)
			s.decoder = msgpack.NewDecoder(s.reader)
			err := problem.DecodeMsgpack(s.decoder)

			if err != nil {
				s.Equal(err.Error(), tt.err)
				return
			}
			s.Equal(tt.expected, problem)
		})
	}
}

func (s *ProblemMsgpackSuite) TestProblemAdditions_DecodeMsgpack() {
	tests := []struct {
		name     string
		args     interface{}
		err      string
		expected *ProblemAdditions
	}{
		{
			name: "base type substitution negative",
			args: []byte{},
			err:  "can't decode msgpack field `ProblemAdditions array len`[0]: msgpack: invalid code c4 decoding array length",
		},
		{
			name: "empty",
			args: []interface{}{},
			err:  "can't decode msgpack field `(basket_item.ProblemAdditions) len doesn't match`[0]: (basket_item.ProblemAdditions) len doesn't match: 0",
		},
		{
			name: "ConfigurationProblemAdditions is not struct negative",
			args: []interface{}{[]interface{}{[]byte{}}},
			err:  "can't decode msgpack field `ProblemAdditions ConfigurationProblemAdditions`[1]: can't decode msgpack field `ConfigurationProblemAdditions NotAvailableProductItemIds`[1]: msgpack: invalid code c4 decoding array length",
		},
		{
			name: "positive",
			args: []interface{}{[]interface{}{[]interface{}{}}},
			expected: &ProblemAdditions{
				ConfigurationProblemAdditions: ConfigurationProblemAdditions{
					NotAvailableProductItemIds: []ItemId{},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			problemAdditions := &ProblemAdditions{}
			buf, _ := msgpack.Marshal(tt.args)
			s.reader = bytes.NewReader(buf)
			s.decoder = msgpack.NewDecoder(s.reader)
			err := problemAdditions.DecodeMsgpack(s.decoder)

			if err != nil {
				s.Equal(err.Error(), tt.err)
				return
			}
			s.Equal(tt.expected, problemAdditions)
		})
	}
}

func (s *ProblemMsgpackSuite) TestConfigurationProblemAdditions_DecodeMsgpack() {
	tests := []struct {
		name     string
		args     interface{}
		err      string
		expected *ConfigurationProblemAdditions
	}{
		{
			name: "base type substitution negative",
			args: []byte{},
			err:  "can't decode msgpack field `ConfigurationProblemAdditions array len`[0]: msgpack: invalid code c4 decoding array length",
		},
		{
			name: "empty",
			args: []interface{}{},
			err:  "can't decode msgpack field `(basket_item.ConfigurationProblemAdditions) incorrect len`[0]: (basket_item.ConfigurationProblemAdditions) incorrect len: 0",
		},
		{
			name: "NotAvailableProductItemIds is not slice ItemId type negative",
			args: []interface{}{[]interface{}{0}},
			err:  "can't decode msgpack field `ConfigurationProblemAdditions NotAvailableProductItemIds`[2]: msgpack: invalid code 0 decoding bytes length",
		},
		{
			name:     "positive",
			args:     []interface{}{[]interface{}{}},
			expected: &ConfigurationProblemAdditions{NotAvailableProductItemIds: []ItemId{}},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			config := &ConfigurationProblemAdditions{}
			buf, _ := msgpack.Marshal(tt.args)
			s.reader = bytes.NewReader(buf)
			s.decoder = msgpack.NewDecoder(s.reader)
			err := config.DecodeMsgpack(s.decoder)

			if err != nil {
				s.Equal(err.Error(), tt.err)
				return
			}
			s.Equal(tt.expected, config)
		})
	}
}
