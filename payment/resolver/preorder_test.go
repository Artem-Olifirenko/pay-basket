package resolver

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/payment"
	"go.citilink.cloud/order/internal/preorder"
	"testing"
)

func TestPreorder_NewPreOrder(t *testing.T) {
	expected := &PreOrder{
		nil,
	}
	got := NewPreOrder(nil)
	assert.Equal(t, expected, got)
}

func TestPreorder_Resolve(t *testing.T) {
	type args struct {
		ctx            context.Context
		resolvedIdsMap payment.ResolvedPaymentIdMap
	}
	type testCase struct {
		name     string
		args     args
		init     func(bskt *order.MockBasket, ordr *order.MockOrder, orderList *MockPreOrderList)
		payments []order.PaymentId
		want     []order.AllowStatus
		err      error
	}
	tests := []testCase{
		{
			name: "success",
			args: args{
				ctx: context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:                   payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCredit:                 payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
					order.PaymentIdCardsOnline:            payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
					order.PaymentIdInstallments:           payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
					order.PaymentIdYandex:                 payment.NewResolvedPaymentId(order.PaymentIdYandex),
					order.PaymentIdTerminalOrCashbox:      payment.NewResolvedPaymentId(order.PaymentIdTerminalOrCashbox),
					order.PaymentIdWebmoney:               payment.NewResolvedPaymentId(order.PaymentIdWebmoney),
					order.PaymentIdTerminalPinpad:         payment.NewResolvedPaymentId(order.PaymentIdTerminalPinpad),
					order.PaymentIdSbp:                    payment.NewResolvedPaymentId(order.PaymentIdSbp),
					order.PaymentIdApplePay:               payment.NewResolvedPaymentId(order.PaymentIdApplePay),
				},
			},
			payments: []order.PaymentId{
				order.PaymentIdCash,
				order.PaymentIdCredit,
				order.PaymentIdCashless,
				order.PaymentIdSberbankBusinessOnline,
				order.PaymentIdCashWithCard,
				order.PaymentIdCardsOnline,
				order.PaymentIdInstallments,
				order.PaymentIdYandex,
				order.PaymentIdTerminalOrCashbox,
				order.PaymentIdWebmoney,
				order.PaymentIdTerminalPinpad,
				order.PaymentIdSbp,
				order.PaymentIdApplePay,
			},
			init: func(bskt *order.MockBasket, ordr *order.MockOrder, orderList *MockPreOrderList) {
				ordr.EXPECT().Basket(gomock.Any()).Return(bskt, nil)

				bskt.EXPECT().All().Return(basket_item.Items{
					basket_item.NewItem(
						"1", basket_item.TypeProduct,
						"name_1", "nil",
						1, 100, 0,
						"0", 0),
					basket_item.NewItem(
						"2", basket_item.TypeProduct,
						"name_2", "nil",
						1, 200, 0,
						"0", 0),
				})
				orderList.EXPECT().FilterByProductIds(gomock.Any(), gomock.Any()).Return(
					preorder.PreOrders{preorder.NewPreOrder(
						1,
						"name_1",
						[]catalog_types.ProductId{"1", "2"},
						[]order.PaymentId{
							order.PaymentIdCash,
							order.PaymentIdCredit,
							order.PaymentIdCashless,
							order.PaymentIdSberbankBusinessOnline,
							order.PaymentIdCashWithCard,
							order.PaymentIdCardsOnline,
							order.PaymentIdInstallments,
							order.PaymentIdYandex,
							order.PaymentIdTerminalOrCashbox,
							order.PaymentIdWebmoney,
							order.PaymentIdTerminalPinpad,
							order.PaymentIdSbp,
							order.PaymentIdApplePay},
						[]preorder.DeliveryId{preorder.DeliveryIdDelivery, preorder.DeliveryIdFast},
					)},
					nil,
				)
			},
			want: []order.AllowStatus{
				order.AllowStatusAllow,
				order.AllowStatusLimited,
				order.AllowStatusAllow,
				order.AllowStatusAllow,
				order.AllowStatusAllow,
				order.AllowStatusAllow,
				order.AllowStatusAllow,
				order.AllowStatusAllow,
				order.AllowStatusAllow,
				order.AllowStatusAllow,
				order.AllowStatusAllow,
				order.AllowStatusAllow,
				order.AllowStatusAllow,
			},
			err: nil,
		},
		{
			name: "basket error",
			args: args{
				ctx:            context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{},
			},
			payments: []order.PaymentId{},
			init: func(bskt *order.MockBasket, ordr *order.MockOrder, orderList *MockPreOrderList) {
				ordr.EXPECT().Basket(gomock.Any()).Return(nil, fmt.Errorf("refresher can refresh only configuration items"))
			},
			want: []order.AllowStatus{},
			err:  fmt.Errorf("refresher can refresh only configuration items"),
		},
		{
			name: "FilterByProductIds error",
			args: args{
				ctx:            context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{},
			},
			payments: []order.PaymentId{},
			init: func(bskt *order.MockBasket, ordr *order.MockOrder, orderList *MockPreOrderList) {
				ordr.EXPECT().Basket(gomock.Any()).Return(bskt, nil)

				bskt.EXPECT().All().Return(basket_item.Items{
					basket_item.NewItem(
						"1", basket_item.TypeProduct,
						"name_1", "nil",
						1, 100, 0,
						"0", 0),
					basket_item.NewItem(
						"2", basket_item.TypeProduct,
						"name_2", "nil",
						1, 200, 0,
						"0", 0),
				})
				orderList.EXPECT().FilterByProductIds(gomock.Any(), gomock.Any()).Return(
					nil,
					fmt.Errorf("error execute db procedure"),
				)
			},
			want: []order.AllowStatus{},
			err:  fmt.Errorf("error execute db procedure"),
		},
		{
			name: "preOrders == 0",
			args: args{
				ctx:            context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{},
			},
			payments: []order.PaymentId{},
			init: func(bskt *order.MockBasket, ordr *order.MockOrder, orderList *MockPreOrderList) {
				ordr.EXPECT().Basket(gomock.Any()).Return(bskt, nil)

				bskt.EXPECT().All().Return(basket_item.Items{
					basket_item.NewItem(
						"1", basket_item.TypeProduct,
						"name_1", "nil",
						1, 100, 0,
						"0", 0),
					basket_item.NewItem(
						"2", basket_item.TypeProduct,
						"name_2", "nil",
						1, 200, 0,
						"0", 0),
				})
				orderList.EXPECT().FilterByProductIds(gomock.Any(), gomock.Any()).Return(
					preorder.PreOrders{},
					nil,
				)
			},
			want: []order.AllowStatus{},
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			bskt := order.NewMockBasket(ctrl)
			preOrderList := NewMockPreOrderList(ctrl)
			ordr := order.NewMockOrder(ctrl)

			tt.init(bskt, ordr, preOrderList)

			r := &PreOrder{preOrderList}
			err := r.Resolve(tt.args.ctx, tt.args.resolvedIdsMap, ordr)
			assert.Equal(t, tt.err, err)
			for i, status := range tt.want {
				assert.Equal(t, status, tt.args.resolvedIdsMap[tt.payments[i]].Status())
			}
		})
	}
}
