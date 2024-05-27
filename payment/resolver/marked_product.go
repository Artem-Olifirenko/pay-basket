package resolver

import (
	"context"

	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

type markedProduct struct{}

// возможные типы оплаты для маркированных продуктов
var allowedPaymentIdsForMarketProducts = map[order.PaymentId]bool{
	order.PaymentIdCashless:               true,
	order.PaymentIdSberbankBusinessOnline: true,
}

func NewMarkedProduct() *markedProduct {
	return &markedProduct{}
}

// Resolve ограничивает способы оплаты для маркированных товаров кроме списка в allowedIdsMap
func (m *markedProduct) Resolve(
	ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap,
	ordr order.Order,
) error {
	bsk, err := ordr.Basket(ctx)
	if err != nil {
		return err
	}
	if bsk.IsMarkingAvailable() {
		// Если пользователь не b2b, то никакие типы оплаты тут не блокируем
		if ordr.User() == nil ||
			ordr.User().GetB2B() == nil ||
			!ordr.User().GetB2B().GetIsB2BState() {
			return nil
		}

		for _, paymentType := range resolvedIdsMap {
			if _, ok := allowedPaymentIdsForMarketProducts[paymentType.Id()]; !ok {
				paymentType.Disallow(
					order.AllowStatusLimited,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentForMarkedProducts, order.SubsystemPaymentMethod).
						WithMessage("Данный способ оплаты недоступен для маркированных товаров."),
				)
			}
		}
	}

	return nil
}
