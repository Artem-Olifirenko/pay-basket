package refresher

import (
	"context"
	"fmt"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/citizap"
	"go.citilink.cloud/librecoverer"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/assembly"
	"go.citilink.cloud/order/internal/b2b"
	"go.citilink.cloud/order/internal/metrics"
	"go.citilink.cloud/order/internal/order/basket"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/basket/refresher/mssql"
	"go.citilink.cloud/order/internal/specs/domains/catalog_facade"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	facade_productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalogfacade/product/v1"
	"go.citilink.cloud/user_types"
	"go.uber.org/zap"
	"sync"
)

type productItemRefresher struct {
	productClient           productv1.ProductAPIClient
	facadeProductClient     facade_productv1.ProductAPIClient
	b2b                     *b2b.B2B
	commodityGroupChecker   mssql.AvailabilityChecker
	configurationCalculator assembly.Calculator
	metrics                 *metrics.Metrics
	isMarkingEnabled        bool // Включена ли функция маркировки для продукта
}

func NewProductItemRefresher(
	isMarkingEnabled bool,
	productClient productv1.ProductAPIClient,
	facadeProductClient facade_productv1.ProductAPIClient,
	b2b *b2b.B2B,
	commodityGroupChecker mssql.AvailabilityChecker,
	configurationCalculator assembly.Calculator,
	metrics *metrics.Metrics,
) *productItemRefresher {
	return &productItemRefresher{
		isMarkingEnabled:        isMarkingEnabled,
		productClient:           productClient,
		facadeProductClient:     facadeProductClient,
		b2b:                     b2b,
		commodityGroupChecker:   commodityGroupChecker,
		configurationCalculator: configurationCalculator,
		metrics:                 metrics,
	}
}

func (p *productItemRefresher) Refreshable(item *basket_item.Item) bool {
	return item.Type() == basket_item.TypeProduct
}

