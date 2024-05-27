package basket_item

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestInfoSuite(t *testing.T) {
	suite.Run(t, &InfoSuite{})
}

type InfoSuite struct {
	suite.Suite

	info *Info
}

func (s *InfoSuite) SetupTest() {}

func (s *InfoSuite) SetupSubTest() {}

func (s *InfoSuite) TestInfo_Additionals() {
	expectedAdditions := InfoAdditions{
		PriceChanged: PriceChangedInfoAddition{
			From: 100,
			To:   200,
		},
		CountMoreThenAvail: CountMoreThenAvailInfoAdditions{
			AvailCount: 100,
		},
	}

	type fields struct {
		id        InfoId
		message   string
		additions *InfoAdditions
	}
	tests := []struct {
		name   string
		fields fields
		want   *InfoAdditions
	}{
		{
			name: "test getting info additions",
			fields: fields{
				additions: &expectedAdditions,
			},
			want: &expectedAdditions,
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			s.info = &Info{
				id:        tt.fields.id,
				message:   tt.fields.message,
				additions: tt.fields.additions,
			}

			s.Equal(tt.want, s.info.Additionals())
		})
	}
}

func (s *InfoSuite) TestInfo_Id() {
	expectedID := InfoIdPriceChanged

	type fields struct {
		id        InfoId
		message   string
		additions *InfoAdditions
	}

	tests := []struct {
		name   string
		fields fields
		want   InfoId
	}{
		{
			name: "test correct id value",
			fields: fields{
				id: expectedID,
			},
			want: expectedID,
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			s.info = &Info{
				id:        tt.fields.id,
				message:   tt.fields.message,
				additions: tt.fields.additions,
			}

			s.Equal(tt.want, s.info.Id())
		})
	}
}

func (s *InfoSuite) TestInfo_Message() {
	type fields struct {
		id        InfoId
		message   string
		additions *InfoAdditions
	}

	expectedMessage := "message"

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test correct message value",
			fields: fields{
				message: expectedMessage,
			},
			want: expectedMessage,
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			s.info = &Info{
				id:        tt.fields.id,
				message:   tt.fields.message,
				additions: tt.fields.additions,
			}

			s.Equal(tt.want, s.info.Message())
		})
	}
}

func (s *InfoSuite) TestInfo_SetAdditions() {
	type fields struct {
		id        InfoId
		message   string
		additions *InfoAdditions
	}

	expectedAdditions := InfoAdditions{
		PriceChanged: PriceChangedInfoAddition{
			From: 100,
			To:   200,
		},
		CountMoreThenAvail: CountMoreThenAvailInfoAdditions{
			AvailCount: 100,
		},
	}

	tests := []struct {
		name   string
		fields fields
		want   *InfoAdditions
	}{
		{
			name: "test set additions",
			want: &expectedAdditions,
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			s.info = &Info{
				id:        tt.fields.id,
				message:   tt.fields.message,
				additions: tt.fields.additions,
			}
			s.info.SetAdditions(tt.want)

			s.Equal(tt.want, s.info.Additionals())
		})
	}
}

func (s *InfoSuite) TestNewInfo() {
	type args struct {
		id      InfoId
		message string
	}
	expectedInfo := &Info{id: 10, message: "message", additions: &InfoAdditions{}}

	tests := []struct {
		name string
		args args
		want *Info
	}{
		{
			name: "test info creation",
			args: args{
				id:      expectedInfo.id,
				message: expectedInfo.message,
			},
			want: expectedInfo,
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			got := NewInfo(expectedInfo.id, expectedInfo.message)
			s.Equal(tt.want, got)
		})
	}
}
