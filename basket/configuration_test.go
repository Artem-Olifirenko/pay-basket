package basket

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/catalog_types"
	database "go.citilink.cloud/libdatabase"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	overallv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/overall/v1"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	productmockv1 "go.citilink.cloud/order/internal/specs/grpcclient/mock/citilink/catalog/product/v1"
	"go.citilink.cloud/store_types"
	"testing"
)

func TestConfiguration_Add(t *testing.T) {
	type fields struct {
		bsk           *Basket
		productClient productv1.ProductAPIClient
		db            database.DB
	}
	type args struct {
		ctx                   context.Context
		confId                basket_item.ConfId
		confType              basket_item.ConfType
		assemblyServiceItemId string
		confItems             []*basket_item.ConfItem
	}

	itemsMap := make(map[basket_item.UniqId]*basket_item.Item)
	itemsMap["50"] = basket_item.NewItem(
		"11",
		basket_item.TypeConfiguration,
		"n",
		"image",
		1,
		1,
		0,
		"msk_cl",
		catalog_types.PriceColumn(1),
	)

	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(productApiMock *productmockv1.MockProductAPIClient)
		want    []*basket_item.Item
		err     error
	}{
		{
			name: "product api error",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{spaceId: "msk_cl"},
				},
			},
			args: args{
				ctx:                   context.Background(),
				confId:                basket_item.ConfId("10"),
				confType:              basket_item.ConfType(1),
				assemblyServiceItemId: "11",
				confItems: []*basket_item.ConfItem{
					{
						ProductId: catalog_types.ProductId("100"),
					},
				},
			},
			prepare: func(productApiMock *productmockv1.MockProductAPIClient) {
				productApiMock.EXPECT().FindFull(context.Background(), &productv1.FindFullRequest{
					Ids:     []string{"100"},
					SpaceId: "msk_cl",
				}).Return(nil, errors.New("test error"))
			},
			want: nil,
			err: internal.NewCatalogError(
				fmt.Errorf("can't assemble item's configuration: %w", fmt.Errorf("can't get products from catalog: %w", errors.New("test error"))),
				"msk_cl",
			),
		},
		{
			name: "can't find all products",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{spaceId: "msk_cl"},
				},
			},
			args: args{
				ctx:                   context.Background(),
				confId:                basket_item.ConfId("10"),
				confType:              basket_item.ConfType(1),
				assemblyServiceItemId: "11",
				confItems: []*basket_item.ConfItem{
					{
						ProductId: catalog_types.ProductId("100"),
					},
				},
			},
			prepare: func(productApiMock *productmockv1.MockProductAPIClient) {
				productApiMock.EXPECT().FindFull(context.Background(), &productv1.FindFullRequest{
					Ids:     []string{"100"},
					SpaceId: "msk_cl",
				}).Return(&productv1.FindFullResponse{
					Infos: []*productv1.FindFullResponse_FullInfo{},
				}, nil)
			},
			want: nil,
			err:  fmt.Errorf("can't assemble item's configuration: %w", fmt.Errorf("can't find all products")),
		},
		{
			name: "can't get price for product",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
					},
				},
			},
			args: args{
				ctx:                   context.Background(),
				confId:                basket_item.ConfId("10"),
				confType:              basket_item.ConfType(1),
				assemblyServiceItemId: "11",
				confItems: []*basket_item.ConfItem{
					{
						ProductId: catalog_types.ProductId("100"),
					},
				},
			},
			prepare: func(productApiMock *productmockv1.MockProductAPIClient) {
				productApiMock.EXPECT().FindFull(context.Background(), &productv1.FindFullRequest{
					Ids:     []string{"100"},
					SpaceId: "msk_cl",
				}).Return(&productv1.FindFullResponse{
					Infos: []*productv1.FindFullResponse_FullInfo{
						{
							Id: "100",
						},
					},
				}, nil)
			},
			want: nil,
			err: fmt.Errorf("can't assemble item's configuration: %w", fmt.Errorf("can't get price for product '%s', with price column '%d'",
				"100", int32(1))),
		},
		{
			name: "success",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
						items:       itemsMap,
					},
				},
			},
			args: args{
				ctx:                   context.Background(),
				confId:                basket_item.ConfId("10"),
				confType:              basket_item.ConfType(1),
				assemblyServiceItemId: "11",
				confItems: []*basket_item.ConfItem{
					{
						ProductId: catalog_types.ProductId("100"),
						Services:  []*basket_item.ConfItemService{{ItemId: "id"}},
					},
				},
			},
			prepare: func(productApiMock *productmockv1.MockProductAPIClient) {
				priceMap := make(map[int32]*overallv1.Price)
				priceMap[1] = &overallv1.Price{
					Column: overallv1.PriceColumn_PRICE_COLUMN_RETAIL,
					Price:  500,
				}

				productApiMock.EXPECT().FindFull(context.Background(), &productv1.FindFullRequest{
					Ids:     []string{"100"},
					SpaceId: "msk_cl",
				}).Return(&productv1.FindFullResponse{
					Infos: []*productv1.FindFullResponse_FullInfo{
						{
							Id: "100",
							Price: &productv1.ProductPriceByRegion{
								ProductId: "100",
								Prices:    priceMap,
							},
							Regional: &productv1.ProductRegional{
								CreditPrograms: []string{"1"},
							},
						},
					},
				}, nil)
			},
			want: []*basket_item.Item{
				basket_item.NewItem(basket_item.ConfSpecialItemId,
					basket_item.TypeConfiguration, "", "", 1, 0, 0, "msk_cl",
					catalog_types.PriceColumn(1)),
				basket_item.NewItem("100",
					basket_item.TypeConfigurationProduct, "", "", 1, 0, 0, "msk_cl",
					catalog_types.PriceColumn(1)),
				basket_item.NewItem("11",
					basket_item.TypeConfigurationAssemblyService, "", "", 1, 0, 0, "msk_cl",
					catalog_types.PriceColumn(1)),
				basket_item.NewItem("11",
					basket_item.TypeConfigurationProductService, "", "", 1, 0, 0, "msk_cl",
					catalog_types.PriceColumn(1)),
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		ctrl := gomock.NewController(t)

		db, _, err := sqlmock.New()
		if err != nil {
			t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
		}

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		productApiMock := productmockv1.NewMockProductAPIClient(ctrl)
		tt.prepare(productApiMock)

		t.Run(tt.name, func(t *testing.T) {
			conf := &Configuration{
				basket:        tt.fields.bsk,
				productClient: productApiMock,
				db:            sqlxDB,
			}

			got, err := conf.Add(tt.args.ctx, tt.args.confId, tt.args.confType, tt.args.assemblyServiceItemId, tt.args.confItems)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			}

			if tt.want != nil {
				assert.Equal(t, len(got), len(tt.want))
			}
		})
		db.Close()
	}
}

