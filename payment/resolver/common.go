package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewCommon() *Common {
	return &Common{}
}

type Common struct {
}

// Resolve блокирует некоторые типы оплаты и выставляет причины блока
func (r *Common) Resolve(
	_ context.Context,
	resolvedIdsMap payment.ResolvedPaymentIdMap,
	_ order.Order,
) error {
	resolvedIdsMap[order.PaymentIdInvalid].Disallow(
		order.AllowStatusDisallowed,
		order.NewDisallowReasonWithInfo(order.DisallowReasonUnknown, order.SubsystemUnknown),
	)
	resolvedIdsMap[order.PaymentIdTerminalPinpad].Disallow(
		order.AllowStatusDisallowed,
		order.NewDisallowReasonWithInfo(order.DisallowReasonUnknown, order.SubsystemUnknown),
	)
	// на данный момент времени оплата через WebMoney заблокирована со стороны бизнеса
	// так как это не проблема заказа, то блокировка была добавлена сюда с "неизвестной причиной"
	resolvedIdsMap[order.PaymentIdWebmoney].Disallow(
		order.AllowStatusDisallowed,
		order.NewDisallowReasonWithInfo(order.DisallowReasonUnknown, order.SubsystemUnknown),
	)

	resolvedIdsMap[order.PaymentIdTerminalOrCashbox].Disallow(
		order.AllowStatusDisallowed,
		order.NewDisallowReasonWithInfo(order.DisallowReasonUnknown, order.SubsystemUnknown),
	)
	return nil
}
