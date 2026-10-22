// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

type marshalTestCase struct {
	wantData []byte
	values   []any
}

type marshalErrorTestCase struct {
	name         string
	value        any
	wantErrorMsg string
}

// CBOR test data are from https://tools.ietf.org/html/rfc7049#appendix-A.
var marshalTestCases = []marshalTestCase{
	// unsigned integer
	{
		wantData: mustHexDecode("00"),
		values:   []any{uint(0), uint8(0), uint16(0), uint32(0), uint64(0), int(0), int8(0), int16(0), int32(0), int64(0)},
	},
	{
		wantData: mustHexDecode("01"),
		values:   []any{uint(1), uint8(1), uint16(1), uint32(1), uint64(1), int(1), int8(1), int16(1), int32(1), int64(1)},
	},
	{
		wantData: mustHexDecode("0a"),
		values:   []any{uint(10), uint8(10), uint16(10), uint32(10), uint64(10), int(10), int8(10), int16(10), int32(10), int64(10)},
	},
	{
		wantData: mustHexDecode("17"),
		values:   []any{uint(23), uint8(23), uint16(23), uint32(23), uint64(23), int(23), int8(23), int16(23), int32(23), int64(23)},
	},
	{
		wantData: mustHexDecode("1818"),
		values:   []any{uint(24), uint8(24), uint16(24), uint32(24), uint64(24), int(24), int8(24), int16(24), int32(24), int64(24)},
	},
	{
		wantData: mustHexDecode("1819"),
		values:   []any{uint(25), uint8(25), uint16(25), uint32(25), uint64(25), int(25), int8(25), int16(25), int32(25), int64(25)},
	},
	{
		wantData: mustHexDecode("1864"),
		values:   []any{uint(100), uint8(100), uint16(100), uint32(100), uint64(100), int(100), int8(100), int16(100), int32(100), int64(100)},
	},
	{
		wantData: mustHexDecode("18ff"),
		values:   []any{uint(255), uint8(255), uint16(255), uint32(255), uint64(255), int(255), int16(255), int32(255), int64(255)},
	},
	{
		wantData: mustHexDecode("190100"),
		values:   []any{uint(256), uint16(256), uint32(256), uint64(256), int(256), int16(256), int32(256), int64(256)},
	},
	{
		wantData: mustHexDecode("1903e8"),
		values:   []any{uint(1000), uint16(1000), uint32(1000), uint64(1000), int(1000), int16(1000), int32(1000), int64(1000)},
	},
	{
		wantData: mustHexDecode("19ffff"),
		values:   []any{uint(65535), uint16(65535), uint32(65535), uint64(65535), int(65535), int32(65535), int64(65535)},
	},
	{
		wantData: mustHexDecode("1a00010000"),
		values:   []any{uint(65536), uint32(65536), uint64(65536), int(65536), int32(65536), int64(65536)},
	},
	{
		wantData: mustHexDecode("1a000f4240"),
		values:   []any{uint(1000000), uint32(1000000), uint64(1000000), int(1000000), int32(1000000), int64(1000000)},
	},
	{
		wantData: mustHexDecode("1affffffff"),
		values:   []any{uint(4294967295), uint32(4294967295), uint64(4294967295), int64(4294967295)},
	},
	{
		wantData: mustHexDecode("1b000000e8d4a51000"),
		values:   []any{uint64(1000000000000), int64(1000000000000)},
	},
	{
		wantData: mustHexDecode("1bffffffffffffffff"),
		values:   []any{uint64(18446744073709551615)},
	},

	// negative integer
	{
		wantData: mustHexDecode("20"),
		values:   []any{int(-1), int8(-1), int16(-1), int32(-1), int64(-1)},
	},
	{
		wantData: mustHexDecode("29"),
		values:   []any{int(-10), int8(-10), int16(-10), int32(-10), int64(-10)},
	},
	{
		wantData: mustHexDecode("37"),
		values:   []any{int(-24), int8(-24), int16(-24), int32(-24), int64(-24)},
	},
	{
		wantData: mustHexDecode("3818"),
		values:   []any{int(-25), int8(-25), int16(-25), int32(-25), int64(-25)},
	},
	{
		wantData: mustHexDecode("3863"),
		values:   []any{int(-100), int8(-100), int16(-100), int32(-100), int64(-100)},
	},
	{
		wantData: mustHexDecode("38ff"),
		values:   []any{int(-256), int16(-256), int32(-256), int64(-256)},
	},
	{
		wantData: mustHexDecode("390100"),
		values:   []any{int(-257), int16(-257), int32(-257), int64(-257)},
	},
	{
		wantData: mustHexDecode("3903e7"),
		values:   []any{int(-1000), int16(-1000), int32(-1000), int64(-1000)},
	},
	{
		wantData: mustHexDecode("39ffff"),
		values:   []any{int(-65536), int32(-65536), int64(-65536)},
	},
	{
		wantData: mustHexDecode("3a00010000"),
		values:   []any{int(-65537), int32(-65537), int64(-65537)},
	},
	{
		wantData: mustHexDecode("3affffffff"),
		values:   []any{int64(-4294967296)},
	},
	{
		wantData: mustHexDecode("3b0000000100000000"),
		values:   []any{int64(-4294967297)},
	},
	{
		wantData: mustHexDecode("3b7fffffffffffffff"),
		values:   []any{int64(math.MinInt64)},
	},

	// byte string
	{
		wantData: mustHexDecode("40"),
		values:   []any{[]byte{}},
	},
	{
		wantData: mustHexDecode("4401020304"),
		values:   []any{[]byte{1, 2, 3, 4}, [...]byte{1, 2, 3, 4}},
	},

	// text string
	{
		wantData: mustHexDecode("60"),
		values:   []any{""},
	},
	{
		wantData: mustHexDecode("6161"),
		values:   []any{"a"},
	},
	{
		wantData: mustHexDecode("6449455446"),
		values:   []any{"IETF"},
	},
	{
		wantData: mustHexDecode("62225c"),
		values:   []any{"\"\\"},
	},
	{
		wantData: mustHexDecode("62c3bc"),
		values:   []any{"ü"},
	},
	{
		wantData: mustHexDecode("63e6b0b4"),
		values:   []any{"水"},
	},
	{
		wantData: mustHexDecode("64f0908591"),
		values:   []any{"𐅑"},
	},

	// array
	{
		wantData: mustHexDecode("80"),
		values: []any{
			[0]int{},
			[]uint{},
			// []uint8{},
			[]uint16{},
			[]uint32{},
			[]uint64{},
			[]int{},
			[]int8{},
			[]int16{},
			[]int32{},
			[]int64{},
			[]string{},
			[]bool{},
			[]float32{},
			[]float64{},
			[]any{},
		},
	},
	{
		wantData: mustHexDecode("83010203"),
		values: []any{
			[...]int{1, 2, 3},
			[]uint{1, 2, 3},
			// []uint8{1, 2, 3},
			[]uint16{1, 2, 3},
			[]uint32{1, 2, 3},
			[]uint64{1, 2, 3},
			[]int{1, 2, 3},
			[]int8{1, 2, 3},
			[]int16{1, 2, 3},
			[]int32{1, 2, 3},
			[]int64{1, 2, 3},
			[]any{1, 2, 3},
		},
	},
	{
		wantData: mustHexDecode("8301820203820405"),
		values: []any{
			[...]any{1, [...]int{2, 3}, [...]int{4, 5}},
			[]any{1, []uint{2, 3}, []uint{4, 5}},
			[]any{1, []uint16{2, 3}, []uint16{4, 5}},
			[]any{1, []uint32{2, 3}, []uint32{4, 5}},
			[]any{1, []uint64{2, 3}, []uint64{4, 5}},
			[]any{1, []int{2, 3}, []int{4, 5}},
			[]any{1, []int8{2, 3}, []int8{4, 5}},
			[]any{1, []int16{2, 3}, []int16{4, 5}},
			[]any{1, []int32{2, 3}, []int32{4, 5}},
			[]any{1, []int64{2, 3}, []int64{4, 5}},
			[]any{1, []any{2, 3}, []any{4, 5}},
		},
	},
	{
		wantData: mustHexDecode("98190102030405060708090a0b0c0d0e0f101112131415161718181819"),
		values: []any{
			[...]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			// []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]uint16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
		},
	},
	{
		wantData: mustHexDecode("826161a161626163"),
		values: []any{
			[...]any{"a", map[string]string{"b": "c"}},
			[]any{"a", map[string]string{"b": "c"}},
			[]any{"a", map[any]any{"b": "c"}},
		},
	},

	// map
	{
		wantData: mustHexDecode("a0"),
		values: []any{
			map[uint]bool{},
			map[uint8]bool{},
			map[uint16]bool{},
			map[uint32]bool{},
			map[uint64]bool{},
			map[int]bool{},
			map[int8]bool{},
			map[int16]bool{},
			map[int32]bool{},
			map[int64]bool{},
			map[float32]bool{},
			map[float64]bool{},
			map[bool]bool{},
			map[string]bool{},
			map[any]any{},
		},
	},
	{
		wantData: mustHexDecode("a201020304"),
		values: []any{
			map[uint]uint{3: 4, 1: 2},
			map[uint8]uint8{3: 4, 1: 2},
			map[uint16]uint16{3: 4, 1: 2},
			map[uint32]uint32{3: 4, 1: 2},
			map[uint64]uint64{3: 4, 1: 2},
			map[int]int{3: 4, 1: 2},
			map[int8]int8{3: 4, 1: 2},
			map[int16]int16{3: 4, 1: 2},
			map[int32]int32{3: 4, 1: 2},
			map[int64]int64{3: 4, 1: 2},
			map[any]any{3: 4, 1: 2},
		},
	},
	{
		wantData: mustHexDecode("a26161016162820203"),
		values: []any{
			map[string]any{"a": 1, "b": []any{2, 3}},
			map[any]any{"b": []any{2, 3}, "a": 1},
		},
	},
	{
		wantData: mustHexDecode("a56161614161626142616361436164614461656145"),
		values: []any{
			map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"},
			map[any]any{"b": "B", "a": "A", "c": "C", "e": "E", "d": "D"},
		},
	},

	// tag
	{
		wantData: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
		values: []any{
			Tag{0, "2013-03-21T20:04:00Z"},
			RawTag{0, mustHexDecode("74323031332d30332d32315432303a30343a30305a")},
		},
	}, // 0: standard date/time
	{
		wantData: mustHexDecode("c11a514b67b0"),
		values: []any{
			Tag{1, uint64(1363896240)},
			RawTag{1, mustHexDecode("1a514b67b0")},
		},
	}, // 1: epoch-based date/time
	{
		wantData: mustHexDecode("c249010000000000000000"),
		values: []any{
			mustBigInt("18446744073709551616"),
			Tag{2, []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			RawTag{2, mustHexDecode("49010000000000000000")},
		},
	}, // 2: positive bignum: 18446744073709551616
	{
		wantData: mustHexDecode("c349010000000000000000"),
		values: []any{
			mustBigInt("-18446744073709551617"),
			Tag{3, []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			RawTag{3, mustHexDecode("49010000000000000000")},
		},
	}, // 3: negative bignum: -18446744073709551617
	{
		wantData: mustHexDecode("c1fb41d452d9ec200000"),
		values: []any{
			Tag{1, float64(1363896240.5)},
			RawTag{1, mustHexDecode("fb41d452d9ec200000")},
		},
	}, // 1: epoch-based date/time
	{
		wantData: mustHexDecode("d74401020304"),
		values: []any{
			Tag{23, []byte{0x01, 0x02, 0x03, 0x04}},
			RawTag{23, mustHexDecode("4401020304")},
		},
	}, // 23: expected conversion to base16 encoding
	{
		wantData: mustHexDecode("d818456449455446"),
		values: []any{
			Tag{24, []byte{0x64, 0x49, 0x45, 0x54, 0x46}},
			RawTag{24, mustHexDecode("456449455446")},
		},
	}, // 24: encoded cborBytes data item
	{
		wantData: mustHexDecode("d82076687474703a2f2f7777772e6578616d706c652e636f6d"),
		values: []any{
			Tag{32, "http://www.example.com"},
			RawTag{32, mustHexDecode("76687474703a2f2f7777772e6578616d706c652e636f6d")},
		},
	}, // 32: URI

	// primitives
	{
		wantData: mustHexDecode("f4"),
		values:   []any{false},
	},
	{
		wantData: mustHexDecode("f5"),
		values:   []any{true},
	},
	{
		wantData: mustHexDecode("f6"),
		values:   []any{nil, []byte(nil), []int(nil), map[uint]bool(nil), (*int)(nil), io.Reader(nil)},
	},
	// simple values
	{
		wantData: mustHexDecode("e0"),
		values:   []any{SimpleValue(0)},
	},
	{
		wantData: mustHexDecode("f0"),
		values:   []any{SimpleValue(16)},
	},
	{
		wantData: mustHexDecode("f820"),
		values:   []any{SimpleValue(32)},
	},
	{
		wantData: mustHexDecode("f8ff"),
		values:   []any{SimpleValue(255)},
	},
	// nan, positive and negative inf
	{
		wantData: mustHexDecode("f97c00"),
		values:   []any{math.Inf(1)},
	},
	{
		wantData: mustHexDecode("f97e00"),
		values:   []any{math.NaN()},
	},
	{
		wantData: mustHexDecode("f9fc00"),
		values:   []any{math.Inf(-1)},
	},
	// float32
	{
		wantData: mustHexDecode("fa47c35000"),
		values:   []any{float32(100000.0)},
	},
	{
		wantData: mustHexDecode("fa7f7fffff"),
		values:   []any{float32(3.4028234663852886e+38)},
	},
	// float64
	{
		wantData: mustHexDecode("fb3ff199999999999a"),
		values:   []any{float64(1.1)},
	},
	{
		wantData: mustHexDecode("fb7e37e43c8800759c"),
		values:   []any{float64(1.0e+300)},
	},
	{
		wantData: mustHexDecode("fbc010666666666666"),
		values:   []any{float64(-4.1)},
	},

	// More testcases not covered by https://tools.ietf.org/html/rfc7049#appendix-A.
	{
		wantData: mustHexDecode("d83dd183010203"), // 61(17([1, 2, 3])), nested tags 61 and 17
		values: []any{
			Tag{61, Tag{17, []any{uint64(1), uint64(2), uint64(3)}}},
			RawTag{61, mustHexDecode("d183010203")},
		},
	},

	{
		wantData: mustHexDecode("83f6f6f6"),
		values:   []any{[]any{nil, nil, nil}},
	}, // [nil, nil, nil]
}

func TestMarshal(t *testing.T) {
	testMarshalWithFP(t, func(t *testing.T) {
		testMarshal(t, marshalTestCases)
	})
}

func TestInvalidTypeMarshal(t *testing.T) {
	testMarshalWithFP(t, testInvalidTypeMarshal)
}

func testInvalidTypeMarshal(t *testing.T) {
	type s1 struct {
		Chan chan bool
	}
	type s2 struct {
		_    struct{} `cbor:",toarray"`
		Chan chan bool
	}
	var marshalErrorTestCases = []marshalErrorTestCase{
		{
			name:         "channel cannot be marshaled",
			value:        make(chan bool),
			wantErrorMsg: "cbor: unsupported type: chan bool",
		},
		{
			name:         "slice of channel cannot be marshaled",
			value:        make([]chan bool, 10),
			wantErrorMsg: "cbor: unsupported type: []chan bool",
		},
		{
			name:         "slice of pointer to channel cannot be marshaled",
			value:        make([]*chan bool, 10),
			wantErrorMsg: "cbor: unsupported type: []*chan bool",
		},
		{
			name:         "map of channel cannot be marshaled",
			value:        make(map[string]chan bool),
			wantErrorMsg: "cbor: unsupported type: map[string]chan bool",
		},
		{
			name:         "struct of channel cannot be marshaled",
			value:        s1{},
			wantErrorMsg: "cbor: unsupported type: cbor.s1",
		},
		{
			name:         "struct of channel cannot be marshaled",
			value:        s2{},
			wantErrorMsg: "cbor: unsupported type: cbor.s2",
		},
		{
			name:         "function cannot be marshaled",
			value:        func(i int) int { return i * i },
			wantErrorMsg: "cbor: unsupported type: func(int) int",
		},
		{
			name:         "complex cannot be marshaled",
			value:        complex(100, 8),
			wantErrorMsg: "cbor: unsupported type: complex128",
		},
	}
	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	for _, tc := range marshalErrorTestCases {
		t.Run(tc.name, func(t *testing.T) {
			v := tc.value
			b, err := Marshal(&v)
			if err == nil {
				t.Errorf("Marshal(%v) didn't return an error, want error %q", tc.value, tc.wantErrorMsg)
			} else if _, ok := err.(*UnsupportedTypeError); !ok {
				t.Errorf("Marshal(%v) error type %T, want *UnsupportedTypeError", tc.value, err)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Marshal(%v) error %q, want %q", tc.value, err.Error(), tc.wantErrorMsg)
			} else if b != nil {
				t.Errorf("Marshal(%v) = 0x%x, want nil", tc.value, b)
			}

			v = tc.value
			b, err = em.Marshal(&v)
			if err == nil {
				t.Errorf("Marshal(%v) didn't return an error, want error %q", tc.value, tc.wantErrorMsg)
			} else if _, ok := err.(*UnsupportedTypeError); !ok {
				t.Errorf("Marshal(%v) error type %T, want *UnsupportedTypeError", tc.value, err)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Marshal(%v) error %q, want %q", tc.value, err.Error(), tc.wantErrorMsg)
			} else if b != nil {
				t.Errorf("Marshal(%v) = 0x%x, want nil", tc.value, b)
			}
		})
	}
}

func TestMarshalLargeByteString(t *testing.T) {
	testMarshalWithFP(t, testMarshalLargeByteString)
}

func testMarshalLargeByteString(t *testing.T) {
	// []byte{100, 100, 100, ...}
	lengths := []int{0, 1, 2, 22, 23, 24, 254, 255, 256, 65534, 65535, 65536, 10000000}
	testCases := make([]marshalTestCase, len(lengths))
	for i, length := range lengths {
		data := bytes.NewBuffer(encodeCborHeader(cborTypeByteString, uint64(length))) //nolint:gosec
		value := make([]byte, length)
		for j := range length {
			data.WriteByte(100)
			value[j] = 100
		}
		testCases[i] = marshalTestCase{data.Bytes(), []any{value}}
	}

	testMarshal(t, testCases)
}

func TestMarshalLargeTextString(t *testing.T) {
	testMarshalWithFP(t, testMarshalLargeTextString)
}

func testMarshalLargeTextString(t *testing.T) {
	// "ddd..."
	lengths := []int{0, 1, 2, 22, 23, 24, 254, 255, 256, 65534, 65535, 65536, 10000000}
	testCases := make([]marshalTestCase, len(lengths))
	for i, length := range lengths {
		data := bytes.NewBuffer(encodeCborHeader(cborTypeTextString, uint64(length))) //nolint:gosec
		value := make([]byte, length)
		for j := range length {
			data.WriteByte(100)
			value[j] = 100
		}
		testCases[i] = marshalTestCase{data.Bytes(), []any{string(value)}}
	}

	testMarshal(t, testCases)
}

func TestMarshalLargeArray(t *testing.T) {
	testMarshalWithFP(t, testMarshalLargeArray)
}

func testMarshalLargeArray(t *testing.T) {
	// []string{"水", "水", "水", ...}
	lengths := []int{0, 1, 2, 22, 23, 24, 254, 255, 256, 65534, 65535, 65536, 131072}
	testCases := make([]marshalTestCase, len(lengths))
	for i, length := range lengths {
		data := bytes.NewBuffer(encodeCborHeader(cborTypeArray, uint64(length))) //nolint:gosec
		value := make([]string, length)
		for j := range length {
			data.Write([]byte{0x63, 0xe6, 0xb0, 0xb4})
			value[j] = "水"
		}
		testCases[i] = marshalTestCase{data.Bytes(), []any{value}}
	}

	testMarshal(t, testCases)
}

func TestMarshalLargeMapCanonical(t *testing.T) {
	testMarshalWithFP(t, testMarshalLargeMapCanonical)
}

func testMarshalLargeMapCanonical(t *testing.T) {
	// map[int]int {0:0, 1:1, 2:2, ...}
	lengths := []int{0, 1, 2, 22, 23, 24, 254, 255, 256, 65534, 65535, 65536, 131072}
	testCases := make([]marshalTestCase, len(lengths))
	for i, length := range lengths {
		data := bytes.NewBuffer(encodeCborHeader(cborTypeMap, uint64(length))) //nolint:gosec
		value := make(map[int]int, length)
		for j := range length {
			d := encodeCborHeader(cborTypePositiveInt, uint64(j))
			data.Write(d)
			data.Write(d)
			value[j] = j
		}
		testCases[i] = marshalTestCase{data.Bytes(), []any{value}}
	}

	testMarshal(t, testCases)
}

func TestMarshalLargeMap(t *testing.T) {
	testMarshalWithFP(t, testMarshalLargeMap)
}

func testMarshalLargeMap(t *testing.T) {
	// map[int]int {0:0, 1:1, 2:2, ...}
	lengths := []int{0, 1, 2, 22, 23, 24, 254, 255, 256, 65534, 65535, 65536, 131072}
	for _, length := range lengths {
		m1 := make(map[int]int, length)
		for i := range length {
			m1[i] = i
		}

		data, err := Marshal(m1)
		if err != nil {
			t.Fatalf("Marshal(%v) returned error %v", m1, err)
		}

		m2 := make(map[int]int)
		if err = unmarshal(t, data, &m2); err != nil {
			t.Fatalf("Unmarshal(0x%x) returned error %v", data, err)
		}

		if !reflect.DeepEqual(m1, m2) {
			t.Errorf("Unmarshal() = %v, want %v", m2, m1)
		}
	}
}

func testMarshal(t *testing.T, testCases []marshalTestCase) {
	t.Helper()

	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	bem, err := EncOptions{Sort: SortCanonical}.UserBufferEncMode()
	if err != nil {
		t.Errorf("UserBufferEncMode() returned an error %v", err)
	}
	for _, tc := range testCases {
		for _, value := range tc.values {
			// Encode value using default options
			if _, err := Marshal(value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", value, err)
			}

			// Encode value to provided buffer using default options
			var buf1 bytes.Buffer
			if err := MarshalToBuffer(value, &buf1); err != nil {
				t.Errorf("MarshalToBuffer(%v) returned error %v", value, err)
			}

			// Encode value using specified options
			if b, err := em.Marshal(value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", value, err)
			} else if !bytes.Equal(b, tc.wantData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", value, b, tc.wantData)
			}

			// Encode value to provided buffer using specified options
			var buf2 bytes.Buffer
			if err := bem.MarshalToBuffer(value, &buf2); err != nil {
				t.Errorf("MarshalToBuffer(%v) returned error %v", value, err)
			} else if !bytes.Equal(buf2.Bytes(), tc.wantData) {
				t.Errorf("MarshalToBuffer(%v) = 0x%x, want 0x%x", value, buf2.Bytes(), tc.wantData)
			}
		}
		r := RawMessage(tc.wantData)
		if b, err := Marshal(r); err != nil {
			t.Errorf("Marshal(%v) returned error %v", r, err)
		} else if !bytes.Equal(b, r) {
			t.Errorf("Marshal(%v) returned %v, want %v", r, b, r)
		}
	}
}

type inner struct {
	X, Y, z int64
}

type outer struct {
	IntField          int
	FloatField        float32
	BoolField         bool
	StringField       string
	ByteStringField   []byte
	ArrayField        []string
	MapField          map[string]bool
	NestedStructField *inner
	unexportedField   int64
}

func TestMarshalStruct(t *testing.T) {
	testMarshalWithFP(t, testMarshalStruct)
}

func testMarshalStruct(t *testing.T) {
	v1 := outer{
		IntField:          123,
		FloatField:        100000.0,
		BoolField:         true,
		StringField:       "test",
		ByteStringField:   []byte{1, 3, 5},
		ArrayField:        []string{"hello", "world"},
		MapField:          map[string]bool{"afternoon": false, "morning": true},
		NestedStructField: &inner{X: 1000, Y: 1000000, z: 10000000},
		unexportedField:   6,
	}
	unmarshalWant := outer{
		IntField:          123,
		FloatField:        100000.0,
		BoolField:         true,
		StringField:       "test",
		ByteStringField:   []byte{1, 3, 5},
		ArrayField:        []string{"hello", "world"},
		MapField:          map[string]bool{"afternoon": false, "morning": true},
		NestedStructField: &inner{X: 1000, Y: 1000000},
	}

	data, err := Marshal(v1)
	if err != nil {
		t.Fatalf("Marshal(%v) returned error %v", v1, err)
	}

	var v2 outer
	if err = unmarshal(t, data, &v2); err != nil {
		t.Fatalf("Unmarshal(0x%x) returned error %v", data, err)
	}

	if !reflect.DeepEqual(unmarshalWant, v2) {
		t.Errorf("Unmarshal() = %v, want %v", v2, unmarshalWant)
	}
}

// TestMarshalStructVariableLength tests marshaling structs that can encode to CBOR maps of varying
// size depending on their field contents.
func TestMarshalStructVariableLength(t *testing.T) {
	testMarshalWithFP(t, testMarshalStructVariableLength)
}

func testMarshalStructVariableLength(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   any
		want []byte
	}{
		{
			name: "zero out of one items",
			in: struct {
				F int `cbor:",omitempty"`
			}{},
			want: mustHexDecode("a0"),
		},
		{
			name: "one out of one items",
			in: struct {
				F int `cbor:",omitempty"`
			}{F: 1},
			want: mustHexDecode("a1614601"),
		},
		{
			name: "23 out of 24 items",
			in: struct {
				A, B, C, D, E, F, G, H, I, J, K, L, M, N, O, P, Q, R, S, T, U, V, W int
				X                                                                   int `cbor:",omitempty"`
			}{},
			want: mustHexDecode("b7614100614200614300614400614500614600614700614800614900614a00614b00614c00614d00614e00614f00615000615100615200615300615400615500615600615700"),
		},
		{
			name: "24 out of 24 items",
			in: struct {
				A, B, C, D, E, F, G, H, I, J, K, L, M, N, O, P, Q, R, S, T, U, V, W int
				X                                                                   int `cbor:",omitempty"`
			}{X: 1},
			want: mustHexDecode("b818614100614200614300614400614500614600614700614800614900614a00614b00614c00614d00614e00614f00615000615100615200615300615400615500615600615700615801"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(tc.want, got) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.in, got, tc.want)
			}
		})
	}
}

func TestMarshalStructCanonical(t *testing.T) {
	testMarshalWithFP(t, testMarshalStructCanonical)
}

