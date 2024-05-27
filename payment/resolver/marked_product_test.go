package resolver

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
)

func TestMarkedProduct_Resolve(t *testing.T) {
	type expectedPropertiesPerPayment map[order.PaymentId]struct {
		status  order.AllowStatus
		reasons []*order.DisallowReasonWithInfo
	}
	type fields struct {
		isMarkingEnabled bool
	}
	type args struct {
		ctx  context.Context
		ordr func(ctrl *gomock.Controller) order.Order
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    expectedPropertiesPerPayment
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "basket returns err",
			fields: fields{
				isMarkingEnabled: false,
			},
			args: args{
				ctx: context.Background(),
				ordr: func(ctrl *gomock.Controller) order.Order {
					o := order.NewMockOrder(ctrl)
					o.EXPECT().Basket(gomock.Any()).Return(
						nil,
						fmt.Errorf("error getting basket"),
					)
					return o
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "marked disable",
			fields: fields{
				isMarkingEnabled: false,
			},
			args: args{
				ctx: context.Background(),
				ordr: func(ctrl *gomock.Controller) order.Order {
					o := order.NewMockOrder(ctrl)
					o.EXPECT().Basket(gomock.Any()).Return(
						func() order.Basket {
							b := order.NewMockBasket(ctrl)
							b.EXPECT().IsMarkingAvailable().Return(false)
							return b
						}(),
						nil,
					)
					return o
				},
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
			wantErr: assert.NoError,
		},
		{
			name: "user not b2b",
			fields: fields{
				isMarkingEnabled: true,
			},
			args: args{
				ctx: context.Background(),
				ordr: func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().Basket(gomock.Any()).Return(
						func() order.Basket {
							b := order.NewMockBasket(ctrl)
							b.EXPECT().IsMarkingAvailable().Return(true)
							return b
						}(),
						nil,
					)
					ordr.EXPECT().User().Return(&userv1.User{
						Id: "1",
						B2B: &userv1.User_B2B{
							IsB2BState: false,
						},
					}).Times(3)
					return ordr
				},
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
			wantErr: assert.NoError,
		},
		{
			name: "marked product in msk",
			fields: fields{
				isMarkingEnabled: true,
			},
			args: args{
				ctx: context.Background(),
				ordr: func(ctrl *gomock.Controller) order.Order {
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
							b.EXPECT().IsMarkingAvailable().Return(true)
							return b
						}(),
						nil,
					).Times(1)
					return ordr
				},
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
					reasons: []*order.DisallowReasonWithInfo{order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentForMarkedProducts, order.SubsystemPaymentMethod).
						WithMessage("Данный способ оплаты недоступен для маркированных товаров.")},
				},
				order.PaymentIdCardsOnline: struct {
					status  order.AllowStatus
					reasons []*order.DisallowReasonWithInfo
				}{
					status: order.AllowStatusLimited,
					reasons: []*order.DisallowReasonWithInfo{order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentForMarkedProducts, order.SubsystemPaymentMethod).
						WithMessage("Данный способ оплаты недоступен для маркированных товаров.")},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		resolvedIdsMap := payment.ResolvedPaymentIdMap{
			order.PaymentIdCashless:    payment.NewResolvedPaymentId(order.PaymentIdCashless),
			order.PaymentIdCash:        payment.NewResolvedPaymentId(order.PaymentIdCash),
			order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
		}
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			r := &markedProduct{}
			err := r.Resolve(tt.args.ctx, resolvedIdsMap, tt.args.ordr(ctrl))
			if !tt.wantErr(t, err, fmt.Sprintf("Resolve expected and actual error values are different")) {
				return
			} else if tt.want != nil {
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
			}
		})
	}
}

func TestNewMarkedProduct(t *testing.T) {
	expected := &markedProduct{}
	actual := NewMarkedProduct()
	assert.Equal(t, actual, expected)
}
