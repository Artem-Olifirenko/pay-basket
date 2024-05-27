package basket_item

import (
	"bytes"
	"github.com/stretchr/testify/suite"
	"gopkg.in/vmihailenco/msgpack.v2"
	"testing"
)

func TestInfoMsgpackSuite(t *testing.T) {
	suite.Run(t, &InfoMsgpackSuite{})
}

type InfoMsgpackSuite struct {
	suite.Suite

	info         *Info
	additions    *InfoAdditions
	priceChanged *PriceChangedInfoAddition
	availInfo    *CountMoreThenAvailInfoAdditions
}

func (s *InfoMsgpackSuite) SetupTest() {}

func (s *InfoMsgpackSuite) SetupSubTest() {
	s.info = &Info{}
	s.additions = &InfoAdditions{}
	s.priceChanged = &PriceChangedInfoAddition{}
	s.availInfo = &CountMoreThenAvailInfoAdditions{}
}

func (s *InfoMsgpackSuite) TestBasketItemInformation_DecodeMsgpack() {
	tests := []struct {
		name     string
		obj      interface{}
		err      string
		expected *Info
	}{
		{
			name: "base type substitution negative",
			obj:  []byte{},
			err:  "can't decode msgpack field `Info array len`[0]: msgpack: invalid code c4 decoding array length",
		},
		{
			name: "empty struct negative",
			obj:  []interface{}{},
			err:  "can't decode msgpack field `incorrect Info length`[0]: incorrect Info length: 0",
		},
		{
			name: "id string negative",
			obj:  []interface{}{"id", 0, &InfoAdditions{}},
			err:  "can't decode msgpack field `Info id`[1]: msgpack: invalid code a2 decoding int64",
		},
		{
			name: "message int negative",
			obj:  []interface{}{0, 0, &InfoAdditions{}},
			err:  "can't decode msgpack field `Info message`[2]: msgpack: invalid code 0 decoding bytes length",
		},
		{
			name: "empty",
			obj:  []interface{}{},
			err:  "can't decode msgpack field `incorrect Info length`[0]: incorrect Info length: 0",
		},
		{
			name:     "positive",
			obj:      []interface{}{27, "message", nil},
			expected: &Info{id: 27, message: "message", additions: nil},
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)
			reader := bytes.NewReader(buf)
			decoder := msgpack.NewDecoder(reader)
			err = s.info.DecodeMsgpack(decoder)

			if err != nil {
				s.Equal(err.Error(), tt.err)
				return
			}

			s.Equal(tt.expected, s.info)
		})
	}
}

func (s *InfoMsgpackSuite) TestInfoAdditions_DecodeMsgpack() {
	tests := []struct {
		name     string
		obj      interface{}
		err      string
		expected *InfoAdditions
	}{
		{
			name: "base type substitution negative",
			obj:  []byte{},
			err:  "can't decode msgpack field `InfoAdditions array len`[0]: msgpack: invalid code c4 decoding array length",
		},
		{
			name: "empty struct negative",
			obj:  []interface{}{},
			err:  "can't decode msgpack field `incorrect InfoAdditions length`[0]: incorrect InfoAdditions length: 0",
		},
		{
			name: "PriceChangedInfoAddition.From is string negative",
			obj:  []interface{}{[]interface{}{"From", 2}, []interface{}{3}},
			err:  "can't decode msgpack field `InfoAdditions PriceChanged`[1]: can't decode msgpack field `PriceChangedInfoAddition From`[1]: msgpack: invalid code a4 decoding int64",
		},
		{
			name: "PriceChangedInfoAddition.To missing negative",
			obj:  []interface{}{[]interface{}{2}, []interface{}{3}},
			err:  "can't decode msgpack field `InfoAdditions PriceChanged`[1]: can't decode msgpack field `incorrect PriceChangedInfoAddition length`[0]: incorrect PriceChangedInfoAddition length: 1",
		},
		{
			name: "empty CountMoreThenAvailInfoAdditions negative",
			obj:  []interface{}{[]interface{}{1, 2}, []interface{}{}},
			err:  "can't decode msgpack field `InfoAdditions CountMoreThenAvail`[2]: can't decode msgpack field `incorrect CountMoreThenAvailInfoAdditions length`[0]: incorrect CountMoreThenAvailInfoAdditions length: 0",
		},
		{
			name: "empty",
			obj:  []interface{}{[]interface{}{}, []interface{}{}},
			err:  "can't decode msgpack field `InfoAdditions PriceChanged`[1]: can't decode msgpack field `incorrect PriceChangedInfoAddition length`[0]: incorrect PriceChangedInfoAddition length: 0",
		},
		{
			name:     "positive",
			obj:      []interface{}{[]interface{}{1, 2}, []interface{}{3}},
			expected: &InfoAdditions{PriceChanged: PriceChangedInfoAddition{1, 2}, CountMoreThenAvail: CountMoreThenAvailInfoAdditions{3}},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)
			reader := bytes.NewReader(buf)
			decoder := msgpack.NewDecoder(reader)
			err = s.additions.DecodeMsgpack(decoder)

			if err != nil {
				s.Equal(err.Error(), tt.err)
				return
			}

			s.Equal(tt.expected, s.additions)
		})
	}
}