func testMarshalStructCanonical(t *testing.T) {
	v := outer{
		IntField:          123,
		FloatField:        100000.0,
		BoolField:         true,
		StringField:       "test",
		ByteStringField:   []byte{1, 3, 5},
		ArrayField:        []string{"hello", "world"},
		MapField:          map[string]bool{"afternoon": false, "morning": true},
		NestedStructField: &inner{X: 1000, Y: 1000000, z: 10000000},
		unexportedField:   6,
	}
	var data bytes.Buffer
	data.WriteByte(byte(cborTypeMap) | 8) // CBOR header: map type with 8 items (exported fields)

	data.WriteByte(byte(cborTypeTextString) | 8) // "IntField"
	data.WriteString("IntField")
	data.WriteByte(byte(cborTypePositiveInt) | 24)
	data.WriteByte(123)

	data.WriteByte(byte(cborTypeTextString) | 8) // "MapField"
	data.WriteString("MapField")
	data.WriteByte(byte(cborTypeMap) | 2)
	data.WriteByte(byte(cborTypeTextString) | 7)
	data.WriteString("morning")
	data.WriteByte(byte(cborTypePrimitives) | 21)
	data.WriteByte(byte(cborTypeTextString) | 9)
	data.WriteString("afternoon")
	data.WriteByte(byte(cborTypePrimitives) | 20)

	data.WriteByte(byte(cborTypeTextString) | 9) // "BoolField"
	data.WriteString("BoolField")
	data.WriteByte(byte(cborTypePrimitives) | 21)

	data.WriteByte(byte(cborTypeTextString) | 10) // "ArrayField"
	data.WriteString("ArrayField")
	data.WriteByte(byte(cborTypeArray) | 2)
	data.WriteByte(byte(cborTypeTextString) | 5)
	data.WriteString("hello")
	data.WriteByte(byte(cborTypeTextString) | 5)
	data.WriteString("world")

	data.WriteByte(byte(cborTypeTextString) | 10) // "FloatField"
	data.WriteString("FloatField")
	data.Write([]byte{0xfa, 0x47, 0xc3, 0x50, 0x00})

	data.WriteByte(byte(cborTypeTextString) | 11) // "StringField"
	data.WriteString("StringField")
	data.WriteByte(byte(cborTypeTextString) | 4)
	data.WriteString("test")

	data.WriteByte(byte(cborTypeTextString) | 15) // "ByteStringField"
	data.WriteString("ByteStringField")
	data.WriteByte(byte(cborTypeByteString) | 3)
	data.Write([]byte{1, 3, 5})

	data.WriteByte(byte(cborTypeTextString) | 17) // "NestedStructField"
	data.WriteString("NestedStructField")
	data.WriteByte(byte(cborTypeMap) | 2)
	data.WriteByte(byte(cborTypeTextString) | 1)
	data.WriteString("X")
	data.WriteByte(byte(cborTypePositiveInt) | 25)
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(1000))
	data.Write(b)
	data.WriteByte(byte(cborTypeTextString) | 1)
	data.WriteString("Y")
	data.WriteByte(byte(cborTypePositiveInt) | 26)
	b = make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(1000000))
	data.Write(b)

	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %q", err)
	}
	if b, err := em.Marshal(v); err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, data.Bytes()) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, data.Bytes())
	}
}

func TestMarshalNullPointerToEmbeddedStruct(t *testing.T) {
	testMarshalWithFP(t, testMarshalNullPointerToEmbeddedStruct)
}

func testMarshalNullPointerToEmbeddedStruct(t *testing.T) {
	type (
		T1 struct {
			X int
		}
		T2 struct {
			*T1
		}
	)
	v := T2{}
	wantCborData := []byte{0xa0} // {}
	data, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%v) returned error %v", v, err)
	}
	if !bytes.Equal(wantCborData, data) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, data, wantCborData)
	}
}

func TestMarshalNullPointerToStruct(t *testing.T) {
	testMarshalWithFP(t, testMarshalNullPointerToStruct)
}

func testMarshalNullPointerToStruct(t *testing.T) {
	type (
		T1 struct {
			X int
		}
		T2 struct {
			T *T1
		}
	)
	v := T2{}
	wantCborData := []byte{0xa1, 0x61, 0x54, 0xf6} // {X: nil}
	data, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%v) returned error %v", v, err)
	}
	if !bytes.Equal(wantCborData, data) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, data, wantCborData)
	}
}

// Struct fields encoding follows the same struct fields visibility
// rules used by JSON encoding package.  Some struct types are from
// tests in JSON encoding package to ensure that the same rules are
// followed.
func TestAnonymousFields1(t *testing.T) {
	testMarshalWithFP(t, testAnonymousFields1)
}

func testAnonymousFields1(t *testing.T) {
	// Fields (T1.X, T2.X) with the same name at the same level are ignored
	type (
		T1 struct{ x, X int }
		T2 struct{ x, X int }
		T  struct {
			T1
			T2
		}
	)
	v := T{T1{1, 2}, T2{3, 4}}
	want := []byte{0xa0} // {}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}
}

func TestAnonymousFields2(t *testing.T) {
	testMarshalWithFP(t, testAnonymousFields2)
}

func testAnonymousFields2(t *testing.T) {
	// Field (T.X) with the same name at a less nested level is serialized
	type (
		T1 struct{ x, X int }
		T2 struct{ x, X int }
		T  struct {
			T1
			T2
			x, X int
		}
	)
	v := T{T1{1, 2}, T2{3, 4}, 5, 6}
	want := []byte{0xa1, 0x61, 0x58, 0x06} // {X:6}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{X: 6}
	if err := unmarshal(t, b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestAnonymousFields3(t *testing.T) {
	testMarshalWithFP(t, testAnonymousFields3)
}

func testAnonymousFields3(t *testing.T) {
	// Unexported embedded field (myInt) of non-struct type is ignored
	type (
		myInt int
		T     struct {
			myInt
		}
	)
	v := T{5}
	want := []byte{0xa0} // {}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}
}

func TestAnonymousFields4(t *testing.T) {
	testMarshalWithFP(t, testAnonymousFields4)
}

func testAnonymousFields4(t *testing.T) {
	// Exported embedded field (MyInt) of non-struct type is serialized
	type (
		MyInt int
		T     struct {
			MyInt
		}
	)
	v := T{5}
	want := []byte{0xa1, 0x65, 0x4d, 0x79, 0x49, 0x6e, 0x74, 0x05} // {MyInt: 5}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	if err = unmarshal(t, b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v, v2) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v, v, v2, v2)
	}
}

func TestAnonymousFields5(t *testing.T) {
	testMarshalWithFP(t, testAnonymousFields5)
}

func testAnonymousFields5(t *testing.T) {
	// Unexported embedded field (*myInt) of pointer to non-struct type is ignored
	type (
		myInt int
		T     struct {
			*myInt
		}
	)
	v := T{new(myInt)}
	*v.myInt = 5
	want := []byte{0xa0} // {}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}
}

func TestAnonymousFields6(t *testing.T) {
	testMarshalWithFP(t, testAnonymousFields6)
}

func testAnonymousFields6(t *testing.T) {
	// Exported embedded field (*MyInt) of pointer to non-struct type should be serialized
	type (
		MyInt int
		T     struct {
			*MyInt
		}
	)
	v := T{new(MyInt)}
	*v.MyInt = 5
	want := []byte{0xa1, 0x65, 0x4d, 0x79, 0x49, 0x6e, 0x74, 0x05} // {MyInt: 5}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	if err = unmarshal(t, b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v, v2) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v, v, v2, v2)
	}
}

func TestAnonymousFields7(t *testing.T) {
	testMarshalWithFP(t, testAnonymousFields7)
}

func testAnonymousFields7(t *testing.T) {
	// Exported fields (t1.X, T2.Y) of embedded structs should have their exported fields be serialized
	type (
		t1 struct{ x, X int }
		T2 struct{ y, Y int }
		T  struct {
			t1
			T2
		}
	)
	v := T{t1{1, 2}, T2{3, 4}}
	want := []byte{0xa2, 0x61, 0x58, 0x02, 0x61, 0x59, 0x04} // {X:2, Y:4}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{t1{X: 2}, T2{Y: 4}}
	if err = unmarshal(t, b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestAnonymousFields8(t *testing.T) {
	testMarshalWithFP(t, testAnonymousFields8)
}

func testAnonymousFields8(t *testing.T) {
	// Exported fields of pointers (t1.X, T2.Y)
	type (
		t1 struct{ x, X int }
		T2 struct{ y, Y int }
		T  struct {
			*t1
			*T2
		}
	)
	v := T{&t1{1, 2}, &T2{3, 4}}
	want := []byte{0xa2, 0x61, 0x58, 0x02, 0x61, 0x59, 0x04} // {X:2, Y:4}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	// v1 cannot be unmarshaled to because reflect cannot allocate unexported field s1.
	var v1 T
	wantErrorMsg := "cannot set embedded pointer to unexported struct"
	wantV := T{T2: &T2{Y: 4}}
	err = unmarshal(t, b, &v1)
	if err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want error %q", b, wantErrorMsg)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error %q", b, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(v1, wantV) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", b, v1, v1, wantV, wantV)
	}

	// v2 can be unmarshaled to because unexported field t1 is already allocated.
	var v2 T
	v2.t1 = &t1{}
	unmarshalWant := T{&t1{X: 2}, &T2{Y: 4}}
	if err = unmarshal(t, b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestAnonymousFields9(t *testing.T) {
	testMarshalWithFP(t, testAnonymousFields9)
}

func testAnonymousFields9(t *testing.T) {
	// Multiple levels of nested anonymous fields
	type (
		MyInt1 int
		MyInt2 int
		myInt  int
		t2     struct {
			MyInt2
			myInt
		}
		t1 struct {
			MyInt1
			myInt
			t2
		}
		T struct {
			t1
			myInt
		}
	)
	v := T{t1{1, 2, t2{3, 4}}, 6}
	want := []byte{0xa2, 0x66, 0x4d, 0x79, 0x49, 0x6e, 0x74, 0x31, 0x01, 0x66, 0x4d, 0x79, 0x49, 0x6e, 0x74, 0x32, 0x03} // {MyInt1: 1, MyInt2: 3}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{t1: t1{MyInt1: 1, t2: t2{MyInt2: 3}}}
	if err = unmarshal(t, b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestAnonymousFields10(t *testing.T) {
	testMarshalWithFP(t, testAnonymousFields10)
}

func testAnonymousFields10(t *testing.T) {
	// Fields of the same struct type at the same level
	type (
		t3 struct {
			Z int
		}
		t1 struct {
			X int
			t3
		}
		t2 struct {
			Y int
			t3
		}
		T struct {
			t1
			t2
		}
	)
	v := T{t1{1, t3{2}}, t2{3, t3{4}}}
	want := []byte{0xa2, 0x61, 0x58, 0x01, 0x61, 0x59, 0x03} // {X: 1, Y: 3}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{t1: t1{X: 1}, t2: t2{Y: 3}}
	if err = unmarshal(t, b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestAnonymousFields11(t *testing.T) {
	testMarshalWithFP(t, testAnonymousFields11)
}

func testAnonymousFields11(t *testing.T) {
	// Fields (T.t2.X, T.t1.t2.X) of the same struct type at different levels
	type (
		t2 struct {
			X int
		}
		t1 struct {
			Y int
			t2
		}
		T struct {
			t1
			t2
		}
	)
	v := T{t1{1, t2{2}}, t2{3}}
	want := []byte{0xa2, 0x61, 0x59, 0x01, 0x61, 0x58, 0x03} // {Y: 1, X: 3}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{t1: t1{Y: 1}, t2: t2{X: 3}}
	if err = unmarshal(t, b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestOmitAndRenameStructField(t *testing.T) {
	testMarshalWithFP(t, testOmitAndRenameStructField)
}

func testOmitAndRenameStructField(t *testing.T) {
	type T struct {
		I   int // never omit
		Io  int `cbor:",omitempty"` // omit empty
		Iao int `cbor:"-"`          // always omit
		R   int `cbor:"omitempty"`  // renamed to omitempty
	}

	v1 := T{}
	// {"I": 0, "omitempty": 0}
	want1 := []byte{0xa2,
		0x61, 0x49, 0x00,
		0x69, 0x6f, 0x6d, 0x69, 0x74, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x00}

	v2 := T{I: 1, Io: 2, Iao: 0, R: 3}
	// {"I": 1, "Io": 2, "omitempty": 3}
	want2 := []byte{0xa3,
		0x61, 0x49, 0x01,
		0x62, 0x49, 0x6f, 0x02,
		0x69, 0x6f, 0x6d, 0x69, 0x74, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x03}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testCases := []roundTripTestCase{
		{"default values", v1, want1},
		{"non-default values", v2, want2}}
	testRoundTrip(t, testCases, em, dm)
}

func TestOmitEmptyForBuiltinType(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForBuiltinType)
}

func testOmitEmptyForBuiltinType(t *testing.T) {
	type T struct {
		B     bool           `cbor:"b"`
		Bo    bool           `cbor:"bo,omitempty"`
		UI    uint           `cbor:"ui"`
		UIo   uint           `cbor:"uio,omitempty"`
		I     int            `cbor:"i"`
		Io    int            `cbor:"io,omitempty"`
		F     float64        `cbor:"f"`
		Fo    float64        `cbor:"fo,omitempty"`
		S     string         `cbor:"s"`
		So    string         `cbor:"so,omitempty"`
		Slc   []string       `cbor:"slc"`
		Slco  []string       `cbor:"slco,omitempty"`
		M     map[int]string `cbor:"m"`
		Mo    map[int]string `cbor:"mo,omitempty"`
		P     *int           `cbor:"p"`
		Po    *int           `cbor:"po,omitempty"`
		Intf  any            `cbor:"intf"`
		Intfo any            `cbor:"intfo,omitempty"`
	}

	v := T{}
	// {"b": false, "ui": 0, "i":0, "f": 0, "s": "", "slc": null, "m": {}, "p": nil, "intf": nil }
	want := []byte{0xa9,
		0x61, 0x62, 0xf4,
		0x62, 0x75, 0x69, 0x00,
		0x61, 0x69, 0x00,
		0x61, 0x66, 0xfb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x61, 0x73, 0x60,
		0x63, 0x73, 0x6c, 0x63, 0xf6,
		0x61, 0x6d, 0xf6,
		0x61, 0x70, 0xf6,
		0x64, 0x69, 0x6e, 0x74, 0x66, 0xf6,
	}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"default values", v, want}}, em, dm)
}

func TestOmitEmptyForAnonymousStruct(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForAnonymousStruct)
}

func testOmitEmptyForAnonymousStruct(t *testing.T) {
	type T struct {
		Str  struct{} `cbor:"str"`
		Stro struct{} `cbor:"stro,omitempty"`
	}

	v := T{}
	want := []byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0} // {"str": {}}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"default values", v, want}}, em, dm)
}

func TestOmitEmptyForStruct1(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForStruct1)
}

func testOmitEmptyForStruct1(t *testing.T) {
	type T1 struct {
		Bo    bool           `cbor:"bo,omitempty"`
		UIo   uint           `cbor:"uio,omitempty"`
		Io    int            `cbor:"io,omitempty"`
		Fo    float64        `cbor:"fo,omitempty"`
		So    string         `cbor:"so,omitempty"`
		Slco  []string       `cbor:"slco,omitempty"`
		Mo    map[int]string `cbor:"mo,omitempty"`
		Po    *int           `cbor:"po,omitempty"`
		Intfo any            `cbor:"intfo,omitempty"`
	}
	type T struct {
		Str  T1 `cbor:"str"`
		Stro T1 `cbor:"stro,omitempty"`
	}

	v := T{}
	want := []byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0} // {"str": {}}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"default values", v, want}}, em, dm)
}

func TestOmitEmptyForStruct2(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForStruct2)
}

func testOmitEmptyForStruct2(t *testing.T) {
	type T1 struct {
		Bo    bool           `cbor:"bo,omitempty"`
		UIo   uint           `cbor:"uio,omitempty"`
		Io    int            `cbor:"io,omitempty"`
		Fo    float64        `cbor:"fo,omitempty"`
		So    string         `cbor:"so,omitempty"`
		Slco  []string       `cbor:"slco,omitempty"`
		Mo    map[int]string `cbor:"mo,omitempty"`
		Po    *int           `cbor:"po,omitempty"`
		Intfo any            `cbor:"intfo"`
	}
	type T struct {
		Stro T1 `cbor:"stro,omitempty"`
	}

	v := T{}
	want := []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x65, 0x69, 0x6e, 0x74, 0x66, 0x6f, 0xf6} // {"stro": {intfo: nil}}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"non-default values", v, want}}, em, dm)
}

func TestOmitEmptyMode(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyMode)
}

func testOmitEmptyMode(t *testing.T) {
	type T1 struct{}
	type T struct {
		B     bool           `cbor:"b"`
		Bo    bool           `cbor:"bo,omitempty"`
		UI    uint           `cbor:"ui"`
		UIo   uint           `cbor:"uio,omitempty"`
		I     int            `cbor:"i"`
		Io    int            `cbor:"io,omitempty"`
		F     float64        `cbor:"f"`
		Fo    float64        `cbor:"fo,omitempty"`
		S     string         `cbor:"s"`
		So    string         `cbor:"so,omitempty"`
		Slc   []string       `cbor:"slc"`
		Slco  []string       `cbor:"slco,omitempty"`
		M     map[int]string `cbor:"m"`
		Mo    map[int]string `cbor:"mo,omitempty"`
		P     *int           `cbor:"p"`
		Po    *int           `cbor:"po,omitempty"`
		Intf  any            `cbor:"intf"`
		Intfo any            `cbor:"intfo,omitempty"`
		Str   T1             `cbor:"str"`
		Stro  T1             `cbor:"stro,omitempty"`
	}

	v := T{}

	// {"b": false, "ui": 0, "i": 0, "f": 0.0, "s": "", "slc": null, "m": null, "p": null, "intf": null, "str": {}, "stro": {}}
	wantDataWithOmitEmptyGoValue := []byte{
		0xab,
		0x61, 0x62, 0xf4,
		0x62, 0x75, 0x69, 0x00,
		0x61, 0x69, 0x00,
		0x61, 0x66, 0xfb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x61, 0x73, 0x60,
		0x63, 0x73, 0x6c, 0x63, 0xf6,
		0x61, 0x6d, 0xf6,
		0x61, 0x70, 0xf6,
		0x64, 0x69, 0x6e, 0x74, 0x66, 0xf6,
		0x63, 0x73, 0x74, 0x72, 0xa0,
		0x64, 0x73, 0x74, 0x72, 0x6F, 0xa0,
	}

	// {"b": false, "ui": 0, "i": 0, "f": 0.0, "s": "", "slc": null, "m": null, "p": null, "intf": null, "str": {}}
	wantDataWithOmitEmptyCBORValue := []byte{
		0xaa,
		0x61, 0x62, 0xf4,
		0x62, 0x75, 0x69, 0x00,
		0x61, 0x69, 0x00,
		0x61, 0x66, 0xfb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x61, 0x73, 0x60,
		0x63, 0x73, 0x6c, 0x63, 0xf6,
		0x61, 0x6d, 0xf6,
		0x61, 0x70, 0xf6,
		0x64, 0x69, 0x6e, 0x74, 0x66, 0xf6,
		0x63, 0x73, 0x74, 0x72, 0xa0,
	}

	emOmitEmptyGoValue, _ := EncOptions{OmitEmpty: OmitEmptyGoValue}.EncMode()
	emOmitEmptyCBORValue, _ := EncOptions{OmitEmpty: OmitEmptyCBORValue}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"OmitEmptyGoValue (default)", v, wantDataWithOmitEmptyGoValue}}, emOmitEmptyGoValue, dm)
	testRoundTrip(t, []roundTripTestCase{{"OmitEmptyCBORValue", v, wantDataWithOmitEmptyCBORValue}}, emOmitEmptyCBORValue, dm)
}

func TestOmitEmptyForNestedStruct(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForNestedStruct)
}

func testOmitEmptyForNestedStruct(t *testing.T) {
	type T1 struct {
		Bo    bool           `cbor:"bo,omitempty"`
		UIo   uint           `cbor:"uio,omitempty"`
		Io    int            `cbor:"io,omitempty"`
		Fo    float64        `cbor:"fo,omitempty"`
		So    string         `cbor:"so,omitempty"`
		Slco  []string       `cbor:"slco,omitempty"`
		Mo    map[int]string `cbor:"mo,omitempty"`
		Po    *int           `cbor:"po,omitempty"`
		Intfo any            `cbor:"intfo,omitempty"`
	}
	type T2 struct {
		Stro T1 `cbor:"stro,omitempty"`
	}
	type T struct {
		Str  T2 `cbor:"str"`
		Stro T2 `cbor:"stro,omitempty"`
	}

	v := T{}
	want := []byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0} // {"str": {}}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"default values", v, want}}, em, dm)
}

func TestOmitEmptyForToArrayStruct1(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForToArrayStruct1)
}

func testOmitEmptyForToArrayStruct1(t *testing.T) {
	type T1 struct {
		_    struct{} `cbor:",toarray"`
		b    bool
		ui   uint
		i    int
		f    float64
		s    string
		slc  []string
		m    map[int]string
		p    *int
		intf any
	}
	type T struct {
		Str  T1 `cbor:"str"`
		Stro T1 `cbor:"stro,omitempty"`
	}

	v := T{
		Str:  T1{b: false, ui: 0, i: 0, f: 0.0, s: "", slc: nil, m: nil, p: nil, intf: nil},
		Stro: T1{b: false, ui: 0, i: 0, f: 0.0, s: "", slc: nil, m: nil, p: nil, intf: nil},
	}
	want := []byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0x80} // {"str": []}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"no exportable fields", v, want}}, em, dm)
}

func TestOmitEmptyForToArrayStruct2(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForToArrayStruct2)
}

func testOmitEmptyForToArrayStruct2(t *testing.T) {
	type T1 struct {
		_     struct{}       `cbor:",toarray"`
		Bo    bool           `cbor:"bo"`
		UIo   uint           `cbor:"uio"`
		Io    int            `cbor:"io"`
		Fo    float64        `cbor:"fo"`
		So    string         `cbor:"so"`
		Slco  []string       `cbor:"slco"`
		Mo    map[int]string `cbor:"mo"`
		Po    *int           `cbor:"po"`
		Intfo any            `cbor:"intfo"`
	}
	type T struct {
		Stro T1 `cbor:"stro,omitempty"`
	}

	v := T{}
	// {"stro": [false, 0, 0, 0.0, "", [], {}, nil, nil]}
	want := []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0x89, 0xf4, 0x00, 0x00, 0xfb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x60, 0xf6, 0xf6, 0xf6, 0xf6}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"has exportable fields", v, want}}, em, dm)
}

func TestOmitEmptyForStructWithPtrToAnonymousField(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForStructWithPtrToAnonymousField)
}

func testOmitEmptyForStructWithPtrToAnonymousField(t *testing.T) {
	type (
		T1 struct {
			X int `cbor:"x,omitempty"`
			Y int `cbor:"y,omitempty"`
		}
		T2 struct {
			*T1
		}
		T struct {
			Stro T2 `cbor:"stro,omitempty"`
		}
	)

	testCases := []struct {
		name         string
		obj          any
		wantCborData []byte
	}{
		{
			name:         "null pointer to anonymous field",
			obj:          T{},
			wantCborData: []byte{0xa0}, // {}
		},
		{
			name:         "not-null pointer to anonymous field",
			obj:          T{T2{&T1{}}},
			wantCborData: []byte{0xa0}, // {}
		},
		{
			name:         "not empty value in field 1",
			obj:          T{T2{&T1{X: 1}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x61, 0x78, 0x01}, // {stro:{x:1}}
		},
		{
			name:         "not empty value in field 2",
			obj:          T{T2{&T1{Y: 2}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x61, 0x79, 0x02}, // {stro:{y:2}}
		},
		{
			name:         "not empty value in all fields",
			obj:          T{T2{&T1{X: 1, Y: 2}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa2, 0x61, 0x78, 0x01, 0x61, 0x79, 0x02}, // {stro:{x:1, y:2}}
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}
		})
	}
}

func TestOmitEmptyForStructWithAnonymousField(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForStructWithAnonymousField)
}

func testOmitEmptyForStructWithAnonymousField(t *testing.T) {
	type (
		T1 struct {
			X int `cbor:"x,omitempty"`
			Y int `cbor:"y,omitempty"`
		}
		T2 struct {
			T1
		}
		T struct {
			Stro T2 `cbor:"stro,omitempty"`
		}
	)

	testCases := []struct {
		name         string
		obj          any
		wantCborData []byte
	}{
		{
			name:         "default values",
			obj:          T{},
			wantCborData: []byte{0xa0}, // {}
		},
		{
			name:         "default values",
			obj:          T{T2{T1{}}},
			wantCborData: []byte{0xa0}, // {}
		},
		{
			name:         "not empty value in field 1",
			obj:          T{T2{T1{X: 1}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x61, 0x78, 0x01}, // {stro:{x:1}}
		},
		{
			name:         "not empty value in field 2",
			obj:          T{T2{T1{Y: 2}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x61, 0x79, 0x02}, // {stro:{y:2}}
		},
		{
			name:         "not empty value in all fields",
			obj:          T{T2{T1{X: 1, Y: 2}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa2, 0x61, 0x78, 0x01, 0x61, 0x79, 0x02}, // {stro:{x:1, y:2}}
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}
		})
	}
}

func TestOmitEmptyForBinaryMarshaler1(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForBinaryMarshaler1)
}

func testOmitEmptyForBinaryMarshaler1(t *testing.T) {
	type T1 struct {
		No number `cbor:"no,omitempty"`
	}
	type T struct {
		Str  T1 `cbor:"str"`
		Stro T1 `cbor:"stro,omitempty"`
	}

	testCases := []roundTripTestCase{
		{
			"empty BinaryMarshaler",
			T1{},
			[]byte{0xa0}, // {}
		},
		{
			"empty struct containing empty BinaryMarshaler",
			T{},
			[]byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0}, // {str: {}}
		},
	}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, testCases, em, dm)
}

func TestOmitEmptyForBinaryMarshaler2(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForBinaryMarshaler2)
}

func testOmitEmptyForBinaryMarshaler2(t *testing.T) {
	type T1 struct {
		So stru `cbor:"so,omitempty"`
	}
	type T struct {
		Str  T1 `cbor:"str"`
		Stro T1 `cbor:"stro,omitempty"`
	}

	testCases := []roundTripTestCase{
		{
			"empty BinaryMarshaler",
			T1{},
			[]byte{0xa0}, // {}
		},
		{
			"empty struct containing empty BinaryMarshaler",
			T{},
			[]byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0}, // {str: {}}
		},
	}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, testCases, em, dm)
}

// omitempty is a no-op for time.Time.
func TestOmitEmptyForTime(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForTime)
}

func testOmitEmptyForTime(t *testing.T) {
	type T struct {
		Tm time.Time `cbor:"t,omitempty"`
	}

	v := T{}
	want := []byte{0xa1, 0x61, 0x74, 0xf6} // {"t": nil}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"default values", v, want}}, em, dm)
}

// omitempty is a no-op for big.Int.
func TestOmitEmptyForBigInt(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForBigInt)
}

func testOmitEmptyForBigInt(t *testing.T) {
	type T struct {
		I big.Int `cbor:"bi,omitempty"`
	}

	v := T{}
	want := []byte{0xa1, 0x62, 0x62, 0x69, 0xc2, 0x40} // {"bi": 2([])}

	em, _ := EncOptions{BigIntConvert: BigIntConvertNone}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"default values", v, want}}, em, dm)
}

func TestOmitZeroForBuiltinType(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForBuiltinType)
}

