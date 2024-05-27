package basket

//go:generate mockgen -source=actualizer.go -destination=actualizer_mock.go -package=basket

import (
	"go.citilink.cloud/order/internal/order/basket/basket_item"
)

type ActualizerItem interface {
	GetName() string
	GetPrice() int
	GetCount() int
	GetItemId() basket_item.ItemId
	GetNotExist() bool
	ReduceInfo() ReduceInfo
	GetParentItemId() basket_item.ItemId
}

type ActualizerItems interface {
	FindByItem(item *basket_item.Item) ActualizerItem
	FindByType(itemType basket_item.Type) []ActualizerItem
}

type ReduceInfo interface {
	Info() string
	Count() int
}
