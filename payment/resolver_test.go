package payment

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/citizap/factory"
	citizap_factory_ctx "go.citilink.cloud/citizap/factory/ctx"
	simplestorage "go.citilink.cloud/libsimple-storage/v2"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/metrics"
	"go.citilink.cloud/order/internal/order"
	"go.citilink.cloud/order/internal/order/mock"
	"go.uber.org/zap"
	"gopkg.in/vmihailenco/msgpack.v2"
	"os"
	"testing"
)

var mtrcs1 *metrics.Metrics

func TestMain(m *testing.M) {
	mtrcs1 = metrics.New()
	os.Exit(m.Run())
}

func TestCompositeResolver_Add(t *testing.T) {
	type fields struct {
		resolvers     func() []Resolver
		cache         func(ctrl *gomock.Controller) simplestorage.Storage
		loggerFactory func() factory.Factory
		metrics       func() *metrics.Metrics
	}
	type args struct {
		resolver []Resolver
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []Resolver
	}{
		{
			name: "add resolver",
			fields: fields{
				resolvers: func() []Resolver {
					return []Resolver{}
				},
				cache: func(ctrl *gomock.Controller) simplestorage.Storage {
					return mock.NewMockStorage(ctrl)
				},
				loggerFactory: func() factory.Factory {
					logger := zap.NewNop()
					loggerFactory := citizap_factory_ctx.New(logger)
					return loggerFactory
				},
				metrics: func() *metrics.Metrics {
					return mtrcs1
				},
			},
			args: args{
				resolver: []Resolver{
					nil,
				},
			},
			want: []Resolver{
				nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CompositeResolver{
				resolvers:     tt.fields.resolvers(),
				cache:         tt.fields.cache(gomock.NewController(t)),
				loggerFactory: tt.fields.loggerFactory(),
				metrics:       tt.fields.metrics(),
			}
			r.Add(tt.args.resolver...)
			assert.Equalf(t, tt.args.resolver, tt.want, "Add()")
		})
	}
}

