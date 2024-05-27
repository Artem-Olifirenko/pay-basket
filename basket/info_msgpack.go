package basket

import (
	"fmt"
	"go.citilink.cloud/order/internal"
	"gopkg.in/vmihailenco/msgpack.v2"
)

func (i *Info) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(2); err != nil {
		return err
	}
	if err := e.Encode(&i.item); err != nil { // 1
		return err
	}
	if err := e.Encode(&i.info); err != nil { // 2
		return err
	}

	return nil
}

func (i *Info) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var length int
	if length, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "Info array len")
	}

	if length != 2 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("incorrect len: %d", length), 0, "incorrect len")
	}

	if err := d.Decode(&i.item); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "Info item")
	}

	if err := d.Decode(&i.info); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "Info info")
	}

	return nil
}
