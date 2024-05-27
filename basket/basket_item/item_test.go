package basket_item

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/catalog_types"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
	"go.citilink.cloud/store_types"
	"testing"
)

func generateItem(id ItemId, iType Type) *Item {
	return NewItem(
		id,
		iType,
		"item_"+string(iType),
		"image",
		0,
		0,
		0,
		store_types.NewSpaceId("msk_cl"),
		catalog_types.PriceColumnRetail,
	)
}

func TestItemId_IsConfiguration(t *testing.T) {
	tests := []struct {
		name string
		item ItemId
		want bool
	}{
		{
			"is configuration",
			ConfSpecialItemId,
			true,
		},
		{
			"is not configuration",
			ItemId("test"),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.item.IsConfiguration())
		})
	}
}

func Test_NewBasketItemFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	factory := NewMockItemFactory(ctrl)
	compFactory := NewBasketItemFactory(factory)
	assert.Equal(t, &CompositeFactory{factories: []ItemFactory{factory}}, compFactory)
}

func TestCompositeFactory_Add(t *testing.T) {
	ctrl := gomock.NewController(t)
	factory := NewMockItemFactory(ctrl)
	compFactory := NewBasketItemFactory()
	compFactory.Add(factory)
	assert.Equal(t, &CompositeFactory{factories: []ItemFactory{factory}}, compFactory)
}

func TestCompositeFactory_Creatable(t *testing.T) {
	tests := []struct {
		name     string
		prepare  func(factory *MockItemFactory)
		itemType Type
		want     bool
	}{
		{
			"is creatable",
			func(factory *MockItemFactory) {
				factory.EXPECT().Creatable(TypeProduct).Return(true).Times(1)
			},
			TypeProduct,
			true,
		},
		{
			"is not creatable",
			func(factory *MockItemFactory) {
				factory.EXPECT().Creatable(TypeProduct).Return(false).Times(1)
			},
			TypeProduct,
			false,
		},
	}

	for _, tt := range tests {
		factory := NewMockItemFactory(gomock.NewController(t))
		compFactory := NewBasketItemFactory(factory)
		tt.prepare(factory)

		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, compFactory.Creatable(tt.itemType))
		})
	}
}

func TestCompositeFactory_Create(t *testing.T) {
	type args struct {
		ctx             context.Context
		itemId          ItemId
		spaceId         store_types.SpaceId
		itemType        Type
		count           int
		parentItem      *Item
		priceColumn     catalog_types.PriceColumn
		user            *userv1.User
		ignoreFairPrice bool
	}
	tests := []struct {
		name    string
		init    func(mockItemFactory *MockItemFactory) (*Item, *CompositeFactory)
		args    args
		wantErr bool
	}{
		{
			"create item",
			func(mockItemFactory *MockItemFactory) (*Item, *CompositeFactory) {
				item := NewItem(
					ItemId("test_item_id"),
					TypeProduct,
					"test product",
					"test image",
					1,
					1000,
					100,
					store_types.SpaceId("test_space"),
					catalog_types.PriceColumnClub,
				)
				mockItemFactory.EXPECT().Creatable(TypeProduct).Return(true).Times(1)
				mockItemFactory.EXPECT().Create(
					context.Background(),
					ItemId("test_item_id"),
					store_types.SpaceId("test_space_id"),
					TypeProduct,
					1,
					nil,
					catalog_types.PriceColumnClub,
					&userv1.User{},
					true,
				).Return(
					item,
					nil,
				).Times(1)

				return item, NewBasketItemFactory(mockItemFactory)
			},
			args{
				ctx:             context.Background(),
				itemId:          ItemId("test_item_id"),
				spaceId:         store_types.SpaceId("test_space_id"),
				itemType:        TypeProduct,
				count:           1,
				parentItem:      nil,
				priceColumn:     catalog_types.PriceColumnClub,
				user:            &userv1.User{},
				ignoreFairPrice: true,
			},
			false,
		},
		{
			"item create error",
			func(mockItemFactory *MockItemFactory) (*Item, *CompositeFactory) {
				item := NewItem(
					ItemId("test_item_id"),
					TypeProduct,
					"test product",
					"test image",
					1,
					1000,
					100,
					store_types.SpaceId("test_space"),
					catalog_types.PriceColumnClub,
				)
				mockItemFactory.EXPECT().Creatable(TypeProduct).Return(true).Times(1)
				mockItemFactory.EXPECT().Create(
					context.Background(),
					ItemId("test_item_id"),
					store_types.SpaceId("test_space_id"),
					TypeProduct,
					1,
					nil,
					catalog_types.PriceColumnClub,
					&userv1.User{},
					true,
				).Return(
					nil,
					errors.New("test error"),
				).Times(1)

				return item, NewBasketItemFactory(mockItemFactory)
			},
			args{
				ctx:             context.Background(),
				itemId:          ItemId("test_item_id"),
				spaceId:         store_types.SpaceId("test_space_id"),
				itemType:        TypeProduct,
				count:           1,
				parentItem:      nil,
				priceColumn:     catalog_types.PriceColumnClub,
				user:            &userv1.User{},
				ignoreFairPrice: true,
			},
			true,
		},
		{
			"item is not creatable",
			func(mockItemFactory *MockItemFactory) (*Item, *CompositeFactory) {
				item := NewItem(
					ItemId("test_item_id"),
					TypeProduct,
					"test product",
					"test image",
					1,
					1000,
					100,
					store_types.SpaceId("test_space"),
					catalog_types.PriceColumnClub,
				)
				mockItemFactory.EXPECT().Creatable(TypeProduct).Return(false).Times(1)

				return item, NewBasketItemFactory(mockItemFactory)
			},
			args{
				ctx:             context.Background(),
				itemId:          ItemId("test_item_id"),
				spaceId:         store_types.SpaceId("test_space_id"),
				itemType:        TypeProduct,
				count:           1,
				parentItem:      nil,
				priceColumn:     catalog_types.PriceColumnClub,
				user:            &userv1.User{},
				ignoreFairPrice: true,
			},
			true,
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		mockItemFactory := NewMockItemFactory(ctrl)
		item, cf := tt.init(mockItemFactory)

		t.Run(tt.name, func(t *testing.T) {

			got, err := cf.Create(
				tt.args.ctx,
				tt.args.itemId,
				tt.args.spaceId,
				tt.args.itemType,
				tt.args.count,
				tt.args.parentItem,
				tt.args.priceColumn,
				tt.args.user,
				tt.args.ignoreFairPrice,
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, item, got)
				assert.NoError(t, err)
			}
		})
	}
}

