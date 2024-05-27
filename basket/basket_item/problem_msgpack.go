package basket_item

import (
	"fmt"
	"go.citilink.cloud/order/internal"
	"gopkg.in/vmihailenco/msgpack.v2"
)

func (p *Problem) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(4); err != nil {
		return err
	}

	if err := e.EncodeInt(int(p.id)); err != nil { // 1
		return err
	}
	if err := e.EncodeString(p.message); err != nil { // 2
		return err
	}
	if err := e.Encode(&p.additions); err != nil { // 3
		return err
	}
	if err := e.EncodeBool(p.isHidden); err != nil { // 4
		return err
	}

	return nil
}

func (p *Problem) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "Problem array len")
	}

	if l < 3 || l > 4 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("(basket_item.Problem) len doesn't match: %d", l), 0, "(basket_item.Problem) len doesn't match")
	}

	if v, err := d.DecodeInt(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "Problem id")
	} else {
		p.id = ProblemId(v)
	}
	if v, err := d.DecodeString(); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "Problem message")
	} else {
		p.message = v
	}
	if err := d.Decode(&p.additions); err != nil { // 3
		return internal.NewMsgPackDecodeError(err, 3, "Problem additions")
	}

	if l > 3 {
		if v, err := d.DecodeBool(); err != nil { // 4
			return internal.NewMsgPackDecodeError(err, 4, "Problem isHidden")
		} else {
			p.isHidden = v
		}
	}

	return nil
}

func (p *ProblemAdditions) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(1); err != nil {
		return err
	}

	if err := e.Encode(&p.ConfigurationProblemAdditions); err != nil { // 1
		return err
	}

	return nil
}

func (p *ProblemAdditions) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "ProblemAdditions array len")
	}

	if l != 1 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("(basket_item.ProblemAdditions) len doesn't match: %d", l), 0, "(basket_item.ProblemAdditions) len doesn't match")
	}

	if err := d.Decode(&p.ConfigurationProblemAdditions); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "ProblemAdditions ConfigurationProblemAdditions")
	}

	return nil
}

func (c *ConfigurationProblemAdditions) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(1); err != nil {
		return err
	}

	if err := e.EncodeArrayLen(len(c.NotAvailableProductItemIds)); err != nil { // 1
		return err
	}
	for _, v := range c.NotAvailableProductItemIds {
		err := e.EncodeString(string(v))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ConfigurationProblemAdditions) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var itemL int
	if itemL, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "ConfigurationProblemAdditions array len")
	}

	if itemL != 1 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("(basket_item.ConfigurationProblemAdditions) incorrect len: %d", itemL), 0, "(basket_item.ConfigurationProblemAdditions) incorrect len")
	}

	l, err := d.DecodeArrayLen() // 1
	if err != nil {
		return internal.NewMsgPackDecodeError(err, 1, "ConfigurationProblemAdditions NotAvailableProductItemIds")
	}
	c.NotAvailableProductItemIds = make([]ItemId, l)
	for j := 0; j < l; j++ {
		if err := d.Decode(&c.NotAvailableProductItemIds[j]); err != nil {
			return internal.NewMsgPackDecodeError(err, 2, "ConfigurationProblemAdditions NotAvailableProductItemIds")
		}
	}

	return nil
}
