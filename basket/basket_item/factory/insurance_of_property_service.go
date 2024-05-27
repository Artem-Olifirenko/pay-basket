package factory

import (
	"context"
	"fmt"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	servicev1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/service/v1"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
	"go.citilink.cloud/store_types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type insuranceOfPropertyServiceItemFactory struct {
	serviceClient servicev1.ServiceAPIClient
}

func NewInsuranceOfPropertyServiceItemFactory(serviceClient servicev1.ServiceAPIClient) *insuranceOfPropertyServiceItemFactory {
	return &insuranceOfPropertyServiceItemFactory{
		serviceClient: serviceClient,
	}
}

func (i *insuranceOfPropertyServiceItemFactory) Creatable(itemType basket_item.Type) bool {
	return itemType == basket_item.TypePropertyInsurance
}

func (i *insuranceOfPropertyServiceItemFactory) Create(
	ctx context.Context,
	itemId basket_item.ItemId,
	spaceId store_types.SpaceId,
	_ basket_item.Type,
	count int,
	_ *basket_item.Item,
	priceColumn catalog_types.PriceColumn,
	_ *userv1.User,
	_ bool,
) (*basket_item.Item, error) {
	response, err := i.serviceClient.FindByIdPropertyInsuranceServices(ctx, &servicev1.FindByIdPropertyInsuranceServicesRequest{
		Id:      string(itemId),
		SpaceId: string(spaceId),
	})
	if err != nil {
		s, _ := status.FromError(err)
		if s.Code() == codes.NotFound {
			return nil, internal.NewCatalogError(
				internal.NewNotFoundError(fmt.Errorf("property insurance service not found")),
				spaceId,
			)
		} else {
			return nil, internal.NewCatalogError(
				fmt.Errorf("can't get property insurance service from catalog microservice: %w", err),
				spaceId,
			)
		}
	}

	propertyInsuranceService := response.GetService()
	price, ok := propertyInsuranceService.GetPrices()[int32(priceColumn)]
	if !ok {
		return nil, fmt.Errorf("can't find price with price column %d in property insurance service (id:%s, space:%s)", priceColumn, itemId, spaceId)
	}

	propertyInsuranceServiceItem := basket_item.NewItem(
		itemId,
		basket_item.TypePropertyInsurance,
		propertyInsuranceService.GetName(),
		"",
		count,
		int(price.GetPrice()),
		0,
		spaceId,
		priceColumn,
	)

	if price.Price == 0 {
		propertyInsuranceServiceItem.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable,
			"услугу купить невозможно"))
	}

	propertyInsuranceServiceItem.Additions().SetService(
		basket_item.NewService(
			propertyInsuranceService.GetIsAvailForCredit(),
			propertyInsuranceService.GetIsAvailForInstallments(),
		),
	)

	err = propertyInsuranceServiceItem.SetCount(1)
	if err != nil {
		return nil, fmt.Errorf("could not set items count to '1'")
	}
	propertyInsuranceServiceItem.Rules().SetMaxCount(1)

	return propertyInsuranceServiceItem, nil
}