func TestCompositeResolver_Resolve(t *testing.T) {
	type fields struct {
		resolvers     func(ctrl *gomock.Controller) []Resolver
		cache         func(ctrl *gomock.Controller) simplestorage.Storage
		loggerFactory func() factory.Factory
		metrics       func() *metrics.Metrics
	}
	// ctrl := gomock.NewController(t)
	type args struct {
		ctx  context.Context
		ordr func(ctrl *gomock.Controller) order.Order
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ResolvedPaymentIdMap
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "error while getting basket",
			fields: fields{
				resolvers: func(ctrl *gomock.Controller) []Resolver {
					return []Resolver{}
				},
				cache: func(ctrl *gomock.Controller) simplestorage.Storage {
					return mock.NewMockStorage(ctrl)
				},
				loggerFactory: func() factory.Factory {
					logger := zap.NewNop()
					loggerFactory := citizap_factory_ctx.New(logger)
					return loggerFactory
				},
				metrics: func() *metrics.Metrics {
					return mtrcs1
				},
			},
			args: args{
				ctx: context.Background(),
				ordr: func(ctrl *gomock.Controller) order.Order {
					o := order.NewMockOrder(ctrl)
					o.EXPECT().Basket(gomock.Any()).Return(nil, fmt.Errorf("some error"))
					return o
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "use has not added items",
			fields: fields{
				resolvers: func(ctrl *gomock.Controller) []Resolver {
					return []Resolver{}
				},
				cache: func(ctrl *gomock.Controller) simplestorage.Storage {
					return mock.NewMockStorage(ctrl)
				},
				loggerFactory: func() factory.Factory {
					logger := zap.NewNop()
					loggerFactory := citizap_factory_ctx.New(logger)
					return loggerFactory
				},
				metrics: func() *metrics.Metrics {
					return mtrcs1
				},
			},
			args: args{
				ctx: context.Background(),
				ordr: func(ctrl *gomock.Controller) order.Order {
					b := order.NewMockBasket(ctrl)
					b.EXPECT().HasUserAddedItems().Times(1).Return(false)
					o := order.NewMockOrder(ctrl)
					o.EXPECT().Basket(gomock.Any()).Return(b, nil)
					return o
				},
			},
			want: func() ResolvedPaymentIdMap {
				resolvedIdMap := ResolvedPaymentIdMap{}
				for _, id := range order.PaymentIds {
					resolvedIdMap[id] = NewResolvedPaymentId(id)
				}
				return resolvedIdMap
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "error while getting cache",
			fields: fields{
				resolvers: func(ctrl *gomock.Controller) []Resolver {
					return []Resolver{}
				},
				cache: func(ctrl *gomock.Controller) simplestorage.Storage {
					s := mock.NewMockStorage(ctrl)
					s.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error"))
					s.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

					return s
				},
				loggerFactory: func() factory.Factory {
					logger := zap.NewNop()
					loggerFactory := citizap_factory_ctx.New(logger)
					return loggerFactory
				},
				metrics: func() *metrics.Metrics {
					return mtrcs1
				},
			},
			args: args{
				ctx: context.Background(),
				ordr: func(ctrl *gomock.Controller) order.Order {
					b := order.NewMockBasket(ctrl)
					b.EXPECT().HasUserAddedItems().Times(1).Return(true)
					o := order.NewMockOrder(ctrl)
					o.EXPECT().Basket(gomock.Any()).Return(b, nil)
					o.EXPECT().Fingerprint().AnyTimes()
					o.EXPECT().CompileCacheKey(gomock.Any()).AnyTimes()
					return o
				},
			},
			want: func() ResolvedPaymentIdMap {
				resolvedIdMap := ResolvedPaymentIdMap{}
				for _, id := range order.PaymentIds {
					resolvedIdMap[id] = NewResolvedPaymentId(id)
				}
				return resolvedIdMap
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "return cached data",
			fields: fields{
				resolvers: func(ctrl *gomock.Controller) []Resolver {
					return []Resolver{}
				},
				cache: func(ctrl *gomock.Controller) simplestorage.Storage {
					s := mock.NewMockStorage(ctrl)

					buf := bytes.NewBuffer(nil)
					encoder := msgpack.NewEncoder(buf)
					resolvedIdMap := ResolvedPaymentIdMap{}
					resolvedIdMap[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)
					err := encoder.Encode(&resolvedIdMap)
					assert.NoError(t, err)
					encodedResolvedIds := buf.Bytes()

					s.EXPECT().Get(gomock.Any(), gomock.Any()).
						Return(encodedResolvedIds, nil)
					return s
				},
				loggerFactory: func() factory.Factory {
					logger := zap.NewNop()
					loggerFactory := citizap_factory_ctx.New(logger)
					return loggerFactory
				},
				metrics: func() *metrics.Metrics {
					return mtrcs1
				},
			},
			args: args{
				ctx: context.Background(),
				ordr: func(ctrl *gomock.Controller) order.Order {
					b := order.NewMockBasket(ctrl)
					b.EXPECT().HasUserAddedItems().Times(1).Return(true)
					o := order.NewMockOrder(ctrl)
					o.EXPECT().Basket(gomock.Any()).Return(b, nil)
					o.EXPECT().Fingerprint().AnyTimes()
					o.EXPECT().CompileCacheKey(gomock.Any()).AnyTimes()
					return o
				},
			},
			want: func() ResolvedPaymentIdMap {
				resolvedIdMap := ResolvedPaymentIdMap{}
				resolvedIdMap[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)
				return resolvedIdMap
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "error while resolving payment ids",
			fields: fields{
				resolvers: func(ctrl *gomock.Controller) []Resolver {
					r1 := NewMockResolver(ctrl)
					r1.EXPECT().Resolve(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("some error"))
					return []Resolver{
						r1,
					}
				},
				cache: func(ctrl *gomock.Controller) simplestorage.Storage {
					s := mock.NewMockStorage(ctrl)

					resolvedIdMap := ResolvedPaymentIdMap{}
					resolvedIdMap[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)

					s.EXPECT().Get(gomock.Any(), gomock.Any()).
						Return(nil, internal.NewNotFoundError(fmt.Errorf("some error")))
					return s
				},
				loggerFactory: func() factory.Factory {
					logger := zap.NewNop()
					loggerFactory := citizap_factory_ctx.New(logger)
					return loggerFactory
				},
				metrics: func() *metrics.Metrics {
					return mtrcs1
				},
			},
			args: args{
				ctx: context.Background(),
				ordr: func(ctrl *gomock.Controller) order.Order {
					b := order.NewMockBasket(ctrl)
					b.EXPECT().HasUserAddedItems().Times(1).Return(true)
					o := order.NewMockOrder(ctrl)
					o.EXPECT().Basket(gomock.Any()).Return(b, nil)
					o.EXPECT().Fingerprint().AnyTimes()
					o.EXPECT().CompileCacheKey(gomock.Any()).AnyTimes()
					return o
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "ok",
			fields: fields{
				resolvers: func(ctrl *gomock.Controller) []Resolver {
					r1 := NewMockResolver(ctrl)
					r1.EXPECT().Resolve(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
					return []Resolver{
						r1,
					}
				},
				cache: func(ctrl *gomock.Controller) simplestorage.Storage {
					s := mock.NewMockStorage(ctrl)

					resolvedIdMap := ResolvedPaymentIdMap{}
					resolvedIdMap[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)

					s.EXPECT().Get(gomock.Any(), gomock.Any()).
						Return(nil, internal.NewNotFoundError(fmt.Errorf("some error")))

					s.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("some error"))
					return s
				},
				loggerFactory: func() factory.Factory {
					logger := zap.NewNop()
					loggerFactory := citizap_factory_ctx.New(logger)
					return loggerFactory
				},
				metrics: func() *metrics.Metrics {
					return mtrcs1
				},
			},
			args: args{
				ctx: context.Background(),
				ordr: func(ctrl *gomock.Controller) order.Order {
					b := order.NewMockBasket(ctrl)
					b.EXPECT().HasUserAddedItems().Times(1).Return(true)
					o := order.NewMockOrder(ctrl)
					o.EXPECT().Basket(gomock.Any()).Return(b, nil)
					o.EXPECT().Fingerprint().AnyTimes()
					o.EXPECT().CompileCacheKey(gomock.Any()).AnyTimes()
					return o
				},
			},
			want: func() ResolvedPaymentIdMap {
				resolvedIdMap := ResolvedPaymentIdMap{}
				for _, id := range order.PaymentIds {
					resolvedIdMap[id] = NewResolvedPaymentId(id)
				}
				return resolvedIdMap
			}(),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		t.Run(tt.name, func(t *testing.T) {
			r := &CompositeResolver{
				resolvers:     tt.fields.resolvers(ctrl),
				cache:         tt.fields.cache(ctrl),
				loggerFactory: tt.fields.loggerFactory(),
				metrics:       tt.fields.metrics(),
			}
			o := tt.args.ordr(ctrl)
			got, err := r.Resolve(tt.args.ctx, o)

			if !tt.wantErr(t, err, fmt.Sprintf("Resolve()")) {
				return
			}
			assert.Equalf(t, tt.want, got, "Resolve()")
		})
	}
}

func TestNewResolvedPaymentId(t *testing.T) {
	type args struct {
		id order.PaymentId
	}
	tests := []struct {
		name string
		args args
		want *ResolvedId
	}{
		{
			name: "creation of new resolved payment id",
			args: args{
				id: order.PaymentIdWebmoney,
			},
			want: &ResolvedId{
				id:     order.PaymentIdWebmoney,
				status: order.AllowStatusAllow,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewResolvedPaymentId(tt.args.id), "NewResolvedPaymentId(%v)", tt.args.id)
		})
	}
}

func TestNewResolver(t *testing.T) {
	type args struct {
		resolvers     []Resolver
		cache         simplestorage.Storage
		loggerFactory factory.Factory
		metrics       *metrics.Metrics
	}

	tests := []struct {
		name string
		args args
		want *CompositeResolver
	}{
		{
			name: "creation ok",
			args: args{
				resolvers:     nil,
				cache:         nil,
				loggerFactory: nil,
				metrics:       nil,
			},
			want: &CompositeResolver{
				resolvers: nil,
				cache: func() simplestorage.Storage {
					return simplestorage.WrapPrefixed("payment.CompositeResolver", nil)
				}(),
				loggerFactory: nil,
				metrics:       nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewResolver(tt.args.cache, tt.args.loggerFactory, tt.args.metrics), "NewResolver(%v, %v, %v)", tt.args.cache, tt.args.loggerFactory, tt.args.metrics)
		})
	}
}

func TestResolvedId_Disallow(t *testing.T) {
	type fields struct {
		id        order.PaymentId
		status    order.AllowStatus
		reasons   []*order.DisallowReasonWithInfo
		isDefault bool
		isChosen  bool
	}
	type args struct {
		status order.AllowStatus
		info   *order.DisallowReasonWithInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "ok",
			fields: fields{
				id:        order.PaymentIdWebmoney,
				status:    order.AllowStatusAllow,
				reasons:   nil,
				isDefault: false,
				isChosen:  false,
			},
			args: args{
				status: order.AllowStatusLimited,
				info:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &ResolvedId{
				id:        tt.fields.id,
				status:    tt.fields.status,
				reasons:   tt.fields.reasons,
				isDefault: tt.fields.isDefault,
				isChosen:  tt.fields.isChosen,
			}
			i.Disallow(tt.args.status, tt.args.info)
		})
	}
}

func TestResolvedId_Id(t *testing.T) {
	type fields struct {
		id        order.PaymentId
		status    order.AllowStatus
		reasons   []*order.DisallowReasonWithInfo
		isDefault bool
		isChosen  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   order.PaymentId
	}{
		{
			name: "ok",
			fields: fields{
				id:        order.PaymentIdWebmoney,
				status:    0,
				reasons:   nil,
				isDefault: false,
				isChosen:  false,
			},
			want: order.PaymentIdWebmoney,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &ResolvedId{
				id:        tt.fields.id,
				status:    tt.fields.status,
				reasons:   tt.fields.reasons,
				isDefault: tt.fields.isDefault,
				isChosen:  tt.fields.isChosen,
			}
			assert.Equalf(t, tt.want, i.Id(), "Id()")
		})
	}
}

func TestResolvedId_IsChosen(t *testing.T) {
	type fields struct {
		id        order.PaymentId
		status    order.AllowStatus
		reasons   []*order.DisallowReasonWithInfo
		isDefault bool
		isChosen  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "ok",
			fields: fields{
				id:        0,
				status:    0,
				reasons:   nil,
				isDefault: false,
				isChosen:  false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &ResolvedId{
				id:        tt.fields.id,
				status:    tt.fields.status,
				reasons:   tt.fields.reasons,
				isDefault: tt.fields.isDefault,
				isChosen:  tt.fields.isChosen,
			}
			assert.Equalf(t, tt.want, i.IsChosen(), "IsChosen()")
		})
	}
}

func TestResolvedId_IsDefault(t *testing.T) {
	type fields struct {
		id        order.PaymentId
		status    order.AllowStatus
		reasons   []*order.DisallowReasonWithInfo
		isDefault bool
		isChosen  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "ok",
			fields: fields{
				id:        0,
				status:    0,
				reasons:   nil,
				isDefault: false,
				isChosen:  false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &ResolvedId{
				id:        tt.fields.id,
				status:    tt.fields.status,
				reasons:   tt.fields.reasons,
				isDefault: tt.fields.isDefault,
				isChosen:  tt.fields.isChosen,
			}
			assert.Equalf(t, tt.want, i.IsDefault(), "IsDefault()")
		})
	}
}

func TestResolvedId_Reasons(t *testing.T) {
	type fields struct {
		id        order.PaymentId
		status    order.AllowStatus
		reasons   []*order.DisallowReasonWithInfo
		isDefault bool
		isChosen  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   []*order.DisallowReasonWithInfo
	}{
		{
			name: "ok",
			fields: fields{
				id:        0,
				status:    0,
				reasons:   nil,
				isDefault: false,
				isChosen:  false,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &ResolvedId{
				id:        tt.fields.id,
				status:    tt.fields.status,
				reasons:   tt.fields.reasons,
				isDefault: tt.fields.isDefault,
				isChosen:  tt.fields.isChosen,
			}
			assert.Equalf(t, tt.want, i.Reasons(), "Reasons()")
		})
	}
}

func TestResolvedId_Status(t *testing.T) {
	type fields struct {
		id        order.PaymentId
		status    order.AllowStatus
		reasons   []*order.DisallowReasonWithInfo
		isDefault bool
		isChosen  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   order.AllowStatus
	}{
		{
			name: "ok",
			fields: fields{
				id:        0,
				status:    0,
				reasons:   nil,
				isDefault: false,
				isChosen:  false,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &ResolvedId{
				id:        tt.fields.id,
				status:    tt.fields.status,
				reasons:   tt.fields.reasons,
				isDefault: tt.fields.isDefault,
				isChosen:  tt.fields.isChosen,
			}
			assert.Equalf(t, tt.want, i.Status(), "Status()")
		})
	}
}

func TestResolvedPaymentIdMap_DisallowMany(t *testing.T) {
	type args struct {
		ids    []order.PaymentId
		status order.AllowStatus
		info   *order.DisallowReasonWithInfo
	}
	tests := []struct {
		name string
		r    ResolvedPaymentIdMap
		args args
	}{
		{
			name: "disallow many",
			r: func() ResolvedPaymentIdMap {
				resolvedIdMap := ResolvedPaymentIdMap{}
				resolvedIdMap[order.PaymentIdCash] = NewResolvedPaymentId(order.PaymentIdCash)
				resolvedIdMap[order.PaymentIdCredit] = NewResolvedPaymentId(order.PaymentIdCredit)
				resolvedIdMap[order.PaymentIdWebmoney] = NewResolvedPaymentId(order.PaymentIdWebmoney)
				return resolvedIdMap
			}(),
			args: args{
				ids: []order.PaymentId{
					order.PaymentIdCash,
					order.PaymentIdCredit,
					order.PaymentIdWebmoney,
				},
				status: order.AllowStatusAllow,
				info:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.DisallowMany(tt.args.ids, tt.args.status, tt.args.info)
		})
	}
}
