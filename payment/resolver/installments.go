package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/payment"
)

type Installments struct {
}

func NewInstallments() *Installments {
	return &Installments{}
}

// Resolve блокирует оплату в рассрочку для некоторых позиций в корзине
func (i *Installments) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	bsk, err := ordr.Basket(ctx)
	if err != nil {
		return err
	}

	var itemsNotAllowed basket_item.Items
	for _, item := range bsk.All() {
		// Услугу сборки и услуги товаров в конфигурации можно покупать в рассрочку в любом случае
		if item.Type() == basket_item.TypeConfigurationAssemblyService || item.Type() == basket_item.TypeConfigurationProductService {
			continue
		}
		// если позиция является сервисом и данная позиция недоступна для покупки в рассрочку, то данная позиция должна быть ограничена для покупки в рассрочку
		if item.Type().IsService() &&
			!item.Additions().GetService().GetIsAvailableForInstallments() {
			itemsNotAllowed = append(itemsNotAllowed, item)
		}

		// если позиция является продуктом и у позиции нет CreditProgram (неясно что это), то данная позиция должна быть ограничена для покупки в рассрочку
		if item.Type().IsProduct() &&
			len(item.Additions().GetProduct().CreditPrograms()) == 0 {
			itemsNotAllowed = append(itemsNotAllowed, item)
		}
	}

	if len(itemsNotAllowed) > 0 {
		resolvedIdsMap[order.PaymentIdInstallments].Disallow(
			order.AllowStatusDisallowed,
			order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdItemNotAvailableForPaymentMethod, order.SubsystemBasket).
				WithAdditions(order.DisallowReasonWithInfoAdditions{
					Basket: order.NewBasketDisallowAdditions(itemsNotAllowed.UniqIds()),
				}),
		)
	}

	return nil
}
