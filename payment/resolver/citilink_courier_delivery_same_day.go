package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewSameDayCitilinkCourierDelivery() *SameDayCitilinkCourierDelivery {
	return &SameDayCitilinkCourierDelivery{}
}

type SameDayCitilinkCourierDelivery struct {
}

// Resolve резолвер блокирует типы оплаты = кредит и рассрочка, если выбран способ доставки = день-в-день
func (r *SameDayCitilinkCourierDelivery) Resolve(
	ctx context.Context,
	resolvedIdsMap payment.ResolvedPaymentIdMap,
	ordr order.Order,
) error {
	// если способ доставки отличный от доставки курьером день-в-день - ничего блокировать не нужно
	if ordr.Delivery().Id() != order.DeliveryIdCitilinkCourier || ordr.Delivery().CitilinkCourier(ctx).Id() != order.CitilinkCourierDeliveryIdSameDay {
		return nil
	}

	for _, resolvedId := range resolvedIdsMap {
		if resolvedId.Id() == order.PaymentIdInstallments {
			resolvedId.Disallow(order.AllowStatusLimited,
				order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdSameDayCitilinkCourierDelivery, order.SubsystemDelivery).
					WithMessage("Рассрочка недоступна при доставке день-в-день."))
		}
		if resolvedId.Id() == order.PaymentIdCredit {
			resolvedId.Disallow(order.AllowStatusLimited,
				order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdSameDayCitilinkCourierDelivery, order.SubsystemDelivery).
					WithMessage("Кредит недоступен при доставке день-в-день."))
		}
	}

	return nil
}
