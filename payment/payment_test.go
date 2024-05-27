package payment

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
	"testing"
)

func TestFactory_Create(t *testing.T) {
	type fields struct {
		resolver CompositePaymentResolver
	}
	type args struct {
		order order.Order
		st    State
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Payment
	}{
		{
			name: "test payment creation",
			fields: fields{
				resolver: nil,
			},
			args: args{
				order: nil,
				st:    nil,
			},
			want: &Payment{
				order:    nil,
				resolver: nil,
				state:    nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Factory{
				resolver: tt.fields.resolver,
			}
			assert.Equalf(t, tt.want, f.Create(tt.args.order, tt.args.st), "Create(%v, %v)", tt.args.order, tt.args.st)
		})
	}
}

func TestNewFactory(t *testing.T) {
	type args struct {
		resolver CompositePaymentResolver
	}
	tests := []struct {
		name string
		args args
		want *Factory
	}{
		{
			name: "factory creation",
			args: args{
				resolver: nil,
			},
			want: &Factory{
				resolver: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewFactory(tt.args.resolver), "NewFactory(%v)", tt.args.resolver)
		})
	}
}

func TestNewPayment(t *testing.T) {
	type args struct {
		order    order.Order
		resolver CompositePaymentResolver
		st       State
	}
	tests := []struct {
		name string
		args args
		want *Payment
	}{
		{
			name: "payment creation",
			args: args{
				order:    nil,
				resolver: nil,
				st:       nil,
			},
			want: &Payment{
				order:    nil,
				resolver: nil,
				state:    nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewPayment(tt.args.order, tt.args.resolver, tt.args.st), "NewPayment(%v, %v, %v)", tt.args.order, tt.args.resolver, tt.args.st)
		})
	}
}

func TestPayment_Check(t *testing.T) {

	type args struct {
		ctx context.Context
		id  order.PaymentId
	}
	tests := []struct {
		name                  string
		prepare               func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment)
		args                  args
		wantIsValid           bool
		wantErr               error
		wantResolvedPaymentId order.ResolvedPaymentId
	}{
		{
			name: "ids resolved with err",
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {
				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, fmt.Errorf("some error"))
			},
			args: args{
				ctx: context.Background(),
				id:  order.PaymentIdCash,
			},
			wantIsValid:           false,
			wantErr:               fmt.Errorf("can't resolve ids: %w", fmt.Errorf("can't resolve ids: %w", fmt.Errorf("some error"))),
			wantResolvedPaymentId: nil,
		},
		{
			name: "payment id of order is invalid (resolved id is not found)",
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {
				mockState.EXPECT().PaymentIdB2B().Return(31).Times(1)
				mockOrder.EXPECT().User().Return(&userv1.User{
					B2B: &userv1.User_B2B{
						IsB2BState: true,
					},
				}).Times(1)

				mockOrder.EXPECT().Payment().Return(p).Times(1)

				payments := make(ResolvedPaymentIdMap)
				payments[order.PaymentIdCash] = NewResolvedPaymentId(order.PaymentIdCash)
				payments[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)

				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Times(1).
					Return(payments, nil)
			},
			args: args{
				ctx: context.Background(),
				id:  order.PaymentIdInstallments,
			},
			wantIsValid:           false,
			wantErr:               nil,
			wantResolvedPaymentId: nil,
		},
		{
			name: "payment id of order is invalid",
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {

				mockState.EXPECT().PaymentIdB2C().Return(4).Times(1)

				mockOrder.EXPECT().User().Return(&userv1.User{
					B2B: &userv1.User_B2B{
						IsB2BState: false,
					},
				}).Times(1)

				mockOrder.EXPECT().Payment().Return(p).Times(1)

				payments := make(ResolvedPaymentIdMap)
				payments[order.PaymentIdCash] = NewResolvedPaymentId(order.PaymentIdCash)
				payments[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)
				payments[order.PaymentIdCredit].Disallow(order.AllowStatusDisallowed, nil)

				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Times(1).
					Return(payments, nil)
			},
			args: args{
				ctx: context.Background(),
				id:  order.PaymentIdCredit,
			},
			wantIsValid: false,
			wantErr:     nil,
			wantResolvedPaymentId: func() order.ResolvedPaymentId {
				p := NewResolvedPaymentId(order.PaymentIdCredit)
				p.Disallow(order.AllowStatusDisallowed, nil)
				return p
			}(),
		},
		{
			name: "payment id of order is valid",
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {

				mockState.EXPECT().PaymentIdB2C().Return(4).Times(1)

				mockOrder.EXPECT().User().Return(&userv1.User{
					B2B: &userv1.User_B2B{
						IsB2BState: false,
					},
				}).Times(1)

				mockOrder.EXPECT().Payment().Return(p).Times(1)

				payments := make(ResolvedPaymentIdMap)
				payments[order.PaymentIdCash] = NewResolvedPaymentId(order.PaymentIdCash)
				payments[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)

				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Times(1).
					Return(payments, nil)
			},
			args: args{
				ctx: context.Background(),
				id:  order.PaymentIdCredit,
			},
			wantIsValid: true,
			wantErr:     nil,
			wantResolvedPaymentId: func() order.ResolvedPaymentId {
				p := NewResolvedPaymentId(order.PaymentIdCredit)
				p.isChosen = true
				return p
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockOrder := order.NewMockOrder(ctrl)
			mockCompositeResolver := NewMockCompositePaymentResolver(ctrl)
			mockState := NewMockState(ctrl)

			p := NewPayment(mockOrder, mockCompositeResolver, mockState)
			tt.prepare(mockOrder, mockCompositeResolver, mockState, p)

			gotIsValid, gotErr, gotResolvedPaymentId := p.Check(tt.args.ctx, tt.args.id)

			assert.Equalf(t, tt.wantIsValid, gotIsValid, "Check(%v)", tt.args.id)
			assert.Equalf(t, tt.wantErr, gotErr, "Check(%v)", tt.args.id)
			assert.Equalf(t, tt.wantResolvedPaymentId, gotResolvedPaymentId, "Check(%v)", tt.args.id)
		})
	}
}

