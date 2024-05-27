package basket

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/suite"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"gopkg.in/vmihailenco/msgpack.v2"
	"testing"
)

type BasketInfoMsgpackSuite struct {
	suite.Suite

	price *Info
}

func (s *BasketInfoMsgpackSuite) SetupSubTest() {
	s.price = &Info{}
}

func TestBasketInfoMsgpackSuite(t *testing.T) {
	suite.Run(t, &BasketInfoMsgpackSuite{})
}

func (s *BasketInfoMsgpackSuite) TestInfoMsgpack_DecodeMsgpack() {
	tests := []struct {
		name string
		data interface{}
		want func() (*Info, error)
	}{
		{
			"wrong data",
			"invalid data",
			func() (*Info, error) {
				return nil, errors.New("can't decode msgpack field `Info array len`[0]: msgpack: invalid code ac decoding array length")
			},
		},
		{
			"incorrect len",
			[]interface{}{""},
			func() (*Info, error) {
				return nil, errors.New("can't decode msgpack field `incorrect len`[0]: incorrect len: 1")
			},
		},
		{
			"incorrect item",
			[]interface{}{"", ""},
			func() (*Info, error) {
				return nil, errors.New("can't decode msgpack field `Info item`[1]: can't decode msgpack field `Item array len`[0]: msgpack: invalid code a0 decoding array length")
			},
		},
		{
			"incorrect info",
			[]interface{}{basket_item.NewItem(
				"1", "t", "n", "i", 0, 0, 0, "msk_cl", 1,
			), ""},
			func() (*Info, error) {
				return nil, errors.New("can't decode msgpack field `Info info`[2]: can't decode msgpack field `Info array len`[0]: msgpack: invalid code a0 decoding array length")
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
			err = s.price.DecodeMsgpack(decoder)
			want, wantErr := tt.want()
			if wantErr != nil {
				s.EqualError(err, wantErr.Error())
			} else {
				s.Nil(err)
				s.Equal(want, s.price)
			}
		})
	}
}
