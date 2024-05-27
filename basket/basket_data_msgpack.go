package basket

import (
	"fmt"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/store_types"
	"gopkg.in/vmihailenco/msgpack.v2"
)

func (b *BasketData) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(9); err != nil {
		return err
	}
	if err := e.EncodeString(string(b.spaceId)); err != nil { // 1
		return err
	}
	if err := e.Encode(&b.items); err != nil { // 2
		return err
	}
	if err := e.EncodeInt(0); err != nil { // 3 deleted
		return err
	}
	if err := e.EncodeInt(0); err != nil { // 4 deleted
		return err
	}
	if err := e.EncodeInt(int(b.priceColumn)); err != nil { // 5
		return err
	}

	if err := e.EncodeArrayLen(len(b.infos)); err != nil { // 6
		return err
	}
	for _, v := range b.infos {
		err := e.Encode(v)
		if err != nil {
			return err
		}
	}

	if err := e.EncodeString(b.commitFingerprint); err != nil { // 7
		return err
	}

	if err := e.EncodeString(string(b.cityId)); err != nil { // 8
		return err
	}

	if err := e.EncodeBool(b.hasPossibleConfiguration); err != nil { // 9
		return err
	}

	return nil
}

func (b *BasketData) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var length int
	if length, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "BasketData array len")
	}

	if length > 9 || length < 2 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("(basket.BasketData) incorrect len: %d", length), 0, "(basket.BasketData) incorrect len")
	}

	if v, err := d.DecodeString(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "BasketData spaceId")
	} else {
		b.spaceId = store_types.SpaceId(v)
	}

	if err := d.Decode(&b.items); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "BasketData items")
	}

	if length > 2 {
		if _, err := d.DecodeInt(); err != nil { // 3 deleted
			return internal.NewMsgPackDecodeError(err, 3, "BasketData DecodeInt")
		}
		if _, err := d.DecodeInt(); err != nil { // 4 deleted
			return internal.NewMsgPackDecodeError(err, 4, "BasketData DecodeInt")
		}
		if v, err := d.DecodeInt(); err != nil { // 5
			return internal.NewMsgPackDecodeError(err, 5, "BasketData priceColumn")
		} else {
			b.priceColumn = catalog_types.PriceColumn(v)
		}
	}

	if length > 5 {
		l, err := d.DecodeArrayLen() // 6
		if err != nil {
			return internal.NewMsgPackDecodeError(err, 6, "BasketData array len")
		}
		b.infos = make([]*Info, l)
		for j := 0; j < l; j++ {
			if err := d.Decode(&b.infos[j]); err != nil {
				return internal.NewMsgPackDecodeError(err, 6, "BasketData decode infos")
			}
		}
	}

	if length > 6 { // 7
		if v, err := d.DecodeString(); err != nil { // 7
			return internal.NewMsgPackDecodeError(err, 7, "BasketData commitFingerprint")
		} else {
			b.commitFingerprint = v
		}
	}

	if length > 7 { // 8
		if v, err := d.DecodeString(); err != nil { // 8
			return internal.NewDecodeErr(err)
		} else {
			b.cityId = CityId(v)
		}
	}

	if length > 8 { // 9
		if v, err := d.DecodeBool(); err != nil {
			return internal.NewMsgPackDecodeError(err, 9, "Basket hasPossibleConfiguration")
		} else {
			b.hasPossibleConfiguration = v
		}
	}

	return nil
}