func TestItem_Fingerprint(t *testing.T) {
	item := NewItem(
		ItemId("test_item_id"),
		TypeProduct,
		"test product",
		"test image",
		1,
		1000,
		10,
		store_types.SpaceId("test_space_id"),
		catalog_types.PriceColumnClub,
	)
	assert.Equal(t, "1149408419158322160", item.Fingerprint())
}

func TestItem_Type(t *testing.T) {
	item := NewItem(
		ItemId("test_item_id"),
		TypeProduct,
		"test product",
		"test image",
		1,
		1000,
		10,
		store_types.SpaceId("test_space_id"),
		catalog_types.PriceColumnClub,
	)
	assert.Equal(t, TypeProduct, item.Type())
}

func TestItem_ParentUniqId(t *testing.T) {
	parent := NewItem(
		ItemId("test_parent_item_id"),
		TypeConfiguration,
		"test parent product",
		"test parent image",
		1,
		1000,
		10,
		store_types.SpaceId("test_space_id"),
		catalog_types.PriceColumnClub,
	)
	item := NewItem(
		ItemId("test_item_id"),
		TypeConfigurationProduct,
		"test product",
		"test image",
		1,
		1000,
		10,
		store_types.SpaceId("test_space_id"),
		catalog_types.PriceColumnClub,
	)

	err := parent.AddChild(item)

	assert.NoError(t, err)
	assert.Equal(t, parent.UniqId(), item.ParentUniqId())
}

