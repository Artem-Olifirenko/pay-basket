package resolver

//go:generate mockgen -source=preorder.go -destination=./preorder_mock.go -package=resolver

import (
	"context"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/payment"
	"go.citilink.cloud/order/internal/preorder"
)

// PreOrderList позволяет получить информацию по предзаказам
type PreOrderList interface {
	// FilterByProductIds получает список предзаказов, которые есть в бд из переданных
	FilterByProductIds(ctx context.Context, productIds ...catalog_types.ProductId) (preorder.PreOrders, error)
}

type PreOrder struct {
	preOrders PreOrderList
}

func NewPreOrder(preOrders PreOrderList) *PreOrder {
	return &PreOrder{preOrders: preOrders}
}

// Resolve ограничивает тип оплаты, если имеется в корзине предзаказанный товар
func (r *PreOrder) Resolve(ctx context.Context, resolvedIdsMap payment.ResolvedPaymentIdMap, ordr order.Order) error {
	bsk, err := ordr.Basket(ctx)
	if err != nil {
		return err
	}

	var productIds []catalog_types.ProductId
	items := basket_item.Items{}
	for _, item := range bsk.All() {
		if item.Type().IsProduct() {
			productIds = append(productIds, catalog_types.ProductId(item.ItemId()))
			items = append(items, item)
		}
	}

	preOrders, err := r.preOrders.FilterByProductIds(ctx, productIds...)
	if err != nil {
		return err
	}

	if len(preOrders) == 0 {
		return nil
	}

	availableIdMap := map[order.PaymentId]bool{}
	for _, paymentId := range preOrders.PaymentIds() {
		switch paymentId {
		case order.PaymentIdCashless:
			availableIdMap[order.PaymentIdCashless] = true
		case order.PaymentIdSberbankBusinessOnline:
			availableIdMap[order.PaymentIdSberbankBusinessOnline] = true
		case order.PaymentIdCash:
			availableIdMap[order.PaymentIdCash] = true
		case order.PaymentIdCashWithCard:
			availableIdMap[order.PaymentIdCashWithCard] = true
		case order.PaymentIdCardsOnline:
			availableIdMap[order.PaymentIdCardsOnline] = true
		case order.PaymentIdCredit:
			availableIdMap[order.PaymentIdCardsOnline] = true
		case order.PaymentIdInstallments:
			availableIdMap[order.PaymentIdCardsOnline] = true
		case order.PaymentIdYandex:
			availableIdMap[order.PaymentIdYandex] = true
		case order.PaymentIdTerminalOrCashbox:
			availableIdMap[order.PaymentIdTerminalOrCashbox] = true
		case order.PaymentIdWebmoney:
			availableIdMap[order.PaymentIdWebmoney] = true
		case order.PaymentIdTerminalPinpad:
			availableIdMap[order.PaymentIdTerminalPinpad] = true
		case order.PaymentIdSbp:
			availableIdMap[order.PaymentIdSbp] = true
		case order.PaymentIdApplePay:
			availableIdMap[order.PaymentIdApplePay] = true
		case order.PaymentIdCards:
			availableIdMap[order.PaymentIdCards] = true
		case order.PaymentIdSberPay:
			availableIdMap[order.PaymentIdSberPay] = true
		}
	}

	for _, id := range resolvedIdsMap {
		if _, ok := availableIdMap[id.Id()]; ok {
			continue
		}

		id.Disallow(
			order.AllowStatusLimited,
			order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdPreOrder, order.SubsystemBasket).
				WithMessage("В связи с присутствием в корзине предзаказного товара ограничен тип оплаты").
				WithAdditions(order.DisallowReasonWithInfoAdditions{
					Basket: order.NewBasketDisallowAdditions(items.UniqIds()),
				}))
	}

	return nil
}
