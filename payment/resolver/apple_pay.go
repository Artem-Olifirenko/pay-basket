package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewApplePay() *ApplePay {
	return &ApplePay{}
}

type ApplePay struct {
}

// Resolve резолвер блокирует тип оплаты Apple Pay, если такой тип оплаты не доступен для пользователя
func (r *ApplePay) Resolve(_ context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, o order.Order) error {
	if o.GetApplePayAllowed() {
		return nil
	}

	resolvedIdsMap[order.PaymentIdApplePay].Disallow(
		order.AllowStatusDisallowed,
		order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdApplePay, order.SubsystemPaymentMethod),
	)

	return nil
}
