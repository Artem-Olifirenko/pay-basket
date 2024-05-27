package payment

//go:generate mockgen -source=resolver.go -destination=./resolver_mock.go -package=payment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	citizap_factory "go.citilink.cloud/citizap/factory"
	simplestoragev2 "go.citilink.cloud/libsimple-storage/v2"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/metrics"
	"go.citilink.cloud/order/internal/order"
	"go.uber.org/zap"
	"gopkg.in/vmihailenco/msgpack.v2"
	"hash/fnv"
	"strconv"
)

type Resolver interface {
	Resolve(ctx context.Context, resolvedIdsMap ResolvedPaymentIdMap, order order.Order) error
}

type CompositePaymentResolver interface {
	Resolve(ctx context.Context, ordr order.Order) (ResolvedPaymentIdMap, error)
	Add(resolver ...Resolver)
}

func NewResolver(
	cache simplestoragev2.Storage,
	loggerFactory citizap_factory.Factory,
	metrics *metrics.Metrics,
) CompositePaymentResolver {
	return &CompositeResolver{
		cache:         simplestoragev2.WrapPrefixed("payment.CompositeResolver", cache),
		loggerFactory: loggerFactory,
		metrics:       metrics,
	}
}

type CompositeResolver struct {
	resolvers     []Resolver
	cache         simplestoragev2.Storage
	loggerFactory citizap_factory.Factory

	metrics *metrics.Metrics
}

func (r *CompositeResolver) Resolve(ctx context.Context, ordr order.Order) (ResolvedPaymentIdMap, error) {
	resolvedIdMap := ResolvedPaymentIdMap{}
	for _, id := range order.PaymentIds {
		resolvedIdMap[id] = NewResolvedPaymentId(id)
	}
	bsk, err := ordr.Basket(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't get basket: %w", err)
	}

	if !bsk.HasUserAddedItems() { // незачем ничего считать, если корзина пустая
		return resolvedIdMap, nil
	}

	logger := r.loggerFactory.Create(ctx)
	h := fnv.New64a()
	// ошибка глушится, т.к. не влияет на дальнейшую работу
	_, _ = h.Write([]byte("Resolve:" + ordr.CompileCacheKey(ordr.Fingerprint())))
	cacheKey := strconv.FormatUint(h.Sum64(), 10)

	encodedResolvedIds, err := r.cache.Get(ctx, cacheKey)
	if err != nil {
		var notFoundErr *simplestoragev2.NotFoundErr
		if !errors.As(err, &notFoundErr) {
			logger.Warn("can't get encoded resolved ids from cache", zap.Error(err))
		}
	} else {
		resolvedIdMap := ResolvedPaymentIdMap{}
		dec := msgpack.NewDecoder(bytes.NewBuffer(encodedResolvedIds))
		err := dec.Decode(&resolvedIdMap)
		if err != nil {
			logger.Warn("can't decode resolved ids", zap.Error(err))
		} else {
			return resolvedIdMap, nil
		}
	}

	errs := internal.NewErrorCollection()
	for _, resolver := range r.resolvers {
		err := resolver.Resolve(ctx, resolvedIdMap, ordr)
		if err != nil {
			errs.Add(err)
		}
	}

	if !errs.Empty() {
		return nil, fmt.Errorf("errors while resolve payment ids: %w", errs)
	}

	buf := bytes.NewBuffer(nil)
	encoder := msgpack.NewEncoder(buf)
	err = encoder.Encode(&resolvedIdMap)
	if err != nil {
		logger.Warn("can't encode resolved ids", zap.Error(err))
	} else {
		err := r.cache.Set(ctx, cacheKey, buf.Bytes(), 600)
		if err != nil {
			logger.Warn("can's set encoded resolved ids", zap.Error(err))
		}
	}

	return resolvedIdMap, nil
}

func (r *CompositeResolver) Add(resolver ...Resolver) {
	r.resolvers = append(r.resolvers, resolver...)
}

type ResolvedId struct {
	id        order.PaymentId                 // 1
	status    order.AllowStatus               // 2
	reasons   []*order.DisallowReasonWithInfo // 3
	isDefault bool                            // 4
	isChosen  bool                            // 5
}

func (i *ResolvedId) Id() order.PaymentId {
	return i.id
}

func NewResolvedPaymentId(id order.PaymentId) *ResolvedId {
	return &ResolvedId{
		id:     id,
		status: order.AllowStatusAllow,
	}
}

func (i *ResolvedId) Status() order.AllowStatus {
	return i.status
}

func (i *ResolvedId) Disallow(status order.AllowStatus, info *order.DisallowReasonWithInfo) {
	// меняем статусы у метода оплаты только в том случае, если он жестче, чем текущий
	if i.status < status {
		i.status = status
	}

	i.reasons = append(i.reasons, info)
}

func (i *ResolvedId) Reasons() []*order.DisallowReasonWithInfo {
	return i.reasons
}

func (i *ResolvedId) IsDefault() bool {
	return i.isDefault
}

func (i *ResolvedId) IsChosen() bool {
	return i.isChosen
}

type ResolvedPaymentIdMap map[order.PaymentId]*ResolvedId

func (r ResolvedPaymentIdMap) DisallowMany(ids []order.PaymentId, status order.AllowStatus, info *order.DisallowReasonWithInfo) {
	for _, id := range ids {
		r[id].Disallow(status, info)
	}
}