func (s *InfoMsgpackSuite) TestPriceChangedInfoAddition_DecodeMsgpack() {
	tests := []struct {
		name     string
		obj      interface{}
		err      string
		expected *PriceChangedInfoAddition
	}{
		{
			name: "base type substitution negative",
			obj:  []byte{},
			err:  "can't decode msgpack field `PriceChangedInfoAddition array len`[0]: msgpack: invalid code c4 decoding array length",
		},
		{
			name: "missing To negative",
			obj:  []interface{}{1},
			err:  "can't decode msgpack field `incorrect PriceChangedInfoAddition length`[0]: incorrect PriceChangedInfoAddition length: 1",
		},
		{
			name: "From is string negative",
			obj:  []interface{}{"From", 2},
			err:  "can't decode msgpack field `PriceChangedInfoAddition From`[1]: msgpack: invalid code a4 decoding int64",
		},
		{
			name: "To is string negative",
			obj:  []interface{}{2, "To"},
			err:  "can't decode msgpack field `PriceChangedInfoAddition To`[2]: msgpack: invalid code a2 decoding int64",
		},
		{
			name: "empty",
			obj:  []interface{}{},
			err:  "can't decode msgpack field `incorrect PriceChangedInfoAddition length`[0]: incorrect PriceChangedInfoAddition length: 0",
		},
		{
			name:     "positive",
			obj:      []interface{}{1, 2},
			expected: &PriceChangedInfoAddition{From: 1, To: 2},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)
			reader := bytes.NewReader(buf)
			decoder := msgpack.NewDecoder(reader)
			err = s.priceChanged.DecodeMsgpack(decoder)

			if err != nil {
				s.Equal(err.Error(), tt.err)
				return
			}

			s.Equal(tt.expected, s.priceChanged)
		})
	}
}

func (s *InfoMsgpackSuite) TestCountMoreThenAvailInfoAdditions_DecodeMsgpack() {
	tests := []struct {
		name     string
		obj      interface{}
		err      string
		expected *CountMoreThenAvailInfoAdditions
	}{
		{
			name: "base type substitution negative",
			obj:  []byte{},
			err:  "can't decode msgpack field `CountMoreThenAvailInfoAdditions array len`[0]: msgpack: invalid code c4 decoding array length",
		},
		{
			name: "AvailCount is not int negative",
			obj:  []interface{}{"AvailCount"},
			err:  "can't decode msgpack field `CountMoreThenAvailInfoAdditions AvailCount`[1]: msgpack: invalid code aa decoding int64",
		},
		{
			name: "empty",
			obj:  []interface{}{},
			err:  "can't decode msgpack field `incorrect CountMoreThenAvailInfoAdditions length`[0]: incorrect CountMoreThenAvailInfoAdditions length: 0",
		},
		{
			name:     "positive",
			obj:      []interface{}{1},
			expected: &CountMoreThenAvailInfoAdditions{AvailCount: 1},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)
			reader := bytes.NewReader(buf)
			decoder := msgpack.NewDecoder(reader)
			err = s.availInfo.DecodeMsgpack(decoder)

			if err != nil {
				s.Equal(err.Error(), tt.err)
				return
			}

			s.Equal(tt.expected, s.availInfo)
		})
	}
}
