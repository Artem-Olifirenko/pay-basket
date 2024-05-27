package basket_item

import (
	"go.citilink.cloud/catalog_types"
)

const ConfSpecialItemId = ItemId("553081")
const DefaultConfId = ConfId("q0000")

type ConfType int

const (
	ConfTypeUnknown  ConfType = 0
	ConfTypeUser     ConfType = 1
	ConfTypeTemplate ConfType = 2
	ConfTypeVendor   ConfType = 3
)

type ConfId string

type ConfItem struct {
	ProductId catalog_types.ProductId
	Count     int
	Services  []*ConfItemService
}

type ConfItemService struct {
	ItemId string
	Name   string
	Price  int
	Count  int
}