func TestConfiguration_MoveItemFrom(t *testing.T) {
	type mocks struct {
		productApiMock  *productmockv1.MockProductAPIClient
		itemFactoryMock *basket_item.MockItemFactory
	}

	type fields struct {
		bsk           *Basket
		productClient productv1.ProductAPIClient
		db            database.DB
	}

	tests := []struct {
		name        string
		fields      func(itemFactoryMock *basket_item.MockItemFactory) fields
		prepare     func(m *mocks)
		basketItems func() map[basket_item.UniqId]*basket_item.Item
		wantErr     bool
		err         error
		mocks       func(ctrl *gomock.Controller) *mocks
	}{
		{
			name: "can't move item out",
			fields: func(itemFactoryMock *basket_item.MockItemFactory) fields {
				return fields{
					bsk: &Basket{
						data: &BasketData{
							spaceId:     "msk_cl",
							priceColumn: catalog_types.PriceColumnRetail,
						},
					},
				}
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsMapConfNotMutable := make(map[basket_item.UniqId]*basket_item.Item)
				itemsMapConfNotMutable["1"] = basket_item.NewItem(
					"11", basket_item.TypeConfiguration, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)

				return itemsMapConfNotMutable
			},
			err: internal.NewLogicErrorWithMsg(errors.New("configuration is template and can't move item out"),
				"Из шаблонной конфигурации нельзя удалять комплектующие."),

			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					productApiMock:  productmockv1.NewMockProductAPIClient(ctrl),
					itemFactoryMock: basket_item.NewMockItemFactory(ctrl),
				}
			},
		},
		{
			name: "no configuration in basket",
			fields: func(itemFactoryMock *basket_item.MockItemFactory) fields {
				return fields{
					bsk: &Basket{
						data: &BasketData{
							spaceId:     "msk_cl",
							priceColumn: catalog_types.PriceColumnRetail,
						},
					},
				}
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsTemplateMap := make(map[basket_item.UniqId]*basket_item.Item)
				itemsTemplateMap["1"] = basket_item.NewItem(
					"11", "t", "", "", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)

				itemsTemplateMap["1"].SetMovableFromConfiguration(true)

				return itemsTemplateMap
			},
			err: fmt.Errorf("no configuration in basket"),
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					productApiMock:  productmockv1.NewMockProductAPIClient(ctrl),
					itemFactoryMock: basket_item.NewMockItemFactory(ctrl),
				}
			},
		},
		{
			name: "items not found",
			fields: func(itemFactoryMock *basket_item.MockItemFactory) fields {
				return fields{
					bsk: &Basket{
						data: &BasketData{
							spaceId:     "msk_cl",
							priceColumn: catalog_types.PriceColumnRetail,
						},
					},
				}
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsMapConf := make(map[basket_item.UniqId]*basket_item.Item)
				item := basket_item.NewItem(
					"11", basket_item.TypeConfiguration, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)
				item.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})
				item.SetMovableFromConfiguration(true)
				itemsMapConf["1"] = item

				return itemsMapConf
			},
			wantErr: true,
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					productApiMock:  productmockv1.NewMockProductAPIClient(ctrl),
					itemFactoryMock: basket_item.NewMockItemFactory(ctrl),
				}
			},
		},
		{
			name: "can't move product from configuration to basket",
			fields: func(itemFactoryMock *basket_item.MockItemFactory) fields {
				return fields{
					bsk: &Basket{
						data: &BasketData{
							spaceId:     "msk_cl",
							priceColumn: catalog_types.PriceColumnRetail,
						},
						itemFactory: itemFactoryMock,
					},
				}
			},
			prepare: func(m *mocks) {
				m.itemFactoryMock.EXPECT().Create(
					context.Background(),
					basket_item.ItemId("11"),
					store_types.SpaceId("msk_cl"),
					basket_item.TypeProduct,
					1,
					gomock.Any(),
					catalog_types.PriceColumnRetail,
					nil,
					false,
				).Return(nil, errors.New("test error"))
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsMapProduct := make(map[basket_item.UniqId]*basket_item.Item)
				itemProduct := basket_item.NewItem(
					"11", basket_item.TypeConfigurationProduct, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)
				itemProduct.SetMovableFromConfiguration(true)

				item := basket_item.NewItem(
					"11", basket_item.TypeConfiguration, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)
				item.SetMovableFromConfiguration(true)
				item.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})

				itemProduct.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})
				itemsMapProduct["1"] = itemProduct
				itemsMapProduct["2"] = item

				return itemsMapProduct
			},
			err: fmt.Errorf("can't move product from configuration to basket: can't create item with item factory: test error"),
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					productApiMock:  productmockv1.NewMockProductAPIClient(ctrl),
					itemFactoryMock: basket_item.NewMockItemFactory(ctrl),
				}
			},
		},
		{
			name: "success",
			fields: func(itemFactoryMock *basket_item.MockItemFactory) fields {
				return fields{
					bsk: &Basket{
						data: &BasketData{
							spaceId:     "msk_cl",
							priceColumn: catalog_types.PriceColumnRetail,
						},
						itemFactory: itemFactoryMock,
					},
				}
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsMapProduct := make(map[basket_item.UniqId]*basket_item.Item)
				itemProduct := basket_item.NewItem(
					"11", basket_item.TypeConfigurationAssemblyService, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)

				item := basket_item.NewItem(
					"11", basket_item.TypeConfiguration, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)
				item.SetMovableFromConfiguration(true)
				item.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})

				itemProduct.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})
				itemsMapProduct["1"] = itemProduct
				itemsMapProduct["2"] = item

				return itemsMapProduct
			},
			err: nil,
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					productApiMock:  productmockv1.NewMockProductAPIClient(ctrl),
					itemFactoryMock: basket_item.NewMockItemFactory(ctrl),
				}
			},
		},
	}
	for _, tt := range tests {
		mocks := tt.mocks(gomock.NewController(t))

		db, _, err := sqlmock.New()
		if err != nil {
			t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
		}

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		if tt.prepare != nil {
			tt.prepare(mocks)
		}

		t.Run(tt.name, func(t *testing.T) {
			conf := &Configuration{
				basket:        tt.fields(mocks.itemFactoryMock).bsk,
				productClient: mocks.productApiMock,
				db:            sqlxDB,
			}
			items := tt.basketItems()
			conf.basket.data.items = items

			err := conf.MoveItemFrom(context.Background(), items["1"].UniqId())
			if tt.wantErr {
				assert.Error(t, err, "")
				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
				}
			}
		})
		db.Close()
	}
}

