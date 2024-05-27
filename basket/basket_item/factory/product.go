package factory

import (
	"context"
	"fmt"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/specs/domains/catalog_facade"
	v1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/overall/v1"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	facade_productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalogfacade/product/v1"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
	"go.citilink.cloud/store_types"
)

type productItemFactory struct {
	productClient       productv1.ProductAPIClient
	facadeProductClient facade_productv1.ProductAPIClient
}

func NewProductItemFactory(
	productClient productv1.ProductAPIClient,
	facadeProductClient facade_productv1.ProductAPIClient,
) *productItemFactory {
	return &productItemFactory{
		productClient:       productClient,
		facadeProductClient: facadeProductClient,
	}
}

func (f *productItemFactory) Creatable(itemType basket_item.Type) bool {
	return itemType == basket_item.TypeProduct
}

func (f *productItemFactory) Create(
	ctx context.Context,
	itemId basket_item.ItemId,
	spaceId store_types.SpaceId,
	_ basket_item.Type,
	count int,
	_ *basket_item.Item,
	priceColumn catalog_types.PriceColumn,
	user *userv1.User,
	ignoreFairPrice bool,
) (*basket_item.Item, error) {
	response, err := f.productClient.FindFull(ctx, &productv1.FindFullRequest{
		Ids:     []string{string(itemId)},
		SpaceId: string(spaceId),
	})
	if err != nil {
		return nil, internal.NewCatalogError(
			fmt.Errorf("can't get product with id '%s' from catalog microservice: %w", itemId, err),
			spaceId,
		)
	}

	// TODO: WEB-58006 на данный момент можно использовать catalog и catalog-facade одновременно, но в последствии
	// необходимо будет сделать переход
	visitor := catalog_facade.NewVisitor(spaceId, user)
	facadeResponse, err := f.facadeProductClient.FilterInfo(ctx, &facade_productv1.FilterInfoRequest{
		VisitorWithProducts: []*facade_productv1.VisitorWithProducts{
			{
				Visitor: visitor.VisitorInfo(),
				Ids:     []string{string(itemId)},
			},
		},
	})
	if err != nil {
		return nil, internal.NewCatalogError(
			fmt.Errorf("can't get product with id '%s' from catalog facade microservice: %w", itemId, err),
			spaceId,
		)
	}

	if len(response.GetInfos()) == 0 || len(facadeResponse.GetProductInfo()) == 0 {
		return nil, internal.NewNotFoundError(fmt.Errorf("product '%s' not found in region '%s'", itemId, spaceId))
	}

	productInfo := response.GetInfos()[0]
	facadeProductInfo := facadeResponse.GetProductInfo()[0]

	price, ok := productInfo.GetPrice().GetPrices()[int32(catalog_types.PriceColumnRetail)]
	if ok && productInfo.GetPrice().GetIsFairPrice() && ignoreFairPrice {
		price.Price, productInfo.GetPrice().OldPrice = productInfo.GetPrice().OldPrice, price.Price
	}

	price, ok = productInfo.GetPrice().GetPrices()[int32(priceColumn)]
	if !ok {
		price = &v1.Price{
			Price: 0,
		}
	}

	productItem := basket_item.NewItem(
		itemId,
		basket_item.TypeProduct,
		productInfo.GetRegional().GetName(),
		productInfo.GetRegional().GetImageName(),
		f.fixCount(count, productInfo.GetRegional()),
		int(price.Price),
		basket_item.CalculateBonus(user, productInfo.GetPrice()),
		spaceId,
		priceColumn,
	)

	if user == nil || user.GetB2B() == nil || !user.GetB2B().GetIsB2BState() {
		productItem.SetPrepaymentMandatory(productInfo.GetRegional().GetIsPrepaymentMandatory())
	}

	productItem.SetIgnoreFairPrice(ignoreFairPrice)

	if !facadeProductInfo.GetInfo().GetIsAvailable() {
		productItem.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable,
			"товара нет в наличии"))
	}

	if price.Price == 0 {
		productItem.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable,
			"товар купить невозможно"))
	}

	productItem.SetCountMultiplicity(int(productInfo.GetRegional().GetMultiplicity()))

	//nolint:staticcheck //https://jira.citilink.ru/browse/WEB-68629
	creditPrograms := make([]catalog_types.CreditProgram, 0, len(productInfo.GetRegional().GetCreditPrograms()))
	//nolint:staticcheck //https://jira.citilink.ru/browse/WEB-68629
	for _, v := range productInfo.GetRegional().GetCreditPrograms() {
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
	product.SetIsDiscounted(productInfo.GetRegional().IsDiscounted)
	productItem.Additions().SetProduct(product)
	product.SetIsFnsTracked(productInfo.GetRegional().GetIsFnsTracked())

	return productItem, nil
}

func (f *productItemFactory) fixCount(count int, productRegional *productv1.ProductRegional) int {
	if count > basket_item.LimitTotalGoods {
		count = basket_item.LimitTotalGoods
	}
	fixedCount := count
	multiplicity := int(productRegional.GetMultiplicity())
	if multiplicity != 0 && count%multiplicity != 0 {
		if count <= multiplicity {
			fixedCount = multiplicity
		} else {
			fixedCount = count / multiplicity * multiplicity
		}
	}

	return fixedCount
}
