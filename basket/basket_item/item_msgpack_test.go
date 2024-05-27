package basket_item

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/suite"
	"go.citilink.cloud/catalog_types"
	"gopkg.in/vmihailenco/msgpack.v2"
	"sync"
	"testing"
	"time"
)

func TestItemMsgpackSuite(t *testing.T) {
	suite.Run(t, &ItemMsgpackSuite{})
}

type ItemMsgpackSuite struct {
	suite.Suite

	reader  *bytes.Reader
	decoder *msgpack.Decoder
}

func (s *ItemMsgpackSuite) SetupTest() {}

func (s *ItemMsgpackSuite) SetupSubTest() {}

func (s *ItemMsgpackSuite) TestRules_DecodeMsgpack() {
	tests := []struct {
		name string
		obj  interface{}
		err  func() error
		want *Rules
	}{
		{
			name: "base type substitution negative",
			obj:  []byte{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Rules array len`[0]: msgpack: invalid code c4 decoding array length")
			},
			want: &Rules{},
		},
		{
			name: "empty",
			obj:  []interface{}{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `len doesn't match`[0]: len doesn't match: 0")
			},
			want: &Rules{},
		},
		{
			name: "maxCount is int negative",
			obj:  []interface{}{"maxCount"},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Rules maxCount`[1]: msgpack: invalid code a8 decoding int64")
			},
			want: &Rules{},
		},
		{
			name: "positive",
			obj:  []interface{}{1},
			err: func() error {
				return nil
			},
			want: &Rules{1, sync.RWMutex{}},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)

			s.reader = bytes.NewReader(buf)
			s.decoder = msgpack.NewDecoder(s.reader)

			rules := &Rules{}
			err = rules.DecodeMsgpack(s.decoder)

			if err != nil {
				s.EqualError(err, tt.err().Error())
			}
			s.Equal(tt.want, rules)
		})
	}
}

func (s *ItemMsgpackSuite) TestService_DecodeMsgpack() {
	tests := []struct {
		name string
		obj  interface{}
		err  func() error
		want *Service
	}{
		{
			name: "base type substitution negative",
			obj:  []byte{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Service array len`[0]: msgpack: invalid code c4 decoding array length")
			},
			want: &Service{},
		},
		{
			name: "empty",
			obj:  []interface{}{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `(basket_item.Service) not enough fields`[0]: (basket_item.Service) not enough fields: 0")
			},
			want: &Service{},
		},
		{
			name: "IsCreditAvail is int negative",
			obj:  []interface{}{1, true},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Service IsCreditAvail`[1]: msgpack: invalid code 1 decoding bool")
			},
			want: &Service{
				IsCreditAvail:              false,
				IsAvailableForInstallments: false,
				mx:                         sync.RWMutex{},
			},
		},
		{
			name: "IsAvailableForInstallments is int negative",
			obj:  []interface{}{true, 1},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Service IsAvailableForInstallments`[2]: msgpack: invalid code 1 decoding bool")
			},
			want: &Service{
				IsCreditAvail:              true,
				IsAvailableForInstallments: false,
				mx:                         sync.RWMutex{},
			},
		},
		{
			name: "positive",
			obj:  []interface{}{true, false},
			err: func() error {
				return nil
			},
			want: &Service{true, false, sync.RWMutex{}},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)

			s.reader = bytes.NewReader(buf)
			s.decoder = msgpack.NewDecoder(s.reader)

			service := &Service{}
			err = service.DecodeMsgpack(s.decoder)

			if err != nil {
				s.EqualError(err, tt.err().Error())
			}
			s.Equal(tt.want, service)
		})
	}
}