func testOmitZeroForBuiltinType(t *testing.T) {
	type T struct {
		B     bool           `cbor:"b"`
		Bo    bool           `cbor:"bo,omitzero"`
		UI    uint           `cbor:"ui"`
		UIo   uint           `cbor:"uio,omitzero"`
		I     int            `cbor:"i"`
		Io    int            `cbor:"io,omitzero"`
		F     float64        `cbor:"f"`
		Fo    float64        `cbor:"fo,omitzero"`
		S     string         `cbor:"s"`
		So    string         `cbor:"so,omitzero"`
		Slc   []string       `cbor:"slc"`
		Slco  []string       `cbor:"slco,omitzero"`
		M     map[int]string `cbor:"m"`
		Mo    map[int]string `cbor:"mo,omitzero"`
		P     *int           `cbor:"p"`
		Po    *int           `cbor:"po,omitzero"`
		Intf  any            `cbor:"intf"`
		Intfo any            `cbor:"intfo,omitzero"`
	}

	v := T{}
	// {"b": false, "ui": 0, "i":0, "f": 0, "s": "", "slc": null, "m": {}, "p": nil, "intf": nil }
	want := []byte{0xa9,
		0x61, 0x62, 0xf4,
		0x62, 0x75, 0x69, 0x00,
		0x61, 0x69, 0x00,
		0x61, 0x66, 0xfb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x61, 0x73, 0x60,
		0x63, 0x73, 0x6c, 0x63, 0xf6,
		0x61, 0x6d, 0xf6,
		0x61, 0x70, 0xf6,
		0x64, 0x69, 0x6e, 0x74, 0x66, 0xf6,
	}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"default values", v, want}}, em, dm)
}

func TestOmitZeroForAnonymousStruct(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForAnonymousStruct)
}

func testOmitZeroForAnonymousStruct(t *testing.T) {
	type T struct {
		Str  struct{} `cbor:"str"`
		Stro struct{} `cbor:"stro,omitzero"`
	}

	v := T{}
	want := []byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0} // {"str": {}}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"default values", v, want}}, em, dm)
}

func TestOmitZeroForStruct1(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForStruct1)
}

func testOmitZeroForStruct1(t *testing.T) {
	type T1 struct {
		Bo    bool           `cbor:"bo,omitzero"`
		UIo   uint           `cbor:"uio,omitzero"`
		Io    int            `cbor:"io,omitzero"`
		Fo    float64        `cbor:"fo,omitzero"`
		So    string         `cbor:"so,omitzero"`
		Slco  []string       `cbor:"slco,omitzero"`
		Mo    map[int]string `cbor:"mo,omitzero"`
		Po    *int           `cbor:"po,omitzero"`
		Intfo any            `cbor:"intfo,omitzero"`
	}
	type T struct {
		Str  T1 `cbor:"str"`
		Stro T1 `cbor:"stro,omitzero"`
	}

	v := T{}
	want := []byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0} // {"str": {}}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"default values", v, want}}, em, dm)
}

func TestOmitZeroForStruct2(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForStruct2)
}

func testOmitZeroForStruct2(t *testing.T) {
	type T1 struct {
		Bo    bool           `cbor:"bo,omitzero"`
		UIo   uint           `cbor:"uio,omitzero"`
		Io    int            `cbor:"io,omitzero"`
		Fo    float64        `cbor:"fo,omitzero"`
		So    string         `cbor:"so,omitzero"`
		Slco  []string       `cbor:"slco,omitzero"`
		Mo    map[int]string `cbor:"mo,omitzero"`
		Po    *int           `cbor:"po,omitzero"`
		Intfo any            `cbor:"intfo"`
	}
	type T struct {
		Stro T1 `cbor:"stro,omitzero"`
	}

	v := T{}
	want := []byte{0xa0} // {}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"non-default values", v, want}}, em, dm)
}

func TestOmitZeroForNestedStruct(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForNestedStruct)
}

func testOmitZeroForNestedStruct(t *testing.T) {
	type T1 struct {
		Bo    bool           `cbor:"bo,omitzero"`
		UIo   uint           `cbor:"uio,omitzero"`
		Io    int            `cbor:"io,omitzero"`
		Fo    float64        `cbor:"fo,omitzero"`
		So    string         `cbor:"so,omitzero"`
		Slco  []string       `cbor:"slco,omitzero"`
		Mo    map[int]string `cbor:"mo,omitzero"`
		Po    *int           `cbor:"po,omitzero"`
		Intfo any            `cbor:"intfo,omitzero"`
	}
	type T2 struct {
		Stro T1 `cbor:"stro,omitzero"`
	}
	type T struct {
		Str  T2 `cbor:"str"`
		Stro T2 `cbor:"stro,omitzero"`
	}

	v := T{}
	want := []byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0} // {"str": {}}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"default values", v, want}}, em, dm)
}

func TestOmitZeroForToArrayStruct1(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForToArrayStruct1)
}

func testOmitZeroForToArrayStruct1(t *testing.T) {
	type T1 struct {
		_    struct{} `cbor:",toarray"`
		b    bool
		ui   uint
		i    int
		f    float64
		s    string
		slc  []string
		m    map[int]string
		p    *int
		intf any
	}
	type T struct {
		Str  T1 `cbor:"str"`
		Stro T1 `cbor:"stro,omitzero"`
	}

	v := T{
		Str:  T1{b: false, ui: 0, i: 0, f: 0.0, s: "", slc: nil, m: nil, p: nil, intf: nil},
		Stro: T1{b: false, ui: 0, i: 0, f: 0.0, s: "", slc: nil, m: nil, p: nil, intf: nil},
	}
	want := []byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0x80} // {"str": []}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"no exportable fields", v, want}}, em, dm)
}

func TestOmitZeroForToArrayStruct2(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForToArrayStruct2)
}

func testOmitZeroForToArrayStruct2(t *testing.T) {
	type T1 struct {
		_     struct{}       `cbor:",toarray"`
		Bo    bool           `cbor:"bo"`
		UIo   uint           `cbor:"uio"`
		Io    int            `cbor:"io"`
		Fo    float64        `cbor:"fo"`
		So    string         `cbor:"so"`
		Slco  []string       `cbor:"slco"`
		Mo    map[int]string `cbor:"mo"`
		Po    *int           `cbor:"po"`
		Intfo any            `cbor:"intfo"`
	}
	type T struct {
		Stro T1 `cbor:"stro,omitzero"`
	}

	v := T{}
	// {"stro": [false, 0, 0, 0.0, "", [], {}, nil, nil]}
	want := []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0x89, 0xf4, 0x00, 0x00, 0xfb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x60, 0xf6, 0xf6, 0xf6, 0xf6}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"has exportable fields", v, want}}, em, dm)
}

func TestOmitZeroForStructWithPtrToAnonymousField(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForStructWithPtrToAnonymousField)
}

func testOmitZeroForStructWithPtrToAnonymousField(t *testing.T) {
	type (
		T1 struct {
			X int `cbor:"x,omitzero"`
			Y int `cbor:"y,omitzero"`
		}
		T2 struct {
			*T1
		}
		T struct {
			Stro T2 `cbor:"stro,omitzero"`
		}
	)

	testCases := []struct {
		name         string
		obj          any
		wantCborData []byte
	}{
		{
			name:         "null pointer to anonymous field",
			obj:          T{},
			wantCborData: []byte{0xa0}, // {}
		},
		{
			name:         "not-null pointer to anonymous field",
			obj:          T{T2{&T1{}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa0}, // {"stro":{}}
		},
		{
			name:         "not empty value in field 1",
			obj:          T{T2{&T1{X: 1}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x61, 0x78, 0x01}, // {stro:{x:1}}
		},
		{
			name:         "not empty value in field 2",
			obj:          T{T2{&T1{Y: 2}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x61, 0x79, 0x02}, // {stro:{y:2}}
		},
		{
			name:         "not empty value in all fields",
			obj:          T{T2{&T1{X: 1, Y: 2}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa2, 0x61, 0x78, 0x01, 0x61, 0x79, 0x02}, // {stro:{x:1, y:2}}
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}
		})
	}
}

func TestOmitZeroForStructWithAnonymousField(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForStructWithAnonymousField)
}

func testOmitZeroForStructWithAnonymousField(t *testing.T) {
	type (
		T1 struct {
			X int `cbor:"x,omitzero"`
			Y int `cbor:"y,omitzero"`
		}
		T2 struct {
			T1
		}
		T struct {
			Stro T2 `cbor:"stro,omitzero"`
		}
	)

	testCases := []struct {
		name         string
		obj          any
		wantCborData []byte
	}{
		{
			name:         "default values",
			obj:          T{},
			wantCborData: []byte{0xa0}, // {}
		},
		{
			name:         "default values",
			obj:          T{T2{T1{}}},
			wantCborData: []byte{0xa0}, // {}
		},
		{
			name:         "not empty value in field 1",
			obj:          T{T2{T1{X: 1}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x61, 0x78, 0x01}, // {stro:{x:1}}
		},
		{
			name:         "not empty value in field 2",
			obj:          T{T2{T1{Y: 2}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa1, 0x61, 0x79, 0x02}, // {stro:{y:2}}
		},
		{
			name:         "not empty value in all fields",
			obj:          T{T2{T1{X: 1, Y: 2}}},
			wantCborData: []byte{0xa1, 0x64, 0x73, 0x74, 0x72, 0x6f, 0xa2, 0x61, 0x78, 0x01, 0x61, 0x79, 0x02}, // {stro:{x:1, y:2}}
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}
		})
	}
}

func TestOmitZeroForBinaryMarshaler1(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForBinaryMarshaler1)
}

func testOmitZeroForBinaryMarshaler1(t *testing.T) {
	type T1 struct {
		No number `cbor:"no,omitzero"`
	}
	type T struct {
		Str  T1 `cbor:"str"`
		Stro T1 `cbor:"stro,omitzero"`
	}

	testCases := []roundTripTestCase{
		{
			"empty BinaryMarshaler",
			T1{},
			[]byte{0xa0}, // {}
		},
		{
			"empty struct containing empty BinaryMarshaler",
			T{},
			[]byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0}, // {str: {}}
		},
	}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, testCases, em, dm)
}

func TestOmitZeroForBinaryMarshaler2(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForBinaryMarshaler2)
}

func testOmitZeroForBinaryMarshaler2(t *testing.T) {
	type T1 struct {
		So stru `cbor:"so,omitzero"`
	}
	type T struct {
		Str  T1 `cbor:"str"`
		Stro T1 `cbor:"stro,omitzero"`
	}

	testCases := []roundTripTestCase{
		{
			"empty BinaryMarshaler",
			T1{},
			[]byte{0xa0}, // {}
		},
		{
			"empty struct containing empty BinaryMarshaler",
			T{},
			[]byte{0xa1, 0x63, 0x73, 0x74, 0x72, 0xa0}, // {str: {}}
		},
	}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, testCases, em, dm)
}

func TestOmitZeroForTime(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForTime)
}

func testOmitZeroForTime(t *testing.T) {
	type T struct {
		Tm time.Time `cbor:"t,omitzero"`
	}

	v := T{}
	want := []byte{0xa0} // {}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"default values", v, want}}, em, dm)
}

func TestOmitZeroForBigInt(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroForBigInt)
}

func testOmitZeroForBigInt(t *testing.T) {
	type T struct {
		I big.Int `cbor:"bi,omitzero"`
	}

	v := T{}
	want := []byte{0xa0} // {}

	em, _ := EncOptions{BigIntConvert: BigIntConvertNone}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, []roundTripTestCase{{"default values", v, want}}, em, dm)
}

func TestIsZero(t *testing.T) {
	var zeroStructZeroer isZeroer = zeroTestTypeCustom{value: 42}

	testCases := []struct {
		name string
		t    reflect.Type
		v    reflect.Value

		want       bool
		wantHasErr bool
	}{
		{
			name: "nil",
			t:    nil,
			v:    reflect.ValueOf(nil),
			want: true,
		},
		{
			name: "string-zero",
			t:    reflect.TypeFor[string](),
			v:    reflect.ValueOf(""),
			want: true,
		},

		{
			name: "string-nonzero",
			t:    reflect.TypeFor[string](),
			v:    reflect.ValueOf("a"),
			want: false,
		},
		{
			name: "int-zero",
			t:    reflect.TypeFor[int](),
			v:    reflect.ValueOf(0),
			want: true,
		},
		{
			name: "int-nonzero",
			t:    reflect.TypeFor[int](),
			v:    reflect.ValueOf(1),
			want: false,
		},

		{
			name: "bool-zero",
			t:    reflect.TypeFor[bool](),
			v:    reflect.ValueOf(false),
			want: true,
		},
		{
			name: "bool-nonzero",
			t:    reflect.TypeFor[bool](),
			v:    reflect.ValueOf(true),
			want: false,
		},

		{
			name: "slice-zero",
			t:    reflect.TypeFor[[]string](),
			v:    reflect.ValueOf([]string(nil)),
			want: true,
		},
		{
			name: "slice-nonzero",
			t:    reflect.TypeFor[[]string](),
			v:    reflect.ValueOf([]string{}),
			want: false,
		},

		{
			name: "map-zero",
			t:    reflect.TypeFor[map[string]string](),
			v:    reflect.ValueOf(map[string]string(nil)),
			want: true,
		},
		{
			name: "map-nonzero",
			t:    reflect.TypeFor[map[string]string](),
			v:    reflect.ValueOf(map[string]string{}),
			want: false,
		},

		{
			name: "struct-zero",
			t:    reflect.TypeFor[zeroTestType](),
			v:    reflect.ValueOf(zeroTestType{}),
			want: true,
		},
		{
			name: "struct-nonzero",
			t:    reflect.TypeFor[zeroTestType](),
			v:    reflect.ValueOf(zeroTestType{value: 42}),
			want: false,
		},

		{
			name: "pointer-zero",
			t:    reflect.TypeFor[*zeroTestType](),
			v:    reflect.ValueOf((*zeroTestType)(nil)),
			want: true,
		},
		{
			name: "pointer-nonzero",
			t:    reflect.TypeFor[*zeroTestType](),
			v:    reflect.ValueOf(&zeroTestType{}),
			want: false,
		},

		{
			name: "any-struct-zero",
			t:    reflect.TypeOf(any(nil)),
			v:    reflect.ValueOf(zeroTestType{}),
			want: true,
		},
		{
			name: "any-struct-nonzero",
			t:    reflect.TypeOf(any(nil)),
			v:    reflect.ValueOf(zeroTestType{value: 42}),
			want: false,
		},

		{
			name: "any-pointer-zero",
			t:    reflect.TypeOf(any(nil)),
			v:    reflect.ValueOf((*zeroTestType)(nil)),
			want: true,
		},
		{
			name: "any-pointer-nonzero",
			t:    reflect.TypeOf(any(nil)),
			v:    reflect.ValueOf(&zeroTestType{}),
			want: false,
		},

		{
			name: "custom-structreceiver-zero-structvalue",
			t:    reflect.TypeFor[zeroTestTypeCustom](),
			v:    reflect.ValueOf(zeroTestTypeCustom{value: 42}),
			want: true,
		},
		{
			name: "custom-structreceiver-nonzero-structvalue",
			t:    reflect.TypeFor[zeroTestTypeCustom](),
			v:    reflect.ValueOf(zeroTestTypeCustom{value: 1}),
			want: false,
		},
		{
			name: "custom-structreceiver-zero-pointervalue",
			t:    reflect.TypeFor[zeroTestTypeCustom](),
			v:    reflect.ValueOf(&zeroTestTypeCustom{value: 42}),
			want: true,
		},
		{
			name: "custom-structreceiver-nonzero-pointervalue",
			t:    reflect.TypeFor[zeroTestTypeCustom](),
			v:    reflect.ValueOf(&zeroTestTypeCustom{value: 1}),
			want: false,
		},
		{
			name: "custom-structreceiver-zero-addressablestructvalue",
			t:    reflect.TypeFor[zeroTestTypeCustom](),
			v:    reflect.ValueOf(&zeroTestTypeCustom{value: 42}).Elem(),
			want: true,
		},
		{
			name: "custom-structreceiver-nonzero-addressablestructvalue",
			t:    reflect.TypeFor[zeroTestTypeCustom](),
			v:    reflect.ValueOf(&zeroTestTypeCustom{value: 1}).Elem(),
			want: false,
		},

		{
			name: "custom-structreceiver-zero-pointervalue",
			t:    reflect.TypeFor[*zeroTestTypeCustom](),
			v:    reflect.ValueOf(&zeroTestTypeCustom{value: 42}),
			want: true,
		},
		{
			name: "custom-structreceiver-nonzero-pointervalue",
			t:    reflect.TypeFor[*zeroTestTypeCustom](),
			v:    reflect.ValueOf(&zeroTestTypeCustom{value: 1}),
			want: false,
		},
		{
			name: "custom-structreceiver-zero-nil-pointervalue",
			t:    reflect.TypeFor[*zeroTestTypeCustom](),
			v:    reflect.ValueOf((*zeroTestTypeCustom)(nil)),
			want: true,
		},

		{
			name: "custom-pointerreceiver-zero-structvalue",
			t:    reflect.TypeFor[zeroTestTypeCustomPointer](),
			v:    reflect.ValueOf(zeroTestTypeCustomPointer{value: 42}),
			want: true,
		},
		{
			name: "custom-pointerreceiver-nonzero-structvalue",
			t:    reflect.TypeFor[zeroTestTypeCustomPointer](),
			v:    reflect.ValueOf(zeroTestTypeCustomPointer{value: 1}),
			want: false,
		},

		{
			name: "custom-pointerreceiver-zero-pointervalue",
			t:    reflect.TypeFor[*zeroTestTypeCustomPointer](),
			v:    reflect.ValueOf(&zeroTestTypeCustomPointer{value: 42}),
			want: true,
		},
		{
			name: "custom-pointerreceiver-nonzero-pointervalue",
			t:    reflect.TypeFor[*zeroTestTypeCustomPointer](),
			v:    reflect.ValueOf(&zeroTestTypeCustomPointer{value: 1}),
			want: false,
		},

		{
			name: "custom-interface-nil-pointer",
			t:    isZeroerType,
			v:    reflect.ValueOf((*zeroTestTypeCustom)(nil)),
			want: true,
		},
		{
			name: "custom-interface-zero-structreceiver-pointer",
			t:    isZeroerType,
			v:    reflect.ValueOf(&zeroTestTypeCustom{value: 42}),
			want: true,
		},
		{
			name: "custom-interface-zero-structreceiver",
			t:    isZeroerType,
			v:    reflect.ValueOf(zeroStructZeroer),
			want: true,
		},
		{
			name: "custom-interface-nonzero-struct",
			t:    isZeroerType,
			v:    reflect.ValueOf(&zeroTestTypeCustom{value: 1}),
			want: false,
		},
		{
			name: "custom-interface-nil-pointerreceiver",
			t:    isZeroerType,
			v:    reflect.ValueOf((*zeroTestTypeCustomPointer)(nil)),
			want: true,
		},
		{
			name: "custom-interface-zero-pointerreceiver",
			t:    isZeroerType,
			v:    reflect.ValueOf(&zeroTestTypeCustomPointer{value: 42}),
			want: true,
		},
		{
			name: "custom-interface-nonzero-pointerreceiver",
			t:    isZeroerType,
			v:    reflect.ValueOf(&zeroTestTypeCustomPointer{value: 1}),
			want: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Log(string(debug.Stack()))
					t.Errorf("unexpected panic %v", err)
				}
			}()
			got, err := getIsZeroFunc(tc.t)(nil, tc.v)
			if tc.wantHasErr != (err != nil) {
				t.Errorf("getIsZeroFunc() returned err=%v, wantErr=%v", err, tc.wantHasErr)
			}
			if tc.want != got {
				t.Errorf("getIsZeroFunc() returned %v, want %v", got, tc.want)
			}
		})
	}
}

type zeroTestType struct {
	value int
}

type zeroTestTypeCustom struct {
	value int
}

func (z zeroTestTypeCustom) IsZero() bool {
	return z.value == 42
}

type zeroTestTypeCustomPointer struct {
	value int
}

func (z *zeroTestTypeCustomPointer) IsZero() bool {
	return z.value == 42
}

func TestJSONStdlibOmitZero(t *testing.T) {
	testMarshalWithFP(t, testJSONStdlibOmitZero)
}

func testJSONStdlibOmitZero(t *testing.T) {
	type CBOR struct {
		S string `cbor:"s,omitzero"`
	}
	type JSON struct {
		S string `json:"s,omitzero"`
	}

	testCases := []struct {
		name string
		obj  any
		want []byte
	}{
		{
			name: "cbor tag",
			obj:  CBOR{},
			want: []byte{0xa0}, // {}
		},
		{
			name: "json stag",
			obj:  JSON{},
			want: []byte{0xa0}, // {}
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			em, _ := EncOptions{}.EncMode()
			dm, _ := DecOptions{}.DecMode()
			testRoundTrip(t, []roundTripTestCase{{tc.name, tc.obj, tc.want}}, em, dm)
		})
	}
}

func TestTaggedField(t *testing.T) {
	testMarshalWithFP(t, testTaggedField)
}

func testTaggedField(t *testing.T) {
	// A field (T2.X) with a tag dominates untagged field.
	type (
		T1 struct {
			S string
		}
		T2 struct {
			X string `cbor:"S"`
		}
		T struct {
			T1
			T2
		}
	)
	v := T{T1{"T1"}, T2{"T2"}}
	want := []byte{0xa1, 0x61, 0x53, 0x62, 0x54, 0x32} // {"S":"T2"}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{T2: T2{"T2"}}
	if err = unmarshal(t, b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestDuplicatedFields(t *testing.T) {
	testMarshalWithFP(t, testDuplicatedFields)
}

func testDuplicatedFields(t *testing.T) {
	// Duplicate fields (T.T1.S, T.T2.S) are ignored.
	type (
		T1 struct {
			S string
		}
		T2 struct {
			S string
		}
		T3 struct {
			X string `cbor:"S"`
		}
		T4 struct {
			T1
			T3
		}
		T struct {
			T1
			T2
			T4 // Contains a tagged S field through T3; should not dominate.
		}
	)
	v := T{
		T1{"T1"},
		T2{"T2"},
		T4{
			T1{"nested T1"},
			T3{"nested T3"},
		},
	}
	want := []byte{0xa0} // {}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}
}

type TReader struct {
	X int
}

func (s TReader) Read(_ []byte) (n int, err error) {
	return 0, nil
}

func TestTaggedAnonymousField(t *testing.T) {
	testMarshalWithFP(t, testTaggedAnonymousField)
}

func testTaggedAnonymousField(t *testing.T) {
	// Anonymous field with a name given in its CBOR tag is treated as having that name, rather than being anonymous.
	type (
		T1 struct {
			X int
		}
		T struct {
			X  int
			T1 `cbor:"T1"`
		}
	)
	v := T{X: 1, T1: T1{X: 2}}
	want := []byte{0xa2, 0x61, 0x58, 0x01, 0x62, 0x54, 0x31, 0xa1, 0x61, 0x58, 0x02} // {X: 1, T1: {X:2}}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	unmarshalWant := T{X: 1, T1: T1{X: 2}}
	if err = unmarshal(t, b, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
	} else if !reflect.DeepEqual(v2, unmarshalWant) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", b, v2, v2, unmarshalWant, unmarshalWant)
	}
}

func TestAnonymousInterfaceField(t *testing.T) {
	testMarshalWithFP(t, testAnonymousInterfaceField)
}

func testAnonymousInterfaceField(t *testing.T) {
	// Anonymous field of interface type is treated the same as having that type as its name, rather than being anonymous.
	type (
		T struct {
			X int
			io.Reader
		}
	)
	v := T{X: 1, Reader: TReader{X: 2}}
	want := []byte{0xa2, 0x61, 0x58, 0x01, 0x66, 0x52, 0x65, 0x61, 0x64, 0x65, 0x72, 0xa1, 0x61, 0x58, 0x02} // {X: 1, Reader: {X:2}}
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", v, b, want)
	}

	var v2 T
	const wantErrorMsg = "cannot unmarshal map into Go struct field cbor.T.Reader of type io.Reader"
	if err = unmarshal(t, b, &v2); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want error (*UnmarshalTypeError)", b)
	} else {
		if typeError, ok := err.(*UnmarshalTypeError); !ok {
			t.Errorf("Unmarshal(0x%x) returned wrong type of error %T, want (*UnmarshalTypeError)", b, err)
		} else if !strings.Contains(typeError.Error(), wantErrorMsg) {
			t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", b, err.Error(), wantErrorMsg)
		}
	}
}

func TestEncodeInterface(t *testing.T) {
	testMarshalWithFP(t, testEncodeInterface)
}

func testEncodeInterface(t *testing.T) {
	var r io.Reader = TReader{X: 2}
	want := []byte{0xa1, 0x61, 0x58, 0x02} // {X:2}
	b, err := Marshal(r)
	if err != nil {
		t.Errorf("Marshal(%+v) returned error %v", r, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", r, b, want)
	}

	var v io.Reader
	const wantErrorMsg = "cannot unmarshal map into Go value of type io.Reader"
	if err = unmarshal(t, b, &v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want error (*UnmarshalTypeError)", b)
	} else {
		if typeError, ok := err.(*UnmarshalTypeError); !ok {
			t.Errorf("Unmarshal(0x%x) returned wrong type of error %T, want (*UnmarshalTypeError)", b, err)
		} else if !strings.Contains(typeError.Error(), wantErrorMsg) {
			t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", b, err.Error(), wantErrorMsg)
		}
	}
}

func TestMarshalStructTag1(t *testing.T) {
	testMarshalWithFP(t, testMarshalStructTag1)
}

func testMarshalStructTag1(t *testing.T) {
	type strc struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
	}
	v := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	want := mustHexDecode("a3616161416162614261636143") // {"a":"A", "b":"B", "c":"C"}
	if b, err := Marshal(v); err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = %v, want %v", v, b, want)
	}
}

func TestMarshalStructTag2(t *testing.T) {
	testMarshalWithFP(t, testMarshalStructTag2)
}

func testMarshalStructTag2(t *testing.T) {
	type strc struct {
		A string `json:"a"`
		B string `json:"b"`
		C string `json:"c"`
	}
	v := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	want := mustHexDecode("a3616161416162614261636143") // {"a":"A", "b":"B", "c":"C"}
	if b, err := Marshal(v); err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = %v, want %v", v, b, want)
	}
}

func TestMarshalStructTag3(t *testing.T) {
	testMarshalWithFP(t, testMarshalStructTag3)
}

func testMarshalStructTag3(t *testing.T) {
	type strc struct {
		A string `json:"x" cbor:"a"`
		B string `json:"y" cbor:"b"`
		C string `json:"z"`
	}
	v := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	want := mustHexDecode("a36161614161626142617a6143") // {"a":"A", "b":"B", "z":"C"}
	if b, err := Marshal(v); err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = %v, want %v", v, b, want)
	}
}

func TestMarshalStructTag4(t *testing.T) {
	testMarshalWithFP(t, testMarshalStructTag4)
}

func testMarshalStructTag4(t *testing.T) {
	type strc struct {
		A string `json:"x" cbor:"a"`
		B string `json:"y" cbor:"b"`
		C string `json:"-"`
	}
	v := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	want := mustHexDecode("a26161614161626142") // {"a":"A", "b":"B"}
	if b, err := Marshal(v); err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = %v, want %v", v, b, want)
	}
}

func TestMarshalStructLongFieldName(t *testing.T) {
	testMarshalWithFP(t, testMarshalStructLongFieldName)
}

