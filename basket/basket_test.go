package basket

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	servicev1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/service/v1"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
	productv1mock "go.citilink.cloud/order/internal/specs/grpcclient/mock/citilink/catalog/product/v1"
	"go.citilink.cloud/store_types"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewBasket(t *testing.T) {
	mockStringsContainer := internal.NewMockStringsContainer(gomock.NewController(t))
	args := struct {
		user       *userv1.User
		basketData *BasketData
		mrkOpts    *markingOptions
		subCtrOpts *subcontractServiceChangeOptions
	}{
		user: &userv1.User{
			SpaceId:     "test_space_id",
			PriceColumn: 2,
		},
		basketData: &BasketData{
			spaceId:     "test_basket_id",
			priceColumn: catalog_types.PriceColumnRetail,
		},
		mrkOpts: NewMarkingOptions(false, mockStringsContainer),
		subCtrOpts: NewSubcontractServiceChangeOptions(
			false,
		),
	}
	want := &Basket{
		data:                            args.basketData,
		user:                            args.user,
		markingOptions:                  args.mrkOpts,
		subcontractServiceChangeOptions: args.subCtrOpts,
	}
	want.configuration = NewConfiguration(want, nil, nil)
	got := NewBasket(
		args.basketData,
		nil,
		args.user,
		nil,
		nil,
		nil,
		nil,
		args.mrkOpts,
		args.subCtrOpts,
		nil,
	)
	assert.Equal(t, want, got)
}