func (s *ItemMsgpackSuite) TestSubcontractApplyServiceInfo_DecodeMsgpack() {
	date := time.Date(2021, 11, 2, 11, 2, 3, 4, &time.Location{})
	ansicStr := date.Format(time.RFC3339)
	ansicStrToDate, _ := time.Parse(time.RFC3339, ansicStr)

	tests := []struct {
		name string
		obj  interface{}
		err  func() error
		want *SubcontractApplyServiceInfo
	}{
		{
			name: "base type substitution negative",
			obj:  []byte{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `SubcontractApplyServiceInfo array len`[0]: msgpack: invalid code c4 decoding array length")
			},
			want: &SubcontractApplyServiceInfo{},
		},
		{
			name: "Date is struct Time negative",
			obj:  []interface{}{date, "street", "777-222", "Birobidzhan"},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `SubcontractApplyServiceInfo Date`[1]: msgpack: invalid code 92 decoding bytes length")
			},
			want: &SubcontractApplyServiceInfo{
				Date:        date,
				Address:     "street",
				CityKladrId: "777-222",
				CityName:    "Birobidzhan",
				mx:          sync.RWMutex{},
			},
		},
		{
			name: "Address is int negative",
			obj:  []interface{}{ansicStr, 1, "777-222", "Birobidzhan"},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `SubcontractApplyServiceInfo Address`[2]: msgpack: invalid code 1 decoding bytes length")
			},
			want: &SubcontractApplyServiceInfo{},
		},
		{
			name: "CityKladrId is int negative",
			obj:  []interface{}{ansicStr, "street", 777, "Birobidzhan"},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `SubcontractApplyServiceInfo CityKladrId`[3]: msgpack: invalid code cd decoding bytes length")
			},
			want: &SubcontractApplyServiceInfo{},
		},
		{
			name: "CityName is int string negative",
			obj:  []interface{}{ansicStr, "street", "777", 1},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `SubcontractApplyServiceInfo CityName`[4]: msgpack: invalid code 1 decoding bytes length")
			},
			want: &SubcontractApplyServiceInfo{},
		},
		{
			name: "empty",
			obj:  []interface{}{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `(basket_item.SubcontractApplyServiceInfo) len doesn't match`[0]: (basket_item.SubcontractApplyServiceInfo) len doesn't match: 0")
			},
			want: &SubcontractApplyServiceInfo{},
		},
		{
			name: "positive",
			obj:  []interface{}{ansicStr, "street", "777-222", "Birobidzhan"},
			err: func() error {
				return nil
			},
			want: &SubcontractApplyServiceInfo{Date: ansicStrToDate, Address: "street", CityKladrId: "777-222", CityName: "Birobidzhan"},
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)

			s.reader = bytes.NewReader(buf)
			s.decoder = msgpack.NewDecoder(s.reader)

			subcontract := &SubcontractApplyServiceInfo{}
			err = subcontract.DecodeMsgpack(s.decoder)

			if err != nil {
				s.EqualError(err, tt.err().Error())
			} else {
				s.Equal(tt.want, subcontract)
			}

		})
	}
}

