package basket

import (
	"go.citilink.cloud/order/internal/order/basket/basket_item"
)

func NewProblem(item *basket_item.Item, problem *basket_item.Problem) *Problem {
	return &Problem{item: item, problem: problem}
}

type Problem struct {
	item    *basket_item.Item
	problem *basket_item.Problem
}

func (p *Problem) Item() *basket_item.Item {
	return p.item
}

func (p *Problem) Problem() *basket_item.Problem {
	return p.problem
}
