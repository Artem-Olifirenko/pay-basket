package resolver

import (
	"context"
	"fmt"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

type fnsTrackedProduct struct {
	isFnsTrackingEnabled bool // включена ли опция прослеживаемости товаров
}

func NewFnsTrackedProduct(fnsTrackedEnabled bool) *fnsTrackedProduct {
	return &fnsTrackedProduct{isFnsTrackingEnabled: fnsTrackedEnabled}
}

// Resolve ограничивает способы оплаты для прослеживаемых товаров
func (f *fnsTrackedProduct) Resolve(
	ctx context.Context,
	resolvedIdsMap payment.ResolvedPaymentIdMap,
	ordr order.Order,
) error {
	if f.isFnsTrackingEnabled {
		// блокируем только для b2b юзеров
		if ordr.User() == nil || ordr.User().GetB2B() == nil || !ordr.User().GetB2B().GetIsB2BState() {
			return nil
		}

		bsk, err := ordr.Basket(ctx)
		if err != nil {
			return fmt.Errorf("can't get basket of order: %w", err)
		}

		// нет прослеживаемого товара - выходим
		if !bsk.HasFnsTrackedProducts() {
			return nil
		}

		allowedIdsMap := map[order.PaymentId]bool{
			order.PaymentIdCashless: true,
		}

		for _, paymentType := range resolvedIdsMap {
			if _, ok := allowedIdsMap[paymentType.Id()]; !ok {
				paymentType.Disallow(
					order.AllowStatusLimited,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentForFnsTrackedProducts, order.SubsystemPaymentMethod).
						WithMessage("Данный способ оплаты недоступен для прослеживаемых товаров."),
				)
			}
		}
	}
	return nil
}
