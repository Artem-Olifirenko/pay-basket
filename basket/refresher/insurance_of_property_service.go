package refresher

import (
	"context"
	"fmt"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order/basket"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	servicev1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/service/v1"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type insuranceOfPropertyServiceItemRefresher struct {
	serviceClient servicev1.ServiceAPIClient
}

func NewInsuranceOfPropertyServiceItemRefresher(serviceClient servicev1.ServiceAPIClient) *insuranceOfPropertyServiceItemRefresher {
	return &insuranceOfPropertyServiceItemRefresher{
		serviceClient: serviceClient,
	}
}

func (i *insuranceOfPropertyServiceItemRefresher) Refreshable(item *basket_item.Item) bool {
	return item.Type() == basket_item.TypePropertyInsurance
}

func (i *insuranceOfPropertyServiceItemRefresher) Refresh(
	ctx context.Context,
	items []*basket_item.Item,
	_ basket.RefresherBasket,
	_ *zap.Logger,
) error {
	errGroup, ctx := errgroup.WithContext(ctx)
	for _, item := range items {
		item := item
		errGroup.Go(func() error {
			response, err := i.serviceClient.FindByIdPropertyInsuranceServices(ctx, &servicev1.FindByIdPropertyInsuranceServicesRequest{
				Id:      string(item.ItemId()),
				SpaceId: string(item.SpaceId()),
			})
			if err != nil {
				s, _ := status.FromError(err)
				if s.Code() == codes.NotFound {
					item.AddProblem(basket_item.NewProblem(
						basket_item.ProblemNotAvailable,
						"услуга страхования имущества не предоставляется"))
					return nil
				} else {
					return internal.NewCatalogError(
						fmt.Errorf("can't get property insurance service from catalog microservice: %w", err),
						item.SpaceId(),
					)
				}
			}

			propertyInsuranceService := response.GetService()
			price, ok := propertyInsuranceService.GetPrices()[int32(item.PriceColumn())]
			if !ok {
				return fmt.Errorf("can't find price with price column %d in property insurance service (id:%s, space:%s)", item.PriceColumn(), item.ItemId(), item.SpaceId())
			}

			if price.Price == 0 {
				item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable,
					"услуга страхования имущества недоступна"))
			}

			if item.Price() != int(price.GetPrice()) {
				if item.IsSelected() {
					info := basket_item.NewInfo(basket_item.InfoIdPriceChanged, "цена на услугу изменилась")
					info.Additionals().PriceChanged = basket_item.PriceChangedInfoAddition{
						From: item.Price(),
						To:   int(price.GetPrice()),
					}
					item.AddInfo(info)
				}

				item.SetPrice(int(price.GetPrice()))
			}
			return nil
		})
	}

	err := errGroup.Wait()
	if err != nil {
		return fmt.Errorf("can't wait errgroup: %w", err)
	}
	return nil
}