func TestItem_IsChild(t *testing.T) {
	tests := []struct {
		name string
		init func() *Item
		want bool
	}{
		{
			"is child",
			func() *Item {
				parent := NewItem(
					ItemId("test_parent_item_id"),
					TypeConfiguration,
					"test parent product",
					"test parent image",
					1,
					1000,
					10,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)
				item := NewItem(
					ItemId("test_item_id"),
					TypeConfigurationProduct,
					"test product",
					"test image",
					1,
					1000,
					10,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				parent.AddChild(item)

				return item
			},
			true,
		},
		{
			"is not child",
			func() *Item {
				item := NewItem(
					ItemId("test_item_id"),
					TypeConfigurationProduct,
					"test product",
					"test image",
					1,
					1000,
					10,
					store_types.SpaceId("test_space_id"),
					catalog_types.PriceColumnClub,
				)

				return item
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tt.init()
			assert.Equal(t, tt.want, item.IsChild())
		})
	}
}

func TestItem_SetBonus(t *testing.T) {
	item := Item{}
	item.SetBonus(100)
	assert.Equal(t, 100, item.bonus)
}

func TestItem_Bonus(t *testing.T) {
	item := Item{bonus: 100}
	assert.Equal(t, 100, item.Bonus())
}

func TestItem_FixBonus(t *testing.T) {
	item := Item{}
	item.FixBonus(100)
	assert.Equal(t, 100, item.bonus)
}

func TestItem_FixPrice(t *testing.T) {
	item := Item{}
	item.FixPrice(100)
	assert.Equal(t, 100, item.price)
}

func TestItem_Count(t *testing.T) {
	item := Item{count: 100}
	assert.Equal(t, 100, item.Count())
}

func TestItem_FixCount(t *testing.T) {
	item := Item{}
	item.FixCount(100)
	assert.Equal(t, 100, item.count)
}

func TestItem_FixName(t *testing.T) {
	item := Item{}
	item.FixName("new name")
	assert.Equal(t, "new name", item.name)
}

func TestItem_Name(t *testing.T) {
	item := Item{name: "test name"}
	assert.Equal(t, "test name", item.Name())
}

func TestItem_Image(t *testing.T) {
	item := Item{image: "test image"}
	assert.Equal(t, "test image", item.Image())
}

func TestItem_SetPrepaymentMandatory(t *testing.T) {
	item := Item{}
	item.SetPrepaymentMandatory(true)
	assert.Equal(t, true, item.isPrepaymentMandatory)
}

func TestItem_IsPrepaymentMandatory(t *testing.T) {
	item := Item{isPrepaymentMandatory: true}
	assert.Equal(t, true, item.IsPrepaymentMandatory())
}

func TestItem_CommitChanges(t *testing.T) {
	item := Item{}
	item.CommitChanges()
	assert.Equal(t, item.Fingerprint(), item.commitFingerprint)
}

func TestItem_IsChanged(t *testing.T) {
	item := Item{commitFingerprint: "test fingerprint"}
	assert.True(t, item.IsChanged())
}

func TestItem_SetMarkedPurchaseReason(t *testing.T) {
	item := Item{}
	item.SetMarkedPurchaseReason(MarkedPurchaseReasonForResale)
	assert.Equal(t, MarkedPurchaseReasonForResale, item.markedPurchaseReason)
}

func TestItem_MarkedPurchaseReason(t *testing.T) {
	item := Item{markedPurchaseReason: MarkedPurchaseReasonForResale}
	assert.Equal(t, MarkedPurchaseReasonForResale, item.MarkedPurchaseReason())
}

func TestItem_SetHasFairPrice(t *testing.T) {
	item := Item{}
	item.SetHasFairPrice(true)
	assert.True(t, item.hasFairPrice)
}

func TestItem_HasFairPrice(t *testing.T) {
	item := Item{hasFairPrice: true}
	assert.True(t, item.HasFairPrice())
}

func TestItem_SetIgnoreFairPrice(t *testing.T) {
	item := Item{ignoreFairPrice: true}

	item.SetIgnoreFairPrice(true)
	assert.True(t, item.ignoreFairPrice)
	assert.False(t, item.ignoreFairPriceChanged)
}

func TestItem_IgnoreFairPrice(t *testing.T) {
	item := Item{ignoreFairPrice: true}
	assert.True(t, item.IgnoreFairPrice())
}

func TestItem_IgnoreFairPriceChanged(t *testing.T) {
	item := Item{ignoreFairPriceChanged: true}
	assert.True(t, item.IgnoreFairPriceChanged())
}

func TestItem_Cost(t *testing.T) {
	item := Item{count: 2, price: 1000}
	assert.Equal(t, 2000, item.Cost())
}

func TestItem_SpaceId(t *testing.T) {
	item := Item{spaceId: store_types.SpaceId("test_space")}
	assert.Equal(t, store_types.SpaceId("test_space"), item.SpaceId())
}

func TestItem_SetSpaceId(t *testing.T) {
	item := Item{}
	item.SetSpaceId(store_types.SpaceId("test_space"))
	assert.Equal(t, store_types.SpaceId("test_space"), item.spaceId)
}

func TestItem_Price(t *testing.T) {
	item := Item{price: 3000}
	assert.Equal(t, 3000, item.Price())
}

func TestItem_SetPrice(t *testing.T) {
	item := Item{}
	item.SetPrice(3000)
	assert.Equal(t, 3000, item.price)
}

func TestItem_SetCount(t *testing.T) {
	tests := []struct {
		name            string
		item            *Item
		wantErr         bool
		maxItemErr      bool
		errMessage      string
		expectedCount   int
		maxCountForRule int
		multiplicity    int
	}{
		{
			"can't set more than maxCount error",
			generateItem("123123", TypeProduct),
			true,
			true,
			"can't set count 3 more then max count 2 for this item",
			3, 2, 1,
		},
		{
			"not changeable count error",
			generateItem("123123", TypePresent),
			true,
			false,
			"count of position with type 'present' can't be changed",
			3, 2, 1,
		},
		{
			"larger than max count but correct",
			generateItem("123123", TypeProduct),
			false,
			false,
			"",
			LimitTotalGoods + 1, 1500, 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(
			tt.name, func(t *testing.T) {
				tt.item.additions = ItemAdditions{
					Product:                      &ProductItemAdditions{},
					Configuration:                &ConfiguratorItemAdditions{},
					SubcontractServiceForProduct: &SubcontractItemAdditions{},
					Service:                      &Service{},
				}
				tt.item.countMultiplicity = tt.multiplicity
				tt.item.Rules().SetMaxCount(tt.maxCountForRule)
				err := tt.item.SetCount(tt.expectedCount)
				if tt.wantErr {
					assert.EqualError(t, err, tt.errMessage)
					if tt.maxItemErr {
						var expectedErrType *MaxItemCountError
						assert.True(t, errors.As(err, &expectedErrType))
						assert.Equal(t, tt.maxCountForRule, expectedErrType.MaxCount())
					}
				} else {
					assert.False(t, tt.item.Additions().GetProduct().IsCountMoreThenAvailChecked())
					assert.Nil(t, err)
				}
			},
		)
	}
}

func TestItem_AddChild(t *testing.T) {
	item := NewItem(
		ConfSpecialItemId,
		TypeConfiguration,
		"conf",
		"image",
		1,
		1,
		1,
		store_types.NewSpaceId("msk_cl"),
		catalog_types.PriceColumnRetail,
	)
	tests := []struct {
		name    string
		child   *Item
		wantErr bool
	}{
		{
			"correct",
			generateItem("123123", TypeConfigurationProduct),
			false,
		},
		{
			"incorrect type",
			generateItem("123123", TypePresent),
			true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(
			tt.name, func(t *testing.T) {
				err := item.AddChild(tt.child)
				if tt.wantErr {
					assert.Errorf(t, err, err.Error())
				} else {
					assert.Nil(t, err)
				}
			},
		)
	}
}

func TestItem_MakeChildOf(t *testing.T) {
	item := NewItem(
		"123123",
		TypeConfigurationProduct,
		"conf_child",
		"image",
		1,
		10,
		1,
		store_types.NewSpaceId("msk_cl"),
		catalog_types.PriceColumnRetail,
	)
	tests := []struct {
		name    string
		parent  *Item
		wantErr bool
	}{
		{
			"correct",
			NewItem(
				ConfSpecialItemId,
				TypeConfiguration,
				"conf",
				"image",
				1,
				1,
				1,
				store_types.NewSpaceId("msk_cl"),
				catalog_types.PriceColumnRetail,
			),
			false,
		},
		{
			"incorrect type",
			NewItem(
				ConfSpecialItemId,
				TypePresent,
				"present",
				"image",
				1,
				1,
				1,
				store_types.NewSpaceId("msk_cl"),
				catalog_types.PriceColumnRetail,
			),
			true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(
			tt.name, func(t *testing.T) {
				err := item.MakeChildOf(tt.parent)
				if tt.wantErr {
					assert.Errorf(t, err, err.Error())
				} else {
					assert.Nil(t, err)
				}
			},
		)
	}
}

func TestItem_IsChildOf(t *testing.T) {
	item := generateItem("123123", TypeConfigurationProduct)
	parentItem := generateItem(ConfSpecialItemId, TypeConfiguration)
	_ = item.MakeChildOf(parentItem)
	want := item.IsChildOf(parentItem)
	assert.True(t, want)
}

func TestItem_CorrectCountFromMultiplicity(t *testing.T) {
	item := generateItem("123123", TypeProduct)
	tests := []struct {
		count          int
		multiplicity   int
		expectedResult int
	}{
		{1, 0, 1},
		{1, 1, 1},
		{1, 10, 10},
		{1, 40, 40},
		{10, 10, 10},
		{10, 40, 40},
		{40, 10, 40},
		{40, 40, 40},
		{50, 10, 50},
		{50, 40, 40},
		{50, 11, 44},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(
			"", func(t *testing.T) {
				item.countMultiplicity = tt.multiplicity
				result := item.CorrectCountFromMultiplicity(tt.count)
				assert.Equal(t, tt.expectedResult, result)
			},
		)
	}
}

func TestItem_SortTypeValue(t *testing.T) {
	tests := []struct {
		name string
		item *Item
		want int
	}{
		{"product", &Item{itemType: TypeProduct}, 1},
		{"subcontract_service", &Item{itemType: TypeSubcontractServiceForProduct}, 2},
		{"insurance_service", &Item{itemType: TypeInsuranceServiceForProduct}, 2},
		{"digital_service", &Item{itemType: TypeDigitalService}, 2},
		{"property_service", &Item{itemType: TypePropertyInsurance}, 2},
		{"delivery_service", &Item{itemType: TypeDeliveryService}, 2},
		{"lifting_service", &Item{itemType: TypeLiftingService}, 2},
		{"configuration", &Item{itemType: TypeConfiguration}, 3},
		{"conf_product", &Item{itemType: TypeConfigurationProduct}, 4},
		{"conf_product_service", &Item{itemType: TypeConfigurationProductService}, 5},
		{"conf_assembly_service", &Item{itemType: TypeConfigurationAssemblyService}, 5},
		{"unknown", &Item{itemType: TypeUnknown}, 100},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(
			tt.name, func(t *testing.T) {
				got := tt.item.SortTypeValue()
				assert.Equal(t, tt.want, got)
			},
		)
	}
}

func TestItems_ItemIds(t *testing.T) {
	var items Items
	items = append(
		items, []*Item{
			generateItem("123456", TypeProduct),
			generateItem("123457", TypeProduct),
			generateItem("123458", TypeProduct),
			generateItem("123459", TypeProduct),
			generateItem("123460", TypeProduct),
		}...,
	)
	got := items.ItemIds()
	assert.Equal(t, got, []ItemId{"123456", "123457", "123458", "123459", "123460"})
}

func TestItems_ItemIdsStrings(t *testing.T) {
	var items Items
	items = append(
		items, []*Item{
			generateItem("123456", TypeProduct),
			generateItem("123457", TypeProduct),
			generateItem("123458", TypeProduct),
			generateItem("123459", TypeProduct),
			generateItem("123460", TypeProduct),
		}...,
	)
	got := items.ItemIdsStrings()
	assert.Equal(t, got, []string{"123456", "123457", "123458", "123459", "123460"})
}

func TestItems_UniqIds(t *testing.T) {
	var items Items
	items = append(
		items, []*Item{
			generateItem("123456", TypeProduct),
			generateItem("123457", TypeProduct),
			generateItem("123458", TypeProduct),
			generateItem("123459", TypeProduct),
			generateItem("123460", TypeProduct),
		}...,
	)
	got := items.UniqIds()
	assert.Len(t, got, 5)
}

func TestItems_ToXItems(t *testing.T) {
	var items Items
	items = append(
		items, []*Item{
			generateItem("123456", TypeProduct),
			generateItem("123457", TypeProduct),
			generateItem("123458", TypeProduct),
			generateItem("123459", TypePresent),
			generateItem("123460", TypeDigitalService),
		}...,
	)
	got := items.ToXItems()
	assert.IsType(t, &XItem{}, got[0])
	assert.Len(t, got, 5)
}

func TestItems_ToItemList(t *testing.T) {
	var items Items
	items = append(
		items, []*Item{
			generateItem("123456", TypeProduct),
			generateItem("123457", TypeProduct),
			generateItem("123458", TypeProduct),
			generateItem("123459", TypePresent),
			generateItem("123460", TypeDigitalService),
		}...,
	)
	got := items.ToItemList()
	assert.IsType(t, &ItemList{}, got)
	assert.Len(t, got.Items, 5)
}

func TestItems_First(t *testing.T) {
	var items Items
	items = append(
		items, []*Item{
			generateItem("123456", TypeProduct),
			generateItem("123457", TypeProduct),
			generateItem("123458", TypeProduct),
			generateItem("123459", TypePresent),
			generateItem("123460", TypeDigitalService),
		}...,
	)
	got := items.First()
	assert.Equal(t, ItemId("123456"), got.itemId)

	emptyItems := make(Items, 0)
	gotNil := emptyItems.First()
	assert.Nil(t, gotNil)
}

func TestItemMap_ToSlice(t *testing.T) {
	items := make(ItemMap)
	items[UniqId("1")] = generateItem("123456", TypeProduct)
	items[UniqId("2")] = generateItem("123457", TypeProduct)
	items[UniqId("3")] = generateItem("123458", TypeProduct)
	items[UniqId("4")] = generateItem("123459", TypePresent)
	items[UniqId("5")] = generateItem("123460", TypeDigitalService)
	got := items.ToSlice()
	assert.Len(t, got, 5)
}

func TestItem_CalculateBonus(t *testing.T) {
	type args struct {
		user   *userv1.User
		prices *productv1.ProductPriceByRegion
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "B2B take correct bonus",
			args: args{
				&userv1.User{
					LpStatusAsString: "SILVER",
					B2B: &userv1.User_B2B{
						IsB2BState: true,
					},
				},
				&productv1.ProductPriceByRegion{
					Bonuses: []*productv1.ProductPriceByRegion_Bonus{
						{
							LoyaltyStatus: &productv1.ProductPriceByRegion_LoyaltyStatus{
								Code: "NOT_THIS",
							},
							BonusB2C: 1,
							BonusB2B: 2,
						},
						{
							LoyaltyStatus: &productv1.ProductPriceByRegion_LoyaltyStatus{
								Code: "SILVER",
							},
							BonusB2C: 10,
							BonusB2B: 20,
						},
						{
							LoyaltyStatus: &productv1.ProductPriceByRegion_LoyaltyStatus{
								Code: "AND_NOT_THAT",
							},
							BonusB2C: 100,
							BonusB2B: 200,
						},
					},
				},
			},
			want: 20,
		},
		{
			name: "B2C take correct bonus",
			args: args{
				&userv1.User{
					LpStatusAsString: "SILVER",
					B2B: &userv1.User_B2B{
						IsB2BState: false,
					},
				},
				&productv1.ProductPriceByRegion{
					Bonuses: []*productv1.ProductPriceByRegion_Bonus{
						{
							LoyaltyStatus: &productv1.ProductPriceByRegion_LoyaltyStatus{
								Code: "NOT_THIS",
							},
							BonusB2C: 1,
							BonusB2B: 2,
						},
						{
							LoyaltyStatus: &productv1.ProductPriceByRegion_LoyaltyStatus{
								Code: "SILVER",
							},
							BonusB2C: 10,
							BonusB2B: 20,
						},
						{
							LoyaltyStatus: &productv1.ProductPriceByRegion_LoyaltyStatus{
								Code: "AND_NOT_THAT",
							},
							BonusB2C: 100,
							BonusB2B: 200,
						},
					},
				},
			},
			want: 10,
		},
		{
			name: "empty bonuses - zero bonus",
			args: args{
				&userv1.User{
					LpStatusAsString: "SILVER",
					B2B: &userv1.User_B2B{
						IsB2BState: true,
					},
				},
				&productv1.ProductPriceByRegion{
					Bonuses: []*productv1.ProductPriceByRegion_Bonus{},
				},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateBonus(tt.args.user, tt.args.prices); got != tt.want {
				t.Errorf("CalculateBonus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfiguratorItemAdditions_IsMutable(t *testing.T) {
	tests := []struct {
		name string
		args *ConfiguratorItemAdditions
		want bool
	}{
		{
			name: "citilink template",
			args: &ConfiguratorItemAdditions{
				ConfId:   "123",
				ConfType: ConfTypeTemplate,
			},
			want: false,
		},
		{
			name: "vendor template",
			args: &ConfiguratorItemAdditions{
				ConfId:   "123",
				ConfType: ConfTypeVendor,
			},
			want: false,
		},
		{
			name: "user template",
			args: &ConfiguratorItemAdditions{
				ConfId:   "123",
				ConfType: ConfTypeUser,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.IsMutable(); got != tt.want {
				t.Errorf("%s: expected - %t, got - %t", tt.name, tt.want, got)
			}
		})
	}
}

func TestRules_IsMaxCount(t *testing.T) {
	rules := Rules{maxCount: 1}
	assert.True(t, rules.IsMaxCount())
	rules = Rules{}
	assert.False(t, rules.IsMaxCount())
}

func TestRules_MaxCount(t *testing.T) {
	rules := Rules{maxCount: 2}
	assert.Equal(t, 2, rules.MaxCount())
}

func TestRules_SetMaxCount(t *testing.T) {
	rules := Rules{}
	rules.SetMaxCount(2)
	assert.Equal(t, 2, rules.maxCount)
}

func TestItem_ItemId(t *testing.T) {
	itm := Item{itemId: ItemId("test")}
	assert.Equal(t, ItemId("test"), itm.ItemId())
}

func TestItem_setParent(t *testing.T) {
	tests := []struct {
		name    string
		itm     *Item
		prnt    *Item
		wantErr bool
	}{
		{
			"parent can not have child",
			&Item{},
			&Item{
				uniqId:   UniqId("test_parent_unique_id"),
				itemId:   ItemId("test_parent_item_id"),
				itemType: TypeProduct,
			},
			true,
		},
		{
			"parent set",
			&Item{
				itemType: TypeInsuranceServiceForProduct,
			},
			&Item{
				uniqId:   UniqId("test_parent_unique_id"),
				itemId:   ItemId("test_parent_item_id"),
				itemType: TypeProduct,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.itm.setParent(tt.prnt)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, UniqId("test_parent_unique_id"), tt.itm.parentUniqId)
				assert.Equal(t, ItemId("test_parent_item_id"), tt.itm.parentItemId)
			}
		})
	}
}

func TestItem_Spec(t *testing.T) {
	itm := Item{itemType: TypeProduct}
	assert.Equal(t, typeToSpecs[TypeProduct], itm.Spec())
}

func TestItem_ToXItem(t *testing.T) {
	tests := []struct {
		name string
		item *Item
		want *XItem
	}{
		{
			"is not present, is not service, not configuration",
			&Item{
				itemId:       ItemId("test_item_id"),
				count:        2,
				price:        100,
				itemType:     TypeProduct,
				bonus:        10,
				parentItemId: ItemId(""),
			},
			&XItem{
				ItemId:                    "test_item_id",
				Count:                     2,
				Count1:                    2,
				Price1:                    100,
				Price2:                    100,
				Price3:                    100,
				Discount:                  0,
				IsPresent:                 0,
				Bonus:                     10,
				NavisionType:              int(NavTypeProduct),
				ParentItemId:              "",
				IsService:                 0,
				Price2WithoutLoyaltyBonus: 100,
			},
		},
		{
			"is present",
			&Item{
				itemId:       ItemId("test_item_id"),
				count:        2,
				price:        100,
				itemType:     TypePresent,
				bonus:        10,
				parentItemId: ItemId(""),
			},
			&XItem{
				ItemId:                    "test_item_id",
				Count:                     2,
				Count1:                    2,
				Price1:                    100,
				Price2:                    100,
				Price3:                    100,
				Discount:                  0,
				IsPresent:                 1,
				Bonus:                     10,
				NavisionType:              int(NavTypeProduct),
				ParentItemId:              "",
				IsService:                 0,
				Price2WithoutLoyaltyBonus: 100,
			},
		},
		{
			"is service",
			&Item{
				itemId:       ItemId("test_item_id"),
				count:        2,
				price:        100,
				itemType:     TypeSubcontractServiceForProduct,
				bonus:        10,
				parentItemId: ItemId(""),
			},
			&XItem{
				ItemId:                    "test_item_id",
				Count:                     2,
				Count1:                    2,
				Price1:                    100,
				Price2:                    100,
				Price3:                    100,
				Discount:                  0,
				IsPresent:                 0,
				Bonus:                     10,
				NavisionType:              int(NavTypeService),
				ParentItemId:              "",
				IsService:                 1,
				Price2WithoutLoyaltyBonus: 100,
			},
		},
		{
			"is configuration",
			&Item{
				itemId:       ItemId("test_item_id"),
				count:        2,
				price:        100,
				itemType:     TypeConfiguration,
				bonus:        10,
				parentItemId: ItemId(""),
				additions: ItemAdditions{
					Configuration: &ConfiguratorItemAdditions{
						ConfId:   "test_conf",
						ConfType: ConfTypeUser,
					},
				},
			},
			&XItem{
				ItemId:                    "test_item_id",
				Count:                     2,
				Count1:                    2,
				Price1:                    100,
				Price2:                    100,
				Price3:                    100,
				Discount:                  0,
				IsPresent:                 0,
				Bonus:                     10,
				NavisionType:              int(NavTypeProduct),
				ParentItemId:              "",
				IsService:                 0,
				Price2WithoutLoyaltyBonus: 100,
				ConfId:                    "test_conf",
				ConfType:                  1,
				IsConfiguration:           1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.item.ToXItem())
		})
	}
}

func TestItem_Problems(t *testing.T) {
	problems := []*Problem{
		{id: ProblemId(1)},
		{id: ProblemId(2)},
	}
	permProblems := []*Problem{
		{id: ProblemId(3)},
	}

	itm := Item{problems: problems, permanentProblems: permProblems}
	assert.Equal(t, []*Problem{
		{id: ProblemId(1)},
		{id: ProblemId(2)},
		{id: ProblemId(3)},
	}, itm.Problems())
}

func TestItem_DeleteProblems(t *testing.T) {
	itm := Item{problems: []*Problem{
		{id: ProblemId(1)},
	}}
	itm.DeleteProblems()
	assert.Nil(t, itm.problems)
}

func TestItem_SimulateProblem(t *testing.T) {
	itm := Item{permanentProblems: []*Problem{
		{id: ProblemId(1)},
	}}
	itm.SimulateProblem(&Problem{id: ProblemId(2)})

	assert.Equal(t, []*Problem{
		{id: ProblemId(1)},
		{id: ProblemId(2)},
	}, itm.permanentProblems)
}

func TestItem_CancelSimulateProblems(t *testing.T) {
	itm := Item{permanentProblems: []*Problem{
		{id: ProblemId(1)},
	}}
	itm.CancelSimulateProblems()

	assert.Nil(t, itm.permanentProblems)
}

func TestItem_CountMultiplicity(t *testing.T) {
	itm := Item{countMultiplicity: 5}
	assert.Equal(t, 5, itm.CountMultiplicity())
}

func TestItem_SetCountMultiplicity(t *testing.T) {
	itm := Item{}
	itm.SetCountMultiplicity(8)
	assert.Equal(t, 8, itm.countMultiplicity)
}

func TestItem_AddProblem(t *testing.T) {
	itm := Item{isSelected: true, problems: []*Problem{{id: ProblemId(1)}}}
	itm.AddProblem(&Problem{id: ProblemId(2)})
	assert.Equal(t, []*Problem{
		{id: ProblemId(1)},
		{id: ProblemId(2)},
	}, itm.problems)
}

func TestItem_Infos(t *testing.T) {
	itm := Item{
		infos: map[InfoId]*Info{
			InfoId(2): {id: InfoId(2)},
		},
	}
	assert.Equal(t, map[InfoId]*Info{InfoId(2): {id: InfoId(2)}}, itm.Infos())
}

func TestItem_AddInfo(t *testing.T) {
	itm := Item{
		infos: map[InfoId]*Info{
			InfoId(2): {id: InfoId(2)},
		},
	}
	itm.AddInfo(&Info{
		id: InfoId(3),
	})
	assert.Equal(t, map[InfoId]*Info{
		InfoId(2): {id: InfoId(2)},
		InfoId(3): {id: InfoId(3)},
	}, itm.infos)
}

func TestItem_CommitInfo(t *testing.T) {
	itm := Item{
		infos: map[InfoId]*Info{
			InfoId(2): {id: InfoId(2)},
			InfoId(3): {id: InfoId(3)},
		},
	}

	itm.CommitInfo(InfoId(3))
	assert.Equal(t, map[InfoId]*Info{
		InfoId(2): {id: InfoId(2)},
	}, itm.infos)
}

func TestItem_Additions(t *testing.T) {
	itm := Item{additions: ItemAdditions{
		Product: &ProductItemAdditions{
			isAvailInStore: true,
			vat:            20,
			categoryId:     catalog_types.CategoryId(10),
			isOEM:          true,
		},
	}}
	assert.Equal(t, &ItemAdditions{
		Product: &ProductItemAdditions{
			isAvailInStore: true,
			vat:            20,
			categoryId:     catalog_types.CategoryId(10),
			isOEM:          true,
		},
	}, itm.Additions())
}

func TestItem_Rules(t *testing.T) {
	itm := Item{
		rules: Rules{
			maxCount: 10,
		},
	}
	assert.Equal(t, &Rules{maxCount: 10}, itm.Rules())
}

func TestItem_SetPriceColumn(t *testing.T) {
	itm := Item{}
	itm.SetPriceColumn(catalog_types.PriceColumnClub)
	assert.Equal(t, catalog_types.PriceColumnClub, itm.priceColumn)
}

func TestItem_PriceColumn(t *testing.T) {
	itm := Item{priceColumn: catalog_types.PriceColumnClub}
	assert.Equal(t, catalog_types.PriceColumnClub, itm.PriceColumn())
}

func TestNewMaxItemCountError(t *testing.T) {
	err := NewMaxItemCountError(errors.New("test error"), 5)
	assert.Equal(t, errors.New("test error"), err.err)
	assert.Equal(t, 5, err.maxCount)
}

func TestMaxItemCountError_MaxCount(t *testing.T) {
	err := MaxItemCountError{maxCount: 5}
	assert.Equal(t, 5, err.MaxCount())
}

func TestMaxItemCountError_Error(t *testing.T) {
	err := MaxItemCountError{err: errors.New("test error")}
	assert.Equal(t, "test error", err.Error())
}

func TestMaxItemCountError_Unwrap(t *testing.T) {
	maxItemCountError := NewMaxItemCountError(errors.New("test error"), 10)
	assert.Equal(t, errors.New("test error"), maxItemCountError.Unwrap())
}

func TestSpec_CanHaveChildren(t *testing.T) {
	s := Spec{childrenTypes: []Type{}}
	assert.False(t, s.CanHaveChildren())
	s = Spec{childrenTypes: []Type{
		Type("test_type"),
	}}
	assert.True(t, s.CanHaveChildren())
}

func TestSpec_MustBeAChild(t *testing.T) {
	s := Spec{mustBeAChild: true}
	assert.True(t, s.MustBeAChild())
}

func TestSpec_CanHaveChild(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		itmType Type
		want    bool
	}{
		{
			"can not have children",
			Spec{
				childrenTypes: []Type{},
			},
			Type("test"),
			false,
		},
		{
			"not valid item type",
			Spec{
				childrenTypes: []Type{
					TypeProduct,
					TypeConfiguration,
				},
			},
			Type("test"),
			false,
		},
		{
			"item type found",
			Spec{
				childrenTypes: []Type{
					TypeProduct,
					TypeConfiguration,
				},
			},
			TypeProduct,
			true,
		},
		{
			"item type not found",
			Spec{
				childrenTypes: []Type{
					TypeProduct,
					TypeConfiguration,
				},
			},
			TypeDeliveryService,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.spec.CanHaveChild(tt.itmType))
		})
	}
}

func TestSpec_ChildrenTypes(t *testing.T) {
	s := Spec{childrenTypes: []Type{
		TypeProduct,
	}}
	assert.Equal(t, []Type{TypeProduct}, s.ChildrenTypes())
}

func TestSpec_IsCountLessOrEqualThenParent(t *testing.T) {
	s := Spec{isCountLessOrEqualThenParent: true}
	assert.True(t, s.IsCountLessOrEqualThenParent())
}

func TestSpec_IsDeletable(t *testing.T) {
	s := Spec{isDeletable: true}
	assert.True(t, s.IsDeletable())
}

func TestSpec_IsCountChangeable(t *testing.T) {
	s := Spec{isCountChangeable: true}
	assert.True(t, s.IsCountChangeable())
}

func TestSpec_IsOnlyOnePositionPossible(t *testing.T) {
	s := Spec{isOnlyOnePositionPossible: true}
	assert.True(t, s.IsOnlyOnePositionPossible())
}

func TestSpec_IsOnlyOnePositionPerParent(t *testing.T) {
	s := Spec{isOnlyOnePositionPerParent: true}
	assert.True(t, s.IsOnlyOnePositionPerParent())
}

func TestSpec_IsAllowedForPerson(t *testing.T) {
	s := Spec{allowedUserTypes: specPersonType}
	assert.True(t, s.IsAllowedForPerson())
	s = Spec{allowedUserTypes: specB2bUserType}
	assert.False(t, s.IsAllowedForPerson())
}

func TestSpec_IsAllowedForB2bUser(t *testing.T) {
	s := Spec{allowedUserTypes: specB2bUserType}
	assert.True(t, s.IsAllowedForB2bUser())
	s = Spec{allowedUserTypes: specPersonType}
	assert.False(t, s.IsAllowedForB2bUser())
}

func TestSpec_IsCountEqualToParentCount(t *testing.T) {
	s := Spec{isCountEqualToParentCount: true}
	assert.True(t, s.IsCountEqualToParentCount())
}

func TestType_Validate(t *testing.T) {
	assert.Nil(t, TypeProduct.Validate())
	assert.Error(t, Type("test").Validate())
}

func TestType_Spec(t *testing.T) {
	tests := []struct {
		name      string
		spec      *Spec
		typ       Type
		wantPanic bool
	}{
		{
			"spec found",
			&Spec{
				childrenTypes: []Type{
					TypeInsuranceServiceForProduct,
					TypeSubcontractServiceForProduct,
					TypeDigitalService,
					TypePresent,
				},
				allowedUserTypes:  bitmap(3),
				isDeletable:       true,
				isCountChangeable: true,
			},
			TypeProduct,
			false,
		},
		{
			"spec not found; panic",
			&Spec{},
			Type("test"),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() { tt.typ.Spec() })
			} else {
				assert.Equal(t, tt.spec, tt.typ.Spec())
			}
		})
	}
}

