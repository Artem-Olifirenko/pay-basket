package basket_item

import (
	"fmt"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/store_types"
	"gopkg.in/vmihailenco/msgpack.v2"
	"time"
)

func (i *Item) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(30); err != nil {
		return err
	}
	if err := e.EncodeString(string(i.uniqId)); err != nil { // 1
		return err
	}
	if err := e.EncodeString(string(i.itemId)); err != nil { // 2
		return err
	}
	if err := e.EncodeString(string(i.itemType)); err != nil { // 3
		return err
	}
	if err := e.EncodeString(string(i.parentUniqId)); err != nil { // 4
		return err
	}
	if err := e.EncodeString(string(i.parentItemId)); err != nil { // 5
		return err
	}
	if err := e.EncodeString(i.name); err != nil { // 6
		return err
	}
	if err := e.EncodeString(i.image); err != nil { // 7
		return err
	}
	if err := e.EncodeInt(i.count); err != nil { // 8
		return err
	}
	if err := e.EncodeInt(i.price); err != nil { // 9
		return err
	}
	if err := e.EncodeInt(i.bonus); err != nil { // 10
		return err
	}
	if err := e.EncodeInt(i.countMultiplicity); err != nil { // 11
		return err
	}

	// удалено
	if err := e.EncodeArrayLen(0); err != nil { // 12
		return err
	}

	if err := e.EncodeArrayLen(len(i.problems)); err != nil { // 13
		return err
	}
	for _, v := range i.problems {
		err := e.Encode(v)
		if err != nil {
			return err
		}
	}

	if err := e.Encode(&i.infos); err != nil { // 14
		return err
	}
	if err := e.Encode(&i.rules); err != nil { // 15
		return err
	}
	if err := e.Encode(&i.additions); err != nil { // 16
		return err
	}
	if err := e.Encode(&i.permanentProblems); err != nil { // 17
		return err
	}

	if err := e.EncodeInt(0); err != nil { // 18 deleted
		return err
	}
	if err := e.EncodeInt(0); err != nil { // 19 deleted
		return err
	}
	if err := e.EncodeString(string(i.spaceId)); err != nil { // 20
		return err
	}
	if err := e.EncodeInt(int(i.priceColumn)); err != nil { // 21
		return err
	}
	if err := e.EncodeString(i.commitFingerprint); err != nil { // 22
		return err
	}
	if err := e.EncodeBool(i.isPrepaymentMandatory); err != nil { // 23
		return err
	}
	if err := e.EncodeBool(i.hasFairPrice); err != nil { // 24
		return err
	}
	if err := e.EncodeBool(i.ignoreFairPrice); err != nil { // 25
		return err
	}
	if err := e.Encode(&i.allowResale); err != nil { // 26
		return err
	}
	if err := e.Encode(&i.discount); err != nil { // 27
		return err
	}
	if err := e.EncodeBool(i.movableToConfiguration); err != nil { // 28
		return err
	}
	if err := e.EncodeBool(i.movableFromConfiguration); err != nil { // 29
		return err
	}
	if err := e.EncodeBool(i.isSelected); err != nil { // 30
		return err
	}

	return nil
}

