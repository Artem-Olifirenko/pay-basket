package basket

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.citilink.cloud/order/internal"
	"reflect"
	"testing"

	"go.citilink.cloud/citizap/factory"
	database "go.citilink.cloud/libdatabase"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
)

func TestBasketFactorySuite(t *testing.T) {
	suite.Run(t, &BasketFactorySuite{})
}

type BasketFactorySuite struct {
	suite.Suite
	ctrl                *gomock.Controller
	stringContainerMock *internal.MockStringsContainer
}

func (s *BasketFactorySuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())

	s.stringContainerMock = internal.NewMockStringsContainer(s.ctrl)
}

func (s *BasketFactorySuite) TestBasketFactory_CreateAnon() {
	type fields struct {
		itemFactory   basket_item.ItemFactory
		productClient productv1.ProductAPIClient
		itemRefresher itemRefresher
		db            database.DB
	}
	type args struct {
		basket        *BasketData
		loggerFactory factory.Factory
	}
	tests := []struct {
		name    string
		fields  fields
		request args
	}{
		{
			name: "test anonymous factory",
			request: args{
				basket:        &BasketData{},
				loggerFactory: nil,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			f := &BasketFactory{
				itemFactory:                     tt.fields.itemFactory,
				productClient:                   tt.fields.productClient,
				itemRefresher:                   tt.fields.itemRefresher,
				db:                              tt.fields.db,
				markingOptions:                  NewMarkingOptions(false, s.stringContainerMock),
				subcontractServiceChangeOptions: NewSubcontractServiceChangeOptions(false),
			}
			got := f.CreateAnon(tt.request.basket, tt.request.loggerFactory)
			s.Nil(got.user)
		})
	}
}

func (s *BasketFactorySuite) TestBasketFactory_CreateUser() {
	type fields struct {
		itemFactory   basket_item.ItemFactory
		productClient productv1.ProductAPIClient
		itemRefresher itemRefresher
		db            database.DB
	}
	type args struct {
		basket        *BasketData
		user          *userv1.User
		loggerFactory factory.Factory
	}
	tests := []struct {
		name    string
		fields  fields
		request args
	}{
		{
			name: "test anonymous factory",
			request: args{
				basket:        &BasketData{},
				loggerFactory: nil,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			f := &BasketFactory{
				itemFactory:                     tt.fields.itemFactory,
				productClient:                   tt.fields.productClient,
				itemRefresher:                   tt.fields.itemRefresher,
				db:                              tt.fields.db,
				markingOptions:                  NewMarkingOptions(false, s.stringContainerMock),
				subcontractServiceChangeOptions: NewSubcontractServiceChangeOptions(false),
			}
			got := f.CreateUser(tt.request.basket, &userv1.User{}, tt.request.loggerFactory)
			s.NotNil(got.user)
		})
	}
}

func (s *BasketFactorySuite) TestNewBasketFactory() {
	type args struct {
		itemFactory                      basket_item.ItemFactory
		productClient                    productv1.ProductAPIClient
		itemRefresher                    itemRefresher
		db                               database.DB
		markingEnabled                   bool
		markingEnabledInCities           internal.StringsContainer
		subcontractServicesChangeEnabled bool
	}
	tests := []struct {
		name    string
		request args
		want    func() *BasketFactory
	}{
		{
			name: "Test basket factory creation",
			request: args{
				itemFactory:                      nil,
				productClient:                    nil,
				itemRefresher:                    nil,
				db:                               nil,
				markingEnabled:                   false,
				markingEnabledInCities:           s.stringContainerMock,
				subcontractServicesChangeEnabled: false,
			},
			want: func() *BasketFactory {
				return &BasketFactory{
					itemFactory:   nil,
					productClient: nil,
					itemRefresher: nil,
					db:            nil,
					markingOptions: NewMarkingOptions(
						false,
						s.stringContainerMock,
					),
					subcontractServiceChangeOptions: NewSubcontractServiceChangeOptions(
						false,
					),
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(
			tt.name, func() {
				got := NewBasketFactory(
					tt.request.itemFactory,
					tt.request.productClient,
					tt.request.itemRefresher,
					tt.request.db,
					NewMarkingOptions(
						tt.request.markingEnabled,
						tt.request.markingEnabledInCities,
					),
					NewSubcontractServiceChangeOptions(
						tt.request.subcontractServicesChangeEnabled,
					),
					nil,
				)
				want := tt.want()
				if !reflect.DeepEqual(got, want) {
					s.Errorf(errors.New("unexpected basket factory"), "NewBasketFactory() = %v, want %v", got, want)
				}
			},
		)
	}
}
