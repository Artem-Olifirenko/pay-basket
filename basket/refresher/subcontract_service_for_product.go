package refresher

import (
	"context"
	"fmt"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order/basket"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	productv1client "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	"go.citilink.cloud/store_types"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serviceChecker interface {
	IsAvailableInCity(ctx context.Context, spaceId store_types.SpaceId, kladrId store_types.KladrId, serviceId string, priceCol catalog_types.PriceColumn) (bool, error)
}

type subcontractServiceForProductItemRefresher struct {
	productClient  productv1client.ProductAPIClient
	serviceChecker serviceChecker
}

func NewSubcontractServiceForProductItemRefresher(productClient productv1client.ProductAPIClient, citiesService serviceChecker) *subcontractServiceForProductItemRefresher {
	return &subcontractServiceForProductItemRefresher{productClient: productClient, serviceChecker: citiesService}
}

func (i *subcontractServiceForProductItemRefresher) Refreshable(item *basket_item.Item) bool {
	return item.Type() == basket_item.TypeSubcontractServiceForProduct
}

func (i *subcontractServiceForProductItemRefresher) Refresh(
	ctx context.Context,
	items []*basket_item.Item,
	bsk basket.RefresherBasket,
	_ *zap.Logger,
) error {
	errGroup, ctx := errgroup.WithContext(ctx)
	for _, item := range items {
		item := item
		errGroup.Go(func() error {
			parentItem := bsk.FindOneById(item.ParentUniqId())
			if parentItem == nil {
				return fmt.Errorf("can't find parent item %s of item %s:%s:%s", item.ParentUniqId(), item.UniqId(), item.ItemId(), item.Type())
			}

			response, err := i.productClient.FindServices(ctx, &productv1client.FindServicesRequest{
				ProductId: string(parentItem.ItemId()),
				SpaceId:   string(item.SpaceId()),
			})
			if err != nil {
				st, ok := status.FromError(err)
				if !ok {
					return internal.NewCatalogError(
						fmt.Errorf("can't get error status from find services response: %w", err),
						item.SpaceId(),
					)
				}

				if st.Code() == codes.NotFound {
					item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "товара с услугой нет в наличии"))
					return nil
				}

				return internal.NewCatalogError(
					fmt.Errorf("can't get subcontract services from catalog microservice: %w", err),
					item.SpaceId(),
				)
			}

			subcontractService, ok := response.GetSubcontractServices()[string(item.ItemId())]
			if !ok {
				item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "услуга субподряда не предоставляется"))
				return nil
			}

			price, ok := subcontractService.GetPrices()[int32(item.PriceColumn())]
			if !ok {
				return fmt.Errorf("can't find price with price column %d in subcontract service (id:%s, space:%s)", item.PriceColumn(), item.ItemId(), item.SpaceId())
			}

			if price.Price == 0 {
				item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable,
					"услуга страхования товара недоступна"))
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

				return nil
			}

			applyServiceInfo := item.Additions().GetSubcontractServiceForProduct().GetApplyServiceInfo()
			if applyServiceInfo != nil {
				kladr := applyServiceInfo.GetCityKladrId()
				res, err := i.serviceChecker.IsAvailableInCity(
					ctx,
					item.SpaceId(),
					kladr,
					string(item.ItemId()),
					item.PriceColumn(),
				)
				if err != nil {
					return fmt.Errorf("can't check availability service in city: %w", err)
				}

				if !res {
					item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailableInSelectedCity,
						"услуга субподряда недоступна в выбранном городе"))
				}
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
