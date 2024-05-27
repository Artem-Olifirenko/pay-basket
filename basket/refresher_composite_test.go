package basket

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.uber.org/zap"
	"testing"
)

type RefresherComposite struct {
	suite.Suite

	ctrl                *gomock.Controller
	ctx                 context.Context
	logger              *zap.Logger
	refresherMock       *Mockrefresher
	refresherBasketMock *MockRefresherBasket
}

func (s *RefresherComposite) SetupTest() {
	s.logger = zap.NewNop()
}

func (s *RefresherComposite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.refresherMock = NewMockrefresher(s.ctrl)
	s.refresherBasketMock = NewMockRefresherBasket(s.ctrl)

}

func (s *RefresherComposite) TestNewItemRefresherComposite() {
	refr1, refr2 := s.refresherMock, s.refresherMock

	refrComposite := NewItemRefresherComposite(refr1, refr2)
	s.Equal(&ItemRefresherComposite{
		refreshers: []refresher{refr1, refr2},
	}, refrComposite)
}

func (s *RefresherComposite) TestItemRefresherComposite_Refresh() {

	tests := []struct {
		name    string
		prepare func(s *RefresherComposite)
		wantErr bool
	}{
		{
			name: "refresh",
			prepare: func(s *RefresherComposite) {

				item := &basket_item.Item{}
				s.refresherBasketMock.EXPECT().All().Return(basket_item.Items{
					item,
				}).Times(1)

				s.refresherMock.EXPECT().Refreshable(&basket_item.Item{}).Return(true).Times(1)
				s.refresherMock.EXPECT().Refresh(
					s.ctx,
					[]*basket_item.Item{item},
					s.refresherBasketMock,
					s.logger,
				).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "not refreshable item",
			prepare: func(s *RefresherComposite) {
				items := basket_item.Items{
					&basket_item.Item{},
				}
				s.refresherBasketMock.EXPECT().All().Return(items).Times(1)

				s.refresherMock.EXPECT().Refreshable(items[0]).Return(false).Times(1)
			},
			wantErr: false,
		},
		{
			name: "refresh error",
			prepare: func(s *RefresherComposite) {
				items := basket_item.Items{
					&basket_item.Item{},
				}
				s.refresherBasketMock.EXPECT().All().Return(items).Times(1)

				s.refresherMock.EXPECT().Refreshable(&basket_item.Item{}).Return(true).Times(1)
				s.refresherMock.EXPECT().Refresh(
					s.ctx,
					items,
					s.refresherBasketMock,
					s.logger,
				).Return(errors.New("test error")).Times(1)
			},

			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		s.Run(tt.name, func() {
			tt.prepare(s)
			r := NewItemRefresherComposite(s.refresherMock)
			err := r.Refresh(s.ctx, s.refresherBasketMock, s.logger)
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func TestRefresherCompositeSuite(t *testing.T) {
	suite.Run(t, &RefresherComposite{})
}
