package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewUser() *User {
	return &User{}
}

type User struct {
}

// Resolve резолвер блокирует типы оплаты, недоступные для авторизованного пользователя (физика)
func (r *User) Resolve(_ context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	if ordr.User() == nil || ordr.User().GetB2B().GetIsB2BState() {
		return nil
	}

	for _, resolvedId := range resolvedIdsMap {
		if !resolvedId.Id().AvailableForUser() {
			resolvedId.Disallow(
				order.AllowStatusDisallowed,
				order.NewDisallowReasonWithInfo(order.DisallowReasonUnknown, order.SubsystemPaymentMethod),
			)
		}
	}
	return nil
}
