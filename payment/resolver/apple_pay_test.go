package resolver

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
	"testing"
)

func TestNewApplePay_Resolve(t *testing.T) {
	applePay := NewApplePay()
	assert.Equal(t, &ApplePay{}, applePay)
}

func TestApplePay_Resolve(t *testing.T) {
	tests := []struct {
		name                 string
		orderMock            func(ctrl *gomock.Controller) order.Order
		resolvedPaymentIdMap func() payment.ResolvedPaymentIdMap
		expect               func() interface{}
	}{
		{
			name: "GetApplePayAllowed is true",
			orderMock: func(ctrl *gomock.Controller) order.Order {
				orderMock := order.NewMockOrder(ctrl)
				orderMock.EXPECT().GetApplePayAllowed().Return(true)
				return orderMock
			},
			resolvedPaymentIdMap: func() payment.ResolvedPaymentIdMap {
				resolvedMap := make(payment.ResolvedPaymentIdMap)
				resolvedMap[order.PaymentIdSberbankBusinessOnline] = &payment.ResolvedId{}
				return make(payment.ResolvedPaymentIdMap)
			},
			expect: func() interface{} {
				return nil
			},
		},
		{
			name: "GetApplePayAllowed is false",
			orderMock: func(ctrl *gomock.Controller) order.Order {
				orderMock := order.NewMockOrder(ctrl)
				orderMock.EXPECT().GetApplePayAllowed().Return(false)
				return orderMock
			},
			resolvedPaymentIdMap: func() payment.ResolvedPaymentIdMap {
				resolvedMap := make(payment.ResolvedPaymentIdMap)
				resolvedMap[order.PaymentIdApplePay] = &payment.ResolvedId{}
				return resolvedMap
			},
			expect: func() interface{} {
				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			applePay := NewApplePay()
			orderMock := test.orderMock(gomock.NewController(t))
			resolvedMap := test.resolvedPaymentIdMap()

			var err error
			err = applePay.Resolve(context.Background(), resolvedMap, orderMock)
			assert.NoError(t, err)

			if len(resolvedMap) > 0 {
				assert.Equal(t, order.AllowStatusDisallowed, resolvedMap[order.PaymentIdApplePay].Status())
				assert.NotEmpty(t, resolvedMap[order.PaymentIdApplePay].Reasons())
			}
		})
	}
}

func TestApplePay_ResolvePanicMapKey(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name                 string
		orderMock            func() order.Order
		resolvedPaymentIdMap func() payment.ResolvedPaymentIdMap
	}{
		{
			name: "GetApplePayAllowed is false and empty ResolvedPaymentIdMap key",
			orderMock: func() order.Order {
				orderMock := order.NewMockOrder(ctrl)
				orderMock.EXPECT().GetApplePayAllowed().Return(false)
				return orderMock
			},
			resolvedPaymentIdMap: func() payment.ResolvedPaymentIdMap {
				return make(payment.ResolvedPaymentIdMap)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			applePay := NewApplePay()

			var ptf assert.PanicTestFunc
			ptf = func() {
				_ = applePay.Resolve(context.Background(), test.resolvedPaymentIdMap(), test.orderMock())
			}
			assert.Panics(t, ptf)
		})
	}
}
