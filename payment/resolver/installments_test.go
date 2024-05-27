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

func TestInstallments_Resolve(t *testing.T) {
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
			name: "all items in basket are available for purchase",
			basketItems: func() basket_item.Items {
				bI := basket_item.Items{basket_item.NewItem(
					"5534",
					basket_item.TypeDigitalService,
					"5534",
					"image",
					0,
					0,
					0,
					store_types.NewSpaceId("msk_cl"),
					catalog_types.PriceColumnRetail,
				), basket_item.NewItem(
					"569",
					basket_item.TypeProduct,
					"569",
					"image",
					0,
					0,
					0,
					store_types.NewSpaceId("msk_cl"),
					catalog_types.PriceColumnRetail,
				)}

				serviceAdditions := &basket_item.Service{}
				bI[0].Additions().SetService(serviceAdditions)
				serviceAdditions.SetIsAvailableForInstallments(true)

				productAdditions := &basket_item.ProductItemAdditions{}
				bI[1].Additions().SetProduct(productAdditions)
				bI[1].Additions().GetProduct().SetCreditPrograms([]catalog_types.CreditProgram{"foo"})
				return bI
			}(),
			args: args{
				ctx: context.Background(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdInstallments:           payment.NewResolvedPaymentId(order.PaymentIdInstallments),
					order.PaymentIdCash:                   payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
				},
				ordr: func(ctrl *gomock.Controller, bskItems basket_item.Items) order.Order {
					o := order.NewMockOrder(ctrl)
					b := order.NewMockBasket(ctrl)
					b.EXPECT().All().Return(bskItems)
					o.EXPECT().Basket(context.Background()).Return(b, nil)

					return o
				},
			},
			wantPaymentStatusMap: func(bskItems basket_item.Items) expectedPropertiesPerPayment {
				return expectedPropertiesPerPayment{
					order.PaymentIdInstallments: struct {
						status  order.AllowStatus
						reasons []*order.DisallowReasonWithInfo
					}{
						status:  order.AllowStatusAllow,
						reasons: nil,
					},
					order.PaymentIdCash: struct {
						status  order.AllowStatus
						reasons []*order.DisallowReasonWithInfo
					}{
						status:  order.AllowStatusAllow,
						reasons: nil,
					},
					order.PaymentIdSberbankBusinessOnline: struct {
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
			name: "some items in basket are blocked by resolver",
			basketItems: func() basket_item.Items {
				bI := basket_item.Items{basket_item.NewItem(
					"5534",
					basket_item.TypeDigitalService,
					"5534",
					"image",
					0,
					0,
					0,
					store_types.NewSpaceId("msk_cl"),
					catalog_types.PriceColumnRetail,
				), basket_item.NewItem(
					"569",
					basket_item.TypeProduct,
					"569",
					"image",
					0,
					0,
					0,
					store_types.NewSpaceId("msk_cl"),
					catalog_types.PriceColumnRetail,
				), basket_item.NewItem(
					"56119",
					basket_item.TypeProduct,
					"56119",
					"image",
					0,
					0,
					0,
					store_types.NewSpaceId("msk_cl"),
					catalog_types.PriceColumnRetail,
				), basket_item.NewItem(
					"1551",
					basket_item.TypeConfigurationAssemblyService,
					"1551",
					"image",
					0,
					0,
					0,
					store_types.NewSpaceId("msk_cl"),
					catalog_types.PriceColumnRetail,
				)}

				serviceAdditions := &basket_item.Service{}
				bI[0].Additions().SetService(serviceAdditions)
				bI[0].Additions().GetService().SetIsAvailableForInstallments(false)

				productAdditions := &basket_item.ProductItemAdditions{}
				bI[1].Additions().SetProduct(productAdditions)
				bI[1].Additions().GetProduct().SetCreditPrograms([]catalog_types.CreditProgram{})

				productAdditions2 := &basket_item.ProductItemAdditions{}
				bI[2].Additions().SetProduct(productAdditions2)
				bI[2].Additions().GetProduct().SetCreditPrograms([]catalog_types.CreditProgram{"foo"})
				return bI
			}(),
			args: args{
				ctx: context.Background(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdInstallments:           payment.NewResolvedPaymentId(order.PaymentIdInstallments),
					order.PaymentIdCash:                   payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
				},
				ordr: func(ctrl *gomock.Controller, bskItems basket_item.Items) order.Order {
					o := order.NewMockOrder(ctrl)
					b := order.NewMockBasket(ctrl)
					b.EXPECT().All().Return(bskItems)
					o.EXPECT().Basket(context.Background()).Return(b, nil)

					return o
				},
			},
			wantPaymentStatusMap: func(bskItems basket_item.Items) expectedPropertiesPerPayment {
				return expectedPropertiesPerPayment{
					order.PaymentIdInstallments: struct {
						status  order.AllowStatus
						reasons []*order.DisallowReasonWithInfo
					}{
						status: order.AllowStatusDisallowed,
						reasons: []*order.DisallowReasonWithInfo{
							order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdItemNotAvailableForPaymentMethod, order.SubsystemBasket).
								WithAdditions(order.DisallowReasonWithInfoAdditions{
									Basket: order.NewBasketDisallowAdditions(bskItems[0:2].UniqIds()),
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
					order.PaymentIdSberbankBusinessOnline: struct {
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
			r := NewInstallments()
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
				w := tt.wantPaymentStatusMap(tt.basketItems)
				assert.Equal(t, w, gotPaymentMap)
			}
		})
	}
}
