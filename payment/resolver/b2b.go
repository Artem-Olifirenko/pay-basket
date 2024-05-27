package resolver

import (
	"context"
	"fmt"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
)

// если сумма заказа менее 100 000 р - разрешена оплата и налом и картой
// если сумма заказа более 100 000 р - разрешена оплата только картой
const cashPaymentLimit = 100_000

func NewB2b(isCashOrCardAvailable bool) *B2b {
	return &B2b{
		isCashOrCardAvailable: isCashOrCardAvailable,
	}
}

type B2b struct {
	// доступность способов оплаты нал/безнал, главное условие, до всех проверок
	isCashOrCardAvailable bool
}

// Resolve резолвер блокирует типы оплаты, если:
// - пользователь корпоративный, то блокировать оплату, у которой нет признака B2b
func (r *B2b) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	cost, err := ordr.Cost(ctx)
	if err != nil {
		return fmt.Errorf("error getting order cost: %w", err)
	}

	// если пользователь НЕ анонимный и находится в режиме b2b
	if ordr.User() != nil && ordr.User().GetB2B().GetIsB2BState() {
		contractor, err := ordr.B2B().Contractor(ctx)
		if err != nil {
			return fmt.Errorf("error getting b2b contractor: %w", err)
		}

		contractorHasSanctions := false
		if contractor != nil {
			contractorHasSanctions = contractor.HasSanctions()
		}

		// перебираем доступные доступные методы оплаты
		for _, resolvedId := range resolvedIdsMap {
			// если метод недоступен для B2b то ввести запрет для метода оплаты
			if !resolvedId.Id().AvailableForB2b() {
				resolvedId.Disallow(
					order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdB2b, order.SubsystemUnknown),
				)
			} else {
				switch resolvedId.Id() {
				case order.PaymentIdCashWithCard, order.PaymentIdCards:
					if !r.isCashOrCardAvailable || contractorHasSanctions {
						resolvedId.Disallow(
							order.AllowStatusDisallowed,
							order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdB2b, order.SubsystemPaymentMethod),
						)
						continue
					}
				}

				if resolvedId.Id() == order.PaymentIdCashWithCard && cost.GetWithDiscount() > cashPaymentLimit {
					resolvedId.Disallow(
						order.AllowStatusDisallowed,
						order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdMaxCost, order.SubsystemPaymentMethod),
					)
				}

				if resolvedId.Id() == order.PaymentIdCards && cost.GetWithDiscount() <= cashPaymentLimit {
					resolvedId.Disallow(
						order.AllowStatusDisallowed,
						order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdMinCost, order.SubsystemPaymentMethod),
					)
				}
			}
		}
	} else if ordr.User() == nil {
		// если пользователь анонимный ИЛИ у пользователя нет доступа к B2b
		for _, resolvedId := range resolvedIdsMap {
			// если метод доступен для B2b то ввести запрет для методы оплаты
			if resolvedId.Id().AvailableForB2b() && !resolvedId.Id().AvailableForAnonUser() {
				resolvedId.Disallow(
					order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdB2b, order.SubsystemUnknown),
				)
			}
		}
	}

	return nil
}
