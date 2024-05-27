package payment

//go:generate mockgen -source=payment.go -destination=./payment_mock.go -package=payment

import (
	"context"
	"fmt"
	"go.citilink.cloud/order/internal/order"
)

//go:generate mockgen -source=payment.go -destination=../mock/payment.go -package=mock

type State interface {
	SetPaymentIdB2B(id int)
	SetPaymentIdB2C(id int)
	PaymentIdB2B() int
	PaymentIdB2C() int
}

func NewFactory(resolver CompositePaymentResolver) *Factory {
	return &Factory{resolver: resolver}
}

type Factory struct {
	resolver CompositePaymentResolver
}

func (f *Factory) Create(order order.Order, st State) *Payment {
	return NewPayment(order, f.resolver, st)
}

func NewPayment(order order.Order, resolver CompositePaymentResolver, st State) *Payment {
	return &Payment{
		order:    order,
		resolver: resolver,
		state:    st,
	}
}

type Payment struct {
	order    order.Order
	resolver CompositePaymentResolver
	state    State
}

func (p *Payment) Id() order.PaymentId {
	if p.order.User().GetB2B().GetIsB2BState() {
		return order.PaymentId(p.state.PaymentIdB2B())
	}

	return order.PaymentId(p.state.PaymentIdB2C())
}

func (p *Payment) SetId(ctx context.Context, id order.PaymentId) (bool, error, order.ResolvedPaymentId) {
	isValid, err, resolvedPaymentId := p.Check(ctx, id)
	if err != nil {
		return false, fmt.Errorf("can't validate id: %w", err), nil
	}

	if !isValid {
		return false, nil, resolvedPaymentId
	}

	if p.order.User().GetB2B().GetIsB2BState() {
		p.state.SetPaymentIdB2B(int(id))
	} else {
		p.state.SetPaymentIdB2C(int(id))
	}

	return true, nil, nil
}

func (p *Payment) Check(
	ctx context.Context,
	id order.PaymentId,
) (isValid bool, err error, resolvedPaymentId order.ResolvedPaymentId) {
	resolvedIds, err := p.ResolveIds(ctx)
	if err != nil {
		return false, fmt.Errorf("can't resolve ids: %w", err), nil
	}

	var foundResolvedId order.ResolvedPaymentId
	for _, resolvedId := range resolvedIds {
		if resolvedId.Id() == id {
			foundResolvedId = resolvedId
			break
		}
	}
	if foundResolvedId == nil {
		return false, nil, nil
	}

	if foundResolvedId.Status() != order.AllowStatusAllow {
		return false, nil, foundResolvedId
	}

	return true, nil, foundResolvedId
}

func (p *Payment) CheckAllowed(ctx context.Context) (isValid bool, err error, resolvedPaymentId order.ResolvedPaymentId) {
	return p.Check(ctx, p.Id())
}

func (p *Payment) NeedFeel() bool {
	return !p.Id().IsValid()
}

func (p *Payment) ResolveIds(ctx context.Context) ([]order.ResolvedPaymentId, error) {
	resolvedIdMap, err := p.resolver.Resolve(ctx, p.order)
	if err != nil {
		return nil, fmt.Errorf("can't resolve ids: %w", err)
	}

	resolvedIds := resolvedIdMap.ToSlice()
	pId := p.order.Payment().Id()
	if pId.IsValid() && resolvedIdMap[pId].Status() == order.AllowStatusAllow {
		resolvedIdMap[pId].isChosen = true
	}

	if len(resolvedIds) > 0 {
		allowedResolvedIds := resolvedIds.Allowed()
		if len(allowedResolvedIds) > 0 {
			allowedResolvedIds[0].isDefault = true
		} else {
			resolvedIds[0].isDefault = true
		}
	}

	idsToReturn := make([]order.ResolvedPaymentId, 0, len(resolvedIds))
	for _, id := range resolvedIds {
		idsToReturn = append(idsToReturn, id)
	}

	return idsToReturn, nil
}

// ToSlice приводит мапу к срезу, одновременно сортируя относительно среза order.PaymentIds
func (r ResolvedPaymentIdMap) ToSlice() ResolvedPaymentIds {
	ids := make([]*ResolvedId, 0, len(r))
	for _, id := range order.PaymentIds {
		if v, ok := r[id]; ok {
			ids = append(ids, v)
		}
	}

	return ids
}

type ResolvedPaymentIds []*ResolvedId

// Allowed возвращает только разрешенные типы оплаты
func (r ResolvedPaymentIds) Allowed() ResolvedPaymentIds {
	var ids ResolvedPaymentIds
	for _, id := range r {
		if id.Status() == order.AllowStatusAllow {
			ids = append(ids, id)
		}
	}

	return ids
}

// Problems проверяет всё ли хорошо с выбранным типом оплаты заказа. Если нет вернуть проблему
func (p *Payment) Problems(ctx context.Context) ([]*order.Problem, error) {
	var problems []*order.Problem
	// если тип оплаты еще не указан, то и проблем никаких быть не может
	if p.NeedFeel() {
		return problems, nil
	}

	isValid, err, _ := p.CheckAllowed(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't check allowed payment: %w", err)
	}

	if isValid {
		return problems, nil
	}

	problem := order.NewProblem(
		order.ProblemIdPaymentIdNotAllowed,
		fmt.Sprintf("Тип оплаты %s недоступен", p.Id().Name()),
		order.SubsystemPaymentMethod,
	)

	return append(problems, problem), nil
}