func (p *productItemRefresher) Refresh(
	ctx context.Context,
	items []*basket_item.Item,
	bsk basket.RefresherBasket,
	logger *zap.Logger,
) error {
	user := bsk.User()
	var userId string
	if user != nil {
		userId = user.Id
	}

	logger = logger.With(citizap.SpaceId(string(bsk.SpaceId())))

	spaceId := bsk.SpaceId()
	var possibleConfigurationItems []assembly.CheckItem
	var isExistUserConf bool
	selectedItems := bsk.SelectedItems()
	for _, item := range selectedItems {
		// Подготовим список товаров для проверки могут ли они быть конфигурацией.
		// Сюда входят как товары текущей конфигурации, так и товары вне конфигурации с типом "Продукт".
		if item.Type() == basket_item.TypeProduct || item.Type() == basket_item.TypeConfigurationProduct {
			possibleConfigurationItems = append(possibleConfigurationItems, assembly.CheckItem{
				Id:    string(item.ItemId()),
				Type:  item.Type(),
				Count: item.Count(),
			})
		}

		// Определяем есть ли в корзине пользовательская конфигурация
		if item.Type() == basket_item.TypeConfiguration {
			if item.Additions().GetConfiguration().GetConfType() == basket_item.ConfTypeUser {
				isExistUserConf = true
			}
		}

		// Сбрасываем значение, флаг будет определен далее при выполнении требуемых условий
		item.SetMovableToConfiguration(false)

		// Если товар является частью конфигурации,
		// то помечаем товар как возможный для извлечения из конфигурации в корзину.
		if item.Type() == basket_item.TypeConfigurationProduct {
			item.SetMovableFromConfiguration(true)
		}
	}

	wg := sync.WaitGroup{}

	// Если в корзине есть конфигурация и другие товары вне конфигурации, то проверим их совместимость.
	// Отправляем в фоне запрос в API монолита на получение возможной конфигурации из текущих товаров корзины
	// и текущей конфигурации.
	if isExistUserConf && len(selectedItems) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer librecoverer.Default()

			checkData := &assembly.PossibleConfigurationCheckData{
				UserId:  userId,
				SpaceId: spaceId,
				Items:   possibleConfigurationItems,
			}
			possibleConfiguration, possibleConfigurationErr := p.configurationCalculator.GetPossibleConfiguration(checkData)
			if possibleConfigurationErr != nil {
				logger.Info("can't get possible configuration from basket products", zap.Error(possibleConfigurationErr))

				return
			}

			if possibleConfiguration != nil {
				for _, item := range selectedItems {
					for _, confItem := range possibleConfiguration.Items {
						// Возможная (потенциально расширенная) конфигурация включает товары уже находящиеся в текущей конфигурации,
						// поэтому не помечаем такие товары как подходящие для переноса в конфигурацию, они уже в конфигурации.
						if item.Type() != basket_item.TypeConfigurationProduct && string(item.ItemId()) == string(confItem.ProductId) {
							// Флаг MovableToConfiguration проставляем, только если число товаров данной позиции не
							// превышает максимальное значение.
							if confItem.Count <= basket.MaxCountOfProductItemsInConf {
								item.SetMovableToConfiguration(true)
								break
							}
						}
					}
				}
			}
		}()
	}

	// Если в корзине еще нет конфигурации и достаточно товаров для возможной конфигурации, то проверим можно ли из
	// этих товаров потенциально сделать конфигурацию.
	if !isExistUserConf && len(possibleConfigurationItems) >= assembly.MinProductsForPossibleConfiguration {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer librecoverer.Default()

			checkData := &assembly.PossibleConfigurationCheckData{
				UserId:  userId,
				SpaceId: spaceId,
				Items:   possibleConfigurationItems,
			}
			_, possibleConfigurationErr := p.configurationCalculator.GetPossibleConfiguration(checkData)
			if possibleConfigurationErr != nil {
				bsk.SetHasPossibleConfiguration(false)
				logger.Info("can't get possible configuration from basket products", zap.Error(possibleConfigurationErr))

				return
			}

			bsk.SetHasPossibleConfiguration(true)
		}()
	} else {
		bsk.SetHasPossibleConfiguration(false)
	}

	productIds := make([]string, 0, len(items))
	for _, item := range bsk.All() {
		if item.Type() == basket_item.TypeProduct {
			productIds = append(productIds, string(item.ItemId()))
		}
	}

	// Получение детализации по товарам из каталога
	response, err := p.productClient.FindFull(
		ctx, &productv1.FindFullRequest{
			Ids:     productIds,
			SpaceId: string(spaceId),
		},
	)
	if err != nil {
		return internal.NewCatalogError(
			fmt.Errorf("can't get products from catalog microservice: %w", err),
			spaceId,
		)
	}

	// TODO: WEB-58006 на данный момент можно использовать catalog и catalog-facade одновременно, но в последствии
	// необходимо будет сделать переход
	visitor := catalog_facade.NewVisitor(spaceId, bsk.User())
	facadeResponse, err := p.facadeProductClient.Filter(ctx, &facade_productv1.FilterRequest{
		Visitor:   visitor.VisitorInfo(),
		Ids:       productIds,
		WithInfo:  true,
		WithStock: true,
	})
	if err != nil {
		return internal.NewCatalogError(
			fmt.Errorf("can't get products from catalog facade microservices: %w", err),
			spaceId,
		)
	}
	productsAvailability := make(map[string]bool, len(facadeResponse.GetProducts()))
	productsMaxAvailable := make(map[string]int32, len(facadeResponse.GetProducts()))
	for _, p := range facadeResponse.GetProducts() {
		if p.GetInfo() != nil {
			productsAvailability[p.GetId()] = p.GetInfo().GetIsAvailable()
		}
		if p.GetStock() != nil {
			productsMaxAvailable[p.GetId()] = p.GetStock().GetMaxAvailableToBuy()
		}
	}

	// Ищем каждый товар в ответе каталога и, если не нашли, считаем что он недоступен в этом регионе (и удаляем его)
