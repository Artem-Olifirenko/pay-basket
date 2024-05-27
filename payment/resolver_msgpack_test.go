package payment

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"go.citilink.cloud/order/internal/order"
	"gopkg.in/vmihailenco/msgpack.v2"
	"testing"
)

func TestResolvedId_EncodeMsgpack(t *testing.T) {
	tests := []struct {
		name          string
		obj           ResolvedId
		expectedBytes []byte
	}{
		{
			name:          "with everything empty",
			obj:           ResolvedId{},
			expectedBytes: []byte{0x95, 0x0, 0x0, 0xc0, 0xc2, 0xc2},
		},
		{
			name: "with everything filled",
			obj: ResolvedId{
				id:        order.PaymentIdCash,
				status:    order.AllowStatusAllow,
				isDefault: true,
				isChosen:  true,
				reasons: []*order.DisallowReasonWithInfo{
					order.NewDisallowReasonWithInfo(
						order.DisallowReasonCitilinkCourierDeliveryIdFastNotAvailable,
						order.SubsystemBasket),
					order.NewDisallowReasonWithInfo(
						order.DisallowReasonCitilinkCourierDeliveryIdSameDayNotAvailable,
						order.SubsystemDelivery),
				},
			},
			expectedBytes: []byte{0x95, 0x2, 0x0, 0x92, 0x94, 0x1e, 0xa6, 0x62, 0x61, 0x73, 0x6b, 0x65, 0x74, 0xa0,
				0x91, 0xc0, 0x94, 0x1d, 0xa8, 0x64, 0x65, 0x6c, 0x69, 0x76, 0x65, 0x72, 0x79, 0xa0, 0x91, 0xc0, 0xc3,
				0xc3},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			enc := msgpack.NewEncoder(buf)
			err := test.obj.EncodeMsgpack(enc)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedBytes, buf.Bytes())
		})
	}
}

func TestResolvedId_DecodeMsgpack(t *testing.T) {
	tests := []struct {
		name     string
		expected *ResolvedId
		data     interface{}
		err      string
	}{
		{
			name: "wrong format",
			data: "s",
			err:  "can't decode msgpack field `ResolvedId array len`[0]: msgpack: invalid code a1 decoding array length",
		},
		{
			name: "wrong length",
			data: []interface{}{0},
			err:  "can't decode msgpack field `len doesn't match`[0]: len doesn't match 1",
		},
		{
			name: "wrong id",
			data: []interface{}{
				"order.PaymentIdCashWithCard",
				order.AllowStatusLimited,
				[]*order.DisallowReasonWithInfo{
					order.NewDisallowReasonWithInfo(
						order.DisallowReasonCitilinkCourierDeliveryIdFastNotAvailable,
						order.SubsystemBasket),
					order.NewDisallowReasonWithInfo(
						order.DisallowReasonCitilinkCourierDeliveryIdSameDayNotAvailable,
						order.SubsystemDelivery),
				},
				false,
				false,
			},
			err: "can't decode msgpack field `ResolvedId id`[1]: msgpack: invalid code bb decoding int64",
		},
		{
			name: "wrong status",
			data: []interface{}{
				order.PaymentIdCashWithCard,
				"order.AllowStatusLimited",
				[]*order.DisallowReasonWithInfo{
					order.NewDisallowReasonWithInfo(
						order.DisallowReasonCitilinkCourierDeliveryIdFastNotAvailable,
						order.SubsystemBasket),
					order.NewDisallowReasonWithInfo(
						order.DisallowReasonCitilinkCourierDeliveryIdSameDayNotAvailable,
						order.SubsystemDelivery),
				},
				false,
				false,
			},
			err: "can't decode msgpack field `ResolvedId status`[2]: msgpack: invalid code b8 decoding int64",
		},
		{
			name: "wrong reasons",
			data: []interface{}{
				order.PaymentIdCashWithCard,
				order.AllowStatusLimited,
				"[]*order.DisallowReasonWithInfo{",
				false,
				false,
			},
			err: "can't decode msgpack field `ResolvedId reasons`[3]: msgpack: invalid code d9 decoding array length",
		},
		{
			name: "wrong isDefault",
			data: []interface{}{
				order.PaymentIdCashWithCard,
				order.AllowStatusLimited,
				[]*order.DisallowReasonWithInfo{
					order.NewDisallowReasonWithInfo(
						order.DisallowReasonCitilinkCourierDeliveryIdFastNotAvailable,
						order.SubsystemBasket),
					order.NewDisallowReasonWithInfo(
						order.DisallowReasonCitilinkCourierDeliveryIdSameDayNotAvailable,
						order.SubsystemDelivery),
				},
				"false",
				false,
			},
			err: "can't decode msgpack field `ResolvedId isDefault`[4]: msgpack: invalid code a5 decoding bool",
		},
		{
			name: "wrong isChosen",
			data: []interface{}{
				order.PaymentIdCashWithCard,
				order.AllowStatusLimited,
				[]*order.DisallowReasonWithInfo{
					order.NewDisallowReasonWithInfo(
						order.DisallowReasonCitilinkCourierDeliveryIdFastNotAvailable,
						order.SubsystemBasket),
					order.NewDisallowReasonWithInfo(
						order.DisallowReasonCitilinkCourierDeliveryIdSameDayNotAvailable,
						order.SubsystemDelivery),
				},
				false,
				"false",
			},
			err: "can't decode msgpack field `ResolvedId isChosen`[5]: msgpack: invalid code a5 decoding bool",
		},
		{
			name: "with everything filled",
			expected: &ResolvedId{
				id:     order.PaymentIdCashWithCard,
				status: order.AllowStatusLimited,
				reasons: []*order.DisallowReasonWithInfo{
					order.NewDisallowReasonWithInfo(
						order.DisallowReasonCitilinkCourierDeliveryIdFastNotAvailable,
						order.SubsystemBasket),
				},
				isDefault: false,
				isChosen:  false,
			},
			data: []interface{}{
				order.PaymentIdCashWithCard,
				order.AllowStatusLimited,
				[]*order.DisallowReasonWithInfo{
					order.NewDisallowReasonWithInfo(
						order.DisallowReasonCitilinkCourierDeliveryIdFastNotAvailable,
						order.SubsystemBasket),
				},
				false,
				false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			product := &ResolvedId{}
			buf, err := msgpack.Marshal(test.data)
			assert.Nil(t, err)
			reader := bytes.NewReader(buf)
			decoder := msgpack.NewDecoder(reader)
			err = product.DecodeMsgpack(decoder)
			if test.err != "" {
				assert.EqualError(t, err, test.err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, product)
			}
		})
	}
}
