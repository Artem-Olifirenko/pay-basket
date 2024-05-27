package resolver

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
	"testing"
)

func TestAnonUser_Resolve(t *testing.T) {
	type args struct {
		resolvedIdsMap payment.ResolvedPaymentIdMap
		user           *userv1.User
	}

	tests := []struct {
		name string
		args args
		want func() payment.ResolvedPaymentIdMap
	}{
		{
			// На вход подаётся набор оплат, где есть типы оплат, доступные и недоступные анонимному пользователю, и ордер, где юзер - НЕ анонимный; на выходе должен быть набор оплат равный набору оплат на входе (т.к. не анонимный пользователь никак не должен влиять на процесс в данном файле)
			"non-anon user does not affect payment processing",
			args{
				payment.ResolvedPaymentIdMap{
					order.PaymentIdWebmoney:    payment.NewResolvedPaymentId(order.PaymentIdWebmoney),
					order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
				},
				&userv1.User{},
			},
			func() payment.ResolvedPaymentIdMap {
				return payment.ResolvedPaymentIdMap{
					order.PaymentIdWebmoney:    payment.NewResolvedPaymentId(order.PaymentIdWebmoney),
					order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
				}
			},
		},
		{
			// На вход подаётся набор оплат, где есть типы оплат, доступные и недоступные анонимному пользователю и ордер, где юзер - анонимный; на выходе должен быть набор оплат НЕ равный набору оплат на входе (проверяем, что анонимный пользователь влияет на набор оплат, где есть типы оплат доступные и недоступные анонимному пользователю)
			"anon user limits non-anon payment type",
			args{
				payment.ResolvedPaymentIdMap{
					order.PaymentIdWebmoney:    payment.NewResolvedPaymentId(order.PaymentIdWebmoney),
					order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
				},
				nil,
			},
			func() payment.ResolvedPaymentIdMap {
				ret := payment.ResolvedPaymentIdMap{
					order.PaymentIdWebmoney:    payment.NewResolvedPaymentId(order.PaymentIdWebmoney),
					order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
				}
				ret[order.PaymentIdWebmoney].Disallow(order.AllowStatusLimited,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdAnonUser, order.SubsystemUnknown))
				return ret
			},
		},
		{
			// На вход подаётся набор оплат, где все типы оплат доступны анонимному пользователю и ордер, где юзер - анонимный; на выходе должен быть набор оплат равный набору оплат на входе (проверяем что анонимный пользователь не влияет на набор оплат, если такой набор состоит только из типов оплат доступных анонимному пользователю)
			"anon user does not limit payment types which are allowed for anon user",
			args{
				payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:        payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
				},
				nil,
			},
			func() payment.ResolvedPaymentIdMap {
				return payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:        payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCardsOnline: payment.NewResolvedPaymentId(order.PaymentIdCardsOnline),
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewAnonUser()
			ordr := order.NewMockOrder(gomock.NewController(t))
			ordr.EXPECT().User().Return(test.args.user).Times(1)
			err := resolver.Resolve(context.TODO(), test.args.resolvedIdsMap, ordr)
			got := test.args.resolvedIdsMap
			want := test.want()

			assert.NoError(t, err)
			assert.Equal(t, got, want)
		})
	}
}
