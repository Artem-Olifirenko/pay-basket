package resolver

import (
	"context"
	"fmt"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

type InstallmentPlanAmountLimiter struct {
	minimumInstallmentPlanAmount int
}

func NewInstallmentPlanAmountLimiter(minimumInstallmentPlanAmount int) *InstallmentPlanAmountLimiter {
	return &InstallmentPlanAmountLimiter{
		minimumInstallmentPlanAmount: minimumInstallmentPlanAmount,
	}
}

// Resolve блокирует способ оплаты в кредит, если минимальная сумма заказа меньше чем указана в featureCfg.CartLimitConfig.MinimumInstallmentPlanAmount
func (limiter *InstallmentPlanAmountLimiter) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, o order.Order) error {
	cost, err := o.Cost(ctx)
	if err != nil {
		return fmt.Errorf("can't get order cost: %w", err)
	}

	if cost.GetWithDiscount() < limiter.minimumInstallmentPlanAmount {
		resolvedIdsMap[order.PaymentIdInstallments].Disallow(
			order.AllowStatusDisallowed,
			order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdItemNotAvailableForPaymentMethod, order.SubsystemBasket),
		)
	}

	return nil
}
