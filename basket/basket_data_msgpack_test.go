package basket

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/suite"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"gopkg.in/vmihailenco/msgpack.v2"
	"reflect"
	"testing"
)

type BasketDataMgspackSuite struct {
	suite.Suite

	basketData BasketData
}

func (s *BasketDataMgspackSuite) SetupSubTest() {
}

func (s *BasketDataMgspackSuite) TestBasketData_DecodeMsgpack() {
	testItem := basket_item.NewItem(
		"item-id",
		basket_item.TypeProduct,
		"test",
		"",
		0,
		0,
		0,
		"",
		0,
	)

	testBasketItemInfo := basket_item.NewInfo(basket_item.InfoIdPositionRemoved, "")

	testInfo := NewInfo(testItem, testBasketItemInfo)

	tests := []struct {
		name    string
		data    interface{}
		want    func() *BasketData
		wantErr string
	}{
		{
			name:    "wrong data",
			data:    "123321",
			wantErr: "can't decode msgpack field `BasketData array len`[0]: msgpack: invalid code a6 decoding array length",
		},
		{
			name:    "incorrect len(short)",
			data:    []interface{}{0},
			wantErr: "can't decode msgpack field `(basket.BasketData) incorrect len`[0]: (basket.BasketData) incorrect len: 1",
		},
		{
			name:    "incorrect len(long)",
			data:    []interface{}{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantErr: "can't decode msgpack field `(basket.BasketData) incorrect len`[0]: (basket.BasketData) incorrect len: 10",
		},
		{
			name:    "incorrect spaceId",
			data:    []interface{}{0, 0, 0, 0},
			wantErr: "can't decode msgpack field `BasketData spaceId`[1]: msgpack: invalid code 0 decoding bytes length",
		},
		{
			name:    "incorrect items",
			data:    []interface{}{"t", 0, 0, 0},
			wantErr: "can't decode msgpack field `BasketData items`[2]: msgpack: invalid code 0 decoding map length",
		},
		{
			name: "incorrect 3 item (deleted)",
			data: []interface{}{"t", map[basket_item.UniqId]*basket_item.Item{
				"test": testItem,
			}, "t"},
			wantErr: "can't decode msgpack field `BasketData DecodeInt`[3]: msgpack: invalid code a1 decoding int64",
		},
		{
			name: "incorrect 4 item (deleted)",
			data: []interface{}{"t", map[basket_item.UniqId]*basket_item.Item{
				"test": testItem,
			}, 0, "e"},
			wantErr: "can't decode msgpack field `BasketData DecodeInt`[4]: msgpack: invalid code a1 decoding int64",
		},
		{
			name: "incorrect price",
			data: []interface{}{"t", map[basket_item.UniqId]*basket_item.Item{
				"test": testItem,
			}, 0, 1, ""},
			wantErr: "can't decode msgpack field `BasketData priceColumn`[5]: msgpack: invalid code a0 decoding int64",
		},
		{
			name: "incorrect info array len",
			data: []interface{}{"t", map[basket_item.UniqId]*basket_item.Item{
				"test": testItem,
			}, 0, 1, 10, ""},
			wantErr: "can't decode msgpack field `BasketData array len`[6]: msgpack: invalid code a0 decoding array length",
		},
		{
			name: "incorrect fingerprint",
			data: []interface{}{"t", map[basket_item.UniqId]*basket_item.Item{
				"test": testItem,
			}, 0, 1, 10, []*Info{testInfo}, 12345},
			wantErr: "can't decode msgpack field `BasketData commitFingerprint`[7]: msgpack: invalid code cd decoding bytes length",
		},
		{
			name: "incorrect infos",
			data: []interface{}{"t", map[basket_item.UniqId]*basket_item.Item{
				"test": testItem,
			}, 0, 1, 10, []string{"test"}, "fingerprint"},
			wantErr: "can't decode msgpack field `BasketData decode infos`[6]: can't decode msgpack field `Info array len`[0]: msgpack: invalid code a4 decoding array length",
		},
		{
			name: "incorrect cityId",
			data: []interface{}{"t", map[basket_item.UniqId]*basket_item.Item{
				"test": testItem,
			}, 0, 1, 10, []*Info{testInfo}, "fingerprint", false, true},
			wantErr: "decode error: msgpack: invalid code c2 decoding bytes length",
		},
		{
			name: "incorrect hasPossibleConfiguration",
			data: []interface{}{"t", map[basket_item.UniqId]*basket_item.Item{
				"test": testItem,
			}, 0, 1, 10, []*Info{testInfo}, "fingerprint", "city id", "not bool"},
			wantErr: "can't decode msgpack field `Basket hasPossibleConfiguration`[9]: msgpack: invalid code a8 decoding bool",
		},
		{
			name: "ok",
			data: []interface{}{"t", map[basket_item.UniqId]*basket_item.Item{
				"test": testItem,
			}, 0, 1, 10, []*Info{testInfo}, "fingerprint", "city Id", true},
			want: func() *BasketData {
				return &BasketData{
					spaceId:                  "t",
					items:                    map[basket_item.UniqId]*basket_item.Item{"test": testItem},
					priceColumn:              10,
					infos:                    []*Info{testInfo},
					commitFingerprint:        "fingerprint",
					cityId:                   "city Id",
					hasPossibleConfiguration: true,
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.data)
			s.Nil(err)
			reader := bytes.NewReader(buf)
			decoder := msgpack.NewDecoder(reader)
			err = s.basketData.DecodeMsgpack(decoder)
			if tt.wantErr != "" {
				s.EqualError(err, tt.wantErr)
			} else {
				s.Nil(err)
				if !reflect.DeepEqual(s.basketData, tt.want) {
					s.Error(errors.New("DecodeMsgpack() = %v, want %v"), s.basketData, tt.want())
				}
			}
		})
	}
}

func TestBasketDataMgspackSuite(t *testing.T) {
	suite.Run(t, new(BasketDataMgspackSuite))
}
