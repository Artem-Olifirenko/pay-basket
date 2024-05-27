package resolver

//go:generate mockgen -source=online_payment_check.go -destination=./online_payment_check_mock.go -package=resolver

import (
	"context"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
	"go.citilink.cloud/order/internal/order/payment/resolver/mssql"
	"go.citilink.cloud/store_types"
)

// OnlinePaymentInfo позволяет получить различную информацию по онлайн платёжным инструментам
type OnlinePaymentInfo interface {
	// GetRegions получает список всех регионов системы и информацию о доступности онлайн оплаты в этих регионах
	GetRegions(ctx context.Context) (map[store_types.SpaceId]*mssql.YandexPayShopData, error)
}

type OnlinePaymentCheck struct {
	onlinePaymentInfo OnlinePaymentInfo
}

// OnlinePaymentIds слайс доступных методов оплат онлайн
var onlinePaymentIds = []order.PaymentId{
	order.PaymentIdWebmoney,
	order.PaymentIdCardsOnline,
	order.PaymentIdYandex,
	order.PaymentIdTerminalOrCashbox,
	order.PaymentIdSbp,
	order.PaymentIdApplePay,
}

func NewOnlinePaymentCheck(onlinePaymentInfo OnlinePaymentInfo) *OnlinePaymentCheck {
	return &OnlinePaymentCheck{onlinePaymentInfo: onlinePaymentInfo}
}

// Resolve блокирует онлайн методы оплат, если регион заказа не поддерживает онлайн оплату
func (r *OnlinePaymentCheck) Resolve(
	ctx context.Context,
	resolvedIdsMap payment.ResolvedPaymentIdMap,
	ordr order.Order,
) error {
	regions, err := r.onlinePaymentInfo.GetRegions(ctx)

	if err != nil {
		resolvedIdsMap.DisallowMany(
			onlinePaymentIds,
			order.AllowStatusDisallowed,
			order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdOnlinePaymentInRegion, order.SubsystemUnknown))
		return err
	}

	if regionInfo, ok := regions[ordr.SpaceId()]; ok {
		if !regionInfo.IsEnabled {
			resolvedIdsMap.DisallowMany(
				onlinePaymentIds,
				order.AllowStatusDisallowed,
				order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdOnlinePaymentInRegion, order.SubsystemUnknown))
		}

		return nil
	}

	// блокируем, если в БД нет информации о регионе
	resolvedIdsMap.DisallowMany(
		onlinePaymentIds,
		order.AllowStatusDisallowed,
		order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdOnlinePaymentInRegion, order.SubsystemUnknown))

	return nil
}
