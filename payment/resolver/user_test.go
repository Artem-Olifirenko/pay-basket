package resolver

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
	"testing"
)

func TestNewUser_Resolve(t *testing.T) {
	type args struct {
		resolvedIdsMap payment.ResolvedPaymentIdMap
		ordr           func(ctrl *gomock.Controller) order.Order
	}

	tests := []struct {
		name string
		args args
		want func() payment.ResolvedPaymentIdMap
	}{
		{
			"user payments only",
			args{
				payment.ResolvedPaymentIdMap{
					order.PaymentIdSberPay:                payment.NewResolvedPaymentId(order.PaymentIdSberPay),
					order.PaymentIdCards:                  payment.NewResolvedPaymentId(order.PaymentIdCards),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
					order.PaymentIdSbp:                    payment.NewResolvedPaymentId(order.PaymentIdSbp),
					order.PaymentIdInstallments:           payment.NewResolvedPaymentId(order.PaymentIdInstallments),
					order.PaymentIdYandex:                 payment.NewResolvedPaymentId(order.PaymentIdYandex),
					order.PaymentIdCredit:                 payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCardsOnline:            payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
					order.PaymentIdCash:                   payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
				},
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					usr := &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: false,
						},
					}
					ordr.EXPECT().User().Times(2).Return(usr)
					return ordr
				},
			},
			func() payment.ResolvedPaymentIdMap {
				ret := payment.ResolvedPaymentIdMap{
					order.PaymentIdSberPay:                payment.NewResolvedPaymentId(order.PaymentIdSberPay),
					order.PaymentIdCards:                  payment.NewResolvedPaymentId(order.PaymentIdCards),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
					order.PaymentIdSbp:                    payment.NewResolvedPaymentId(order.PaymentIdSbp),
					order.PaymentIdInstallments:           payment.NewResolvedPaymentId(order.PaymentIdInstallments),
					order.PaymentIdYandex:                 payment.NewResolvedPaymentId(order.PaymentIdYandex),
					order.PaymentIdCredit:                 payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCardsOnline:            payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
					order.PaymentIdCash:                   payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
				}
				ret[order.PaymentIdCashless].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonUnknown, order.SubsystemPaymentMethod))
				ret[order.PaymentIdSberbankBusinessOnline].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonUnknown, order.SubsystemPaymentMethod))
				return ret
			},
		},
		{
			"empty payments list",
			args{
				payment.ResolvedPaymentIdMap{},
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					usr := &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: false,
						},
					}
					ordr.EXPECT().User().Times(2).Return(usr)
					return ordr
				},
			},
			func() payment.ResolvedPaymentIdMap {
				return payment.ResolvedPaymentIdMap{}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewUser()

			err := resolver.Resolve(context.TODO(), test.args.resolvedIdsMap, test.args.ordr(gomock.NewController(t)))
			got := test.args.resolvedIdsMap
			want := test.want()

			assert.NoError(t, err)
			assert.Equal(t, got, want)
		})
	}
}
