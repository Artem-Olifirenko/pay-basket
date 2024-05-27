package basket_item

import (
	"fmt"
	"go.citilink.cloud/order/internal"
	"gopkg.in/vmihailenco/msgpack.v2"
)

func (i *Info) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(3); err != nil {
		return err
	}

	if err := e.EncodeInt(int(i.Id())); err != nil { // 1
		return err
	}

	if err := e.EncodeString(i.Message()); err != nil { // 2
		return err
	}

	if err := e.Encode(i.Additionals()); err != nil { // 3
		return err
	}

	return nil
}

func (i *Info) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "Info array len")
	}

	if l != 3 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("incorrect Info length: %d", l), 0, "incorrect Info length")
	}

	if v, err := d.DecodeInt(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "Info id")
	} else {
		i.id = InfoId(v)
	}
	if v, err := d.DecodeString(); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "Info message")
	} else {
		i.message = v
	}
	if err := d.Decode(&i.additions); err != nil { // 3
		return internal.NewMsgPackDecodeError(err, 3, "Info additions")
	}

	return nil
}

func (a *InfoAdditions) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(3); err != nil {
		return err
	}

	if err := e.Encode(&a.PriceChanged); err != nil { // 1
		return err
	}

	if err := e.Encode(&a.CountMoreThenAvail); err != nil { // 2
		return err
	}

	if err := e.Encode(&a.ChangedItem); err != nil { // 3
		return err
	}

	return nil
}

func (a *InfoAdditions) DecodeMsgpack(d *msgpack.Decoder) error {
	length, err := d.DecodeArrayLen()
	if err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "InfoAdditions array len")
	}

	if length < 1 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("incorrect InfoAdditions length: %d", length), 0, "incorrect InfoAdditions length")
	}

	if err := d.Decode(&a.PriceChanged); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "InfoAdditions PriceChanged")
	}

	if length > 1 {
		if err := d.Decode(&a.CountMoreThenAvail); err != nil { // 2
			return internal.NewMsgPackDecodeError(err, 2, "InfoAdditions CountMoreThenAvail")
		}
	}

	if length > 2 {
		if err := d.Decode(&a.ChangedItem); err != nil { // 3
			return internal.NewDecodeErr(err)
		}
	}

	return nil
}

func (a *PriceChangedInfoAddition) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(2); err != nil {
		return err
	}

	if err := e.EncodeInt(a.From); err != nil {
		return err
	}
	if err := e.EncodeInt(a.To); err != nil {
		return err
	}

	return nil
}

func (a *PriceChangedInfoAddition) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "PriceChangedInfoAddition array len")
	}

	if l != 2 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("incorrect PriceChangedInfoAddition length: %d", l), 0, "incorrect PriceChangedInfoAddition length")
	}

	if v, err := d.DecodeInt(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "PriceChangedInfoAddition From")
	} else {
		a.From = v
	}
	if v, err := d.DecodeInt(); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "PriceChangedInfoAddition To")
	} else {
		a.To = v
	}

	return nil
}

func (a *CountMoreThenAvailInfoAdditions) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(1); err != nil {
		return err
	}

	if err := e.EncodeInt(a.AvailCount); err != nil {
		return err
	}

	return nil
}

func (a *CountMoreThenAvailInfoAdditions) DecodeMsgpack(d *msgpack.Decoder) error {
	length, err := d.DecodeArrayLen()
	if err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "CountMoreThenAvailInfoAdditions array len")
	}

	if length != 1 {
		return internal.NewMsgPackDecodeError(
			fmt.Errorf("incorrect CountMoreThenAvailInfoAdditions length: %d", length),
			0,
			"incorrect CountMoreThenAvailInfoAdditions length")
	}

	if v, err := d.DecodeInt(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "CountMoreThenAvailInfoAdditions AvailCount")
	} else {
		a.AvailCount = v
	}

	return nil
}

func (c *ChangedItemInfoAdditions) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(5); err != nil {
		return err
	}

	if err := e.EncodeString(c.ItemId); err != nil { // 1
		return err
	}

	if err := e.EncodeString(c.UniqId); err != nil { // 2
		return err
	}

	if err := e.EncodeInt(c.Count); err != nil { // 3
		return err
	}

	if err := e.EncodeString(c.Name); err != nil { // 4
		return err
	}

	if err := e.EncodeInt(c.Price); err != nil { // 5
		return err
	}

	return nil
}

func (c *ChangedItemInfoAdditions) DecodeMsgpack(d *msgpack.Decoder) error {
	length, err := d.DecodeArrayLen()
	if err != nil {
		return internal.NewDecodeErr(err)
	}

	if length != 5 {
		return internal.NewDecodeErr(
			fmt.Errorf("(basket_item.ChangedItemInfoAdditions) len doesn't match: %d", length))
	}

	if v, err := d.DecodeString(); err != nil { // 1
		return internal.NewDecodeErr(err)
	} else {
		c.ItemId = v
	}

	if v, err := d.DecodeString(); err != nil { // 2
		return internal.NewDecodeErr(err)
	} else {
		c.UniqId = v
	}

	if v, err := d.DecodeInt(); err != nil { // 3
		return internal.NewDecodeErr(err)
	} else {
		c.Count = v
	}

	if v, err := d.DecodeString(); err != nil { // 4
		return internal.NewDecodeErr(err)
	} else {
		c.Name = v
	}

	if v, err := d.DecodeInt(); err != nil { // 5
		return internal.NewDecodeErr(err)
	} else {
		c.Price = v
	}

	return nil
}
