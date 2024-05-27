package basket

import (
	"go.citilink.cloud/order/internal/order/basket/basket_item"
)

type Finder func([]*basket_item.Item) basket_item.Items

var Finders = &finders{}

type finders struct{}

func (f *finders) ChildrenOf(parent *basket_item.Item) Finder {
	return func(items []*basket_item.Item) basket_item.Items {
		if !parent.Spec().CanHaveChildren() {
			return nil
		}

		var foundedItems []*basket_item.Item
		for _, item := range items {
			if item.IsChildOf(parent) {
				foundedItems = append(foundedItems, item)
			}
		}

		return foundedItems
	}
}

func (f *finders) ChildrenOfRecursive(parent *basket_item.Item) Finder {
	return func(items []*basket_item.Item) basket_item.Items {
		return findChildrenOfRecursive(parent, items)
	}
}

func findChildrenOfRecursive(parent *basket_item.Item, items []*basket_item.Item) []*basket_item.Item {
	if !parent.Spec().CanHaveChildren() {
		return nil
	}

	var directChildren []*basket_item.Item
	for _, item := range items {
		if item.IsChildOf(parent) {
			directChildren = append(directChildren, item)
		}
	}

	foundedItems := make([]*basket_item.Item, 0, len(directChildren))
	foundedItems = append(foundedItems, directChildren...)
	for _, item := range directChildren {
		foundedItems = append(foundedItems, findChildrenOfRecursive(item, items)...)
	}

	return foundedItems
}

func (f *finders) ByType(itemTypes ...basket_item.Type) Finder {
	return func(items []*basket_item.Item) basket_item.Items {
		var foundedItems []*basket_item.Item
		for _, item := range items {
			for _, itemType := range itemTypes {
				if item.Type() == itemType {
					foundedItems = append(foundedItems, item)
				}
			}
		}

		return foundedItems
	}
}

func (f *finders) ByItemIds(itemIds ...basket_item.ItemId) Finder {
	return func(items []*basket_item.Item) basket_item.Items {
		var foundedItems []*basket_item.Item
		for _, item := range items {
			founded := false
			for _, itemId := range itemIds {
				if item.ItemId() == itemId {
					founded = true
					break
				}
			}
			if founded {
				foundedItems = append(foundedItems, item)
			}
		}

		return foundedItems
	}
}

// FilterForXItems специальный метод, который отсекает ненужные позиции, оставляя только те, которые можно потом
// превратить в XItems и отправить в специальные процедуры. Это сделано в связи с тем, чтобы были добавлены новые
// услуги в корзину и некоторые процедуры не могут их переваривать правильно, из-за этого приходится отфильтровывать
// такие позиции, чтобы все работало как раньше
func (f *finders) FilterForXItems() Finder {
	return func(items []*basket_item.Item) basket_item.Items {
		var foundedItems []*basket_item.Item
		for _, item := range items {
			if item.Type() == basket_item.TypeDeliveryService || item.Type() == basket_item.TypeLiftingService {
				continue
			}
			foundedItems = append(foundedItems, item)
		}

		return foundedItems
	}
}

// FilterForSubcontractServices находит все услуги установки
func (f *finders) FilterForSubcontractServices() Finder {
	return func(items []*basket_item.Item) basket_item.Items {
		var foundedItems []*basket_item.Item
		for _, item := range items {
			if item.Type() != basket_item.TypeSubcontractServiceForProduct {
				continue
			}
			foundedItems = append(foundedItems, item)
		}

		return foundedItems
	}
}

func (f *finders) FindUserAddedItems() Finder {
	return func(items []*basket_item.Item) basket_item.Items {
		var foundedItems []*basket_item.Item
		for _, item := range items {
			if !item.Type().IsAddedByServer() {
				foundedItems = append(foundedItems, item)
			}
		}

		return foundedItems
	}
}
