package basket

import (
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"testing"
)

func TestNewProblem(t *testing.T) {
	type args struct {
		item    *basket_item.Item
		problem *basket_item.Problem
	}
	expectedInfo := &Problem{
		item:    &basket_item.Item{},
		problem: &basket_item.Problem{},
	}

	tests := []struct {
		name string
		args args
		want *Problem
	}{
		{
			name: "test creation",
			args: args{
				item:    expectedInfo.item,
				problem: expectedInfo.problem,
			},
			want: expectedInfo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProblem(tt.args.item, tt.args.problem); !assert.Equal(t, tt.want, got) {
				t.Errorf("NewProblem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProblem_Item(t *testing.T) {
	type fields struct {
		item    *basket_item.Item
		problem *basket_item.Problem
	}

	expectedItem := &basket_item.Item{}

	tests := []struct {
		name   string
		fields fields
		want   *basket_item.Item
	}{
		{
			name: "test getting item",
			fields: fields{
				item: expectedItem,
			},
			want: expectedItem,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Problem{
				item:    tt.fields.item,
				problem: tt.fields.problem,
			}
			if got := p.Item(); !assert.Equal(t, tt.want, got) {
				t.Errorf("Item() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProblem_Problem(t *testing.T) {
	type fields struct {
		item    *basket_item.Item
		problem *basket_item.Problem
	}

	expectedProblem := &basket_item.Problem{}

	tests := []struct {
		name   string
		fields fields
		want   *basket_item.Problem
	}{
		{
			name: "test getting item",
			fields: fields{
				problem: expectedProblem,
			},
			want: expectedProblem,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Problem{
				item:    tt.fields.item,
				problem: tt.fields.problem,
			}
			if got := p.Problem(); !assert.Equal(t, tt.want, got) {
				t.Errorf("Problem() = %v, want %v", got, tt.want)
			}
		})
	}
}