func testMarshalStructLongFieldName(t *testing.T) {
	type strc struct {
		A string `cbor:"a"`
		B string `cbor:"abcdefghijklmnopqrstuvwxyz"`
		C string `cbor:"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmn"`
	}
	v := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	want := mustHexDecode("a361616141781a6162636465666768696a6b6c6d6e6f707172737475767778797a614278426162636465666768696a6b6c6d6e6f707172737475767778797a6162636465666768696a6b6c6d6e6f707172737475767778797a6162636465666768696a6b6c6d6e6143") // {"a":"A", "abcdefghijklmnopqrstuvwxyz":"B", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmn":"C"}
	if b, err := Marshal(v); err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%+v) = %v, want %v", v, b, want)
	}
}

func TestMarshalRawMessageValue(t *testing.T) {
	testMarshalWithFP(t, testMarshalRawMessageValue)
}

func testMarshalRawMessageValue(t *testing.T) {
	type (
		T1 struct {
			M RawMessage `cbor:",omitempty"`
		}
		T2 struct {
			M *RawMessage `cbor:",omitempty"`
		}
	)

	var (
		rawNil   = RawMessage(nil)
		rawEmpty = RawMessage([]byte{})
		raw      = RawMessage([]byte{0x01})
	)

	testCases := []struct {
		obj  any
		want []byte
	}{
		// Test with nil RawMessage.
		{
			obj:  rawNil,
			want: []byte{0xf6},
		},
		{
			obj:  &rawNil,
			want: []byte{0xf6},
		},
		{
			obj:  []any{rawNil},
			want: []byte{0x81, 0xf6},
		},
		{
			obj:  &[]any{rawNil},
			want: []byte{0x81, 0xf6},
		},
		{
			obj:  []any{&rawNil},
			want: []byte{0x81, 0xf6},
		},
		{
			obj:  &[]any{&rawNil},
			want: []byte{0x81, 0xf6},
		},
		{
			obj:  struct{ M RawMessage }{rawNil},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  &struct{ M RawMessage }{rawNil},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  struct{ M *RawMessage }{&rawNil},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  &struct{ M *RawMessage }{&rawNil},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  map[string]any{"M": rawNil},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  &map[string]any{"M": rawNil},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  map[string]any{"M": &rawNil},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  &map[string]any{"M": &rawNil},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  T1{rawNil},
			want: []byte{0xa0},
		},
		{
			obj:  T2{&rawNil},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  &T1{rawNil},
			want: []byte{0xa0},
		},
		{
			obj:  &T2{&rawNil},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},

		// Test with empty, but non-nil, RawMessage.
		{
			obj:  rawEmpty,
			want: []byte{0xf6},
		},
		{
			obj:  &rawEmpty,
			want: []byte{0xf6},
		},
		{
			obj:  []any{rawEmpty},
			want: []byte{0x81, 0xf6},
		},
		{
			obj:  &[]any{rawEmpty},
			want: []byte{0x81, 0xf6},
		},
		{
			obj:  []any{&rawEmpty},
			want: []byte{0x81, 0xf6},
		},
		{
			obj:  &[]any{&rawEmpty},
			want: []byte{0x81, 0xf6},
		},
		{
			obj:  struct{ M RawMessage }{rawEmpty},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  &struct{ M RawMessage }{rawEmpty},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  struct{ M *RawMessage }{&rawEmpty},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  &struct{ M *RawMessage }{&rawEmpty},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  map[string]any{"M": rawEmpty},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  &map[string]any{"M": rawEmpty},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  map[string]any{"M": &rawEmpty},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  &map[string]any{"M": &rawEmpty},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  T1{rawEmpty},
			want: []byte{0xa0},
		},
		{
			obj:  T2{&rawEmpty},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},
		{
			obj:  &T1{rawEmpty},
			want: []byte{0xa0},
		},
		{
			obj:  &T2{&rawEmpty},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6},
		},

		// Test with RawMessage with some data.
		{
			obj:  raw,
			want: []byte{0x01},
		},
		{
			obj:  &raw,
			want: []byte{0x01},
		},
		{
			obj:  []any{raw},
			want: []byte{0x81, 0x01},
		},
		{
			obj:  &[]any{raw},
			want: []byte{0x81, 0x01},
		},
		{
			obj:  []any{&raw},
			want: []byte{0x81, 0x01},
		},
		{
			obj:  &[]any{&raw},
			want: []byte{0x81, 0x01},
		},
		{
			obj:  struct{ M RawMessage }{raw},
			want: []byte{0xa1, 0x61, 0x4d, 0x01},
		},
		{
			obj:  &struct{ M RawMessage }{raw},
			want: []byte{0xa1, 0x61, 0x4d, 0x01},
		},
		{
			obj:  struct{ M *RawMessage }{&raw},
			want: []byte{0xa1, 0x61, 0x4d, 0x01},
		},
		{
			obj:  &struct{ M *RawMessage }{&raw},
			want: []byte{0xa1, 0x61, 0x4d, 0x01},
		},
		{
			obj:  map[string]any{"M": raw},
			want: []byte{0xa1, 0x61, 0x4d, 0x01},
		},
		{
			obj:  &map[string]any{"M": raw},
			want: []byte{0xa1, 0x61, 0x4d, 0x01},
		},
		{
			obj:  map[string]any{"M": &raw},
			want: []byte{0xa1, 0x61, 0x4d, 0x01},
		},
		{
			obj:  &map[string]any{"M": &raw},
			want: []byte{0xa1, 0x61, 0x4d, 0x01},
		},
		{
			obj:  T1{raw},
			want: []byte{0xa1, 0x61, 0x4d, 0x01},
		},
		{
			obj:  T2{&raw},
			want: []byte{0xa1, 0x61, 0x4d, 0x01},
		},
		{
			obj:  &T1{raw},
			want: []byte{0xa1, 0x61, 0x4d, 0x01},
		},
		{
			obj:  &T2{&raw},
			want: []byte{0xa1, 0x61, 0x4d, 0x01},
		},
	}

	for _, tc := range testCases {
		b, err := Marshal(tc.obj)
		if err != nil {
			t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
		}
		if !bytes.Equal(b, tc.want) {
			t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.want)
		}
	}
}

func TestCyclicDataStructure(t *testing.T) {
	testMarshalWithFP(t, testCyclicDataStructure)
}

func testCyclicDataStructure(t *testing.T) {
	type Node struct {
		V int   `cbor:"v"`
		N *Node `cbor:"n,omitempty"`
	}
	v := Node{1, &Node{2, &Node{3, nil}}}                                                                                  // linked list: 1, 2, 3
	wantCborData := []byte{0xa2, 0x61, 0x76, 0x01, 0x61, 0x6e, 0xa2, 0x61, 0x76, 0x02, 0x61, 0x6e, 0xa1, 0x61, 0x76, 0x03} // {v: 1, n: {v: 2, n: {v: 3}}}
	data, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%v) returned error %v", v, err)
	}
	if !bytes.Equal(wantCborData, data) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, data, wantCborData)
	}
	var v1 Node
	if err = unmarshal(t, data, &v1); err != nil {
		t.Fatalf("Unmarshal(0x%x) returned error %v", data, err)
	}
	if !reflect.DeepEqual(v, v1) {
		t.Errorf("Unmarshal(0x%x) returned %+v, want %+v", data, v1, v)
	}
}

func TestMarshalUnmarshalStructKeyAsInt(t *testing.T) {
	testMarshalWithFP(t, testMarshalUnmarshalStructKeyAsInt)
}

func testMarshalUnmarshalStructKeyAsInt(t *testing.T) {
	type T struct {
		F1 int `cbor:"1,omitempty,keyasint"`
		F2 int `cbor:"2,omitempty"`
		F3 int `cbor:"-3,omitempty,keyasint"`
	}
	testCases := []struct {
		name         string
		obj          any
		wantCborData []byte
	}{
		{
			name:         "zero value struct",
			obj:          T{},
			wantCborData: mustHexDecode("a0"), // {}
		},
		{
			name:         "initialized value struct",
			obj:          T{F1: 1, F2: 2, F3: 3},
			wantCborData: mustHexDecode("a301012203613202"), // {1: 1, -3: 3, "2": 2}
		},
	}
	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned error %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := em.Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}

			var v2 T
			if err := unmarshal(t, b, &v2); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
			}
			if !reflect.DeepEqual(tc.obj, v2) {
				t.Errorf("Unmarshal(0x%x) returned %+v, want %+v", b, v2, tc.obj)
			}
		})
	}
}

func TestMarshalStructKeyAsIntNumError(t *testing.T) {
	testMarshalWithFP(t, testMarshalStructKeyAsIntNumError)
}

func testMarshalStructKeyAsIntNumError(t *testing.T) {
	type T1 struct {
		F1 int `cbor:"2.0,keyasint"`
	}
	type T2 struct {
		F1 int `cbor:"-18446744073709551616,keyasint"`
	}
	type T3 struct {
		F1 int `cbor:"99999999999999999999,keyasint"`
	}
	testCases := []struct {
		name         string
		obj          any
		wantErrorMsg string
	}{
		{
			name:         "float as key",
			obj:          T1{},
			wantErrorMsg: "cbor: failed to parse field name \"2.0\" to int",
		},
		{
			name:         "int key < math.MinInt",
			obj:          T2{},
			wantErrorMsg: "cbor: failed to parse field name \"-18446744073709551616\" to int",
		},
		{
			name:         "int key > math.MaxInt",
			obj:          T3{},
			wantErrorMsg: "cbor: failed to parse field name \"99999999999999999999\" to int",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.obj)
			if err == nil {
				t.Errorf("Marshal(%+v) didn't return an error, want error %q", tc.obj, tc.wantErrorMsg)
			} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Marshal(%v) error %v, want %v", tc.obj, err.Error(), tc.wantErrorMsg)
			} else if b != nil {
				t.Errorf("Marshal(%v) = 0x%x, want nil", tc.obj, b)
			}
		})
	}
}

func TestMarshalUnmarshalStructToArray(t *testing.T) {
	testMarshalWithFP(t, testMarshalUnmarshalStructToArray)
}

func testMarshalUnmarshalStructToArray(t *testing.T) {
	type T1 struct {
		M int `cbor:",omitempty"`
	}
	type T2 struct {
		N int `cbor:",omitempty"`
		O int `cbor:",omitempty"`
	}
	type T struct {
		_   struct{} `cbor:",toarray"`
		A   int      `cbor:",omitempty"`
		B   T1       // nested struct
		T1           // embedded struct
		*T2          // embedded struct
	}
	testCases := []struct {
		name         string
		obj          T
		wantCborData []byte
	}{
		{
			name:         "zero value struct (test omitempty)",
			obj:          T{},
			wantCborData: mustHexDecode("8500a000f6f6"), // [0, {}, 0, nil, nil]
		},
		{
			name:         "initialized struct",
			obj:          T{A: 24, B: T1{M: 1}, T1: T1{M: 2}, T2: &T2{N: 3, O: 4}},
			wantCborData: mustHexDecode("851818a1614d01020304"), // [24, {M: 1}, 2, 3, 4]
		},
		{
			name:         "null pointer to embedded struct",
			obj:          T{A: 24, B: T1{M: 1}, T1: T1{M: 2}},
			wantCborData: mustHexDecode("851818a1614d0102f6f6"), // [24, {M: 1}, 2, nil, nil]
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}

			// SortMode should be ignored for struct to array encoding
			em, err := EncOptions{Sort: SortCanonical}.EncMode()
			if err != nil {
				t.Errorf("EncMode() returned error %v", err)
			}
			b, err = em.Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}

			var v2 T
			if err := unmarshal(t, b, &v2); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
			}
			if tc.obj.T2 == nil {
				if !reflect.DeepEqual(*(v2.T2), T2{}) {
					t.Errorf("Unmarshal(0x%x) returned %+v, want %+v", b, v2, tc.obj)
				}
				v2.T2 = nil
			}
			if !reflect.DeepEqual(tc.obj, v2) {
				t.Errorf("Unmarshal(0x%x) returned %+v, want %+v", b, v2, tc.obj)
			}
		})
	}
}

func TestMapSort(t *testing.T) {
	testMarshalWithFP(t, testMapSort)
}

func testMapSort(t *testing.T) {
	m := make(map[any]any)
	m[10] = true
	m[100] = true
	m[-1] = true
	m["z"] = "zzz"
	m["aa"] = true
	m[[1]int{100}] = true
	m[[1]int{-1}] = true
	m[false] = true

	lenFirstSortedCborData := mustHexDecode("a80af520f5f4f51864f5617a637a7a7a8120f5626161f5811864f5") // sorted keys: 10, -1, false, 100, "z", [-1], "aa", [100]
	bytewiseSortedCborData := mustHexDecode("a80af51864f520f5617a637a7a7a626161f5811864f58120f5f4f5") // sorted keys: 10, 100, -1, "z", "aa", [100], [-1], false

	testCases := []struct {
		name         string
		opts         EncOptions
		wantCborData []byte
	}{
		{
			name:         "length first sort",
			opts:         EncOptions{Sort: SortLengthFirst},
			wantCborData: lenFirstSortedCborData,
		},
		{
			name:         "bytewise sort",
			opts:         EncOptions{Sort: SortBytewiseLexical},
			wantCborData: bytewiseSortedCborData,
		},
		{
			name:         "CBOR canonical sort",
			opts:         EncOptions{Sort: SortCanonical},
			wantCborData: lenFirstSortedCborData,
		},
		{
			name:         "CTAP2 canonical sort",
			opts:         EncOptions{Sort: SortCTAP2},
			wantCborData: bytewiseSortedCborData,
		},
		{
			name:         "core deterministic sort",
			opts:         EncOptions{Sort: SortCoreDeterministic},
			wantCborData: bytewiseSortedCborData,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Errorf("EncMode() returned error %v", err)
			}
			b, err := em.Marshal(m)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", m, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", m, b, tc.wantCborData)
			}
		})
	}
}

func TestStructSort(t *testing.T) {
	testMarshalWithFP(t, testStructSort)
}

func testStructSort(t *testing.T) {
	type T struct {
		A bool `cbor:"aa"`
		B bool `cbor:"z"`
		C bool `cbor:"-1,keyasint"`
		D bool `cbor:"100,keyasint"`
		E bool `cbor:"10,keyasint"`
	}
	var v T

	unsortedCborData := mustHexDecode("a5626161f4617af420f41864f40af4")       // unsorted fields: "aa", "z", -1, 100, 10
	lenFirstSortedCborData := mustHexDecode("a50af420f41864f4617af4626161f4") // sorted fields: 10, -1, 100, "z", "aa",
	bytewiseSortedCborData := mustHexDecode("a50af41864f420f4617af4626161f4") // sorted fields: 10, 100, -1, "z", "aa"

	testCases := []struct {
		name         string
		opts         EncOptions
		wantCborData []byte
	}{
		{
			name:         "no sort",
			opts:         EncOptions{},
			wantCborData: unsortedCborData,
		},
		{
			name:         "no sort",
			opts:         EncOptions{Sort: SortNone},
			wantCborData: unsortedCborData,
		},
		{
			name:         "length first sort",
			opts:         EncOptions{Sort: SortLengthFirst},
			wantCborData: lenFirstSortedCborData,
		},
		{
			name:         "bytewise sort",
			opts:         EncOptions{Sort: SortBytewiseLexical},
			wantCborData: bytewiseSortedCborData,
		},
		{
			name:         "CBOR canonical sort",
			opts:         EncOptions{Sort: SortCanonical},
			wantCborData: lenFirstSortedCborData,
		},
		{
			name:         "CTAP2 canonical sort",
			opts:         EncOptions{Sort: SortCTAP2},
			wantCborData: bytewiseSortedCborData,
		},
		{
			name:         "core deterministic sort",
			opts:         EncOptions{Sort: SortCoreDeterministic},
			wantCborData: bytewiseSortedCborData,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Errorf("EncMode() returned error %v", err)
			}
			b, err := em.Marshal(v)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", v, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, tc.wantCborData)
			}
		})
	}
}

func TestTypeAlias(t *testing.T) {
	testMarshalWithFP(t, testTypeAlias)
}

func testTypeAlias(t *testing.T) {
	type myBool = bool
	type myUint = uint
	type myUint8 = uint8
	type myUint16 = uint16
	type myUint32 = uint32
	type myUint64 = uint64
	type myInt = int
	type myInt8 = int8
	type myInt16 = int16
	type myInt32 = int32
	type myInt64 = int64
	type myFloat32 = float32
	type myFloat64 = float64
	type myString = string
	type myByteSlice = []byte
	type myIntSlice = []int
	type myIntArray = [4]int
	type myMapIntInt = map[int]int

	testCases := []roundTripTestCase{
		{
			name:         "bool alias",
			obj:          myBool(true),
			wantCborData: mustHexDecode("f5"),
		},
		{
			name:         "uint alias",
			obj:          myUint(0),
			wantCborData: mustHexDecode("00"),
		},
		{
			name:         "uint8 alias",
			obj:          myUint8(0),
			wantCborData: mustHexDecode("00"),
		},
		{
			name:         "uint16 alias",
			obj:          myUint16(1000),
			wantCborData: mustHexDecode("1903e8"),
		},
		{
			name:         "uint32 alias",
			obj:          myUint32(1000000),
			wantCborData: mustHexDecode("1a000f4240"),
		},
		{
			name:         "uint64 alias",
			obj:          myUint64(1000000000000),
			wantCborData: mustHexDecode("1b000000e8d4a51000"),
		},
		{
			name:         "int alias",
			obj:          myInt(-1),
			wantCborData: mustHexDecode("20"),
		},
		{
			name:         "int8 alias",
			obj:          myInt8(-1),
			wantCborData: mustHexDecode("20"),
		},
		{
			name:         "int16 alias",
			obj:          myInt16(-1000),
			wantCborData: mustHexDecode("3903e7"),
		},
		{
			name:         "int32 alias",
			obj:          myInt32(-1000),
			wantCborData: mustHexDecode("3903e7"),
		},
		{
			name:         "int64 alias",
			obj:          myInt64(-1000),
			wantCborData: mustHexDecode("3903e7"),
		},
		{
			name:         "float32 alias",
			obj:          myFloat32(100000.0),
			wantCborData: mustHexDecode("fa47c35000"),
		},
		{
			name:         "float64 alias",
			obj:          myFloat64(1.1),
			wantCborData: mustHexDecode("fb3ff199999999999a"),
		},
		{
			name:         "string alias",
			obj:          myString("a"),
			wantCborData: mustHexDecode("6161"),
		},
		{
			name:         "[]byte alias",
			obj:          myByteSlice([]byte{1, 2, 3, 4}), //nolint:unconvert
			wantCborData: mustHexDecode("4401020304"),
		},
		{
			name:         "[]int alias",
			obj:          myIntSlice([]int{1, 2, 3, 4}), //nolint:unconvert
			wantCborData: mustHexDecode("8401020304"),
		},
		{
			name:         "[4]int alias",
			obj:          myIntArray([...]int{1, 2, 3, 4}), //nolint:unconvert
			wantCborData: mustHexDecode("8401020304"),
		},
		{
			name:         "map[int]int alias",
			obj:          myMapIntInt(map[int]int{1: 2, 3: 4}), //nolint:unconvert
			wantCborData: mustHexDecode("a201020304"),
		},
	}
	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	dm, err := DecOptions{}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned an error %v", err)
	}
	testRoundTrip(t, testCases, em, dm)
}

func TestNewTypeWithBuiltinUnderlyingType(t *testing.T) {
	testMarshalWithFP(t, testNewTypeWithBuiltinUnderlyingType)
}

func testNewTypeWithBuiltinUnderlyingType(t *testing.T) { //nolint:dupl
	type myBool bool
	type myUint uint
	type myUint8 uint8
	type myUint16 uint16
	type myUint32 uint32
	type myUint64 uint64
	type myInt int
	type myInt8 int8
	type myInt16 int16
	type myInt32 int32
	type myInt64 int64
	type myFloat32 float32
	type myFloat64 float64
	type myString string
	type myByteSlice []byte
	type myIntSlice []int
	type myIntArray [4]int
	type myMapIntInt map[int]int

	testCases := []roundTripTestCase{
		{
			name:         "bool alias",
			obj:          myBool(true),
			wantCborData: mustHexDecode("f5"),
		},
		{
			name:         "uint alias",
			obj:          myUint(0),
			wantCborData: mustHexDecode("00"),
		},
		{
			name:         "uint8 alias",
			obj:          myUint8(0),
			wantCborData: mustHexDecode("00"),
		},
		{
			name:         "uint16 alias",
			obj:          myUint16(1000),
			wantCborData: mustHexDecode("1903e8"),
		},
		{
			name:         "uint32 alias",
			obj:          myUint32(1000000),
			wantCborData: mustHexDecode("1a000f4240"),
		},
		{
			name:         "uint64 alias",
			obj:          myUint64(1000000000000),
			wantCborData: mustHexDecode("1b000000e8d4a51000"),
		},
		{
			name:         "int alias",
			obj:          myInt(-1),
			wantCborData: mustHexDecode("20"),
		},
		{
			name:         "int8 alias",
			obj:          myInt8(-1),
			wantCborData: mustHexDecode("20"),
		},
		{
			name:         "int16 alias",
			obj:          myInt16(-1000),
			wantCborData: mustHexDecode("3903e7"),
		},
		{
			name:         "int32 alias",
			obj:          myInt32(-1000),
			wantCborData: mustHexDecode("3903e7"),
		},
		{
			name:         "int64 alias",
			obj:          myInt64(-1000),
			wantCborData: mustHexDecode("3903e7"),
		},
		{
			name:         "float32 alias",
			obj:          myFloat32(100000.0),
			wantCborData: mustHexDecode("fa47c35000"),
		},
		{
			name:         "float64 alias",
			obj:          myFloat64(1.1),
			wantCborData: mustHexDecode("fb3ff199999999999a"),
		},
		{
			name:         "string alias",
			obj:          myString("a"),
			wantCborData: mustHexDecode("6161"),
		},
		{
			name:         "[]byte alias",
			obj:          myByteSlice([]byte{1, 2, 3, 4}),
			wantCborData: mustHexDecode("4401020304"),
		},
		{
			name:         "[]int alias",
			obj:          myIntSlice([]int{1, 2, 3, 4}),
			wantCborData: mustHexDecode("8401020304"),
		},
		{
			name:         "[4]int alias",
			obj:          myIntArray([...]int{1, 2, 3, 4}),
			wantCborData: mustHexDecode("8401020304"),
		},
		{
			name:         "map[int]int alias",
			obj:          myMapIntInt(map[int]int{1: 2, 3: 4}),
			wantCborData: mustHexDecode("a201020304"),
		},
	}
	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	dm, err := DecOptions{}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned an error %v", err)
	}
	testRoundTrip(t, testCases, em, dm)
}

func TestShortestFloat16(t *testing.T) {
	testMarshalWithFP(t, testShortestFloat16)
}

func testShortestFloat16(t *testing.T) {
	testCases := []struct {
		name         string
		f64          float64
		wantCborData []byte
	}{
		// Data from RFC 7049 appendix A
		{
			name:         "shrink to float16",
			f64:          0.0,
			wantCborData: mustHexDecode("f90000"),
		},
		{
			name:         "shrink to float16",
			f64:          1.0,
			wantCborData: mustHexDecode("f93c00"),
		},
		{
			name:         "shrink to float16",
			f64:          1.5,
			wantCborData: mustHexDecode("f93e00"),
		},
		{
			name:         "shrink to float16",
			f64:          65504.0,
			wantCborData: mustHexDecode("f97bff"),
		},
		{
			name:         "shrink to float16",
			f64:          5.960464477539063e-08,
			wantCborData: mustHexDecode("f90001"),
		},
		{
			name:         "shrink to float16",
			f64:          6.103515625e-05,
			wantCborData: mustHexDecode("f90400"),
		},
		{
			name:         "shrink to float16",
			f64:          -4.0,
			wantCborData: mustHexDecode("f9c400"),
		},
		// Data from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
		{
			name:         "shrink to float16",
			f64:          0.333251953125,
			wantCborData: mustHexDecode("f93555"),
		},
		// Data from 7049bis 4.2.1 and 5.5
		{
			name:         "shrink to float16",
			f64:          5.5,
			wantCborData: mustHexDecode("f94580"),
		},
		// Data from RFC 7049 appendix A
		{
			name:         "shrink to float32",
			f64:          100000.0,
			wantCborData: mustHexDecode("fa47c35000"),
		},
		{
			name:         "shrink to float32",
			f64:          3.4028234663852886e+38,
			wantCborData: mustHexDecode("fa7f7fffff"),
		},
		// Data from 7049bis 4.2.1 and 5.5
		{
			name:         "shrink to float32",
			f64:          5555.5,
			wantCborData: mustHexDecode("fa45ad9c00"),
		},
		{
			name:         "shrink to float32",
			f64:          1000000.5,
			wantCborData: mustHexDecode("fa49742408"),
		},
		// Data from RFC 7049 appendix A
		{
			name:         "shrink to float64",
			f64:          1.0e+300,
			wantCborData: mustHexDecode("fb7e37e43c8800759c"),
		},
	}
	em, err := EncOptions{ShortestFloat: ShortestFloat16}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := em.Marshal(tc.f64)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.f64, err)
			} else if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.f64, b, tc.wantCborData)
			}
			var f64 float64
			if err = unmarshal(t, b, &f64); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
			} else if f64 != tc.f64 {
				t.Errorf("Unmarshal(0x%x) = %f, want %f", b, f64, tc.f64)
			}
		})
	}
}

func TestShortestFloatNone(t *testing.T) {
	testMarshalWithFP(t, testShortestFloatNone)
}

