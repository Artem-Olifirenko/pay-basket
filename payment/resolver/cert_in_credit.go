package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewCertInCredit() *CertInCredit {
	return &CertInCredit{}
}

type CertInCredit struct {
}

// Resolve resolver блокирует типы оплаты = кредит,
// если в корзине есть продукты из 219 категории ("Подарочные сертификаты")
// и тип таких продуктов НЕ является частью конфигурации
func (r *CertInCredit) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	bsk, err := ordr.Basket(ctx)
	if err != nil {
		return err
	}

	var itemsNotAllowed basket_item.Items
	for _, item := range bsk.All() {
		if item.Type().IsPartOfConfiguration() {
			continue
		}

		if !item.Type().IsProduct() {
			continue
		}

		if item.Additions().GetProduct().CategoryId() == basket_item.GiftCertificateCategoryID {
			itemsNotAllowed = append(itemsNotAllowed, item)
		}
	}

	if len(itemsNotAllowed) > 0 {
		for _, pmt := range resolvedIdsMap {
			if !pmt.Id().IsCredit() {
				continue
			}

			pmt.Disallow(
				order.AllowStatusLimited,
				order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdItemNotAvailableForPaymentMethod, order.SubsystemBasket).
					WithAdditions(order.DisallowReasonWithInfoAdditions{
						Basket: order.NewBasketDisallowAdditions(itemsNotAllowed.UniqIds()),
					}),
			)
		}
	}

	return nil
}
