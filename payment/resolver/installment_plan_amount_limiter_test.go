package resolver

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
	"testing"
)

func TestInstallmentPlanAmountLimiter_Resolve(t *testing.T) {
	type mocks struct {
		order *order.MockOrder
	}

	type args struct {
		ctx            context.Context
		resolvedIdsMap payment.ResolvedPaymentIdMap
		amount         int
	}

	tests := []struct {
		args    args
		mocks   func(ctrl *gomock.Controller) *mocks
		want    func(m *mocks, ctx context.Context) *order.MockOrder
		name    string
		wantErr bool
		err     error
	}{
		{
			name:    "order cost error",
			wantErr: true,
			err:     fmt.Errorf("can't get order cost: can't get basket"),
			args: args{
				ctx:            context.Background(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{},
				amount:         1199,
			},
			want: func(m *mocks, ctx context.Context) *order.MockOrder {
				m.order.EXPECT().
					Cost(ctx).
					Return(nil, fmt.Errorf("can't get basket")).
					Times(1)
				return m.order
			},
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					order: order.NewMockOrder(ctrl),
				}
			},
		},
		{
			name:    "order price below the minimum for credit - reasons",
			wantErr: false,
			args: args{
				ctx: context.Background(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
				},
				amount: 1200,
			},
			want: func(m *mocks, ctx context.Context) *order.MockOrder {
				m.order.EXPECT().
					Cost(ctx).
					Return(&order.OrderCost{WithoutDiscount: 1300, WithDiscount: 1199}, nil).
					Times(1)
				return m.order
			},
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					order: order.NewMockOrder(ctrl),
				}
			},
		},
		{
			name:    "the price of the order is higher than the minimum for credit - no reasons",
			wantErr: false,
			args: args{
				ctx:            context.Background(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{},
				amount:         1200,
			},
			want: func(m *mocks, ctx context.Context) *order.MockOrder {
				m.order.EXPECT().
					Cost(ctx).
					Return(&order.OrderCost{WithoutDiscount: 1300, WithDiscount: 1200}, nil).
					Times(1)
				return m.order
			},
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					order: order.NewMockOrder(ctrl),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mocks := tt.mocks(ctrl)
			wantOrder := tt.want(mocks, tt.args.ctx)

			resolver := NewInstallmentPlanAmountLimiter(tt.args.amount)
			resolverMap := tt.args.resolvedIdsMap

			err := resolver.Resolve(tt.args.ctx, resolverMap, wantOrder)
			if tt.wantErr {
				assert.EqualError(t, tt.err, err.Error())
				return
			}

			paymentId, ok := resolverMap[order.PaymentIdInstallments]
			if !ok {
				assert.Empty(t, resolverMap)
				return
			}

			factReason := paymentId.Reasons()[0].Reason()
			assert.Equal(t, order.DisallowReasonPaymentIdItemNotAvailableForPaymentMethod, factReason)
		})
	}

}
