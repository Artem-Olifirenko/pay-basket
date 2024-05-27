package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

type CreditAndInstallments struct {
}

func NewCreditAndInstallments() *CreditAndInstallments {
	return &CreditAndInstallments{}
}

// Resolve резолвер блокирует тип оплаты = кредит, если возможна рассрочка для заказа
func (i *CreditAndInstallments) Resolve(_ context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, _ order.Order) error {
	// не показываем кредит, если рассрочка возможна для заказа
	isInstallmentsAvailable := true
	for _, disallowReason := range resolvedIdsMap[order.PaymentIdInstallments].Reasons() {
		if disallowReason.Reason() == order.DisallowReasonPaymentIdItemNotAvailableForPaymentMethod {
			isInstallmentsAvailable = false
			break
		}
	}
	if isInstallmentsAvailable {
		resolvedIdsMap[order.PaymentIdCredit].Disallow(
			order.AllowStatusDisallowed,
			order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdItemNotAvailableForPaymentMethod, order.SubsystemBasket),
		)
	}

	return nil
}
