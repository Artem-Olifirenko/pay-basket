package resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

func NewFastCitilinkCourierDelivery() *FastCitilinkCourierDelivery {
	return &FastCitilinkCourierDelivery{}
}

// FastCitilinkCourierDelivery Резолвер курьерской доставки день в день
type FastCitilinkCourierDelivery struct {
}

// Resolve сверяет заказы по айди и меняет статус в случае несоответствия
func (r *FastCitilinkCourierDelivery) Resolve(
	ctx context.Context,
	resolvedIdsMap payment.ResolvedPaymentIdMap,
	ordr order.Order,
) error {
	if ordr.Delivery().Id() != order.DeliveryIdCitilinkCourier || ordr.Delivery().CitilinkCourier(ctx).Id() != order.CitilinkCourierDeliveryIdFast {
		return nil
	}

	for _, resolvedId := range resolvedIdsMap {
		switch resolvedId.Id() {
		case order.PaymentIdInstallments:
			{
				resolvedId.Disallow(order.AllowStatusLimited,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdFastCitilinkCourierDelivery, order.SubsystemDelivery).
						WithMessage("Рассрочка недоступна при срочной доставке."))
			}
		case order.PaymentIdCredit:
			{
				resolvedId.Disallow(order.AllowStatusLimited,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdFastCitilinkCourierDelivery, order.SubsystemDelivery).
						WithMessage("Кредит недоступен при срочной доставке."))
			}
		case order.PaymentIdCardsOnline:
			{
				resolvedId.Disallow(order.AllowStatusLimited,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdFastCitilinkCourierDelivery, order.SubsystemDelivery).
						WithMessage("Данный способ оплаты недоступен при срочной доставке."))
			}
		case order.PaymentIdSbp:
			{
				resolvedId.Disallow(order.AllowStatusLimited,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdFastCitilinkCourierDelivery, order.SubsystemDelivery).
						WithMessage("Данный способ оплаты недоступен при срочной доставке."))
			}
		case order.PaymentIdYandex:
			{
				resolvedId.Disallow(order.AllowStatusLimited,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdFastCitilinkCourierDelivery, order.SubsystemDelivery).
						WithMessage("Данный способ оплаты недоступен при срочной доставке."))
			}
		}
	}

	return nil
}
