package resolver

import (
	"context"
	"fmt"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewMaxCost() *MaxCost {
	return &MaxCost{}
}

type MaxCost struct {
}

func (r *MaxCost) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	cost, err := ordr.Cost(ctx)
	if err != nil {
		return fmt.Errorf("error getting order cost: %w", err)
	}

	for _, resolvedId := range resolvedIdsMap {
		if resolvedId.Id().MaxCost() != 0 && cost.GetWithDiscount() > resolvedId.Id().MaxCost() {
			resolvedId.Disallow(
				order.AllowStatusLimited,
				order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdMaxCost, order.SubsystemBasket),
			)
		}
	}
	return nil
}