func (i *Item) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var itemL int
	if itemL, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "Item array len")
	}

	if itemL > 30 || itemL < 17 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("(basket_item.Item) incorrect len: %d", itemL), 0, "(basket_item.Item) incorrect len")
	}

	if v, err := d.DecodeString(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "Item uniqId")
	} else {
		i.uniqId = UniqId(v)
	}
	if v, err := d.DecodeString(); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "Item itemId")
	} else {
		i.itemId = ItemId(v)
	}
	if v, err := d.DecodeString(); err != nil { // 3
		return internal.NewMsgPackDecodeError(err, 3, "Item itemType")
	} else {
		i.itemType = Type(v)
	}
	if v, err := d.DecodeString(); err != nil { // 4
		return internal.NewMsgPackDecodeError(err, 4, "Item parentUniqId")
	} else {
		i.parentUniqId = UniqId(v)
	}
	if v, err := d.DecodeString(); err != nil { // 5
		return internal.NewMsgPackDecodeError(err, 5, "Item parentItemId")
	} else {
		i.parentItemId = ItemId(v)
	}
	if v, err := d.DecodeString(); err != nil { // 6
		return internal.NewMsgPackDecodeError(err, 6, "Item name")
	} else {
		i.name = v
	}
	if v, err := d.DecodeString(); err != nil { // 7
		return internal.NewMsgPackDecodeError(err, 7, "Item image")
	} else {
		i.image = v
	}
	if v, err := d.DecodeInt(); err != nil { // 8
		return internal.NewMsgPackDecodeError(err, 8, "Item count")
	} else {
		i.count = v
	}
	if v, err := d.DecodeInt(); err != nil { // 9
		return internal.NewMsgPackDecodeError(err, 9, "Item price")
	} else {
		i.price = v
	}
	if v, err := d.DecodeInt(); err != nil { // 10
		return internal.NewMsgPackDecodeError(err, 10, "Item bonus")
	} else {
		i.bonus = v
	}
	if v, err := d.DecodeInt(); err != nil { // 11
		return internal.NewMsgPackDecodeError(err, 11, "Item countMultiplicity")
	} else {
		i.countMultiplicity = v
	}

	// удалено
	l, err := d.DecodeArrayLen() // 12
	if err != nil {
		return internal.NewMsgPackDecodeError(err, 12, "Item decode array len")
	}
	for j := 0; j < l; j++ {
		if _, err := d.DecodeInt(); err != nil {
			return internal.NewMsgPackDecodeError(err, 12, "Item DecodeInt")
		}
	}

	problemLen, err := d.DecodeArrayLen() // 13
	if err != nil {
		return internal.NewMsgPackDecodeError(err, 13, "Item DecodeArrayLen")
	}

	if problemLen > 0 {
		i.problems = make([]*Problem, problemLen)
		for j := 0; j < problemLen; j++ {
			if err := d.Decode(&i.problems[j]); err != nil {
				return internal.NewMsgPackDecodeError(err, 13, "Item problems")
			}
		}
	}

	if err := d.Decode(&i.infos); err != nil { // 14
		return internal.NewMsgPackDecodeError(err, 14, "Item infos")
	}
	if err := d.Decode(&i.rules); err != nil { // 15
		return internal.NewMsgPackDecodeError(err, 15, "Item rules")
	}
	if err := d.Decode(&i.additions); err != nil { // 16
		return internal.NewMsgPackDecodeError(err, 16, "Item additions")
	}
	if err := d.Decode(&i.permanentProblems); err != nil { // 17
		return internal.NewMsgPackDecodeError(err, 17, "Item permanentProblems")
	}
	if itemL > 17 {
		if _, err := d.DecodeInt(); err != nil { // 18 deleted
			return internal.NewMsgPackDecodeError(err, 18, "Item DecodeInt")
		}
		if _, err := d.DecodeInt(); err != nil { // 19 deleted
			return internal.NewMsgPackDecodeError(err, 19, "Item DecodeInt")
		}
		if v, err := d.DecodeString(); err != nil { // 20
			return internal.NewMsgPackDecodeError(err, 20, "Item spaceId")
		} else {
			i.spaceId = store_types.SpaceId(v)
		}
		if v, err := d.DecodeInt(); err != nil { // 21
			return internal.NewMsgPackDecodeError(err, 21, "Item priceColumn")
		} else {
			i.priceColumn = catalog_types.PriceColumn(v)
		}
	}

	if itemL > 21 {
		if v, err := d.DecodeString(); err != nil { // 22
			return internal.NewMsgPackDecodeError(err, 22, "Item commitFingerprint")
		} else {
			i.commitFingerprint = v
		}
	}

	if itemL > 22 {
		if v, err := d.DecodeBool(); err != nil {
			return internal.NewMsgPackDecodeError(err, 23, "Item isPrepaymentMandatory")
		} else {
			i.isPrepaymentMandatory = v
		}
	}

	if itemL > 23 {
		if v, err := d.DecodeBool(); err != nil {
			return internal.NewMsgPackDecodeError(err, 24, "Item hasFairPrice")
		} else {
			i.hasFairPrice = v
		}
	}

	if itemL > 24 {
		if v, err := d.DecodeBool(); err != nil {
			return internal.NewMsgPackDecodeError(err, 25, "Item ignoreFairPrice")
		} else {
			i.ignoreFairPrice = v
		}
	}

	if itemL > 25 {
		if err := d.Decode(&i.allowResale); err != nil { // 26
			return internal.NewMsgPackDecodeError(err, 26, "Item allowResale")
		}
	}

	if itemL > 26 {
		if err := d.Decode(&i.discount); err != nil { // 27
			return internal.NewMsgPackDecodeError(err, 27, "Item discount")
		}
	}

	if itemL > 27 { // 28
		if v, err := d.DecodeBool(); err != nil {
			return internal.NewMsgPackDecodeError(err, 28, "Item movableToConfiguration")
		} else {
			i.movableToConfiguration = v
		}
	}

	if itemL > 28 { // 29
		if v, err := d.DecodeBool(); err != nil {
			return internal.NewMsgPackDecodeError(err, 29, "Item movableFromConfiguration")
		} else {
			i.movableFromConfiguration = v
		}
	}

	if itemL > 29 { // 30
		if v, err := d.DecodeBool(); err != nil {
			return internal.NewMsgPackDecodeError(err, 30, "Item isSelected")
		} else {
			i.isSelected = v
		}
	} else {
		i.isSelected = true
	}

	return nil
}

