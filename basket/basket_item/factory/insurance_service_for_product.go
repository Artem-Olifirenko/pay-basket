package factory

import (
	"context"
	"fmt"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/specs/domains/catalog_facade"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalogfacade/product/v1"
	servicev1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalogfacade/service/v1"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
	"go.citilink.cloud/store_types"
)

type InsuranceServiceForProductItemFactory struct {
	productClient ProductClient
}

func NewInsuranceServiceForProductItemFactory(productClient productv1.ProductAPIClient) *InsuranceServiceForProductItemFactory {
	return &InsuranceServiceForProductItemFactory{productClient: productClient}
}

func (i *InsuranceServiceForProductItemFactory) Creatable(itemType basket_item.Type) bool {
	return itemType == basket_item.TypeInsuranceServiceForProduct
}

func (i *InsuranceServiceForProductItemFactory) Create(
	ctx context.Context,
	itemId basket_item.ItemId,
	spaceId store_types.SpaceId,
	_ basket_item.Type,
	count int,
	parentItem *basket_item.Item,
	priceColumn catalog_types.PriceColumn,
	user *userv1.User,
	_ bool,
) (*basket_item.Item, error) {
	if parentItem == nil {
		return nil, fmt.Errorf("parent item is neccessary for '%s'", basket_item.TypeInsuranceServiceForProduct)
	}

	visitor := catalog_facade.NewVisitor(spaceId, user)
	response, err := i.productClient.FilterServices(
		ctx,
		&productv1.FilterServicesRequest{
			VisitorWithProducts: []*productv1.VisitorWithProducts{
				{
					Ids:     []string{string(parentItem.ItemId())},
					Visitor: visitor.VisitorInfo(),
				},
			},
		})
	if err != nil {
		return nil, internal.NewCatalogError(
			fmt.Errorf("can't get subcontract services from catalog microservice: %w", err),
			spaceId,
		)
	}

	if len(response.GetProductServices()) == 0 {
		return nil, internal.NewNotFoundError(
			fmt.Errorf("insurance services not found for product %s in space %s",
				parentItem.ItemId(),
				parentItem.SpaceId()))
	}

	var service *servicev1.Service
searchServiceLoop:
	for _, productService := range response.GetProductServices()[0].GetServices().GetServicesComposition() {
		for _, variant := range productService.GetVariants() {
			if variant.GetService().GetId() == string(itemId) {
				service = variant.GetService()
				break searchServiceLoop
			}
		}
	}
	if service == nil {
		return nil, internal.NewNotFoundError(
			fmt.Errorf("insurance services not found for product %s in space %s with id %s",
				parentItem.ItemId(),
				parentItem.SpaceId(),
				itemId))
	}

	price := int(service.GetPrice().GetPrice())
	insuranceServiceItem := basket_item.NewItem(
		itemId,
		basket_item.TypeInsuranceServiceForProduct,
		service.GetName(),
		"",
		count,
		price,
		0,
		spaceId,
		priceColumn,
	)

	if price == 0 {
		insuranceServiceItem.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable,
			"услугу купить невозможно"))
	}

	insuranceServiceItem.Additions().SetService(
		basket_item.NewService(
			service.GetIsAvailForCredit(),
			service.GetIsAvailForInstallments(),
		),
	)

	err = parentItem.AddChild(insuranceServiceItem)
	if err != nil {
		return nil, fmt.Errorf("can't make item '%s' child of item '%s'", insuranceServiceItem.Type(), parentItem.Type())
	}

	return insuranceServiceItem, nil
}
