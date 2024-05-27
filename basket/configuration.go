package basket

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"go.citilink.cloud/catalog_types"
	database "go.citilink.cloud/libdatabase"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	"time"
)

const MaxCountOfProductItemsInConf = 10

func NewConfiguration(
	basket *Basket,
	productClient productv1.ProductAPIClient,
	db database.DB,
) *Configuration {
	return &Configuration{basket: basket, productClient: productClient, db: db}
}

type Configuration struct {
	basket        *Basket
	productClient productv1.ProductAPIClient
	db            database.DB
}

func (c *Configuration) Add(
	ctx context.Context,
	confId basket_item.ConfId,
	confType basket_item.ConfType,
	assemblyServiceItemId string,
	confItems []*basket_item.ConfItem,
) ([]*basket_item.Item, error) {
	items, err := c.assembleConfigurationItems(ctx, confId, confType, assemblyServiceItemId, confItems)
	if err != nil {
		return nil, fmt.Errorf("can't assemble item's configuration: %w", err)
	}

	configurations := c.basket.Find(Finders.ByType(basket_item.TypeConfiguration))
	for _, conf := range configurations {
		err := c.basket.Remove(conf, false)
		if err != nil {
			return nil, fmt.Errorf("can't remove item's configuration: %w", err)
		}
	}

	for _, item := range items {
		_, err := c.basket.data.Add(item)
		if err != nil {
			return nil, fmt.Errorf("can't add item's configuration: %w", err)
		}
	}

	return items, nil
}