func (s *ItemMsgpackSuite) TestSubcontractItemAdditions_DecodeMsgpack() {
	date := time.Date(2021, 11, 2, 11, 2, 3, 4, &time.Location{})
	ansicStr := date.Format(time.RFC3339)
	ansicStrToDate, _ := time.Parse(time.RFC3339, ansicStr)

	tests := []struct {
		name string
		obj  interface{}
		err  func() error
		want *SubcontractItemAdditions
	}{
		{
			name: "base type substitution negative",
			obj:  []byte{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `SubcontractItemAdditions array len`[0]: msgpack: invalid code c4 decoding array length")
			},
			want: &SubcontractItemAdditions{},
		},
		{
			name: "empty",
			obj:  []interface{}{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `(basket_item.SubcontractItemAdditions) len doesn't match`[0]: (basket_item.SubcontractItemAdditions) len doesn't match: 0")
			},
			want: &SubcontractItemAdditions{},
		},
		{
			name: "ApplyServiceInfo is empty negative",
			obj:  []interface{}{[]interface{}{}},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `SubcontractItemAdditions ApplyServiceInfo`[1]: can't decode msgpack field `(basket_item.SubcontractApplyServiceInfo) len doesn't match`[0]: (basket_item.SubcontractApplyServiceInfo) len doesn't match: 0")
			},
			want: &SubcontractItemAdditions{},
		},
		{
			name: "positive",
			obj:  []interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
			err: func() error {
				return nil
			},
			want: &SubcontractItemAdditions{
				ApplyServiceInfo: &SubcontractApplyServiceInfo{
					Date:        ansicStrToDate,
					Address:     "street",
					CityKladrId: "197",
					CityName:    "Birobidzhan",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)

			s.reader = bytes.NewReader(buf)
			s.decoder = msgpack.NewDecoder(s.reader)

			subcontract := &SubcontractItemAdditions{}
			err = subcontract.DecodeMsgpack(s.decoder)

			if err != nil {
				s.EqualError(err, tt.err().Error())
			} else {
				s.Equal(tt.want, subcontract)
			}

		})
	}
}

func (s *ItemMsgpackSuite) TestConfiguratorItemAdditions_DecodeMsgpack() {
	tests := []struct {
		name string
		obj  interface{}
		err  func() error
		want *ConfiguratorItemAdditions
	}{
		{
			name: "base type substitution negative",
			obj:  []byte{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ConfiguratorItemAdditions array len`[0]: msgpack: " +
					"invalid code c4 decoding array length")
			},
			want: &ConfiguratorItemAdditions{},
		},
		{
			name: "ConfId is int negative",
			obj:  []interface{}{1, ConfTypeUser},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ConfiguratorItemAdditions ConfId`[1]: msgpack: " +
					"invalid code 1 decoding bytes length")
			},
			want: &ConfiguratorItemAdditions{},
		},
		{
			name: "ConfType is string negative",
			obj:  []interface{}{"ConfId", "ConfType"},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ConfiguratorItemAdditions ConfType`[2]: msgpack: " +
					"invalid code a8 decoding int64")
			},
			want: &ConfiguratorItemAdditions{},
		},
		{
			name: "empty",
			obj:  []interface{}{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `(basket_item.ConfiguratorItemAdditions) len " +
					"doesn't match`[0]: (basket_item.ConfiguratorItemAdditions) len doesn't match: 0")
			},
			want: &ConfiguratorItemAdditions{},
		},
		{
			name: "positive",
			obj:  []interface{}{"config", ConfTypeUser},
			err: func() error {
				return nil
			},
			want: &ConfiguratorItemAdditions{ConfId: "config", ConfType: ConfTypeUser},
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)

			s.reader = bytes.NewReader(buf)
			s.decoder = msgpack.NewDecoder(s.reader)

			configurator := &ConfiguratorItemAdditions{}
			err = configurator.DecodeMsgpack(s.decoder)

			if err != nil {
				s.EqualError(err, tt.err().Error())
			} else {
				s.Equal(tt.want, configurator)
			}

		})
	}
}

