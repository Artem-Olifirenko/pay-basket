package resolver

import (
	"context"
	"fmt"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

// CreditAndInstallmentInStore резолвер определяющий доступность кредита и рассрочки в текущей выбранной точке самовывоза
type CreditAndInstallmentInStore struct {
	storeFinder internal.StoreFinder
}

func NewCreditAndInstallmentInStore(storeFinder internal.StoreFinder) *CreditAndInstallmentInStore {
	return &CreditAndInstallmentInStore{storeFinder: storeFinder}
}

func (c *CreditAndInstallmentInStore) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	deliveryId := ordr.Delivery().Id()
	if deliveryId != order.DeliveryIdSelf && deliveryId != order.DeliveryIdTerminal {
		return nil
	}

	selfDelivery, err := ordr.Delivery().Self(ctx)
	if err != nil {
		return fmt.Errorf("error getting self delivery: %w", err)
	}

	pupId := selfDelivery.PupId()

	if pupId == "" {
		return nil
	}

	store := c.storeFinder.FindByPupId(pupId)
	if store == nil {
		return fmt.Errorf("can't find store with pup_id: %s", string(pupId))
	}

	if !store.GetIsCreditAvailable() {
		resolvedIdsMap[order.PaymentIdCredit].Disallow(
			order.AllowStatusLimited,
			order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCreditAndInstallmentInStoreNotAvailable, order.SubsystemBasket).
				WithMessage("Оплата кредитом на данной точке выдачи запрещена!"),
		)
		resolvedIdsMap[order.PaymentIdInstallments].Disallow(
			order.AllowStatusLimited,
			order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCreditAndInstallmentInStoreNotAvailable, order.SubsystemBasket).
				WithMessage("Оплата рассрочкой на данной точке выдачи запрещена!"),
		)
	}

	return nil
}
