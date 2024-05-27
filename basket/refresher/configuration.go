package refresher

import (
	"context"
	"fmt"
	"go.citilink.cloud/citizap"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order/basket"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/specs/domains/catalog_facade"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	facade_productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalogfacade/product/v1"
	"go.uber.org/zap"
)

type configurationItemRefresher struct {
	productClient       productv1.ProductAPIClient
	facadeProductClient facade_productv1.ProductAPIClient
}

func NewConfigurationItemRefresher(
	productClient productv1.ProductAPIClient,
	facadeProductClient facade_productv1.ProductAPIClient,
) *configurationItemRefresher {
	return &configurationItemRefresher{
		productClient:       productClient,
		facadeProductClient: facadeProductClient,
	}
}

func (c *configurationItemRefresher) Refreshable(item *basket_item.Item) bool {
	return item.Type() == basket_item.TypeConfiguration
}

// Refresh обновляет конфигурацию в корзине. Учитывая то, что конфигурация у нас является совершенно уникальной
// сущностью, то и обновлять ее приходится не менее уникально
func (c *configurationItemRefresher) Refresh(
	ctx context.Context,
	items []*basket_item.Item,
	bsk basket.RefresherBasket,
	logger *zap.Logger,
) error {
	logger = logger.With(citizap.SpaceId(string(bsk.SpaceId())))

	// На текущий момент у нас в корзине может быть только одна конфигурация, то есть можно сделать и conf := items[0].
	// Но тогда надо будет делать на всякий случай проверку на len(items) > 0, а это уже смотрится не так лаконично
	for _, conf := range items {
		// Учитывая текущее состояние конфигурации, мы можем проверять только товары в конфигурации, услугами и прочим
		// занимается актуалайзер, да это костыль, но пока без него никак.
		children := bsk.Find(basket.Finders.ChildrenOfRecursive(conf))
		productIds := make([]string, 0, len(children))
		productItems := make([]*basket_item.Item, 0, len(children))
		// Максимальное количество конфигураций должно быть не больше максимального количества одного из товаров в её составе
		maxCountConf := basket.MaxCountOfProductItemsInConf
		for _, child := range children {
			// пока можем работать только с позициями с типом товар
			if child.Type() != basket_item.TypeConfigurationProduct {
				continue
			}

			productIds = append(productIds, string(child.ItemId()))
			productItems = append(productItems, child)
		}

		findFullResponse, err := c.productClient.FindFull(
			ctx, &productv1.FindFullRequest{
				Ids:     productIds,
				SpaceId: string(conf.SpaceId()),
			},
		)
		if err != nil {
			return internal.NewCatalogError(
				fmt.Errorf("can't get products from catalog: %w", err),
				conf.SpaceId(),
			)
		}

		// TODO: WEB-58006 на данный момент можно использовать catalog и catalog-facade одновременно, но в последствии
		// необходимо будет сделать переход
		visitor := catalog_facade.NewVisitor(conf.SpaceId(), bsk.User())
		facadeResponse, err := c.facadeProductClient.Filter(ctx, &facade_productv1.FilterRequest{
			Visitor:   visitor.VisitorInfo(),
			Ids:       productIds,
			WithInfo:  true,
			WithStock: true,
		})
		if err != nil {
			return internal.NewCatalogError(
				fmt.Errorf("can't get products from catalog facade: %w", err),
				conf.SpaceId(),
			)
		}

		// Чтобы поиск был O(1)
		itemsDetails := make(map[string]*productv1.FindFullResponse_FullInfo, len(findFullResponse.GetInfos()))
		productsAvailability := make(map[string]bool, len(facadeResponse.GetProducts()))
		productsMaxAvailable := make(map[string]int32, len(facadeResponse.GetProducts()))
		for _, i := range findFullResponse.GetInfos() {
			itemsDetails[i.GetId()] = i
		}
		for _, p := range facadeResponse.GetProducts() {
			if p.GetInfo() != nil {
				productsAvailability[p.GetId()] = p.GetInfo().GetIsAvailable()
			}
			if p.GetStock() != nil {
				productsMaxAvailable[p.GetId()] = p.GetStock().GetMaxAvailableToBuy()
			}
		}

		var notAvailableProductItemIds []basket_item.ItemId
		for _, productItem := range productItems {
			var productInfo *productv1.FindFullResponse_FullInfo
			if info, ok := itemsDetails[string(productItem.ItemId())]; ok {
				productInfo = info
			}

			productItem.SetMovableFromConfiguration(true)

			productItemMaxCount := productItem.CalculateMaxCount(int(productsMaxAvailable[string(productItem.ItemId())]))
			// Определяем максимальное количество конфигурации по меньшему количеству одного из товаров в составе
			if productItemMaxCount < maxCountConf {
				maxCountConf = productItemMaxCount
			}

			// Если мы не нашли этот товар, считаем что позиция недоступна в этом регионе и удаляем ее,
			// конфигурацию разбираем
			if productInfo == nil {
				logger.Info(
					"product can't be founded in catalog. Maybe it's because of changed region. "+
						"Item will be deleted from basket and configuration will be dissassembled",
					citizap.ProductId(string(productItem.ItemId())),
					citizap.SpaceId(string(productItem.SpaceId())),
				)

				err = bsk.Configuration().Disassemble()
				if err != nil {
					return fmt.Errorf("can't dissassemble configuration: %w", err)
				}

				err := bsk.Remove(productItem, false)
				if err != nil {
					return fmt.Errorf("can't remove product: %w", err)
				}

				bsk.AddInfo(
					basket.NewInfo(
						productItem,
						basket_item.NewInfo(
							basket_item.InfoIdPositionRemoved,
							"Позиция удалена в связи с отсутствием в выбранном городе. "+
								"Конфигурация, в которой присутствовала данная позиция, разобрана.",
						),
					),
				)
				continue
			}

			// Если товар найден, но он не в наличии, то так же помечаем позицию, как недоступную к покупке
			isAvailable := productsAvailability[string(productItem.ItemId())]
			if !isAvailable {
				notAvailableProductItemIds = append(notAvailableProductItemIds, productItem.ItemId())
				continue
			}

			// Если товар имеет нулевую цену то он не доступен к покупке
			if productItem.Price() == 0 {
				notAvailableProductItemIds = append(notAvailableProductItemIds, productItem.ItemId())
				continue
			}

			if price, ok := productInfo.GetPrice().GetPrices()[int32(productItem.PriceColumn())]; ok {
				if productItem.Price() != int(price.GetPrice()) {
					if productItem.IsSelected() {
						info := basket_item.NewInfo(
							basket_item.InfoIdPriceChanged,
							"цена на комплектующую конфигурации изменилась",
						)
						info.Additionals().PriceChanged = basket_item.PriceChangedInfoAddition{
							From: productItem.Price(),
							To:   int(price.GetPrice()),
						}
						productItem.AddInfo(info)
					}

					productItem.SetPrice(int(price.GetPrice()))
					productItem.SetBonus(basket_item.CalculateBonus(bsk.User(), productInfo.GetPrice()))
				}
			} else {
				logger.Warn(
					"can't find price of product, but it's available...",
					citizap.ProductId(string(productItem.ItemId())),
					zap.Int("price_column", int(productItem.PriceColumn())),
				)
			}
		}

		// в случае, если какие-то товары стали недоступны, то эту проблему мы добавляем к самой конфигурации, так как
		// только удалением самой конфигурации можно решить данную проблему
		if len(notAvailableProductItemIds) > 0 {
			problem := basket_item.NewProblem(
				basket_item.ProblemProductItemInConfigurationNotAvailable,
				"одна из позиций в конфигурации недоступна",
			)
			problem.Additions().ConfigurationProblemAdditions = basket_item.ConfigurationProblemAdditions{
				NotAvailableProductItemIds: notAvailableProductItemIds,
			}
			conf.AddProblem(problem)
		}

		conf.Rules().SetMaxCount(maxCountConf)
	}

	return nil
}
