package refresher

import (
	"context"
	"fmt"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order/basket"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// insuranceServiceForProductItemRefresher рефрешер для услуг страхования товаров (продление гарантии, кража и т.д.)
type insuranceServiceForProductItemRefresher struct {
	productClient productv1.ProductAPIClient
}

func NewInsuranceServiceForProductItemRefresher(productClient productv1.ProductAPIClient) *insuranceServiceForProductItemRefresher {
	return &insuranceServiceForProductItemRefresher{productClient: productClient}
}

func (i *insuranceServiceForProductItemRefresher) Refreshable(item *basket_item.Item) bool {
	return item.Type() == basket_item.TypeInsuranceServiceForProduct
}

func (i *insuranceServiceForProductItemRefresher) Refresh(
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

			response, err := i.productClient.FindServices(ctx, &productv1.FindServicesRequest{
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
					fmt.Errorf("can't get insurance services from catalog microservice: %w", err),
					item.SpaceId(),
				)
			}

			insuranceService, ok := response.GetInsuranceServices()[string(item.ItemId())]
			if !ok {
				item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "услуга страхования товара не предоставляется"))
				return nil
			}

			avail, ok := insuranceService.GetAvailability()[int32(item.PriceColumn())]
			if !ok || !avail {
				item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "услуга страхования товара не предоставляется"))
				return nil
			}

			price, ok := insuranceService.GetPrices()[int32(item.PriceColumn())]
			if !ok {
				return fmt.Errorf(
					"can't find price with price column %d in insurance service (id:%s, space:%s, product id:%s)",
					item.PriceColumn(),
					item.ItemId(),
					item.SpaceId(),
					parentItem.ItemId(),
				)
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
