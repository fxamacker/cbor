// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

package cbor_test

import (
	"bytes"
	"encoding/hex"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/fxamacker/cbor"
)

var (
	typeBool            = reflect.TypeOf(true)
	typeUint            = reflect.TypeOf(uint(0))
	typeUint8           = reflect.TypeOf(uint8(0))
	typeUint16          = reflect.TypeOf(uint16(0))
	typeUint32          = reflect.TypeOf(uint32(0))
	typeUint64          = reflect.TypeOf(uint64(0))
	typeInt             = reflect.TypeOf(int(0))
	typeInt8            = reflect.TypeOf(int8(0))
	typeInt16           = reflect.TypeOf(int16(0))
	typeInt32           = reflect.TypeOf(int32(0))
	typeInt64           = reflect.TypeOf(int64(0))
	typeFloat32         = reflect.TypeOf(float32(0))
	typeFloat64         = reflect.TypeOf(float64(0))
	typeString          = reflect.TypeOf("")
	typeByteSlice       = reflect.TypeOf([]byte(nil))
	typeIntSlice        = reflect.TypeOf([]int{})
	typeStringSlice     = reflect.TypeOf([]string{})
	typeMapStringInt    = reflect.TypeOf(map[string]int{})
	typeMapStringString = reflect.TypeOf(map[string]string{})
	typeMapStringIntf   = reflect.TypeOf(map[string]interface{}{})
	typeIntf            = reflect.TypeOf([]interface{}(nil)).Elem()
)

type unmarshalTest struct {
	cborData            []byte
	emptyInterfaceValue interface{}
	values              []interface{}
	wrongTypes          []reflect.Type
}

type unmarshalFloatTest struct {
	cborData            []byte
	emptyInterfaceValue interface{}
	diff                float64
	values              []interface{}
	wrongTypes          []reflect.Type
}

