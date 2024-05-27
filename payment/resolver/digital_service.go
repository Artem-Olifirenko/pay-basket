package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/basket"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewDigitalService() *DigitalService {
	return &DigitalService{}
}

type DigitalService struct {
}

// Resolve блокирует некоторые типы оплаты для позиций digitalService
func (r *DigitalService) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	bsk, err := ordr.Basket(ctx)
	if err != nil {
		return err
	}

	digitalServiceItems := bsk.Find(basket.Finders.ByType(basket_item.TypeDigitalService))
	if len(digitalServiceItems) > 0 {
		resolvedIdsMap[order.PaymentIdSberbankBusinessOnline].Disallow(
			order.AllowStatusLimited,
			order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdItemNotAvailableForPaymentMethod, order.SubsystemBasket).
				WithAdditions(order.DisallowReasonWithInfoAdditions{
					Basket: order.NewBasketDisallowAdditions(digitalServiceItems.UniqIds()),
				}),
		)
	}

	return nil
}
