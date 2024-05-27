package resolver

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
	"testing"
)

func TestMinCost_Resolve(t *testing.T) {
	type args struct {
		ctx            context.Context
		resolvedIdsMap payment.ResolvedPaymentIdMap
		orders         func(ctrl *gomock.Controller) map[order.PaymentId]order.Order
	}

	tests := []struct {
		name string
		args *args
		want func() (order.AllowStatus, *order.DisallowReasonWithInfo, error)
	}{
		{
			name: "allowing cases for min payments",
			args: &args{
				ctx: context.Background(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdInvalid:                payment.NewResolvedPaymentId(order.PaymentIdInvalid),
					order.PaymentIdUnknown:                payment.NewResolvedPaymentId(order.PaymentIdUnknown),
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdCash:                   payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCardsOnline:            payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
					order.PaymentIdCredit:                 payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdYandex:                 payment.NewResolvedPaymentId(order.PaymentIdYandex),
					order.PaymentIdTerminalOrCashbox:      payment.NewResolvedPaymentId(order.PaymentIdTerminalOrCashbox),
					order.PaymentIdWebmoney:               payment.NewResolvedPaymentId(order.PaymentIdWebmoney),
					order.PaymentIdTerminalPinpad:         payment.NewResolvedPaymentId(order.PaymentIdTerminalPinpad),
					order.PaymentIdInstallments:           payment.NewResolvedPaymentId(order.PaymentIdInstallments),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
					order.PaymentIdSbp:                    payment.NewResolvedPaymentId(order.PaymentIdSbp),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
					order.PaymentIdApplePay:               payment.NewResolvedPaymentId(order.PaymentIdApplePay),
				},
				orders: func(ctrl *gomock.Controller) map[order.PaymentId]order.Order {
					amounts := map[order.PaymentId]int{
						order.PaymentIdInvalid:                1,
						order.PaymentIdUnknown:                1,
						order.PaymentIdCashless:               0,
						order.PaymentIdCash:                   0,
						order.PaymentIdCardsOnline:            0,
						order.PaymentIdCredit:                 100,
						order.PaymentIdYandex:                 0,
						order.PaymentIdTerminalOrCashbox:      0,
						order.PaymentIdWebmoney:               0,
						order.PaymentIdTerminalPinpad:         0,
						order.PaymentIdInstallments:           100,
						order.PaymentIdSberbankBusinessOnline: 0,
						order.PaymentIdSbp:                    0,
						order.PaymentIdCashWithCard:           0,
						order.PaymentIdApplePay:               0,
					}
					orders := make(map[order.PaymentId]order.Order)
					for paymentId, amount := range amounts {
						orderMock := order.NewMockOrder(ctrl)
						orderMock.EXPECT().Cost(gomock.Any()).Return(&order.OrderCost{
							WithDiscount:    amount,
							WithoutDiscount: amount,
						}, nil).Times(1)
						orders[paymentId] = orderMock
					}
					return orders
				},
			},
			want: func() (order.AllowStatus, *order.DisallowReasonWithInfo, error) {
				return order.AllowStatusAllow, nil, nil
			},
		},
		{
			name: "disallowing cases for min payments",
			args: &args{
				ctx: context.Background(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdInvalid:      payment.NewResolvedPaymentId(order.PaymentIdInvalid),
					order.PaymentIdUnknown:      payment.NewResolvedPaymentId(order.PaymentIdUnknown),
					order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
				},
				orders: func(ctrl *gomock.Controller) map[order.PaymentId]order.Order {
					amounts := map[order.PaymentId]int{
						order.PaymentIdInvalid:      0,
						order.PaymentIdUnknown:      0,
						order.PaymentIdCredit:       99,
						order.PaymentIdInstallments: 99,
					}
					orders := make(map[order.PaymentId]order.Order)
					for paymentId, amount := range amounts {
						orderMock := order.NewMockOrder(ctrl)
						orderMock.EXPECT().Cost(gomock.Any()).Return(&order.OrderCost{
							WithDiscount:    amount,
							WithoutDiscount: amount,
						}, nil).Times(1)
						orders[paymentId] = orderMock
					}
					return orders
				},
			},
			want: func() (order.AllowStatus, *order.DisallowReasonWithInfo, error) {
				return order.AllowStatusLimited, order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdMinCost, order.SubsystemBasket), nil
			},
		},
		{
			name: "error case",
			args: &args{
				ctx: context.Background(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless: payment.NewResolvedPaymentId(order.PaymentIdCashless),
				},
				orders: func(ctrl *gomock.Controller) map[order.PaymentId]order.Order {
					amounts := map[order.PaymentId]int{
						order.PaymentIdCashless: 0,
					}
					orders := make(map[order.PaymentId]order.Order)
					for paymentId, _ := range amounts {
						orderMock := order.NewMockOrder(ctrl)
						orderMock.EXPECT().Cost(gomock.Any()).Return(nil, errors.New("error")).Times(1)
						orders[paymentId] = orderMock
					}
					return orders
				},
			},
			want: func() (order.AllowStatus, *order.DisallowReasonWithInfo, error) {
				return order.AllowStatusAllow, nil, errors.New("error")
			},
		},
	}

	for _, tt := range tests {
		resolver := NewMinCost()

		t.Run(tt.name, func(t *testing.T) {
			orders := tt.args.orders(gomock.NewController(t))
			if len(tt.args.resolvedIdsMap) != len(orders) {
				t.Errorf("length map (%d) and orders (%d) are different", len(tt.args.resolvedIdsMap), len(orders))
			}
			wantStatus, wantReason, wantErr := tt.want()
			for paymentId, resolvedId := range tt.args.resolvedIdsMap {
				resolvedIdsMap := payment.ResolvedPaymentIdMap{paymentId: resolvedId}
				err := resolver.Resolve(tt.args.ctx, resolvedIdsMap, orders[paymentId])
				if err != nil {
					assert.EqualError(t, err, fmt.Errorf("error getting order cost: %w", wantErr).Error())
				}
				for _, resolvedId := range resolvedIdsMap {
					var resolvedReason *order.DisallowReasonWithInfo
					if len(resolvedId.Reasons()) > 0 {
						resolvedReason = resolvedId.Reasons()[0]
					}
					assert.Equal(t, wantStatus, resolvedId.Status())
					assert.Equal(t, wantReason, resolvedReason)
				}
			}
		})
	}
}
