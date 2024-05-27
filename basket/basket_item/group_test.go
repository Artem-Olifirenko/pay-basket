package basket_item

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestGroupSuite(t *testing.T) {
	suite.Run(t, &GroupSuite{})
}

type GroupSuite struct {
	suite.Suite
}

func (s *GroupSuite) SetupTest() {}

func (s *GroupSuite) SetupSubTest() {}

func (s *GroupSuite) TestGroup_Validate() {
	tests := []struct {
		name string
		g    Group
		err  error
	}{
		{
			name: "invalid group",
			g:    GroupInvalid,
			err:  fmt.Errorf("group is not in valid state invalid"),
		},
		{
			name: "valid group",
			g:    GroupProduct,
			err:  nil,
		},
		{
			name: "valid group",
			g:    GroupService,
			err:  nil,
		},
		{
			name: "valid group",
			g:    GroupConfiguration,
			err:  nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			err := tt.g.Validate()

			if err != nil {
				s.Equal(err.Error(), tt.err.Error())
				return
			}

			s.Nil(err)
		})
	}
}
