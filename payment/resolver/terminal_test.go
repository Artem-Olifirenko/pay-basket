package resolver

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
	"testing"
)

func TestNewTerminal_Resolve(t *testing.T) {
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
			"terminal payments restriction",
			args{
				payment.ResolvedPaymentIdMap{
					order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
					order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCardsOnline:  payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
					order.PaymentIdYandex:       payment.NewResolvedPaymentId(order.PaymentIdYandex),
					order.PaymentIdCashWithCard: payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
					order.PaymentIdSbp:          payment.NewResolvedPaymentId(order.PaymentIdSbp),
				},
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().GetKioskSource(gomock.Any()).Return(order.Terminal)
					return ordr
				},
			},
			func() payment.ResolvedPaymentIdMap {
				ret := payment.ResolvedPaymentIdMap{
					order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
					order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCardsOnline:  payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
					order.PaymentIdYandex:       payment.NewResolvedPaymentId(order.PaymentIdYandex),
					order.PaymentIdCashWithCard: payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
					order.PaymentIdSbp:          payment.NewResolvedPaymentId(order.PaymentIdSbp),
				}

				ret[order.PaymentIdSbp].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentTerminalRestriction, order.SubsystemPaymentMethod))
				ret[order.PaymentIdInstallments].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentTerminalRestriction, order.SubsystemPaymentMethod))
				ret[order.PaymentIdCredit].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentTerminalRestriction, order.SubsystemPaymentMethod))
				ret[order.PaymentIdCardsOnline].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentTerminalRestriction, order.SubsystemPaymentMethod))
				ret[order.PaymentIdYandex].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentTerminalRestriction, order.SubsystemPaymentMethod))
				return ret
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewTerminal()

			err := resolver.Resolve(context.TODO(), test.args.resolvedIdsMap, test.args.ordr(gomock.NewController(t)))
			got := test.args.resolvedIdsMap
			want := test.want()

			assert.NoError(t, err)
			assert.Equal(t, got, want)
		})
	}
}
