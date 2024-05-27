package basket_item

//go:generate mockgen -source=type_service.go -destination=./mock/type_service_mock.go -package=basket_item_mock

import (
	"context"
	"fmt"
	"go.citilink.cloud/order/internal"
	servicev1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/service/v1"
	"go.citilink.cloud/store_types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TypeService interface {
	Determine(ctx context.Context, spaceId store_types.SpaceId, serviceId ItemId, itemGroup Group, getToConfiguration bool) (Type, error)
}

// typeService Сервис типов позиции
type typeService struct {
	serviceClient servicev1.ServiceAPIClient
	mapper        *CatalogServicesToDomainMapper
}

func NewTypeServices(serviceClient servicev1.ServiceAPIClient, mapper *CatalogServicesToDomainMapper) *typeService {
	return &typeService{serviceClient: serviceClient, mapper: mapper}
}

// Determine Определяет тип позиции
func (s *typeService) Determine(
	ctx context.Context,
	spaceId store_types.SpaceId,
	serviceId ItemId,
	itemGroup Group,
	getToConfiguration bool,
) (Type, error) {
	itemType := TypeUnknown

	switch itemGroup {
	case GroupProduct:
		// Товар внутри конфигурации
		if getToConfiguration {
			itemType = TypeConfigurationProduct
		} else {
			itemType = TypeProduct
		}
	case GroupConfiguration:
		itemType = TypeConfiguration
		// услуга - получаем информацию о типе, из mic catalog
	case GroupService:
		response, err := s.serviceClient.FindServices(ctx, &servicev1.FindServicesRequest{
			Offset:     0,
			Limit:      1,
			SpaceIds:   []string{string(spaceId)},
			ServiceIds: []string{string(serviceId)},
		})
		if err != nil {
			s, _ := status.FromError(err)
			if s.Code() == codes.NotFound {
				return itemType, internal.NewCatalogError(
					internal.NewNotFoundError(fmt.Errorf("can't find service in catalog microservice")),
					spaceId,
				)
			} else {
				return itemType, internal.NewCatalogError(
					fmt.Errorf("can't get error status from find services response: %w", err),
					spaceId,
				)
			}
		}

		itemType = s.mapper.MapCatalogServicesTypeToBasketItemType(response.GetServices())
	}

	return itemType, nil
}