func testShortestFloatNone(t *testing.T) {
	testCases := []struct {
		name         string
		f            any
		wantCborData []byte
	}{
		// Data from RFC 7049 appendix A
		{
			name:         "float32",
			f:            float32(0.0),
			wantCborData: mustHexDecode("fa00000000"),
		},
		{
			name:         "float64",
			f:            float64(0.0),
			wantCborData: mustHexDecode("fb0000000000000000"),
		},
		{
			name:         "float32",
			f:            float32(1.0),
			wantCborData: mustHexDecode("fa3f800000"),
		},
		{
			name:         "float64",
			f:            float64(1.0),
			wantCborData: mustHexDecode("fb3ff0000000000000"),
		},
		{
			name:         "float32",
			f:            float32(1.5),
			wantCborData: mustHexDecode("fa3fc00000"),
		},
		{
			name:         "float64",
			f:            float64(1.5),
			wantCborData: mustHexDecode("fb3ff8000000000000"),
		},
		{
			name:         "float32",
			f:            float32(65504.0),
			wantCborData: mustHexDecode("fa477fe000"),
		},
		{
			name:         "float64",
			f:            float64(65504.0),
			wantCborData: mustHexDecode("fb40effc0000000000"),
		},
		{
			name:         "float32",
			f:            float32(5.960464477539063e-08),
			wantCborData: mustHexDecode("fa33800000"),
		},
		{
			name:         "float64",
			f:            float64(5.960464477539063e-08),
			wantCborData: mustHexDecode("fb3e70000000000000"),
		},
		{
			name:         "float32",
			f:            float32(6.103515625e-05),
			wantCborData: mustHexDecode("fa38800000"),
		},
		{
			name:         "float64",
			f:            float64(6.103515625e-05),
			wantCborData: mustHexDecode("fb3f10000000000000"),
		},
		{
			name:         "float32",
			f:            float32(-4.0),
			wantCborData: mustHexDecode("fac0800000"),
		},
		{
			name:         "float64",
			f:            float64(-4.0),
			wantCborData: mustHexDecode("fbc010000000000000"),
		},
		// Data from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
		{
			name:         "float32",
			f:            float32(0.333251953125),
			wantCborData: mustHexDecode("fa3eaaa000"),
		},
		{
			name:         "float64",
			f:            float64(0.333251953125),
			wantCborData: mustHexDecode("fb3fd5540000000000"),
		},
		// Data from 7049bis 4.2.1 and 5.5
		{
			name:         "float32",
			f:            float32(5.5),
			wantCborData: mustHexDecode("fa40b00000"),
		},
		{
			name:         "float64",
			f:            float64(5.5),
			wantCborData: mustHexDecode("fb4016000000000000"),
		},
		// Data from RFC 7049 appendix A
		{
			name:         "float32",
			f:            float32(100000.0),
			wantCborData: mustHexDecode("fa47c35000"),
		},
		{
			name:         "float64",
			f:            float64(100000.0),
			wantCborData: mustHexDecode("fb40f86a0000000000"),
		},
		{
			name:         "float32",
			f:            float32(3.4028234663852886e+38),
			wantCborData: mustHexDecode("fa7f7fffff"),
		},
		{
			name:         "float64",
			f:            float64(3.4028234663852886e+38),
			wantCborData: mustHexDecode("fb47efffffe0000000"),
		},
		// Data from 7049bis 4.2.1 and 5.5
		{
			name:         "float32",
			f:            float32(5555.5),
			wantCborData: mustHexDecode("fa45ad9c00"),
		},
		{
			name:         "float64",
			f:            float64(5555.5),
			wantCborData: mustHexDecode("fb40b5b38000000000"),
		},
		{
			name:         "float32",
			f:            float32(1000000.5),
			wantCborData: mustHexDecode("fa49742408"),
		},
		{
			name:         "float64",
			f:            float64(1000000.5),
			wantCborData: mustHexDecode("fb412e848100000000"),
		},
		{
			name:         "float64",
			f:            float64(1.0e+300),
			wantCborData: mustHexDecode("fb7e37e43c8800759c"),
		},
	}
	em, err := EncOptions{ShortestFloat: ShortestFloatNone}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := em.Marshal(tc.f)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.f, err)
			} else if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.f, b, tc.wantCborData)
			}
			if reflect.ValueOf(tc.f).Kind() == reflect.Float32 {
				var f32 float32
				if err = unmarshal(t, b, &f32); err != nil {
					t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
				} else if f32 != tc.f {
					t.Errorf("Unmarshal(0x%x) = %f, want %f", b, f32, tc.f)
				}
			} else {
				var f64 float64
				if err = unmarshal(t, b, &f64); err != nil {
					t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
				} else if f64 != tc.f {
					t.Errorf("Unmarshal(0x%x) = %f, want %f", b, f64, tc.f)
				}
			}
		})
	}
}

func TestInfConvert(t *testing.T) {
	testMarshalWithFP(t, testInfConvert)
}

func testInfConvert(t *testing.T) {
	infConvertNoneOpt := EncOptions{InfConvert: InfConvertNone}
	infConvertFloat16Opt := EncOptions{InfConvert: InfConvertFloat16}
	testCases := []struct {
		name         string
		v            any
		opts         EncOptions
		wantCborData []byte
	}{
		{
			name:         "float32 -inf no conversion",
			v:            float32(math.Inf(-1)),
			opts:         infConvertNoneOpt,
			wantCborData: mustHexDecode("faff800000"),
		},
		{
			name:         "float32 +inf no conversion",
			v:            float32(math.Inf(1)),
			opts:         infConvertNoneOpt,
			wantCborData: mustHexDecode("fa7f800000"),
		},
		{
			name:         "float64 -inf no conversion",
			v:            math.Inf(-1),
			opts:         infConvertNoneOpt,
			wantCborData: mustHexDecode("fbfff0000000000000"),
		},
		{
			name:         "float64 +inf no conversion",
			v:            math.Inf(1),
			opts:         infConvertNoneOpt,
			wantCborData: mustHexDecode("fb7ff0000000000000"),
		},
		{
			name:         "float32 -inf to float16",
			v:            float32(math.Inf(-1)),
			opts:         infConvertFloat16Opt,
			wantCborData: mustHexDecode("f9fc00"),
		},
		{
			name:         "float32 +inf to float16",
			v:            float32(math.Inf(1)),
			opts:         infConvertFloat16Opt,
			wantCborData: mustHexDecode("f97c00"),
		},
		{
			name:         "float64 -inf to float16",
			v:            math.Inf(-1),
			opts:         infConvertFloat16Opt,
			wantCborData: mustHexDecode("f9fc00"),
		},
		{
			name:         "float64 +inf to float16",
			v:            math.Inf(1),
			opts:         infConvertFloat16Opt,
			wantCborData: mustHexDecode("f97c00"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Fatalf("EncMode() returned an error %v", err)
			}
			b, err := em.Marshal(tc.v)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.v, err)
			} else if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.v, b, tc.wantCborData)
			}
		})
		var vName string
		switch v := tc.v.(type) {
		case float32:
			vName = fmt.Sprintf("0x%x", math.Float32bits(v))
		case float64:
			vName = fmt.Sprintf("0x%x", math.Float64bits(v))
		}
		t.Run("reject inf "+vName, func(t *testing.T) {
			em, err := EncOptions{InfConvert: InfConvertReject}.EncMode()
			if err != nil {
				t.Fatalf("EncMode() returned an error %v", err)
			}
			want := &UnsupportedValueError{msg: "floating-point infinity"}
			if _, got := em.Marshal(tc.v); !reflect.DeepEqual(want, got) {
				t.Errorf("Marshal(%v) returned error %v, want %v", tc.v, got, want)
			}
		})
	}
}

func TestNilContainers(t *testing.T) {
	testMarshalWithFP(t, testNilContainers)
}

func testNilContainers(t *testing.T) {
	nilContainersNull := EncOptions{NilContainers: NilContainerAsNull}
	nilContainersEmpty := EncOptions{NilContainers: NilContainerAsEmpty}

	testCases := []struct {
		name         string
		v            any
		opts         EncOptions
		wantCborData []byte
	}{
		{
			name:         "map(nil) as CBOR null",
			v:            map[string]string(nil),
			opts:         nilContainersNull,
			wantCborData: mustHexDecode("f6"),
		},
		{
			name:         "map(nil) as CBOR empty map",
			v:            map[string]string(nil),
			opts:         nilContainersEmpty,
			wantCborData: mustHexDecode("a0"),
		},
		{
			name:         "slice(nil) as CBOR null",
			v:            []int(nil),
			opts:         nilContainersNull,
			wantCborData: mustHexDecode("f6"),
		},
		{
			name:         "slice(nil) as CBOR empty array",
			v:            []int(nil),
			opts:         nilContainersEmpty,
			wantCborData: mustHexDecode("80"),
		},
		{
			name:         "[]byte(nil) as CBOR null",
			v:            []byte(nil),
			opts:         nilContainersNull,
			wantCborData: mustHexDecode("f6"),
		},
		{
			name:         "[]byte(nil) as CBOR empty bytestring",
			v:            []byte(nil),
			opts:         nilContainersEmpty,
			wantCborData: mustHexDecode("40"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Errorf("EncMode() returned an error %v", err)
			}
			b, err := em.Marshal(tc.v)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.v, err)
			} else if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.v, b, tc.wantCborData)
			}
		})
	}
}

// Keith Randall's workaround for constant propagation issue https://github.com/golang/go/issues/36400
const (
	// qnan 32 bits constants
	qnanConst0xffc00001 uint32 = 0xffc00001
	qnanConst0x7fc00001 uint32 = 0x7fc00001
	qnanConst0xffc02000 uint32 = 0xffc02000
	qnanConst0x7fc02000 uint32 = 0x7fc02000
	// snan 32 bits constants
	snanConst0xff800001 uint32 = 0xff800001
	snanConst0x7f800001 uint32 = 0x7f800001
	snanConst0xff802000 uint32 = 0xff802000
	snanConst0x7f802000 uint32 = 0x7f802000
	// qnan 64 bits constants
	qnanConst0xfff8000000000001 uint64 = 0xfff8000000000001
	qnanConst0x7ff8000000000001 uint64 = 0x7ff8000000000001
	qnanConst0xfff8000020000000 uint64 = 0xfff8000020000000
	qnanConst0x7ff8000020000000 uint64 = 0x7ff8000020000000
	qnanConst0xfffc000000000000 uint64 = 0xfffc000000000000
	qnanConst0x7ffc000000000000 uint64 = 0x7ffc000000000000
	// snan 64 bits constants
	snanConst0xfff0000000000001 uint64 = 0xfff0000000000001
	snanConst0x7ff0000000000001 uint64 = 0x7ff0000000000001
	snanConst0xfff0000020000000 uint64 = 0xfff0000020000000
	snanConst0x7ff0000020000000 uint64 = 0x7ff0000020000000
	snanConst0xfff4000000000000 uint64 = 0xfff4000000000000
	snanConst0x7ff4000000000000 uint64 = 0x7ff4000000000000
)

var (
	// qnan 32 bits variables
	qnanVar0xffc00001 = qnanConst0xffc00001
	qnanVar0x7fc00001 = qnanConst0x7fc00001
	qnanVar0xffc02000 = qnanConst0xffc02000
	qnanVar0x7fc02000 = qnanConst0x7fc02000
	// snan 32 bits variables
	snanVar0xff800001 = snanConst0xff800001
	snanVar0x7f800001 = snanConst0x7f800001
	snanVar0xff802000 = snanConst0xff802000
	snanVar0x7f802000 = snanConst0x7f802000
	// qnan 64 bits variables
	qnanVar0xfff8000000000001 = qnanConst0xfff8000000000001
	qnanVar0x7ff8000000000001 = qnanConst0x7ff8000000000001
	qnanVar0xfff8000020000000 = qnanConst0xfff8000020000000
	qnanVar0x7ff8000020000000 = qnanConst0x7ff8000020000000
	qnanVar0xfffc000000000000 = qnanConst0xfffc000000000000
	qnanVar0x7ffc000000000000 = qnanConst0x7ffc000000000000
	// snan 64 bits variables
	snanVar0xfff0000000000001 = snanConst0xfff0000000000001
	snanVar0x7ff0000000000001 = snanConst0x7ff0000000000001
	snanVar0xfff0000020000000 = snanConst0xfff0000020000000
	snanVar0x7ff0000020000000 = snanConst0x7ff0000020000000
	snanVar0xfff4000000000000 = snanConst0xfff4000000000000
	snanVar0x7ff4000000000000 = snanConst0x7ff4000000000000
)

func TestNaNConvert(t *testing.T) {
	testMarshalWithFP(t, testNaNConvert)
}

func testNaNConvert(t *testing.T) {
	nanConvert7e00Opt := EncOptions{NaNConvert: NaNConvert7e00}
	nanConvertNoneOpt := EncOptions{NaNConvert: NaNConvertNone}
	nanConvertPreserveSignalOpt := EncOptions{NaNConvert: NaNConvertPreserveSignal}
	nanConvertQuietOpt := EncOptions{NaNConvert: NaNConvertQuiet}

	type nanConvert struct {
		opt          EncOptions
		wantCborData []byte
	}
	testCases := []struct {
		v       any
		convert []nanConvert
	}{
		// float32 qNaN dropped payload not zero
		{
			v: math.Float32frombits(qnanVar0xffc00001),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("faffc00001"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("faffc00001"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("faffc00001"),
				},
			},
		},
		// float32 qNaN dropped payload not zero
		{
			v: math.Float32frombits(qnanVar0x7fc00001),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fa7fc00001"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("fa7fc00001"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("fa7fc00001"),
				},
			},
		},
		// float32 -qNaN dropped payload zero
		{
			v: math.Float32frombits(qnanVar0xffc02000),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("faffc02000"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("f9fe01"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("f9fe01"),
				},
			},
		},
		// float32 qNaN dropped payload zero
		{
			v: math.Float32frombits(qnanVar0x7fc02000),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fa7fc02000"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("f97e01"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("f97e01"),
				},
			},
		},
		// float32 -sNaN dropped payload not zero
		{
			v: math.Float32frombits(snanVar0xff800001),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("faff800001"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("faff800001"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("faffc00001"),
				},
			},
		},
		// float32 sNaN dropped payload not zero
		{
			v: math.Float32frombits(snanVar0x7f800001),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fa7f800001"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("fa7f800001"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("fa7fc00001"),
				},
			},
		},
		// float32 -sNaN dropped payload zero
		{
			v: math.Float32frombits(snanVar0xff802000),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("faff802000"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("f9fc01"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("f9fe01"),
				},
			},
		},
		// float32 sNaN dropped payload zero
		{
			v: math.Float32frombits(snanVar0x7f802000),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fa7f802000"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("f97c01"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("f97e01"),
				},
			},
		},
		// float64 -qNaN dropped payload not zero
		{
			v: math.Float64frombits(qnanVar0xfff8000000000001),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fbfff8000000000001"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("fbfff8000000000001"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("fbfff8000000000001"),
				},
			},
		},
		// float64 qNaN dropped payload not zero
		{
			v: math.Float64frombits(qnanVar0x7ff8000000000001),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fb7ff8000000000001"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("fb7ff8000000000001"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("fb7ff8000000000001"),
				},
			},
		},
		// float64 -qNaN dropped payload zero
		{
			v: math.Float64frombits(qnanVar0xfff8000020000000),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fbfff8000020000000"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("faffc00001"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("faffc00001"),
				},
			},
		},
		// float64 qNaN dropped payload zero
		{
			v: math.Float64frombits(qnanVar0x7ff8000020000000),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fb7ff8000020000000"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("fa7fc00001"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("fa7fc00001"),
				},
			},
		},
		// float64 -qNaN dropped payload zero
		{
			v: math.Float64frombits(qnanVar0xfffc000000000000),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fbfffc000000000000"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("f9ff00"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("f9ff00"),
				},
			},
		},
		// float64 qNaN dropped payload zero
		{
			v: math.Float64frombits(qnanVar0x7ffc000000000000),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fb7ffc000000000000"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("f97f00"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("f97f00"),
				},
			},
		},
		// float64 -sNaN dropped payload not zero
		{
			v: math.Float64frombits(snanVar0xfff0000000000001),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fbfff0000000000001"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("fbfff0000000000001"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("fbfff8000000000001"),
				},
			},
		},
		// float64 sNaN dropped payload not zero
		{
			v: math.Float64frombits(snanVar0x7ff0000000000001),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fb7ff0000000000001"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("fb7ff0000000000001"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("fb7ff8000000000001"),
				},
			},
		},
		// float64 -sNaN dropped payload zero
		{
			v: math.Float64frombits(snanVar0xfff0000020000000),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fbfff0000020000000"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("faff800001"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("faffc00001"),
				},
			},
		},
		// float64 sNaN dropped payload zero
		{
			v: math.Float64frombits(snanVar0x7ff0000020000000),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fb7ff0000020000000"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("fa7f800001"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("fa7fc00001"),
				},
			},
		},
		// float64 -sNaN dropped payload zero
		{
			v: math.Float64frombits(snanVar0xfff4000000000000),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fbfff4000000000000"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("f9fd00"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("f9ff00"),
				},
			},
		},
		// float64 sNaN dropped payload zero
		{
			v: math.Float64frombits(snanVar0x7ff4000000000000),
			convert: []nanConvert{
				{
					opt:          nanConvert7e00Opt,
					wantCborData: mustHexDecode("f97e00"),
				},
				{
					opt:          nanConvertNoneOpt,
					wantCborData: mustHexDecode("fb7ff4000000000000"),
				},
				{
					opt:          nanConvertPreserveSignalOpt,
					wantCborData: mustHexDecode("f97d00"),
				},
				{
					opt:          nanConvertQuietOpt,
					wantCborData: mustHexDecode("f97f00"),
				},
			},
		},
	}
	for _, tc := range testCases {
		var vName string
		switch v := tc.v.(type) {
		case float32:
			vName = fmt.Sprintf("0x%x", math.Float32bits(v))
		case float64:
			vName = fmt.Sprintf("0x%x", math.Float64bits(v))
		}
		for _, convert := range tc.convert {
			var convertName string
			switch convert.opt.NaNConvert {
			case NaNConvert7e00:
				convertName = "Convert7e00"
			case NaNConvertNone:
				convertName = "ConvertNone"
			case NaNConvertPreserveSignal:
				convertName = "ConvertPreserveSignal"
			case NaNConvertQuiet:
				convertName = "ConvertQuiet"
			}
			name := convertName + "_" + vName
			t.Run(name, func(t *testing.T) {
				em, err := convert.opt.EncMode()
				if err != nil {
					t.Fatalf("EncMode() returned an error %v", err)
				}
				b, err := em.Marshal(tc.v)
				if err != nil {
					t.Errorf("Marshal(%v) returned error %v", tc.v, err)
				} else if !bytes.Equal(b, convert.wantCborData) {
					t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.v, b, convert.wantCborData)
				}
			})
		}

		t.Run("ConvertReject_"+vName, func(t *testing.T) {
			em, err := EncOptions{NaNConvert: NaNConvertReject}.EncMode()
			if err != nil {
				t.Fatalf("EncMode() returned an error %v", err)
			}
			want := &UnsupportedValueError{msg: "floating-point NaN"}
			if _, got := em.Marshal(tc.v); !reflect.DeepEqual(want, got) {
				t.Errorf("Marshal(%v) returned error %v, want %v", tc.v, got, want)
			}
		})
	}
}

func TestMarshalSenML(t *testing.T) {
	testMarshalWithFP(t, testMarshalSenML)
}

func testMarshalSenML(t *testing.T) {
	// Data from https://tools.ietf.org/html/rfc8428#section-6
	// Data contains 13 floating-point numbers.
	data := mustHexDecode("87a721781b75726e3a6465763a6f773a3130653230373361303130383030363a22fb41d303a15b00106223614120050067766f6c7461676501615602fb405e066666666666a3006763757272656e74062402fb3ff3333333333333a3006763757272656e74062302fb3ff4cccccccccccda3006763757272656e74062202fb3ff6666666666666a3006763757272656e74062102f93e00a3006763757272656e74062002fb3ff999999999999aa3006763757272656e74060002fb3ffb333333333333")
	testCases := []struct {
		name string
		opts EncOptions
	}{
		{
			name: "EncOptions ShortestFloatNone",
			opts: EncOptions{},
		},
		{
			name: "EncOptions ShortestFloat16",
			opts: EncOptions{ShortestFloat: ShortestFloat16},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v []SenMLRecord
			if err := unmarshal(t, data, &v); err != nil {
				t.Errorf("Marshal() returned error %v", err)
			}
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Errorf("EncMode() returned an error %v", err)
			}
			b, err := em.Marshal(v)
			if err != nil {
				t.Errorf("Unmarshal() returned error %v ", err)
			}
			var v2 []SenMLRecord
			if err := unmarshal(t, b, &v2); err != nil {
				t.Errorf("Marshal() returned error %v", err)
			}
			if !reflect.DeepEqual(v, v2) {
				t.Errorf("SenML round-trip failed: v1 %+v, v2 %+v", v, v2)
			}
		})
	}
}

func TestEncIndefiniteLengthOption(t *testing.T) {
	// Default option allows indefinite-length items.
	t.Run("default option", func(t *testing.T) {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		if err := enc.StartIndefiniteByteString(); err != nil {
			t.Errorf("StartIndefiniteByteString() returned an error %v", err)
		}
		if err := enc.EndIndefinite(); err != nil {
			t.Errorf("EndIndefinite() returned an error %v", err)
		}

		if err := enc.StartIndefiniteTextString(); err != nil {
			t.Errorf("StartIndefiniteTextString() returned an error %v", err)
		}
		if err := enc.EndIndefinite(); err != nil {
			t.Errorf("EndIndefinite() returned an error %v", err)
		}

		if err := enc.StartIndefiniteArray(); err != nil {
			t.Errorf("StartIndefiniteArray() returned an error %v", err)
		}
		if err := enc.EndIndefinite(); err != nil {
			t.Errorf("EndIndefinite() returned an error %v", err)
		}

		if err := enc.StartIndefiniteMap(); err != nil {
			t.Errorf("StartIndefiniteMap() returned an error %v", err)
		}
		if err := enc.EndIndefinite(); err != nil {
			t.Errorf("EndIndefinite() returned an error %v", err)
		}
	})

	// StartIndefiniteXXX returns error when IndefLength = IndefLengthForbidden
	t.Run("forbid indefinite length", func(t *testing.T) {
		var buf bytes.Buffer
		em, _ := EncOptions{IndefLength: IndefLengthForbidden}.EncMode()
		enc := em.NewEncoder(&buf)
		wantErrorMsg := "cbor: indefinite-length byte string isn't allowed"
		if err := enc.StartIndefiniteByteString(); err == nil {
			t.Errorf("StartIndefiniteByteString() didn't return an error")
		} else if _, ok := err.(*IndefiniteLengthError); !ok {
			t.Errorf("StartIndefiniteByteString() error type %T, want *IndefiniteLengthError", err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("StartIndefiniteByteString() returned error %q, want %q", err.Error(), wantErrorMsg)
		}

		wantErrorMsg = "cbor: indefinite-length UTF-8 text string isn't allowed"
		if err := enc.StartIndefiniteTextString(); err == nil {
			t.Errorf("StartIndefiniteTextString() didn't return an error")
		} else if _, ok := err.(*IndefiniteLengthError); !ok {
			t.Errorf("StartIndefiniteTextString() error type %T, want *IndefiniteLengthError", err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("StartIndefiniteTextString() returned error %q, want %q", err.Error(), wantErrorMsg)
		}

		wantErrorMsg = "cbor: indefinite-length array isn't allowed"
		if err := enc.StartIndefiniteArray(); err == nil {
			t.Errorf("StartIndefiniteArray() didn't return an error")
		} else if _, ok := err.(*IndefiniteLengthError); !ok {
			t.Errorf("StartIndefiniteArray() error type %T, want *IndefiniteLengthError", err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("StartIndefiniteArray() returned error %q, want %q", err.Error(), wantErrorMsg)
		}

		wantErrorMsg = "cbor: indefinite-length map isn't allowed"
		if err := enc.StartIndefiniteMap(); err == nil {
			t.Errorf("StartIndefiniteMap() didn't return an error")
		} else if _, ok := err.(*IndefiniteLengthError); !ok {
			t.Errorf("StartIndefiniteMap() error type %T, want *IndefiniteLengthError", err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("StartIndefiniteMap() returned error %q, want %q", err.Error(), wantErrorMsg)
		}
	})
}

func TestEncTagsMdOption(t *testing.T) {
	testMarshalWithFP(t, testEncTagsMdOption)
}

func testEncTagsMdOption(t *testing.T) {
	// Default option allows encoding CBOR tags
	tag := Tag{123, "hello"}
	if _, err := Marshal(tag); err != nil {
		t.Errorf("Marshal() returned an error %v", err)
	}

	// Create EncMode with TimeTag = EncTagRequired and TagsForbidden option returns error
	wantErrorMsg := "cbor: cannot set TagsMd to TagsForbidden when TimeTag is EncTagRequired"
	_, err := EncOptions{TimeTag: EncTagRequired, TagsMd: TagsForbidden}.EncMode()
	if err == nil {
		t.Errorf("EncModeWithTags() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncModeWithTags() returned error %q, want %q", err.Error(), wantErrorMsg)
	}

	// Create EncMode with TagSet and TagsForbidden option returns error
	wantErrorMsg = "cbor: cannot create EncMode with TagSet when TagsMd is TagsForbidden"
	tags := NewTagSet()
	_, err = EncOptions{TagsMd: TagsForbidden}.EncModeWithTags(tags)
	if err == nil {
		t.Errorf("EncModeWithTags() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncModeWithTags() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
	_, err = EncOptions{TagsMd: TagsForbidden}.EncModeWithSharedTags(tags)
	if err == nil {
		t.Errorf("EncModeWithSharedTags() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("EncModeWithSharedTags() returned error %q, want %q", err.Error(), wantErrorMsg)
	}

	// Encoding Tag and TagsForbidden option returns error
	wantErrorMsg = "cbor: cannot encode cbor.Tag when TagsMd is TagsForbidden"
	em, _ := EncOptions{TagsMd: TagsForbidden}.EncMode()
	if _, err := em.Marshal(&tag); err == nil {
		t.Errorf("Marshal() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Marshal() returned error %q, want %q", err.Error(), wantErrorMsg)
	}

	// Encoding RawTag and TagsForbidden option returns error
	wantErrorMsg = "cbor: cannot encode cbor.RawTag when TagsMd is TagsForbidden"
	rawTag := RawTag{123, []byte{01}}
	if _, err := em.Marshal(&rawTag); err == nil {
		t.Errorf("Marshal() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Marshal() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestMarshalPosBigInt(t *testing.T) {
	testMarshalWithFP(t, testMarshalPosBigInt)
}

func testMarshalPosBigInt(t *testing.T) {
	testCases := []struct {
		name             string
		cborDataShortest []byte
		cborDataBigInt   []byte
		value            big.Int
	}{
		{
			name:             "fit uint8",
			cborDataShortest: mustHexDecode("00"),
			cborDataBigInt:   mustHexDecode("c240"),
			value:            mustBigInt("0"),
		},
		{
			name:             "fit uint16",
			cborDataShortest: mustHexDecode("1903e8"),
			cborDataBigInt:   mustHexDecode("c24203e8"),
			value:            mustBigInt("1000"),
		},
		{
			name:             "fit uint32",
			cborDataShortest: mustHexDecode("1a000f4240"),
			cborDataBigInt:   mustHexDecode("c2430f4240"),
			value:            mustBigInt("1000000"),
		},
		{
			name:             "fit uint64",
			cborDataShortest: mustHexDecode("1b000000e8d4a51000"),
			cborDataBigInt:   mustHexDecode("c245e8d4a51000"),
			value:            mustBigInt("1000000000000"),
		},
		{
			name:             "max uint64",
			cborDataShortest: mustHexDecode("1bffffffffffffffff"),
			cborDataBigInt:   mustHexDecode("c248ffffffffffffffff"),
			value:            mustBigInt("18446744073709551615"),
		},
		{
			name:             "overflow uint64",
			cborDataShortest: mustHexDecode("c249010000000000000000"),
			cborDataBigInt:   mustHexDecode("c249010000000000000000"),
			value:            mustBigInt("18446744073709551616"),
		},
	}

	dmShortest, err := EncOptions{}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	dmShortestWithTagsForbidden, err := EncOptions{TagsMd: TagsForbidden}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	dmBigInt, err := EncOptions{BigIntConvert: BigIntConvertNone}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}

	tagsForbiddenErrorMsg := "cbor: cannot encode big.Int outside CBOR integer range when TagsMd is TagsForbidden"

	type s struct{ B big.Int }
	wantStructPrefix := []byte{0xa1, 0x61, 0x42}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if b, err := dmShortest.Marshal(tc.value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.value, err)
			} else if !bytes.Equal(b, tc.cborDataShortest) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.cborDataShortest)
			}

			if b, err := dmBigInt.Marshal(tc.value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.value, err)
			} else if !bytes.Equal(b, tc.cborDataBigInt) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.cborDataBigInt)
			}

			b, err := dmShortestWithTagsForbidden.Marshal(tc.value)
			if bytes.Equal(tc.cborDataShortest, tc.cborDataBigInt) {
				if err == nil {
					t.Errorf("Marshal(%v): expected error, got nil", tc.value)
				} else if err.Error() != tagsForbiddenErrorMsg {
					t.Errorf("Marshal(%v): got %q, want %q", tc.value, err.Error(), tagsForbiddenErrorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Marshal(%v) returned error %v", tc.value, err)
				} else if !bytes.Equal(b, tc.cborDataShortest) {
					t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.cborDataShortest)
				}
			}

			v := s{B: tc.value}
			want := append([]byte(nil), wantStructPrefix...)
			want = append(want, tc.cborDataBigInt...)
			if b, err := dmBigInt.Marshal(&v); err != nil {
				t.Errorf("Marshal(%v) returned error %v", v, err)
			} else if !bytes.Equal(b, want) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
			}
		})
	}
}