itemsLoop:
	for _, item := range items {
		for _, productInfo := range response.GetInfos() {
			if productInfo.Id == string(item.ItemId()) {
				continue itemsLoop
			}
		}

		logger.Info(
			"product can't be founded in catalog. Maybe it's because of changed region. Item will be deleted from basket",
			citizap.ProductId(string(item.ItemId())),
			zap.String("uniq_id", string(item.UniqId())),
			citizap.SpaceId(string(item.SpaceId())),
		)
		err := bsk.Remove(item, false)
		if err != nil {
			return fmt.Errorf("can't remove product: %w", err)
		}

		bsk.AddInfo(
			basket.NewInfo(
				item,
				basket_item.NewInfo(
					basket_item.InfoIdPositionRemoved,
					"Позиция удалена в связи с отсутствием в выбранном городе",
				),
			),
		)

		return nil
	}

	if user != nil && user.GetB2B() != nil && user.GetB2B().GetIsB2BState() {
		companyId := user.GetB2B().GetState().GetCompanyId()
		contractors, err := p.b2b.FindContractor(
			ctx,
			b2b.CompanyId(companyId),
			user,
			b2b.ContractorId(user.B2B.State.ContractorId),
		)
		if err != nil {
			return fmt.Errorf("error getting contractor: %w", err)
		}

		if len(contractors) == 0 {
			logger.Error(
				"b2b returned no contractors, at least one must be presented",
				zap.String("company_id", companyId),
				zap.String("contractor_id", user.B2B.State.ContractorId),
			)

			return fmt.Errorf("empty contractors list")
		}
		contractor := contractors[0]
		companies, err := p.b2b.Companies(
			ctx,
			user_types.UserId(user.GetId()),
			contractor.CompanyId,
			b2b.ContractorId(contractor.Id),
		)
		if err != nil {
			return fmt.Errorf("error getting companies: %w", err)
		}

		var company *b2b.Company
		for _, c := range companies {
			if c.Id == b2b.CompanyId(companyId) {
				company = c
				break
			}
		}

		if company == nil {
			return fmt.Errorf("can't find company: %s", b2b.CompanyId(companyId))
		}

		for _, productInfo := range response.GetInfos() {
			for _, item := range items {
				if productInfo.Id != string(item.ItemId()) || !productInfo.GetRegional().GetIsMarked() {
					continue
				}

				if p.isMarkingEnabled {
					// Массив идентификаторов товаров, товарную группу которых нужно проверить на доступность перепродажи
					var itemIdsForCommodityGroupCheck []basket_item.ItemId
					for _, item := range selectedItems {
						if catalog_types.ProductId(productInfo.GetId()) == catalog_types.ProductId(item.ItemId()) {
							hasEdoOrGis := company.HasEdo || company.HasGisHonestSign
							productAdditions := item.Additions().GetProduct()
							if productAdditions.MarkedPurchaseReason() == basket_item.MarkedPurchaseReasonForResale && !hasEdoOrGis {
								// если отсутствует ЭДО или ГИС
								item.AddProblem(basket_item.NewProblem(
									basket_item.ProblemPurchaseReasonNotAvailableForUser,
									"отсутствует ЭДО или ГИС",
								))
								return nil
							}

							if hasEdoOrGis {
								// Если цель приобретения товара ещё не выбрана (при добавлении в корзину, например),
								// то будем проверять товарную группу таких товаров на возможность перепродажи.
								if productAdditions.MarkedPurchaseReason() == basket_item.MarkedPurchaseReasonUnknown {
									itemIdsForCommodityGroupCheck = append(itemIdsForCommodityGroupCheck, item.ItemId())
								}
							}
						}
					}

					// Если есть товары для проверки товарных групп на возможность перепродажи
					if len(itemIdsForCommodityGroupCheck) > 0 {
						commodityGroupAvailabilities, err := p.commodityGroupChecker.CheckAvailabilityForResale(
							ctx,
							itemIdsForCommodityGroupCheck,
							contractor.Inn,
						)
						if err != nil {
							return fmt.Errorf("can't check commodity group for availability for resale: %w", err)
						}

						for _, itemId := range itemIdsForCommodityGroupCheck {
							availabilityForResale, ok := commodityGroupAvailabilities[itemId]
							// если товарная группа товара доступна для перепродажи, то указываем это для позиции
							if ok && availabilityForResale.IsAvailabilityForResale {
								item.SetAllowResale(basket_item.NewAllowResale(true, ""))
								continue
							}

							var commodityGroupName string
							if availabilityForResale != nil {
								commodityGroupName = availabilityForResale.Name
							}

							allowResale := basket_item.NewAllowResale(false, commodityGroupName)
							item.SetAllowResale(allowResale)

							// Если цель приобретения для товара выбрана "перепродажа", а товарная группа недоступна
							// для перепродажи, то добавляем еще и проблему.
							if item.Additions().GetProduct().MarkedPurchaseReason() == basket_item.MarkedPurchaseReasonForResale {
								item.AddProblem(basket_item.NewProblem(
									basket_item.ProblemPurchaseReasonNotAvailableForUser, allowResale.Message(),
								))
							}
						}
					}
				} else {
					err := bsk.Remove(item, false)
					if err != nil {
						return fmt.Errorf("can't remove marked product for b2b: %w", err)
					}
					return nil
				}
			}
		}
	}

	for _, productInfo := range response.GetInfos() {
		for _, item := range items {
			if productInfo.Id != string(item.ItemId()) {
				continue
			}

			// Если товар найден, но он не в наличии, то так же помечаем позицию, как недоступную к покупке
			isAvailable := productsAvailability[string(item.ItemId())]
			if !isAvailable {
				item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "товара нет в наличии"))
				continue
			}

			price, ok := productInfo.GetPrice().GetPrices()[int32(catalog_types.PriceColumnRetail)]
			if ok && productInfo.GetPrice().GetIsFairPrice() && item.IgnoreFairPrice() {
				price.Price, productInfo.GetPrice().OldPrice = productInfo.GetPrice().OldPrice, price.Price
			}

			if price, ok := productInfo.GetPrice().GetPrices()[int32(item.PriceColumn())]; ok {
				if item.Price() != int(price.GetPrice()) {
					if item.IsSelected() {
						logger.Info(
							"price of the item has been changed",
							citizap.ProductId(string(item.ItemId())),
							zap.Int("old_price", item.Price()),
							zap.Int("new_price", int(price.GetPrice())),
						)
					}

					if !item.IgnoreFairPriceChanged() && item.IsSelected() {
						info := basket_item.NewInfo(basket_item.InfoIdPriceChanged, "цена на товар изменилась")
						info.Additionals().PriceChanged = basket_item.PriceChangedInfoAddition{
							From: item.Price(),
							To:   int(price.GetPrice()),
						}
						item.AddInfo(info)
					}
					item.SetPrice(int(price.GetPrice()))
				}
				item.SetBonus(basket_item.CalculateBonus(bsk.User(), productInfo.GetPrice()))

				if price.Price == 0 {
					item.AddProblem(
						basket_item.NewProblem(
							basket_item.ProblemNotAvailable,
							"покупка товара недоступна",
						),
					)
				}
			} else {
				item.AddProblem(
					basket_item.NewProblem(
						basket_item.ProblemNotAvailable,
						"покупка товара недоступна",
					),
				)

				logger.Warn(
					"can't find price of product, but it's available...",
					citizap.ProductId(string(item.ItemId())),
					zap.Int("price_column", int(item.PriceColumn())),
				)
			}

			total := item.CalculateMaxCount(int(productsMaxAvailable[string(item.ItemId())]))

			// Помимо использования для особых случаев, теперь это поле используется также для отображения общего доступного количества товара
			if !item.Rules().IsMaxCount() || item.Rules().MaxCount() != total {
				item.Rules().SetMaxCount(total)
			}
			// Проверка, что запрашиваемое кол-во товара стало меньше, чем его наличие на складах. Но тут есть проблема, как
			// только это произойдет (наличие станет меньше, чем запрашиваемое) мы постоянно будем оповещать об этом
			// (путем создания информации). Чтобы этого не произошло, мы запоминаем наличие товара в additions. Таким образом
			// мы можем понять изменилось ли наличие товара или нет с момента добавления info.
			productAdditions := item.Additions().GetProduct()
			if item.Count() > total && (productAdditions.AvailTotal() != total || !productAdditions.IsCountMoreThenAvailChecked()) {
				if item.IsSelected() {
					info := basket_item.NewInfo(
						basket_item.InfoIdCountMoreThanAvail,
						"Запрашиваемое кол-во больше, чем товара в наличии.",
					)
					info.Additionals().CountMoreThenAvail = basket_item.CountMoreThenAvailInfoAdditions{
						AvailCount: total,
					}
					item.AddInfo(info)
				}

				productAdditions.SetAvailTotal(total)
				productAdditions.SetIsCountMoreThenAvailChecked(true)
			}
			// если info было добавлено, но кол-во товаров изменилось на доступное (в случае UpdateItem), удаляем info
			if item.Count() <= total {
				for _, v := range item.Infos() {
					if v.Id() == basket_item.InfoIdCountMoreThanAvail {
						item.CommitInfo(v.Id())
						productAdditions.SetIsCountMoreThenAvailChecked(false)
						break
					}
				}
			}

			// Если пользователь б2б, у него все вариант оплаты либо предоплатные (оплата юр. лица), либо онлайн
			// В таком случае мы убираем флаг предоплатности у позиций
			if user != nil && user.GetB2B() != nil && user.GetB2B().GetIsB2BState() {
				if item.IsPrepaymentMandatory() {
					item.SetPrepaymentMandatory(false)
				}
			} else if item.IsPrepaymentMandatory() != productInfo.GetRegional().GetIsPrepaymentMandatory() {
				item.SetPrepaymentMandatory(productInfo.GetRegional().GetIsPrepaymentMandatory())
			}
		}
	}

	// Если требуется, то подождем завершения запроса возможной конфигурации из API монолита
	wg.Wait()

	return nil
}
