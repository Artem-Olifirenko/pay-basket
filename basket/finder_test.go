package basket

import (
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/store_types"
	"testing"
)

func Test_finders_ChildrenOf(t *testing.T) {
	tests := []struct {
		name string
		want func() (parent *basket_item.Item, itemsList basket_item.Items, result basket_item.Items)
	}{
		{
			"is child",
			func() (*basket_item.Item, basket_item.Items, basket_item.Items) {
				parent := basket_item.NewItem(
					"test_parent_item_id",
					basket_item.TypeConfiguration,
					"test parent product name",
					"test parent image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypeConfigurationProduct,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				err := item.MakeChildOf(parent)
				assert.NoError(t, err)

				items := basket_item.Items{
					item,
				}

				wanted := basket_item.Items{
					item,
				}

				return parent, items, wanted
			},
		},
		{
			"is not child",
			func() (*basket_item.Item, basket_item.Items, basket_item.Items) {
				parent := basket_item.NewItem(
					"test_parent_item_id",
					basket_item.TypeConfiguration,
					"test parent product name",
					"test parent image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypeConfigurationProduct,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				items := basket_item.Items{
					item,
				}

				return parent, items, nil
			},
		},
		{
			"parent can't have children",
			func() (*basket_item.Item, basket_item.Items, basket_item.Items) {
				parent := basket_item.NewItem(
					"test_parent_item_id",
					basket_item.TypeInsuranceServiceForProduct,
					"test parent product name",
					"test parent image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				return parent, basket_item.Items{}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &finders{}
			parent, items, want := tt.want()
			childrenFinder := f.ChildrenOf(parent)
			got := childrenFinder(items)

			assert.Equal(t, want, got)
		})
	}
}

func Test_finders_ChildrenOfRecursive(t *testing.T) {
	tests := []struct {
		name string
		want func() (parent *basket_item.Item, itemsList basket_item.Items, result basket_item.Items)
	}{
		{
			"parent have direct child",
			func() (*basket_item.Item, basket_item.Items, basket_item.Items) {
				parent := basket_item.NewItem(
					"test_parent_item_id",
					basket_item.TypeConfiguration,
					"test parent product name",
					"test parent image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypeConfigurationProduct,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				err := item.MakeChildOf(parent)
				assert.NoError(t, err)

				items := basket_item.Items{
					item,
				}

				wanted := basket_item.Items{
					item,
				}

				return parent, items, wanted
			},
		},
		{
			"parent have nested child",
			func() (*basket_item.Item, basket_item.Items, basket_item.Items) {
				parent := basket_item.NewItem(
					"test_parent_item_id",
					basket_item.TypeConfiguration,
					"test parent product name",
					"test parent image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypeConfigurationProduct,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				service := basket_item.NewItem(
					"test_service_id",
					basket_item.TypeConfigurationProductService,
					"test service",
					"test image",
					1,
					2000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				err := item.MakeChildOf(parent)
				assert.NoError(t, err)

				err = service.MakeChildOf(item)
				assert.NoError(t, err)

				items := basket_item.Items{
					item,
					service,
				}

				wanted := basket_item.Items{
					item,
					service,
				}

				return parent, items, wanted
			},
		},
		{
			"parent can't have children",
			func() (*basket_item.Item, basket_item.Items, basket_item.Items) {
				parent := basket_item.NewItem(
					"test_parent_item_id",
					basket_item.TypeInsuranceServiceForProduct,
					"test parent product name",
					"test parent image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				items := basket_item.Items{}

				return parent, items, nil
			},
		},
		{
			"empty items list",
			func() (*basket_item.Item, basket_item.Items, basket_item.Items) {
				parent := basket_item.NewItem(
					"test_parent_item_id",
					basket_item.TypeConfiguration,
					"test parent product name",
					"test parent image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				items := basket_item.Items{}

				return parent, items, basket_item.Items{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &finders{}
			parent, items, want := tt.want()
			childrenFinder := f.ChildrenOfRecursive(parent)
			got := childrenFinder(items)

			assert.Equal(t, want, got)
		})
	}
}

func Test_finders_ByType(t *testing.T) {
	tests := []struct {
		name string
		want func() (types []basket_item.Type, itemsList basket_item.Items, result basket_item.Items)
	}{
		{
			"empty items list",
			func() ([]basket_item.Type, basket_item.Items, basket_item.Items) {
				types := []basket_item.Type{}
				items := basket_item.Items{}
				return types, items, nil
			},
		},
		{
			"empty types list",
			func() ([]basket_item.Type, basket_item.Items, basket_item.Items) {
				types := []basket_item.Type{}
				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypeConfigurationProduct,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)
				items := basket_item.Items{
					item,
				}
				return types, items, nil
			},
		},
		{
			"found by type",
			func() ([]basket_item.Type, basket_item.Items, basket_item.Items) {
				types := []basket_item.Type{basket_item.TypeProduct}
				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypeProduct,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)
				items := basket_item.Items{
					item,
				}
				wanted := basket_item.Items{
					item,
				}
				return types, items, wanted
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &finders{}
			types, items, want := tt.want()
			finder := f.ByType(types...)
			got := finder(items)

			assert.Equal(t, want, got)
		})
	}
}

func Test_finders_ByItemIds(t *testing.T) {
	tests := []struct {
		name string
		want func() (ids []basket_item.ItemId, itemsList basket_item.Items, result basket_item.Items)
	}{
		{
			"found by ids",
			func() ([]basket_item.ItemId, basket_item.Items, basket_item.Items) {
				ids := []basket_item.ItemId{"test_item_id"}
				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypeProduct,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)
				items := basket_item.Items{
					item,
				}
				wanted := basket_item.Items{
					item,
				}
				return ids, items, wanted
			},
		},
		{
			"not found",
			func() ([]basket_item.ItemId, basket_item.Items, basket_item.Items) {
				ids := []basket_item.ItemId{"wrong_test_item_id"}
				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypeProduct,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)
				items := basket_item.Items{
					item,
				}
				return ids, items, nil
			},
		},
		{
			"empty item ids",
			func() ([]basket_item.ItemId, basket_item.Items, basket_item.Items) {
				ids := []basket_item.ItemId{}
				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypeProduct,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)
				items := basket_item.Items{
					item,
				}
				return ids, items, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &finders{}
			itemIds, items, want := tt.want()
			finder := f.ByItemIds(itemIds...)
			got := finder(items)

			assert.Equal(t, want, got)
		})
	}
}

func Test_finders_FilterForXItems(t *testing.T) {
	tests := []struct {
		name string
		want func() (itemsList basket_item.Items, result basket_item.Items)
	}{
		{
			"wrong types",
			func() (itemsList basket_item.Items, result basket_item.Items) {
				items := basket_item.Items{
					basket_item.NewItem(
						"test_item_id",
						basket_item.TypeDeliveryService,
						"test product name",
						"test image",
						1,
						1000,
						0,
						store_types.SpaceId("test_space_id"),
						catalog_types.PriceColumnClub,
					),
					basket_item.NewItem(
						"test_item_id",
						basket_item.TypeLiftingService,
						"test product name",
						"test image",
						1,
						1000,
						0,
						store_types.SpaceId("test_space_id"),
						catalog_types.PriceColumnClub,
					),
				}
				return items, nil
			},
		},
		{
			"filtered",
			func() (itemsList basket_item.Items, result basket_item.Items) {
				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypeProduct,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)
				items := basket_item.Items{
					item,
				}
				wanted := basket_item.Items{
					item,
				}
				return items, wanted
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &finders{}
			items, want := tt.want()
			finder := f.FilterForXItems()
			got := finder(items)

			assert.Equal(t, want, got)
		})
	}
}

func Test_finders_FilterForSubcontractServices(t *testing.T) {
	tests := []struct {
		name string
		want func() (itemsList basket_item.Items, result basket_item.Items)
	}{
		{
			"wrong types",
			func() (itemsList basket_item.Items, result basket_item.Items) {
				items := basket_item.Items{
					basket_item.NewItem(
						"test_item_id",
						basket_item.TypeDeliveryService,
						"test product name",
						"test image",
						1,
						1000,
						0,
						store_types.SpaceId("test_space_id"),
						catalog_types.PriceColumnClub,
					),
					basket_item.NewItem(
						"test_item_id",
						basket_item.TypeLiftingService,
						"test product name",
						"test image",
						1,
						1000,
						0,
						store_types.SpaceId("test_space_id"),
						catalog_types.PriceColumnClub,
					),
				}
				return items, nil
			},
		},
		{
			"filtered",
			func() (itemsList basket_item.Items, result basket_item.Items) {
				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypeSubcontractServiceForProduct,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)
				items := basket_item.Items{
					item,
				}
				wanted := basket_item.Items{
					item,
				}
				return items, wanted
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &finders{}
			items, want := tt.want()
			finder := f.FilterForSubcontractServices()
			got := finder(items)

			assert.Equal(t, want, got)
		})
	}
}

func Test_finders_FindUserAddedItems(t *testing.T) {
	tests := []struct {
		name string
		want func() (itemsList basket_item.Items, result basket_item.Items)
	}{
		{
			"not found",
			func() (itemsList basket_item.Items, result basket_item.Items) {
				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypePresent,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)
				items := basket_item.Items{item}
				return items, nil
			},
		},
		{
			"found",
			func() (itemsList basket_item.Items, result basket_item.Items) {
				item := basket_item.NewItem(
					"test_item_id",
					basket_item.TypeProduct,
					"test product name",
					"test image",
					1,
					1000,
					0,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)
				items := basket_item.Items{item}
				wanted := basket_item.Items{item}
				return items, wanted
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &finders{}
			items, want := tt.want()
			finder := f.FindUserAddedItems()
			got := finder(items)

			assert.Equal(t, want, got)
		})
	}
}
