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

func NewDigitalServiceItemFactory(
	serviceClient servicev1.ServiceAPIClient,
) *digitalServiceItemFactory {
	return &digitalServiceItemFactory{serviceClient: serviceClient}
}

type digitalServiceItemFactory struct {
	serviceClient servicev1.ServiceAPIClient
}

func (f *digitalServiceItemFactory) Creatable(itemType basket_item.Type) bool {
	return itemType == basket_item.TypeDigitalService
}

func (f *digitalServiceItemFactory) Create(
	ctx context.Context,
	itemId basket_item.ItemId,
	spaceId store_types.SpaceId,
	_ basket_item.Type,
	count int,
	parentItem *basket_item.Item,
	priceColumn catalog_types.PriceColumn,
	_ *userv1.User,
	_ bool,
) (*basket_item.Item, error) {
	response, err := f.serviceClient.FindByIdDigitalServices(ctx, &servicev1.FindByIdDigitalServicesRequest{
		Id:      string(itemId),
		SpaceId: string(spaceId),
	})
	if err != nil {
		s, _ := status.FromError(err)
		if s.Code() == codes.NotFound {
			return nil, internal.NewCatalogError(
				internal.NewNotFoundError(fmt.Errorf("digital service not found")),
				spaceId,
			)
		} else {
			return nil, internal.NewCatalogError(
				fmt.Errorf("can't get digital services from catalog microservice: %w", err),
				spaceId,
			)
		}
	}

	digitalService := response.Service
	price, ok := digitalService.GetPrices()[int32(priceColumn)]
	if !ok {
		return nil, fmt.Errorf("can't find price with price column %d in digital service (id:%s, space:%s)", priceColumn, itemId, spaceId)
	}

	digitalServiceItem := basket_item.NewItem(
		itemId,
		basket_item.TypeDigitalService,
		digitalService.GetName(),
		"",
		count,
		int(price.GetPrice()),
		0,
		spaceId,
		priceColumn,
	)

	if price.Price == 0 {
		digitalServiceItem.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable,
			"услугу купить невозможно"))
	}

	// цифровая услуга так же может быть прикреплена
	if parentItem != nil {
		err := parentItem.AddChild(digitalServiceItem)
		if err != nil {
			return nil, fmt.Errorf("can't make item '%s' child of item '%s'", digitalServiceItem.Type(), parentItem.Type())
		}
	}

	digitalServiceItem.Additions().SetService(
		basket_item.NewService(
			digitalService.GetIsAvailForCredit(),
			digitalService.GetIsAvailForInstallments(),
		),
	)

	digitalServiceItem.SetCountMultiplicity(1)

	return digitalServiceItem, nil
}
