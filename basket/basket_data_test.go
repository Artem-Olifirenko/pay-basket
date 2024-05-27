package basket

import (
	"github.com/stretchr/testify/suite"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"testing"
)

type BasketDataSuite struct {
	suite.Suite
}

func (b *BasketDataSuite) SetupTest() {
}

func (b *BasketDataSuite) SetupSubTest() {
}

func (b *BasketDataSuite) TestBasketData_Fingerprint() {
	tests := []struct {
		name    string
		request *BasketData
		want    func() string
	}{
		{
			name:    "success: empty msk cl retail",
			request: NewBasketData("msk_cl", catalog_types.PriceColumnRetail, "msk"),
			want: func() string {
				return "4216353138595167701"
			},
		},
		{
			name:    "success: empty spb cl retail",
			request: NewBasketData("spb_cl", catalog_types.PriceColumnRetail, "spb"),
			want: func() string {
				return "1534197782466464695"
			},
		},
		{
			name:    "success: empty msk cl club",
			request: NewBasketData("spb_cl", catalog_types.PriceColumnClub, "spb"),
			want: func() string {
				return "3844539885842716164"
			},
		},
		{
			name: "success: with one item count 1",
			request: func() *BasketData {
				b := NewBasketData("msk_cl", catalog_types.PriceColumnRetail, "msk")
				_, _ = b.Add(basket_item.NewItem(
					"123",
					basket_item.TypeProduct,
					"name",
					"image",
					1,
					1,
					1,
					"msk_cl",
					catalog_types.PriceColumnRetail))
				return b
			}(),
			want: func() string {
				return "4754936641902033992"
			},
		},
		{
			name: "success: with one item count 2",
			request: func() *BasketData {
				b := NewBasketData("msk_cl", catalog_types.PriceColumnRetail, "msk")
				_, _ = b.Add(basket_item.NewItem(
					"123",
					basket_item.TypeProduct,
					"name",
					"image",
					2,
					1,
					1,
					"msk_cl",
					catalog_types.PriceColumnRetail))
				return b
			}(),
			want: func() string {
				return "15399763920430204393"
			},
		},
		{
			name: "success: with one item count 2 and price 2",
			request: func() *BasketData {
				b := NewBasketData("msk_cl", catalog_types.PriceColumnRetail, "msk")
				_, _ = b.Add(basket_item.NewItem(
					"123",
					basket_item.TypeProduct,
					"name",
					"image",
					2,
					2,
					1,
					"msk_cl",
					catalog_types.PriceColumnRetail))
				return b
			}(),
			want: func() string {
				return "15780090000240964663"
			},
		},
		{
			name: "success: with one item count 2 and price 2 and bonus 2",
			request: func() *BasketData {
				b := NewBasketData("msk_cl", catalog_types.PriceColumnRetail, "msk")
				_, _ = b.Add(basket_item.NewItem(
					"123",
					basket_item.TypeProduct,
					"name",
					"image",
					2,
					2,
					2,
					"msk_cl",
					catalog_types.PriceColumnRetail))
				return b
			}(),
			want: func() string {
				return "15780090000240964663"
			},
		},
		{
			name: "success: with two items",
			request: func() *BasketData {
				b := NewBasketData("msk_cl", catalog_types.PriceColumnRetail, "msk")
				_, _ = b.Add(basket_item.NewItem(
					"123",
					basket_item.TypeProduct,
					"name",
					"image",
					1,
					1,
					1,
					"msk_cl",
					catalog_types.PriceColumnRetail))
				_, _ = b.Add(basket_item.NewItem(
					"1234",
					basket_item.TypeProduct,
					"name",
					"image",
					1,
					1,
					1,
					"msk_cl",
					catalog_types.PriceColumnRetail))
				return b
			}(),
			want: func() string {
				return "585458320222589841"
			},
		},
		{
			name: "success: with two items in reverse",
			request: func() *BasketData {
				b := NewBasketData("msk_cl", catalog_types.PriceColumnRetail, "msk")
				_, _ = b.Add(basket_item.NewItem(
					"1234",
					basket_item.TypeProduct,
					"name",
					"image",
					1,
					1,
					1,
					"msk_cl",
					catalog_types.PriceColumnRetail))
				_, _ = b.Add(basket_item.NewItem(
					"123",
					basket_item.TypeProduct,
					"name",
					"image",
					1,
					1,
					1,
					"msk_cl",
					catalog_types.PriceColumnRetail))
				return b
			}(),
			want: func() string {
				return "585458320222589841"
			},
		},
		{
			name: "success: with equal items id",
			request: func() *BasketData {
				b := NewBasketData("msk_cl", catalog_types.PriceColumnRetail, "msk")
				for i, itemId := range []basket_item.ItemId{"123", "1234", "5555", "6666"} {
					item, _ := b.Add(basket_item.NewItem(
						itemId,
						basket_item.TypeProduct,
						"name",
						"image",
						1,
						1,
						1,
						"msk_cl",
						catalog_types.PriceColumnRetail))
					insurance, _ := b.Add(basket_item.NewItem(
						"J5403",
						basket_item.TypeInsuranceServiceForProduct,
						"name",
						"image",
						1,
						i,
						1,
						"msk_cl",
						catalog_types.PriceColumnRetail))
					_ = item.AddChild(insurance)
				}

				return b
			}(),
			want: func() string {
				return "2016431460704393095"
			},
		},
	}
	for _, test := range tests {
		b.Run(test.name, func() {
			b.Assert().Equal(test.want(), test.request.Fingerprint())
		})
	}
}

func TestBasketDataSuite(t *testing.T) {
	suite.Run(t, new(BasketDataSuite))
}
