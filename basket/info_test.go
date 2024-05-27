package basket

import (
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"testing"
)

func TestInfo_Info(t *testing.T) {
	type fields struct {
		item *basket_item.Item
		info *basket_item.Info
	}

	expectedInfo := &basket_item.Info{}

	tests := []struct {
		name   string
		fields fields
		want   *basket_item.Info
	}{
		{
			name: "test getting info",
			fields: fields{
				info: expectedInfo,
			},
			want: expectedInfo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Info{
				item: tt.fields.item,
				info: tt.fields.info,
			}
			if got := i.Info(); !assert.Equal(t, tt.want, got) {
				t.Errorf("Info() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfo_Item(t *testing.T) {
	type fields struct {
		item *basket_item.Item
		info *basket_item.Info
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
			i := &Info{
				item: tt.fields.item,
				info: tt.fields.info,
			}
			if got := i.Item(); !assert.Equal(t, tt.want, got) {
				t.Errorf("Item() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewInfo(t *testing.T) {
	type args struct {
		item *basket_item.Item
		info *basket_item.Info
	}
	expectedInfo := &Info{
		item: &basket_item.Item{},
		info: &basket_item.Info{},
	}

	tests := []struct {
		name string
		args args
		want *Info
	}{
		{
			name: "test creation",
			args: args{
				item: expectedInfo.item,
				info: expectedInfo.info,
			},
			want: expectedInfo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewInfo(tt.args.item, tt.args.info); !assert.Equal(t, tt.want, got) {
				t.Errorf("NewInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
