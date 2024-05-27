package resolver

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/mock"
	"go.citilink.cloud/order/internal/order/payment"
	storev1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/store/store/v1"
	"go.citilink.cloud/store_types"
	"testing"
)

func TestNewCreditAndInstallmentInStore(t *testing.T) {
	type args struct {
		storeFinder internal.StoreFinder
	}
	tests := []struct {
		name string
		args args
		want func(storeFinder *mock.MockStoreFinder) *CreditAndInstallmentInStore
	}{
		{
			name: "correct new resolver create",
			want: func(storeFinder *mock.MockStoreFinder) *CreditAndInstallmentInStore {
				return &CreditAndInstallmentInStore{
					storeFinder: storeFinder,
				}
			},
		},
	}
	for _, tt := range tests {
		storeFinder := mock.NewMockStoreFinder(gomock.NewController(t))
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, NewCreditAndInstallmentInStore(storeFinder), tt.want(storeFinder))
		})
	}
}

func TestCreditAndInstallmentInStore_Resolve(t *testing.T) {
	type args struct {
		ctx         context.Context
		resolvedIds func() payment.ResolvedPaymentIdMap
		order       func(ctrl *gomock.Controller) order.Order
		storeFinder func(ctrl *gomock.Controller) internal.StoreFinder
	}

	tests := []struct {
		name string
		args args
		want func() (payment.ResolvedPaymentIdMap, error)
	}{
		{
			name: "allow credit and installment",
			args: args{
				ctx: context.Background(),
				resolvedIds: func() payment.ResolvedPaymentIdMap {
					return payment.ResolvedPaymentIdMap{
						order.PaymentIdCash:         payment.NewResolvedPaymentId(order.PaymentIdCash),
						order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
						order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
					}
				},
				order: func(ctrl *gomock.Controller) order.Order {
					self := mock.NewMockSelfDelivery(ctrl)
					self.EXPECT().PupId().Return(store_types.NewPupId("kremlin")).Times(1)
					delivery := mock.NewMockDelivery(ctrl)
					delivery.EXPECT().Id().Return(order.DeliveryIdSelf).Times(1)
					delivery.EXPECT().Self(context.Background()).Return(self, nil).Times(1)
					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().Delivery().Return(delivery).Times(2)

					return ordr
				},
				storeFinder: func(ctrl *gomock.Controller) internal.StoreFinder {
					store := &storev1.Store{
						PupId:             "kremlin",
						SpaceId:           "msk_cl",
						IsCreditAvailable: true,
					}
					storeFinder := mock.NewMockStoreFinder(ctrl)
					storeFinder.EXPECT().FindByPupId(store_types.PupId("kremlin")).Return(store).Times(1)

					return storeFinder
				},
			},
			want: func() (payment.ResolvedPaymentIdMap, error) {
				return payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:         payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
					order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
				}, nil
			},
		},
		{
			name: "disallow credit and installment",
			args: args{
				ctx: context.Background(),
				resolvedIds: func() payment.ResolvedPaymentIdMap {
					return payment.ResolvedPaymentIdMap{
						order.PaymentIdCash:         payment.NewResolvedPaymentId(order.PaymentIdCash),
						order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
						order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
					}
				},
				order: func(ctrl *gomock.Controller) order.Order {
					self := mock.NewMockSelfDelivery(ctrl)
					self.EXPECT().PupId().Return(store_types.NewPupId("kremlin")).Times(1)
					delivery := mock.NewMockDelivery(ctrl)
					delivery.EXPECT().Id().Return(order.DeliveryIdSelf).Times(1)
					delivery.EXPECT().Self(context.Background()).Return(self, nil).Times(1)
					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().Delivery().Return(delivery).Times(2)

					return ordr
				},
				storeFinder: func(ctrl *gomock.Controller) internal.StoreFinder {
					store := &storev1.Store{
						PupId:             "kremlin",
						SpaceId:           "msk_cl",
						IsCreditAvailable: false,
					}
					storeFinder := mock.NewMockStoreFinder(ctrl)
					storeFinder.EXPECT().FindByPupId(store_types.PupId("kremlin")).Return(store).Times(1)

					return storeFinder
				},
			},
			want: func() (payment.ResolvedPaymentIdMap, error) {
				resolvedIds := payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:         payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
					order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
				}
				resolvedIds[order.PaymentIdInstallments].Disallow(
					order.AllowStatusLimited,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCreditAndInstallmentInStoreNotAvailable, order.SubsystemBasket).
						WithMessage("Оплата рассрочкой на данной точке выдачи запрещена!"),
				)
				resolvedIds[order.PaymentIdCredit].Disallow(
					order.AllowStatusLimited,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdCreditAndInstallmentInStoreNotAvailable, order.SubsystemBasket).
						WithMessage("Оплата кредитом на данной точке выдачи запрещена!"),
				)

				return resolvedIds, nil
			},
		},
		{
			name: "delivery_id isn't self",
			args: args{
				ctx: context.Background(),
				resolvedIds: func() payment.ResolvedPaymentIdMap {
					return payment.ResolvedPaymentIdMap{
						order.PaymentIdCash:         payment.NewResolvedPaymentId(order.PaymentIdCash),
						order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
						order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
					}
				},
				order: func(ctrl *gomock.Controller) order.Order {
					delivery := mock.NewMockDelivery(ctrl)
					delivery.EXPECT().Id().Return(order.DeliveryIdCitilinkCourier).Times(1)
					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().Delivery().Return(delivery).Times(1)

					return ordr
				},
				storeFinder: func(ctrl *gomock.Controller) internal.StoreFinder {
					return mock.NewMockStoreFinder(ctrl)
				},
			},
			want: func() (payment.ResolvedPaymentIdMap, error) {
				return payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:         payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
					order.PaymentIdCredit:       payment.NewResolvedPaymentId(order.PaymentIdCredit),
				}, nil
			},
		},
		{
			name: "can't get self delivery",
			args: args{
				ctx: context.Background(),
				resolvedIds: func() payment.ResolvedPaymentIdMap {
					return payment.ResolvedPaymentIdMap{}
				},
				order: func(ctrl *gomock.Controller) order.Order {
					delivery := mock.NewMockDelivery(ctrl)
					delivery.EXPECT().Id().Return(order.DeliveryIdSelf).Times(1)
					delivery.EXPECT().Self(context.Background()).Return(nil, fmt.Errorf("booom")).Times(1)
					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().Delivery().Return(delivery).Times(2)

					return ordr
				},
				storeFinder: func(ctrl *gomock.Controller) internal.StoreFinder {
					return mock.NewMockStoreFinder(ctrl)
				},
			},
			want: func() (payment.ResolvedPaymentIdMap, error) {
				return payment.ResolvedPaymentIdMap{}, fmt.Errorf("error getting self delivery: booom")
			},
		},
		{
			name: "can't find pup_id",
			args: args{
				ctx: context.Background(),
				resolvedIds: func() payment.ResolvedPaymentIdMap {
					return payment.ResolvedPaymentIdMap{}
				},
				order: func(ctrl *gomock.Controller) order.Order {
					self := mock.NewMockSelfDelivery(ctrl)
					self.EXPECT().PupId().Return(store_types.NewPupId("kremlin")).Times(1)
					delivery := mock.NewMockDelivery(ctrl)
					delivery.EXPECT().Id().Return(order.DeliveryIdSelf).Times(1)
					delivery.EXPECT().Self(context.Background()).Return(self, nil).Times(1)
					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().Delivery().Return(delivery).Times(2)

					return ordr
				},
				storeFinder: func(ctrl *gomock.Controller) internal.StoreFinder {
					storeFinder := mock.NewMockStoreFinder(ctrl)
					storeFinder.EXPECT().FindByPupId(store_types.PupId("kremlin")).Return(nil).Times(1)

					return storeFinder
				},
			},
			want: func() (payment.ResolvedPaymentIdMap, error) {
				return payment.ResolvedPaymentIdMap{}, fmt.Errorf("can't find store with pup_id: kremlin")
			},
		},
		{
			name: "pup_id empty",
			args: args{
				ctx: context.Background(),
				resolvedIds: func() payment.ResolvedPaymentIdMap {
					return payment.ResolvedPaymentIdMap{}
				},
				order: func(ctrl *gomock.Controller) order.Order {
					self := mock.NewMockSelfDelivery(ctrl)
					self.EXPECT().PupId().Return(store_types.NewPupId("")).Times(1)
					delivery := mock.NewMockDelivery(ctrl)
					delivery.EXPECT().Id().Return(order.DeliveryIdSelf).Times(1)
					delivery.EXPECT().Self(context.Background()).Return(self, nil).Times(1)
					ordr := order.NewMockOrder(ctrl)
					ordr.EXPECT().Delivery().Return(delivery).Times(2)

					return ordr
				},
				storeFinder: func(ctrl *gomock.Controller) internal.StoreFinder {
					storeFinder := mock.NewMockStoreFinder(ctrl)

					return storeFinder
				},
			},
			want: func() (payment.ResolvedPaymentIdMap, error) {
				return payment.ResolvedPaymentIdMap{}, nil
			},
		},
	}

	for _, tt := range tests {
		ctrl := gomock.NewController(t)

		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.args.ctx
			creditAndInstallmentInStore := NewCreditAndInstallmentInStore(tt.args.storeFinder(ctrl))
			want, wantErr := tt.want()
			got := tt.args.resolvedIds()
			err := creditAndInstallmentInStore.Resolve(ctx, got, tt.args.order(ctrl))
			if err != nil {
				assert.EqualError(t, err, wantErr.Error())
			}
			assert.Equal(t, len(want), len(got))
			for i := range want {
				assert.Equal(t, want[i].Status(), got[i].Status())
				assert.Equal(t, len(want[i].Reasons()), len(got[i].Reasons()))
			}
		})
	}
}