func TestType_IsProduct(t *testing.T) {
	tests := []struct {
		name      string
		typ       Type
		isProduct bool
	}{
		{
			"type product",
			TypeProduct,
			true,
		},
		{
			"subcontract service for product",
			TypeSubcontractServiceForProduct,
			false,
		},
		{
			"insurance service for product",
			TypeInsuranceServiceForProduct,
			false,
		},
		{
			"digital service",
			TypeDigitalService,
			false,
		},
		{
			"present",
			TypePresent,
			false,
		},
		{
			"property insurance",
			TypePropertyInsurance,
			false,
		},
		{
			"configuration",
			TypeConfiguration,
			false,
		},
		{
			"configuration product",
			TypeConfigurationProduct,
			true,
		},
		{
			"configuration product service",
			TypeConfigurationProductService,
			false,
		},
		{
			"configuration assembly service",
			TypeConfigurationAssemblyService,
			false,
		},
		{
			"delivery service",
			TypeDeliveryService,
			false,
		},
		{
			"lifting service",
			TypeLiftingService,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isProduct, tt.typ.IsProduct())
		})
	}
}

func TestItems_Sort(t *testing.T) {
	tests := []struct {
		name   string
		is     Items
		parent *Item
		want   Items
	}{
		{
			"empty items list",
			Items{},
			&Item{},
			Items{},
		},
		{
			"parent is nil", // stack overflow
			Items{
				{
					uniqId: UniqId("test_item_2"),
					name:   "test item 2",
				},
				{
					uniqId: UniqId("test_item_1"),
					name:   "test item 1",
				},
			},
			nil,
			Items{
				{
					uniqId: UniqId("test_item_1"),
					name:   "test item 1",
				},
				{
					uniqId: UniqId("test_item_2"),
					name:   "test item 2",
				},
			},
		},
		{
			"sort children",
			Items{
				{
					uniqId:       UniqId("test_item_2"),
					name:         "test item 2",
					parentUniqId: UniqId("parent_item"),
				},
				{
					uniqId:       UniqId("test_item_1"),
					name:         "test item 1",
					parentUniqId: UniqId("parent_item"),
				},
				{
					uniqId: UniqId("parent_item"),
					name:   "parent item",
				},
			},
			&Item{
				uniqId: UniqId("parent_item"),
				name:   "parent item",
			},
			Items{
				{
					uniqId:       UniqId("test_item_1"),
					name:         "test item 1",
					parentUniqId: UniqId("parent_item"),
				},
				{
					uniqId:       UniqId("test_item_2"),
					name:         "test item 2",
					parentUniqId: UniqId("parent_item"),
				},
			},
		},
		{
			"sort by type",
			Items{
				{
					uniqId:   UniqId("test_item_1"),
					name:     "test item 1",
					itemType: TypeConfiguration,
				},
				{
					uniqId:   UniqId("test_item_2"),
					name:     "test item 2",
					itemType: TypeProduct,
				},
			},
			nil,
			Items{
				{
					uniqId:   UniqId("test_item_2"),
					name:     "test item 2",
					itemType: TypeProduct,
				},
				{
					uniqId:   UniqId("test_item_1"),
					name:     "test item 1",
					itemType: TypeConfiguration,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.is.Sort(tt.parent)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculateBonus(t *testing.T) {
	type args struct {
		user   *userv1.User
		prices *productv1.ProductPriceByRegion
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			"user is nil",
			args{
				nil,
				&productv1.ProductPriceByRegion{},
			},
			0,
		},
		{
			"bonus not found",
			args{
				&userv1.User{
					LpStatusAsString: "test status",
				},
				&productv1.ProductPriceByRegion{
					Bonuses: []*productv1.ProductPriceByRegion_Bonus{
						{},
					},
				},
			},
			0,
		},
		{
			"bonus b2c",
			args{
				&userv1.User{
					LpStatusAsString: "test status",
					B2B: &userv1.User_B2B{
						IsB2BState: false,
					},
				},
				&productv1.ProductPriceByRegion{
					Bonuses: []*productv1.ProductPriceByRegion_Bonus{
						{
							LoyaltyStatus: &productv1.ProductPriceByRegion_LoyaltyStatus{
								Code: "test status",
							},
							BonusB2C: 100,
						},
					},
				},
			},
			100,
		},
		{
			"bonus b2b",
			args{
				&userv1.User{
					LpStatusAsString: "test status",
					B2B: &userv1.User_B2B{
						IsB2BState: true,
					},
				},
				&productv1.ProductPriceByRegion{
					Bonuses: []*productv1.ProductPriceByRegion_Bonus{
						{
							LoyaltyStatus: &productv1.ProductPriceByRegion_LoyaltyStatus{
								Code: "test status",
							},
							BonusB2B: 100,
						},
					},
				},
			},
			100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateBonus(tt.args.user, tt.args.prices)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestItem_CalculateMaxCount(t *testing.T) {
	type fields struct {
		countMultiplicity int
	}
	type args struct {
		maxAvailable int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "correct",
			fields: fields{
				countMultiplicity: 1,
			},
			args: args{
				maxAvailable: 10,
			},
			want: 10,
		},
		{
			name: "correct count multiplicity",
			fields: fields{
				countMultiplicity: 2,
			},
			args: args{
				maxAvailable: 15,
			},
			want: 30,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Item{
				countMultiplicity: tt.fields.countMultiplicity,
			}
			assert.Equal(t, tt.want, i.CalculateMaxCount(tt.args.maxAvailable))
		})
	}
}
