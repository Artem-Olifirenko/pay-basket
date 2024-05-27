package resolver

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
	"go.citilink.cloud/order/internal/order/payment/resolver/mssql"
	"go.citilink.cloud/store_types"
	"testing"
)

func TestOnlinePaymentCheck_Resolve(t *testing.T) {
	type paymentStatusMap map[order.PaymentId]order.AllowStatus

	tests := []struct {
		name   string
		fields struct {
			resolvedIdsMap     payment.ResolvedPaymentIdMap
			order              func(ctrl *gomock.Controller) order.Order
			onlinePaymentCheck func(ctrl *gomock.Controller) *OnlinePaymentCheck
		}
		want    paymentStatusMap
		wantErr error
	}{
		{
			name: "regions info is not available",
			fields: struct {
				resolvedIdsMap     payment.ResolvedPaymentIdMap
				order              func(ctrl *gomock.Controller) order.Order
				onlinePaymentCheck func(ctrl *gomock.Controller) *OnlinePaymentCheck
			}{
				resolvedIdsMap: func() payment.ResolvedPaymentIdMap {
					return payment.ResolvedPaymentIdMap{
						order.PaymentIdWebmoney:          payment.NewResolvedPaymentId(order.PaymentIdWebmoney),
						order.PaymentIdCardsOnline:       payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
						order.PaymentIdYandex:            payment.NewResolvedPaymentId(order.PaymentIdYandex),
						order.PaymentIdTerminalOrCashbox: payment.NewResolvedPaymentId(order.PaymentIdTerminalOrCashbox),
						order.PaymentIdSbp:               payment.NewResolvedPaymentId(order.PaymentIdSbp),
						order.PaymentIdApplePay:          payment.NewResolvedPaymentId(order.PaymentIdApplePay),
						order.PaymentIdCashless:          payment.NewResolvedPaymentId(order.PaymentIdCashless),
					}
				}(),
				order: func(ctrl *gomock.Controller) order.Order {
					return order.NewMockOrder(ctrl)
				},
				onlinePaymentCheck: func(ctrl *gomock.Controller) *OnlinePaymentCheck {
					p := NewMockOnlinePaymentInfo(ctrl)
					p.EXPECT().GetRegions(gomock.Any()).Return(nil, fmt.Errorf("regions are not available"))
					return NewOnlinePaymentCheck(p)
				},
			},
			want: paymentStatusMap{
				order.PaymentIdWebmoney:          order.AllowStatusDisallowed,
				order.PaymentIdCardsOnline:       order.AllowStatusDisallowed,
				order.PaymentIdYandex:            order.AllowStatusDisallowed,
				order.PaymentIdTerminalOrCashbox: order.AllowStatusDisallowed,
				order.PaymentIdSbp:               order.AllowStatusDisallowed,
				order.PaymentIdApplePay:          order.AllowStatusDisallowed,
				order.PaymentIdCashless:          order.AllowStatusAllow,
			},
			wantErr: errors.New("regions are not available"),
		},
		{
			name: "space of order does not support online payment",
			fields: struct {
				resolvedIdsMap     payment.ResolvedPaymentIdMap
				order              func(ctrl *gomock.Controller) order.Order
				onlinePaymentCheck func(ctrl *gomock.Controller) *OnlinePaymentCheck
			}{
				resolvedIdsMap: func() payment.ResolvedPaymentIdMap {
					return payment.ResolvedPaymentIdMap{
						order.PaymentIdWebmoney:          payment.NewResolvedPaymentId(order.PaymentIdWebmoney),
						order.PaymentIdCardsOnline:       payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
						order.PaymentIdYandex:            payment.NewResolvedPaymentId(order.PaymentIdYandex),
						order.PaymentIdTerminalOrCashbox: payment.NewResolvedPaymentId(order.PaymentIdTerminalOrCashbox),
						order.PaymentIdSbp:               payment.NewResolvedPaymentId(order.PaymentIdSbp),
						order.PaymentIdApplePay:          payment.NewResolvedPaymentId(order.PaymentIdApplePay),
						order.PaymentIdCashless:          payment.NewResolvedPaymentId(order.PaymentIdCashless),
					}
				}(),
				order: func(ctrl *gomock.Controller) order.Order {
					o := order.NewMockOrder(ctrl)
					o.EXPECT().SpaceId().Return(store_types.SpaceId("some_space"))
					return o
				},
				onlinePaymentCheck: func(ctrl *gomock.Controller) *OnlinePaymentCheck {
					p := NewMockOnlinePaymentInfo(ctrl)
					p.EXPECT().GetRegions(gomock.Any()).Return(
						map[store_types.SpaceId]*mssql.YandexPayShopData{
							"some_space": {
								SpaceId:   "some_space",
								IsEnabled: false,
							},
						}, nil)
					return NewOnlinePaymentCheck(p)
				},
			},
			want: paymentStatusMap{
				order.PaymentIdWebmoney:          order.AllowStatusDisallowed,
				order.PaymentIdCardsOnline:       order.AllowStatusDisallowed,
				order.PaymentIdYandex:            order.AllowStatusDisallowed,
				order.PaymentIdTerminalOrCashbox: order.AllowStatusDisallowed,
				order.PaymentIdSbp:               order.AllowStatusDisallowed,
				order.PaymentIdApplePay:          order.AllowStatusDisallowed,
				order.PaymentIdCashless:          order.AllowStatusAllow,
			},
			wantErr: nil,
		},
		{
			name: "space of order is not recognized by system",
			fields: struct {
				resolvedIdsMap     payment.ResolvedPaymentIdMap
				order              func(ctrl *gomock.Controller) order.Order
				onlinePaymentCheck func(ctrl *gomock.Controller) *OnlinePaymentCheck
			}{
				resolvedIdsMap: func() payment.ResolvedPaymentIdMap {
					return payment.ResolvedPaymentIdMap{
						order.PaymentIdWebmoney:          payment.NewResolvedPaymentId(order.PaymentIdWebmoney),
						order.PaymentIdCardsOnline:       payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
						order.PaymentIdYandex:            payment.NewResolvedPaymentId(order.PaymentIdYandex),
						order.PaymentIdTerminalOrCashbox: payment.NewResolvedPaymentId(order.PaymentIdTerminalOrCashbox),
						order.PaymentIdSbp:               payment.NewResolvedPaymentId(order.PaymentIdSbp),
						order.PaymentIdApplePay:          payment.NewResolvedPaymentId(order.PaymentIdApplePay),
						order.PaymentIdCashless:          payment.NewResolvedPaymentId(order.PaymentIdCashless),
					}
				}(),
				order: func(ctrl *gomock.Controller) order.Order {
					o := order.NewMockOrder(ctrl)
					o.EXPECT().SpaceId().Return(store_types.SpaceId("some_another_space"))
					return o
				},
				onlinePaymentCheck: func(ctrl *gomock.Controller) *OnlinePaymentCheck {
					p := NewMockOnlinePaymentInfo(ctrl)
					p.EXPECT().GetRegions(gomock.Any()).Return(
						map[store_types.SpaceId]*mssql.YandexPayShopData{
							"some_space": {
								SpaceId:   "some_space",
								IsEnabled: false,
							},
						}, nil)
					return &OnlinePaymentCheck{
						onlinePaymentInfo: p,
					}
				},
			},
			want: paymentStatusMap{
				order.PaymentIdWebmoney:          order.AllowStatusDisallowed,
				order.PaymentIdCardsOnline:       order.AllowStatusDisallowed,
				order.PaymentIdYandex:            order.AllowStatusDisallowed,
				order.PaymentIdTerminalOrCashbox: order.AllowStatusDisallowed,
				order.PaymentIdSbp:               order.AllowStatusDisallowed,
				order.PaymentIdApplePay:          order.AllowStatusDisallowed,
				order.PaymentIdCashless:          order.AllowStatusAllow,
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gotPaymentMap := make(paymentStatusMap)

			gotErr := test.fields.onlinePaymentCheck(ctrl).
				Resolve(context.Background(), test.fields.resolvedIdsMap, test.fields.order(ctrl))

			for pmntId, pmnt := range test.fields.resolvedIdsMap {
				gotPaymentMap[pmntId] = pmnt.Status()
			}
			assert.Equal(t, gotErr, test.wantErr)
			assert.Equal(t, gotPaymentMap, test.want)
		})
	}
}