func (c *Configuration) assembleConfigurationItems(
	ctx context.Context,
	confId basket_item.ConfId,
	confType basket_item.ConfType,
	assemblyServiceItemId string,
	confItems []*basket_item.ConfItem,
) ([]*basket_item.Item, error) {
	price := 0
	// 1 - конфигурация; 1 - услуга сборки; N - комплектующие конфигурации
	items := make([]*basket_item.Item, 0, 2+len(confItems))

	confItemAddition := basket_item.NewConfiguratorItemAdditions(string(confId), confType)

	configurationItem := basket_item.NewConfigurationItem(
		price,
		c.basket.SpaceId(),
		c.basket.PriceColumn(),
	)
	configurationItem.Additions().SetConfiguration(confItemAddition)

	items = append(items, configurationItem)

	productIds := make([]string, 0, len(confItems))
	for _, confItem := range confItems {
		productIds = append(productIds, string(confItem.ProductId))
	}

	findFullResponse, err := c.productClient.FindFull(ctx, &productv1.FindFullRequest{
		Ids:     productIds,
		SpaceId: string(c.basket.SpaceId()),
	})
	if err != nil {
		return nil, internal.NewCatalogError(
			fmt.Errorf("can't get products from catalog: %w", err),
			c.basket.SpaceId(),
		)
	}

	if len(findFullResponse.Infos) != len(productIds) {
		return nil, fmt.Errorf("can't find all products")
	}

	productInfoMap := make(map[catalog_types.ProductId]*productv1.FindFullResponse_FullInfo)
	for _, productInfo := range findFullResponse.Infos {
		productInfoMap[catalog_types.ProductId(productInfo.GetId())] = productInfo
	}

	for _, confItem := range confItems {
		productInfo := productInfoMap[confItem.ProductId]

		price, ok := productInfo.GetPrice().GetPrices()[int32(c.basket.PriceColumn())]
		if !ok {
			return nil, fmt.Errorf("can't get price for product '%s', with price column '%d'",
				confItem.ProductId, int32(c.basket.PriceColumn()))
		}

		productItem := basket_item.NewItem(
			basket_item.ItemId(productInfo.Id),
			basket_item.TypeConfigurationProduct,
			productInfo.GetRegional().GetName(),
			productInfo.GetRegional().GetImageName(),
			confItem.Count,
			int(price.GetPrice()),
			basket_item.CalculateBonus(c.basket.User(), productInfo.GetPrice()),
			c.basket.SpaceId(),
			c.basket.PriceColumn(),
		)

		productItem.SetCountMultiplicity(int(productInfo.GetRegional().Multiplicity))
		productItem.SetPrepaymentMandatory(productInfo.GetRegional().GetIsPrepaymentMandatory())
		productItem.Additions().SetConfiguration(confItemAddition)

		//nolint:staticcheck //https://jira.citilink.ru/browse/WEB-68629
		creditPrograms := make([]catalog_types.CreditProgram, 0, len(productInfo.GetRegional().CreditPrograms))
		//nolint:staticcheck //https://jira.citilink.ru/browse/WEB-68629
		for _, v := range productInfo.GetRegional().CreditPrograms {
			creditPrograms = append(creditPrograms, catalog_types.CreditProgram(v))
		}
		product := basket_item.NewProductItemAdditions(
			catalog_types.CategoryId(productInfo.GetCategory().GetId()),
			creditPrograms,
			int(productInfo.GetRegional().GetVat()),
			int(productInfo.GetStock().GetTotal()),
		)
		product.SetIsAvailInStore(productInfo.GetStock().GetIsAvailInStore())
		product.SetIsOEM(productInfo.GetRegional().GetIsOem())
		product.SetIsCountMoreThenAvailChecked(true)
		product.SetIsAvailForDPD(true)
		product.SetIsMarked(productInfo.GetRegional().GetIsMarked())
		product.SetCategoryName(productInfo.GetCategory().GetName())
		product.SetBrandName(productInfo.GetRegional().GetBrandName())
		product.SetCategoryPath(productInfo.GetCategory().GetShortCategoryPath())
		product.SetShortName(productInfo.GetRegional().GetShortName())
		productItem.Additions().SetProduct(product)
		product.SetIsFnsTracked(productInfo.GetRegional().GetIsFnsTracked())

		err := configurationItem.AddChild(productItem)
		if err != nil {
			return nil, fmt.Errorf("can't add child of configurationItem: %w", err)
		}

		items = append(items, productItem)

		for _, featureService := range confItem.Services {
			featureServiceItem := basket_item.NewItem(
				basket_item.ItemId(featureService.ItemId),
				basket_item.TypeConfigurationProductService,
				featureService.Name,
				"",
				featureService.Count,
				featureService.Price,
				0,
				c.basket.SpaceId(),
				c.basket.PriceColumn(),
			)

			featureServiceItem.Additions().SetConfiguration(confItemAddition)

			err := productItem.AddChild(featureServiceItem)
			if err != nil {
				return nil, fmt.Errorf("can't add feature service item to product item: %w", err)
			}

			items = append(items, featureServiceItem)
		}
	}

	assemblyServiceItem := basket_item.NewItem(
		basket_item.ItemId(assemblyServiceItemId),
		basket_item.TypeConfigurationAssemblyService,
		"",
		"",
		1,
		0,
		0,
		c.basket.SpaceId(),
		c.basket.PriceColumn(),
	)
	assemblyServiceItem.Additions().SetConfiguration(confItemAddition)
	items = append(items, assemblyServiceItem)

	err = configurationItem.AddChild(assemblyServiceItem)
	if err != nil {
		return nil, fmt.Errorf("can't add assembly service item"+
			" to configuration item: %w", err)
	}

	return items, nil
}

