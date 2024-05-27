package resolver

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/mock"
	"go.citilink.cloud/order/internal/order/payment"
	"testing"
)

func TestSameDayCitilinkCourierDelivery_Resolve(t *testing.T) {
	type expectedPropertiesPerPayment map[order.PaymentId]struct {
		status  order.AllowStatus
		reasons []*order.DisallowReasonWithInfo
	}
	type args struct {
		ctx            context.Context
		resolvedIdsMap payment.ResolvedPaymentIdMap
		ordr           func(ctrl *gomock.Controller) order.Order
	}

	tests := []struct {
		name string
		args args
		want expectedPropertiesPerPayment
	}{
		{
			"same day delivery",
			args{
				context.Background(),
				func() payment.ResolvedPaymentIdMap {
					ret := payment.ResolvedPaymentIdMap{
						order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
						order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
					}

					return ret
				}(), func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					dlvr := mock.NewMockDelivery(ctrl)
					courier := mock.NewMockCitilinkCourierDelivery(ctrl)

					dlvr.EXPECT().Id().Return(order.DeliveryIdCitilinkCourier)
					courier.EXPECT().Id().Return(order.CitilinkCourierDeliveryIdSameDay)
					dlvr.EXPECT().CitilinkCourier(context.Background()).Return(courier)
					ordr.EXPECT().Delivery().Return(dlvr).AnyTimes()
					return ordr
				},
			},
			expectedPropertiesPerPayment{
				order.PaymentIdInstallments: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status: order.AllowStatusLimited,
					reasons: []*order.DisallowReasonWithInfo{order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdSameDayCitilinkCourierDelivery, order.SubsystemDelivery).
						WithMessage("Рассрочка недоступна при доставке день-в-день.")},
				},
				order.PaymentIdCredit: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status: order.AllowStatusLimited,
					reasons: []*order.DisallowReasonWithInfo{order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdSameDayCitilinkCourierDelivery, order.SubsystemDelivery).
						WithMessage("Кредит недоступен при доставке день-в-день.")},
				},
			},
		},
		{
			"not same day delivery",
			args{
				context.Background(),
				func() payment.ResolvedPaymentIdMap {
					ret := payment.ResolvedPaymentIdMap{
						order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
						order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
					}
					return ret
				}(),
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					dlvr := mock.NewMockDelivery(ctrl)
					courier := mock.NewMockCitilinkCourierDelivery(ctrl)

					dlvr.EXPECT().Id().Return(order.DeliveryIdCitilinkCourier)
					courier.EXPECT().Id().Return(order.CitilinkCourierDeliveryIdFast)
					dlvr.EXPECT().CitilinkCourier(context.Background()).Return(courier)
					ordr.EXPECT().Delivery().Return(dlvr).AnyTimes()
					return ordr
				},
			},
			expectedPropertiesPerPayment{
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
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := NewSameDayCitilinkCourierDelivery()
			err := r.Resolve(test.args.ctx, test.args.resolvedIdsMap, test.args.ordr(gomock.NewController(t)))
			assert.NoError(t, err)
			gotPaymentMap := make(expectedPropertiesPerPayment)
			for pmntId, pmnt := range test.args.resolvedIdsMap {
				gotPaymentMap[pmntId] = struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status:  pmnt.Status(),
					reasons: pmnt.Reasons(),
				}
			}
			assert.Equal(t, gotPaymentMap, test.want)
		})
	}
}