func TestPayment_Problems(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment)
		want    []*order.Problem
		wantErr error
	}{
		{
			name: "no need to feel",
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {
				mockState.EXPECT().PaymentIdB2C().Return(0).Times(1)
				mockOrder.EXPECT().User().Return(&userv1.User{
					B2B: &userv1.User_B2B{
						IsB2BState: false,
					},
				}).Times(1)

				payments := make(ResolvedPaymentIdMap)
				payments[order.PaymentIdCash] = NewResolvedPaymentId(order.PaymentIdCash)
			},
			wantErr: nil,
			want:    nil,
		},
		{
			name: "payment ids resolved with err",
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {
				mockState.EXPECT().PaymentIdB2C().Return(4).Times(2)
				mockOrder.EXPECT().User().Return(&userv1.User{
					B2B: &userv1.User_B2B{
						IsB2BState: false,
					},
				}).Times(2)

				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, fmt.Errorf("some error"))
			},
			wantErr: fmt.Errorf("can't check allowed payment: %w", fmt.Errorf("can't resolve ids: %w", fmt.Errorf("can't resolve ids: %w", fmt.Errorf("some error")))),
			want:    nil,
		},
		{
			name: "payment ids resolved correctly",
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {
				mockState.EXPECT().PaymentIdB2C().Return(4).Times(3)
				mockOrder.EXPECT().User().Return(&userv1.User{
					B2B: &userv1.User_B2B{
						IsB2BState: false,
					},
				}).Times(3)

				mockOrder.EXPECT().Payment().Return(p).Times(1)

				payments := make(ResolvedPaymentIdMap)
				payments[order.PaymentIdCash] = NewResolvedPaymentId(order.PaymentIdCash)
				payments[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)

				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Times(1).
					Return(payments, nil)
			},
			wantErr: nil,
			want:    nil,
		},
		{
			name: "payment ids resolved with problems",
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {
				mockState.EXPECT().PaymentIdB2C().Return(4).Times(4)
				mockOrder.EXPECT().User().Return(&userv1.User{
					B2B: &userv1.User_B2B{
						IsB2BState: false,
					},
				}).Times(4)

				mockOrder.EXPECT().Payment().Return(p).Times(1)

				payments := make(ResolvedPaymentIdMap)
				payments[order.PaymentIdCash] = NewResolvedPaymentId(order.PaymentIdCash)
				payments[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)
				payments[order.PaymentIdCredit].Disallow(order.AllowStatusDisallowed, nil)

				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Times(1).
					Return(payments, nil)
			},
			wantErr: nil,
			want: func() []*order.Problem {
				p := NewResolvedPaymentId(order.PaymentIdCredit)
				pr := order.NewProblem(
					order.ProblemIdPaymentIdNotAllowed,
					fmt.Sprintf("Тип оплаты %s недоступен", p.Id().Name()),
					order.SubsystemPaymentMethod,
				)
				return []*order.Problem{pr}
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockOrder := order.NewMockOrder(ctrl)
			mockCompositeResolver := NewMockCompositePaymentResolver(ctrl)
			mockState := NewMockState(ctrl)
			p := NewPayment(mockOrder, mockCompositeResolver, mockState)
			tt.prepare(mockOrder, mockCompositeResolver, mockState, p)

			got, err := p.Problems(context.Background())

			// tt.wantErr(t, err, fmt.Sprintf("Problems(%v)", tt.name))
			assert.Equalf(t, tt.wantErr, err, "Errors comparison")
			assert.Equalf(t, tt.want, got, "Problems(%v)", tt.name)
		})
	}
}

