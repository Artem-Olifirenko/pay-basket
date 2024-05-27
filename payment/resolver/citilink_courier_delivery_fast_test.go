package resolver

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/mock"
	"go.citilink.cloud/order/internal/order/payment"
	"testing"
)

func TestCitilinkCourierDeliveryFast_NewFastCitilinkCourierDelivery(t *testing.T) {
	expected := &FastCitilinkCourierDelivery{}
	actual := NewFastCitilinkCourierDelivery()

	assert.Equal(t, expected, actual)
}

func TestCitilinkCourierDeliveryFast_Resolve(t *testing.T) {
	type args struct {
		ctx            context.Context
		resolvedIdsMap payment.ResolvedPaymentIdMap
	}
	type testCase struct {
		name    string
		args    args
		init    func(ordr *order.MockOrder, dlvr *mock.MockDelivery, citilinkCourDel *mock.MockCitilinkCourierDelivery)
		want    []order.AllowStatus
		payment []order.PaymentId
		err     error
	}
	/*
		TODO: Пока так, но тут, по-хорошему, надо мокать вызов resolvedId.Disallow, чтобы unit-тест не выходил за границы
		 тестируемого класса/структуры (иначе это уже не unit). Для этого надо изменить тип payment.ResolvedPaymentIdMap,
		 чтобы внутри был не *ResolvedId, а его интерфейс, которого пока нет (чтобы можно было подсунуть mock).
	*/
	tests := []testCase{
		{
			name: "installments case",
			args: args{
				ctx: context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:         payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
				},
			},
			init: func(ordr *order.MockOrder, dlvr *mock.MockDelivery, citilinkCourDel *mock.MockCitilinkCourierDelivery) {
				ordr.EXPECT().Delivery().Times(2).Return(dlvr)
				dlvr.EXPECT().Id().Return(order.DeliveryIdCitilinkCourier)
				dlvr.EXPECT().CitilinkCourier(gomock.Any()).Return(citilinkCourDel)
				citilinkCourDel.EXPECT().Id().Return(order.CitilinkCourierDeliveryIdFast)
			},
			want: []order.AllowStatus{
				order.AllowStatusAllow,
				order.AllowStatusLimited,
			},
			payment: []order.PaymentId{
				order.PaymentIdCash,
				order.PaymentIdInstallments,
			},
			err: nil,
		},
		{
			name: "credit case",
			args: args{
				ctx: context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:   payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCredit: payment.NewResolvedPaymentId(order.PaymentIdCredit),
				},
			},
			init: func(ordr *order.MockOrder, dlvr *mock.MockDelivery, citilinkCourDel *mock.MockCitilinkCourierDelivery) {
				ordr.EXPECT().Delivery().Times(2).Return(dlvr)
				dlvr.EXPECT().Id().Return(order.DeliveryIdCitilinkCourier)
				dlvr.EXPECT().CitilinkCourier(gomock.Any()).Return(citilinkCourDel)
				citilinkCourDel.EXPECT().Id().Return(order.CitilinkCourierDeliveryIdFast)
			},
			want: []order.AllowStatus{
				order.AllowStatusAllow,
				order.AllowStatusLimited,
			},
			payment: []order.PaymentId{
				order.PaymentIdCash,
				order.PaymentIdCredit,
			},
			err: nil,
		},
		{
			name: "cards online case",
			args: args{
				ctx: context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:        payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
				},
			},
			init: func(ordr *order.MockOrder, dlvr *mock.MockDelivery, citilinkCourDel *mock.MockCitilinkCourierDelivery) {
				ordr.EXPECT().Delivery().Times(2).Return(dlvr)
				dlvr.EXPECT().Id().Return(order.DeliveryIdCitilinkCourier)
				dlvr.EXPECT().CitilinkCourier(gomock.Any()).Return(citilinkCourDel)
				citilinkCourDel.EXPECT().Id().Return(order.CitilinkCourierDeliveryIdFast)
			},
			want: []order.AllowStatus{
				order.AllowStatusAllow,
				order.AllowStatusLimited,
			},
			payment: []order.PaymentId{
				order.PaymentIdCash,
				order.PaymentIdCardsOnline,
			},
			err: nil,
		},
		{
			name: "sbp case",
			args: args{
				ctx: context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCash: payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdSbp:  payment.NewResolvedPaymentId(order.PaymentIdSbp),
				},
			},
			init: func(ordr *order.MockOrder, dlvr *mock.MockDelivery, citilinkCourDel *mock.MockCitilinkCourierDelivery) {
				ordr.EXPECT().Delivery().Times(2).Return(dlvr)
				dlvr.EXPECT().Id().Return(order.DeliveryIdCitilinkCourier)
				dlvr.EXPECT().CitilinkCourier(gomock.Any()).Return(citilinkCourDel)
				citilinkCourDel.EXPECT().Id().Return(order.CitilinkCourierDeliveryIdFast)
			},
			want: []order.AllowStatus{
				order.AllowStatusAllow,
				order.AllowStatusLimited,
			},
			payment: []order.PaymentId{
				order.PaymentIdCash,
				order.PaymentIdSbp,
			},
			err: nil,
		},
		{
			name: "yandex case",
			args: args{
				ctx: context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:   payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdYandex: payment.NewResolvedPaymentId(order.PaymentIdYandex),
				},
			},
			init: func(ordr *order.MockOrder, dlvr *mock.MockDelivery, citilinkCourDel *mock.MockCitilinkCourierDelivery) {
				ordr.EXPECT().Delivery().Times(2).Return(dlvr)
				dlvr.EXPECT().Id().Return(order.DeliveryIdCitilinkCourier)
				dlvr.EXPECT().CitilinkCourier(gomock.Any()).Return(citilinkCourDel)
				citilinkCourDel.EXPECT().Id().Return(order.CitilinkCourierDeliveryIdFast)
			},
			want: []order.AllowStatus{
				order.AllowStatusAllow,
				order.AllowStatusLimited,
			},
			payment: []order.PaymentId{
				order.PaymentIdCash,
				order.PaymentIdYandex,
			},
			err: nil,
		},
		{
			name: "not courier delivery",
			args: args{
				ctx: context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:         payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
				},
			},
			init: func(ordr *order.MockOrder, dlvr *mock.MockDelivery, citilinkCourDel *mock.MockCitilinkCourierDelivery) {
				ordr.EXPECT().Delivery().Times(1).Return(dlvr)
				dlvr.EXPECT().Id().Return(order.DeliveryIdSelf)
			},
			want:    []order.AllowStatus{},
			payment: []order.PaymentId{},
			err:     nil,
		},
		{
			name: "not fast delivery",
			args: args{
				ctx: context.TODO(),
				resolvedIdsMap: payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:         payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdInstallments: payment.NewResolvedPaymentId(order.PaymentIdInstallments),
				},
			},
			init: func(ordr *order.MockOrder, dlvr *mock.MockDelivery, citilinkCourDel *mock.MockCitilinkCourierDelivery) {
				ordr.EXPECT().Delivery().Times(2).Return(dlvr)
				dlvr.EXPECT().Id().Return(order.DeliveryIdCitilinkCourier)
				dlvr.EXPECT().CitilinkCourier(gomock.Any()).Return(citilinkCourDel)
				citilinkCourDel.EXPECT().Id().Return(order.CitilinkCourierDeliveryIdSameDay)
			},
			want:    []order.AllowStatus{},
			payment: []order.PaymentId{},
			err:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			ordr := order.NewMockOrder(ctrl)
			dlvr := mock.NewMockDelivery(ctrl)
			citilinkCourDel := mock.NewMockCitilinkCourierDelivery(ctrl)

			tt.init(ordr, dlvr, citilinkCourDel)

			r := &FastCitilinkCourierDelivery{}
			err := r.Resolve(tt.args.ctx, tt.args.resolvedIdsMap, ordr)
			assert.Equal(t, tt.err, err)
			for i, status := range tt.want {
				assert.Equal(t, status, tt.args.resolvedIdsMap[tt.payment[i]].Status())
			}
		})
	}
}
