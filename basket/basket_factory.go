package basket

import (
	citizap_factory "go.citilink.cloud/citizap/factory"
	database "go.citilink.cloud/libdatabase"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/bonus/bonuses_for_payment"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
)

type BasketFactory struct {
	itemFactory   basket_item.ItemFactory
	productClient productv1.ProductAPIClient
	itemRefresher itemRefresher
	db            database.DB
	*markingOptions
	*subcontractServiceChangeOptions
	bonusesForPaymentsCalculator *bonuses_for_payment.BonusesForPaymentAgent
}

func NewBasketFactory(
	itemFactory basket_item.ItemFactory,
	productClient productv1.ProductAPIClient,
	itemRefresher itemRefresher,
	db database.DB,
	mrkOpts *markingOptions,
	sbcrOpts *subcontractServiceChangeOptions,
	bonusesForPaymentsCalculator *bonuses_for_payment.BonusesForPaymentAgent,
) *BasketFactory {
	return &BasketFactory{
		itemFactory:                     itemFactory,
		productClient:                   productClient,
		itemRefresher:                   itemRefresher,
		db:                              db,
		markingOptions:                  mrkOpts,
		subcontractServiceChangeOptions: sbcrOpts,
		bonusesForPaymentsCalculator:    bonusesForPaymentsCalculator,
	}
}

func (b *BasketFactory) CreateAnon(
	basket *BasketData,
	loggerFactory citizap_factory.Factory,
) *Basket {
	return NewBasket(
		basket, b.itemFactory, nil, b.productClient, b.itemRefresher,
		loggerFactory, b.db,
		b.markingOptions,
		b.subcontractServiceChangeOptions,
		b.bonusesForPaymentsCalculator,
	)
}

func (b *BasketFactory) CreateUser(
	basket *BasketData, user *userv1.User,
	loggerFactory citizap_factory.Factory,
) *Basket {
	return NewBasket(
		basket, b.itemFactory, user, b.productClient, b.itemRefresher,
		loggerFactory, b.db,
		b.markingOptions,
		b.subcontractServiceChangeOptions,
		b.bonusesForPaymentsCalculator,
	)
}