func TestConfiguration_Disassemble(t *testing.T) {
	type fields struct {
		bsk           *Basket
		productClient productv1.ProductAPIClient
		db            database.DB
	}

	tests := []struct {
		name        string
		fields      fields
		basketItems func() map[basket_item.UniqId]*basket_item.Item
		err         error
	}{
		{
			name: "can't disassemble",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
					},
				},
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsMapConfNotMutable := make(map[basket_item.UniqId]*basket_item.Item)
				itemsMapConfNotMutable["11"] = basket_item.NewItem(
					"11", basket_item.TypeConfiguration, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)

				return itemsMapConfNotMutable
			},
			err: internal.NewLogicErrorWithMsg(errors.New("configuration is template and can't disassemble"),
				"Из шаблонной конфигурации нельзя удалять комплектующие."),
		},
		{
			name: "no configuration in basket",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
					},
				},
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsTemplateMap := make(map[basket_item.UniqId]*basket_item.Item)
				itemsTemplateMap["50"] = basket_item.NewItem(
					"11", "t", "", "", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)

				return itemsTemplateMap
			},
			err: fmt.Errorf("no configuration in basket"),
		},
		{
			name: "success with children",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
					},
				},
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsMapConf := make(map[basket_item.UniqId]*basket_item.Item)
				item := basket_item.NewItem(
					"11", basket_item.TypeConfiguration, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)
				item.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})
				itemsMapConf["1"] = item

				child := basket_item.NewItem(
					"12", basket_item.TypeConfigurationProduct, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)

				_ = item.AddChild(child)
				itemsMapConf["2"] = child

				return itemsMapConf
			},
		},
		{
			name: "success",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
					},
				},
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsMapConf := make(map[basket_item.UniqId]*basket_item.Item)
				item := basket_item.NewItem(
					"11", basket_item.TypeConfiguration, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)
				item.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})
				itemsMapConf["11"] = item

				return itemsMapConf
			},
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)

		db, _, err := sqlmock.New()
		if err != nil {
			t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
		}

		sqlxDB := sqlx.NewDb(db, "sqlmock")

		productApiMock := productmockv1.NewMockProductAPIClient(ctrl)

		t.Run(tt.name, func(t *testing.T) {
			conf := &Configuration{
				basket:        tt.fields.bsk,
				productClient: productApiMock,
				db:            sqlxDB,
			}
			items := tt.basketItems()
			conf.basket.data.items = items

			err := conf.Disassemble()
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
		db.Close()
	}
}