func (s *ItemMsgpackSuite) TestProductItemAdditions_DecodeMsgpack() {
	tests := []struct {
		name string
		obj  interface{}
		err  func() error
		want *ProductItemAdditions
	}{
		{
			name: "base type substitution negative",
			obj:  []byte{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ProductItemAdditions array len`[0]: msgpack: invalid code c4 decoding array length")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "isAvailInStore is not bool negative",
			obj: &[]interface{}{
				1, 1, 1, "", 0, false,
				1, true, []catalog_types.CreditProgram{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				true, "name", "brand", "path", "short",
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ProductItemAdditions IsAvailInStore`[1]: msgpack: invalid code 1 decoding bool")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "vat is not int negative",
			obj: &[]interface{}{
				true, "int", 1, "", 0, false,
				1, true, []catalog_types.CreditProgram{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				true, "name", "brand", "path", "short",
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ProductItemAdditions Vat`[2]: msgpack: invalid code a3 decoding int64")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "categoryId is not int negative",
			obj: &[]interface{}{
				true, 1, "int", "", 0, false,
				1, true, []catalog_types.CreditProgram{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				true, "name", "brand", "path", "short",
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ProductItemAdditions CategoryId`[3]: msgpack: invalid code a3 decoding int64")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "isOEM is not bool negative",
			obj: &[]interface{}{
				true, 1, 1, "", 0, "bool",
				1, true, []catalog_types.CreditProgram{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				true, "name", "brand", "path", "short",
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ProductItemAdditions IsOEM`[6]: msgpack: invalid code a4 decoding bool")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "availTotal is not int negative",
			obj: &[]interface{}{
				true, 1, 1, "", 0, false,
				"int", true, []catalog_types.CreditProgram{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				true, "name", "brand", "path", "short",
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ProductItemAdditions AvailTotal`[7]: msgpack: invalid code a3 decoding int64")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "isCountMoreThenAvailChecked is not int negative",
			obj: &[]interface{}{
				true, 1, 1, "", 0, false,
				1, "true", []catalog_types.CreditProgram{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				true, "name", "brand", "path", "short",
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ProductItemAdditions IsCountMoreThenAvailChecked`[8]: msgpack: invalid code a4 decoding bool")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "creditPrograms is not CreditProgram negative",
			obj: &[]interface{}{
				true, 1, 1, "", 0, false,
				1, true, []byte{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				true, "name", "brand", "path", "short",
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ProductItemAdditions DecodeArrayLen`[9]: msgpack: invalid code c4 decoding array length")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "isAvailForDPD is not bool negative",
			obj: &[]interface{}{
				true, 1, 1, "", 0, false,
				1, true, []catalog_types.CreditProgram{}, "", "",
				"true", true, MarkedPurchaseReasonUnknown,
				true, "name", "brand", "path", "short",
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ProductItemAdditions DecodeBool`[12]: msgpack: invalid code a4 decoding bool")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "isMarked is not bool negative",
			obj: &[]interface{}{
				true, 1, 1, "", 0, false,
				1, true, []catalog_types.CreditProgram{}, "", "",
				true, "true", MarkedPurchaseReasonUnknown,
				true, "name", "brand", "path", "short",
			},
			err: func() error {
				return fmt.Errorf("decode error: msgpack: invalid code a4 decoding bool")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "markedPurchaseReason is not int negative",
			obj: &[]interface{}{
				true, 1, 1, "", 0, false,
				1, true, []catalog_types.CreditProgram{}, "", "",
				true, true, "MarkedPurchaseReasonUnknown",
				true, "name", "brand", "path", "short",
			},
			err: func() error {
				return fmt.Errorf("decode error: msgpack: invalid code bb decoding int64")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "isDiscounted is not bool negative",
			obj: &[]interface{}{
				true, 1, 1, "", 0, false,
				1, true, []catalog_types.CreditProgram{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				"true", "name", "brand", "path", "short",
			},
			err: func() error {
				return fmt.Errorf("decode error: msgpack: invalid code a4 decoding bool")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "categoryName is not string negative",
			obj: &[]interface{}{
				true, 1, 1, "", 0, false,
				1, true, []catalog_types.CreditProgram{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				true, 1, "brand", "path", "short",
			},
			err: func() error {
				return fmt.Errorf("decode error: msgpack: invalid code 1 decoding bytes length")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "brandName is not string negative",
			obj: &[]interface{}{
				true, 1, 1, "", 0, false,
				1, true, []catalog_types.CreditProgram{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				true, "name", 1, "path", "short",
			},
			err: func() error {
				return fmt.Errorf("decode error: msgpack: invalid code 1 decoding bytes length")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "categoryPath is not string negative",
			obj: &[]interface{}{
				true, 1, 1, "", 0, false,
				1, true, []catalog_types.CreditProgram{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				true, "name", "brand", 1, "short",
			},
			err: func() error {
				return fmt.Errorf("decode error: msgpack: invalid code 1 decoding bytes length")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "shortName is not string negative",
			obj: &[]interface{}{
				true, 1, 1, "", 0, false,
				1, true, []catalog_types.CreditProgram{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				true, "name", "brand", "path", 1,
			},
			err: func() error {
				return fmt.Errorf("decode error: msgpack: invalid code 1 decoding bytes length")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "empty",
			obj:  []interface{}{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ProductItemAdditions IsAvailInStore`[1]: EOF")
			},
			want: &ProductItemAdditions{},
		},
		{
			name: "positive",
			obj: &[]interface{}{
				true, 1, 1, "", 0, false,
				1, true, []catalog_types.CreditProgram{}, "", "",
				true, true, MarkedPurchaseReasonUnknown,
				true, "name", "brand", "path", "short",
			},
			err: func() error {
				return nil
			},
			want: &ProductItemAdditions{
				isAvailInStore: true, vat: 1, categoryId: 1, isOEM: false,
				availTotal: 1, isCountMoreThenAvailChecked: true, creditPrograms: []catalog_types.CreditProgram{},
				isAvailForDPD: true, isMarked: true, markedPurchaseReason: MarkedPurchaseReasonUnknown,
				isDiscounted: true, categoryName: "name", brandName: "brand", categoryPath: "path", shortName: "short",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)

			s.reader = bytes.NewReader(buf)
			s.decoder = msgpack.NewDecoder(s.reader)

			productItem := &ProductItemAdditions{}
			err = productItem.DecodeMsgpack(s.decoder)

			if err != nil {
				s.EqualError(err, tt.err().Error())
			} else {
				s.Equal(tt.want, productItem)
			}
		})
	}
}

func (s *ItemMsgpackSuite) TestItemAdditions_DecodeMsgpack() {
	date := time.Date(2021, 11, 2, 11, 2, 3, 4, &time.Location{})
	ansicStr := date.Format(time.RFC3339)
	ansicStrToDate, _ := time.Parse(time.RFC3339, ansicStr)

	tests := []struct {
		name string
		obj  interface{}
		err  func() error
		want *ItemAdditions
	}{
		{
			name: "base type substitution negative",
			obj:  []byte{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ItemAdditions array len`[0]: msgpack: invalid code " +
					"c4 decoding array length")
			},
			want: &ItemAdditions{},
		},
		{
			name: "Product_is_not_valalid_negative",
			obj: []interface{}{
				"",
				[]interface{}{"config", ConfTypeUser},
				[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
				[]interface{}{true, false},
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ItemAdditions Product`[1]: can't decode msgpack " +
					"field `ProductItemAdditions array len`[0]: msgpack: invalid code a0 decoding array length")
			},
			want: &ItemAdditions{},
		},
		{
			name: "Configuration is not valalid negative",
			obj: []interface{}{
				&[]interface{}{
					true, 1, 1, "", 0, false,
					1, true, []catalog_types.CreditProgram{}, "", "",
					true, true, MarkedPurchaseReasonUnknown,
					true, "name", "brand", "path", "short",
				},
				"",
				[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
				[]interface{}{true, false},
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ItemAdditions Configuration`[2]: can't decode " +
					"msgpack field `ConfiguratorItemAdditions array len`[0]: msgpack: invalid code a0 decoding array length")
			},
			want: &ItemAdditions{},
		},
		{
			name: "SubcontractServiceForProduct is not valalid negative",
			obj: []interface{}{
				&[]interface{}{
					true, 1, 1, "", 0, false,
					1, true, []catalog_types.CreditProgram{}, "", "",
					true, true, MarkedPurchaseReasonUnknown,
					true, "name", "brand", "path", "short",
				},
				[]interface{}{"config", ConfTypeUser},
				"",
				[]interface{}{true, false},
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ItemAdditions SubcontractServiceForProduct`[3]: " +
					"can't decode msgpack field `SubcontractItemAdditions array len`[0]: msgpack: invalid code a0 decoding array length")
			},
			want: &ItemAdditions{},
		},
		{
			name: "Service is not valid negative",
			obj: []interface{}{
				&[]interface{}{
					true, 1, 1, "", 0, false,
					1, true, []catalog_types.CreditProgram{}, "", "",
					true, true, MarkedPurchaseReasonUnknown,
					true, "name", "brand", "path", "short",
				},
				[]interface{}{"config", ConfTypeUser},
				[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
				"",
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `ItemAdditions Service`[4]: can't decode msgpack " +
					"field `Service array len`[0]: msgpack: invalid code a0 decoding array length")
			},
			want: &ItemAdditions{},
		},
		{
			name: "empty",
			obj:  []interface{}{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `(basket_item.ItemAdditions) incorrect len`[0]: " +
					"(basket_item.ItemAdditions) incorrect len: 0")
			},
			want: &ItemAdditions{},
		},
		{
			name: "positive",
			obj: []interface{}{
				&[]interface{}{
					true, 1, 1, "", 0, false,
					1, true, []catalog_types.CreditProgram{}, "", "",
					true, true, MarkedPurchaseReasonUnknown,
					true, "name", "brand", "path", "short",
				},
				[]interface{}{"config", ConfTypeUser},
				[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
				[]interface{}{true, false},
			},
			err: func() error {
				return nil
			},
			want: &ItemAdditions{
				Product: &ProductItemAdditions{
					isAvailInStore: true, vat: 1, categoryId: 1, isOEM: false,
					availTotal: 1, isCountMoreThenAvailChecked: true, creditPrograms: []catalog_types.CreditProgram{},
					isAvailForDPD: true, isMarked: true, markedPurchaseReason: MarkedPurchaseReasonUnknown,
					isDiscounted: true, categoryName: "name", brandName: "brand", categoryPath: "path", shortName: "short",
				},
				Configuration: &ConfiguratorItemAdditions{ConfId: "config", ConfType: ConfTypeUser},
				SubcontractServiceForProduct: &SubcontractItemAdditions{
					ApplyServiceInfo: &SubcontractApplyServiceInfo{
						Date:        ansicStrToDate,
						Address:     "street",
						CityKladrId: "197",
						CityName:    "Birobidzhan",
					},
				},
				Service: &Service{true, false, sync.RWMutex{}},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)

			s.reader = bytes.NewReader(buf)
			s.decoder = msgpack.NewDecoder(s.reader)

			itemAdditions := &ItemAdditions{}
			err = itemAdditions.DecodeMsgpack(s.decoder)

			if err != nil {
				s.EqualError(err, tt.err().Error())
			} else {
				s.Equal(tt.want, itemAdditions)
			}
		})
	}
}

func (s *ItemMsgpackSuite) TestItem_DecodeMsgpack() {
	date := time.Date(2021, 11, 2, 11, 2, 3, 4, &time.Location{})
	ansicStr := date.Format(time.RFC3339)
	ansicStrToDate, _ := time.Parse(time.RFC3339, ansicStr)

	expectedItem := NewItem(
		"", TypeUnknown, "", "", 0,
		0, 0, "", catalog_types.PriceColumnRetail,
	)

	expectedItem.additions = ItemAdditions{}
	itemAdditions := expectedItem.Additions()
	itemAdditions.SetProduct(&ProductItemAdditions{
		isAvailInStore: true, vat: 1, categoryId: 1, isOEM: false,
		availTotal: 1, isCountMoreThenAvailChecked: true, creditPrograms: []catalog_types.CreditProgram{},
		isAvailForDPD: true, isMarked: true, markedPurchaseReason: MarkedPurchaseReasonUnknown,
		isDiscounted: true, categoryName: "name", brandName: "brand", categoryPath: "path", shortName: "short",
	})
	itemAdditions.SetConfiguration(&ConfiguratorItemAdditions{ConfId: "config", ConfType: ConfTypeUser})
	itemAdditions.SetSubcontractServiceForProduct(&SubcontractItemAdditions{
		ApplyServiceInfo: &SubcontractApplyServiceInfo{
			Date:        ansicStrToDate,
			Address:     "street",
			CityKladrId: "197",
			CityName:    "Birobidzhan",
		},
	})
	itemAdditions.SetService(&Service{true, false, sync.RWMutex{}})

	expectedItem.SetAllowResale(NewAllowResale(false, "commodity group name"))

	uniqId := expectedItem.UniqId()

	tests := []struct {
		name string
		obj  interface{}
		err  func() error
		want *Item
	}{
		{
			name: "base type substitution negative",
			obj:  []byte{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item array len`[0]: msgpack: invalid code c4 decoding array length")
			},
			want: &Item{},
		},
		{
			name: "itemId is not string negative",
			obj: []interface{}{
				uniqId,      // uniqId 1
				1,           // itemId 2
				TypeUnknown, // itemType 3
				"",          // parentUniqId 4
				"",          // parentItemId 5
				"",          // name 6
				"",          // image 7
				0,           // count 8
				0,           // price 9
				0,           // bonus 10
				0,           // countMultiplicity 11
				[]int{0},    // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
				[]interface{}{
					false, "commodity group name",
				},
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item itemId`[2]: msgpack: invalid code 1 decoding bytes length")
			},
			want: &Item{},
		},
		{
			name: "itemType is not string negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				1,        // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item itemType`[3]: msgpack: invalid code 1 decoding bytes length")
			},
			want: &Item{},
		},
		{
			name: "parentUniqId is not string negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				1,        // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item parentUniqId`[4]: msgpack: invalid code 1 decoding bytes length")
			},
			want: &Item{},
		},
		{
			name: "parentItemId is not string negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				1,        // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item parentItemId`[5]: msgpack: invalid code 1 decoding bytes length")
			},
			want: &Item{},
		},
		{
			name: "name is not string negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				1,        // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item name`[6]: msgpack: invalid code 1 decoding bytes length")
			},
			want: &Item{},
		},
		{
			name: "image is not string negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				1,        // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item image`[7]: msgpack: invalid code 1 decoding bytes length")
			},
			want: &Item{},
		},
		{
			name: "count is not int negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				"",       // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item count`[8]: msgpack: invalid code a0 decoding int64")
			},
			want: &Item{},
		},
		{
			name: "price is not int negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				"",       // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item price`[9]: msgpack: invalid code a0 decoding int64")
			},
			want: &Item{},
		},
		{
			name: "bonus is not int negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				"",       // price 9
				"",       // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item price`[9]: msgpack: invalid code a0 decoding int64")
			},
			want: &Item{},
		},
		{
			name: "countMultiplicity is not int negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				"",       // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item countMultiplicity`[11]: msgpack: invalid " +
					"code a0 decoding int64")
			},
			want: &Item{},
		},
		{
			name: "problems is not slice negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]byte{},           // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item DecodeArrayLen`[13]: msgpack: invalid code " +
					"c4 decoding array length")
			},
			want: &Item{},
		},
		{
			name: "infos is not map negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				[]byte{},           // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item infos`[14]: msgpack: invalid code c4 decoding " +
					"map length")
			},
			want: &Item{},
		},
		{
			name: "rules is empty negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{},    // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item rules`[15]: can't decode msgpack field `len doesn't match`[0]: len doesn't match: 0")
			},
			want: &Item{},
		},
		{
			name: "additions is empty negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil),              // problems 13
				map[InfoId]*Info{},              // infos 14
				[]interface{}{0},                // rules 15
				[]interface{}{},                 // additions 16
				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item additions`[16]: can't decode msgpack field " +
					"`(basket_item.ItemAdditions) incorrect len`[0]: (basket_item.ItemAdditions) incorrect len: 0")
			},
			want: &Item{},
		},
		{
			name: "permanentProblems is not *Problem_slice negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{1},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]byte{},                        // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item permanentProblems`[17]: msgpack: invalid code c4 decoding array length")
			},
			want: &Item{},
		},
		{
			name: "spaceId is not string negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{1},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				1,                               // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item spaceId`[20]: msgpack: invalid code 1 decoding bytes length")
			},
			want: &Item{},
		},
		{
			name: "commitFingerprint is not string negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{1},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil), // permanentProblems 17
				0,                  // 18 deleted
				0,                  // 19 deleted
				"",                 // spaceId 20
				"",                 // priceColumn 21
				"",                 // commitFingerprint 22
				false,              // isPrepaymentMandatory 23
				false,              // hasFairPrice 24
				false,              // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item priceColumn`[21]: msgpack: invalid code a0 decoding int64")
			},
			want: &Item{},
		},
		{
			name: "commitFingerprint is not string negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{1},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},
				[]interface{}(nil), // permanentProblems 17

				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				1,                               // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item commitFingerprint`[22]: msgpack: invalid " +
					"code 1 decoding bytes length")
			},
			want: &Item{},
		},
		{
			name: "isPrepaymentMandatory is not bool negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{1},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				1,                               // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item isPrepaymentMandatory`[23]: msgpack: " +
					"invalid code 1 decoding bool")
			},
			want: &Item{},
		},
		{
			name: "hasFairPrice is not bool negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{1},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},

				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				"",                              // hasFairPrice 24
				false,                           // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item hasFairPrice`[24]: msgpack: invalid code a0 " +
					"decoding bool")
			},
			want: &Item{},
		},
		{
			name: "ignoreFairPrice is not bool negative",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{1},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},
				[]interface{}(nil), // permanentProblems 17

				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				"",                              // ignoreFairPrice 25
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item ignoreFairPrice`[25]: msgpack: invalid code " +
					"a0 decoding bool")
			},
			want: &Item{},
		},
		{
			name: "empty",
			obj:  []interface{}{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `(basket_item.Item) incorrect len`[0]: " +
					"(basket_item.Item) incorrect len: 0")
			},
			want: &Item{},
		},
		{
			name: "allow resale is empty",
			obj: []interface{}{
				uniqId,   // uniqId 1
				"",       // itemId 2
				"",       // itemType 3
				"",       // parentUniqId 4
				"",       // parentItemId 5
				"",       // name 6
				"",       // image 7
				0,        // count 8
				0,        // price 9
				0,        // bonus 10
				0,        // countMultiplicity 11
				[]int{0}, // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},
				[]interface{}(nil),              // permanentProblems 17
				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
				[]interface{}{},                 // allowResale 26
			},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item allowResale`[26]: can't decode msgpack field " +
					"`Item allowResale`[0]: (basket_item.AllowResale) incorrect len: 0")
			},
			want: &Item{},
		},
		{
			name: "positive",
			obj: []interface{}{
				uniqId,      // uniqId 1
				"",          // itemId 2
				TypeUnknown, // itemType 3
				"",          // parentUniqId 4
				"",          // parentItemId 5
				"",          // name 6
				"",          // image 7
				0,           // count 8
				0,           // price 9
				0,           // bonus 10
				0,           // countMultiplicity 11
				[]int{0},    // deleted 12

				[]interface{}(nil), // problems 13
				map[InfoId]*Info{}, // infos 14
				[]interface{}{0},   // rules 15
				[]interface{}{ // additions 16
					&[]interface{}{
						true, 1, 1, "", 0, false,
						1, true, []catalog_types.CreditProgram{}, "", "",
						true, true, MarkedPurchaseReasonUnknown,
						true, "name", "brand", "path", "short",
					},
					[]interface{}{"config", ConfTypeUser},
					[]interface{}{[]interface{}{ansicStr, "street", "197", "Birobidzhan"}},
					[]interface{}{true, false},
				},
				[]interface{}(nil), // permanentProblems 17

				0,                               // 18 deleted
				0,                               // 19 deleted
				"",                              // spaceId 20
				catalog_types.PriceColumnRetail, // priceColumn 21
				"",                              // commitFingerprint 22
				false,                           // isPrepaymentMandatory 23
				false,                           // hasFairPrice 24
				false,                           // ignoreFairPrice 25
				[]interface{}{ // allowResale 26
					false, "commodity group name",
				},
				[]interface{}{0, 0, 0, ""},
				false,
				false,
				true,
			},
			err: func() error {
				return nil
			},
			want: expectedItem,
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)

			s.reader = bytes.NewReader(buf)
			s.decoder = msgpack.NewDecoder(s.reader)

			item := &Item{}
			err = item.DecodeMsgpack(s.decoder)

			if err != nil {
				s.EqualError(err, tt.err().Error())
			} else {
				s.Equal(tt.want, item)
			}
		})
	}
}

