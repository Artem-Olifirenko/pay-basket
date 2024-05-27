package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewAnonUser() *AnonUser {
	return &AnonUser{}
}

type AnonUser struct {
}

// Resolve резолвер блокирует типы оплаты, недоступные для анонимного пользователя
func (r *AnonUser) Resolve(_ context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	if ordr.User() != nil {
		return nil
	}

	for _, resolvedId := range resolvedIdsMap {
		if !resolvedId.Id().AvailableForAnonUser() {
			resolvedId.Disallow(
				order.AllowStatusLimited,
				order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdAnonUser, order.SubsystemUnknown),
			)
		}
	}
	return nil
}