// MoveItemFrom перемещает позицию из конфигурации. Особенность данной операции заключается в том, что в
// случае если передвигается позиция типа "товар", то аналог этой позиции добавляется в корзину. Услуги и прочие типы
// позиций просто удаляются
func (c *Configuration) MoveItemFrom(ctx context.Context, uniqId basket_item.UniqId) error {
	configurations := c.basket.Find(Finders.ByType(basket_item.TypeConfiguration))
	for _, conf := range configurations {
		if !conf.Additions().GetConfiguration().IsMutable() {
			return internal.NewLogicErrorWithMsg(errors.New("configuration is template and can't move item out"),
				"Из шаблонной конфигурации нельзя удалять комплектующие.")
		}
	}

	var configurationItem *basket_item.Item
	for _, item := range c.basket.All() {
		if item.Type() == basket_item.TypeConfiguration {
			configurationItem = item
		}
	}

	if configurationItem == nil {
		return fmt.Errorf("no configuration in basket")
	}

	var itemToMove *basket_item.Item
	for _, item := range c.basket.All() {
		if item.Type().IsPartOfConfiguration() && item.UniqId() == uniqId {
			itemToMove = item
		}
	}

	if itemToMove == nil {
		return fmt.Errorf("item %s not found", uniqId)
	}

	if !itemToMove.IsMovableFromConfiguration() {
		return fmt.Errorf("can't move not movable item from configuration")
	}

	if itemToMove.Type().IsProduct() {
		_, err := c.basket.Add(ctx, itemToMove.ItemId(), basket_item.TypeProduct, "", itemToMove.Count(), false)
		if err != nil {
			return fmt.Errorf("can't move product from configuration to basket: %w", err)
		}
	}

	c.basket.data.Remove(itemToMove)

	return nil
}

// MoveItemIn перемещает позицию в конфигурацию. (на данный момент возможно перемещение только товаров)
func (c *Configuration) MoveItemIn(uniqId basket_item.UniqId) error {
	configurations := c.basket.Find(Finders.ByType(basket_item.TypeConfiguration))
	for _, conf := range configurations {
		if !conf.Additions().GetConfiguration().IsMutable() {
			return internal.NewLogicErrorWithMsg(errors.New("configuration is template and can't move item in"),
				"В шаблонную конфигурации нельзя добавлять комплектующие.")
		}
	}

	var configurationItem *basket_item.Item
	for _, item := range c.basket.All() {
		if item.Type() == basket_item.TypeConfiguration {
			configurationItem = item
			break
		}
	}

	if configurationItem == nil {
		return fmt.Errorf("no configuration in basket")
	}

	var itemToMove *basket_item.Item
	for _, item := range c.basket.All() {
		if item.Type() == basket_item.TypeProduct && item.UniqId() == uniqId {
			itemToMove = item
		}
	}

	if itemToMove == nil {
		return fmt.Errorf("item %s not found", uniqId)
	}

	if itemToMove.Type() != basket_item.TypeProduct {
		return fmt.Errorf("can't move not product type item to configuration")
	}

	if !itemToMove.IsMovableToConfiguration() {
		return fmt.Errorf("can't move not movable item to configuration")
	}

	children := c.basket.Find(Finders.ChildrenOf(configurationItem))
	var foundedChild *basket_item.Item
	for _, child := range children {
		if child.Type() == basket_item.TypeConfigurationProduct && child.ItemId() == itemToMove.ItemId() {
			foundedChild = child
			break
		}
	}

	var newItem *basket_item.Item
	if foundedChild == nil {
		newItem = basket_item.NewItem(
			itemToMove.ItemId(),
			basket_item.TypeConfigurationProduct,
			itemToMove.Name(),
			itemToMove.Image(),
			c.fixMoveInItemCount(itemToMove.Count()),
			itemToMove.Price(),
			itemToMove.Bonus(),
			c.basket.SpaceId(),
			c.basket.PriceColumn(),
		)
		newItem.SetPrepaymentMandatory(itemToMove.IsPrepaymentMandatory())
		newItem.Additions().SetProduct(itemToMove.Additions().GetProduct())
		newItem.Additions().SetConfiguration(configurationItem.Additions().GetConfiguration())
		err := configurationItem.AddChild(newItem)
		if err != nil {
			return err
		}
		_, err = c.basket.data.Add(newItem)
		if err != nil {
			return err
		}
	} else {
		foundedChild.FixCount(c.fixMoveInItemCount(foundedChild.Count() + itemToMove.Count()))
	}

	c.basket.data.Remove(itemToMove)

	return nil
}

