package resolver

import (
	"context"
	"fmt"
	"go.citilink.cloud/order/internal"

	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

type CashWithCardCheck struct {
	withoutCardCityIdsStorage internal.StringsContainer
}

// NewCashWithCardCheck принимает на вход список городов citiesWithoutCardList, где не принимают оплату картой
func NewCashWithCardCheck(citiesWithoutCardListStorage internal.StringsContainer) *CashWithCardCheck {
	return &CashWithCardCheck{withoutCardCityIdsStorage: citiesWithoutCardListStorage}
}

func (r *CashWithCardCheck) isOrderCityWithoutCardOrDeliveryCourier(ordr order.Order) bool {
	return r.withoutCardCityIdsStorage.Contains(string(ordr.CityId())) && ordr.Delivery().Id() == order.DeliveryIdCitilinkCourier
}

// Resolve резолвер блокирует способ оплаты "оплата наличными" если выбран город не из списка ИЛИ доставка самовывоз,
// блокирует способ оплаты "оплата наличными или картой" если выбран город из списка и доставка не самовывоз
func (r *CashWithCardCheck) Resolve(
	ctx context.Context,
	resolvedIdsMap payment.ResolvedPaymentIdMap,
	ordr order.Order,
) error {
	// не запускаем данные проверки для B2B
	if ordr.User() != nil && ordr.User().GetB2B().GetIsB2BState() {
		return nil
	}

	// Для магазинов Сберлогистики и DPD флаги способов оплаты (картой и наличкой) надо брать из
	// БД (см. #WEB-48004 и #WEB-51999). Они туда приходят из Navision. По нашим магазинам данные приходят
	// из другой таблицы - и информации по оплате картой или только наличными нет
	if ordr.Delivery().Id() == order.DeliveryIdSelf {
		selfDelivery, err := ordr.Delivery().Self(ctx)
		if err != nil {
			return fmt.Errorf("can't get self delivery: %w", err)
		}

		chosenStore, err := selfDelivery.ChosenStore(ctx)
		if err != nil {
			return fmt.Errorf("can't get chosen store: %w", err)
		}

		if chosenStore != nil {
			// В магазине есть и оплата наличкой, и оплата картой - показываем "оплата картой или наличкой"
			if chosenStore.Store().IsCashPaymentAvailable && chosenStore.Store().IsPaymentByCardAvailable {
				resolvedIdsMap[order.PaymentIdCash].Disallow(
					order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCashNotAvailable, order.SubsystemUnknown),
				)
				resolvedIdsMap[order.PaymentIdCards].Disallow(
					order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCardNotAvailable, order.SubsystemUnknown),
				)
				return nil
			}

			// В магазине нет оплаты наличкой (только карта)
			if !chosenStore.Store().IsCashPaymentAvailable {
				resolvedIdsMap[order.PaymentIdCash].Disallow(
					order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCashNotAvailable, order.SubsystemUnknown),
				)
			}

			// В магазине нет оплаты картой (только наличка)
			if !chosenStore.Store().IsPaymentByCardAvailable {
				resolvedIdsMap[order.PaymentIdCards].Disallow(
					order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCardNotAvailable, order.SubsystemUnknown),
				)
			}

			resolvedIdsMap[order.PaymentIdCashWithCard].Disallow(
				order.AllowStatusDisallowed,
				order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCashWithCardNotAvailable, order.SubsystemUnknown),
			)
			return nil
		}
	}

	if r.isOrderCityWithoutCardOrDeliveryCourier(ordr) {
		// Если выбран город из списка и доставка не самовывоз - выводится "оплата наличными"
		resolvedIdsMap[order.PaymentIdCards].Disallow(
			order.AllowStatusDisallowed,
			order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCardNotAvailable, order.SubsystemUnknown),
		)
		resolvedIdsMap[order.PaymentIdCashWithCard].Disallow(
			order.AllowStatusDisallowed,
			order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCashWithCardNotAvailable, order.SubsystemUnknown),
		)
		return nil
	}

	// Во всех остальных случаях выводится "оплата наличными или картой"
	resolvedIdsMap[order.PaymentIdCards].Disallow(
		order.AllowStatusDisallowed,
		order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCardNotAvailable, order.SubsystemUnknown),
	)
	resolvedIdsMap[order.PaymentIdCash].Disallow(
		order.AllowStatusDisallowed,
		order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCashNotAvailable, order.SubsystemUnknown),
	)

	return nil
}
