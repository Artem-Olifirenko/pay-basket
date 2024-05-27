package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewTerminal() *Terminal {
	return &Terminal{}
}

type Terminal struct {
}

// Resolve резолвер блокирует типы оплаты для терминалов
func (t *Terminal) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	if ordr.GetKioskSource(ctx) == order.Terminal {
		// перебираем доступные методы оплаты
		for _, resolvedId := range resolvedIdsMap {
			switch resolvedId.Id() {
			case order.PaymentIdCashWithCard, order.PaymentIdCash, order.PaymentIdCards:
				continue
			default:
				resolvedId.Disallow(
					order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentTerminalRestriction, order.SubsystemPaymentMethod),
				)
			}
		}
	}
	return nil
}