func (a *AllowResale) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(2); err != nil {
		return err
	}

	if err := e.EncodeBool(a.isAllow); err != nil { // 1
		return err
	}
	if err := e.EncodeString(a.commodityGroupName); err != nil { // 2
		return err
	}

	return nil
}

func (a *AllowResale) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "AllowResale array len")
	}

	if l != 2 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("(basket_item.AllowResale) incorrect len: %d", l), 0, "Item allowResale")
	}

	if v, err := d.DecodeBool(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "Item allowResale isAllow")
	} else {
		a.isAllow = v
	}
	if v, err := d.DecodeString(); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "Item allowResale commodityGroupName")
	} else {
		a.commodityGroupName = v
	}

	return nil
}

func (i *ItemDiscount) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(4); err != nil {
		return err
	}

	if err := e.EncodeInt(i.Coupon); err != nil { // 1
		return err
	}
	if err := e.EncodeInt(i.Action); err != nil { // 2
		return err
	}
	if err := e.EncodeInt(i.Total); err != nil { // 3
		return err
	}
	if err := e.EncodeString(i.AppliedPromotions); err != nil { // 4
		return err
	}

	return nil
}

func (i *ItemDiscount) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "ItemDiscount array len")
	}

	if l != 4 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("(basket_item.ItemDiscount) incorrect len: %d", l), 0, "(basket_item.ItemDiscount) incorrect len")
	}

	if v, err := d.DecodeInt(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "ItemDiscount Coupon")
	} else {
		i.Coupon = v
	}

	if v, err := d.DecodeInt(); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "ItemDiscount Action")
	} else {
		i.Action = v
	}

	if v, err := d.DecodeInt(); err != nil { // 3
		return internal.NewMsgPackDecodeError(err, 3, "ItemDiscount Total")
	} else {
		i.Total = v
	}

	if v, err := d.DecodeString(); err != nil { // 4
		return internal.NewMsgPackDecodeError(err, 4, "ItemDiscount AppliedPromotions")
	} else {
		i.AppliedPromotions = v
	}

	return nil
}

func (i *ItemAdditions) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(4); err != nil {
		return err
	}

	if err := e.Encode(i.GetProduct()); err != nil { // 1
		return err
	}
	if err := e.Encode(i.GetConfiguration()); err != nil { // 2
		return err
	}
	if err := e.Encode(i.GetSubcontractServiceForProduct()); err != nil { // 3
		return err
	}
	if err := e.Encode(i.GetService()); err != nil { // 4
		return err
	}

	return nil
}

func (i *ItemAdditions) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "ItemAdditions array len")
	}

	if l != 4 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("(basket_item.ItemAdditions) incorrect len: %d", l), 0, "(basket_item.ItemAdditions) incorrect len")
	}

	if err := d.Decode(&i.Product); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "ItemAdditions Product")
	}
	if err := d.Decode(&i.Configuration); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "ItemAdditions Configuration")
	}
	if err := d.Decode(&i.SubcontractServiceForProduct); err != nil { // 3
		return internal.NewMsgPackDecodeError(err, 3, "ItemAdditions SubcontractServiceForProduct")
	}

	if l > 3 {
		if err := d.Decode(&i.Service); err != nil { // 4
			return internal.NewMsgPackDecodeError(err, 4, "ItemAdditions Service")
		}
	}

	return nil
}