func TestConfiguration_MoveItemIn(t *testing.T) {
	type fields struct {
		bsk           *Basket
		productClient productv1.ProductAPIClient
		db            database.DB
	}
	type args struct {
		uniqId basket_item.UniqId
	}

	tests := []struct {
		name        string
		fields      fields
		args        args
		basketItems func() map[basket_item.UniqId]*basket_item.Item
		wantErr     bool
		err         error
	}{
		{
			name: "can't move item out",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
					},
				},
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsNotMutable := make(map[basket_item.UniqId]*basket_item.Item)
				itemsNotMutable["1"] = basket_item.NewItem(
					"11", basket_item.TypeConfiguration, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)

				return itemsNotMutable
			},
			wantErr: true,
			err: internal.NewLogicErrorWithMsg(errors.New("configuration is template and can't move item in"),
				"В шаблонную конфигурации нельзя добавлять комплектующие."),
		},
		{
			name: "no configuration in basket",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
					},
				},
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsTemplate := make(map[basket_item.UniqId]*basket_item.Item)
				itemsTemplate["1"] = basket_item.NewItem(
					"11", "t", "", "", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)

				return itemsTemplate
			},
			wantErr: true,
			err:     fmt.Errorf("no configuration in basket"),
		},
		{
			name: "item not found",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
					},
				},
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsConf := make(map[basket_item.UniqId]*basket_item.Item)
				item := basket_item.NewItem(
					"1", basket_item.TypeConfiguration, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)
				item.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})
				itemsConf["1"] = item

				return itemsConf
			},
			wantErr: true,
		},
		{
			name: "can't find parent item error",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
					},
				},
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsMapProduct := make(map[basket_item.UniqId]*basket_item.Item)
				itemProduct := basket_item.NewItem(
					"1", basket_item.TypeProduct, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)

				itemProduct.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})
				itemsMapProduct["1"] = itemProduct

				item := basket_item.NewItem(
					"1", basket_item.TypeConfiguration, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)
				item.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})
				itemsMapProduct["2"] = item

				return itemsMapProduct
			},
			wantErr: true,
		},
		{
			name: "max count items in conf",
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
					},
				},
			},
			basketItems: func() map[basket_item.UniqId]*basket_item.Item {
				itemsMapProduct := make(map[basket_item.UniqId]*basket_item.Item)
				itemProduct := basket_item.NewItem(
					"1", basket_item.TypeProduct, "n", "image", 11, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)

				itemProduct.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})
				itemsMapProduct["1"] = itemProduct

				item := basket_item.NewItem(
					"1", basket_item.TypeConfiguration, "n", "image", 1, 1, 0,
					"msk_cl", catalog_types.PriceColumn(1),
				)
				item.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})
				itemsMapProduct["2"] = item

				return itemsMapProduct
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		ctrl := gomock.NewController(t)

		db, _, err := sqlmock.New()
		if err != nil {
			t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
		}
		sqlxDB := sqlx.NewDb(db, "sqlmock")

		productApiMock := productmockv1.NewMockProductAPIClient(ctrl)

		t.Run(tt.name, func(t *testing.T) {
			conf := &Configuration{
				basket:        tt.fields.bsk,
				productClient: productApiMock,
				db:            sqlxDB,
			}
			items := tt.basketItems()
			conf.basket.data.items = items

			err := conf.MoveItemIn(items["1"].UniqId())
			if tt.wantErr {
				assert.Error(t, err, "")
				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
				}
			}
		})
		db.Close()
	}
}