func TestPayment_SetId(t *testing.T) {

	type args struct {
		ctx context.Context
		id  order.PaymentId
	}

	tests := []struct {
		name    string
		args    args
		prepare func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment)
		want    bool
		want1   error
		want2   order.ResolvedPaymentId
	}{
		{
			name: "an error occurred",
			args: args{
				ctx: context.Background(),
				id:  0,
			},
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {
				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, fmt.Errorf("some error"))
			},
			want:  false,
			want1: fmt.Errorf("can't validate id: %w", fmt.Errorf("can't resolve ids: %w", fmt.Errorf("can't resolve ids: %w", fmt.Errorf("some error")))),
			want2: nil,
		},
		{
			name: "set failed",
			args: args{
				ctx: context.Background(),
				id:  order.PaymentIdCredit,
			},
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {
				mockState.EXPECT().PaymentIdB2C().Return(4).Times(1)

				mockOrder.EXPECT().User().Return(&userv1.User{
					B2B: &userv1.User_B2B{
						IsB2BState: false,
					},
				}).Times(1)

				mockOrder.EXPECT().Payment().Return(p).Times(1)

				payments := make(ResolvedPaymentIdMap)
				payments[order.PaymentIdCash] = NewResolvedPaymentId(order.PaymentIdCash)
				payments[order.PaymentIdCash].Disallow(order.AllowStatusDisallowed, nil)
				payments[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)
				payments[order.PaymentIdCredit].Disallow(order.AllowStatusDisallowed, nil)

				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Times(1).
					Return(payments, nil)
			},
			want:  false,
			want1: nil,
			want2: func() order.ResolvedPaymentId {
				p := NewResolvedPaymentId(order.PaymentIdCredit)
				p.Disallow(order.AllowStatusDisallowed, nil)
				return p
			}(),
		},
		{
			name: "set ok for B2C",
			args: args{
				ctx: context.Background(),
				id:  order.PaymentIdCredit,
			},
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {
				mockState.EXPECT().PaymentIdB2C().Return(4).Times(1)
				mockState.EXPECT().SetPaymentIdB2C(gomock.Any()).Times(1)

				mockOrder.EXPECT().User().Return(&userv1.User{
					B2B: &userv1.User_B2B{
						IsB2BState: false,
					},
				}).Times(2)

				mockOrder.EXPECT().Payment().Return(p).Times(1)

				payments := make(ResolvedPaymentIdMap)
				payments[order.PaymentIdCash] = NewResolvedPaymentId(order.PaymentIdCash)
				payments[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)

				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Times(1).
					Return(payments, nil)
			},
			want:  true,
			want1: nil,
			want2: nil,
		},
		{
			name: "set ok for B2B",
			args: args{
				ctx: context.Background(),
				id:  order.PaymentIdCredit,
			},
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {
				mockState.EXPECT().PaymentIdB2B().Return(30).Times(1)
				mockState.EXPECT().SetPaymentIdB2B(gomock.Any()).Times(1)

				mockOrder.EXPECT().User().Return(&userv1.User{
					B2B: &userv1.User_B2B{
						IsB2BState: true,
					},
				}).Times(2)

				mockOrder.EXPECT().Payment().Return(p).Times(1)

				payments := make(ResolvedPaymentIdMap)
				payments[order.PaymentIdCash] = NewResolvedPaymentId(order.PaymentIdCash)
				payments[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)

				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), gomock.Any()).
					Times(1).
					Return(payments, nil)
			},
			want:  true,
			want1: nil,
			want2: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockOrder := order.NewMockOrder(ctrl)
			mockCompositeResolver := NewMockCompositePaymentResolver(ctrl)
			mockState := NewMockState(ctrl)

			p := NewPayment(mockOrder, mockCompositeResolver, mockState)
			tt.prepare(mockOrder, mockCompositeResolver, mockState, p)

			got, got1, got2 := p.SetId(tt.args.ctx, tt.args.id)
			assert.Equalf(t, tt.want, got, "SetId(%v)", tt.args.id)
			assert.Equalf(t, tt.want1, got1, "SetId(%v)", tt.args.id)
			assert.Equalf(t, tt.want2, got2, "SetId(%v)", tt.args.id)
		})
	}
}