func TestMarshalNegBigInt(t *testing.T) {
	testMarshalWithFP(t, testMarshalNegBigInt)
}

func testMarshalNegBigInt(t *testing.T) {
	testCases := []struct {
		name             string
		cborDataShortest []byte
		cborDataBigInt   []byte
		value            big.Int
	}{
		{
			name:             "fit int8",
			cborDataShortest: mustHexDecode("20"),
			cborDataBigInt:   mustHexDecode("c340"),
			value:            mustBigInt("-1"),
		},
		{
			name:             "fit int16",
			cborDataShortest: mustHexDecode("3903e7"),
			cborDataBigInt:   mustHexDecode("c34203e7"),
			value:            mustBigInt("-1000"),
		},
		{
			name:             "fit int32",
			cborDataShortest: mustHexDecode("3a000f423f"),
			cborDataBigInt:   mustHexDecode("c3430f423f"),
			value:            mustBigInt("-1000000"),
		},
		{
			name:             "fit int64",
			cborDataShortest: mustHexDecode("3b000000e8d4a50fff"),
			cborDataBigInt:   mustHexDecode("c345e8d4a50fff"),
			value:            mustBigInt("-1000000000000"),
		},
		{
			name:             "min int64",
			cborDataShortest: mustHexDecode("3b7fffffffffffffff"),
			cborDataBigInt:   mustHexDecode("c3487fffffffffffffff"),
			value:            mustBigInt("-9223372036854775808"),
		},
		{
			name:             "overflow Go int64 fit CBOR neg int",
			cborDataShortest: mustHexDecode("3b8000000000000000"),
			cborDataBigInt:   mustHexDecode("c3488000000000000000"),
			value:            mustBigInt("-9223372036854775809"),
		},
		{
			name:             "min CBOR neg int",
			cborDataShortest: mustHexDecode("3bffffffffffffffff"),
			cborDataBigInt:   mustHexDecode("c348ffffffffffffffff"),
			value:            mustBigInt("-18446744073709551616"),
		},
		{
			name:             "overflow CBOR neg int",
			cborDataShortest: mustHexDecode("c349010000000000000000"),
			cborDataBigInt:   mustHexDecode("c349010000000000000000"),
			value:            mustBigInt("-18446744073709551617"),
		},
	}

	dmShortest, err := EncOptions{}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	dmShortestWithTagsForbidden, err := EncOptions{TagsMd: TagsForbidden}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	dmBigInt, err := EncOptions{BigIntConvert: BigIntConvertNone}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}

	tagsForbiddenErrorMsg := "cbor: cannot encode big.Int outside CBOR integer range when TagsMd is TagsForbidden"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if b, err := dmShortest.Marshal(tc.value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.value, err)
			} else if !bytes.Equal(b, tc.cborDataShortest) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.cborDataShortest)
			}

			if b, err := dmBigInt.Marshal(tc.value); err != nil {
				t.Errorf("Marshal(%v) returned error %v", tc.value, err)
			} else if !bytes.Equal(b, tc.cborDataBigInt) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.cborDataBigInt)
			}

			b, err := dmShortestWithTagsForbidden.Marshal(tc.value)
			if bytes.Equal(tc.cborDataShortest, tc.cborDataBigInt) {
				if err == nil {
					t.Errorf("Marshal(%v): expected error, got nil", tc.value)
				} else if err.Error() != tagsForbiddenErrorMsg {
					t.Errorf("Marshal(%v): got %q, want %q", tc.value, err.Error(), tagsForbiddenErrorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Marshal(%v) returned error %v", tc.value, err)
				} else if !bytes.Equal(b, tc.cborDataShortest) {
					t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.cborDataShortest)
				}
			}
		})
	}
}

func TestStructWithSimpleValueFields(t *testing.T) {
	testMarshalWithFP(t, testStructWithSimpleValueFields)
}

func testStructWithSimpleValueFields(t *testing.T) {
	type T struct {
		SV1 SimpleValue `cbor:",omitempty"` // omit empty
		SV2 SimpleValue
	}

	v1 := T{}
	want1 := []byte{0xa1, 0x63, 0x53, 0x56, 0x32, 0xe0}

	v2 := T{SV1: SimpleValue(1), SV2: SimpleValue(255)}
	want2 := []byte{
		0xa2,
		0x63, 0x53, 0x56, 0x31, 0xe1,
		0x63, 0x53, 0x56, 0x32, 0xf8, 0xff,
	}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testCases := []roundTripTestCase{
		{"default values", v1, want1},
		{"non-default values", v2, want2}}
	testRoundTrip(t, testCases, em, dm)
}

func TestMapWithSimpleValueKey(t *testing.T) {
	testMarshalWithFP(t, testMapWithSimpleValueKey)
}

func testMapWithSimpleValueKey(t *testing.T) {
	data := []byte{0xa2, 0x00, 0x00, 0xe0, 0x00} // {0: 0, simple(0): 0}

	// Decode CBOR map with positive integer 0 and simple value 0 as keys.
	// No map key duplication is detected because keys are of different CBOR types.
	// RFC 8949 Section 5.6.1 says "a simple value 2 is not equivalent to an integer 2".
	decOpts := DecOptions{
		DupMapKey: DupMapKeyEnforcedAPF, // duplicated key not allowed
	}
	decMode, _ := decOpts.DecMode()

	var v map[any]any
	err := unmarshalWithDM(t, decMode, data, &v)
	if err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", data, err)
	}

	// Encode decoded Go map.
	encOpts := EncOptions{
		Sort: SortBytewiseLexical,
	}
	encMode, _ := encOpts.EncMode()

	encodedData, err := encMode.Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v) returned error %v", v, err)
	}

	// Test roundtrip produces identical CBOR data.
	if !bytes.Equal(data, encodedData) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, encodedData, data)
	}
}

func TestMarshalStringType(t *testing.T) {
	testMarshalWithFP(t, testMarshalStringType)
}

func testMarshalStringType(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts EncOptions
		in   string
		want []byte
	}{
		{
			name: "to byte string",
			opts: EncOptions{String: StringToByteString},
			in:   "01234",
			want: mustHexDecode("453031323334"),
		},
		{
			name: "to text string",
			opts: EncOptions{String: StringToTextString},
			in:   "01234",
			want: mustHexDecode("653031323334"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Fatal(err)
			}

			got, err := em.Marshal(tc.in)
			if err != nil {
				t.Errorf("unexpected error from Marshal(%q): %v", tc.in, err)
			}

			if !bytes.Equal(got, tc.want) {
				t.Errorf("Marshal(%q): wanted %x, got %x", tc.in, tc.want, got)
			}
		})
	}
}

func TestMarshalFieldNameType(t *testing.T) {
	testMarshalWithFP(t, testMarshalFieldNameType)
}

func testMarshalFieldNameType(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts EncOptions
		in   any
		want []byte
	}{
		{
			name: "fixed-length to text string",
			opts: EncOptions{FieldName: FieldNameToTextString},
			in: struct {
				F1 int `cbor:"1,keyasint"`
				F2 int `cbor:"a"`
				F3 int `cbor:"-3,keyasint"`
			}{},
			want: mustHexDecode("a301006161002200"),
		},
		{
			name: "fixed-length to byte string",
			opts: EncOptions{FieldName: FieldNameToByteString},
			in: struct {
				F1 int `cbor:"1,keyasint"`
				F2 int `cbor:"a"`
				F3 int `cbor:"-3,keyasint"`
			}{},
			want: mustHexDecode("a301004161002200"),
		},
		{
			name: "variable-length to text string",
			opts: EncOptions{FieldName: FieldNameToTextString},
			in: struct {
				F1 int `cbor:"1,omitempty,keyasint"`
				F2 int `cbor:"a,omitempty"`
				F3 int `cbor:"-3,omitempty,keyasint"`
			}{F1: 7, F2: 7, F3: 7},
			want: mustHexDecode("a301076161072207"),
		},
		{
			name: "variable-length to byte string",
			opts: EncOptions{FieldName: FieldNameToByteString},
			in: struct {
				F1 int `cbor:"1,omitempty,keyasint"`
				F2 int `cbor:"a,omitempty"`
				F3 int `cbor:"-3,omitempty,keyasint"`
			}{F1: 7, F2: 7, F3: 7},
			want: mustHexDecode("a301074161072207"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Fatal(err)
			}

			got, err := em.Marshal(tc.in)
			if err != nil {
				t.Errorf("unexpected error from Marshal(%q): %v", tc.in, err)
			}

			if !bytes.Equal(got, tc.want) {
				t.Errorf("Marshal(%q): wanted %x, got %x", tc.in, tc.want, got)
			}
		})
	}
}

func TestMarshalRawMessageContainingMalformedCBORData(t *testing.T) {
	testMarshalWithFP(t, testMarshalRawMessageContainingMalformedCBORData)
}

func testMarshalRawMessageContainingMalformedCBORData(t *testing.T) {
	testCases := []struct {
		name         string
		value        any
		wantErrorMsg string
	}{
		// Nil RawMessage and empty RawMessage are encoded as CBOR nil.
		{
			name:         "truncated data",
			value:        RawMessage{0xa6},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.RawMessage: unexpected EOF",
		},
		{
			name:         "malformed data",
			value:        RawMessage{0x1f},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.RawMessage: cbor: invalid additional information 31 for type positive integer",
		},
		{
			name:         "extraneous data",
			value:        RawMessage{0x01, 0x01},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.RawMessage: cbor: 1 bytes of extraneous data starting at index 1",
		},
		{
			name:         "invalid builtin tag",
			value:        RawMessage{0xc0, 0x01},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.RawMessage: cbor: tag number 0 must be followed by text string, got positive integer",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.value)
			if err == nil {
				t.Errorf("Marshal(%v) didn't return an error, want error %q", tc.value, tc.wantErrorMsg)
			} else if _, ok := err.(*MarshalerError); !ok {
				t.Errorf("Marshal(%v) error type %T, want *MarshalerError", tc.value, err)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Marshal(%v) error %q, want %q", tc.value, err.Error(), tc.wantErrorMsg)
			}
			if b != nil {
				t.Errorf("Marshal(%v) = 0x%x, want nil", tc.value, b)
			}
		})
	}
}

type marshaler struct {
	data []byte
}

func (m marshaler) MarshalCBOR() (data []byte, err error) {
	return m.data, nil
}

func TestMarshalerReturnsMalformedCBORData(t *testing.T) {
	testMarshalWithFP(t, testMarshalerReturnsMalformedCBORData)
}

func testMarshalerReturnsMalformedCBORData(t *testing.T) {

	testCases := []struct {
		name         string
		value        any
		wantErrorMsg string
	}{
		{
			name:         "truncated data",
			value:        marshaler{data: []byte{0xa6}},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.marshaler: unexpected EOF",
		},
		{
			name:         "malformed data",
			value:        marshaler{data: []byte{0x1f}},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.marshaler: cbor: invalid additional information 31 for type positive integer",
		},
		{
			name:         "extraneous data",
			value:        marshaler{data: []byte{0x01, 0x01}},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.marshaler: cbor: 1 bytes of extraneous data starting at index 1",
		},
		{
			name:         "invalid builtin tag",
			value:        marshaler{data: []byte{0xc0, 0x01}},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.marshaler: cbor: tag number 0 must be followed by text string, got positive integer",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.value)
			if err == nil {
				t.Errorf("Marshal(%v) didn't return an error, want error %q", tc.value, tc.wantErrorMsg)
			} else if _, ok := err.(*MarshalerError); !ok {
				t.Errorf("Marshal(%v) error type %T, want *MarshalerError", tc.value, err)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Marshal(%v) error %q, want %q", tc.value, err.Error(), tc.wantErrorMsg)
			}
			if b != nil {
				t.Errorf("Marshal(%v) = 0x%x, want nil", tc.value, b)
			}
		})
	}
}

func TestMarshalerReturnsDisallowedCBORData(t *testing.T) {
	testMarshalWithFP(t, testMarshalerReturnsDisallowedCBORData)
}

func testMarshalerReturnsDisallowedCBORData(t *testing.T) {

	testCases := []struct {
		name         string
		encOpts      EncOptions
		value        any
		wantErrorMsg string
	}{
		{
			name:         "enc mode forbids indefinite-length, data has indefinite-length",
			encOpts:      EncOptions{IndefLength: IndefLengthForbidden},
			value:        marshaler{data: mustHexDecode("5f42010243030405ff")},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.marshaler: cbor: indefinite-length byte string isn't allowed",
		},
		{
			name:    "enc mode allows indefinite-length, data has indefinite-length",
			encOpts: EncOptions{IndefLength: IndefLengthAllowed},
			value:   marshaler{data: mustHexDecode("5f42010243030405ff")},
		},
		{
			name:         "enc mode forbids tags, data has tags",
			encOpts:      EncOptions{TagsMd: TagsForbidden},
			value:        marshaler{data: mustHexDecode("c074323031332d30332d32315432303a30343a30305a")},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.marshaler: cbor: CBOR tag isn't allowed",
		},
		{
			name:    "enc mode allows tags, data has tags",
			encOpts: EncOptions{TagsMd: TagsAllowed},
			value:   marshaler{data: mustHexDecode("c074323031332d30332d32315432303a30343a30305a")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.encOpts.EncMode()
			if err != nil {
				t.Fatal(err)
			}

			b, err := em.Marshal(tc.value)
			if tc.wantErrorMsg == "" {
				if err != nil {
					t.Errorf("Marshal(%v) returned error %q", tc.value, err)
				}
			} else {
				if err == nil {
					t.Errorf("Marshal(%v) didn't return an error, want error %q", tc.value, tc.wantErrorMsg)
				} else if _, ok := err.(*MarshalerError); !ok {
					t.Errorf("Marshal(%v) error type %T, want *MarshalerError", tc.value, err)
				} else if err.Error() != tc.wantErrorMsg {
					t.Errorf("Marshal(%v) error %q, want %q", tc.value, err.Error(), tc.wantErrorMsg)
				}
				if b != nil {
					t.Errorf("Marshal(%v) = 0x%x, want nil", tc.value, b)
				}
			}
		})
	}
}

func TestSortModeFastShuffle(t *testing.T) {
	testMarshalWithFP(t, testSortModeFastShuffle)
}

func testSortModeFastShuffle(t *testing.T) {
	em, err := EncOptions{Sort: SortFastShuffle}.EncMode()
	if err != nil {
		t.Fatal(err)
	}

	// These cases are based on the assumption that even a constant-time shuffle algorithm can
	// give an unbiased permutation of the keys when there are exactly 2 keys, so each trial
	// should succeed with probability 1/2.

	for _, tc := range []struct {
		name   string
		trials int
		in     any
	}{
		{
			name:   "fixed length struct",
			trials: 1024,
			in:     struct{ A, B int }{},
		},
		{
			name:   "variable length struct",
			trials: 1024,
			in: struct {
				A int
				B int `cbor:",omitempty"`
			}{B: 1},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first, err := em.Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			for i := 1; i <= tc.trials; i++ {
				next, err := em.Marshal(tc.in)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(first, next) {
					return
				}
			}

			t.Errorf("object encoded identically in %d consecutive trials using SortFastShuffle", tc.trials)
		})
	}
}

func TestSortModeFastShuffleEmptyStruct(t *testing.T) {
	testMarshalWithFP(t, testSortModeFastShuffleEmptyStruct)
}

func testSortModeFastShuffleEmptyStruct(t *testing.T) {
	em, err := EncOptions{Sort: SortFastShuffle}.EncMode()
	if err != nil {
		t.Fatal(err)
	}

	got, err := em.Marshal(struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{0xa0}; !bytes.Equal(got, want) {
		t.Errorf("got 0x%x, want 0x%x", got, want)
	}
}

func TestMarshalByteArrayMode(t *testing.T) {
	testMarshalWithFP(t, testMarshalByteArrayMode)
}

func testMarshalByteArrayMode(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts EncOptions
		in   any
		want []byte
	}{
		{
			name: "byte array treated as byte slice by default",
			opts: EncOptions{},
			in:   [1]byte{},
			want: []byte{0x41, 0x00},
		},
		{
			name: "byte array treated as byte slice with ByteArrayAsByteSlice",
			opts: EncOptions{ByteArray: ByteArrayToByteSlice},
			in:   [1]byte{},
			want: []byte{0x41, 0x00},
		},
		{
			name: "byte array treated as array of integers with ByteArrayToArray",
			opts: EncOptions{ByteArray: ByteArrayToArray},
			in:   [1]byte{},
			want: []byte{0x81, 0x00},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Fatal(err)
			}

			out, err := em.Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(out, tc.want) {
				t.Errorf("unexpected output, got 0x%x want 0x%x", out, tc.want)
			}
		})
	}
}

func TestMarshalByteSliceMode(t *testing.T) {
	testMarshalWithFP(t, testMarshalByteSliceMode)
}

func testMarshalByteSliceMode(t *testing.T) {
	type namedByteSlice []byte
	ts := NewTagSet()
	if err := ts.Add(TagOptions{EncTag: EncTagRequired}, reflect.TypeFor[namedByteSlice](), 0xcc); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		tags TagSet
		opts EncOptions
		in   any
		want []byte
	}{
		{
			name: "byte slice marshals to byte string by default",
			opts: EncOptions{},
			in:   []byte{0xbb},
			want: []byte{0x41, 0xbb},
		},
		{
			name: "byte slice marshals to byte string by with ByteSliceToByteString",
			opts: EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatNone},
			in:   []byte{0xbb},
			want: []byte{0x41, 0xbb},
		},
		{
			name: "byte slice marshaled to byte string enclosed in base64url expected encoding tag",
			opts: EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase64URL},
			in:   []byte{0xbb},
			want: []byte{0xd5, 0x41, 0xbb},
		},
		{
			name: "byte slice marshaled to byte string enclosed in base64 expected encoding tag",
			opts: EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase64},
			in:   []byte{0xbb},
			want: []byte{0xd6, 0x41, 0xbb},
		},
		{
			name: "byte slice marshaled to byte string enclosed in base16 expected encoding tag",
			opts: EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase16},
			in:   []byte{0xbb},
			want: []byte{0xd7, 0x41, 0xbb},
		},
		{
			name: "user-registered tag numbers are encoded with no expected encoding tag",
			tags: ts,
			opts: EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatNone},
			in:   namedByteSlice{0xbb},
			want: []byte{0xd8, 0xcc, 0x41, 0xbb},
		},
		{
			name: "user-registered tag numbers are encoded after base64url expected encoding tag",
			tags: ts,
			opts: EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase64URL},
			in:   namedByteSlice{0xbb},
			want: []byte{0xd5, 0xd8, 0xcc, 0x41, 0xbb},
		},
		{
			name: "user-registered tag numbers are encoded after base64 expected encoding tag",
			tags: ts,
			opts: EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase64},
			in:   namedByteSlice{0xbb},
			want: []byte{0xd6, 0xd8, 0xcc, 0x41, 0xbb},
		},
		{
			name: "user-registered tag numbers are encoded after base16 expected encoding tag",
			tags: ts,
			opts: EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase16},
			in:   namedByteSlice{0xbb},
			want: []byte{0xd7, 0xd8, 0xcc, 0x41, 0xbb},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var em EncMode
			if tc.tags != nil {
				var err error
				if em, err = tc.opts.EncModeWithTags(tc.tags); err != nil {
					t.Fatal(err)
				}
			} else {
				var err error
				if em, err = tc.opts.EncMode(); err != nil {
					t.Fatal(err)
				}
			}

			out, err := em.Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(out, tc.want) {
				t.Errorf("unexpected output, got 0x%x want 0x%x", out, tc.want)
			}
		})
	}
}

func TestBigIntConvertReject(t *testing.T) {
	testMarshalWithFP(t, testBigIntConvertReject)
}

func testBigIntConvertReject(t *testing.T) {
	em, err := EncOptions{BigIntConvert: BigIntConvertReject}.EncMode()
	if err != nil {
		t.Fatal(err)
	}

	want := &UnsupportedTypeError{Type: typeBigInt}

	if _, err := em.Marshal(big.Int{}); !reflect.DeepEqual(want, err) {
		t.Errorf("want: %v, got: %v", want, err)
	}

	if _, err := em.Marshal(&big.Int{}); !reflect.DeepEqual(want, err) {
		t.Errorf("want: %v, got: %v", want, err)
	}
}

type testBinaryMarshaler struct {
	StringField  string `cbor:"s,omitempty"`
	IntegerField int64  `cbor:"i,omitempty"`
}

func (testBinaryMarshaler) MarshalBinary() ([]byte, error) {
	return []byte("MarshalBinary"), nil
}

func TestBinaryMarshalerMode(t *testing.T) {
	testMarshalWithFP(t, testBinaryMarshalerMode)
}

func testBinaryMarshalerMode(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts EncOptions
		in   any
		want []byte
	}{
		{
			name: "struct implementing BinaryMarshaler is encoded as MarshalBinary's output in a byte string by default",
			opts: EncOptions{},
			in: testBinaryMarshaler{
				StringField:  "z",
				IntegerField: 3,
			},
			want: []byte("\x4dMarshalBinary"), // 'MarshalBinary'
		},
		{
			name: "struct implementing BinaryMarshaler is encoded as MarshalBinary's output in a byte string with BinaryMarshalerByteString",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerByteString},
			in: testBinaryMarshaler{
				StringField:  "z",
				IntegerField: 3,
			},
			want: []byte("\x4dMarshalBinary"), // 'MarshalBinary'
		},
		{
			name: "struct implementing BinaryMarshaler is encoded to map with BinaryMarshalerNone",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerNone},
			in: testBinaryMarshaler{
				StringField:  "z",
				IntegerField: 3,
			},
			want: mustHexDecode("a26173617a616903"), // {"s": "z", "i": 3}
		},
		{
			name: "struct implementing BinaryMarshaler is encoded to map with BinaryMarshalerNone",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerNone},
			in: testBinaryMarshaler{
				StringField:  "z",
				IntegerField: 3,
			},
			want: mustHexDecode("a26173617a616903"), // {"s": "z", "i": 3}
		},
		{
			name: "struct with a field implementing BinaryMarshaler with omitempty flag encoded with BinaryMarshalerByteString",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerByteString},
			in: struct {
				M testBinaryMarshaler `cbor:"m,omitempty"`
			}{
				M: testBinaryMarshaler{},
			},
			want: []byte{0xa1, 0x61, 0x6d, 0x4d, 'M', 'a', 'r', 's', 'h', 'a', 'l', 'B', 'i', 'n', 'a', 'r', 'y'},
		},
		{
			name: "struct with a field implementing BinaryMarshaler with omitempty flag encoded with BinaryMarshalerNone",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerNone},
			in: struct {
				M testBinaryMarshaler `cbor:"m,omitempty"`
			}{
				M: testBinaryMarshaler{},
			},
			want: mustHexDecode("a0"),
		},
		{
			name: "struct with a field implementing BinaryMarshaler with omitzero flag encoded with BinaryMarshalerByteString",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerByteString},
			in: struct {
				M testBinaryMarshaler `cbor:"m,omitzero"`
			}{
				M: testBinaryMarshaler{},
			},
			want: mustHexDecode("a0"),
		},
		{
			name: "struct with a field implementing BinaryMarshaler with omitzero flag encoded with BinaryMarshalerNone",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerNone},
			in: struct {
				M testBinaryMarshaler `cbor:"m,omitzero"`
			}{
				M: testBinaryMarshaler{},
			},
			want: mustHexDecode("a0"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Fatal(err)
			}

			got, err := em.Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(tc.want, got) {
				t.Errorf("unexpected output, want: 0x%x, got 0x%x", tc.want, got)
			}
		})
	}
}

type binaryMarshalerWithUnsupportedType map[chan bool]any

func (binaryMarshalerWithUnsupportedType) MarshalBinary() ([]byte, error) {
	return []byte("MarshalBinary"), nil
}

type binaryMarshalerUintptr uintptr

func (binaryMarshalerUintptr) MarshalBinary() ([]byte, error) {
	return []byte("MarshalBinary"), nil
}

func TestBinaryMarshalerModeWithUnsupportedType(t *testing.T) {
	testMarshalWithFP(t, testBinaryMarshalerModeWithUnsupportedType)
}

func testBinaryMarshalerModeWithUnsupportedType(t *testing.T) {
	for _, tc := range []struct {
		name      string
		opts      EncOptions
		in        any
		want      []byte
		wantError bool
	}{
		{
			name: "unsupported type implementing BinaryMarshaler is encoded as MarshalBinary's output in a byte string by default",
			opts: EncOptions{},
			in:   binaryMarshalerWithUnsupportedType{},
			want: []byte("\x4dMarshalBinary"), // 'MarshalBinary'
		},
		{
			name: "unsupported type implementing BinaryMarshaler is encoded as MarshalBinary's output in a byte string with BinaryMarshalerByteString",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerByteString},
			in:   binaryMarshalerWithUnsupportedType{},
			want: []byte("\x4dMarshalBinary"), // 'MarshalBinary'
		},
		{
			name:      "unsupported type implementing BinaryMarshaler can't be encoded with BinaryMarshalerNone",
			opts:      EncOptions{BinaryMarshaler: BinaryMarshalerNone},
			in:        binaryMarshalerWithUnsupportedType{},
			wantError: true,
		},
		{
			name: "struct with unsupported type implementing BinaryMarshaler can't be encoded with BinaryMarshalerNone",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerNone},
			in: struct {
				M binaryMarshalerWithUnsupportedType
			}{
				M: binaryMarshalerWithUnsupportedType{},
			},
			wantError: true,
		},
		{
			name: "struct with unsupported type implementing BinaryMarshaler with omitempty can't be encoded with BinaryMarshalerNone",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerNone},
			in: struct {
				M binaryMarshalerWithUnsupportedType `cbor:",omitempty"`
			}{
				M: binaryMarshalerWithUnsupportedType{},
			},
			wantError: true,
		},
		{
			name: "struct with unsupported type implementing BinaryMarshaler with omitzero can't be encoded with BinaryMarshalerNone",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerNone},
			in: struct {
				M binaryMarshalerWithUnsupportedType `cbor:",omitzero"`
			}{
				M: binaryMarshalerWithUnsupportedType{},
			},
			wantError: true,
		},
		{
			name: "struct with uintptr type implementing BinaryMarshaler with omitzero is encoded with BinaryMarshalerByteString",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerByteString},
			in: struct {
				B binaryMarshalerUintptr `cbor:"b,omitzero"`
			}{
				B: 1,
			},
			want: mustHexDecode("a161624d4d61727368616c42696e617279"), // {"b": 'MarshalBinary'}
		},
		{
			name: "struct with zero uintptr type implementing BinaryMarshaler with omitzero is omitted with BinaryMarshalerByteString",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerByteString},
			in: struct {
				B binaryMarshalerUintptr `cbor:"b,omitzero"`
			}{},
			want: mustHexDecode("a0"), // {}
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Fatalf("EncMode(): unexpected error: %v", err)
			}

			got, err := em.Marshal(tc.in)
			if tc.wantError {
				if err == nil {
					t.Fatalf("Marshal(%v): expected *UnsupportedTypeError, got nil", tc.in)
				} else if _, ok := err.(*UnsupportedTypeError); !ok {
					t.Errorf("Marshal(%v) returned %T, want *UnsupportedTypeError", tc.in, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Marshal(%v): unexpected error: %v", tc.in, err)
				} else if !bytes.Equal(tc.want, got) {
					t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.in, got, tc.want)
				}
			}
		})
	}
}

