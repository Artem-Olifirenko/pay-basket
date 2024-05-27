package resolver

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/payment"
	"go.citilink.cloud/store_types"
	"testing"
)

func TestDigitalService_Resolve(t *testing.T) {

	type expectedPropertiesPerPayment map[order.PaymentId]struct {
		status  order.AllowStatus
		reasons []*order.DisallowReasonWithInfo
	}

	type args struct {
		ctx            context.Context
		resolvedIdsMap payment.ResolvedPaymentIdMap
		ordr           func(ctrl *gomock.Controller, bskItems basket_item.Items) order.Order
	}

	tests := []struct {
		name                 string
		basketItems          basket_item.Items
		args                 args
		wantPaymentStatusMap func(bskItems basket_item.Items) expectedPropertiesPerPayment
		wantErr              error
	}{
		{
			name: "digital items in basket",
			basketItems: basket_item.Items{basket_item.NewItem(
				"1234",
				basket_item.TypeDigitalService,
				"1234",
				"image",
				0,
				0,
				0,
				store_types.NewSpaceId("msk_cl"),
				catalog_types.PriceColumnRetail,
			), basket_item.NewItem(
				"6116",
				basket_item.TypeProduct,
				"6116",
				"image",
				0,
				0,
				0,
				store_types.NewSpaceId("msk_cl"),
				catalog_types.PriceColumnRetail,
			)},
			args: args{
				ctx: context.Background(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:                   payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
				},
				ordr: func(ctrl *gomock.Controller, bskItems basket_item.Items) order.Order {
					o := order.NewMockOrder(ctrl)
					b := order.NewMockBasket(ctrl)
					b.EXPECT().Find(gomock.Any()).Return(bskItems)
					o.EXPECT().Basket(context.Background()).Return(b, nil)

					return o
				},
			},
			wantPaymentStatusMap: func(bskItems basket_item.Items) expectedPropertiesPerPayment {
				return expectedPropertiesPerPayment{
					order.PaymentIdSberbankBusinessOnline: struct {
						status  order.AllowStatus
						reasons []*order.DisallowReasonWithInfo
					}{
						status: order.AllowStatusLimited,
						reasons: []*order.DisallowReasonWithInfo{
							order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdItemNotAvailableForPaymentMethod, order.SubsystemBasket).
								WithAdditions(order.DisallowReasonWithInfoAdditions{
									Basket: order.NewBasketDisallowAdditions(bskItems.UniqIds()),
								}),
						},
					},
					order.PaymentIdCash: struct {
						status  order.AllowStatus
						reasons []*order.DisallowReasonWithInfo
					}{
						status:  order.AllowStatusAllow,
						reasons: nil,
					},
				}
			},
		},
		{
			name:        "basket returned error",
			basketItems: basket_item.Items{},
			args: args{
				ctx:            context.Background(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{},
				ordr: func(ctrl *gomock.Controller, bskItems basket_item.Items) order.Order {
					o := order.NewMockOrder(ctrl)
					o.EXPECT().Basket(context.Background()).Times(1).Return(nil, fmt.Errorf("some error"))

					return o
				},
			},
			wantPaymentStatusMap: func(bskItems basket_item.Items) expectedPropertiesPerPayment {
				return expectedPropertiesPerPayment{}
			},
			wantErr: fmt.Errorf("some error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			r := NewDigitalService()
			err := r.Resolve(tt.args.ctx, tt.args.resolvedIdsMap, tt.args.ordr(ctrl, tt.basketItems))
			if err != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				gotPaymentMap := make(expectedPropertiesPerPayment)
				for pmntId, pmnt := range tt.args.resolvedIdsMap {
					gotPaymentMap[pmntId] = struct {
						status  order.AllowStatus
						reasons []*order.DisallowReasonWithInfo
					}{
						status:  pmnt.Status(),
						reasons: pmnt.Reasons(),
					}
				}
				assert.Equal(t, tt.wantPaymentStatusMap(tt.basketItems), gotPaymentMap)
			}
		})
	}
}