func (c *Configuration) Disassemble() error {
	configurations := c.basket.Find(Finders.ByType(basket_item.TypeConfiguration))
	var configurationItem *basket_item.Item
	for _, conf := range configurations {
		if !conf.Additions().GetConfiguration().IsMutable() {
			return internal.NewLogicErrorWithMsg(errors.New("configuration is template and can't disassemble"),
				"Шаблонную конфигурацию нельзя разобрать.")
		}
		configurationItem = conf
	}

	if configurationItem == nil {
		return fmt.Errorf("no configuration in basket")
	}

	for _, item := range c.basket.Find(Finders.ChildrenOfRecursive(configurationItem)) {
		if !item.Type().IsProduct() {
			continue
		}

		newProductItem := basket_item.NewItem(
			item.ItemId(),
			basket_item.TypeProduct,
			item.Name(),
			item.Image(),
			item.Count()*configurationItem.Count(),
			item.Price(),
			item.Bonus(),
			c.basket.SpaceId(),
			c.basket.PriceColumn())
		newProductItem.SetCountMultiplicity(item.CountMultiplicity())
		newProductItem.SetPrepaymentMandatory(item.IsPrepaymentMandatory())
		newProductItem.Additions().SetProduct(item.Additions().GetProduct())
		_, err := c.basket.data.Add(newProductItem)
		if err != nil {
			return err
		}
	}

	c.basket.data.Remove(configurationItem)

	return nil
}

func (c *Configuration) Assemble(
	ctx context.Context,
	assemblyServiceItemId string,
	confItems []*basket_item.ConfItem,
	count int,
) error {
	var configurationItem *basket_item.Item
	for _, item := range c.basket.All() {
		if item.Type() == basket_item.TypeConfiguration {
			configurationItem = item
			break
		}
	}

	if configurationItem != nil {
		return fmt.Errorf("configuration already exists in basket")
	}

	itemListStr, err := xml.Marshal(c.basket.All().ToItemList())
	if err != nil {
		return fmt.Errorf("can't marshal itemList: %w", err)
	}

	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	row := c.db.QueryRowxContext(dbCtx, "Configurator.set_temporary_configuration",
		sql.Named("conf_id", basket_item.DefaultConfId),
		sql.Named("item_list", string(itemListStr)),
	)
	if row.Err() != nil {
		return fmt.Errorf("error on query from db: %w", row.Err())
	}

	rawData := struct {
		ConfId         sql.NullString `db:"conf_id"`
		AssemblyTypeId sql.NullInt64  `db:"assembly_type_id"`
		Compatible     sql.NullBool   `db:"compatible"`
	}{}
	err = row.StructScan(&rawData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("error on scanned row: %w", err)
	}

	if !rawData.Compatible.Bool {
		return fmt.Errorf("invalid configuration")
	}

	confId := basket_item.ConfId(rawData.ConfId.String)
	items, err := c.assembleConfigurationItems(ctx, confId, basket_item.ConfTypeUser, assemblyServiceItemId, confItems)
	if err != nil {
		return fmt.Errorf("error on assembling configuration: %w", err)
	}

	for _, item := range items {
		if item.Type() == basket_item.TypeConfiguration {
			err := item.SetCount(count)
			if err != nil {
				return err
			}
		}
	}

	changeableItems := make([]*basket_item.Item, 0, len(confItems))
	for _, confItem := range confItems {
		for _, item := range c.basket.All() {
			if item.Type() == basket_item.TypeProduct && string(item.ItemId()) == string(confItem.ProductId) {
				changeableItems = append(changeableItems, item)
			}
		}
	}

	for _, item := range items {
		for _, changeableItem := range changeableItems {
			if item.ItemId() != changeableItem.ItemId() {
				continue
			}

			if changeableItem.Count() > item.Count()*count {
				err := changeableItem.SetCount(changeableItem.Count() - item.Count()*count)
				if err != nil {
					return err
				}
			} else {
				c.basket.data.Remove(changeableItem)
			}
		}
	}

	for _, item := range items {
		_, err := c.basket.data.Add(item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Configuration) fixMoveInItemCount(count int) int {
	if count > MaxCountOfProductItemsInConf {
		return MaxCountOfProductItemsInConf
	}

	return count
}