func (s *ItemMsgpackSuite) TestItemAllowResale_DecodeMsgpack() {
	tests := []struct {
		name string
		obj  interface{}
		err  func() error
		want *AllowResale
	}{
		{
			name: "error decode len",
			obj:  []byte{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `AllowResale array len`[0]: msgpack: invalid code " +
					"c4 decoding array length")
			},
			want: &AllowResale{},
		},
		{
			name: "error invalid len",
			obj:  []interface{}{false},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item allowResale`[0]: (basket_item.AllowResale) incorrect len: 1")
			},
			want: &AllowResale{},
		},
		{
			name: "error IsAllow not bool",
			obj:  []interface{}{"not bool", "commodityGroupName"},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item allowResale isAllow`[1]: msgpack: invalid code a8 decoding bool")
			},
			want: &AllowResale{},
		},
		{
			name: "error CommodityGroupName not string",
			obj:  []interface{}{false, 10},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item allowResale commodityGroupName`[2]: msgpack: invalid code a decoding bytes length")
			},
			want: &AllowResale{},
		},
		{
			name: "empty",
			obj:  []interface{}{},
			err: func() error {
				return fmt.Errorf("can't decode msgpack field `Item allowResale`[0]: (basket_item.AllowResale) incorrect len: 0")
			},
			want: &AllowResale{},
		},
		{
			name: "success",
			obj:  []interface{}{false, "commodityGroupName"},
			err: func() error {
				return nil
			},
			want: &AllowResale{isAllow: false, commodityGroupName: "commodityGroupName"},
		},
	}
	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			buf, err := msgpack.Marshal(tt.obj)
			s.Nil(err)

			s.reader = bytes.NewReader(buf)
			s.decoder = msgpack.NewDecoder(s.reader)

			configurator := &AllowResale{}
			err = configurator.DecodeMsgpack(s.decoder)

			if err != nil {
				s.EqualError(err, tt.err().Error())
			} else {
				s.Equal(tt.want, configurator)
			}
		})
	}
}