// CBOR test data are from https://tools.ietf.org/html/rfc7049#appendix-A.
var unmarshalTests = []unmarshalTest{
	// positive integer
	{
		hexDecode("00"),
		uint64(0),
		[]interface{}{uint8(0), uint16(0), uint32(0), uint64(0), uint(0), int8(0), int16(0), int32(0), int64(0), int(0)},
		[]reflect.Type{typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("01"),
		uint64(1),
		[]interface{}{uint8(1), uint16(1), uint32(1), uint64(1), uint(1), int8(1), int16(1), int32(1), int64(1), int(1)},
		[]reflect.Type{typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("0a"),
		uint64(10),
		[]interface{}{uint8(10), uint16(10), uint32(10), uint64(10), uint(10), int8(10), int16(10), int32(10), int64(10), int(10)},
		[]reflect.Type{typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("17"),
		uint64(23),
		[]interface{}{uint8(23), uint16(23), uint32(23), uint64(23), uint(23), int8(23), int16(23), int32(23), int64(23), int(23)},
		[]reflect.Type{typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("1818"),
		uint64(24),
		[]interface{}{uint8(24), uint16(24), uint32(24), uint64(24), uint(24), int8(24), int16(24), int32(24), int64(24), int(24)},
		[]reflect.Type{typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("1819"),
		uint64(25),
		[]interface{}{uint8(25), uint16(25), uint32(25), uint64(25), uint(25), int8(25), int16(25), int32(25), int64(25), int(25)},
		[]reflect.Type{typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("1864"),
		uint64(100),
		[]interface{}{uint8(100), uint16(100), uint32(100), uint64(100), uint(100), int8(100), int16(100), int32(100), int64(100), int(100)},
		[]reflect.Type{typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("1903e8"),
		uint64(1000),
		[]interface{}{uint16(1000), uint32(1000), uint64(1000), uint(1000), int16(1000), int32(1000), int64(1000), int(1000)},
		[]reflect.Type{typeUint8, typeInt8, typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("1a000f4240"),
		uint64(1000000),
		[]interface{}{uint32(1000000), uint64(1000000), uint(1000000), int32(1000000), int64(1000000), int(1000000)},
		[]reflect.Type{typeUint8, typeUint16, typeInt8, typeInt16, typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("1b000000e8d4a51000"),
		uint64(1000000000000),
		[]interface{}{uint64(1000000000000), uint(1000000000000), int64(1000000000000), int(1000000000000)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeInt8, typeInt16, typeInt32, typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("1bffffffffffffffff"),
		uint64(18446744073709551615),
		[]interface{}{uint64(18446744073709551615), uint(18446744073709551615)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeInt8, typeInt16, typeInt32, typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	// negative integer
	{
		hexDecode("20"),
		int64(-1),
		[]interface{}{int8(-1), int16(-1), int32(-1), int64(-1), int(-1)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("29"),
		int64(-10),
		[]interface{}{int8(-10), int16(-10), int32(-10), int64(-10), int(-10)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("3863"),
		int64(-100),
		[]interface{}{int8(-100), int16(-100), int32(-100), int64(-100), int(-100)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("3903e7"),
		int64(-1000),
		[]interface{}{int16(-1000), int32(-1000), int64(-1000), int(-1000)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeFloat32, typeFloat64, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	//{"3bffffffffffffffff", int64(-18446744073709551616)}, // value overflows int64
	// byte string
	{
		hexDecode("40"),
		[]byte{},
		[]interface{}{[]byte{}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("4401020304"),
		[]byte{1, 2, 3, 4},
		[]interface{}{[]byte{1, 2, 3, 4}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("5f42010243030405ff"),
		[]byte{1, 2, 3, 4, 5},
		[]interface{}{[]byte{1, 2, 3, 4, 5}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	},
	// text string
	{
		hexDecode("60"),
		"",
		[]interface{}{""},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("6161"),
		"a",
		[]interface{}{"a"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("6449455446"),
		"IETF",
		[]interface{}{"IETF"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("62225c"),
		"\"\\",
		[]interface{}{"\"\\"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("62c3bc"),
		"ü",
		[]interface{}{"ü"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("63e6b0b4"),
		"水",
		[]interface{}{"水"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("64f0908591"),
		"𐅑",
		[]interface{}{"𐅑"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("7f657374726561646d696e67ff"),
		"streaming",
		[]interface{}{"streaming"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	},
	// array
	{
		hexDecode("80"),
		[]interface{}{},
		[]interface{}{[]interface{}{}, []byte{}, []string{}, []int{}, [...]int{}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeMapStringInt},
	},
	{
		hexDecode("83010203"),
		[]interface{}{uint64(1), uint64(2), uint64(3)},
		[]interface{}{[]interface{}{uint64(1), uint64(2), uint64(3)}, []byte{1, 2, 3}, []int{1, 2, 3}, []uint{1, 2, 3}, [...]int{1, 2, 3}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{})},
	},
	{
		hexDecode("8301820203820405"),
		[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}},
		[]interface{}{[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}, [...]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{})},
	},
	{
		hexDecode("83018202039f0405ff"),
		[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}},
		[]interface{}{[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}, [...]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{})},
	},
	{
		hexDecode("83019f0203ff820405"),
		[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}},
		[]interface{}{[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}, [...]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{})},
	},
	{
		hexDecode("98190102030405060708090a0b0c0d0e0f101112131415161718181819"),
		[]interface{}{uint64(1), uint64(2), uint64(3), uint64(4), uint64(5), uint64(6), uint64(7), uint64(8), uint64(9), uint64(10), uint64(11), uint64(12), uint64(13), uint64(14), uint64(15), uint64(16), uint64(17), uint64(18), uint64(19), uint64(20), uint64(21), uint64(22), uint64(23), uint64(24), uint64(25)},
		[]interface{}{
			[]interface{}{uint64(1), uint64(2), uint64(3), uint64(4), uint64(5), uint64(6), uint64(7), uint64(8), uint64(9), uint64(10), uint64(11), uint64(12), uint64(13), uint64(14), uint64(15), uint64(16), uint64(17), uint64(18), uint64(19), uint64(20), uint64(21), uint64(22), uint64(23), uint64(24), uint64(25)},
			[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[...]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{})},
	},
	{
		hexDecode("9fff"),
		[]interface{}{},
		[]interface{}{[]interface{}{}, []byte{}, []string{}, []int{}, [...]int{}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeMapStringInt},
	},
	{
		hexDecode("9f018202039f0405ffff"),
		[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}},
		[]interface{}{[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}, [...]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{})},
	},
	{
		hexDecode("9f01820203820405ff"),
		[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}},
		[]interface{}{[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}, [...]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{})},
	},
	{
		hexDecode("9f0102030405060708090a0b0c0d0e0f101112131415161718181819ff"),
		[]interface{}{uint64(1), uint64(2), uint64(3), uint64(4), uint64(5), uint64(6), uint64(7), uint64(8), uint64(9), uint64(10), uint64(11), uint64(12), uint64(13), uint64(14), uint64(15), uint64(16), uint64(17), uint64(18), uint64(19), uint64(20), uint64(21), uint64(22), uint64(23), uint64(24), uint64(25)},
		[]interface{}{
			[]interface{}{uint64(1), uint64(2), uint64(3), uint64(4), uint64(5), uint64(6), uint64(7), uint64(8), uint64(9), uint64(10), uint64(11), uint64(12), uint64(13), uint64(14), uint64(15), uint64(16), uint64(17), uint64(18), uint64(19), uint64(20), uint64(21), uint64(22), uint64(23), uint64(24), uint64(25)},
			[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[...]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{})},
	},
	{
		hexDecode("826161a161626163"),
		[]interface{}{"a", map[interface{}]interface{}{"b": "c"}},
		[]interface{}{[]interface{}{"a", map[interface{}]interface{}{"b": "c"}}, [...]interface{}{"a", map[interface{}]interface{}{"b": "c"}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{})},
	},
	{
		hexDecode("826161bf61626163ff"),
		[]interface{}{"a", map[interface{}]interface{}{"b": "c"}},
		[]interface{}{[]interface{}{"a", map[interface{}]interface{}{"b": "c"}}, [...]interface{}{"a", map[interface{}]interface{}{"b": "c"}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{})},
	},
	// map
	{
		hexDecode("a0"),
		map[interface{}]interface{}{},
		[]interface{}{map[interface{}]interface{}{}, map[string]bool{}, map[string]int{}, map[int]string{}, map[int]bool{}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeString, typeBool, typeIntSlice},
	},
	{
		hexDecode("a201020304"),
		map[interface{}]interface{}{uint64(1): uint64(2), uint64(3): uint64(4)},
		[]interface{}{map[interface{}]interface{}{uint64(1): uint64(2), uint64(3): uint64(4)}, map[uint]int{1: 2, 3: 4}, map[int]uint{1: 2, 3: 4}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("a26161016162820203"),
		map[interface{}]interface{}{"a": uint64(1), "b": []interface{}{uint64(2), uint64(3)}},
		[]interface{}{map[interface{}]interface{}{"a": uint64(1), "b": []interface{}{uint64(2), uint64(3)}},
			map[string]interface{}{"a": uint64(1), "b": []interface{}{uint64(2), uint64(3)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("a56161614161626142616361436164614461656145"),
		map[interface{}]interface{}{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"},
		[]interface{}{map[interface{}]interface{}{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"},
			map[string]interface{}{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"},
			map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("bf61610161629f0203ffff"),
		map[interface{}]interface{}{"a": uint64(1), "b": []interface{}{uint64(2), uint64(3)}},
		[]interface{}{map[interface{}]interface{}{"a": uint64(1), "b": []interface{}{uint64(2), uint64(3)}},
			map[string]interface{}{"a": uint64(1), "b": []interface{}{uint64(2), uint64(3)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("bf6346756ef563416d7421ff"),
		map[interface{}]interface{}{"Fun": true, "Amt": int64(-2)},
		[]interface{}{map[interface{}]interface{}{"Fun": true, "Amt": int64(-2)},
			map[string]interface{}{"Fun": true, "Amt": int64(-2)}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	// tag
	{
		hexDecode("c074323031332d30332d32315432303a30343a30305a"),
		"2013-03-21T20:04:00Z",
		[]interface{}{"2013-03-21T20:04:00Z"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	}, // 0: standard date/time
	{
		hexDecode("c11a514b67b0"),
		uint64(1363896240),
		[]interface{}{uint32(1363896240), uint64(1363896240), int32(1363896240), int64(1363896240)},
		[]reflect.Type{typeUint8, typeUint16, typeInt8, typeInt16, typeFloat32, typeFloat64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	}, // 1: epoch-based date/time
	{
		hexDecode("c249010000000000000000"),
		[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		[]interface{}{[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	}, // 2: positive bignum: 18446744073709551616
	{
		hexDecode("c349010000000000000000"),
		[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		[]interface{}{[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	}, // 3: negative bignum: -18446744073709551617
	{
		hexDecode("c1fb41d452d9ec200000"),
		float64(1363896240.5),
		[]interface{}{float32(1363896240.5), float64(1363896240.5)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	}, // 1: epoch-based date/time
	{
		hexDecode("d74401020304"),
		[]byte{0x01, 0x02, 0x03, 0x04},
		[]interface{}{[]byte{0x01, 0x02, 0x03, 0x04}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	}, // 23: expected conversion to base16 encoding
	{
		hexDecode("d818456449455446"),
		[]byte{0x64, 0x49, 0x45, 0x54, 0x46},
		[]interface{}{[]byte{0x64, 0x49, 0x45, 0x54, 0x46}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	}, // 24: encoded cborBytes data item
	{
		hexDecode("d82076687474703a2f2f7777772e6578616d706c652e636f6d"),
		"http://www.example.com",
		[]interface{}{"http://www.example.com"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	}, // 32: URI
	// primitives
	{
		hexDecode("f4"),
		false,
		[]interface{}{false},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeString, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f5"),
		true,
		[]interface{}{true},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeString, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f6"),
		nil,
		[]interface{}{[]byte(nil), []int(nil), []string(nil), map[string]int(nil)},
		[]reflect.Type{},
	},
	{
		hexDecode("f7"),
		nil,
		[]interface{}{[]byte(nil), []int(nil), []string(nil), map[string]int(nil)},
		[]reflect.Type{},
	},
	{
		hexDecode("f0"),
		uint64(16),
		[]interface{}{uint8(16), uint16(16), uint32(16), uint64(16), uint(16), int8(16), int16(16), int32(16), int64(16), int(16)},
		[]reflect.Type{typeFloat32, typeFloat64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f818"),
		uint64(24),
		[]interface{}{uint8(24), uint16(24), uint32(24), uint64(24), uint(24), int8(24), int16(24), int32(24), int64(24), int(24)},
		[]reflect.Type{typeFloat32, typeFloat64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f8ff"),
		uint64(255),
		[]interface{}{uint8(255), uint16(255), uint32(255), uint64(255), uint(255), int16(255), int32(255), int64(255), int(255)},
		[]reflect.Type{typeFloat32, typeFloat64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
}

// CBOR test data are from https://tools.ietf.org/html/rfc7049#appendix-A.
var unmarshalFloatTests = []unmarshalFloatTest{
	// float16
	{
		hexDecode("f90000"),
		float64(0.0),
		float64(0.0),
		[]interface{}{float32(0.0), float64(0.0)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f98000"),
		float64(-0.0),
		float64(0.0),
		[]interface{}{float32(-0.0), float64(-0.0)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f93c00"),
		float64(1.0),
		float64(0.0),
		[]interface{}{float32(1.0), float64(1.0)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f93e00"),
		float64(1.5),
		float64(0.00001),
		[]interface{}{float32(1.5), float64(1.5)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f97bff"),
		float64(65504.0),
		float64(0.0),
		[]interface{}{float32(65504.0), float64(65504.0)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f90001"),
		float64(5.960464477539063e-08),
		float64(0.00001),
		[]interface{}{float32(5.960464477539063e-08), float64(5.960464477539063e-08)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f90400"),
		float64(6.103515625e-05),
		float64(0.00001),
		[]interface{}{float32(6.103515625e-05), float64(6.103515625e-05)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f9c400"),
		float64(-4.0),
		float64(0.0),
		[]interface{}{float32(-4.0), float64(-4.0)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f97c00"),
		math.Inf(1),
		float64(0),
		[]interface{}{float32(math.Float32frombits(0x7f800000)), float64(math.Inf(1))},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f97e00"),
		math.NaN(),
		float64(0),
		[]interface{}{float32(math.Float32frombits(0x7fc00000)), float64(math.NaN())},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("f9fc00"),
		math.Inf(-1),
		float64(0),
		[]interface{}{float32(math.Float32frombits(0xff800000)), float64(math.Inf(-1))},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	// float32
	{
		hexDecode("fa47c35000"),
		float64(100000.0),
		float64(0.0),
		[]interface{}{float32(100000.0), float64(100000.0)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("fa7f7fffff"),
		float64(3.4028234663852886e+38),
		float64(0.00001),
		[]interface{}{float32(3.4028234663852886e+38), float64(3.4028234663852886e+38)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("fa7f800000"),
		math.Inf(1),
		float64(0),
		[]interface{}{float32(math.Float32frombits(0x7f800000)), float64(math.Inf(1))},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("fa7fc00000"),
		math.NaN(),
		float64(0),
		[]interface{}{float32(math.Float32frombits(0x7fc00000)), float64(math.NaN())},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt}},
	{
		hexDecode("faff800000"),
		math.Inf(-1),
		float64(0),
		[]interface{}{float32(math.Float32frombits(0xff800000)), float64(math.Inf(-1))},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	// float64
	{
		hexDecode("fb3ff199999999999a"),
		float64(1.1),
		float64(0.00001),
		[]interface{}{float32(1.1), float64(1.1)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("fb7e37e43c8800759c"),
		float64(1.0e+300),
		float64(0.00001),
		[]interface{}{float64(1.0e+300)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("fbc010666666666666"),
		float64(-4.1),
		float64(0.00001),
		[]interface{}{float32(-4.1), float64(-4.1)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("fb7ff0000000000000"),
		math.Inf(1),
		float64(0),
		[]interface{}{float32(math.Float32frombits(0x7f800000)), float64(math.Inf(1))},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("fb7ff8000000000000"),
		math.NaN(),
		float64(0),
		[]interface{}{float32(math.Float32frombits(0x7fc00000)), float64(math.NaN())},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("fbfff0000000000000"),
		math.Inf(-1),
		float64(0),
		[]interface{}{float32(math.Float32frombits(0xff800000)), float64(math.Inf(-1))},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
	},
}

func hexDecode(s string) []byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func TestUnmarshal(t *testing.T) {
	for _, tc := range unmarshalTests {
		// Test unmarshalling CBOR into empty interface.
		var v interface{}
		if err := cbor.Unmarshal(tc.cborData, &v); err != nil {
			t.Errorf("Unmarshal(0x%0x) returns error %v", tc.cborData, err)
		} else if !reflect.DeepEqual(v, tc.emptyInterfaceValue) {
			t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", tc.cborData, v, v, tc.emptyInterfaceValue, tc.emptyInterfaceValue)
		}
		// Test unmarshalling CBOR into compatible data types.
		for _, value := range tc.values {
			v := reflect.New(reflect.TypeOf(value))
			vPtr := v.Interface()
			if err := cbor.Unmarshal(tc.cborData, vPtr); err != nil {
				t.Errorf("Unmarshal(0x%0x) returns error %v", tc.cborData, err)
			} else if !reflect.DeepEqual(v.Elem().Interface(), value) {
				t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", tc.cborData, v.Elem().Interface(), v.Elem().Interface(), value, value)
			}
		}
		// Test unmarshalling CBOR into incompatible data types.
		for _, typ := range tc.wrongTypes {
			v := reflect.New(typ)
			vPtr := v.Interface()
			if err := cbor.Unmarshal(tc.cborData, vPtr); err == nil {
				t.Errorf("Unmarshal(0x%0x) returns %v (%T), want (*cbor.UnmarshalTypeError)", tc.cborData, v.Elem().Interface(), v.Elem().Interface())
			} else if _, ok := err.(*cbor.UnmarshalTypeError); !ok {
				t.Errorf("Unmarshal(0x%0x) returns wrong error %T, want (*cbor.UnmarshalTypeError)", tc.cborData, err)
			} else if !strings.Contains(err.Error(), "cannot unmarshal") {
				t.Errorf("Unmarshal(0x%0x) returns error %s, want error containing %q", tc.cborData, err.Error(), "cannot unmarshal")
			}
		}
	}
}

func TestUnmarshalFloat(t *testing.T) {
	for _, tc := range unmarshalFloatTests {
		// Test unmarshalling CBOR into empty interface.
		var v interface{}
		if err := cbor.Unmarshal(tc.cborData, &v); err != nil {
			t.Errorf("Unmarshal(0x%0x) returns error %v", tc.cborData, err)
		} else {
			if f, ok := v.(float64); !ok {
				t.Errorf("Unmarshal(0x%0x) returns value of type %T, want float64", tc.cborData, f)
			} else {
				if math.Abs(f-tc.emptyInterfaceValue.(float64)) > tc.diff {
					t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", tc.cborData, v, v, tc.emptyInterfaceValue, tc.emptyInterfaceValue)
				}
			}
		}
		// Test unmarshalling CBOR into compatible data types.
		for _, value := range tc.values {
			v := reflect.New(reflect.TypeOf(value))
			vPtr := v.Interface()
			if err := cbor.Unmarshal(tc.cborData, vPtr); err != nil {
				t.Errorf("Unmarshal(0x%0x) returns error %v", tc.cborData, err)
			} else {
				switch reflect.TypeOf(value).Kind() {
				case reflect.Float32:
					f := v.Elem().Interface().(float32)
					diff := f - value.(float32)
					if math.Abs(float64(diff)) > tc.diff {
						t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", tc.cborData, v.Elem().Interface(), v.Elem().Interface(), value, value)
					}
				case reflect.Float64:
					f := v.Elem().Interface().(float64)
					diff := f - value.(float64)
					if math.Abs(float64(diff)) > tc.diff {
						t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", tc.cborData, v.Elem().Interface(), v.Elem().Interface(), value, value)
					}
				}
			}
		}
		// Test unmarshalling CBOR into incompatible data types.
		for _, typ := range tc.wrongTypes {
			v := reflect.New(typ)
			vPtr := v.Interface()
			if err := cbor.Unmarshal(tc.cborData, vPtr); err == nil {
				t.Errorf("Unmarshal(0x%0x) returns %v (%T), want (*cbor.UnmarshalTypeError)", tc.cborData, v.Elem().Interface(), v.Elem().Interface())
			} else if _, ok := err.(*cbor.UnmarshalTypeError); !ok {
				t.Errorf("Unmarshal(0x%0x) returns wrong error %T, want (*cbor.UnmarshalTypeError)", tc.cborData, err)
			} else if !strings.Contains(err.Error(), "cannot unmarshal") {
				t.Errorf("Unmarshal(0x%0x) returns error %s, want error containing %q", tc.cborData, err.Error(), "cannot unmarshal")
			}
		}
	}
}

func TestUnmarshalIntoPointer(t *testing.T) {
	cborDataNil := []byte{0xf6}                                                                            // nil
	cborDataInt := []byte{0x18, 0x18}                                                                      // 24
	cborDataString := []byte{0x7f, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x67, 0xff} // "streaming"

	var p1 *int
	var p2 *string

	var i int
	pi := &i
	ppi := &pi

	var s string
	ps := &s
	pps := &ps

	// Unmarshal CBOR nil into a pointer.
	if err := cbor.Unmarshal(cborDataNil, &p1); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborDataNil, err)
	} else if p1 != nil {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want nil", cborDataNil, p1, p1)
	}

	// Unmarshal CBOR integer into a non-nil pointer.
	if err := cbor.Unmarshal(cborDataInt, &ppi); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborDataNil, err)
	} else if i != 24 {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want 24", cborDataNil, i, i)
	}

	// Unmarshal CBOR integer into a nil pointer.
	if err := cbor.Unmarshal(cborDataInt, &p1); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborDataNil, err)
	} else if *p1 != 24 {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want 24", cborDataNil, *pi, pi)
	}

	// Unmarshal CBOR string into a non-nil pointer.
	if err := cbor.Unmarshal(cborDataString, &pps); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborDataNil, err)
	} else if s != "streaming" {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want \"streaming\"", cborDataNil, s, s)
	}

	// Unmarshal CBOR string into a nil pointer.
	if err := cbor.Unmarshal(cborDataString, &p2); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborDataNil, err)
	} else if *p2 != "streaming" {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want \"streaming\"", cborDataNil, *p2, p2)
	}
}

func TestUnmarshalIntoArray(t *testing.T) {
	cborData := hexDecode("83010203") // []int{1, 2, 3}

	// Unmarshal CBOR array into Go array.
	var arr1 [3]int
	if err := cbor.Unmarshal(cborData, &arr1); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborData, err)
	} else if arr1 != [3]int{1, 2, 3} {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want [3]int{1, 2, 3}", cborData, arr1, arr1)
	}

	// Unmarshal CBOR array into Go array with more elements.
	var arr2 [5]int
	if err := cbor.Unmarshal(cborData, &arr2); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborData, err)
	} else if arr2 != [5]int{1, 2, 3, 0, 0} {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want [5]int{1, 2, 3, 0, 0}", cborData, arr2, arr2)
	}

	// Unmarshal CBOR array into Go array with less elements.
	var arr3 [1]int
	if err := cbor.Unmarshal(cborData, &arr3); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborData, err)
	} else if arr3 != [1]int{1} {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want [0]int{1}", cborData, arr3, arr3)
	}
}

func TestUnmarshalNil(t *testing.T) {
	cborData := [][]byte{hexDecode("f6"), hexDecode("f7")} // null, undefined
	bSlice := []byte{1, 2, 3}
	strSlice := []string{"hello", "world"}
	m := map[string]bool{"hello": true, "goodbye": false}
	nilValuesAfterUnmarshal := []interface{}{bSlice, strSlice, m}

	for _, data := range cborData {
		for _, v := range nilValuesAfterUnmarshal {
			if err := cbor.Unmarshal(data, &v); err != nil {
				t.Errorf("Unmarshal(0x%0x) returns error %v", data, err)
			} else if v != nil {
				t.Errorf("Unmarshal(0x%0x) = %v (%T), want nil", data, v, v)
			}
		}
	}

	for _, data := range cborData {
		i := 10
		if err := cbor.Unmarshal(data, &i); err != nil {
			t.Errorf("Unmarshal(0x%0x) returns error %v", data, err)
		} else if i != 10 {
			t.Errorf("Unmarshal(0x%0x) = %v (%T), want 10", data, i, i)
		}
		f := 1.23
		if err := cbor.Unmarshal(data, &f); err != nil {
			t.Errorf("Unmarshal(0x%0x) returns error %v", data, err)
		} else if f != 1.23 {
			t.Errorf("Unmarshal(0x%0x) = %v (%T), want 1.23", data, f, f)
		}
		s := "hello"
		if err := cbor.Unmarshal(data, &s); err != nil {
			t.Errorf("Unmarshal(0x%0x) returns error %v", data, err)
		} else if s != "hello" {
			t.Errorf("Unmarshal(0x%0x) = %v (%T), want \"hello\"", data, s, s)
		}
		b := true
		if err := cbor.Unmarshal(data, &t); err != nil {
			t.Errorf("Unmarshal(0x%0x) returns error %v", data, err)
		} else if b != true {
			t.Errorf("Unmarshal(0x%0x) = %v (%T), want true", data, b, b)
		}
	}
}

var invalidUnmarshalTests = []struct {
	name         string
	v            interface{}
	wantErrorMsg string
}{
	{"unmarshal into nil interface{}", nil, "cbor: Unmarshal(nil)"},
	{"unmarshal into non-pointer value", 5, "cbor: Unmarshal(non-pointer int)"},
	{"unmarshal into nil pointer", (*int)(nil), "cbor: Unmarshal(nil *int)"},
}

func TestInvalidUnmarshal(t *testing.T) {
	cborData := []byte{0x00}

	for _, tc := range invalidUnmarshalTests {
		t.Run(tc.name, func(t *testing.T) {
			err := cbor.Unmarshal(cborData, tc.v)
			if err == nil {
				t.Errorf("Unmarshal(0x%0x, %v) expecting error, got nil", cborData, tc.v)
			} else if _, ok := err.(*cbor.InvalidUnmarshalError); !ok {
				t.Errorf("Unmarshal(0x%0x, %v) error type %T, want *cbor.InvalidUnmarshalError", cborData, tc.v, err)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Unmarshal(0x%0x, %v) error %s, want %s", cborData, tc.v, err, tc.wantErrorMsg)
			}
		})
	}
}

var invalidCBORUnmarshalTests = []struct {
	name                 string
	cborData             []byte
	wantErrorMsg         string
	errorMsgPartialMatch bool
}{
	{"data is nil", []byte(nil), "EOF", false},
	{"data is empty", []byte{}, "EOF", false},
	{"incomplete header, want 2 bytes", []byte{0x18}, "unexpected EOF", false},
	{"incomplete header, want 3 bytes", []byte{0x19, 0x03}, "unexpected EOF", false},
	{"incomplete header, want 5 bytes", []byte{0x1a, 0x00, 0x0f, 0x42}, "unexpected EOF", false},
	{"incomplete header, want 9 bytes", []byte{0x1b, 0x00, 0x00, 0x00, 0xe8, 0xd4, 0xa5, 0x10}, "unexpected EOF", false},
	{"data type and additional information mismatch", []byte{0x1c}, "cbor: invalid additional information", true},
	{"data type and additional information mismatch", []byte{0x3f}, "cbor: invalid additional information", true},
	{"unexpected \"break\" code", []byte{0xff}, "cbor: unexpected \"break\" code", false},
	{"byte string: reach EOF before completing payload", []byte{0x48, 0x00, 0x01, 0x02, 0x03}, "unexpected EOF", false},
	{"array: reach EOF before completing payload", []byte{0x88, 0x00, 0x01, 0x02, 0x03}, "unexpected EOF", false},
	{"map: reach EOF before completing payload", []byte{0xa8, 0x00, 0x01, 0x02, 0x03}, "unexpected EOF", false},
	{"array: invalid element", []byte{0x81, 0x1f}, "cbor: invalid additional information", true},
	{"map: invalid element", []byte{0xa1, 0x00, 0x1f}, "cbor: invalid additional information", true},
	{"tag: no tagged data item", []byte{0xc0}, "unexpected EOF", false},
	{"indefinite-length byte string: element type is not byte string", []byte{0x5f, 0x42, 0x01, 0x02, 0x62, 0x61, 0x62}, "wrong element type", true},
	{"indefinite-length array: no \"break\" code", []byte{0x9f, 0x01, 0x02, 0x03, 0x04, 0x05}, "unexpected EOF", false},
	{"indefinite-length array: invalid element", []byte{0x9f, 0x1f}, "cbor: invalid additional information", true},
	{"indefinite-length array: incomplete element", []byte{0x9f, 0x19, 0x03}, "unexpected EOF", false},
	{"indefinite-length map: no \"break\" code", []byte{0xbf, 0x01}, "unexpected EOF", false},
	{"indefinite-length map: read \"break\" code before completing key-value pair", []byte{0xbf, 0x01, 0xff}, "cbor: unexpected \"break\" code", false},
	{"indefinite-length map: invalid element", []byte{0xbf, 0x01, 0x1f}, "cbor: invalid additional information", true},
	{"text string: invalid UTF-8 string", []byte{0x61, 0xfe}, "cbor: invalid UTF-8 string", false},
}

func TestInvalidCBORUnmarshal(t *testing.T) {
	var i interface{}
	for _, tc := range invalidCBORUnmarshalTests {
		t.Run(tc.name, func(t *testing.T) {
			err := cbor.Unmarshal(tc.cborData, &i)
			if err == nil {
				t.Errorf("Unmarshal(0x%0x) expecting error, got nil", tc.cborData)
			} else if !tc.errorMsgPartialMatch && err.Error() != tc.wantErrorMsg {
				t.Errorf("Unmarshal(0x%0x) error %s, want %s", tc.cborData, err, tc.wantErrorMsg)
			} else if tc.errorMsgPartialMatch && !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Unmarshal(0x%0x) error %s, want %s", tc.cborData, err, tc.wantErrorMsg)
			}
		})
	}
}

func TestUnmarshalStruct(t *testing.T) {
	want := outer{
		IntField:          123,
		FloatField:        100000.0,
		BoolField:         true,
		StringField:       "test",
		ByteStringField:   []byte{1, 3, 5},
		ArrayField:        []string{"hello", "world"},
		MapField:          map[string]bool{"morning": true, "afternoon": false},
		NestedStructField: &inner{X: 1000, Y: 1000000},
		unexportedField:   0,
	}

	tests := []struct {
		name     string
		cborData []byte
		want     interface{}
	}{
		{"case-insensitive field name match", hexDecode("a868696e746669656c64187b6a666c6f61746669656c64fa47c3500069626f6f6c6669656c64f56b537472696e674669656c6464746573746f42797465537472696e674669656c64430103056a41727261794669656c64826568656c6c6f65776f726c64684d61704669656c64a2676d6f726e696e67f56961667465726e6f6f6ef4714e65737465645374727563744669656c64a261581903e861591a000f4240"), want},
		{"exact field name match", hexDecode("a868496e744669656c64187b6a466c6f61744669656c64fa47c3500069426f6f6c4669656c64f56b537472696e674669656c6464746573746f42797465537472696e674669656c64430103056a41727261794669656c64826568656c6c6f65776f726c64684d61704669656c64a2676d6f726e696e67f56961667465726e6f6f6ef4714e65737465645374727563744669656c64a261581903e861591a000f4240"), want},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var v outer
			if err := cbor.Unmarshal(tc.cborData, &v); err != nil {
				t.Errorf("Unmarshal(0x%0x) returns error %v", tc.cborData, err)
			} else if !reflect.DeepEqual(v, want) {
				t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", tc.cborData, v, v, want, want)
			}
		})
	}
}

func TestUnmarshalStructError1(t *testing.T) {
	type outer2 struct {
		IntField          int
		FloatField        float32
		BoolField         bool
		StringField       string
		ByteStringField   []byte
		ArrayField        []int // wrong type
		MapField          map[string]bool
		NestedStructField map[int]string // wrong type
		unexportedField   int64
	}
	want := outer2{
		IntField:          123,
		FloatField:        100000.0,
		BoolField:         true,
		StringField:       "test",
		ByteStringField:   []byte{1, 3, 5},
		ArrayField:        []int{0, 0},
		MapField:          map[string]bool{"morning": true, "afternoon": false},
		NestedStructField: map[int]string{},
		unexportedField:   0,
	}

	cborData := hexDecode("a868496e744669656c64187b6a466c6f61744669656c64fa47c3500069426f6f6c4669656c64f56b537472696e674669656c6464746573746f42797465537472696e674669656c64430103056a41727261794669656c64826568656c6c6f65776f726c64684d61704669656c64a2676d6f726e696e67f56961667465726e6f6f6ef4714e65737465645374727563744669656c64a261581903e861591a000f4240")

	var v outer2
	if err := cbor.Unmarshal(cborData, &v); err == nil {
		t.Errorf("Unmarshal(0x%0x) doesn't return an error", cborData)
	} else {
		if typeError, ok := err.(*cbor.UnmarshalTypeError); !ok {
			t.Errorf("Unmarshal(0x%0x) returns wrong type of error %T, want (*cbor.UnmarshalTypeError)", cborData, err)
		} else {
			if typeError.Value != "UTF-8 text string" {
				t.Errorf("Unmarshal(0x%0x) (*cbor.UnmarshalTypeError).Value %s, want %s", cborData, typeError.Value, "UTF-8 text string")
			}
			if typeError.Type != typeInt {
				t.Errorf("Unmarshal(0x%0x) (*cbor.UnmarshalTypeError).Type %s, want %s", cborData, typeError.Type.String(), typeInt.String())
			}
			if typeError.Struct != "cbor_test.outer2" {
				t.Errorf("Unmarshal(0x%0x) (*cbor.UnmarshalTypeError).Struct %s, want %s", cborData, typeError.Struct, "cbor_test.outer2")
			}
			if typeError.Field != "ArrayField" {
				t.Errorf("Unmarshal(0x%0x) (*cbor.UnmarshalTypeError).Field %s, want %s", cborData, typeError.Field, "ArrayField")
			}
			if !strings.Contains(err.Error(), "cannot unmarshal UTF-8 text string into Go struct field") {
				t.Errorf("Unmarshal(0x%0x) returns error %s, want error containing %q", cborData, err.Error(), "cannot unmarshal UTF-8 text string into Go struct field")
			}
		}
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", cborData, v, v, want, want)
	}
}

func TestUnmarshalStructError2(t *testing.T) {
	// Unmarshal map key of integer type into struct
	type strc struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
	}
	want := strc{
		A: "A",
	}
	cborData := hexDecode("a2010261616141") // {1:2, "a":"A"}

	var v strc
	if err := cbor.Unmarshal(cborData, &v); err == nil {
		t.Errorf("Unmarshal(0x%0x) doesn't return an error", cborData)
	} else {
		if typeError, ok := err.(*cbor.UnmarshalTypeError); !ok {
			t.Errorf("Unmarshal(0x%0x) returns wrong type of error %T, want (*cbor.UnmarshalTypeError)", cborData, err)
		} else {
			if typeError.Value != "positive integer" {
				t.Errorf("Unmarshal(0x%0x) (*cbor.UnmarshalTypeError).Value %s, want %s", cborData, typeError.Value, "positive integer")
			}
			if typeError.Type != typeString {
				t.Errorf("Unmarshal(0x%0x) (*cbor.UnmarshalTypeError).Type %s, want %s", cborData, typeError.Type, typeString)
			}
			if !strings.Contains(err.Error(), "cannot unmarshal positive integer into Go value of type string") {
				t.Errorf("Unmarshal(0x%0x) returns error %s, want error containing %q", cborData, err.Error(), "cannot unmarshal positive integer into Go value of type string")
			}
		}
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", cborData, v, v, want, want)
	}
}

func TestUnmarshalPrefilledArray(t *testing.T) {
	prefilledArr := []int{1, 2, 3, 4, 5}
	want := []int{10, 11, 3, 4, 5}
	cborData := hexDecode("820a0b") // []int{10, 11}
	if err := cbor.Unmarshal(cborData, &prefilledArr); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborData, err)
	}
	if len(prefilledArr) != 2 || cap(prefilledArr) != 5 {
		t.Errorf("Unmarshal(0x%0x) = %v (len %d, cap %d), want len == 2, cap == 5", cborData, prefilledArr, len(prefilledArr), cap(prefilledArr))
	}
	if !reflect.DeepEqual(prefilledArr[:cap(prefilledArr)], want) {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", cborData, prefilledArr, prefilledArr, want, want)
	}

	cborData = hexDecode("80") // empty array
	if err := cbor.Unmarshal(cborData, &prefilledArr); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborData, err)
	}
	if len(prefilledArr) != 0 || cap(prefilledArr) != 0 {
		t.Errorf("Unmarshal(0x%0x) = %v (len %d, cap %d), want len == 0, cap == 0", cborData, prefilledArr, len(prefilledArr), cap(prefilledArr))
	}
}

func TestUnmarshalPrefilledMap(t *testing.T) {
	prefilledMap := map[string]string{"key": "value", "a": "1"}
	want := map[string]string{"key": "value", "a": "A", "b": "B", "c": "C", "d": "D", "e": "E"}
	cborData := hexDecode("a56161614161626142616361436164614461656145") // map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"}
	if err := cbor.Unmarshal(cborData, &prefilledMap); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborData, err)
	}
	if !reflect.DeepEqual(prefilledMap, want) {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", cborData, prefilledMap, prefilledMap, want, want)
	}

	prefilledMap = map[string]string{"key": "value"}
	want = map[string]string{"key": "value"}
	cborData = hexDecode("a0") // map[string]string{}
	if err := cbor.Unmarshal(cborData, &prefilledMap); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborData, err)
	}
	if !reflect.DeepEqual(prefilledMap, want) {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", cborData, prefilledMap, prefilledMap, want, want)
	}
}

func TestUnmarshalPrefilledStruct(t *testing.T) {
	type s struct {
		a int
		B []int
		C bool
	}
	prefilledStruct := s{a: 100, B: []int{200, 300, 400, 500}, C: true}
	want := s{a: 100, B: []int{2, 3}, C: true}
	cborData := hexDecode("a26161016162820203") // map[string]interface{} {"a": 1, "b": []int{2, 3}}
	if err := cbor.Unmarshal(cborData, &prefilledStruct); err != nil {
		t.Errorf("Unmarshal(0x%0x) returns error %v", cborData, err)
	}
	if !reflect.DeepEqual(prefilledStruct, want) {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", cborData, prefilledStruct, prefilledStruct, want, want)
	}
	if len(prefilledStruct.B) != 2 || cap(prefilledStruct.B) != 4 {
		t.Errorf("Unmarshal(0x%0x) = %v (len %d, cap %d), want len == 2, cap == 5", cborData, prefilledStruct.B, len(prefilledStruct.B), cap(prefilledStruct.B))
	}
	if !reflect.DeepEqual(prefilledStruct.B[:cap(prefilledStruct.B)], []int{2, 3, 400, 500}) {
		t.Errorf("Unmarshal(0x%0x) = %v (%T), want %v (%T)", cborData, prefilledStruct.B, prefilledStruct.B, []int{2, 3, 400, 500}, []int{2, 3, 400, 500})
	}
}

func TestValid(t *testing.T) {
	var buf bytes.Buffer
	for _, t := range marshalTests {
		buf.Write(t.cborData)
	}
	cborData := buf.Bytes()
	var err error
	for i := 0; i < len(marshalTests); i++ {
		if cborData, err = cbor.Valid(cborData); err != nil {
			t.Errorf("Valid() returns error %s", err)
		}
	}
	if len(cborData) != 0 {
		t.Errorf("Valid() returns leftover data 0x%x, want none", cborData)
	}
}

func TestFuzzCrash1(t *testing.T) {
	// Crash1: string/slice/map length in uint64 cast to int causes integer overflow.
	hexData := "bbcf30303030303030cfd697829782"
	data := hexDecode(hexData)
	var intf interface{}
	wantErrorMsg := "is too large"
	if err := cbor.Unmarshal(data, &intf); err == nil {
		t.Errorf("Unmarshal(0x%02x) returns no error, want error containing substring %s", data, wantErrorMsg)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%02x) returns error %s, want error containing substring %s", data, err, wantErrorMsg)
	}
}

func TestFuzzCrash2(t *testing.T) {
	// Crash2: map key (slice or map) is unhashable.
	testData := []string{
		"b0303030308030303030303030303030303030303030303030303030303030303030",
		"b030303030303030a1413030303030303030303030303030306230303030303030303030303030",
		"b03030303030303030403030303030303030303030306230303030303030303030303030",
		"8f303030a7303a30303030a2303030303030303030303030303030303030303030303030303030",
		"bf30bf8030ffff",
		"bf30a1a030ff",
		"8f3030a730304430303030303030303030303030303030303030303030303030303030",
		"8f303030a730303030303030308530303030303030303030303030303030303030303030",
		"bfb0303030303030303030303030303030303030303030303030303030303030303030ff",
	}
	wantErrorMsg := "invalid map key type"
	for _, hexData := range testData {
		data := hexDecode(hexData)
		var intf interface{}
		if err := cbor.Unmarshal(data, &intf); err == nil {
			t.Errorf("Unmarshal(0x%02x) returns no error, want error containing substring %s", data, wantErrorMsg)
		} else if !strings.Contains(err.Error(), wantErrorMsg) {
			t.Errorf("Unmarshal(0x%02x) returns error %s, want error containing substring %s", data, err, wantErrorMsg)
		}
	}
}

func TestFuzzCrash3(t *testing.T) {
	// Crash3: encoding nil as collection (slice, array, or map) element.
	hexData := "b0303030303030303030303030303030303038303030faffff30303030303030303030303030"
	data := hexDecode(hexData)
	var intf interface{}
	if err := cbor.Unmarshal(data, &intf); err != nil {
		t.Fatalf("Unmarshal(0x%02x) returns error %s\n", data, err)
	}
	if _, err := cbor.Marshal(intf, cbor.EncOptions{Canonical: true}); err != nil {
		t.Errorf("Marshal(%v) returns error %s", intf, err)
	}
}

func TestFuzzCrash4(t *testing.T) {
	// Crash3: parsing nil/undefined as collection (slice, array, or map) element.
	data := []byte("\xbfѣ\x88\xf70000000000000\xff")
	var intf interface{}
	wantErrorMsg := "invalid map key type"
	if err := cbor.Unmarshal(data, &intf); err == nil {
		t.Errorf("Unmarshal(0x%02x) returns no error, want error containing substring %s", data, wantErrorMsg)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%02x) returns error %s, want error containing substring %s", data, err, wantErrorMsg)
	}
}
