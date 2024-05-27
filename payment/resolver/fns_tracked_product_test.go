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

func TestFnsTrackedProduct_Resolve(t *testing.T) {
	ctrl := gomock.NewController(t)
	type expectedPropertiesPerPayment map[order.PaymentId]struct {
		status  order.AllowStatus
		reasons []*order.DisallowReasonWithInfo
	}
	type fields struct {
		isFnsTrackingEnabled bool
	}
	type args struct {
		ctx  context.Context
		ordr order.Order
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   expectedPropertiesPerPayment
	}{
		{
			name: "fns tracked disable",
			fields: fields{
				isFnsTrackingEnabled: false,
			},
			args: args{
				ctx:  context.Background(),
				ordr: order.NewMockOrder(ctrl),
			},
			want: expectedPropertiesPerPayment{
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
			},
		},
		{
			name: "user not b2b",
			fields: fields{
				isFnsTrackingEnabled: true,
			},
			args: args{
				ctx: context.Background(),
				ordr: func() order.Order {
					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().User().Return(&userv1.User{
						Id: "1",
						B2B: &userv1.User_B2B{
							IsB2BState: false,
						},
					}).Times(3)
					return ordr
				}(),
			},
			want: expectedPropertiesPerPayment{
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
			},
		},
		{
			name: "without fns tracked product",
			fields: fields{
				isFnsTrackingEnabled: true,
			},
			args: args{
				ctx: context.Background(),
				ordr: func() order.Order {
					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().User().Return(&userv1.User{
						Id: "1",
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					}).Times(3)
					ordr.EXPECT().Basket(gomock.Any()).Return(
						func() order.Basket {
							b := order.NewMockBasket(ctrl)
							b.EXPECT().HasFnsTrackedProducts().Return(false)
							return b
						}(),
						nil,
					).Times(1)
					return ordr
				}(),
			},
			want: expectedPropertiesPerPayment{
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
			},
		},
		{
			name: "with fns tracked product",
			fields: fields{
				isFnsTrackingEnabled: true,
			},
			args: args{
				ctx: context.Background(),
				ordr: func() order.Order {
					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().User().Return(&userv1.User{
						Id: "1",
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					}).Times(3)
					ordr.EXPECT().Basket(gomock.Any()).Return(
						func() order.Basket {
							b := order.NewMockBasket(ctrl)
							b.EXPECT().HasFnsTrackedProducts().Return(true)
							return b
						}(),
						nil,
					).Times(1)
					return ordr
				}(),
			},
			want: expectedPropertiesPerPayment{
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
					status: order.AllowStatusLimited,
					reasons: []*order.DisallowReasonWithInfo{order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentForFnsTrackedProducts, order.SubsystemPaymentMethod).
						WithMessage("Данный способ оплаты недоступен для прослеживаемых товаров.")},
				},
				order.PaymentIdCardsOnline: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status: order.AllowStatusLimited,
					reasons: []*order.DisallowReasonWithInfo{order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentForFnsTrackedProducts, order.SubsystemPaymentMethod).
						WithMessage("Данный способ оплаты недоступен для прослеживаемых товаров.")},
				},
			},
		},
	}
	for _, tt := range tests {
		resolvedIdsMap := payment.ResolvedPaymentIdMap{
			order.PaymentIdCashless:    payment.NewResolvedPaymentId(order.PaymentIdCashless),
			order.PaymentIdCash:        payment.NewResolvedPaymentId(order.PaymentIdCash),
			order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
		}
		t.Run(tt.name, func(t *testing.T) {
			r := &fnsTrackedProduct{
				isFnsTrackingEnabled: tt.fields.isFnsTrackingEnabled,
			}
			err := r.Resolve(tt.args.ctx, resolvedIdsMap, tt.args.ordr)
			assert.NoError(t, err)
			gotPaymentMap := make(expectedPropertiesPerPayment)
			for pmntId, pmnt := range resolvedIdsMap {
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

func TestNewFnsTrackedProduct(t *testing.T) {
	expected := &fnsTrackedProduct{isFnsTrackingEnabled: true}
	actual := NewFnsTrackedProduct(true)
	assert.Equal(t, actual, expected)
}
