package resolver

import (
	"context"
	"fmt"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

type CreditAmountLimiter struct {
	minimumCreditAmount int
}

func NewCreditAmountLimiter(minimumCreditAmount int) *CreditAmountLimiter {
	return &CreditAmountLimiter{
		minimumCreditAmount: minimumCreditAmount,
	}
}

// Resolve блокирует способ оплаты в кредит, если минимальная сумма заказа меньше чем указана в featureCfg.CartLimitConfig.MinimumCreditAmount
func (limiter *CreditAmountLimiter) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, o order.Order) error {
	cost, err := o.Cost(ctx)
	if err != nil {
		return fmt.Errorf("error can't get order cost: %w", err)
	}

	if cost.GetWithDiscount() < limiter.minimumCreditAmount {
		resolvedIdsMap[order.PaymentIdCredit].Disallow(
			order.AllowStatusDisallowed,
			order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdItemNotAvailableForPaymentMethod, order.SubsystemBasket),
		)
	}

	return nil
}
