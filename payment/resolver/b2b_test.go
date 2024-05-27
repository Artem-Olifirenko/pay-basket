package resolver

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	b2b2 "go.citilink.cloud/order/internal/b2b"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/payment"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
	"testing"
)

func TestNewB2b_Resolve(t *testing.T) {
	type args struct {
		isCashOrCardAvailable bool
		resolvedIdsMap        payment.ResolvedPaymentIdMap
		ordr                  func(ctrl *gomock.Controller) order.Order
	}

	tests := []struct {
		name string
		args args
		want func() payment.ResolvedPaymentIdMap
	}{
		{
			"user is b2b but cash or card is not available in general",
			args{
				false,
				payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
					order.PaymentIdCards:                  payment.NewResolvedPaymentId(order.PaymentIdCards),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
				},
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					usr := &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					}

					b2b := order.NewMockB2B(ctrl)
					b2b.EXPECT().Contractor(gomock.Any()).Return(nil, nil)

					ordr.EXPECT().B2B().Return(b2b)
					ordr.EXPECT().Cost(gomock.Any()).Return(nil, nil)
					ordr.EXPECT().User().Times(2).Return(usr)
					return ordr
				},
			},
			func() payment.ResolvedPaymentIdMap {
				ret := payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
					order.PaymentIdCards:                  payment.NewResolvedPaymentId(order.PaymentIdCards),
				}

				ret[order.PaymentIdCards].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdB2b, order.SubsystemPaymentMethod))
				ret[order.PaymentIdCashWithCard].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdB2b, order.SubsystemPaymentMethod))
				return ret
			},
		},
		{
			"user is b2b and payments are ONLY for b2b",
			args{
				true,
				payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
				},
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					usr := &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					}

					b2b := order.NewMockB2B(ctrl)
					b2b.EXPECT().Contractor(gomock.Any()).Return(nil, nil)

					ordr.EXPECT().B2B().Return(b2b)
					ordr.EXPECT().Cost(gomock.Any()).Return(nil, nil)
					ordr.EXPECT().User().Times(2).Return(usr)
					return ordr
				},
			},
			func() payment.ResolvedPaymentIdMap {
				return payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
				}
			},
		},
		{
			"user is b2b and all payments are not for b2b",
			args{
				true,
				payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:   payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCredit: payment.NewResolvedPaymentId(order.PaymentIdCredit),
				},
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					usr := &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					}
					b2b := order.NewMockB2B(ctrl)
					b2b.EXPECT().Contractor(gomock.Any()).Return(nil, nil)

					ordr.EXPECT().B2B().Return(b2b)
					ordr.EXPECT().User().Times(2).Return(usr)
					ordr.EXPECT().Cost(gomock.Any()).Return(nil, nil)
					return ordr
				},
			},
			func() payment.ResolvedPaymentIdMap {
				ret := payment.ResolvedPaymentIdMap{
					order.PaymentIdCash:   payment.NewResolvedPaymentId(order.PaymentIdCash),
					order.PaymentIdCredit: payment.NewResolvedPaymentId(order.PaymentIdCredit),
				}

				ret[order.PaymentIdCash].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdB2b, order.SubsystemUnknown))
				ret[order.PaymentIdCredit].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdB2b, order.SubsystemUnknown))
				return ret
			},
		},
		{
			"anon user",
			args{
				true,
				payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
				},
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)

					ordr.EXPECT().User().Times(2).Return(nil)
					ordr.EXPECT().Cost(gomock.Any()).Return(nil, nil)

					return ordr
				},
			},
			func() payment.ResolvedPaymentIdMap {
				ret := payment.ResolvedPaymentIdMap{
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
				}

				ret[order.PaymentIdCashless].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdB2b, order.SubsystemUnknown))
				ret[order.PaymentIdSberbankBusinessOnline].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdB2b, order.SubsystemUnknown))
				return ret
			},
		},
		{
			"user is b2b and order cost <= 100000",
			args{
				true,
				payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
					order.PaymentIdCards:                  payment.NewResolvedPaymentId(order.PaymentIdCards),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
				},
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					usr := &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					}
					b2b := order.NewMockB2B(ctrl)
					b2b.EXPECT().Contractor(gomock.Any()).Return(nil, nil)

					ordr.EXPECT().B2B().Return(b2b)
					ordr.EXPECT().User().Times(2).Return(usr)
					ordr.EXPECT().Cost(gomock.Any()).Return(&order.OrderCost{
						WithDiscount: 100000,
					}, nil).Times(1)
					return ordr
				},
			},
			func() payment.ResolvedPaymentIdMap {
				ret := payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
					order.PaymentIdCards:                  payment.NewResolvedPaymentId(order.PaymentIdCards),
				}
				ret[order.PaymentIdCards].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdMinCost, order.SubsystemPaymentMethod))
				return ret
			},
		},
		{
			"user is b2b and order cost > 100000",
			args{
				true,
				payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
					order.PaymentIdCards:                  payment.NewResolvedPaymentId(order.PaymentIdCards),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
				},
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					usr := &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					}
					b2b := order.NewMockB2B(ctrl)
					b2b.EXPECT().Contractor(gomock.Any()).Return(&b2b2.Contractor{
						Providers: []*b2b2.Provider{
							{
								IsMain:      false,
								IsSanctions: false,
							},
						},
					}, nil)

					ordr.EXPECT().B2B().Return(b2b)
					ordr.EXPECT().User().Times(2).Return(usr)
					ordr.EXPECT().Cost(gomock.Any()).Return(&order.OrderCost{
						WithDiscount: 100001,
					}, nil).Times(1)
					return ordr
				},
			},
			func() payment.ResolvedPaymentIdMap {
				ret := payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
					order.PaymentIdCards:                  payment.NewResolvedPaymentId(order.PaymentIdCards),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
				}
				ret[order.PaymentIdCashWithCard].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdMaxCost, order.SubsystemPaymentMethod))
				return ret
			},
		},
		{
			"user is b2b and contractor has sanctions",
			args{
				true,
				payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
					order.PaymentIdCards:                  payment.NewResolvedPaymentId(order.PaymentIdCards),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
				},
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					usr := &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					}
					b2b := order.NewMockB2B(ctrl)
					b2b.EXPECT().Contractor(gomock.Any()).Return(&b2b2.Contractor{
						Providers: []*b2b2.Provider{
							{
								IsMain:      true,
								IsSanctions: true,
							},
						},
					}, nil)

					ordr.EXPECT().B2B().Return(b2b)
					ordr.EXPECT().User().Times(2).Return(usr)
					ordr.EXPECT().Cost(gomock.Any()).Return(&order.OrderCost{
						WithDiscount: 100001,
					}, nil).Times(1)
					return ordr
				},
			},
			func() payment.ResolvedPaymentIdMap {
				ret := payment.ResolvedPaymentIdMap{
					order.PaymentIdCashless:               payment.NewResolvedPaymentId(order.PaymentIdCashless),
					order.PaymentIdSberbankBusinessOnline: payment.NewResolvedPaymentId(order.PaymentIdSberbankBusinessOnline),
					order.PaymentIdCards:                  payment.NewResolvedPaymentId(order.PaymentIdCards),
					order.PaymentIdCashWithCard:           payment.NewResolvedPaymentId(order.PaymentIdCashWithCard),
				}
				ret[order.PaymentIdCashWithCard].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdB2b, order.SubsystemPaymentMethod))
				ret[order.PaymentIdCards].Disallow(order.AllowStatusDisallowed,
					order.NewDisallowReasonWithInfo(order.DisallowReasonPaymentIdB2b, order.SubsystemPaymentMethod))
				return ret
			},
		},
		{
			"user is b2b and empty payments list",
			args{
				true,
				payment.ResolvedPaymentIdMap{},
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					usr := &userv1.User{
						B2B: &userv1.User_B2B{
							IsB2BState: true,
						},
					}
					b2b := order.NewMockB2B(ctrl)
					b2b.EXPECT().Contractor(gomock.Any()).Return(&b2b2.Contractor{
						Providers: []*b2b2.Provider{
							{
								IsMain:      false,
								IsSanctions: false,
							},
						},
					}, nil)

					ordr.EXPECT().B2B().Return(b2b)
					ordr.EXPECT().User().Times(2).Return(usr)
					ordr.EXPECT().Cost(gomock.Any()).Return(nil, nil)
					return ordr
				},
			},
			func() payment.ResolvedPaymentIdMap {
				return payment.ResolvedPaymentIdMap{}
			},
		},
		{
			"user is not b2b and empty payments list",
			args{
				true,
				payment.ResolvedPaymentIdMap{},
				func(ctrl *gomock.Controller) order.Order {
					ordr := order.NewMockOrder(ctrl)
					usr := &userv1.User{}
					ordr.EXPECT().User().Times(3).Return(usr)
					ordr.EXPECT().Cost(gomock.Any()).Return(nil, nil)
					return ordr
				},
			},
			func() payment.ResolvedPaymentIdMap {
				return payment.ResolvedPaymentIdMap{}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewB2b(test.args.isCashOrCardAvailable)

			err := resolver.Resolve(context.TODO(), test.args.resolvedIdsMap, test.args.ordr(gomock.NewController(t)))
			got := test.args.resolvedIdsMap
			want := test.want()

			assert.NoError(t, err)
			assert.Equal(t, got, want)
		})
	}
}
