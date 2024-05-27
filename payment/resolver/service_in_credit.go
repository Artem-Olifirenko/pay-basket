package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewServiceInCredit() *ServiceInCredit {
	return &ServiceInCredit{}
}

type ServiceInCredit struct {
}

// Resolve блокирует кредитные типы оплаты, если в корзине есть позиции, которые являются сервисом (IsService=true) и
// для таких позиций недоступен признак доступного кредита (IsCreditAvail=false)
func (r *ServiceInCredit) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	bsk, err := ordr.Basket(ctx)
	if err != nil {
		return err
	}

	var itemsNotAllowed basket_item.Items
	for _, item := range bsk.All() {
		if item.Type().IsPartOfConfiguration() {
			continue
		}

		if !item.Type().IsService() {
			continue
		}

		if !item.Additions().GetService().GetIsCreditAvail() {
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