func (p *ProductItemAdditions) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(20); err != nil {
		return err
	}

	if err := e.EncodeBool(p.IsAvailInStore()); err != nil { // 1
		return err
	}
	if err := e.EncodeInt(p.Vat()); err != nil { // 2
		return err
	}
	if err := e.EncodeInt(int(p.CategoryId())); err != nil { // 3
		return err
	}
	if err := e.EncodeString(""); err != nil { // 4 удаленное поле BrandName
		return err
	}
	if err := e.EncodeInt(0); err != nil { // 5
		return err
	}
	if err := e.EncodeBool(p.IsOEM()); err != nil { // 6
		return err
	}
	if err := e.EncodeInt(p.AvailTotal()); err != nil { // 7
		return err
	}
	if err := e.EncodeBool(p.IsCountMoreThenAvailChecked()); err != nil { // 8
		return err
	}

	creditPrograms := p.CreditPrograms()
	if err := e.EncodeArrayLen(len(creditPrograms)); err != nil { // 9
		return err
	}
	for _, v := range creditPrograms {
		err := e.EncodeString(string(v))
		if err != nil {
			return err
		}
	}

	if err := e.EncodeString(""); err != nil { // 10 удаленное поле CategoryName
		return err
	}
	if err := e.EncodeString(""); err != nil { // 11 удаленное поле CategoryPath
		return err
	}
	if err := e.EncodeBool(p.IsAvailForDPD()); err != nil { // 12
		return err
	}
	if err := e.EncodeBool(p.IsMarked()); err != nil { // 13
		return err
	}
	if err := e.EncodeInt(int(p.MarkedPurchaseReason())); err != nil { // 14
		return err
	}
	if err := e.EncodeBool(p.IsDiscounted()); err != nil { // 15
		return err
	}
	if err := e.EncodeString(p.CategoryName()); err != nil { // 16
		return err
	}
	if err := e.EncodeString(p.BrandName()); err != nil { // 17
		return err
	}
	if err := e.EncodeString(p.CategoryPath()); err != nil { // 18
		return err
	}
	if err := e.EncodeString(p.ShortName()); err != nil { // 19
		return err
	}
	if err := e.EncodeBool(p.IsFnsTracked()); err != nil { // 20
		return err
	}

	return nil
}

