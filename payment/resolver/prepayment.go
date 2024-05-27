package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewPrepayment() *Prepayment {
	return &Prepayment{}
}

type Prepayment struct {
}

func (r *Prepayment) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	bsk, err := ordr.Basket(ctx)
	if err != nil {
		return err
	}

	itemsWithPrepayment := make(basket_item.Items, 0)
	for _, item := range bsk.All() {
		if item.IsPrepaymentMandatory() {
			itemsWithPrepayment = append(itemsWithPrepayment, item)
		}
	}

	allowedIdsMap := map[order.PaymentId]bool{
		order.PaymentIdCardsOnline:            true,
		order.PaymentIdYandex:                 true,
		order.PaymentIdSberbankBusinessOnline: true,
		order.PaymentIdCashless:               true,
		order.PaymentIdCredit:                 true,
		order.PaymentIdInstallments:           true,
		order.PaymentIdSbp:                    true,
		order.PaymentIdSberPay:                true,
	}

	if len(itemsWithPrepayment) > 0 {
		for _, paymentType := range resolvedIdsMap {
			if _, ok := allowedIdsMap[paymentType.Id()]; !ok {
				paymentType.Disallow(
					order.AllowStatusLimited,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdPrepayment, order.SubsystemBasket).
						WithMessage("В корзине есть товары, доступные только по стопроцентной предоплате").
						WithAdditions(order.DisallowReasonWithInfoAdditions{
							Basket: order.NewBasketDisallowAdditions(itemsWithPrepayment.UniqIds()),
						}))
			}
		}
	}

	return nil
}
