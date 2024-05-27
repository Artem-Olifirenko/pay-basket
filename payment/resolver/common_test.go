package resolver

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
	"testing"
)

func TestCommon_Resolve(t *testing.T) {
	type expectedPropertiesPerPayment map[order.PaymentId]struct {
		status  order.AllowStatus
		reasons []*order.DisallowReasonWithInfo
	}

	type args struct {
		ctx            context.Context
		resolvedIdsMap payment.ResolvedPaymentIdMap
		order          func(ctrl *gomock.Controller) order.Order
	}
	tests := []struct {
		name                 string
		args                 args
		wantPaymentStatusMap expectedPropertiesPerPayment
		wantErr              error
	}{
		{
			name: "payment invalid, terminal pinpad, webmoney, terminalorcashbox must be disallowed with corresponding reasons and any other payment id's status is not affected",
			args: args{
				ctx: context.Background(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdInvalid:                payment.NewResolvedPaymentId(order.PaymentIdInvalid),
					order.PaymentIdTerminalPinpad:         payment.NewResolvedPaymentId(order.PaymentIdTerminalPinpad),
					order.PaymentIdWebmoney:               payment.NewResolvedPaymentId(order.PaymentIdWebmoney),
					order.PaymentIdTerminalOrCashbox:      payment.NewResolvedPaymentId(order.PaymentIdTerminalOrCashbox),
					order.PaymentIdInstallments:           payment.NewResolvedPaymentId(order.PaymentIdInstallments),
					order.PaymentIdCredit:                 payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdUnknown:                payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCash:                   payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCardsOnline:            payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdYandex:                 payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdSbp:                    payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdApplePay:               payment.NewResolvedPaymentId(order.PaymentIdCredit),
				},
				order: func(ctrl *gomock.Controller) order.Order {
					return order.NewMockOrder(ctrl)
				},
			},
			wantPaymentStatusMap: expectedPropertiesPerPayment{
				order.PaymentIdInvalid: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status: order.AllowStatusDisallowed,
					reasons: []*order.DisallowReasonWithInfo{
						order.NewDisallowReasonWithInfo(order.DisallowReasonUnknown, order.SubsystemUnknown),
					},
				},
				order.PaymentIdTerminalPinpad: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status: order.AllowStatusDisallowed,
					reasons: []*order.DisallowReasonWithInfo{
						order.NewDisallowReasonWithInfo(order.DisallowReasonUnknown, order.SubsystemUnknown),
					},
				},
				order.PaymentIdWebmoney: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status: order.AllowStatusDisallowed,
					reasons: []*order.DisallowReasonWithInfo{
						order.NewDisallowReasonWithInfo(order.DisallowReasonUnknown, order.SubsystemUnknown),
					},
				},
				order.PaymentIdTerminalOrCashbox: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status: order.AllowStatusDisallowed,
					reasons: []*order.DisallowReasonWithInfo{
						order.NewDisallowReasonWithInfo(order.DisallowReasonUnknown, order.SubsystemUnknown),
					},
				},
				order.PaymentIdInstallments: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status:  order.AllowStatusAllow,
					reasons: nil,
				},
				order.PaymentIdCredit: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status:  order.AllowStatusAllow,
					reasons: nil,
				},
				order.PaymentIdUnknown: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status:  order.AllowStatusAllow,
					reasons: nil,
				},
				order.PaymentIdCashless: struct {
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
				order.PaymentIdCardsOnline: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status:  order.AllowStatusAllow,
					reasons: nil,
				},
				order.PaymentIdYandex: struct {
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
				order.PaymentIdSbp: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status:  order.AllowStatusAllow,
					reasons: nil,
				},
				order.PaymentIdCashWithCard: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status:  order.AllowStatusAllow,
					reasons: nil,
				},
				order.PaymentIdApplePay: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status:  order.AllowStatusAllow,
					reasons: nil,
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			r := NewCommon()
			if err := r.Resolve(tt.args.ctx, tt.args.resolvedIdsMap, tt.args.order(ctrl)); (err != nil) && (err != tt.wantErr) {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
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
			assert.Equal(t, gotPaymentMap, tt.wantPaymentStatusMap)
		})
	}
}
