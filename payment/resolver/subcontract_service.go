package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/basket"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewSubcontractService() *SubcontractService {
	return &SubcontractService{}
}

type SubcontractService struct {
}

// Resolve блокирует некоторые типы оплаты для позиций subcontractService
func (r *SubcontractService) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	bsk, err := ordr.Basket(ctx)
	if err != nil {
		return err
	}

	subcontractServiceItems := bsk.Find(basket.Finders.ByType(basket_item.TypeSubcontractServiceForProduct))
	if len(subcontractServiceItems) == 0 {
		return nil
	}

	// для b2b блокируем PaymentIdSberbankBusinessOnline если в корзине есть услуги субподряда
	if ordr.User() != nil && ordr.User().GetB2B().GetIsB2BState() {
		resolvedIdsMap[order.PaymentIdSberbankBusinessOnline].Disallow(
			order.AllowStatusLimited,
			order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdItemNotAvailableForPaymentMethod, order.SubsystemBasket).
				WithAdditions(order.DisallowReasonWithInfoAdditions{
					Basket: order.NewBasketDisallowAdditions(subcontractServiceItems.UniqIds()),
				}),
		)
	}

	return nil
}
