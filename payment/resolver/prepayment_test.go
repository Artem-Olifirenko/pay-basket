package resolver

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/payment"
	"go.citilink.cloud/store_types"
	"testing"
)

func generateItem(id basket_item.ItemId, iType basket_item.Type, prepaymentMandatory bool) *basket_item.Item {
	i := basket_item.NewItem(
		id,
		iType,
		"item_"+string(iType),
		"image",
		0,
		0,
		0,
		store_types.NewSpaceId("msk_cl"),
		catalog_types.PriceColumnRetail,
	)
	i.SetPrepaymentMandatory(prepaymentMandatory)
	return i
}

func TestPrepayment_Resolve(t *testing.T) {
	type args struct {
		resolvedIdsMap payment.ResolvedPaymentIdMap
		ordr           func(ctrl *gomock.Controller, ctx context.Context) order.Order
	}
	tests := []struct {
		name string
		args *args
		want func() (order.AllowStatus, error)
	}{
		{
			name: "allowed payments for prepayment",
			args: &args{
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdCardsOnline:            payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
					order.PaymentIdYandex:                 payment.NewResolvedPaymentId(order.PaymentIdYandex),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
					order.PaymentIdInstallments:           payment.NewResolvedPaymentId(order.PaymentIdInstallments),
					order.PaymentIdCredit:                 payment.NewResolvedPaymentId(order.PaymentIdCredit),
					order.PaymentIdSbp:                    payment.NewResolvedPaymentId(order.PaymentIdSbp),
					order.PaymentIdSberPay:                payment.NewResolvedPaymentId(order.PaymentIdSberPay),
				},
				ordr: func(ctrl *gomock.Controller, ctx context.Context) order.Order {
					o := order.NewMockOrder(ctrl)
					b := order.NewMockBasket(ctrl)
					b.EXPECT().All().Return(basket_item.Items{generateItem("123123", "product", true)})
					o.EXPECT().Basket(ctx).Return(b, nil)
					return o
				},
			},
			want: func() (order.AllowStatus, error) {
				return order.AllowStatusAllow, nil
			},
		},
		{
			name: "limited payments for prepayment",
			args: &args{
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdInvalid:           payment.NewResolvedPaymentId(order.PaymentIdInvalid),
					order.PaymentIdUnknown:           payment.NewResolvedPaymentId(order.PaymentIdUnknown),
					order.PaymentIdCash:              payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdTerminalOrCashbox: payment.NewResolvedPaymentId(order.PaymentIdTerminalOrCashbox),
					order.PaymentIdWebmoney:          payment.NewResolvedPaymentId(order.PaymentIdWebmoney),
					order.PaymentIdTerminalPinpad:    payment.NewResolvedPaymentId(order.PaymentIdTerminalPinpad),
					order.PaymentIdCashWithCard:      payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
					order.PaymentIdApplePay:          payment.NewResolvedPaymentId(order.PaymentIdApplePay),
				},
				ordr: func(ctrl *gomock.Controller, ctx context.Context) order.Order {
					o := order.NewMockOrder(ctrl)
					b := order.NewMockBasket(ctrl)
					b.EXPECT().All().Return(basket_item.Items{generateItem("123123", "product", true)})
					o.EXPECT().Basket(ctx).Return(b, nil)
					return o
				},
			},
			want: func() (order.AllowStatus, error) {
				return order.AllowStatusLimited, nil
			},
		},
		{
			name: "no prepayment mandatory",
			args: &args{
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdInvalid:                payment.NewResolvedPaymentId(order.PaymentIdInvalid),
					order.PaymentIdUnknown:                payment.NewResolvedPaymentId(order.PaymentIdUnknown),
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdCash:                   payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCardsOnline:            payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
					order.PaymentIdYandex:                 payment.NewResolvedPaymentId(order.PaymentIdYandex),
					order.PaymentIdTerminalOrCashbox:      payment.NewResolvedPaymentId(order.PaymentIdTerminalOrCashbox),
					order.PaymentIdWebmoney:               payment.NewResolvedPaymentId(order.PaymentIdWebmoney),
					order.PaymentIdTerminalPinpad:         payment.NewResolvedPaymentId(order.PaymentIdTerminalPinpad),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
					order.PaymentIdSbp:                    payment.NewResolvedPaymentId(order.PaymentIdSbp),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
					order.PaymentIdApplePay:               payment.NewResolvedPaymentId(order.PaymentIdApplePay),
					order.PaymentIdSberPay:                payment.NewResolvedPaymentId(order.PaymentIdSberPay),
				},
				ordr: func(ctrl *gomock.Controller, ctx context.Context) order.Order {
					o := order.NewMockOrder(ctrl)
					b := order.NewMockBasket(ctrl)
					b.EXPECT().All().Return(basket_item.Items{generateItem("123123", "product", false)})
					o.EXPECT().Basket(ctx).Return(b, nil)
					return o
				},
			},
			want: func() (order.AllowStatus, error) {
				return order.AllowStatusAllow, nil
			},
		},
		{
			name: "basket error",
			args: &args{
				resolvedIdsMap: payment.ResolvedPaymentIdMap{},
				ordr: func(ctrl *gomock.Controller, ctx context.Context) order.Order {
					o := order.NewMockOrder(ctrl)
					o.EXPECT().Basket(ctx).Return(nil, fmt.Errorf("some error"))
					return o
				},
			},
			want: func() (order.AllowStatus, error) {
				return order.AllowStatusAllow, fmt.Errorf("some error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()
			r := &Prepayment{}
			want, wantErr := tt.want()
			err := r.Resolve(ctx, tt.args.resolvedIdsMap, tt.args.ordr(ctrl, ctx))
			if err != nil {
				assert.EqualError(t, err, wantErr.Error())
			} else {
				assert.NoError(t, err)
				for _, resolvedId := range tt.args.resolvedIdsMap {
					assert.Equal(t, want, resolvedId.Status())
				}
			}
		})
	}
}

func TestPrepayment_NewPrepayment(t *testing.T) {
	expected := &Prepayment{}
	actual := NewPrepayment()

	assert.Equal(t, expected, actual)
}