func (p *ProductItemAdditions) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var length int
	if length, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "ProductItemAdditions array len")
	}

	if v, err := d.DecodeBool(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "ProductItemAdditions IsAvailInStore")
	} else {
		p.SetIsAvailInStore(v)
	}
	if v, err := d.DecodeInt(); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "ProductItemAdditions Vat")
	} else {
		p.SetVat(v)
	}
	if v, err := d.DecodeInt(); err != nil { // 3
		return internal.NewMsgPackDecodeError(err, 3, "ProductItemAdditions CategoryId")
	} else {
		p.SetCategoryId(catalog_types.CategoryId(v))
	}
	if _, err := d.DecodeString(); err != nil { // 4
		return internal.NewMsgPackDecodeError(err, 4, "ProductItemAdditions DecodeString")
	}

	if _, err := d.DecodeInt(); err != nil { // 5 (удалено)
		return internal.NewMsgPackDecodeError(err, 5, "ProductItemAdditions DecodeInt")
	}

	if length > 5 {
		if v, err := d.DecodeBool(); err != nil { // 6
			return internal.NewMsgPackDecodeError(err, 6, "ProductItemAdditions IsOEM")
		} else {
			p.SetIsOEM(v)
		}
	}

	if length > 6 {
		if v, err := d.DecodeInt(); err != nil { // 7
			return internal.NewMsgPackDecodeError(err, 7, "ProductItemAdditions AvailTotal")
		} else {
			p.SetAvailTotal(v)
		}
		if v, err := d.DecodeBool(); err != nil { // 8
			return internal.NewMsgPackDecodeError(err, 8, "ProductItemAdditions IsCountMoreThenAvailChecked")
		} else {
			p.SetIsCountMoreThenAvailChecked(v)
		}
	}

	if length > 8 {
		l, err := d.DecodeArrayLen() // 9
		if err != nil {
			return internal.NewMsgPackDecodeError(err, 9, "ProductItemAdditions DecodeArrayLen")
		}
		creditPrograms := make([]catalog_types.CreditProgram, l)
		for j := 0; j < l; j++ {
			if v, err := d.DecodeString(); err != nil {
				return internal.NewMsgPackDecodeError(err, 9, "ProductItemAdditions DecodeString")
			} else {
				creditPrograms[j] = catalog_types.CreditProgram(v)
			}
		}
		p.SetCreditPrograms(creditPrograms)
	}

	if length > 9 {
		if _, err := d.DecodeString(); err != nil { // 10
			return internal.NewMsgPackDecodeError(err, 10, "ProductItemAdditions DecodeString")
		}
	}

	if length > 10 {
		if _, err := d.DecodeString(); err != nil { // 11
			return internal.NewMsgPackDecodeError(err, 11, "ProductItemAdditions DecodeString")
		}
	}

	if length > 11 {
		if v, err := d.DecodeBool(); err != nil { // 12
			return internal.NewMsgPackDecodeError(err, 12, "ProductItemAdditions DecodeBool")
		} else {
			p.SetIsAvailForDPD(v)
		}
	}

	if length > 12 {
		if v, err := d.DecodeBool(); err != nil { // 13
			return internal.NewDecodeErr(err)
		} else {
			p.SetIsMarked(v)
		}
	}

	if length > 13 {
		if v, err := d.DecodeInt(); err != nil { // 14
			return internal.NewDecodeErr(err)
		} else {
			p.SetMarkedPurchaseReason(MarkedPurchaseReason(v))
		}
	}

	if length > 14 {
		if v, err := d.DecodeBool(); err != nil { // 15
			return internal.NewDecodeErr(err)
		} else {
			p.SetIsDiscounted(v)
		}
	}

	if length > 15 {
		if v, err := d.DecodeString(); err != nil { // 15
			return internal.NewDecodeErr(err)
		} else {
			p.SetCategoryName(v)
		}
	}

	if length > 16 {
		if v, err := d.DecodeString(); err != nil { // 16
			return internal.NewDecodeErr(err)
		} else {
			p.SetBrandName(v)
		}
	}

	if length > 17 {
		if v, err := d.DecodeString(); err != nil { // 17
			return internal.NewDecodeErr(err)
		} else {
			p.SetCategoryPath(v)
		}
	}

	if length > 18 {
		if v, err := d.DecodeString(); err != nil { // 18
			return internal.NewDecodeErr(err)
		} else {
			p.SetShortName(v)
		}
	}

	if length > 19 {
		if v, err := d.DecodeBool(); err != nil { // 20
			return internal.NewDecodeErr(err)
		} else {
			p.SetIsFnsTracked(v)
		}
	}

	return nil
}

func (c *ConfiguratorItemAdditions) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(2); err != nil {
		return err
	}

	if err := e.EncodeString(c.GetConfId()); err != nil { // 1
		return err
	}
	if err := e.EncodeInt(int(c.GetConfType())); err != nil { // 2
		return err
	}

	return nil
}

func (c *ConfiguratorItemAdditions) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "ConfiguratorItemAdditions array len")
	}

	if l != 2 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("(basket_item.ConfiguratorItemAdditions) len doesn't match: %d", l), 0, "(basket_item.ConfiguratorItemAdditions) len doesn't match")
	}

	if v, err := d.DecodeString(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "ConfiguratorItemAdditions ConfId")
	} else {
		c.SetConfId(v)
	}
	if v, err := d.DecodeInt(); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "ConfiguratorItemAdditions ConfType")
	} else {
		c.SetConfType(ConfType(v))
	}

	return nil
}

func (s *SubcontractItemAdditions) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(1); err != nil {
		return err
	}

	if err := e.Encode(s.GetApplyServiceInfo()); err != nil { // 1
		return err
	}

	return nil
}