func TestBasket_Add(t *testing.T) {
	type args struct {
		ctx             context.Context
		itemId          basket_item.ItemId
		itemType        basket_item.Type
		parentUniqId    basket_item.UniqId
		count           int
		ignoreFairPrice bool
	}
	tests := []struct {
		name    string
		basket  *Basket
		init    func(ctrl *gomock.Controller) *Basket
		args    args
		want    *basket_item.Item
		wantErr error
	}{
		{
			name: "empty itemId argument",
			args: args{},
			init: func(ctrl *gomock.Controller) *Basket {
				return &Basket{}
			},
			wantErr: internal.NewValidationError(errors.New("itemId is empty")),
		},
		{
			name: "invalid itemType argument",
			args: args{
				itemId: "test_itemId",
			},
			init: func(ctrl *gomock.Controller) *Basket {
				return &Basket{}
			},
			wantErr: internal.NewValidationError(errors.New("itemType invalid")),
		},
		{
			name: "invalid count argument",
			args: args{
				itemId:   "test_itemId",
				itemType: basket_item.TypeDigitalService,
			},
			init: func(ctrl *gomock.Controller) *Basket {
				return &Basket{}
			},
			wantErr: internal.NewValidationError(errors.New("count less or equal 0")),
		},
		{
			name: "wrong itemType argument for item",
			args: args{
				itemId:   "test_itemId",
				itemType: basket_item.TypeInsuranceServiceForProduct,
				count:    1,
			},
			init: func(ctrl *gomock.Controller) *Basket {
				return &Basket{}
			},
			wantErr: fmt.Errorf("item 'insurance_service_for_product' must be a child"),
		},
		{
			name: "can't find item in basket data",
			args: args{
				itemId:       "test_itemId",
				itemType:     basket_item.TypeInsuranceServiceForProduct,
				parentUniqId: basket_item.UniqId("1"),
				count:        1,
			},
			init: func(ctrl *gomock.Controller) *Basket {
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{},
					},
				}
			},
			wantErr: internal.NewNotFoundError(errors.New("parent item '1' not found in basket")),
		},
		{
			name: "wrong itemType argument for item",
			args: args{
				itemId:       "test_itemId",
				itemType:     basket_item.TypeInsuranceServiceForProduct,
				parentUniqId: basket_item.UniqId("1"),
				count:        1,
			},
			init: func(ctrl *gomock.Controller) *Basket {
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							basket_item.UniqId("1"): basket_item.NewItem("", basket_item.TypeInsuranceServiceForProduct, "", "", 0, 0, 0, "", 0),
						},
					},
				}
			},
			wantErr: fmt.Errorf("parent item 'insurance_service_for_product' can't have children of type insurance_service_for_product"),
		},
		{
			name: "error basket haven't user",
			args: args{
				itemId:   "test_itemId",
				itemType: basket_item.TypeLiftingService,
				count:    20,
			},
			init: func(ctrl *gomock.Controller) *Basket {
				items := make(map[basket_item.UniqId]*basket_item.Item, 20)
				for i := 0; i <= 20; i++ {
					item := &basket_item.Item{}
					item.SetIsSelected(true)
					items[basket_item.UniqId(fmt.Sprint(i))] = item
				}
				return &Basket{
					data: &BasketData{
						items: items,
					},
				}
			},
			wantErr: internal.NewLogicErrorWithMsg(errors.New("anon user can`t add more than 20 positions"),
				"Вы добавили 20 разных позиций в корзину. Авторизуйтесь или зарегистрируйтесь, чтобы добавить больше"),
		},
		{
			name: "error basket have user but user not B2B",
			args: args{
				itemId:   "test_itemId",
				itemType: basket_item.TypeLiftingService,
				count:    20,
			},
			init: func(ctrl *gomock.Controller) *Basket {
				items := make(map[basket_item.UniqId]*basket_item.Item, 50)
				for i := 0; i <= 50; i++ {
					item := &basket_item.Item{}
					item.SetIsSelected(true)
					items[basket_item.UniqId(fmt.Sprint(i))] = item
				}
				return &Basket{
					user: &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: false,
						},
					},
					data: &BasketData{
						items: items,
					},
				}
			},
			wantErr: internal.NewLogicErrorWithMsg(errors.New("user can`t add more than 50 positions"),
				"Вы добавили 50 разных позиций в корзину. Удалите ненужные позиции и добавьте товар повторно"),
		},
		{
			name: "error basket have B2B but basket include 100 item positions",
			args: args{
				itemId:   "test_itemId",
				itemType: basket_item.TypeLiftingService,
				count:    20,
			},
			init: func(ctrl *gomock.Controller) *Basket {
				items := make(map[basket_item.UniqId]*basket_item.Item, 100)
				for i := 0; i <= 100; i++ {
					item := &basket_item.Item{}
					item.SetIsSelected(true)
					items[basket_item.UniqId(fmt.Sprint(i))] = item
				}
				return &Basket{
					user: &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					},
					data: &BasketData{
						items: items,
					},
				}
			},
			wantErr: internal.NewLogicErrorWithMsg(errors.New("B2B user can`t add more than 100 positions"),
				"Вы добавили 100 разных позиций в корзину. Удалите ненужные позиции и добавьте товар повторно"),
		},
		{
			name: "error create item factory",
			args: args{
				ctx:      context.Background(),
				itemId:   basket_item.ItemId("test_itemId"),
				itemType: basket_item.TypeLiftingService,
				count:    20,
			},
			init: func(ctrl *gomock.Controller) *Basket {
				mockItemFactory := basket_item.NewMockItemFactory(ctrl)
				mockItemFactory.EXPECT().Create(
					context.Background(),
					basket_item.ItemId("test_itemId"),
					store_types.SpaceId("spb"),
					basket_item.TypeLiftingService,
					20,
					nil,
					catalog_types.PriceColumnClub,
					nil,
					false,
				).Return(
					nil,
					errors.New("test_error"),
				).Times(1)
				return &Basket{
					data: &BasketData{
						spaceId:     store_types.SpaceId("spb"),
						items:       map[basket_item.UniqId]*basket_item.Item{},
						priceColumn: catalog_types.PriceColumnClub,
					},
					itemFactory: mockItemFactory,
				}
			},
			wantErr: fmt.Errorf("can't create item with item factory: %w", errors.New("test_error")),
		},
	}
	for _, tt := range tests {
		tt := tt
		ctrl := gomock.NewController(t)
		t.Run(tt.name, func(t *testing.T) {
			b := tt.init(ctrl)
			got, err := b.Add(tt.args.ctx, tt.args.itemId, tt.args.itemType, tt.args.parentUniqId, tt.args.count, tt.args.ignoreFairPrice)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestBasket_AddItem(t *testing.T) {
	type args struct {
		item *basket_item.Item
	}
	tests := []struct {
		name    string
		args    func() args
		basket  func(args *args) *Basket
		want    *basket_item.Item
		wantErr func(args *args) error
	}{
		{
			name: "error to bought item",
			args: func() args {
				item := basket_item.NewItem("", basket_item.TypeProduct, "", "", 0, 0, 0, "", 0)
				productAdditions := &basket_item.ProductItemAdditions{}
				item.Additions().SetProduct(productAdditions)
				productAdditions.SetIsOEM(true)
				return args{
					item: item,
				}
			},
			basket: func(args *args) *Basket {
				return &Basket{}
			},
			wantErr: func(args *args) error {
				return fmt.Errorf("item can't be bought by not b2b user")
			},
		},
		{
			name: "error add item to basket data",
			args: func() args {
				item := basket_item.NewItem("", basket_item.TypeDigitalService, "", "", 0, 0, 0, "", 0)
				parentItem := basket_item.NewItem("", basket_item.TypeProduct, "", "", 0, 0, 0, "", 0)
				parentItem.AddChild(item)
				return args{
					item: item,
				}
			},
			basket: func(args *args) *Basket {
				return &Basket{
					data: &BasketData{
						items: basket_item.ItemMap{},
					},
				}
			},
			wantErr: func(args *args) error {
				return fmt.Errorf("can't add item to basket: %w", fmt.Errorf(
					"can't find parent item of %s by id %s",
					args.item.UniqId(), args.item.ParentUniqId(),
				))
			},
		},
		{
			name: "successful test",
			args: func() args {
				item := basket_item.NewItem("", basket_item.TypeDigitalService, "", "", 0, 0, 0, "", 0)
				return args{
					item: item,
				}
			},
			basket: func(args *args) *Basket {
				return &Basket{
					data: &BasketData{
						items: basket_item.ItemMap{},
					},
				}
			},
			wantErr: func(args *args) error {
				return nil
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args()
			b := tt.basket(&args)
			got, err := b.AddItem(args.item)
			if tt.wantErr(&args) == nil {
				assert.Equal(t, args.item, got)
			} else {
				assert.Equal(t, tt.wantErr(&args), err)
			}
		})
	}
}

func TestBasket_AccruedBonus(t *testing.T) {
	tests := []struct {
		name string
		init func() *Basket
		want int
	}{
		{
			name: "successful test",
			init: func() *Basket {
				return &Basket{
					data: &BasketData{
						priceColumn: catalog_types.PriceColumnRetail,
					},
				}
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := tt.init()
			got := b.AccruedBonus()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_AddInfo(t *testing.T) {
	type fields struct {
		data *BasketData
	}
	type args struct {
		infos []*Info
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func(args *args) *Basket
	}{
		{
			name: "successful test",
			args: args{
				infos: []*Info{
					{
						item: &basket_item.Item{},
						info: basket_item.NewInfo(basket_item.InfoIdPriceChanged, "test_message"),
					},
				},
			},
			fields: fields{
				data: &BasketData{
					items: map[basket_item.UniqId]*basket_item.Item{},
				},
			},
			want: func(args *args) *Basket {
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{},
						infos: args.infos,
					},
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.fields.data,
			}
			b.AddInfo(tt.args.infos...)
			assert.Equal(t, tt.want(&tt.args), b)
		})
	}
}

func TestBasket_All(t *testing.T) {
	type fields struct {
		data *BasketData
	}
	tests := []struct {
		name   string
		fields fields
		want   basket_item.Items
	}{
		{
			name: "successful test",
			fields: fields{
				data: &BasketData{
					items: map[basket_item.UniqId]*basket_item.Item{
						basket_item.UniqId("0"): &basket_item.Item{},
						basket_item.UniqId("1"): &basket_item.Item{},
					},
				},
			},
			want: basket_item.Items{
				&basket_item.Item{},
				&basket_item.Item{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.fields.data,
			}
			got := b.All()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_CancelSimulateProblems(t *testing.T) {
	type fields struct {
		data *BasketData
	}
	type args struct {
		item *basket_item.Item
	}
	tests := []struct {
		name   string
		args   args
		fields func(args *args) fields
		want   func(args *args) *Basket
	}{
		{
			name: "successful test",
			fields: func(args *args) fields {
				return fields{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							"0": args.item,
						},
					},
				}
			},
			args: args{
				item: basket_item.NewItem("", "product", "", "", 0, 0, 0, "", 0),
			},
			want: func(args *args) *Basket {
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							"0": args.item,
						},
					},
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fields := tt.fields(&tt.args)
			b := &Basket{
				data: fields.data,
			}
			b.CancelSimulateProblems()
			assert.Equal(t, tt.want(&tt.args), b)
		})
	}
}

func TestBasket_Clear(t *testing.T) {
	tests := []struct {
		name   string
		basket *Basket
		want   *Basket
	}{
		{
			name: "successful test",
			basket: &Basket{
				data: &BasketData{
					items: map[basket_item.UniqId]*basket_item.Item{
						"0": basket_item.NewItem("", "product", "", "", 0, 0, 0, "", 0),
					},
				},
			},
			want: &Basket{
				data: &BasketData{
					items: map[basket_item.UniqId]*basket_item.Item{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.basket
			b.Clear()
			assert.Equal(t, tt.want, b)
		})
	}
}

func TestBasket_CommitAllInfos(t *testing.T) {
	type fields struct {
		data *BasketData
	}
	tests := []struct {
		name   string
		fields fields
		want   *Basket
	}{
		{
			name: "successful test",
			fields: fields{
				data: &BasketData{
					infos: []*Info{
						NewInfo(
							basket_item.NewItem("", basket_item.TypeDigitalService, "", "", 0, 0, 0, "", 0),
							basket_item.NewInfo(basket_item.InfoIdPriceChanged, "test_message"),
						),
					},
				},
			},
			want: &Basket{
				data: &BasketData{},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.fields.data,
			}
			b.CommitAllInfos()
			assert.Equal(t, tt.want, b)
		})
	}
}

func TestBasket_CommitChanges(t *testing.T) {
	type fields struct {
		data *BasketData
	}
	tests := []struct {
		name   string
		fields fields
		want   func(fingerpring string) *Basket
	}{
		{
			name: "successful test",
			fields: fields{
				data: &BasketData{
					spaceId:     "msk",
					priceColumn: catalog_types.PriceColumnClub,
					items:       map[basket_item.UniqId]*basket_item.Item{},
				},
			},
			want: func(fingerpring string) *Basket {
				return &Basket{
					data: &BasketData{
						spaceId:           "msk",
						priceColumn:       catalog_types.PriceColumnClub,
						items:             map[basket_item.UniqId]*basket_item.Item{},
						commitFingerprint: fingerpring,
					},
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.fields.data,
			}
			b.CommitChanges()
			assert.Equal(t, tt.want(b.Fingerprint()), b)
		})
	}
}

func TestBasket_CommitInfo(t *testing.T) {
	type args struct {
		infoItem *Info
	}
	tests := []struct {
		name string
		args args
		want *Basket
	}{
		{
			name: "successful test",
			args: args{
				infoItem: NewInfo(&basket_item.Item{}, basket_item.NewInfo(basket_item.InfoIdPriceChanged, "test_info")),
			},
			want: &Basket{
				data: &BasketData{
					infos: []*Info{},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: &BasketData{
					infos: []*Info{
						tt.args.infoItem,
					},
				},
			}
			b.CommitInfo(tt.args.infoItem.info.Id())
			assert.Equal(t, tt.want, b)
		})
	}
}

func TestBasket_Configuration(t *testing.T) {
	type args struct {
		configuration *Configuration
	}
	tests := []struct {
		name string
		args args
		want *Configuration
	}{
		{
			name: "successful get configuration",
			args: args{
				configuration: &Configuration{},
			},
			want: &Configuration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				configuration: tt.args.configuration,
			}
			got := b.Configuration()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_Cost(t *testing.T) {
	type args struct {
		data *BasketData
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "successful get cost",
			args: args{
				data: &BasketData{},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.args.data,
			}
			got := b.Cost()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_Count(t *testing.T) {
	type args struct {
		data *BasketData
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "successful get count",
			args: args{
				data: &BasketData{},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.args.data,
			}
			got := b.Count()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_Counts(t *testing.T) {
	type fields struct {
		data *BasketData
	}
	tests := []struct {
		name   string
		fields fields
		want   *Counts
	}{
		{
			name: "successful test",
			fields: fields{
				data: &BasketData{
					items: map[basket_item.UniqId]*basket_item.Item{
						"0": basket_item.NewItem("", basket_item.TypeConfigurationProduct, "", "", 0, 0, 0, "", 0),
						"1": basket_item.NewItem("", basket_item.TypeProduct, "", "", 2, 0, 0, "", 0),
						"2": basket_item.NewItem("", basket_item.TypeConfiguration, "", "", 3, 0, 0, "", 0),
						"3": basket_item.NewItem("", basket_item.TypeDigitalService, "", "", 4, 0, 0, "", 0),
					},
				},
			},
			want: &Counts{
				All:              9,
				AllPositions:     3,
				Products:         2,
				ProductPositions: 1,
				Configurations:   3,
				Services:         4,
				ServicePositions: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.fields.data,
			}
			got := b.Counts()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_Data(t *testing.T) {
	type args struct {
		data *BasketData
	}
	tests := []struct {
		name string
		args args
		want *BasketData
	}{
		{
			name: "successful get basket data",
			args: args{
				data: &BasketData{},
			},
			want: &BasketData{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.args.data,
			}
			got := b.Data()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_Find(t *testing.T) {
	type args struct {
		finder Finder
		items  []*basket_item.Item
	}
	tests := []struct {
		name string
		args args
		init func(items []*basket_item.Item) *Basket
		want func(items []*basket_item.Item) basket_item.Items
	}{
		{
			name: "test to find items in basket data",
			args: args{
				finder: func(items []*basket_item.Item) basket_item.Items {
					var foundedItems []*basket_item.Item
					for _, item := range items {
						if item.Type() == basket_item.TypeDigitalService {
							foundedItems = append(foundedItems, item)
						}
					}
					return foundedItems
				},
				items: []*basket_item.Item{
					basket_item.NewItem("item_id_1", "digital_service", "", "", 0, 0, 0, "", 0),
					basket_item.NewItem("item_id_2", "product", "", "", 0, 0, 0, "", 0),
				},
			},
			init: func(items []*basket_item.Item) *Basket {
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							basket_item.UniqId("0"): items[0],
							basket_item.UniqId("1"): items[1],
						},
					},
				}
			},
			want: func(items []*basket_item.Item) basket_item.Items {
				return []*basket_item.Item{
					items[0],
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := tt.init(tt.args.items)
			got := b.Find(tt.args.finder)
			assert.Equal(t, tt.want(tt.args.items), got)
		})
	}
}

func TestBasket_FindByIds(t *testing.T) {
	type args struct {
		items   []*basket_item.Item
		itemIDs []basket_item.UniqId
	}
	tests := []struct {
		name string
		args args
		init func([]*basket_item.Item) *Basket
		want func([]*basket_item.Item) []*basket_item.Item
	}{
		{
			name: "find items by ids",
			args: args{
				items: []*basket_item.Item{
					basket_item.NewItem("item_id_1", "product", "", "", 0, 0, 0, "", 0),
					basket_item.NewItem("item_id_2", "product", "", "", 0, 0, 0, "", 0),
					basket_item.NewItem("item_id_3", "product", "", "", 0, 0, 0, "", 0),
				},
				itemIDs: []basket_item.UniqId{
					basket_item.UniqId("0"),
					basket_item.UniqId("2"),
				},
			},
			init: func(items []*basket_item.Item) *Basket {
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							"0": items[0],
							"1": items[1],
							"2": items[2],
						},
					},
				}
			},
			want: func(items []*basket_item.Item) []*basket_item.Item {
				return []*basket_item.Item{
					items[0],
					items[2],
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := tt.init(tt.args.items)
			got := b.FindByIds(tt.args.itemIDs...)
			assert.Equal(t, tt.want(tt.args.items), got)
		})
	}
}

func TestBasket_FindOneById(t *testing.T) {
	type args struct {
		item   *basket_item.Item
		itemID basket_item.UniqId
	}
	tests := []struct {
		name string
		args args
		init func(*basket_item.Item) *Basket
		want *basket_item.Item
	}{
		{
			name: "get problems from basket data",
			args: args{
				item:   basket_item.NewItem("", "product", "", "", 0, 0, 0, "", 0),
				itemID: basket_item.UniqId("0"),
			},
			init: func(item *basket_item.Item) *Basket {
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							basket_item.UniqId("0"): item,
						},
					},
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			item := tt.args.item
			b := tt.init(item)
			tt.want = item
			got := b.FindOneById(tt.args.itemID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_HasUserAddedItems(t *testing.T) {
	type args struct {
		data *BasketData
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "successful added items",
			args: args{
				data: &BasketData{
					items: map[basket_item.UniqId]*basket_item.Item{
						basket_item.UniqId("0"): basket_item.NewItem("", "product", "", "", 0, 0, 0, "", 0),
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.args.data,
			}
			got := b.HasUserAddedItems()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_Infos(t *testing.T) {
	type args struct {
		data *BasketData
	}
	tests := []struct {
		name string
		args args
		want []*Info
	}{
		{
			name: "successful get infos",
			args: args{
				data: &BasketData{
					infos: []*Info{},
				},
			},
			want: []*Info{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.args.data,
			}
			got := b.Infos()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_IsAllProductsInStore(t *testing.T) {
	type args struct {
		data *BasketData
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "check all products in store",
			args: args{
				data: &BasketData{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.args.data,
			}
			got := b.IsAllProductsInStore()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_IsChanged(t *testing.T) {
	type args struct {
		data *BasketData
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "check is changed",
			args: args{
				data: &BasketData{
					priceColumn: catalog_types.PriceColumnThree,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.args.data,
			}
			got := b.IsChanged()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_IsUser(t *testing.T) {
	type args struct {
		user *userv1.User
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "check nil user",
			want: false,
		},
		{
			name: "check present user",
			args: args{
				user: &userv1.User{},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				user: tt.args.user,
			}
			got := b.IsUser()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_PriceColumn(t *testing.T) {
	type args struct {
		data *BasketData
	}
	tests := []struct {
		name string
		args args
		want catalog_types.PriceColumn
	}{
		{
			name: "get price column from basket data",
			args: args{
				data: &BasketData{
					priceColumn: catalog_types.PriceColumnThree,
				},
			},
			want: catalog_types.PriceColumnThree,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.args.data,
			}
			got := b.PriceColumn()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_Problems(t *testing.T) {
	type args struct {
		item func() *basket_item.Item
	}
	tests := []struct {
		name string
		args args
		init func(*basket_item.Item) *Basket
		want func(*basket_item.Item) []*Problem
	}{
		{
			name: "get problems from basket data",
			args: args{
				item: func() *basket_item.Item {
					item := basket_item.NewItem("", "product", "", "", 0, 0, 0, "", 0)
					item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "test_message"))
					return item
				},
			},
			init: func(item *basket_item.Item) *Basket {
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							basket_item.UniqId("0"): item,
						},
					},
				}
			},
			want: func(item *basket_item.Item) []*Problem {
				return []*Problem{
					{
						item:    item,
						problem: basket_item.NewProblem(basket_item.ProblemNotAvailable, "test_message"),
					},
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			item := tt.args.item()
			b := tt.init(item)
			got := b.Problems()
			assert.Equal(t, tt.want(item), got)
		})
	}
}

func TestBasket_Refresh(t *testing.T) {
	type args struct {
		ctx             context.Context
		actualizerItems ActualizerItems
		logger          *zap.Logger
		item            *basket_item.Item
		childItem       *basket_item.Item
	}
	tests := []struct {
		name    string
		args    args
		init    func(ctrl *gomock.Controller, args *args) *Basket
		wantErr func(args *args) string
	}{
		{
			name: "basket space does not equal to subcontract change service",
			args: args{
				ctx: context.Background(),
			},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				mockItemRefresher := NewMockitemRefresher(ctrl)
				mockItemRefresher.EXPECT().Refresh(
					args.ctx,
					gomock.Any(),
					args.logger,
				).Return(
					fmt.Errorf("test error"),
				).Times(1)

				return &Basket{
					data: &BasketData{
						spaceId: store_types.SpaceId("msk_cl"),
						items:   map[basket_item.UniqId]*basket_item.Item{},
					},
					user: &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					},
					productClient: productv1mock.NewMockProductAPIClient(ctrl),
					itemRefresher: mockItemRefresher,
					subcontractServiceChangeOptions: &subcontractServiceChangeOptions{
						subcontractServicesChangeEnabled: true,
					},
				}
			},
			wantErr: func(args *args) string {
				return "test error"
			},
		},
		{
			name: "product client error",
			args: args{
				ctx: context.Background(),
			},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				item := basket_item.NewItem("test_item_id", basket_item.TypeSubcontractServiceForProduct, "", "", 0, 0, 0, "msk_cl", 0)
				parentItem := basket_item.NewItem("test_parent_item", basket_item.TypeProduct, "", "", 0, 0, 0, "", 0)
				parentItem.AddChild(item)
				mockProductAPIClient := productv1mock.NewMockProductAPIClient(ctrl)
				mockProductAPIClient.EXPECT().FindServices(
					args.ctx,
					&productv1.FindServicesRequest{
						ProductId: "test_parent_item",
						SpaceId:   "msk_cl",
					},
				).Return(
					nil,
					errors.New("test error"),
				).Times(1)

				return &Basket{
					data: &BasketData{
						spaceId: store_types.SpaceId("msk_cl"),
						items: map[basket_item.UniqId]*basket_item.Item{
							item.UniqId():       item,
							parentItem.UniqId(): parentItem,
						},
					},
					user: &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					},
					productClient: mockProductAPIClient,
					subcontractServiceChangeOptions: &subcontractServiceChangeOptions{
						subcontractServicesChangeEnabled: true,
					},
				}
			},
			wantErr: func(args *args) string {
				return "can't get error status from find services response: test error"
			},
		},
		{
			name: "product client internal error",
			args: args{
				ctx: context.Background(),
			},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				item := basket_item.NewItem("test_item_id", basket_item.TypeSubcontractServiceForProduct, "", "", 0, 0, 0, "msk_cl", 0)
				parentItem := basket_item.NewItem("test_parent_item", basket_item.TypeProduct, "", "", 0, 0, 0, "", 0)
				parentItem.AddChild(item)
				mockProductAPIClient := productv1mock.NewMockProductAPIClient(ctrl)
				mockProductAPIClient.EXPECT().FindServices(
					args.ctx,
					&productv1.FindServicesRequest{
						ProductId: "test_parent_item",
						SpaceId:   "msk_cl",
					},
				).Return(
					nil,
					status.Error(codes.Aborted, "error"),
				).Times(1)

				return &Basket{
					data: &BasketData{
						spaceId: store_types.SpaceId("msk_cl"),
						items: map[basket_item.UniqId]*basket_item.Item{
							item.UniqId():       item,
							parentItem.UniqId(): parentItem,
						},
					},
					user: &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					},
					productClient: mockProductAPIClient,
					subcontractServiceChangeOptions: &subcontractServiceChangeOptions{
						subcontractServicesChangeEnabled: true,
					},
				}
			},
			wantErr: func(args *args) string {
				return fmt.Errorf(
					"can't get subcontract services from catalog microservice: %w",
					status.Error(codes.Aborted, "error"),
				).Error()
			},
		},
		{
			name: "error to add subcontract service",
			args: args{
				ctx:    context.Background(),
				logger: zap.NewNop(),
			},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				mockProductAPIClient := productv1mock.NewMockProductAPIClient(ctrl)
				parentItem := basket_item.NewItem("test_parent_item", basket_item.TypeProduct, "", "", 0, 0, 0, "msk_cl", 0)
				wrongChildItem := basket_item.NewItem("test_child_item", basket_item.TypeSubcontractServiceForProduct, "", "", 0, 0, 0, "msk_cl", 0)
				notFoundItem := basket_item.NewItem("test_child_item", basket_item.TypeSubcontractServiceForProduct, "", "", 0, 0, 0, "msk_cl", 0)
				itemWithoutService := basket_item.NewItem("test_child_item", basket_item.TypeSubcontractServiceForProduct, "", "", 0, 0, 0, "msk_cl", 0)
				itemWithoutSubcontact := basket_item.NewItem("test_child_item", basket_item.TypeSubcontractServiceForProduct, "", "", 0, 0, 0, "msk_cl", 0)
				item := basket_item.NewItem("test_child_item", basket_item.TypeSubcontractServiceForProduct, "", "", 0, 0, 0, "msk_cl", 0)
				lostParentItem := basket_item.NewItem("test_lost_parent_item", basket_item.TypeProduct, "", "", 0, 0, 0, "msk_cl", 0)
				lostParentItem.AddChild(wrongChildItem)
				parentItem.AddChild(notFoundItem)
				parentItem.AddChild(itemWithoutService)
				parentItem.AddChild(itemWithoutSubcontact)
				parentItem.AddChild(item)
				// notFoundItem
				mockProductAPIClient.EXPECT().FindServices(
					args.ctx,
					&productv1.FindServicesRequest{
						ProductId: string(parentItem.ItemId()),
						SpaceId:   "msk_cl",
					},
				).Return(
					nil,
					status.Error(codes.NotFound, "error"),
				).Times(1)
				// itemWithoutService
				mockProductAPIClient.EXPECT().FindServices(
					args.ctx,
					&productv1.FindServicesRequest{
						ProductId: string(parentItem.ItemId()),
						SpaceId:   "msk_cl",
					},
				).Return(
					&productv1.FindServicesResponse{
						SubcontractServices: map[string]*servicev1.SubcontractService{},
					},
					nil,
				).Times(1)
				// itemWithoutSubcontact
				mockProductAPIClient.EXPECT().FindServices(
					args.ctx,
					&productv1.FindServicesRequest{
						ProductId: string(parentItem.ItemId()),
						SpaceId:   "msk_cl",
					},
				).Return(
					&productv1.FindServicesResponse{
						SubcontractServices: map[string]*servicev1.SubcontractService{
							string(itemWithoutSubcontact.ItemId()): &servicev1.SubcontractService{
								CustomerTypes: []servicev1.CustomerType{
									servicev1.CustomerType_CUSTOMER_TYPE_B2C,
								},
								LinkedServiceId: "test_linked_service_id",
							},
						},
					},
					nil,
				).Times(1)
				// item
				mockProductAPIClient.EXPECT().FindServices(
					args.ctx,
					&productv1.FindServicesRequest{
						ProductId: "test_parent_item",
						SpaceId:   "msk_cl",
					},
				).Return(
					&productv1.FindServicesResponse{
						SubcontractServices: map[string]*servicev1.SubcontractService{
							"test_child_item": &servicev1.SubcontractService{
								CustomerTypes: []servicev1.CustomerType{
									servicev1.CustomerType_CUSTOMER_TYPE_B2C,
								},
								LinkedServiceId: "test_linked_service_id",
							},
							"test_linked_service_id": &servicev1.SubcontractService{
								Id: "test_item_id",
								CustomerTypes: []servicev1.CustomerType{
									servicev1.CustomerType_CUSTOMER_TYPE_B2C,
								},
							},
						},
					},
					nil,
				).Times(1)

				return &Basket{
					data: &BasketData{
						spaceId: store_types.SpaceId("msk_cl"),
						items: map[basket_item.UniqId]*basket_item.Item{
							wrongChildItem.UniqId():        wrongChildItem,
							notFoundItem.UniqId():          notFoundItem,
							itemWithoutService.UniqId():    itemWithoutService,
							itemWithoutSubcontact.UniqId(): itemWithoutSubcontact,
							item.UniqId():                  item,
							parentItem.UniqId():            parentItem,
						},
					},
					user: &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					},
					productClient: mockProductAPIClient,
					subcontractServiceChangeOptions: &subcontractServiceChangeOptions{
						subcontractServicesChangeEnabled: true,
					},
				}
			},
			wantErr: func(args *args) string {
				return fmt.Errorf(
					"can't add subcontract service: %w",
					fmt.Errorf(
						"can't add subcontractService to basket: %w",
						internal.NewValidationError(errors.New("count less or equal 0")),
					),
				).Error()
			},
		},
		{
			name: "error in itemRefresher",
			args: args{
				ctx:    context.Background(),
				logger: zap.NewNop(),
				item:   basket_item.NewItem("test_child_item", basket_item.TypeSubcontractServiceForProduct, "", "", 1, 0, 0, "msk_cl", 0),
			},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				mockItemRefresher := NewMockitemRefresher(ctrl)
				mockProductAPIClient := productv1mock.NewMockProductAPIClient(ctrl)
				mockItemFactory := basket_item.NewMockItemFactory(ctrl)
				item := args.item
				wrongItem := basket_item.NewItem("test_child_item", basket_item.TypeConfigurationProduct, "", "", 0, 0, 0, "msk_cl", 0)
				lostParent := basket_item.NewItem("test_lost_parent", basket_item.TypeConfiguration, "", "", 0, 0, 0, "msk_cl", 0)
				childItem := basket_item.NewItem("test_child_item", basket_item.TypeSubcontractServiceForProduct, "", "", 1, 0, 0, "msk_cl", 0)
				parentItem := basket_item.NewItem("test_parent_item", basket_item.TypeProduct, "", "", 0, 0, 0, "msk_cl", 0)
				parentItem.AddChild(item)
				parentItem.AddChild(childItem)
				lostParent.AddChild(wrongItem)
				mockProductAPIClient.EXPECT().FindServices(
					args.ctx,
					&productv1.FindServicesRequest{
						ProductId: string(parentItem.ItemId()),
						SpaceId:   string(item.SpaceId()),
					},
				).Return(
					&productv1.FindServicesResponse{
						SubcontractServices: map[string]*servicev1.SubcontractService{
							string(item.ItemId()): &servicev1.SubcontractService{
								CustomerTypes: []servicev1.CustomerType{
									servicev1.CustomerType_CUSTOMER_TYPE_B2C,
								},
								LinkedServiceId: "test_linked_service_id",
							},
							"test_linked_service_id": &servicev1.SubcontractService{
								Id: "test_child_item",
								CustomerTypes: []servicev1.CustomerType{
									servicev1.CustomerType_CUSTOMER_TYPE_B2C,
								},
							},
						},
					},
					nil,
				).Times(1)
				mockProductAPIClient.EXPECT().FindServices(
					args.ctx,
					&productv1.FindServicesRequest{
						ProductId: string(parentItem.ItemId()),
						SpaceId:   string(childItem.SpaceId()),
					},
				).Return(
					&productv1.FindServicesResponse{
						SubcontractServices: map[string]*servicev1.SubcontractService{
							string(childItem.ItemId()): &servicev1.SubcontractService{
								CustomerTypes: []servicev1.CustomerType{
									servicev1.CustomerType_CUSTOMER_TYPE_B2C,
								},
							},
						},
					},
					nil,
				).Times(1)
				mockItemFactory.EXPECT().Create(
					context.Background(),
					basket_item.ItemId("test_child_item"),
					store_types.SpaceId("msk_cl"),
					basket_item.TypeSubcontractServiceForProduct,
					1,
					parentItem,
					catalog_types.PriceColumnRetail,
					&userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					},
					false,
				).Return(
					basket_item.NewItem(
						"created_test_item_id",
						basket_item.TypeSubcontractServiceForProduct,
						"",
						"",
						1,
						0,
						0,
						"msk_cl",
						0,
					),
					nil,
				).Times(1)
				mockItemRefresher.EXPECT().Refresh(
					args.ctx,
					gomock.Any(),
					args.logger,
				).Return(
					fmt.Errorf("test error"),
				).Times(1)

				return &Basket{
					data: &BasketData{
						spaceId:     store_types.SpaceId("msk_cl"),
						priceColumn: catalog_types.PriceColumnRetail,
						items: map[basket_item.UniqId]*basket_item.Item{
							wrongItem.UniqId():  wrongItem,
							item.UniqId():       item,
							childItem.UniqId():  childItem,
							parentItem.UniqId(): parentItem,
						},
					},
					user: &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					},
					itemRefresher: mockItemRefresher,
					itemFactory:   mockItemFactory,
					productClient: mockProductAPIClient,
					subcontractServiceChangeOptions: &subcontractServiceChangeOptions{
						subcontractServicesChangeEnabled: true,
					},
				}
			},
			wantErr: func(args *args) string {
				return "test error"
			},
		},
		{
			name: "error to fixServiceForConfProduct",
			args: args{
				ctx:    context.Background(),
				logger: zap.NewNop(),
			},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				mockItemRefresher := NewMockitemRefresher(ctrl)
				mockActualizerItems := NewMockActualizerItems(ctrl)
				mockActualizerItem := NewMockActualizerItem(ctrl)
				mockReduceInfo := NewMockReduceInfo(ctrl)
				mockItemRefresher.EXPECT().Refresh(
					args.ctx,
					gomock.Any(),
					args.logger,
				).Return(
					nil,
				).Times(1)
				emptyPresentItem := basket_item.NewItem("test_empty_present_item", basket_item.TypePresent, "", "", 0, 0, 0, "msk_cl", 0)
				emptyProductItem := basket_item.NewItem("test_empty_product_item", basket_item.TypeProduct, "", "", 1, 0, 0, "msk_cl", 0)
				presentItem := basket_item.NewItem("test_present_item", basket_item.TypePresent, "", "", 1, 0, 0, "msk_cl", 0)
				productItem := basket_item.NewItem("test_parent_item", basket_item.TypeProduct, "", "", 1, 0, 0, "msk_cl", 0)
				productItem.AddChild(presentItem)
				// Часть конфигурации
				skippedConfigurationItem := basket_item.NewItem("test_configuration_product_item", basket_item.TypeConfigurationProduct, "", "", 1, 0, 1, "msk_cl", 0)
				serviceConfigurationItem := basket_item.NewItem("test_configuration_product_service_item", basket_item.TypeConfigurationProductService, "", "", 1, 0, 0, "msk_cl", 0)
				configurationAssemblyItem := basket_item.NewItem("test_configuration_assembly_item", basket_item.TypeConfigurationAssemblyService, "", "", 1, 10, 0, "msk_cl", 0)
				assemblyConfigurationItem := basket_item.NewItem("test_assembly_configuration_item", basket_item.TypeConfigurationAssemblyService, "", "", 1, 0, 0, "msk_cl", 0)
				configurationItem := basket_item.NewItem("test_parent_item_configuration", basket_item.TypeConfiguration, "", "", 1, 0, 0, "msk_cl", 0)
				configurationItem.AddChild(emptyPresentItem)
				configurationItem.AddChild(configurationAssemblyItem)
				configurationItem.AddChild(assemblyConfigurationItem)
				mockActualizerItems.EXPECT().FindByItem(emptyProductItem).Return(nil).Times(1)
				mockActualizerItems.EXPECT().FindByItem(presentItem).Return(nil).Times(1)
				mockActualizerItems.EXPECT().FindByItem(productItem).Return(mockActualizerItem).Times(1)
				mockActualizerItem.EXPECT().GetNotExist().Return(true).Times(1)
				mockActualizerItem.EXPECT().ReduceInfo().Return(nil).Times(1)
				// моки для TypeConfigurationProduct
				mockActualizerItems.EXPECT().FindByItem(skippedConfigurationItem).Return(nil).Times(1)
				// моки для TypeConfigurationProductService
				mockActualizerItems.EXPECT().FindByItem(serviceConfigurationItem).Return(mockActualizerItem).Times(1)
				mockActualizerItem.EXPECT().GetName().Return("configuration_product_service").Times(1)
				mockActualizerItem.EXPECT().GetPrice().Return(8).Times(3)
				mockActualizerItem.EXPECT().GetCount().Return(1).Times(1)
				mockActualizerItem.EXPECT().GetNotExist().Return(false).Times(1)
				mockActualizerItem.EXPECT().ReduceInfo().Return(nil).Times(1)
				// моки для TypeConfigurationAssemblyService
				mockActualizerItems.EXPECT().FindByItem(configurationAssemblyItem).Return(mockActualizerItem).Times(1)
				mockActualizerItem.EXPECT().GetName().Return("configuration_assembly_item").Times(1)
				mockActualizerItem.EXPECT().GetPrice().Return(9).Times(3)
				mockActualizerItem.EXPECT().GetCount().Return(1).Times(1)
				mockActualizerItem.EXPECT().GetNotExist().Return(false).Times(1)
				mockActualizerItem.EXPECT().ReduceInfo().Return(mockReduceInfo).Times(1)
				mockReduceInfo.EXPECT().Count().Return(0).Times(2)
				mockReduceInfo.EXPECT().Info().Return("reduce info").Times(1)
				// моки для TypeConfigurationAssemblyService с нулевой ценой
				mockActualizerItems.EXPECT().FindByItem(assemblyConfigurationItem).Return(mockActualizerItem).Times(1)
				mockActualizerItem.EXPECT().GetName().Return("configuration_assembly_item").Times(1)
				mockActualizerItem.EXPECT().GetPrice().Return(9).Times(1)
				mockActualizerItem.EXPECT().GetCount().Return(1).Times(1)
				mockActualizerItem.EXPECT().GetNotExist().Return(false).Times(1)
				mockActualizerItem.EXPECT().ReduceInfo().Return(mockReduceInfo).Times(1)
				mockReduceInfo.EXPECT().Count().Return(0).Times(2)
				mockReduceInfo.EXPECT().Info().Return("reduce info").Times(1)
				// моки для TypeConfiguration
				mockActualizerItems.EXPECT().FindByItem(configurationItem).Return(mockActualizerItem).Times(1)
				mockActualizerItem.EXPECT().GetName().Return("configuration_item").Times(1)
				mockActualizerItem.EXPECT().GetNotExist().Return(false).Times(1)
				mockActualizerItem.EXPECT().ReduceInfo().Return(nil).Times(1)
				mockActualizerItems.EXPECT().FindByType(basket_item.TypePresent).Return(
					[]ActualizerItem{
						mockActualizerItem,
					},
				).Times(1)
				mockActualizerItem.EXPECT().GetItemId().Return(basket_item.ItemId("not_found_item_id")).Times(1)
				mockActualizerItems.EXPECT().FindByType(basket_item.TypeConfigurationProductService).Return(
					[]ActualizerItem{
						mockActualizerItem,
					},
				).Times(1)
				mockActualizerItem.EXPECT().GetParentItemId().Return(basket_item.ItemId("test_actualizer_id")).Times(1)
				args.actualizerItems = mockActualizerItems

				return &Basket{
					data: &BasketData{
						spaceId:     store_types.SpaceId("msk_cl"),
						priceColumn: catalog_types.PriceColumnRetail,
						items: map[basket_item.UniqId]*basket_item.Item{
							emptyPresentItem.UniqId():          emptyPresentItem,
							emptyProductItem.UniqId():          emptyProductItem,
							presentItem.UniqId():               presentItem,
							productItem.UniqId():               productItem,
							serviceConfigurationItem.UniqId():  serviceConfigurationItem,
							skippedConfigurationItem.UniqId():  skippedConfigurationItem,
							configurationAssemblyItem.UniqId(): configurationAssemblyItem,
							assemblyConfigurationItem.UniqId(): assemblyConfigurationItem,
							configurationItem.UniqId():         configurationItem,
						},
					},
					itemRefresher: mockItemRefresher,
					subcontractServiceChangeOptions: &subcontractServiceChangeOptions{
						subcontractServicesChangeEnabled: true,
					},
				}
			},
			wantErr: func(args *args) string {
				return "can't found parent item for configuration service"
			},
		},
		{
			name: "successful test with zero configuration price",
			args: args{
				ctx:    context.Background(),
				logger: zap.NewNop(),
			},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				mockItemRefresher := NewMockitemRefresher(ctrl)
				mockActualizerItems := NewMockActualizerItems(ctrl)
				item := basket_item.NewItem("test_type_configuration", basket_item.TypeConfigurationProduct, "", "", 1, 1, 12, "test_space_id", 0)
				configItem := basket_item.NewItem("test_type_configuration", basket_item.TypeConfiguration, "", "", 1, 0, 0, "test_space_id", 0)
				mockItemRefresher.EXPECT().Refresh(
					args.ctx,
					gomock.Any(),
					args.logger,
				).Return(
					nil,
				).Times(1)
				mockActualizerItems.EXPECT().FindByItem(item).Return(nil).Times(1)
				mockActualizerItems.EXPECT().FindByItem(configItem).Return(nil).Times(1)
				mockActualizerItems.EXPECT().FindByType(basket_item.TypePresent).Return(
					[]ActualizerItem{},
				).Times(1)
				mockActualizerItems.EXPECT().FindByType(basket_item.TypeConfigurationProductService).Return(
					[]ActualizerItem{},
				).Times(1)
				args.actualizerItems = mockActualizerItems

				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							configItem.UniqId(): configItem,
							item.UniqId():       item,
						},
					},
					itemRefresher: mockItemRefresher,
					subcontractServiceChangeOptions: &subcontractServiceChangeOptions{
						subcontractServicesChangeEnabled: false,
					},
				}
			},
			wantErr: func(args *args) string {
				return ""
			},
		},
		{
			name: "successful test without zero configuration price",
			args: args{
				ctx:    context.Background(),
				logger: zap.NewNop(),
			},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				mockItemRefresher := NewMockitemRefresher(ctrl)
				mockActualizerItems := NewMockActualizerItems(ctrl)
				configItem := basket_item.NewItem("test_configuration", basket_item.TypeConfiguration, "", "", 1, 15, 0, "test_space_id", 0)
				mockItemRefresher.EXPECT().Refresh(
					args.ctx,
					gomock.Any(),
					args.logger,
				).Return(
					nil,
				).Times(1)
				mockActualizerItems.EXPECT().FindByItem(configItem).Return(nil).Times(1)
				mockActualizerItems.EXPECT().FindByType(basket_item.TypePresent).Return(
					[]ActualizerItem{},
				).Times(1)
				mockActualizerItems.EXPECT().FindByType(basket_item.TypeConfigurationProductService).Return(
					[]ActualizerItem{},
				).Times(1)
				args.actualizerItems = mockActualizerItems

				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							configItem.UniqId(): configItem,
						},
					},
					itemRefresher: mockItemRefresher,
					subcontractServiceChangeOptions: &subcontractServiceChangeOptions{
						subcontractServicesChangeEnabled: false,
					},
				}
			},
			wantErr: func(args *args) string {
				return ""
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		ctrl := gomock.NewController(t)
		t.Run(tt.name, func(t *testing.T) {
			b := tt.init(ctrl, &tt.args)
			err := b.Refresh(tt.args.ctx, tt.args.actualizerItems, tt.args.logger)
			if tt.wantErr(&tt.args) == "" {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr(&tt.args))
			}
		})
	}
}

func TestBasket_Remove(t *testing.T) {
	type args struct {
		item  *basket_item.Item
		child *basket_item.Item
		force bool
	}
	tests := []struct {
		name    string
		init    func(args *args) *Basket
		args    args
		wantErr func(args *args) error
	}{
		{
			name: "remove configuration",
			args: args{
				item: basket_item.NewItem("config_item", basket_item.TypeConfiguration, "", "", 0, 0, 0, "", 0),
			},
			init: func(args *args) *Basket {
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							args.item.UniqId(): args.item,
						},
					},
				}
			},
			wantErr: func(args *args) error {
				return nil
			},
		},
		{
			name: "error to remove item",
			args: args{
				item: basket_item.NewItem("", basket_item.TypeConfigurationProduct, "", "", 0, 0, 0, "", 0),
			},
			init: func(args *args) *Basket {
				return &Basket{}
			},
			wantErr: func(args *args) error {
				return internal.NewLogicError(fmt.Errorf("item with type %s can't be deleted", args.item.Type()))
			},
		},
		{
			name: "successful remove",
			args: args{
				item: basket_item.NewItem("", basket_item.TypeProduct, "", "", 0, 0, 0, "", 0),
			},
			init: func(args *args) *Basket {
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							args.item.UniqId(): args.item,
						},
					},
				}
			},
			wantErr: func(args *args) error {
				return nil
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := tt.init(&tt.args)
			err := b.Remove(tt.args.item, tt.args.force)
			assert.Equal(t, tt.wantErr(&tt.args), err)
		})
	}
}

func TestBasket_SetSpaceId(t *testing.T) {
	type fields struct {
		data *BasketData
		user *userv1.User
	}
	type args struct {
		spaceId store_types.SpaceId
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "error to change space id in basket",
			fields: fields{
				user: &userv1.User{
					SpaceId: "test_space_ID",
				},
			},
			args: args{
				spaceId: "test_space_id",
			},
			wantErr: fmt.Errorf("can't change spaceId to basket differenct from user"),
		},
		{
			name: "successful test",
			fields: fields{
				data: &BasketData{
					items: map[basket_item.UniqId]*basket_item.Item{},
				},
				user: &userv1.User{
					SpaceId: "test_space_id",
				},
			},
			args: args{
				spaceId: "test_space_id",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.fields.data,
				user: tt.fields.user,
			}
			err := b.SetSpaceId(tt.args.spaceId)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestBasket_SpaceId(t *testing.T) {
	type args struct {
		data *BasketData
	}
	tests := []struct {
		name string
		args args
		want store_types.SpaceId
	}{
		{
			name: "get space id from basket data",
			args: args{
				data: &BasketData{
					spaceId: store_types.SpaceId("test_stage_id"),
				},
			},
			want: store_types.SpaceId("test_stage_id"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.args.data,
			}
			got := b.SpaceId()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_ToXItems(t *testing.T) {
	type args struct {
		data *BasketData
	}
	tests := []struct {
		name string
		args args
		want []*basket_item.XItem
	}{
		{
			name: "get XItems from basket data",
			args: args{
				data: &BasketData{
					items: map[basket_item.UniqId]*basket_item.Item{},
				},
			},
			want: []*basket_item.XItem{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				data: tt.args.data,
			}
			got := b.ToXItems()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_User(t *testing.T) {
	type fields struct {
		user *userv1.User
	}
	tests := []struct {
		name   string
		fields fields
		want   *userv1.User
	}{
		{
			name: "successful test",
			fields: fields{
				user: &userv1.User{},
			},
			want: &userv1.User{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{
				user: tt.fields.user,
			}
			got := b.User()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_fixServiceForConfProduct(t *testing.T) {
	type fields struct {
		data *BasketData
	}
	type args struct {
		actualizerItems ActualizerItems
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		init    func(ctrl *gomock.Controller, args *args) *Basket
		wantErr error
	}{
		{
			name: "empty actualizer items",
			args: args{},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				mockActualizerItems := NewMockActualizerItems(ctrl)
				mockActualizerItems.EXPECT().FindByType(
					basket_item.TypeConfigurationProductService,
				).Return(
					[]ActualizerItem{},
				).Times(1)
				args.actualizerItems = mockActualizerItems
				return &Basket{}
			},
		},
		{
			name: "not found parent parent item",
			args: args{},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				item := basket_item.NewItem("test_item", basket_item.TypeConfigurationProductService, "", "", 0, 0, 0, "", 0)
				mockActualizerItems := NewMockActualizerItems(ctrl)
				mockActualizerItem := NewMockActualizerItem(ctrl)
				mockActualizerItems.EXPECT().FindByType(
					basket_item.TypeConfigurationProductService,
				).Return(
					[]ActualizerItem{
						mockActualizerItem,
					},
				).Times(1)
				mockActualizerItem.EXPECT().GetItemId().Return(basket_item.ItemId("actualizer_item")).Times(1)
				args.actualizerItems = mockActualizerItems
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							item.UniqId(): item,
						},
					},
				}
			},
			wantErr: fmt.Errorf("can't found parent item for configuration service"),
		},
		{
			name: "found actualized item",
			args: args{},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				item := basket_item.NewItem("test_item", basket_item.TypeConfigurationProductService, "", "", 0, 0, 0, "", 0)
				mockActualizerItems := NewMockActualizerItems(ctrl)
				mockActualizerItem := NewMockActualizerItem(ctrl)
				mockActualizerItems.EXPECT().FindByType(
					basket_item.TypeConfigurationProductService,
				).Return(
					[]ActualizerItem{
						mockActualizerItem,
					},
				).Times(1)
				mockActualizerItem.EXPECT().GetItemId().Return(basket_item.ItemId("test_item")).Times(1)
				args.actualizerItems = mockActualizerItems
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							item.UniqId(): item,
						},
					},
				}
			},
		},
		{
			name: "successful test",
			args: args{},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				item := basket_item.NewItem("test_item", basket_item.TypeConfigurationProductService, "", "", 0, 0, 0, "", 0)
				prItem := basket_item.NewItem("test_parent_item", basket_item.TypeConfigurationProduct, "", "", 0, 0, 0, "", 0)
				prItem.AddChild(item)
				mockActualizerItems := NewMockActualizerItems(ctrl)
				mockActualizerItem := NewMockActualizerItem(ctrl)
				mockActualizerItems.EXPECT().FindByType(
					basket_item.TypeConfigurationProductService,
				).Return(
					[]ActualizerItem{
						mockActualizerItem,
					},
				).Times(1)
				mockActualizerItem.EXPECT().GetItemId().Return(basket_item.ItemId("act_id")).Times(1)
				mockActualizerItem.EXPECT().GetParentItemId().Return(basket_item.ItemId("test_parent_item")).Times(1)
				mockActualizerItem.EXPECT().GetItemId().Return(basket_item.ItemId("test_item")).Times(1)
				mockActualizerItem.EXPECT().GetName().Return("test_name").Times(1)
				mockActualizerItem.EXPECT().GetCount().Return(1).Times(1)
				mockActualizerItem.EXPECT().GetPrice().Return(12).Times(1)
				args.actualizerItems = mockActualizerItems
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							prItem.UniqId(): prItem,
							item.UniqId():   item,
						},
						spaceId:     store_types.SpaceId("msk_cl"),
						priceColumn: catalog_types.PriceColumnRetail,
					},
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		ctrl := gomock.NewController(t)
		t.Run(tt.name, func(t *testing.T) {
			b := tt.init(ctrl, &tt.args)
			err := b.fixServiceForConfProduct(tt.args.actualizerItems)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestBasket_subcontractServiceAdd(t *testing.T) {
	type args struct {
		ctx          context.Context
		parentUniqId basket_item.UniqId
		count        int
		service      *servicev1.SubcontractService
		parentItem   *basket_item.Item
		childItem    *basket_item.Item
	}
	tests := []struct {
		name    string
		args    func() args
		init    func(ctrl *gomock.Controller, args *args) *Basket
		wantErr error
	}{
		{
			name: "error to add position to basket",
			args: func() args {
				return args{}
			},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				return &Basket{}
			},
			wantErr: fmt.Errorf(
				"can't add subcontractService to basket: %w",
				internal.NewValidationError(errors.New("itemId is empty")),
			),
		},
		{
			name: "successful test",
			args: func() args {
				parentItem := basket_item.NewItem("parent_item", basket_item.TypeProduct, "", "", 0, 0, 0, "", 0)
				childItem := basket_item.NewItem("child_item", basket_item.TypeSubcontractServiceForProduct, "", "", 0, 0, 0, "", 0)
				parentItem.AddChild(childItem)
				return args{
					ctx:          context.Background(),
					parentUniqId: parentItem.UniqId(),
					count:        1,
					service: &servicev1.SubcontractService{
						Id: string(childItem.ItemId()),
					},
					parentItem: parentItem,
					childItem:  childItem,
				}
			},
			init: func(ctrl *gomock.Controller, args *args) *Basket {
				itemFactory := basket_item.NewMockItemFactory(ctrl)
				itemFactory.EXPECT().Create(
					args.ctx,
					args.childItem.ItemId(),
					store_types.SpaceId("msk_cl"),
					basket_item.TypeSubcontractServiceForProduct,
					1,
					args.parentItem,
					catalog_types.PriceColumnRetail,
					nil,
					false,
				).Return(
					args.childItem,
					nil,
				).Times(1)
				return &Basket{
					data: &BasketData{
						items: map[basket_item.UniqId]*basket_item.Item{
							args.parentItem.UniqId(): args.parentItem,
							args.childItem.UniqId():  args.childItem,
						},
						spaceId:     store_types.SpaceId("msk_cl"),
						priceColumn: catalog_types.PriceColumnRetail,
					},
					itemFactory: itemFactory,
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		ctrl := gomock.NewController(t)
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args()
			b := tt.init(ctrl, &args)
			got, err := b.subcontractServiceAdd(args.ctx, args.parentUniqId, args.count, args.service)
			if tt.wantErr == nil {
				assert.Equal(t, args.childItem, got)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
		})
	}
}

func TestBasket_subcontractServiceAllowedToUser(t *testing.T) {
	type args struct {
		userType servicev1.CustomerType
		service  *servicev1.SubcontractService
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "negative result",
			args: args{
				userType: servicev1.CustomerType_CUSTOMER_TYPE_INVALID,
				service: &servicev1.SubcontractService{
					CustomerTypes: []servicev1.CustomerType{
						servicev1.CustomerType_CUSTOMER_TYPE_B2C,
						servicev1.CustomerType_CUSTOMER_TYPE_B2B,
					},
				},
			},
			want: false,
		},
		{
			name: "positive result",
			args: args{
				userType: servicev1.CustomerType_CUSTOMER_TYPE_B2B,
				service: &servicev1.SubcontractService{
					CustomerTypes: []servicev1.CustomerType{
						servicev1.CustomerType_CUSTOMER_TYPE_B2C,
						servicev1.CustomerType_CUSTOMER_TYPE_B2B,
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Basket{}
			got := b.subcontractServiceAllowedToUser(tt.args.userType, tt.args.service)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBasket_IsMarkingAvailable(t *testing.T) {
	type fields struct {
		data    *BasketData
		mrkOpts func(ctrl *gomock.Controller) *markingOptions
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "no marked products in basket",
			fields: fields{
				data: func() *BasketData {
					item := basket_item.NewItem("", basket_item.TypeProduct, "", "", 0, 0, 0, "", 0)
					item.Additions().SetProduct(&basket_item.ProductItemAdditions{})

					item2 := basket_item.NewItem("", basket_item.TypeConfiguration, "", "", 0, 0, 0, "", 0)
					item.Additions().SetConfiguration(&basket_item.ConfiguratorItemAdditions{})

					return &BasketData{
						spaceId: "msk_cl",
						items: map[basket_item.UniqId]*basket_item.Item{
							item.UniqId():  item,
							item2.UniqId(): item2,
						},
					}
				}(),
				mrkOpts: func(ctrl *gomock.Controller) *markingOptions {
					mrkCitiesContainer := internal.NewMockStringsContainer(ctrl)

					return &markingOptions{
						markingEnabled:         true,
						markingEnabledInCities: mrkCitiesContainer,
					}
				},
			},
			want: false,
		},
		{
			name: "marked products are in basket",
			fields: fields{
				data: func() *BasketData {
					item := basket_item.NewItem("", basket_item.TypeProduct, "", "", 0, 0, 0, "", 0)
					productAdditions := &basket_item.ProductItemAdditions{}
					item.Additions().SetProduct(productAdditions)
					productAdditions.SetIsMarked(true)

					return &BasketData{
						spaceId: "msk_cl",
						items: map[basket_item.UniqId]*basket_item.Item{
							item.UniqId(): item,
						},
					}
				}(),
				mrkOpts: func(ctrl *gomock.Controller) *markingOptions {
					mkrCitiesContainer := internal.NewMockStringsContainer(ctrl)

					return &markingOptions{
						markingEnabled:         true,
						markingEnabledInCities: mkrCitiesContainer,
					}
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		ctrl := gomock.NewController(t)

		t.Run(tt.name, func(t *testing.T) {
			mrkOpts := tt.fields.mrkOpts(ctrl)
			b := &Basket{
				data: tt.fields.data,
				markingOptions: NewMarkingOptions(
					mrkOpts.markingEnabled,
					mrkOpts.markingEnabledInCities,
				),
			}
			assert.Equalf(t, tt.want, b.IsMarkingAvailable(), "IsMarkingAvailable()")
		})
	}
}