type testTextMarshaler struct {
	String string `cbor:"s"`
	Error  error  `cbor:"-"`
}

func (tm *testTextMarshaler) MarshalText() ([]byte, error) {
	return []byte(tm.String), tm.Error
}

func TestTextMarshalerMode(t *testing.T) {
	testMarshalWithFP(t, testTextMarshalerMode)
}

func testTextMarshalerMode(t *testing.T) {
	testTags := NewTagSet()
	if err := testTags.Add(TagOptions{EncTag: EncTagRequired}, reflect.TypeFor[testTextMarshaler](), 9999); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		opts EncOptions
		tags TagSet
		in   any
		want []byte
	}{
		{
			name: "struct implementing TextMarshaler is encoded to map by default",
			opts: EncOptions{},
			in:   testTextMarshaler{String: "z"},
			want: []byte{0xa1, 0x61, 's', 0x61, 'z'}, // {"s": "z"}
		},
		{
			name: "struct implementing TextMarshaler is encoded to map with TextMarshalerNone",
			opts: EncOptions{TextMarshaler: TextMarshalerNone},
			in:   testTextMarshaler{String: "z"},
			want: []byte{0xa1, 0x61, 's', 0x61, 'z'}, // {"s": "z"}
		},
		{
			name: "struct implementing TextMarshaler is encoded as MarshalText's output in a text string with TextMarshalerTextString",
			opts: EncOptions{TextMarshaler: TextMarshalerTextString},
			tags: testTags,
			in:   testTextMarshaler{String: "z"},
			want: []byte{0xd9, 0x27, 0x0f, 0x61, 'z'}, // 9999("z")
		},
		{
			name: "TextMarshaler struct field with omitempty is omitted if empty slice is returned using TextMarshalerTextString",
			opts: EncOptions{TextMarshaler: TextMarshalerTextString},
			in: struct {
				M testTextMarshaler `cbor:"m,omitempty"`
			}{
				M: testTextMarshaler{String: ""},
			},
			want: []byte{0xa0}, // {}
		},
		{
			name: "TextMarshaler struct field with omitempty is not omitted if empty slice is returned using TextMarshalerNone",
			opts: EncOptions{TextMarshaler: TextMarshalerNone},
			in: struct {
				M testTextMarshaler `cbor:"m,omitempty"`
			}{
				M: testTextMarshaler{String: ""},
			},
			want: []byte{0xa1, 0x61, 'm', 0xa1, 0x61, 's', 0x60}, // {"m": {"s": ""}}
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				enc EncMode
				err error
			)

			if tc.tags != nil {
				enc, err = tc.opts.EncModeWithTags(tc.tags)
			} else {
				enc, err = tc.opts.EncMode()
			}
			if err != nil {
				t.Fatal(err)
			}

			got, err := enc.Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(tc.want, got) {
				t.Errorf("unexpected output, want: 0x%x, got 0x%x", tc.want, got)
			}
		})
	}
}

type textMarshalerWithUnsupportedType map[chan bool]any

func (tm *textMarshalerWithUnsupportedType) MarshalText() ([]byte, error) {
	return []byte("MarshalText"), nil
}

func TestTextMarshalerModeWithUnsupportedType(t *testing.T) {
	testMarshalWithFP(t, testTextMarshalerModeWithUnsupportedType)
}

func testTextMarshalerModeWithUnsupportedType(t *testing.T) {
	for _, tc := range []struct {
		name      string
		opts      EncOptions
		in        any
		want      []byte
		wantError bool
	}{
		{
			name:      "unsupported type implementing TextMarshaler can't be encoded by default",
			opts:      EncOptions{},
			in:        textMarshalerWithUnsupportedType{},
			wantError: true,
		},
		{
			name:      "unsupported type implementing TextMarshaler can't be encoded with TextMarshalerNone",
			opts:      EncOptions{TextMarshaler: TextMarshalerNone},
			in:        textMarshalerWithUnsupportedType{},
			wantError: true,
		},
		{
			name: "unsupported type implementing TextMarshaler is encoded as MarshalText's output in a text string with TextMarshalerTextString",
			opts: EncOptions{TextMarshaler: TextMarshalerTextString},
			in:   textMarshalerWithUnsupportedType{},
			want: []byte("\x6bMarshalText"),
		},
		{
			name: "struct with unsupported type can't be encoded using TextMarshalerNone",
			opts: EncOptions{TextMarshaler: TextMarshalerNone},
			in: struct {
				M textMarshalerWithUnsupportedType
			}{
				M: textMarshalerWithUnsupportedType{},
			},
			wantError: true,
		},
		{
			name: "struct with unsupported type with omitempty can't be encoded using TextMarshalerNone",
			opts: EncOptions{TextMarshaler: TextMarshalerNone},
			in: struct {
				M textMarshalerWithUnsupportedType `cbor:"m,omitempty"`
			}{
				M: textMarshalerWithUnsupportedType{},
			},
			wantError: true,
		},
		{
			name: "struct with unsupported type with omitzero can't be encoded using TextMarshalerNone",
			opts: EncOptions{TextMarshaler: TextMarshalerNone},
			in: struct {
				M textMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{
				M: textMarshalerWithUnsupportedType{},
			},
			wantError: true,
		},
		{
			name: "struct with unsupported type with omitzero using TextMarshalerTextString",
			opts: EncOptions{TextMarshaler: TextMarshalerTextString},
			in: struct {
				M textMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{
				M: textMarshalerWithUnsupportedType(nil),
			},
			want: []byte{0xa0},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Fatalf("EncMode(): unexpected error: %v", err)
			}

			got, err := em.Marshal(tc.in)
			if tc.wantError {
				if err == nil {
					t.Fatalf("Marshal(%v): expected *UnsupportedTypeError, got nil", tc.in)
				} else if _, ok := err.(*UnsupportedTypeError); !ok {
					t.Errorf("Marshal(%v) returned %T, want *UnsupportedTypeError", tc.in, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Marshal(%v): unexpected error: %v", tc.in, err)
				} else if !bytes.Equal(tc.want, got) {
					t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.in, got, tc.want)
				}
			}
		})
	}
}

func TestTextMarshalerModeError(t *testing.T) {
	testMarshalWithFP(t, testTextMarshalerModeError)
}

func testTextMarshalerModeError(t *testing.T) {
	testTags := NewTagSet()
	if err := testTags.Add(TagOptions{EncTag: EncTagRequired}, reflect.TypeFor[testTextMarshaler](), 9999); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		opts EncOptions
		in   any
		want string
	}{
		{
			name: "non-nil error returned when MarshalText returns non-nil error",
			opts: EncOptions{TextMarshaler: TextMarshalerTextString},
			in:   testTextMarshaler{Error: errors.New("test")},
			want: "cbor: cannot marshal text for cbor.testTextMarshaler: test",
		},
		{
			name: "non-nil error returned when MarshalText returns non-nil error during struct field emptiness check",
			opts: EncOptions{TextMarshaler: TextMarshalerTextString},
			in: struct {
				M testTextMarshaler `cbor:"m,omitempty"`
			}{
				M: testTextMarshaler{Error: errors.New("test")},
			},
			want: "cbor: cannot marshal text for cbor.testTextMarshaler: test",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			enc, err := tc.opts.EncMode()
			if err != nil {
				t.Fatal(err)
			}

			_, err = enc.Marshal(tc.in)
			if err == nil {
				t.Fatal("expected non-nil error")
			}

			if got := err.Error(); got != tc.want {
				t.Errorf("want: %q, got: %q", tc.want, got)
			}
		})
	}
}

type stubJSONMarshaler struct {
	JSON  string
	Error error
}

func (m stubJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(m.JSON), m.Error
}

type stubJSONMarshalerPointerReceiver struct {
	JSON string
}

func (m *stubJSONMarshalerPointerReceiver) MarshalJSON() ([]byte, error) {
	return []byte(m.JSON), nil
}

type transcodeFunc func(io.Writer, io.Reader) error

func (f transcodeFunc) Transcode(w io.Writer, r io.Reader) error {
	return f(w, r)
}

type jsonMarshalerWithUnsupportedType map[chan bool]any

func (m jsonMarshalerWithUnsupportedType) MarshalJSON() ([]byte, error) {
	return []byte("MarshalJSON"), nil
}

func TestJSONMarshalerTranscoderNil(t *testing.T) {
	testMarshalWithFP(t, testJSONMarshalerTranscoderNil)
}

func testJSONMarshalerTranscoderNil(t *testing.T) {
	enc, err := EncOptions{}.EncMode()
	if err != nil {
		t.Fatal(err)
	}

	{
		// default encode behavior of underlying type
		got, err := enc.Marshal(&stubJSONMarshalerPointerReceiver{JSON: "z"})
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{0xa1, 0x64, 'J', 'S', 'O', 'N', 0x61, 'z'}
		if !bytes.Equal(got, want) {
			t.Errorf("want 0x%x, got 0x%x", want, got)
		}
	}

	{
		// default empty condition of underlying type
		got, err := enc.Marshal(struct {
			M stubJSONMarshalerPointerReceiver `cbor:"m,omitempty"`
		}{})
		if err != nil {
			t.Fatal(err)
		}

		want := []byte{0xa1, 0x61, 'm', 0xa1, 0x64, 'J', 'S', 'O', 'N', 0x60}
		if !bytes.Equal(got, want) {
			t.Errorf("want 0x%x, got 0x%x", want, got)
		}
	}

	{
		// default encode behavior of underlying unsupported type
		value := jsonMarshalerWithUnsupportedType{}
		_, err := enc.Marshal(value)
		if err == nil {
			t.Errorf("Marshal(%v): expected *UnsupportedTypeError, got nil", value)
		} else if _, ok := err.(*UnsupportedTypeError); !ok {
			t.Errorf("Marshal(%v) returned %T, want *UnsupportedTypeError", value, err)
		}
	}
	{
		// default encode behavior of struct with underlying unsupported type
		value := struct {
			M jsonMarshalerWithUnsupportedType
		}{}
		_, err := enc.Marshal(value)
		if err == nil {
			t.Errorf("Marshal(%v): expected *UnsupportedTypeError, got nil", value)
		} else if _, ok := err.(*UnsupportedTypeError); !ok {
			t.Errorf("Marshal(%v) returned %T, want *UnsupportedTypeError", value, err)
		}
	}
	{
		// default empty condition of underlying unsupported type
		value := struct {
			M jsonMarshalerWithUnsupportedType `cbor:",omitempty"`
		}{}
		_, err := enc.Marshal(value)
		if err == nil {
			t.Errorf("Marshal(%v): expected *UnsupportedTypeError, got nil", value)
		} else if _, ok := err.(*UnsupportedTypeError); !ok {
			t.Errorf("Marshal(%v) returned %T, want *UnsupportedTypeError", value, err)
		}
	}
	{
		// default zero condition of underlying unsupported type
		value := struct {
			M jsonMarshalerWithUnsupportedType `cbor:",omitzero"`
		}{}
		_, err := enc.Marshal(value)
		if err == nil {
			t.Errorf("Marshal(%v): expected *UnsupportedTypeError, got nil", value)
		} else if _, ok := err.(*UnsupportedTypeError); !ok {
			t.Errorf("Marshal(%v) returned %T, want *UnsupportedTypeError", value, err)
		}
	}
}

func TestJSONMarshalerTranscoder(t *testing.T) {
	testMarshalWithFP(t, testJSONMarshalerTranscoder)
}

func testJSONMarshalerTranscoder(t *testing.T) {
	testTags := NewTagSet()
	if err := testTags.Add(TagOptions{EncTag: EncTagRequired}, reflect.TypeFor[stubJSONMarshaler](), 9999); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name  string
		value any
		tags  TagSet

		transcodeInput  []byte
		transcodeOutput []byte
		transcodeError  error

		wantCborData []byte
		wantErrorMsg string
	}{
		{
			name:            "value-receiver marshaler",
			value:           stubJSONMarshaler{JSON: `"a"`},
			transcodeInput:  []byte(`"a"`),
			transcodeOutput: []byte{0x61, 'a'},
			wantCborData:    []byte{0x61, 'a'},
		},
		{
			name:           "transcoder returns non-nil error",
			value:          stubJSONMarshaler{JSON: `"a"`},
			transcodeInput: []byte(`"a"`),
			transcodeError: errors.New("test"),
			wantErrorMsg: TranscodeError{
				err:          errors.New("test"),
				rtype:        reflect.TypeFor[stubJSONMarshaler](),
				sourceFormat: "json",
				targetFormat: "cbor",
			}.Error(),
		},
		{
			name:            "transcoder produces invalid cbor",
			value:           stubJSONMarshaler{JSON: `"a"`},
			transcodeInput:  []byte(`"a"`),
			transcodeOutput: []byte{0xff},
			wantErrorMsg: TranscodeError{
				err:          errors.New(`cbor: unexpected "break" code`),
				rtype:        reflect.TypeFor[stubJSONMarshaler](),
				sourceFormat: "json",
				targetFormat: "cbor",
			}.Error(),
		},
		{
			name:            "transcoder produces short cbor",
			value:           stubJSONMarshaler{JSON: `"a"`},
			transcodeInput:  []byte(`"a"`),
			transcodeOutput: []byte{0x61},
			wantErrorMsg: TranscodeError{
				err:          io.ErrUnexpectedEOF,
				rtype:        reflect.TypeFor[stubJSONMarshaler](),
				sourceFormat: "json",
				targetFormat: "cbor",
			}.Error(),
		},
		{
			name:            "transcoder produces extraneous cbor",
			value:           stubJSONMarshaler{JSON: `"a"`},
			transcodeInput:  []byte(`"a"`),
			transcodeOutput: []byte{0x61, 'a', 0x61, 'b'},
			wantErrorMsg: TranscodeError{
				err:          &ExtraneousDataError{numOfBytes: 2, index: 2},
				rtype:        reflect.TypeFor[stubJSONMarshaler](),
				sourceFormat: "json",
				targetFormat: "cbor",
			}.Error(),
		},
		{
			name:         "marshaler returns non-nil error",
			value:        stubJSONMarshaler{Error: errors.New("test")},
			wantErrorMsg: "test",
		},
		{
			name:            "value-receiver marshaler with registered tag",
			tags:            testTags,
			value:           stubJSONMarshaler{JSON: `"a"`},
			transcodeInput:  []byte(`"a"`),
			transcodeOutput: []byte{0x61, 'a'},                   // "a"
			wantCborData:    []byte{0xd9, 0x27, 0x0f, 0x61, 'a'}, // 9999("a")
		},
		{
			name:            "pointer-receiver marshaler",
			value:           stubJSONMarshalerPointerReceiver{JSON: `"a"`},
			transcodeInput:  []byte(`"a"`),
			transcodeOutput: []byte{0x61, 'a'},
			wantCborData:    []byte{0x61, 'a'},
		},
		{
			name: "never omitempty",
			value: struct {
				M stubJSONMarshaler `cbor:"m,omitempty"`
			}{M: stubJSONMarshaler{JSON: `"a"`}},
			transcodeInput:  []byte(`"a"`),
			transcodeOutput: []byte{0x61, 'a'},
			wantCborData:    []byte{0xa1, 0x61, 'm', 0x61, 'a'},
		},
		{
			name: "omitzero with unsupported underlying type and non-zero value",
			value: struct {
				M jsonMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{M: jsonMarshalerWithUnsupportedType{}},
			transcodeInput:  []byte("MarshalJSON"),
			transcodeOutput: []byte{0x61, 'a'},
			wantCborData:    []byte{0xa1, 0x61, 'm', 0x61, 'a'}, // {"m": "a"}
		},
		{
			name: "omitzero with unsupported underlying type and zero value",
			value: struct {
				M jsonMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{},
			wantCborData: []byte{0xa0}, // {}
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := EncOptions{
				JSONMarshalerTranscoder: transcodeFunc(func(w io.Writer, r io.Reader) error {
					source, err := io.ReadAll(r)
					if err != nil {
						t.Fatal(err)
					}
					if got := string(source); got != string(tc.transcodeInput) {
						t.Errorf("transcoder got input %q, want %q", got, string(tc.transcodeInput))
					}

					if tc.transcodeError != nil {
						return tc.transcodeError
					}

					_, err = w.Write(tc.transcodeOutput)
					return err
				}),
			}
			var (
				enc EncMode
				err error
			)
			if tc.tags != nil {
				enc, err = opts.EncModeWithTags(tc.tags)
			} else {
				enc, err = opts.EncMode()
			}
			if err != nil {
				t.Fatal(err)
			}

			b, err := enc.Marshal(tc.value)
			if tc.wantErrorMsg != "" {
				if err == nil {
					t.Errorf("Marshal(%v) didn't return an error, want error %q", tc.value, tc.wantErrorMsg)
				} else if gotErrorMsg := err.Error(); gotErrorMsg != tc.wantErrorMsg {
					t.Errorf("Marshal(%v) returned error %q, want %q", tc.value, gotErrorMsg, tc.wantErrorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Marshal(%v) returned non-nil error %v", tc.value, err)
				} else if !bytes.Equal(b, tc.wantCborData) {
					t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.wantCborData)
				}
			}
		})
	}
}

type binaryTextMarshalerWithUnsupportedType map[chan bool]any

func (binaryTextMarshalerWithUnsupportedType) MarshalBinary() ([]byte, error) {
	return []byte("MarshalBinary"), nil
}

func (binaryTextMarshalerWithUnsupportedType) MarshalText() ([]byte, error) {
	return []byte("MarshalText"), nil
}

type binaryJSONMarshalerWithUnsupportedType map[chan bool]any

func (binaryJSONMarshalerWithUnsupportedType) MarshalBinary() ([]byte, error) {
	return []byte("MarshalBinary"), nil
}

func (binaryJSONMarshalerWithUnsupportedType) MarshalJSON() ([]byte, error) {
	return []byte("MarshalJSON"), nil
}

type binaryTextJSONMarshalerWithUnsupportedType map[chan bool]any

func (binaryTextJSONMarshalerWithUnsupportedType) MarshalBinary() ([]byte, error) {
	return []byte("MarshalBinary"), nil
}

func (binaryTextJSONMarshalerWithUnsupportedType) MarshalText() ([]byte, error) {
	return []byte("MarshalText"), nil
}

func (binaryTextJSONMarshalerWithUnsupportedType) MarshalJSON() ([]byte, error) {
	return []byte("MarshalJSON"), nil
}

type isZeroBinaryTextMarshalerWithUnsupportedType map[chan bool]any

func (isZeroBinaryTextMarshalerWithUnsupportedType) MarshalBinary() ([]byte, error) {
	return []byte("MarshalBinary"), nil
}

func (isZeroBinaryTextMarshalerWithUnsupportedType) MarshalText() ([]byte, error) {
	return []byte("MarshalText"), nil
}

func (isZeroBinaryTextMarshalerWithUnsupportedType) IsZero() bool {
	return true
}

type textMarshalerInt int

func (textMarshalerInt) MarshalText() ([]byte, error) {
	return []byte("MarshalText"), nil
}

type jsonMarshalerInt int

func (jsonMarshalerInt) MarshalJSON() ([]byte, error) {
	return []byte("MarshalJSON"), nil
}

type binaryMarshalerToArrayStruct struct {
	_ struct{} `cbor:",toarray"`
	A int
}

func (binaryMarshalerToArrayStruct) MarshalBinary() ([]byte, error) {
	return []byte("MarshalBinary"), nil
}

func TestOmitZeroWithMultipleMarshalers(t *testing.T) {
	testMarshalWithFP(t, testOmitZeroWithMultipleMarshalers)
}

func testOmitZeroWithMultipleMarshalers(t *testing.T) {
	jsonTranscoder := transcodeFunc(func(w io.Writer, _ io.Reader) error {
		_, err := w.Write([]byte{0x61, 'j'}) // "j"
		return err
	})

	for _, tc := range []struct {
		name      string
		opts      EncOptions
		in        any
		want      []byte
		wantError bool
	}{
		// Outermost marshaler is enabled, inner marshalers are disabled, and underlying type is unsupported.
		{
			name: "non-zero unsupported type implementing binary and text marshalers with BinaryMarshalerByteString and TextMarshalerNone",
			opts: EncOptions{},
			in: struct {
				M binaryTextMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{
				M: binaryTextMarshalerWithUnsupportedType{},
			},
			want: mustHexDecode("a1616d4d4d61727368616c42696e617279"), // {"m": h'4D61727368616C42696E617279'}
		},
		{
			name: "zero unsupported type implementing binary and text marshalers with BinaryMarshalerByteString and TextMarshalerNone",
			opts: EncOptions{},
			in: struct {
				M binaryTextMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{},
			want: mustHexDecode("a0"), // {}
		},
		{
			name: "non-zero unsupported type implementing binary and json marshalers with BinaryMarshalerByteString and nil JSONMarshalerTranscoder",
			opts: EncOptions{},
			in: struct {
				M binaryJSONMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{
				M: binaryJSONMarshalerWithUnsupportedType{},
			},
			want: mustHexDecode("a1616d4d4d61727368616c42696e617279"), // {"m": h'4D61727368616C42696E617279'}
		},
		{
			name: "non-zero unsupported type implementing all marshalers with BinaryMarshalerByteString, TextMarshalerNone, and nil JSONMarshalerTranscoder",
			opts: EncOptions{},
			in: struct {
				M binaryTextJSONMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{
				M: binaryTextJSONMarshalerWithUnsupportedType{},
			},
			want: mustHexDecode("a1616d4d4d61727368616c42696e617279"), // {"m": h'4D61727368616C42696E617279'}
		},
		{
			name: "zero unsupported type implementing all marshalers with BinaryMarshalerByteString, TextMarshalerNone, and nil JSONMarshalerTranscoder",
			opts: EncOptions{},
			in: struct {
				M binaryTextJSONMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{},
			want: mustHexDecode("a0"), // {}
		},

		// Outermost marshaler is disabled, an inner marshalers are enabled, and underlying type is unsupported.
		{
			name: "non-zero unsupported type implementing all marshalers with BinaryMarshalerNone and TextMarshalerTextString",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerNone, TextMarshaler: TextMarshalerTextString},
			in: struct {
				M binaryTextJSONMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{
				M: binaryTextJSONMarshalerWithUnsupportedType{},
			},
			want: mustHexDecode("a1616d6b4d61727368616c54657874"), // {"m": "MarshalText"}
		},
		{
			name: "zero unsupported type implementing all marshalers with BinaryMarshalerNone and TextMarshalerTextString",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerNone, TextMarshaler: TextMarshalerTextString},
			in: struct {
				M binaryTextJSONMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{},
			want: mustHexDecode("a0"), // {}
		},
		{
			name: "non-zero unsupported type implementing all marshalers with BinaryMarshalerNone and non-nil jsonTranscoder",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerNone, JSONMarshalerTranscoder: jsonTranscoder},
			in: struct {
				M binaryTextJSONMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{
				M: binaryTextJSONMarshalerWithUnsupportedType{},
			},
			want: mustHexDecode("a1616d616a"), // {"m": "j"}
		},

		// Every marshaler is disabled and the underlying type is unsupported.
		{
			name: "non-zero unsupported type implementing all marshalers with all marshalers disabled",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerNone, TextMarshaler: TextMarshalerNone},
			in: struct {
				M binaryTextJSONMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{
				M: binaryTextJSONMarshalerWithUnsupportedType{},
			},
			wantError: true,
		},
		{
			name: "zero unsupported type implementing all marshalers with all marshalers disabled",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerNone, TextMarshaler: TextMarshalerNone},
			in: struct {
				M binaryTextJSONMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{},
			wantError: true,
		},

		// IsZero() method is honored when a marshaler is enabled.
		{
			name: "non-nil unsupported type implementing IsZeroer and binary and text marshalers with BinaryMarshalerByteString and TextMarshalerNone",
			opts: EncOptions{},
			in: struct {
				M isZeroBinaryTextMarshalerWithUnsupportedType `cbor:"m,omitzero"`
			}{M: isZeroBinaryTextMarshalerWithUnsupportedType{}},
			want: mustHexDecode("a0"), // {}
		},

		// Marshaler is disabled and the underlying type is supported, so the
		// zero check uses the underlying type.
		{
			name: "zero supported type implementing TextMarshaler with TextMarshalerNone",
			opts: EncOptions{TextMarshaler: TextMarshalerNone},
			in: struct {
				M textMarshalerInt `cbor:"m,omitzero"`
			}{},
			want: mustHexDecode("a0"), // {}
		},
		{
			name: "non-zero supported type implementing TextMarshaler with TextMarshalerNone",
			opts: EncOptions{TextMarshaler: TextMarshalerNone},
			in: struct {
				M textMarshalerInt `cbor:"m,omitzero"`
			}{M: 1},
			want: mustHexDecode("a1616d01"), // {"m": 1}
		},
		{
			name: "zero supported type implementing JSONMarshaler with nil JSONMarshalerTranscoder",
			opts: EncOptions{},
			in: struct {
				M jsonMarshalerInt `cbor:"m,omitzero"`
			}{},
			want: mustHexDecode("a0"), // {}
		},
		{
			name: "non-zero supported type implementing JSONMarshaler with nil JSONMarshalerTranscoder",
			opts: EncOptions{},
			in: struct {
				M jsonMarshalerInt `cbor:"m,omitzero"`
			}{M: 1},
			want: mustHexDecode("a1616d01"), // {"m": 1}
		},

		// Zero check of the underlying type is preserved when a marshaler is
		// enabled: a toarray struct with fields is never zero.
		{
			name: "zero toarray struct implementing BinaryMarshaler with BinaryMarshalerByteString",
			opts: EncOptions{BinaryMarshaler: BinaryMarshalerByteString},
			in: struct {
				M binaryMarshalerToArrayStruct `cbor:"m,omitzero"`
			}{},
			want: mustHexDecode("a1616d4d4d61727368616c42696e617279"), // {"m": h'4D61727368616C42696E617279'}
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Fatalf("EncMode(): unexpected error: %v", err)
			}

			got, err := em.Marshal(tc.in)
			if tc.wantError {
				if err == nil {
					t.Fatalf("Marshal(%v): expected *UnsupportedTypeError, got nil", tc.in)
				} else if _, ok := err.(*UnsupportedTypeError); !ok {
					t.Errorf("Marshal(%v) returned %T, want *UnsupportedTypeError", tc.in, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Marshal(%v): unexpected error: %v", tc.in, err)
				} else if !bytes.Equal(tc.want, got) {
					t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.in, got, tc.want)
				}
			}
		})
	}
}

func TestMarshalerErrorUnwrap(t *testing.T) {
	testMarshalWithFP(t, testMarshalerErrorUnwrap)
}

func testMarshalerErrorUnwrap(t *testing.T) {
	innerErr := errors.New("MarshalCBOR: error")
	v := marshalCBORError(innerErr.Error())
	_, err := Marshal(v)
	if err == nil {
		t.Fatal("Marshal() didn't return an error")
	}
	var me *MarshalerError
	if errors.As(err, &me) {
		if unwrapped := me.Unwrap(); unwrapped == nil || unwrapped.Error() != innerErr.Error() {
			t.Errorf("MarshalerError.Unwrap() = %v, want error with message %q", unwrapped, innerErr.Error())
		}
	}
	// The marshalCBORError type returns the raw error, not a MarshalerError,
	// so also test Unwrap directly by constructing a MarshalerError.
	me2 := &MarshalerError{typ: reflect.TypeFor[int](), err: innerErr}
	if unwrapped := me2.Unwrap(); unwrapped != innerErr {
		t.Errorf("MarshalerError.Unwrap() = %v, want %v", unwrapped, innerErr)
	}
}

func TestTranscodeErrorUnwrap(t *testing.T) {
	innerErr := errors.New("transcode failed")
	te := TranscodeError{err: innerErr, rtype: reflect.TypeFor[int](), sourceFormat: "json", targetFormat: "cbor"}
	if unwrapped := te.Unwrap(); unwrapped != innerErr {
		t.Errorf("TranscodeError.Unwrap() = %v, want %v", unwrapped, innerErr)
	}
}

func TestUnsupportedValueErrorMessage(t *testing.T) {
	err := &UnsupportedValueError{msg: "floating-point infinity"}
	want := "cbor: unsupported value: floating-point infinity"
	if got := err.Error(); got != want {
		t.Errorf("UnsupportedValueError.Error() = %q, want %q", got, want)
	}
}

func TestMarshalToBufferNilBuffer(t *testing.T) {
	// Test package-level function
	if err := MarshalToBuffer(42, nil); err == nil {
		t.Error("MarshalToBuffer(42, nil) didn't return an error")
	}

	// Test UserBufferEncMode method
	bem, err := EncOptions{}.UserBufferEncMode()
	if err != nil {
		t.Fatal(err)
	}
	if err := bem.MarshalToBuffer(42, nil); err == nil {
		t.Error("UserBufferEncMode.MarshalToBuffer(42, nil) didn't return an error")
	}
}

func TestEncodedHeadLength(t *testing.T) {
	for _, tc := range []struct {
		name       string
		args       []uint64
		wantLength int
	}{
		{
			name:       "arg <= 23",
			args:       []uint64{0, 1, 22, 23},
			wantLength: 1,
		},
		{
			name:       "23 < arg <= math.MaxUint8",
			args:       []uint64{24, 25, math.MaxUint8 - 1, math.MaxUint8},
			wantLength: 2,
		},
		{
			name:       "math.MaxUint8 < arg <= math.MaxUint16",
			args:       []uint64{math.MaxUint8 + 1, math.MaxUint8 + 2, math.MaxUint16 - 1, math.MaxUint16},
			wantLength: 3,
		},
		{
			name:       "math.MaxUint16 < arg <= math.MaxUint32",
			args:       []uint64{math.MaxUint16 + 1, math.MaxUint16 + 2, math.MaxUint32 - 1, math.MaxUint32},
			wantLength: 5,
		},
		{
			name:       "math.MaxUint32 < arg <= math.MaxUint64",
			args:       []uint64{math.MaxUint32 + 1, math.MaxUint32 + 2, math.MaxUint64 - 1, math.MaxUint64},
			wantLength: 9,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, n := range tc.args {
				got := encodedHeadLength(n)
				if got != tc.wantLength {
					t.Errorf("encodedHeadLength(%d) = %d, want %d", n, got, tc.wantLength)
				}
				b := appendHead(nil, byte(cborTypePositiveInt), n)
				if len(b) != tc.wantLength {
					t.Errorf("appendHead(%d) encoded %d bytes, want %d bytes", n, len(b), tc.wantLength)
				}
			}
		})
	}
}

func TestMarshalSelfReferenceDataTypes(t *testing.T) {
	testMarshalWithFP(t, testMarshalSelfReferenceDataTypes)
}

func testMarshalSelfReferenceDataTypes(t *testing.T) {
	testCases := []struct {
		name     string
		value    any
		wantData []byte
	}{
		{
			name:     "self-ref slice type",
			value:    selfRefSlice{selfRefSlice{}},
			wantData: []byte{0x81, 0x80},
		},
		{
			name:     "self-ref slice type with pointer",
			value:    selfRefSliceWithP{&selfRefSliceWithP{}},
			wantData: []byte{0x81, 0x80},
		},
		{
			name:     "slice type with self-ref map",
			value:    sliceWithSelfRefMap{selfRefMap{"a": selfRefMap{}}},
			wantData: []byte{0x81, 0xa1, 0x61, 0x61, 0xa0},
		},
		{
			name:     "self-ref map type",
			value:    selfRefMap{"a": selfRefMap{}},
			wantData: []byte{0xa1, 0x61, 0x61, 0xa0},
		},
		{
			name:     "map type with self-ref slice",
			value:    mapWithSelfRefSlice{"a": selfRefSlice{selfRefSlice{}}},
			wantData: []byte{0xa1, 0x61, 0x61, 0x81, 0x80},
		},
		{
			name:     "mutual-ref map and slice types",
			value:    mutualRefMap{"a": mutualRefSlice{}},
			wantData: []byte{0xa1, 0x61, 0x61, 0x80},
		},
		{
			name:     "binary marshaler over unsupported self-ref type",
			value:    marshalerOverSelfRefUnsupported{},
			wantData: []byte{0x41, 0x74},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if b, err := Marshal(tc.value); err != nil {
				t.Errorf("Marshal(%v): unexpected error: %v", tc.value, err)
			} else if !bytes.Equal(b, tc.wantData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.value, b, tc.wantData)
			}
		})
	}
}

