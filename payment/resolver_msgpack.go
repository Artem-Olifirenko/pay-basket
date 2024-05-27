package payment

import (
	"fmt"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order"
	"gopkg.in/vmihailenco/msgpack.v2"
)

func (i *ResolvedId) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(5); err != nil {
		return err
	}
	if err := e.EncodeInt(int(i.id)); err != nil { // 1
		return err
	}
	if err := e.EncodeInt(int(i.status)); err != nil { // 2
		return err
	}
	if err := e.Encode(&i.reasons); err != nil { // 3
		return err
	}
	if err := e.EncodeBool(i.isDefault); err != nil { // 4
		return err
	}
	if err := e.EncodeBool(i.isChosen); err != nil { // 5
		return err
	}

	return nil
}

func (i *ResolvedId) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var length int
	if length, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "ResolvedId array len")
	}

	if length < 3 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("len doesn't match %d", length), 0, "len doesn't match")
	}

	if v, err := d.DecodeInt(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "ResolvedId id")
	} else {
		i.id = order.PaymentId(v)
	}

	if v, err := d.DecodeInt(); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "ResolvedId status")
	} else {
		i.status = order.AllowStatus(v)
	}

	if err := d.Decode(&i.reasons); err != nil { // 3
		return internal.NewMsgPackDecodeError(err, 3, "ResolvedId reasons")
	}

	if length > 3 {
		if v, err := d.DecodeBool(); err != nil { // 4
			return internal.NewMsgPackDecodeError(err, 4, "ResolvedId isDefault")
		} else {
			i.isDefault = v
		}
	}

	if length == 5 {
		if v, err := d.DecodeBool(); err != nil { // 5
			return internal.NewMsgPackDecodeError(err, 5, "ResolvedId isChosen")
		} else {
			i.isChosen = v
		}
	}

	return nil
}
