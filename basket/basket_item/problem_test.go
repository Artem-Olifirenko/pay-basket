package basket_item

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestProblemSuite(t *testing.T) {
	suite.Run(t, &ProblemSuite{})
}

type ProblemSuite struct {
	suite.Suite
}

func (s *ProblemSuite) SetupTest() {}

func (s *ProblemSuite) SetupSubTest() {}

func (s *ProblemSuite) TestNewProblem() {
	type args struct {
		id      ProblemId
		message string
	}
	expectedObj := &Problem{
		id:      ProblemNotAvailable,
		message: "message",
	}

	tests := []struct {
		name string
		args args
		want *Problem
	}{
		{
			name: "create new problem",
			args: args{
				id:      expectedObj.id,
				message: expectedObj.message,
			},
			want: expectedObj,
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			got := NewProblem(tt.args.id, tt.args.message)

			s.Equal(got, tt.want)
		})
	}
}

func (s *ProblemSuite) TestProblem_Additions() {
	type args struct {
		id        ProblemId
		message   string
		additions ProblemAdditions
	}

	expectedObj := args{
		id:      ProblemNotAvailable,
		message: "message",
		additions: ProblemAdditions{
			ConfigurationProblemAdditions{
				NotAvailableProductItemIds: []ItemId{"item-id-1", "item-id-2", "item-id-3"},
			},
		},
	}

	tests := []struct {
		name   string
		fields args
		want   *ProblemAdditions
	}{
		{
			name: "check if problem additions exists",
			fields: args{
				id:        expectedObj.id,
				message:   expectedObj.message,
				additions: expectedObj.additions,
			},
			want: &expectedObj.additions,
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			p := &Problem{
				id:        tt.fields.id,
				message:   tt.fields.message,
				additions: tt.fields.additions,
			}
			got := p.Additions()

			s.Equal(got, tt.want)
		})
	}
}

func (s *ProblemSuite) TestProblem_Id() {
	type args struct {
		id        ProblemId
		message   string
		additions ProblemAdditions
	}

	expectedObj := &Problem{
		id: ProblemNotAvailable,
	}

	tests := []struct {
		name   string
		fields args
		want   ProblemId
	}{
		{
			name: "test correct problemId value",
			fields: args{
				id: expectedObj.id,
			},
			want: expectedObj.id,
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			p := &Problem{
				id:        tt.fields.id,
				message:   tt.fields.message,
				additions: tt.fields.additions,
			}
			got := p.Id()

			s.Equal(got, tt.want)
		})
	}
}

func (s *ProblemSuite) TestProblem_Message() {
	type fields struct {
		id        ProblemId
		message   string
		additions ProblemAdditions
	}

	expectedObj := &Problem{
		message: "message",
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test correct problemMessage value",
			fields: fields{
				message: expectedObj.message,
			},
			want: expectedObj.message,
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			p := &Problem{
				id:        tt.fields.id,
				message:   tt.fields.message,
				additions: tt.fields.additions,
			}
			got := p.Message()

			s.Equal(got, tt.want)
		})
	}
}
