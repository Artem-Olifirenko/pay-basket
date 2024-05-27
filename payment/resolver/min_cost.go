package resolver

import (
	"context"
	"fmt"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewMinCost() *MinCost {
	return &MinCost{}
}

type MinCost struct {
}

func (r *MinCost) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	for _, resolvedId := range resolvedIdsMap {
		cost, err := ordr.Cost(ctx)
		if err != nil {
			return fmt.Errorf("error getting order cost: %w", err)
		}

		if resolvedId.Id().MinCost() != 0 && cost.GetWithDiscount() < resolvedId.Id().MinCost() {
			resolvedId.Disallow(
				order.AllowStatusLimited,
				order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdMinCost, order.SubsystemBasket),
			)
		}
	}
	return nil
}