// Unsupported (chan key) types in a cycle through a slice, a pointer, etc.
type (
	selfRefUnsupportedSliceMap map[chan bool][]selfRefUnsupportedSliceMap
	selfRefUnsupportedPtrMap   map[chan bool]*selfRefUnsupportedPtrMap

	selfRefUnsupportedMap    map[chan bool][]selfRefUnsupportedMap
	selfRefUnsupportedArray1 [1]selfRefUnsupportedMap
	selfRefUnsupportedArray2 [2][]selfRefUnsupportedMap
	selfRefUnsupportedRoot   map[*selfRefUnsupportedArray1]selfRefUnsupportedArray2

	selfRefUnsupportedSliceMap2     map[chan bool][]selfRefUnsupportedSliceMap2
	marshalerOverSelfRefUnsupported []selfRefUnsupportedSliceMap2
)

func (marshalerOverSelfRefUnsupported) MarshalBinary() ([]byte, error) {
	return []byte("t"), nil
}

func TestMarshalSelfReferenceUnsupportedType(t *testing.T) {
	testMarshalWithFP(t, testMarshalSelfReferenceUnsupportedType)
}

func testMarshalSelfReferenceUnsupportedType(t *testing.T) {
	testCases := []struct {
		name   string
		values []any
	}{
		{
			name: "empty slice after its element type",
			values: []any{
				selfRefUnsupportedSliceMap{},
				[]selfRefUnsupportedSliceMap{},
			},
		},
		{
			name: "nil pointer after its element type",
			values: []any{
				selfRefUnsupportedPtrMap{},
				(*selfRefUnsupportedPtrMap)(nil),
			},
		},
		{
			name: "type reusing an in-build type from another branch",
			values: []any{
				selfRefUnsupportedRoot{},
				selfRefUnsupportedArray2{},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, v := range tc.values {
				_, err := Marshal(v)
				if err == nil {
					t.Errorf("Marshal(%v) = nil error, want *UnsupportedTypeError", v)
				} else if _, ok := err.(*UnsupportedTypeError); !ok {
					t.Errorf("Marshal(%v) error type = %T, want *UnsupportedTypeError", v, err)
				}
			}
		})
	}
}

type vCBORMarshaler struct{}

func (r vCBORMarshaler) MarshalCBOR() ([]byte, error) {
	return []byte{0xf6}, nil
}

type pCBORMarshaler struct{}

func (r *pCBORMarshaler) MarshalCBOR() ([]byte, error) {
	return []byte{0xf7}, nil
}

type vBinaryMarshaler struct{}

func (r vBinaryMarshaler) MarshalBinary() ([]byte, error) {
	return mustHexDecode("010203"), nil
}

type pBinaryMarshaler struct{}

func (r *pBinaryMarshaler) MarshalBinary() ([]byte, error) {
	return mustHexDecode("040506"), nil
}

type vTextMarshaler struct{}

func (r vTextMarshaler) MarshalText() ([]byte, error) {
	return mustHexDecode("61"), nil
}

type pTextMarshaler struct{}

func (r *pTextMarshaler) MarshalText() ([]byte, error) {
	return mustHexDecode("62"), nil
}

type vJSONMarshaler struct{}

func (r vJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte("true"), nil
}

type pJSONMarshaler struct{}

func (r *pJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte("false"), nil
}

func TestMarshalAddressableMarshaler(t *testing.T) {
	testMarshalWithFP(t, testMarshalAddressableMarshaler)
}

func testMarshalAddressableMarshaler(t *testing.T) {
	em, err := EncOptions{
		TextMarshaler: TextMarshalerTextString,
		JSONMarshalerTranscoder: transcodeFunc(func(w io.Writer, r io.Reader) error {
			b, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			if bytes.Equal(b, []byte("true")) {
				_, err = w.Write([]byte{0xf5})
			} else if bytes.Equal(b, []byte("false")) {
				_, err = w.Write([]byte{0xf4})
			} else {
				err = fmt.Errorf("unrecognized json data %x", b)
			}
			return err
		}),
	}.EncMode()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Marshaler", func(t *testing.T) {
		v := struct {
			V vCBORMarshaler `cbor:",omitempty"`
			P pCBORMarshaler `cbor:",omitempty"`
		}{
			V: vCBORMarshaler{},
			P: pCBORMarshaler{},
		}
		want := mustHexDecode("a26156f66150f7") // {"V": null, "P": undefined}
		b, err := em.Marshal(&v)
		if err != nil {
			t.Fatalf("Marshal(%+v): unexpected error: %v", v, err)
		}
		if !bytes.Equal(b, want) {
			t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", v, b, want)
		}
	})

	t.Run("BinaryMarshaler", func(t *testing.T) {
		v := struct {
			V vBinaryMarshaler `cbor:",omitempty"`
			P pBinaryMarshaler `cbor:",omitempty"`
		}{
			V: vBinaryMarshaler{},
			P: pBinaryMarshaler{},
		}
		want := mustHexDecode("a2615643010203615043040506") // {"V": h'010203', "P": h'040506'}
		b, err := em.Marshal(&v)
		if err != nil {
			t.Fatalf("Marshal(%+v): unexpected error: %v", v, err)
		}
		if !bytes.Equal(b, want) {
			t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", v, b, want)
		}
	})

	t.Run("TextMarshaler", func(t *testing.T) {
		v := struct {
			V vTextMarshaler `cbor:",omitempty"`
			P pTextMarshaler `cbor:",omitempty"`
		}{
			V: vTextMarshaler{},
			P: pTextMarshaler{},
		}
		want := mustHexDecode("a26156616161506162") // {"V": "a", "P": "b"}
		b, err := em.Marshal(&v)
		if err != nil {
			t.Fatalf("Marshal(%+v): unexpected error: %v", v, err)
		}
		if !bytes.Equal(b, want) {
			t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", v, b, want)
		}
	})

	t.Run("json.Marshaler", func(t *testing.T) {
		v := struct {
			V vJSONMarshaler `cbor:",omitempty"`
			P pJSONMarshaler `cbor:",omitempty"`
		}{
			V: vJSONMarshaler{},
			P: pJSONMarshaler{},
		}
		want := mustHexDecode("a26156f56150f4") // {"V": true, "P": false}
		b, err := em.Marshal(&v)
		if err != nil {
			t.Fatalf("Marshal(%+v): unexpected error: %v", v, err)
		}
		if !bytes.Equal(b, want) {
			t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", v, b, want)
		}
	})
}

func TestMarshalNilByteStringArrayAndMap(t *testing.T) {
	testMarshalWithFP(t, testMarshalNilByteStringArrayAndMap)
}

func testMarshalNilByteStringArrayAndMap(t *testing.T) {
	type s1 struct {
		B []byte
	}
	type s2 struct {
		A []any
	}
	type s3 struct {
		M map[string]any
	}

	nilContainersAsNullEM, _ := EncOptions{
		NilContainers: NilContainerAsNull,
	}.EncMode()

	nilContainersAsEmptyEM, _ := EncOptions{
		NilContainers: NilContainerAsEmpty,
	}.EncMode()

	for _, tc := range []struct {
		name string
		em   EncMode
		in   any
		want []byte
	}{
		{
			name: "[]byte(nil) with NilContainerAsNull opt",
			em:   nilContainersAsNullEM,
			in:   []byte(nil),
			want: []byte{0xf6}, // null
		},
		{
			name: "[]byte(nil) with NilContainerAsEmpty opt",
			em:   nilContainersAsEmptyEM,
			in:   []byte(nil),
			want: []byte{0x40}, // h''
		},
		{
			name: "struct{[]byte(nil)} with NilContainerAsNull opt",
			em:   nilContainersAsNullEM,
			in:   s1{B: []byte(nil)},
			want: []byte{0xa1, 0x61, 0x42, 0xf6}, // {"B": null}
		},
		{
			name: "struct{[]byte(nil)} with NilContainerAsEmpty opt",
			em:   nilContainersAsEmptyEM,
			in:   s1{B: []byte(nil)},
			want: []byte{0xa1, 0x61, 0x42, 0x40}, // {"B": h''}
		},
		{
			name: "[]any(nil) with NilContainerAsNull opt",
			em:   nilContainersAsNullEM,
			in:   []any(nil),
			want: []byte{0xf6}, // null
		},
		{
			name: "[]any(nil) with NilContainerAsEmpty opt",
			em:   nilContainersAsEmptyEM,
			in:   []any(nil),
			want: []byte{0x80}, // []
		},
		{
			name: "struct{[]any(nil)} with NilContainerAsNull opt",
			em:   nilContainersAsNullEM,
			in:   s2{A: []any(nil)},
			want: []byte{0xa1, 0x61, 0x41, 0xf6}, // {"A": null}
		},
		{
			name: "struct{[]any(nil)} with NilContainerAsEmpty opt",
			em:   nilContainersAsEmptyEM,
			in:   s2{A: []any(nil)},
			want: []byte{0xa1, 0x61, 0x41, 0x80}, // {"A": []}
		},
		{
			name: "map[string]any(nil) with NilContainerAsNull opt",
			em:   nilContainersAsNullEM,
			in:   map[string]any(nil),
			want: []byte{0xf6}, // null
		},
		{
			name: "map[string]any(nil) with NilContainerAsEmpty opt",
			em:   nilContainersAsEmptyEM,
			in:   map[string]any(nil),
			want: []byte{0xa0}, // {}
		},
		{
			name: "struct{map[string]any(nil)} with NilContainerAsNull opt",
			em:   nilContainersAsNullEM,
			in:   s3{M: map[string]any(nil)},
			want: []byte{0xa1, 0x61, 0x4d, 0xf6}, // {"M": null}
		},
		{
			name: "struct{map[string]any(nil)} with NilContainerAsEmpty opt",
			em:   nilContainersAsEmptyEM,
			in:   s3{M: map[string]any(nil)},
			want: []byte{0xa1, 0x61, 0x4d, 0xa0}, // {"M": {}}
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.em.Marshal(tc.in)
			if err != nil {
				t.Fatalf("Marshal(%#v): unexpected error: %v", tc.in, err)
			}

			if !bytes.Equal(tc.want, got) {
				t.Errorf("Marshal(%#v) = 0x%x, want 0x%x", tc.in, got, tc.want)
			}
		})
	}
}

func TestMarshalNilInterface(t *testing.T) {
	testMarshalWithFP(t, testMarshalNilInterface)
}

func testMarshalNilInterface(t *testing.T) {
	type s struct {
		R io.Reader
	}

	v := s{}
	want := []byte{0xa1, 0x61, 0x52, 0xf6} // {"R": null}

	got, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%v): unexpected error: %v", v, err)
	}

	if !bytes.Equal(want, got) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, got, want)
	}
}

func TestMarshalArrayMapError(t *testing.T) {
	testMarshalWithFP(t, testMarshalArrayMapError)
}

func testMarshalArrayMapError(t *testing.T) {
	sortNoneEM, _ := EncOptions{
		Sort: SortNone,
	}.EncMode()

	sortLengthFirstEM, _ := EncOptions{
		Sort: SortLengthFirst,
	}.EncMode()

	c := make(chan bool)

	for _, tc := range []struct {
		name string
		em   EncMode
		in   any
	}{
		{
			name: "[]any",
			em:   defaultEncMode,
			in:   []any{c},
		},
		{
			name: "map[string]any with SortNone",
			em:   sortNoneEM,
			in:   map[string]any{"a": c, "b": c},
		},
		{
			name: "map[string]any with SortLengthFirst",
			em:   sortLengthFirstEM,
			in:   map[string]any{"a": c, "b": c},
		},
		{
			name: "map[any]any with SortNone (unsupported key)",
			em:   sortNoneEM,
			in:   map[any]any{c: 1, "b": 2},
		},
		{
			name: "map[any]any with SortLengthFirst (unsupported key)",
			em:   sortLengthFirstEM,
			in:   map[any]any{c: 1, "b": 2},
		},
		{
			name: "map[any]any with SortNone (unsupported value)",
			em:   sortNoneEM,
			in:   map[any]any{"a": c, "b": c},
		},
		{
			name: "map[any]any with SortLengthFirst (unsupported value)",
			em:   sortLengthFirstEM,
			in:   map[any]any{"a": c, "b": c},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.em.Marshal(tc.in)
			if err == nil {
				t.Errorf("Marshal(%v): expected *UnsupportedTypeError, got nil", tc.in)
			} else if _, ok := err.(*UnsupportedTypeError); !ok {
				t.Errorf("Marshal(%v) returned %T, want *UnsupportedTypeError", tc.in, err)
			}
		})
	}
}

type (
	namedSliceAny        []any
	namedMapAnyAny       map[any]any
	namedMapStringAny    map[string]any
	namedMapStringString map[string]string
)

// fastPathTypes has fields whose types are encoded by the fast path.
type fastPathTypes struct {
	A    any                `cbor:"a"`
	S    []any              `cbor:"b"`
	PS   *[]any             `cbor:"c"`
	MAA  map[any]any        `cbor:"d"`
	PMAA *map[any]any       `cbor:"e"`
	MSA  map[string]any     `cbor:"f"`
	PMSA *map[string]any    `cbor:"g"`
	MSS  map[string]string  `cbor:"h"`
	PMSS *map[string]string `cbor:"i"`
}

// reflectPathTypes has the same fields as fastPathTypes with named types,
// so the fields are encoded by the reflect path.  Values nested inside
// these fields can still be encoded by the fast path.
type reflectPathTypes struct {
	A    namedAny              `cbor:"a"`
	S    namedSliceAny         `cbor:"b"`
	PS   *namedSliceAny        `cbor:"c"`
	MAA  namedMapAnyAny        `cbor:"d"`
	PMAA *namedMapAnyAny       `cbor:"e"`
	MSA  namedMapStringAny     `cbor:"f"`
	PMSA *namedMapStringAny    `cbor:"g"`
	MSS  namedMapStringString  `cbor:"h"`
	PMSS *namedMapStringString `cbor:"i"`
}

func newReflectPathTypes(v *fastPathTypes) reflectPathTypes {
	return reflectPathTypes{
		A:    v.A,
		S:    v.S,
		PS:   (*namedSliceAny)(v.PS),
		MAA:  v.MAA,
		PMAA: (*namedMapAnyAny)(v.PMAA),
		MSA:  v.MSA,
		PMSA: (*namedMapStringAny)(v.PMSA),
		MSS:  v.MSS,
		PMSS: (*namedMapStringString)(v.PMSS),
	}
}

func TestMarshalStructWithFastPathTypes(t *testing.T) {
	testMarshalWithFP(t, testMarshalStructWithFastPathTypes)
}

func testMarshalStructWithFastPathTypes(t *testing.T) {
	empty := fastPathTypes{
		A:    map[string]any{},
		S:    []any{},
		PS:   &[]any{},
		MAA:  map[any]any{},
		PMAA: &map[any]any{},
		MSA:  map[string]any{},
		PMSA: &map[string]any{},
		MSS:  map[string]string{},
		PMSS: &map[string]string{},
	}

	populated := fastPathTypes{
		A:    map[string]any{"bb": 1, "a": []any{}},
		S:    []any{1, "a", nil, true, -1, 1.5},
		PS:   &[]any{[]byte{1}},
		MAA:  map[any]any{1: "a", -1: "b", "a": 1, 1000: nil},
		PMAA: &map[any]any{true: false},
		MSA:  map[string]any{"c": map[any]any{"x": "y"}, "bb": -2, "a": "s"},
		PMSA: &map[string]any{"k": []any{1}},
		MSS:  map[string]string{"ccc": "3", "a": "1", "bb": "2"},
		PMSS: &map[string]string{"k": "v"},
	}

	testCases := []struct {
		name      string
		opts      EncOptions
		in        fastPathTypes
		want      []byte
		roundTrip bool
	}{
		{
			name: "nil values with NilContainerAsNull",
			opts: EncOptions{NilContainers: NilContainerAsNull},
			in:   fastPathTypes{},
			// {"a": null, "b": null, "c": null, "d": null, "e": null, "f": null, "g": null, "h": null, "i": null}
			want:      mustHexDecode("a96161f66162f66163f66164f66165f66166f66167f66168f66169f6"),
			roundTrip: true,
		},
		{
			name: "nil values with NilContainerAsEmpty",
			opts: EncOptions{NilContainers: NilContainerAsEmpty},
			in:   fastPathTypes{},
			// {"a": null, "b": [], "c": null, "d": {}, "e": null, "f": {}, "g": null, "h": {}, "i": null}
			want:      mustHexDecode("a96161f66162806163f66164a06165f66166a06167f66168a06169f6"),
			roundTrip: true,
		},
		{
			name: "empty values",
			opts: EncOptions{},
			in:   empty,
			// {"a": {}, "b": [], "c": [], "d": {}, "e": {}, "f": {}, "g": {}, "h": {}, "i": {}}
			want:      mustHexDecode("a96161a06162806163806164a06165a06166a06167a06168a06169a0"),
			roundTrip: true,
		},
		{
			name: "populated values with SortLengthFirst",
			opts: EncOptions{Sort: SortLengthFirst},
			in:   populated,
			// {"a": {"a": [], "bb": 1}, "b": [1, "a", null, true, -1, 1.5], "c": [h'01'], "d": {1: "a", -1: "b", "a": 1, 1000: null}, "e": {true: false}, "f": {"a": "s", "c": {"x": "y"}, "bb": -2}, "g": {"k": [1]}, "h": {"a": "1", "bb": "2", "ccc": "3"}, "i": {"k": "v"}}
			want:      mustHexDecode("a96161a261618062626201616286016161f6f520fb3ff800000000000061638141016164a40161612061626161011903e8f66165a1f5f46166a3616161736163a161786179626262216167a1616b81016168a36161613162626261326363636361336169a1616b6176"),
			roundTrip: true,
		},
		{
			name: "populated values with SortBytewiseLexical",
			opts: EncOptions{Sort: SortBytewiseLexical},
			in:   populated,
			// {"a": {"a": [], "bb": 1}, "b": [1, "a", null, true, -1, 1.5], "c": [h'01'], "d": {1: "a", 1000: null, -1: "b", "a": 1}, "e": {true: false}, "f": {"a": "s", "c": {"x": "y"}, "bb": -2}, "g": {"k": [1]}, "h": {"a": "1", "bb": "2", "ccc": "3"}, "i": {"k": "v"}}
			want:      mustHexDecode("a96161a261618062626201616286016161f6f520fb3ff800000000000061638141016164a40161611903e8f62061626161016165a1f5f46166a3616161736163a161786179626262216167a1616b81016168a36161613162626261326363636361336169a1616b6176"),
			roundTrip: true,
		},
		{
			name: "populated values with SortLengthFirst and StringToByteString",
			opts: EncOptions{Sort: SortLengthFirst, String: StringToByteString},
			in:   populated,
			// {"a": {h'61': [], h'6262': 1}, "b": [1, h'61', null, true, -1, 1.5], "c": [h'01'], "d": {1: h'61', -1: h'62', h'61': 1, 1000: null}, "e": {true: false}, "f": {h'61': h'73', h'63': {h'78': h'79'}, h'6262': -2}, "g": {h'6b': [1]}, "h": {h'61': h'31', h'6262': h'32', h'636363': h'33'}, "i": {h'6b': h'76'}}
			want: mustHexDecode("a96161a241618042626201616286014161f6f520fb3ff800000000000061638141016164a40141612041624161011903e8f66165a1f5f46166a3416141734163a141784179426262216167a1416b81016168a34161413142626241324363636341336169a1416b4176"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			em, err := tc.opts.EncMode()
			if err != nil {
				t.Fatalf("EncMode(): unexpected error: %v", err)
			}

			// Encode struct fields that can use fast path
			got, err := em.Marshal(tc.in)
			if err != nil {
				t.Fatalf("Marshal(%v): unexpected error: %v", tc.in, err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", tc.in, got, tc.want)
			}

			// Encode struct fields that can't use fast path
			inWithNamedTypes := newReflectPathTypes(&tc.in)
			got, err = em.Marshal(inWithNamedTypes)
			if err != nil {
				t.Fatalf("Marshal(%v): unexpected error: %v", inWithNamedTypes, err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", inWithNamedTypes, got, tc.want)
			}

			if !tc.roundTrip {
				return
			}

			// Round-trip fast path types
			var fpv fastPathTypes
			if err = Unmarshal(tc.want, &fpv); err != nil {
				t.Fatalf("Unmarshal(0x%x): unexpected error: %v", tc.want, err)
			}
			got, err = em.Marshal(fpv)
			if err != nil {
				t.Fatalf("Marshal(%v): unexpected error: %v", fpv, err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", fpv, got, tc.want)
			}

			// Round-trip reflect path types
			var rpv reflectPathTypes
			if err = Unmarshal(tc.want, &rpv); err != nil {
				t.Fatalf("Unmarshal(0x%x): unexpected error: %v", tc.want, err)
			}
			got, err = em.Marshal(rpv)
			if err != nil {
				t.Fatalf("Marshal(%v): unexpected error: %v", rpv, err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", rpv, got, tc.want)
			}
		})
	}
}

func TestOmitEmptyForFastPathTypes(t *testing.T) {
	testMarshalWithFP(t, testOmitEmptyForFastPathTypes)
}

func testOmitEmptyForFastPathTypes(t *testing.T) {
	type T struct {
		S   []any             `cbor:"s,omitempty"`
		MAA map[any]any       `cbor:"maa,omitempty"`
		MSA map[string]any    `cbor:"msa,omitempty"`
		MSS map[string]string `cbor:"mss,omitempty"`
	}

	v := T{S: []any{}, MAA: map[any]any{}, MSA: map[string]any{}, MSS: map[string]string{}}
	want := []byte{0xa0} // {}

	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%v): unexpected error: %v", v, err)
	} else if !bytes.Equal(b, want) {
		t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, b, want)
	}
}