func TestPayment_ResolveIds(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name    string
		args    args
		prepare func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment)
		want    func() ([]order.ResolvedPaymentId, error)
	}{
		{
			name: "correct",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {
				payments := make(ResolvedPaymentIdMap)
				payments[order.PaymentIdCash] = NewResolvedPaymentId(order.PaymentIdCash)
				payments[order.PaymentIdCardsOnline] = NewResolvedPaymentId(order.PaymentIdCardsOnline)
				payments[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)

				mockOrder.EXPECT().Payment().Return(&Payment{
					order:    mockOrder,
					resolver: mockCompositeResolver,
					state:    mockState,
				})

				mockOrder.EXPECT().User().Return(&userv1.User{
					Id: "IL123",
				})

				mockState.EXPECT().PaymentIdB2C().Return(2)

				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), mockOrder).
					Times(1).
					Return(payments, nil)

			},
			want: func() ([]order.ResolvedPaymentId, error) {
				return []order.ResolvedPaymentId{
					&ResolvedId{
						id:        order.PaymentIdCash,
						isDefault: true,
						isChosen:  true,
					},
					&ResolvedId{
						id: order.PaymentIdCardsOnline,
					},
					&ResolvedId{
						id: order.PaymentIdCredit,
					},
				}, nil
			},
		},
		{
			name: "error",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(mockOrder *order.MockOrder, mockCompositeResolver *MockCompositePaymentResolver, mockState *MockState, p order.Payment) {
				mockCompositeResolver.EXPECT().
					Resolve(gomock.Any(), mockOrder).
					Times(1).
					Return(nil, errors.New("no available payment types"))

			},
			want: func() ([]order.ResolvedPaymentId, error) {
				return nil, fmt.Errorf("can't resolve ids: %w", errors.New("no available payment types"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockOrder := order.NewMockOrder(ctrl)
			mockCompositeResolver := NewMockCompositePaymentResolver(ctrl)
			mockState := NewMockState(ctrl)

			p := NewPayment(mockOrder, mockCompositeResolver, mockState)
			tt.prepare(mockOrder, mockCompositeResolver, mockState, p)

			wantResult, wantErr := tt.want()

			result, err := p.ResolveIds(tt.args.ctx)
			assert.Equal(t, wantResult, result)
			assert.Equal(t, wantErr, err)
		})
	}
}
