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
	"testing"
)

func TestCertInCredit_Resolve(t *testing.T) {
	type args struct {
		resolvedIdsMap payment.ResolvedPaymentIdMap
		ordrHandler    func(ctx context.Context, ctrl *gomock.Controller) order.Order
	}

	type paymentMap map[order.PaymentId]order.AllowStatus

	tests := []struct {
		name string
		args args
		want func() (paymentMap, error)
	}{
		{
			// передаём на вход корзину, где есть позиция с 219 категорией, нет флага частичной конфигурации и есть флаг Продукта,
			//    + набор оплат, где есть и кредитный тип оплаты и некредитный тип оплаты,
			//    на выходе должен быть модифицированный набор оплат (т.к. доступ кредитного типа оплаты будет лимитирован)
			"basket contains item with 219 category",
			args{
				payment.ResolvedPaymentIdMap{
					order.PaymentIdCredit:      payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
				},
				func(ctx context.Context, ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					//
					bskt := order.NewMockBasket(ctrl)
					bItem1 := basket_item.NewItem(
						"123",
						basket_item.TypeProduct,
						"",
						"",
						1,
						5,
						2,
						"",
						catalog_types.PriceColumnRetail,
					)

					productAdditional := &basket_item.ProductItemAdditions{}
					bItem1.Additions().SetProduct(productAdditional)
					productAdditional.SetCategoryId(basket_item.GiftCertificateCategoryID)

					bItems := basket_item.Items{
						bItem1,
					}
					bskt.EXPECT().All().Return(bItems).Times(1)
					ordr.EXPECT().Basket(ctx).Return(bskt, nil).Times(1)

					return ordr
				},
			},
			func() (paymentMap, error) {
				pM := make(paymentMap)
				pM[order.PaymentIdCredit] = order.AllowStatusLimited
				pM[order.PaymentIdCardsOnline] = order.AllowStatusAllow
				return pM, nil
			},
		},
		{
			// передаём на вход корзину, где есть позиция с 219 категорией, есть флаг частичной конфигурации и есть флаг Продукта,
			//    + набор оплат, где есть и кредитный тип оплаты и некредитный тип оплаты,
			//    на выходе должен быть НЕмодифицированный набор оплат (т.к. резолвер работает только когда есть позиции без частично конфигурации)
			"item type option isPartOfConfiguration=true",
			args{
				payment.ResolvedPaymentIdMap{
					order.PaymentIdCredit:      payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
				},
				func(ctx context.Context, ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					//
					bskt := order.NewMockBasket(ctrl)
					bItem1 := basket_item.NewItem(
						"123",
						basket_item.TypeConfigurationProductService,
						"",
						"",
						1,
						5,
						2,
						"",
						catalog_types.PriceColumnRetail,
					)

					productAdditions := &basket_item.ProductItemAdditions{}
					bItem1.Additions().SetProduct(productAdditions)
					bItem1.Additions().GetProduct().SetCategoryId(basket_item.GiftCertificateCategoryID)

					bItems := basket_item.Items{
						bItem1,
					}
					bskt.EXPECT().All().Return(bItems).Times(1)
					ordr.EXPECT().Basket(ctx).Return(bskt, nil).Times(1)

					return ordr
				},
			},
			func() (paymentMap, error) {
				pM := make(paymentMap)
				pM[order.PaymentIdCredit] = order.AllowStatusAllow
				pM[order.PaymentIdCardsOnline] = order.AllowStatusAllow
				return pM, nil
			},
		},
		{
			// передаём на вход корзину, где есть позиция с 219 категорией, нет флага частичной конфигурации и нет флаг Продукта,
			//    + набор оплат, где есть и кредитный тип оплаты и некредитный тип оплаты,
			//    на выходе должен быть НЕмодифицированный набор оплат (т.к. резолвер работает только когда есть позиции с флагом продукта)
			"item type option isProduct=false",
			args{
				payment.ResolvedPaymentIdMap{
					order.PaymentIdCredit:      payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
				},
				func(ctx context.Context, ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					//
					bskt := order.NewMockBasket(ctrl)
					bItem1 := basket_item.NewItem(
						"123",
						basket_item.TypePresent,
						"",
						"",
						1,
						5,
						2,
						"",
						catalog_types.PriceColumnRetail,
					)

					productAdditions := &basket_item.ProductItemAdditions{}
					bItem1.Additions().SetProduct(productAdditions)
					bItem1.Additions().GetProduct().SetCategoryId(basket_item.GiftCertificateCategoryID)

					bItems := basket_item.Items{
						bItem1,
					}
					bskt.EXPECT().All().Return(bItems).Times(1)
					ordr.EXPECT().Basket(ctx).Return(bskt, nil).Times(1)

					return ordr
				},
			},
			func() (paymentMap, error) {
				pM := make(paymentMap)
				pM[order.PaymentIdCredit] = order.AllowStatusAllow
				pM[order.PaymentIdCardsOnline] = order.AllowStatusAllow
				return pM, nil
			},
		},
		{
			// передаём на вход корзину, где есть позиция с 219 категорией, нет флага частичной конфигурации и есть флаг Продукта,
			//    + набор оплат, где есть ТОЛЬКО некредитный тип оплаты,
			//    на выходе должен быть НЕмодифицированный набор оплат (т.к. резолвер работает только с кредитными типами оплат)
			"payment map does not contain credit type",
			args{
				payment.ResolvedPaymentIdMap{
					order.PaymentIdApplePay:    payment.NewResolvedPaymentId(order.PaymentIdApplePay),
					order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
				},
				func(ctx context.Context, ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					//
					bskt := order.NewMockBasket(ctrl)
					bItem1 := basket_item.NewItem(
						"123",
						basket_item.TypeProduct,
						"",
						"",
						1,
						5,
						2,
						"",
						catalog_types.PriceColumnRetail,
					)

					productAdditions := &basket_item.ProductItemAdditions{}
					bItem1.Additions().SetProduct(productAdditions)
					bItem1.Additions().GetProduct().SetCategoryId(basket_item.GiftCertificateCategoryID)

					bItems := basket_item.Items{
						bItem1,
					}
					bskt.EXPECT().All().Return(bItems).Times(1)
					ordr.EXPECT().Basket(ctx).Return(bskt, nil).Times(1)

					return ordr
				},
			},
			func() (paymentMap, error) {
				pM := make(paymentMap)
				pM[order.PaymentIdApplePay] = order.AllowStatusAllow
				pM[order.PaymentIdCardsOnline] = order.AllowStatusAllow
				return pM, nil
			},
		},
		{
			// передаём на вход корзину, где НЕТ позиции с 219 категорией, нет флага частичной конфигурации и есть флаг Продукта,
			//    + набор оплат, где есть и кредитный тип оплаты и некредитный тип оплаты,
			//    на выходе будет НЕмодифицированный набор оплат (т.к. данный резолвер работает только с позициями из 219 категории)
			"basket does not contain item with 219 category",
			args{
				payment.ResolvedPaymentIdMap{
					order.PaymentIdCredit:      payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
				},
				func(ctx context.Context, ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					//
					bskt := order.NewMockBasket(ctrl)
					bItem1 := basket_item.NewItem(
						"123",
						basket_item.TypeProduct,
						"",
						"",
						1,
						5,
						2,
						"",
						catalog_types.PriceColumnRetail,
					)

					productAdditions := &basket_item.ProductItemAdditions{}
					bItem1.Additions().SetProduct(productAdditions)
					bItem1.Additions().GetProduct().SetCategoryId(5)

					bItems := basket_item.Items{
						bItem1,
					}
					bskt.EXPECT().All().Return(bItems).Times(1)
					ordr.EXPECT().Basket(ctx).Return(bskt, nil).Times(1)

					return ordr
				},
			},
			func() (paymentMap, error) {
				pM := make(paymentMap)
				pM[order.PaymentIdCredit] = order.AllowStatusAllow
				pM[order.PaymentIdCardsOnline] = order.AllowStatusAllow
				return pM, nil
			},
		},
		{
			// вызов корзины возвращает ошибку
			"basket error",
			args{
				payment.ResolvedPaymentIdMap{},
				func(ctx context.Context, ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().Basket(ctx).Return(nil, fmt.Errorf("some error")).Times(1)
					return ordr
				},
			},
			func() (paymentMap, error) {
				pM := make(paymentMap)
				return pM, fmt.Errorf("some error")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewCertInCredit()
			ctx := context.Background()
			err := resolver.Resolve(ctx, test.args.resolvedIdsMap, test.args.ordrHandler(ctx, gomock.NewController(t)))
			wantPaymentMap, wantErr := test.want()

			gotPaymentMap := make(paymentMap)
			for pmntId, pmnt := range test.args.resolvedIdsMap {
				gotPaymentMap[pmntId] = pmnt.Status()
			}

			if err != nil {
				assert.EqualError(t, err, wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, gotPaymentMap, wantPaymentMap)
			}
		})
	}
}
