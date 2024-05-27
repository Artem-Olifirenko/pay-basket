package resolver

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
	"testing"
)

func TestCreditAndInstallments_Resolve(t *testing.T) {
	type resolverArgs struct {
		pmtMap payment.ResolvedPaymentIdMap
		ordr   func(ctrl *gomock.Controller) order.Order
	}

	type paymentStatusMap map[order.PaymentId]order.AllowStatus

	tests := []struct {
		name string
		args resolverArgs
		want paymentStatusMap
	}{
		{
			"installments not available",
			resolverArgs{
				func() payment.ResolvedPaymentIdMap {
					ret := payment.ResolvedPaymentIdMap{
						order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
						order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
					}

					ret[order.PaymentIdInstallments].Disallow(
						order.AllowStatusLimited,
						order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdItemNotAvailableForPaymentMethod, order.SubsystemUnknown),
					)

					return ret
				}(),
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					return ordr
				},
			},
			paymentStatusMap{
				order.PaymentIdInstallments: order.AllowStatusLimited,
				order.PaymentIdCredit:       order.AllowStatusAllow,
			},
		},
		{
			"installments available",
			resolverArgs{
				func() payment.ResolvedPaymentIdMap {
					ret := payment.ResolvedPaymentIdMap{
						order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
						order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
					}

					return ret
				}(),
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					return ordr
				},
			},
			paymentStatusMap{
				order.PaymentIdInstallments: order.AllowStatusAllow,
				order.PaymentIdCredit:       order.AllowStatusDisallowed,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			CreditAndInstallments := NewCreditAndInstallments()
			err := CreditAndInstallments.Resolve(ctx, test.args.pmtMap, test.args.ordr(gomock.NewController(t)))

			gotPaymentMap := make(paymentStatusMap)
			for pmntId, pmnt := range test.args.pmtMap {
				gotPaymentMap[pmntId] = pmnt.Status()
			}

			assert.NoError(t, err)
			assert.Equal(t, gotPaymentMap, test.want)

		})
	}
}