func TestConfiguration_Assemble(t *testing.T) {
	type mocks struct {
		productApiMock *productmockv1.MockProductAPIClient
	}

	type fields struct {
		bsk           *Basket
		productClient productv1.ProductAPIClient
		db            database.DB
	}

	itemsTemplateMap := make(map[basket_item.UniqId]*basket_item.Item)
	itemsTemplateMap["50"] = basket_item.NewItem(
		"11", "t", "", "", 1, 1, 0,
		"msk_cl", catalog_types.PriceColumn(1),
	)

	itemsMapConf := make(map[basket_item.UniqId]*basket_item.Item)
	itemsMapConf["11"] = basket_item.NewItem(
		"11", basket_item.TypeConfiguration, "n", "image", 1, 1, 0,
		"msk_cl", catalog_types.PriceColumn(1),
	)

	itemsProductMap := make(map[basket_item.UniqId]*basket_item.Item)
	item := basket_item.NewItem(
		"11", basket_item.TypeProduct, "", "", 1, 1, 0,
		"msk_cl", catalog_types.PriceColumn(1),
	)
	itemsProductMap["1"] = item

	item.Additions().SetProduct(basket_item.NewProductItemAdditions(
		catalog_types.CategoryId(12), []catalog_types.CreditProgram{""}, 1, 1,
	))

	type args struct {
		ctx                   context.Context
		assemblyServiceItemId string
		count                 int
		confItems             []*basket_item.ConfItem
	}

	tests := []struct {
		name    string
		args    args
		fields  fields
		prepare func(m *mocks, dbMock sqlmock.Sqlmock)
		err     error
		mocks   func(ctrl *gomock.Controller) *mocks
	}{
		{
			name: "configuration already exists",
			args: args{
				ctx:                   context.Background(),
				assemblyServiceItemId: "11",
				confItems: []*basket_item.ConfItem{
					{
						ProductId: catalog_types.ProductId("100"),
						Services:  []*basket_item.ConfItemService{{ItemId: "id"}},
					},
				},
				count: 1,
			},
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
						items:       itemsMapConf,
					},
				},
			},
			err: fmt.Errorf("configuration already exists in basket"),
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					productApiMock: productmockv1.NewMockProductAPIClient(ctrl),
				}
			},
		},
		{
			name: "query error",
			args: args{
				ctx:                   context.Background(),
				assemblyServiceItemId: "11",
				confItems: []*basket_item.ConfItem{
					{
						ProductId: catalog_types.ProductId("100"),
						Services:  []*basket_item.ConfItemService{{ItemId: "id"}},
					},
				},
				count: 1,
			},
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
						items:       itemsTemplateMap,
					},
				},
			},
			prepare: func(m *mocks, dbMock sqlmock.Sqlmock) {
				dbMock.ExpectQuery("Configurator.set_temporary_configuration").
					WithArgs(
						sql.Named("conf_id", basket_item.DefaultConfId),
						sql.Named("item_list", "<items><item><id>11</id><quantity>1</quantity></item></items>")).
					WillReturnError(errors.New("test error"))
			},
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					productApiMock: productmockv1.NewMockProductAPIClient(ctrl),
				}
			},
			err: fmt.Errorf("error on query from db: test error"),
		},
		{
			name: "struct scan error",
			args: args{
				ctx:                   context.Background(),
				assemblyServiceItemId: "11",
				confItems: []*basket_item.ConfItem{
					{
						ProductId: catalog_types.ProductId("100"),
						Services:  []*basket_item.ConfItemService{{ItemId: "id"}},
					},
				},
				count: 1,
			},
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
						items:       itemsTemplateMap,
					},
				},
			},
			prepare: func(m *mocks, dbMock sqlmock.Sqlmock) {
				dbMock.ExpectQuery("Configurator.set_temporary_configuration").
					WithArgs(
						sql.Named("conf_id", basket_item.DefaultConfId),
						sql.Named("item_list", "<items><item><id>11</id><quantity>1</quantity></item></items>")).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"conf_id", "compatible", "assembly_type_id"}).
							AddRow("11", "not bool", 1),
					)
			},
			err: fmt.Errorf("error on scanned row: sql: Scan error on column index 1, name \"compatible\":" +
				" sql/driver: couldn't convert \"not bool\" into type bool"),
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					productApiMock: productmockv1.NewMockProductAPIClient(ctrl),
				}
			},
		},
		{
			name: "no rows",
			args: args{
				ctx:                   context.Background(),
				assemblyServiceItemId: "11",
				confItems: []*basket_item.ConfItem{
					{
						ProductId: catalog_types.ProductId("100"),
						Services:  []*basket_item.ConfItemService{{ItemId: "id"}},
					},
				},
				count: 1,
			},
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
						items:       itemsTemplateMap,
					},
				},
			},
			prepare: func(m *mocks, dbMock sqlmock.Sqlmock) {
				dbMock.ExpectQuery("Configurator.set_temporary_configuration").
					WithArgs(
						sql.Named("conf_id", basket_item.DefaultConfId),
						sql.Named("item_list", "<items><item><id>11</id><quantity>1</quantity></item></items>")).
					WillReturnRows(sqlmock.NewRows([]string{"conf_id", "compatible", "assembly_type_id"}))
			},
			err: nil,
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					// dbMock:         dbMock,
					productApiMock: productmockv1.NewMockProductAPIClient(ctrl),
				}
			},
		},
		{
			name: "invalid configuration",
			args: args{
				ctx:                   context.Background(),
				assemblyServiceItemId: "11",
				confItems: []*basket_item.ConfItem{
					{
						ProductId: catalog_types.ProductId("100"),
						Services:  []*basket_item.ConfItemService{{ItemId: "id"}},
					},
				},
				count: 1,
			},
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
						items:       itemsTemplateMap,
					},
				},
			},
			prepare: func(m *mocks, dbMock sqlmock.Sqlmock) {
				dbMock.ExpectQuery("Configurator.set_temporary_configuration").
					WithArgs(
						sql.Named("conf_id", basket_item.DefaultConfId),
						sql.Named("item_list", "<items><item><id>11</id><quantity>1</quantity></item></items>")).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"conf_id", "compatible", "assembly_type_id"}).
							AddRow("11", false, 1),
					)
			},
			err: fmt.Errorf("invalid configuration"),
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					productApiMock: productmockv1.NewMockProductAPIClient(ctrl),
				}
			},
		},
		{
			name: "can't find all products",
			args: args{
				ctx:                   context.Background(),
				assemblyServiceItemId: "11",
				confItems: []*basket_item.ConfItem{
					{
						ProductId: catalog_types.ProductId("100"),
						Services:  []*basket_item.ConfItemService{{ItemId: "id"}},
					},
				},
				count: 1,
			},
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
						items:       itemsTemplateMap,
					},
				},
			},
			prepare: func(m *mocks, dbMock sqlmock.Sqlmock) {
				dbMock.ExpectQuery("Configurator.set_temporary_configuration").
					WithArgs(
						sql.Named("conf_id", basket_item.DefaultConfId),
						sql.Named("item_list", "<items><item><id>11</id><quantity>1</quantity></item></items>")).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"conf_id", "compatible", "assembly_type_id"}).
							AddRow("11", true, 1),
					)

				m.productApiMock.EXPECT().FindFull(context.Background(), &productv1.FindFullRequest{
					Ids:     []string{"100"},
					SpaceId: "msk_cl",
				}).Return(&productv1.FindFullResponse{
					Infos: []*productv1.FindFullResponse_FullInfo{},
				}, nil)
			},
			err: fmt.Errorf("error on assembling configuration: can't find all products"),
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					productApiMock: productmockv1.NewMockProductAPIClient(ctrl),
				}
			},
		},
		{
			name: "success",
			args: args{
				ctx:                   context.Background(),
				assemblyServiceItemId: "11",
				confItems: []*basket_item.ConfItem{
					{
						ProductId: catalog_types.ProductId("100"),
						Services:  []*basket_item.ConfItemService{{ItemId: "id"}},
					},
				},
				count: 1,
			},
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
						items:       itemsTemplateMap,
					},
				},
			},
			prepare: func(m *mocks, dbMock sqlmock.Sqlmock) {
				priceMap := make(map[int32]*overallv1.Price)
				priceMap[1] = &overallv1.Price{
					Column: overallv1.PriceColumn_PRICE_COLUMN_RETAIL,
					Price:  500,
				}

				dbMock.ExpectQuery("Configurator.set_temporary_configuration").
					WithArgs(
						sql.Named("conf_id", basket_item.DefaultConfId),
						sql.Named("item_list", "<items><item><id>11</id><quantity>1</quantity></item></items>")).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"conf_id", "compatible", "assembly_type_id"}).
							AddRow("11", true, 1),
					)

				m.productApiMock.EXPECT().FindFull(context.Background(), &productv1.FindFullRequest{
					Ids:     []string{"100"},
					SpaceId: "msk_cl",
				}).Return(&productv1.FindFullResponse{
					Infos: []*productv1.FindFullResponse_FullInfo{
						{
							Id: "100",
							Price: &productv1.ProductPriceByRegion{
								ProductId: "100",
								Prices:    priceMap,
							},
							Regional: &productv1.ProductRegional{
								CreditPrograms: []string{"1"},
							},
						},
					},
				}, nil)
			},
			err: nil,
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					productApiMock: productmockv1.NewMockProductAPIClient(ctrl),
				}
			},
		},
		{
			name: "success with changeable items",
			args: args{
				ctx:                   context.Background(),
				assemblyServiceItemId: "11",
				confItems: []*basket_item.ConfItem{
					{
						ProductId: catalog_types.ProductId("11"),
						Services:  []*basket_item.ConfItemService{{ItemId: "id"}},
					},
				},
				count: 1,
			},
			fields: fields{
				bsk: &Basket{
					data: &BasketData{
						spaceId:     "msk_cl",
						priceColumn: catalog_types.PriceColumnRetail,
						items:       itemsProductMap,
					},
				},
			},
			prepare: func(m *mocks, dbMock sqlmock.Sqlmock) {
				priceMap := make(map[int32]*overallv1.Price)
				priceMap[1] = &overallv1.Price{
					Column: overallv1.PriceColumn_PRICE_COLUMN_RETAIL,
					Price:  500,
				}

				dbMock.ExpectQuery("Configurator.set_temporary_configuration").
					WithArgs(
						sql.Named("conf_id", basket_item.DefaultConfId),
						sql.Named("item_list", "<items><item><id>11</id><quantity>1</quantity></item></items>")).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"conf_id", "compatible", "assembly_type_id"}).
							AddRow("11", true, 1),
					)

				m.productApiMock.EXPECT().FindFull(context.Background(), &productv1.FindFullRequest{
					Ids:     []string{"11"},
					SpaceId: "msk_cl",
				}).Return(&productv1.FindFullResponse{
					Infos: []*productv1.FindFullResponse_FullInfo{
						{
							Id: "11",
							Price: &productv1.ProductPriceByRegion{
								ProductId: "11",
								Prices:    priceMap,
							},
							Regional: &productv1.ProductRegional{
								CreditPrograms: []string{"1"},
							},
						},
					},
				}, nil)
			},
			err: nil,
			mocks: func(ctrl *gomock.Controller) *mocks {
				return &mocks{
					productApiMock: productmockv1.NewMockProductAPIClient(ctrl),
				}
			},
		},
	}
	for _, tt := range tests {
		ctrl := gomock.NewController(t)

		db, dbMock, err := sqlmock.New()
		if err != nil {
			t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
		}

		sqlxDB := sqlx.NewDb(db, "sqlmock")

		mocks := tt.mocks(ctrl)
		if tt.prepare != nil {
			tt.prepare(mocks, dbMock)
		}

		t.Run(tt.name, func(t *testing.T) {
			conf := &Configuration{
				basket:        tt.fields.bsk,
				productClient: mocks.productApiMock,
				db:            sqlxDB,
			}

			err := conf.Assemble(tt.args.ctx, tt.args.assemblyServiceItemId, tt.args.confItems, tt.args.count)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
		db.Close()
	}
}
