package resolver

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/mock"
	"go.citilink.cloud/order/internal/order/payment"
	"testing"
)

func TestCashWithCardCheck_NewCashWithCardCheck(t *testing.T) {
	type args struct {
		withoutCardCityIdsStorage internal.StringsContainer
	}
	tests := []struct {
		name string
		want func(withoutCardCityIdsStorage *internal.MockStringsContainer) *CashWithCardCheck
	}{
		{
			name: "correct new resolver create",
			want: func(withoutCardCityIdsStorage *internal.MockStringsContainer) *CashWithCardCheck {
				return &CashWithCardCheck{
					withoutCardCityIdsStorage: withoutCardCityIdsStorage,
				}
			},
		},
	}
	for _, tt := range tests {
		withoutCardCityIdsStorage := internal.NewMockStringsContainer(gomock.NewController(t))
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, NewCashWithCardCheck(withoutCardCityIdsStorage), tt.want(withoutCardCityIdsStorage))
		})
	}
}

func TestCashWithCardCheck_Resolve(t *testing.T) {
	type expectedPropertiesPerPayment map[order.PaymentId]struct {
		status  order.AllowStatus
		reasons []*order.DisallowReasonWithInfo
	}
	type args struct {
		ctx                       context.Context
		resolvedIdsMap            payment.ResolvedPaymentIdMap
		ordr                      func(ctrl *gomock.Controller) order.Order
		withoutCardCityIdsStorage func(ctrl *gomock.Controller) internal.StringsContainer
	}
	type testCase struct {
		name string
		args args
		want expectedPropertiesPerPayment
	}

	tests := []testCase{
		{
			name: "some_rnd_cl is avail cash",
			args: args{
				ctx: context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:         payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCards:        payment.NewResolvedPaymentId(order.PaymentIdCards),
					order.PaymentIdCashWithCard: payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
				},
				ordr: func(ctrl *gomock.Controller) order.Order {
					delivery := order.NewMockDelivery(ctrl)
					delivery.EXPECT().Id().Return(order.DeliveryIdCitilinkCourier).Times(1)

					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().Delivery().Return(delivery).Times(1)
					ordr.EXPECT().CityId().Return(order.CityId("some_rnd_cl")).Times(1)
					ordr.EXPECT().User().Times(1).Return(nil)
					return ordr
				},
				withoutCardCityIdsStorage: func(ctrl *gomock.Controller) internal.StringsContainer {
					withoutCardCityIdsStrg := internal.NewMockStringsContainer(ctrl)
					withoutCardCityIdsStrg.EXPECT().Contains("some_rnd_cl").Return(false).Times(1)
					return withoutCardCityIdsStrg
				},
			},
			want: expectedPropertiesPerPayment{
				order.PaymentIdCash: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status:  order.AllowStatusDisallowed,
					reasons: []*order.DisallowReasonWithInfo{order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCashNotAvailable, order.SubsystemUnknown)},
				},
				order.PaymentIdCards: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status:  order.AllowStatusDisallowed,
					reasons: []*order.DisallowReasonWithInfo{order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCardNotAvailable, order.SubsystemUnknown)},
				},
				order.PaymentIdCashWithCard: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status:  order.AllowStatusAllow,
					reasons: nil,
				},
			},
		},
	}

	withoutCardCityIds := []order.CityId{"exist_1", "exist_2", "exist_3"}

	for _, v := range withoutCardCityIds {
		tests = append(tests, testCase{
			name: "city list is not avail cash - " + string(v),
			args: args{
				ctx: context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:         payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCards:        payment.NewResolvedPaymentId(order.PaymentIdCards),
					order.PaymentIdCashWithCard: payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
				},
				ordr: func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					dlvr := mock.NewMockDelivery(ctrl)
					dlvr.EXPECT().Id().Return(order.DeliveryIdCitilinkCourier).Times(2)
					ordr.EXPECT().Delivery().Return(dlvr).Times(2)
					ordr.EXPECT().User().Times(1).Return(nil)
					ordr.EXPECT().CityId().Return(v).Times(1)
					return ordr
				},
				withoutCardCityIdsStorage: func(ctrl *gomock.Controller) internal.StringsContainer {
					withoutCardCityIdsStrg := internal.NewMockStringsContainer(ctrl)
					withoutCardCityIdsStrg.EXPECT().Contains(string(v)).Return(true).Times(1)
					return withoutCardCityIdsStrg
				},
			},
			want: expectedPropertiesPerPayment{
				order.PaymentIdCash: struct {
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
					status:  order.AllowStatusDisallowed,
					reasons: []*order.DisallowReasonWithInfo{order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCashWithCardNotAvailable, order.SubsystemUnknown)},
				},
				order.PaymentIdCards: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status:  order.AllowStatusDisallowed,
					reasons: []*order.DisallowReasonWithInfo{order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCardNotAvailable, order.SubsystemUnknown)},
				},
			},
		})
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)

		t.Run(tt.name, func(t *testing.T) {
			r := NewCashWithCardCheck(tt.args.withoutCardCityIdsStorage(ctrl))
			err := r.Resolve(tt.args.ctx, tt.args.resolvedIdsMap, tt.args.ordr(ctrl))
			assert.NoError(t, err)
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

			assert.Equal(t, tt.want, gotPaymentMap)
		})
	}
}
