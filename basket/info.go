package basket

import (
	"go.citilink.cloud/order/internal/order/basket/basket_item"
)

func NewInfo(item *basket_item.Item, info *basket_item.Info) *Info {
	return &Info{item: item, info: info}
}

type Info struct {
	item *basket_item.Item // 1
	info *basket_item.Info // 2
}

func (i *Info) Item() *basket_item.Item {
	return i.item
}

func (i *Info) Info() *basket_item.Info {
	return i.info
}
