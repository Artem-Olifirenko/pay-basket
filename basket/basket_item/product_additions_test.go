package basket_item

import (
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/catalog_types"
	"testing"
)

// TestNewProductItemAdditions
func TestNewProductItemAdditions(t *testing.T) {
	got := NewProductItemAdditions(
		catalog_types.CategoryId(1),
		[]catalog_types.CreditProgram{"test_prog_1", "test_prog_2"},
		20,
		1000,
	)
	assert.Equal(t, &ProductItemAdditions{
		categoryId:           catalog_types.CategoryId(1),
		vat:                  20,
		availTotal:           1000,
		creditPrograms:       []catalog_types.CreditProgram{"test_prog_1", "test_prog_2"},
		markedPurchaseReason: MarkedPurchaseReasonUnknown,
	}, got)
}

func TestProductItemAdditions_IsAvailInStore(t *testing.T) {
	adds := ProductItemAdditions{
		isAvailInStore: true,
	}
	assert.True(t, adds.IsAvailInStore())
}

func TestProductItemAdditions_SetIsAvailInStore(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetIsAvailInStore(true)
	assert.True(t, adds.isAvailInStore)
}

func TestProductItemAdditions_Vat(t *testing.T) {
	adds := ProductItemAdditions{vat: 20}
	assert.Equal(t, 20, adds.Vat())
}

func TestProductItemAdditions_SetVat(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetVat(20)
	assert.Equal(t, 20, adds.vat)
}

func TestProductItemAdditions_CategoryId(t *testing.T) {
	adds := ProductItemAdditions{categoryId: catalog_types.CategoryId(5)}
	assert.Equal(t, catalog_types.CategoryId(5), adds.CategoryId())
}

func TestProductItemAdditions_SetCategoryId(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetCategoryId(5)
	assert.Equal(t, catalog_types.CategoryId(5), adds.categoryId)
}

func TestProductItemAdditions_CategoryName(t *testing.T) {
	adds := ProductItemAdditions{categoryName: "test_category_name"}
	assert.Equal(t, "test_category_name", adds.CategoryName())
}

func TestProductItemAdditions_SetCategoryName(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetCategoryName("test_category_name")
	assert.Equal(t, "test_category_name", adds.categoryName)
}

func TestProductItemAdditions_BrandName(t *testing.T) {
	adds := ProductItemAdditions{brandName: "test_brand_name"}
	assert.Equal(t, "test_brand_name", adds.BrandName())
}

func TestProductItemAdditions_SetBrandName(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetBrandName("test_brand_name")
	assert.Equal(t, "test_brand_name", adds.brandName)
}

func TestProductItemAdditions_CategoryPath(t *testing.T) {
	adds := ProductItemAdditions{categoryPath: "test_category_path"}
	assert.Equal(t, "test_category_path", adds.CategoryPath())
}

func TestProductItemAdditions_SetCategoryPath(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetCategoryPath("test_category_path")
	assert.Equal(t, "test_category_path", adds.categoryPath)
}

func TestProductItemAdditions_ShortName(t *testing.T) {
	adds := ProductItemAdditions{shortName: "test_short_name"}
	assert.Equal(t, "test_short_name", adds.ShortName())
}

func TestProductItemAdditions_SetShortName(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetShortName("test_short_name")
	assert.Equal(t, "test_short_name", adds.shortName)
}

func TestProductItemAdditions_IsOEM(t *testing.T) {
	adds := ProductItemAdditions{isOEM: true}
	assert.True(t, adds.IsOEM())
}

func TestProductItemAdditions_SetIsOEM(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetIsOEM(true)
	assert.True(t, adds.isOEM)
}

func TestProductItemAdditions_AvailTotal(t *testing.T) {
	adds := ProductItemAdditions{availTotal: 1000}
	assert.Equal(t, 1000, adds.AvailTotal())
}

func TestProductItemAdditions_SetAvailTotal(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetAvailTotal(1000)
	assert.Equal(t, 1000, adds.availTotal)
}

func TestProductItemAdditions_IsCountMoreThenAvailChecked(t *testing.T) {
	adds := ProductItemAdditions{isCountMoreThenAvailChecked: true}
	assert.True(t, adds.IsCountMoreThenAvailChecked())
}

func TestProductItemAdditions_SetIsCountMoreThenAvailChecked(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetIsCountMoreThenAvailChecked(true)
	assert.True(t, adds.isCountMoreThenAvailChecked)
}

func TestProductItemAdditions_CreditPrograms(t *testing.T) {
	adds := ProductItemAdditions{
		creditPrograms: []catalog_types.CreditProgram{"test_prog_1", "test_prog_2"},
	}
	assert.Equal(
		t,
		[]catalog_types.CreditProgram{"test_prog_1", "test_prog_2"},
		adds.CreditPrograms(),
	)
}

func TestProductItemAdditions_SetCreditPrograms(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetCreditPrograms([]catalog_types.CreditProgram{"test_prog_1", "test_prog_2"})
	assert.Equal(t, []catalog_types.CreditProgram{"test_prog_1", "test_prog_2"}, adds.creditPrograms)
}

func TestProductItemAdditions_IsAvailForDPD(t *testing.T) {
	adds := ProductItemAdditions{isAvailForDPD: true}
	assert.True(t, adds.IsAvailForDPD())
}

func TestProductItemAdditions_SetIsAvailForDPD(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetIsAvailForDPD(true)
	assert.True(t, adds.isAvailForDPD)
}

func TestProductItemAdditions_IsMarked(t *testing.T) {
	adds := ProductItemAdditions{isMarked: true}
	assert.True(t, adds.IsMarked())
}

func TestProductItemAdditions_SetIsMarked(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetIsMarked(true)
	assert.True(t, adds.isMarked)
}

func TestProductItemAdditions_IsDiscounted(t *testing.T) {
	adds := ProductItemAdditions{isDiscounted: true}
	assert.True(t, adds.IsDiscounted())
}

func TestProductItemAdditions_SetIsDiscounted(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetIsDiscounted(true)
	assert.True(t, adds.isDiscounted)
}

func TestProductItemAdditions_MarkedPurchaseReason(t *testing.T) {
	adds := ProductItemAdditions{markedPurchaseReason: MarkedPurchaseReasonForSelfNeeds}
	assert.Equal(t, MarkedPurchaseReasonForSelfNeeds, adds.MarkedPurchaseReason())
}

func TestProductItemAdditions_SetMarkedPurchaseReason(t *testing.T) {
	adds := ProductItemAdditions{}
	adds.SetMarkedPurchaseReason(MarkedPurchaseReasonForResale)
	assert.Equal(t, MarkedPurchaseReasonForResale, adds.markedPurchaseReason)
}