func (s *SubcontractItemAdditions) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "SubcontractItemAdditions array len")
	}

	if l != 1 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("(basket_item.SubcontractItemAdditions) len doesn't match: %d", l), 0, "(basket_item.SubcontractItemAdditions) len doesn't match")
	}

	if err := d.Decode(&s.ApplyServiceInfo); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "SubcontractItemAdditions ApplyServiceInfo")
	}

	return nil
}

func (s *SubcontractApplyServiceInfo) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(4); err != nil {
		return err
	}

	if err := e.EncodeString(s.Date.Format(time.RFC3339)); err != nil { // 1
		return err
	}
	if err := e.EncodeString(s.Address); err != nil { // 2
		return err
	}
	if err := e.EncodeString(string(s.CityKladrId)); err != nil { // 3
		return err
	}
	if err := e.EncodeString(s.CityName); err != nil { // 4
		return err
	}

	return nil
}

func (s *SubcontractApplyServiceInfo) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "SubcontractApplyServiceInfo array len")
	}

	if l != 4 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("(basket_item.SubcontractApplyServiceInfo) len doesn't match: %d", l), 0, "(basket_item.SubcontractApplyServiceInfo) len doesn't match")
	}

	if v, err := d.DecodeString(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "SubcontractApplyServiceInfo Date")
	} else {
		// Использован формат стандарта RFC3339 в соответствии с регламентом
		parsedTime, err := time.Parse(time.RFC3339, v)
		if err != nil {
			// Для записей которые были сохранены ранее, в формате стандарта ANSIC
			parsedTime, err = time.Parse(time.ANSIC, v)
			if err != nil {
				return internal.NewMsgPackDecodeError(err, 1, "SubcontractApplyServiceInfo time.Parse")
			}
		}
		s.Date = parsedTime
	}
	if v, err := d.DecodeString(); err != nil { // 2
		return internal.NewMsgPackDecodeError(err, 2, "SubcontractApplyServiceInfo Address")
	} else {
		s.Address = v
	}
	if v, err := d.DecodeString(); err != nil { // 3
		return internal.NewMsgPackDecodeError(err, 3, "SubcontractApplyServiceInfo CityKladrId")
	} else {
		s.CityKladrId = store_types.KladrId(v)
	}
	if v, err := d.DecodeString(); err != nil { // 4
		return internal.NewMsgPackDecodeError(err, 4, "SubcontractApplyServiceInfo CityName")
	} else {
		s.CityName = v
	}

	return nil
}

func (s *Service) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(2); err != nil {
		return err
	}

	if err := e.EncodeBool(s.GetIsCreditAvail()); err != nil { // 1
		return err
	}
	if err := e.EncodeBool(s.GetIsAvailableForInstallments()); err != nil { // 2
		return err
	}

	return nil
}

func (s *Service) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "Service array len")
	}

	if l < 1 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("(basket_item.Service) not enough fields: %d", l), 0, "(basket_item.Service) not enough fields")
	}

	if v, err := d.DecodeBool(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "Service IsCreditAvail")
	} else {
		s.SetIsCreditAvail(v)
	}

	if l > 1 {
		if v, err := d.DecodeBool(); err != nil { // 2
			return internal.NewMsgPackDecodeError(err, 2, "Service IsAvailableForInstallments")
		} else {
			s.SetIsAvailableForInstallments(v)
		}
	}

	return nil
}

func (r *Rules) EncodeMsgpack(e *msgpack.Encoder) error {
	if err := e.EncodeArrayLen(1); err != nil {
		return err
	}

	if err := e.EncodeInt(r.MaxCount()); err != nil {
		return err
	}

	return nil
}

func (r *Rules) DecodeMsgpack(d *msgpack.Decoder) error {
	var err error
	var l int
	if l, err = d.DecodeArrayLen(); err != nil {
		return internal.NewMsgPackDecodeError(err, 0, "Rules array len")
	}

	if l != 1 {
		return internal.NewMsgPackDecodeError(fmt.Errorf("len doesn't match: %d", l), 0, "len doesn't match")
	}

	if v, err := d.DecodeInt(); err != nil { // 1
		return internal.NewMsgPackDecodeError(err, 1, "Rules maxCount")
	} else {
		r.SetMaxCount(v)
	}

	return nil
}
