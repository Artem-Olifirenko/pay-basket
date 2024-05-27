package basket

//go:generate mockgen -source=refresher_composite.go -destination=refresher_composite_mock.go -package=basket

import (
	"context"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.uber.org/zap"
)

func NewItemRefresherComposite(refreshers ...refresher) *ItemRefresherComposite {
	return &ItemRefresherComposite{refreshers}
}

// refresher обновитель цен/наличия и т.п. у позиций
type refresher interface {
	Refreshable(item *basket_item.Item) bool
	Refresh(ctx context.Context, items []*basket_item.Item, bsk RefresherBasket, logger *zap.Logger) error
}

// ItemRefresherComposite композитный обновитель позиций. Применяется для обновления цен, наличия и т.п. у позиций
type ItemRefresherComposite struct {
	refreshers []refresher
}

func (r *ItemRefresherComposite) Refresh(ctx context.Context, basket RefresherBasket, logger *zap.Logger) error {
	errsCollection := internal.NewErrorCollection()
	for _, refresher := range r.refreshers {
		var refreshableItems basket_item.Items
		// Отфильтровываем позиции корзины в соответствии с типом, обрабатываемым рефрешером
		for _, item := range basket.All() {
			if !refresher.Refreshable(item) {
				continue
			}
			refreshableItems = append(refreshableItems, item)
		}

		if len(refreshableItems) > 0 {
			err := refresher.Refresh(ctx, refreshableItems, basket, logger)
			if err != nil {
				errsCollection.Add(err)
			}
		}
	}
	if !errsCollection.Empty() {
		return errsCollection
	}

	return nil
}
