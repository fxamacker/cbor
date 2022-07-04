// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	typeBool            = reflect.TypeOf(true)
	typeUint8           = reflect.TypeOf(uint8(0))
	typeUint16          = reflect.TypeOf(uint16(0))
	typeUint32          = reflect.TypeOf(uint32(0))
	typeUint64          = reflect.TypeOf(uint64(0))
	typeInt8            = reflect.TypeOf(int8(0))
	typeInt16           = reflect.TypeOf(int16(0))
	typeInt32           = reflect.TypeOf(int32(0))
	typeInt64           = reflect.TypeOf(int64(0))
	typeFloat32         = reflect.TypeOf(float32(0))
	typeFloat64         = reflect.TypeOf(float64(0))
	typeString          = reflect.TypeOf("")
	typeByteSlice       = reflect.TypeOf([]byte(nil))
	typeByteArray       = reflect.TypeOf([5]byte{})
	typeIntSlice        = reflect.TypeOf([]int{})
	typeStringSlice     = reflect.TypeOf([]string{})
	typeMapStringInt    = reflect.TypeOf(map[string]int{})
	typeMapStringString = reflect.TypeOf(map[string]string{})
	typeMapStringIntf   = reflect.TypeOf(map[string]interface{}{})
)

type unmarshalTest struct {
	cborData            []byte
	emptyInterfaceValue interface{}
	values              []interface{}
	wrongTypes          []reflect.Type
}

var unmarshalTests = []unmarshalTest{
	// CBOR test data are from https://tools.ietf.org/html/rfc7049#appendix-A.
	// positive integer
	{
		hexDecode("00"),
		uint64(0),
		[]interface{}{uint8(0), uint16(0), uint32(0), uint64(0), uint(0), int8(0), int16(0), int32(0), int64(0), int(0), float32(0), float64(0), bigIntOrPanic("0")},
		[]reflect.Type{typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("01"),
		uint64(1),
		[]interface{}{uint8(1), uint16(1), uint32(1), uint64(1), uint(1), int8(1), int16(1), int32(1), int64(1), int(1), float32(1), float64(1), bigIntOrPanic("1")},
		[]reflect.Type{typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("0a"),
		uint64(10),
		[]interface{}{uint8(10), uint16(10), uint32(10), uint64(10), uint(10), int8(10), int16(10), int32(10), int64(10), int(10), float32(10), float64(10), bigIntOrPanic("10")},
		[]reflect.Type{typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("17"),
		uint64(23),
		[]interface{}{uint8(23), uint16(23), uint32(23), uint64(23), uint(23), int8(23), int16(23), int32(23), int64(23), int(23), float32(23), float64(23), bigIntOrPanic("23")},
		[]reflect.Type{typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("1818"),
		uint64(24),
		[]interface{}{uint8(24), uint16(24), uint32(24), uint64(24), uint(24), int8(24), int16(24), int32(24), int64(24), int(24), float32(24), float64(24), bigIntOrPanic("24")},
		[]reflect.Type{typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("1819"),
		uint64(25),
		[]interface{}{uint8(25), uint16(25), uint32(25), uint64(25), uint(25), int8(25), int16(25), int32(25), int64(25), int(25), float32(25), float64(25), bigIntOrPanic("25")},
		[]reflect.Type{typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("1864"),
		uint64(100),
		[]interface{}{uint8(100), uint16(100), uint32(100), uint64(100), uint(100), int8(100), int16(100), int32(100), int64(100), int(100), float32(100), float64(100), bigIntOrPanic("100")},
		[]reflect.Type{typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("1903e8"),
		uint64(1000),
		[]interface{}{uint16(1000), uint32(1000), uint64(1000), uint(1000), int16(1000), int32(1000), int64(1000), int(1000), float32(1000), float64(1000), bigIntOrPanic("1000")},
		[]reflect.Type{typeUint8, typeInt8, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("1a000f4240"),
		uint64(1000000),
		[]interface{}{uint32(1000000), uint64(1000000), uint(1000000), int32(1000000), int64(1000000), int(1000000), float32(1000000), float64(1000000), bigIntOrPanic("1000000")},
		[]reflect.Type{typeUint8, typeUint16, typeInt8, typeInt16, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("1b000000e8d4a51000"),
		uint64(1000000000000),
		[]interface{}{uint64(1000000000000), uint(1000000000000), int64(1000000000000), int(1000000000000), float32(1000000000000), float64(1000000000000), bigIntOrPanic("1000000000000")},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeInt8, typeInt16, typeInt32, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("1bffffffffffffffff"),
		uint64(18446744073709551615),
		[]interface{}{uint64(18446744073709551615), uint(18446744073709551615), float32(18446744073709551615), float64(18446744073709551615), bigIntOrPanic("18446744073709551615")},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeInt8, typeInt16, typeInt32, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	// negative integer
	{
		hexDecode("20"),
		int64(-1),
		[]interface{}{int8(-1), int16(-1), int32(-1), int64(-1), int(-1), float32(-1), float64(-1), bigIntOrPanic("-1")},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("29"),
		int64(-10),
		[]interface{}{int8(-10), int16(-10), int32(-10), int64(-10), int(-10), float32(-10), float64(-10), bigIntOrPanic("-10")},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("3863"),
		int64(-100),
		[]interface{}{int8(-100), int16(-100), int32(-100), int64(-100), int(-100), float32(-100), float64(-100), bigIntOrPanic("-100")},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("3903e7"),
		int64(-1000),
		[]interface{}{int16(-1000), int32(-1000), int64(-1000), int(-1000), float32(-1000), float64(-1000), bigIntOrPanic("-1000")},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("3bffffffffffffffff"),
		bigIntOrPanic("-18446744073709551616"),
		[]interface{}{bigIntOrPanic("-18446744073709551616")},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	}, // CBOR value -18446744073709551616 overflows Go's int64, see TestNegIntOverflow
	// byte string
	{
		hexDecode("40"),
		[]byte{},
		[]interface{}{[]byte{}, [0]byte{}, [1]byte{0}, [5]byte{0, 0, 0, 0, 0}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("4401020304"),
		[]byte{1, 2, 3, 4},
		[]interface{}{[]byte{1, 2, 3, 4}, [0]byte{}, [1]byte{1}, [5]byte{1, 2, 3, 4, 0}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("5f42010243030405ff"),
		[]byte{1, 2, 3, 4, 5},
		[]interface{}{[]byte{1, 2, 3, 4, 5}, [0]byte{}, [1]byte{1}, [5]byte{1, 2, 3, 4, 5}, [6]byte{1, 2, 3, 4, 5, 0}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	// text string
	{
		hexDecode("60"),
		"",
		[]interface{}{""},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("6161"),
		"a",
		[]interface{}{"a"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("6449455446"),
		"IETF",
		[]interface{}{"IETF"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("62225c"),
		"\"\\",
		[]interface{}{"\"\\"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("62c3bc"),
		"ü",
		[]interface{}{"ü"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("63e6b0b4"),
		"水",
		[]interface{}{"水"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("64f0908591"),
		"𐅑",
		[]interface{}{"𐅑"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("7f657374726561646d696e67ff"),
		"streaming",
		[]interface{}{"streaming"},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	// array
	{
		hexDecode("80"),
		[]interface{}{},
		[]interface{}{[]interface{}{}, []byte{}, []string{}, []int{}, [0]int{}, [1]int{0}, [5]int{0}, []float32{}, []float64{}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("83010203"),
		[]interface{}{uint64(1), uint64(2), uint64(3)},
		[]interface{}{[]interface{}{uint64(1), uint64(2), uint64(3)}, []byte{1, 2, 3}, []int{1, 2, 3}, []uint{1, 2, 3}, [0]int{}, [1]int{1}, [3]int{1, 2, 3}, [5]int{1, 2, 3, 0, 0}, []float32{1, 2, 3}, []float64{1, 2, 3}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{}), typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("8301820203820405"),
		[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}},
		[]interface{}{[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}, [...]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{}), typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("83018202039f0405ff"),
		[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}},
		[]interface{}{[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}, [...]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{}), typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("83019f0203ff820405"),
		[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}},
		[]interface{}{[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}, [...]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{}), typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("98190102030405060708090a0b0c0d0e0f101112131415161718181819"),
		[]interface{}{uint64(1), uint64(2), uint64(3), uint64(4), uint64(5), uint64(6), uint64(7), uint64(8), uint64(9), uint64(10), uint64(11), uint64(12), uint64(13), uint64(14), uint64(15), uint64(16), uint64(17), uint64(18), uint64(19), uint64(20), uint64(21), uint64(22), uint64(23), uint64(24), uint64(25)},
		[]interface{}{
			[]interface{}{uint64(1), uint64(2), uint64(3), uint64(4), uint64(5), uint64(6), uint64(7), uint64(8), uint64(9), uint64(10), uint64(11), uint64(12), uint64(13), uint64(14), uint64(15), uint64(16), uint64(17), uint64(18), uint64(19), uint64(20), uint64(21), uint64(22), uint64(23), uint64(24), uint64(25)},
			[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[0]int{},
			[1]int{1},
			[...]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[30]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 0, 0, 0, 0, 0},
			[]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{}), typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("9fff"),
		[]interface{}{},
		[]interface{}{[]interface{}{}, []byte{}, []string{}, []int{}, [0]int{}, [1]int{0}, [5]int{0, 0, 0, 0, 0}, []float32{}, []float64{}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("9f018202039f0405ffff"),
		[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}},
		[]interface{}{[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}, [...]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{}), typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("9f01820203820405ff"),
		[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}},
		[]interface{}{[]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}, [...]interface{}{uint64(1), []interface{}{uint64(2), uint64(3)}, []interface{}{uint64(4), uint64(5)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{}), typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("9f0102030405060708090a0b0c0d0e0f101112131415161718181819ff"),
		[]interface{}{uint64(1), uint64(2), uint64(3), uint64(4), uint64(5), uint64(6), uint64(7), uint64(8), uint64(9), uint64(10), uint64(11), uint64(12), uint64(13), uint64(14), uint64(15), uint64(16), uint64(17), uint64(18), uint64(19), uint64(20), uint64(21), uint64(22), uint64(23), uint64(24), uint64(25)},
		[]interface{}{
			[]interface{}{uint64(1), uint64(2), uint64(3), uint64(4), uint64(5), uint64(6), uint64(7), uint64(8), uint64(9), uint64(10), uint64(11), uint64(12), uint64(13), uint64(14), uint64(15), uint64(16), uint64(17), uint64(18), uint64(19), uint64(20), uint64(21), uint64(22), uint64(23), uint64(24), uint64(25)},
			[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[0]int{},
			[1]int{1},
			[...]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[30]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 0, 0, 0, 0, 0},
			[]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
			[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{}), typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("826161a161626163"),
		[]interface{}{"a", map[interface{}]interface{}{"b": "c"}},
		[]interface{}{[]interface{}{"a", map[interface{}]interface{}{"b": "c"}}, [...]interface{}{"a", map[interface{}]interface{}{"b": "c"}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeByteArray, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{}), typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("826161bf61626163ff"),
		[]interface{}{"a", map[interface{}]interface{}{"b": "c"}},
		[]interface{}{[]interface{}{"a", map[interface{}]interface{}{"b": "c"}}, [...]interface{}{"a", map[interface{}]interface{}{"b": "c"}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeByteArray, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{}), typeTag, typeRawTag, typeBigInt},
	},
	// map
	{
		hexDecode("a0"),
		map[interface{}]interface{}{},
		[]interface{}{map[interface{}]interface{}{}, map[string]bool{}, map[string]int{}, map[int]string{}, map[int]bool{}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeByteArray, typeString, typeBool, typeIntSlice, typeTag, typeRawTag},
	},
	{
		hexDecode("a201020304"),
		map[interface{}]interface{}{uint64(1): uint64(2), uint64(3): uint64(4)},
		[]interface{}{map[interface{}]interface{}{uint64(1): uint64(2), uint64(3): uint64(4)}, map[uint]int{1: 2, 3: 4}, map[int]uint{1: 2, 3: 4}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeByteArray, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("a26161016162820203"),
		map[interface{}]interface{}{"a": uint64(1), "b": []interface{}{uint64(2), uint64(3)}},
		[]interface{}{map[interface{}]interface{}{"a": uint64(1), "b": []interface{}{uint64(2), uint64(3)}},
			map[string]interface{}{"a": uint64(1), "b": []interface{}{uint64(2), uint64(3)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeByteArray, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("a56161614161626142616361436164614461656145"),
		map[interface{}]interface{}{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"},
		[]interface{}{map[interface{}]interface{}{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"},
			map[string]interface{}{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"},
			map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeByteArray, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("bf61610161629f0203ffff"),
		map[interface{}]interface{}{"a": uint64(1), "b": []interface{}{uint64(2), uint64(3)}},
		[]interface{}{map[interface{}]interface{}{"a": uint64(1), "b": []interface{}{uint64(2), uint64(3)}},
			map[string]interface{}{"a": uint64(1), "b": []interface{}{uint64(2), uint64(3)}}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeByteArray, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("bf6346756ef563416d7421ff"),
		map[interface{}]interface{}{"Fun": true, "Amt": int64(-2)},
		[]interface{}{map[interface{}]interface{}{"Fun": true, "Amt": int64(-2)},
			map[string]interface{}{"Fun": true, "Amt": int64(-2)}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeByteArray, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	// tag
	{
		hexDecode("c074323031332d30332d32315432303a30343a30305a"),
		time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC), // 2013-03-21 20:04:00 +0000 UTC
		[]interface{}{"2013-03-21T20:04:00Z", time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC), Tag{0, "2013-03-21T20:04:00Z"}, RawTag{0, hexDecode("74323031332d30332d32315432303a30343a30305a")}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeBigInt},
	}, // 0: standard date/time
	{
		hexDecode("c11a514b67b0"),
		time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC), // 2013-03-21 20:04:00 +0000 UTC
		[]interface{}{uint32(1363896240), uint64(1363896240), int32(1363896240), int64(1363896240), float32(1363896240), float64(1363896240), time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC), Tag{1, uint64(1363896240)}, RawTag{1, hexDecode("1a514b67b0")}},
		[]reflect.Type{typeUint8, typeUint16, typeInt8, typeInt16, typeByteSlice, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt},
	}, // 1: epoch-based date/time
	{
		hexDecode("c249010000000000000000"),
		bigIntOrPanic("18446744073709551616"),
		[]interface{}{
			// Decode to byte slice
			[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			// Decode to array of various lengths
			[0]byte{},
			[1]byte{0x01},
			[3]byte{0x01, 0x00, 0x00},
			[...]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[10]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			// Decode to Tag and RawTag
			Tag{2, []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			RawTag{2, hexDecode("49010000000000000000")},
			// Decode to big.Int
			bigIntOrPanic("18446744073709551616"),
		},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	}, // 2: positive bignum: 18446744073709551616
	{
		hexDecode("c349010000000000000000"),
		bigIntOrPanic("-18446744073709551617"),
		[]interface{}{
			// Decode to byte slice
			[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			// Decode to array of various lengths
			[0]byte{},
			[1]byte{0x01},
			[3]byte{0x01, 0x00, 0x00},
			[...]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[10]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			// Decode to Tag and RawTag
			Tag{3, []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			RawTag{3, hexDecode("49010000000000000000")},
			// Decode to big.Int
			bigIntOrPanic("-18446744073709551617"),
		},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt},
	}, // 3: negative bignum: -18446744073709551617
	{
		hexDecode("c1fb41d452d9ec200000"),
		time.Date(2013, 3, 21, 20, 4, 0, 500000000, time.UTC), // 2013-03-21 20:04:00.5 +0000 UTC
		[]interface{}{float32(1363896240.5), float64(1363896240.5), time.Date(2013, 3, 21, 20, 4, 0, 500000000, time.UTC), Tag{1, float64(1363896240.5)}, RawTag{1, hexDecode("fb41d452d9ec200000")}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteSlice, typeByteArray, typeString, typeBool, typeIntSlice, typeMapStringInt, typeBigInt},
	}, // 1: epoch-based date/time
	{
		hexDecode("d74401020304"),
		Tag{23, []byte{0x01, 0x02, 0x03, 0x04}},
		[]interface{}{[]byte{0x01, 0x02, 0x03, 0x04}, [0]byte{}, [1]byte{0x01}, [3]byte{0x01, 0x02, 0x03}, [...]byte{0x01, 0x02, 0x03, 0x04}, [5]byte{0x01, 0x02, 0x03, 0x04, 0x00}, Tag{23, []byte{0x01, 0x02, 0x03, 0x04}}, RawTag{23, hexDecode("4401020304")}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt, typeBigInt},
	}, // 23: expected conversion to base16 encoding
	{
		hexDecode("d818456449455446"),
		Tag{24, []byte{0x64, 0x49, 0x45, 0x54, 0x46}},
		[]interface{}{[]byte{0x64, 0x49, 0x45, 0x54, 0x46}, [0]byte{}, [1]byte{0x64}, [3]byte{0x64, 0x49, 0x45}, [...]byte{0x64, 0x49, 0x45, 0x54, 0x46}, [6]byte{0x64, 0x49, 0x45, 0x54, 0x46, 0x00}, Tag{24, []byte{0x64, 0x49, 0x45, 0x54, 0x46}}, RawTag{24, hexDecode("456449455446")}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt, typeBigInt},
	}, // 24: encoded cborBytes data item
	{
		hexDecode("d82076687474703a2f2f7777772e6578616d706c652e636f6d"),
		Tag{32, "http://www.example.com"},
		[]interface{}{"http://www.example.com", Tag{32, "http://www.example.com"}, RawTag{32, hexDecode("76687474703a2f2f7777772e6578616d706c652e636f6d")}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeBigInt},
	}, // 32: URI
	// primitives
	{
		hexDecode("f4"),
		false,
		[]interface{}{false},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteArray, typeByteSlice, typeString, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("f5"),
		true,
		[]interface{}{true},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteArray, typeByteSlice, typeString, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("f6"),
		nil,
		[]interface{}{false, uint(0), uint8(0), uint16(0), uint32(0), uint64(0), int(0), int8(0), int16(0), int32(0), int64(0), float32(0.0), float64(0.0), "", []byte(nil), []int(nil), []string(nil), map[string]int(nil), time.Time{}, bigIntOrPanic("0"), Tag{}, RawTag{}},
		nil,
	},
	{
		hexDecode("f7"),
		nil,
		[]interface{}{false, uint(0), uint8(0), uint16(0), uint32(0), uint64(0), int(0), int8(0), int16(0), int32(0), int64(0), float32(0.0), float64(0.0), "", []byte(nil), []int(nil), []string(nil), map[string]int(nil), time.Time{}, bigIntOrPanic("0"), Tag{}, RawTag{}},
		nil,
	},
	{
		hexDecode("f0"),
		uint64(16),
		[]interface{}{uint8(16), uint16(16), uint32(16), uint64(16), uint(16), int8(16), int16(16), int32(16), int64(16), int(16), float32(16), float64(16), bigIntOrPanic("16")},
		[]reflect.Type{typeByteSlice, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	// This example is not well-formed because Simple value (with 5-bit value 24) must be >= 32.
	// See RFC 7049 section 2.3 for details, instead of the incorrect example in RFC 7049 Appendex A.
	// I reported an errata to RFC 7049 and Carsten Bormann confirmed at https://github.com/fxamacker/cbor/issues/46
	/*
		{
			hexDecode("f818"),
			uint64(24),
			[]interface{}{uint8(24), uint16(24), uint32(24), uint64(24), uint(24), int8(24), int16(24), int32(24), int64(24), int(24), float32(24), float64(24)},
			[]reflect.Type{typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt},
		},
	*/
	{
		hexDecode("f820"),
		uint64(32),
		[]interface{}{uint8(32), uint16(32), uint32(32), uint64(32), uint(32), int8(32), int16(32), int32(32), int64(32), int(32), float32(32), float64(32), bigIntOrPanic("32")},
		[]reflect.Type{typeByteSlice, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	{
		hexDecode("f8ff"),
		uint64(255),
		[]interface{}{uint8(255), uint16(255), uint32(255), uint64(255), uint(255), int16(255), int32(255), int64(255), int(255), float32(255), float64(255), bigIntOrPanic("255")},
		[]reflect.Type{typeByteSlice, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
	},
	// More testcases not covered by https://tools.ietf.org/html/rfc7049#appendix-A.
	{
		hexDecode("5fff"), // empty indefinite length byte string
		[]byte{},
		[]interface{}{[]byte{}, [0]byte{}, [1]byte{0x00}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("7fff"), // empty indefinite length text string
		"",
		[]interface{}{""},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeBool, typeByteArray, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
	},
	{
		hexDecode("bfff"), // empty indefinite length map
		map[interface{}]interface{}{},
		[]interface{}{map[interface{}]interface{}{}, map[string]bool{}, map[string]int{}, map[int]string{}, map[int]bool{}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeTag, typeRawTag},
	},
	// More test data with tags
	{
		hexDecode("c13a0177f2cf"), // 1969-03-21T20:04:00Z, tag 1 with negative integer as epoch time
		time.Date(1969, 3, 21, 20, 4, 0, 0, time.UTC),
		[]interface{}{int32(-24638160), int64(-24638160), int32(-24638160), int64(-24638160), float32(-24638160), float64(-24638160), time.Date(1969, 3, 21, 20, 4, 0, 0, time.UTC), Tag{1, int64(-24638160)}, RawTag{1, hexDecode("3a0177f2cf")}, bigIntOrPanic("-24638160")},
		[]reflect.Type{typeUint8, typeUint16, typeInt8, typeInt16, typeByteSlice, typeString, typeBool, typeByteArray, typeIntSlice, typeMapStringInt},
	},
	{
		hexDecode("d83dd183010203"), // 61(17([1, 2, 3])), nested tags 61 and 17
		Tag{61, Tag{17, []interface{}{uint64(1), uint64(2), uint64(3)}}},
		[]interface{}{[]interface{}{uint64(1), uint64(2), uint64(3)}, []byte{1, 2, 3}, [0]byte{}, [1]byte{1}, [3]byte{1, 2, 3}, [5]byte{1, 2, 3, 0, 0}, []int{1, 2, 3}, []uint{1, 2, 3}, [...]int{1, 2, 3}, []float32{1, 2, 3}, []float64{1, 2, 3}, Tag{61, Tag{17, []interface{}{uint64(1), uint64(2), uint64(3)}}}, RawTag{61, hexDecode("d183010203")}},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{})},
	},
}

type unmarshalFloatTest struct {
	cborData            []byte
	emptyInterfaceValue interface{}
	values              []interface{}
	wrongTypes          []reflect.Type
	equalityThreshold   float64 // Not used for +inf, -inf, and NaN.
}

// unmarshalFloatTests includes test values for float16, float32, and float64.
// Note: the function for float16 to float32 conversion was tested with all
// 65536 values, which is too many to include here.
var unmarshalFloatTests = []unmarshalFloatTest{
	// CBOR test data are from https://tools.ietf.org/html/rfc7049#appendix-A.
	// float16
	{
		hexDecode("f90000"),
		float64(0.0),
		[]interface{}{float32(0.0), float64(0.0)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("f98000"),
		float64(-0.0), //nolint:staticcheck // we know -0.0 is 0.0 in Go
		[]interface{}{float32(-0.0), float64(-0.0)}, //nolint:staticcheck // we know -0.0 is 0.0 in Go
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("f93c00"),
		float64(1.0),
		[]interface{}{float32(1.0), float64(1.0)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("f93e00"),
		float64(1.5),
		[]interface{}{float32(1.5), float64(1.5)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("f97bff"),
		float64(65504.0),
		[]interface{}{float32(65504.0), float64(65504.0)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("f90001"), // float16 subnormal value
		float64(5.960464477539063e-08),
		[]interface{}{float32(5.960464477539063e-08), float64(5.960464477539063e-08)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		1e-16,
	},
	{
		hexDecode("f90400"),
		float64(6.103515625e-05),
		[]interface{}{float32(6.103515625e-05), float64(6.103515625e-05)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		1e-16,
	},
	{
		hexDecode("f9c400"),
		float64(-4.0),
		[]interface{}{float32(-4.0), float64(-4.0)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("f97c00"),
		math.Inf(1),
		[]interface{}{math.Float32frombits(0x7f800000), math.Inf(1)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("f97e00"),
		math.NaN(),
		[]interface{}{math.Float32frombits(0x7fc00000), math.NaN()},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("f9fc00"),
		math.Inf(-1),
		[]interface{}{math.Float32frombits(0xff800000), math.Inf(-1)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	// float32
	{
		hexDecode("fa47c35000"),
		float64(100000.0),
		[]interface{}{float32(100000.0), float64(100000.0)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("fa7f7fffff"),
		float64(3.4028234663852886e+38),
		[]interface{}{float32(3.4028234663852886e+38), float64(3.4028234663852886e+38)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		1e-9,
	},
	{
		hexDecode("fa7f800000"),
		math.Inf(1),
		[]interface{}{math.Float32frombits(0x7f800000), math.Inf(1)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("fa7fc00000"),
		math.NaN(),
		[]interface{}{math.Float32frombits(0x7fc00000), math.NaN()},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("faff800000"),
		math.Inf(-1),
		[]interface{}{math.Float32frombits(0xff800000), math.Inf(-1)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	// float64
	{
		hexDecode("fb3ff199999999999a"),
		float64(1.1),
		[]interface{}{float32(1.1), float64(1.1)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		1e-9,
	},
	{
		hexDecode("fb7e37e43c8800759c"),
		float64(1.0e+300),
		[]interface{}{float64(1.0e+300)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		1e-9,
	},
	{
		hexDecode("fbc010666666666666"),
		float64(-4.1),
		[]interface{}{float32(-4.1), float64(-4.1)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		1e-9,
	},
	{
		hexDecode("fb7ff0000000000000"),
		math.Inf(1),
		[]interface{}{math.Float32frombits(0x7f800000), math.Inf(1)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("fb7ff8000000000000"),
		math.NaN(),
		[]interface{}{math.Float32frombits(0x7fc00000), math.NaN()},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("fbfff0000000000000"),
		math.Inf(-1),
		[]interface{}{math.Float32frombits(0xff800000), math.Inf(-1)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},

	// float16 test data from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
	{
		hexDecode("f903ff"),
		float64(0.000060976),
		[]interface{}{float32(0.000060976), float64(0.000060976)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		1e-9,
	},
	{
		hexDecode("f93bff"),
		float64(0.999511719),
		[]interface{}{float32(0.999511719), float64(0.999511719)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		1e-9,
	},
	{
		hexDecode("f93c01"),
		float64(1.000976563),
		[]interface{}{float32(1.000976563), float64(1.000976563)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		1e-9,
	},
	{
		hexDecode("f93555"),
		float64(0.333251953125),
		[]interface{}{float32(0.333251953125), float64(0.333251953125)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		1e-9,
	},
	// CBOR test data "canonNums" are from https://github.com/cbor-wg/cbor-test-vectors
	{
		hexDecode("f9bd00"),
		float64(-1.25),
		[]interface{}{float32(-1.25), float64(-1.25)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("f93e00"),
		float64(1.5),
		[]interface{}{float32(1.5), float64(1.5)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("fb4024333333333333"),
		float64(10.1),
		[]interface{}{float32(10.1), float64(10.1)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("f90001"),
		float64(5.960464477539063e-8),
		[]interface{}{float32(5.960464477539063e-8), float64(5.960464477539063e-8)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("fa7f7fffff"),
		float64(3.4028234663852886e+38),
		[]interface{}{float32(3.4028234663852886e+38), float64(3.4028234663852886e+38)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("f90400"),
		float64(0.00006103515625),
		[]interface{}{float32(0.00006103515625), float64(0.00006103515625)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("f933ff"),
		float64(0.2498779296875),
		[]interface{}{float32(0.2498779296875), float64(0.2498779296875)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("fa33000000"),
		float64(2.9802322387695312e-8),
		[]interface{}{float32(2.9802322387695312e-8), float64(2.9802322387695312e-8)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("fa33333866"),
		float64(4.1727979294137185e-8),
		[]interface{}{float32(4.1727979294137185e-8), float64(4.1727979294137185e-8)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
	{
		hexDecode("fa37002000"),
		float64(0.000007636845111846924),
		[]interface{}{float32(0.000007636845111846924), float64(0.000007636845111846924)},
		[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeByteArray, typeByteSlice, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag, typeBigInt},
		0.0,
	},
}

const invalidUTF8ErrorMsg = "cbor: invalid UTF-8 string"

func hexDecode(s string) []byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func bigIntOrPanic(s string) big.Int {
	bi, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("failed to convert " + s + " to big.Int")
	}
	return *bi
}

func TestUnmarshal(t *testing.T) {
	for _, tc := range unmarshalTests {
		// Test unmarshalling CBOR into empty interface.
		var v interface{}
		if err := Unmarshal(tc.cborData, &v); err != nil {
			t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
		} else {
			if tm, ok := tc.emptyInterfaceValue.(time.Time); ok {
				if vt, ok := v.(time.Time); !ok || !tm.Equal(vt) {
					t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborData, v, v, tc.emptyInterfaceValue, tc.emptyInterfaceValue)
				}
			} else if !reflect.DeepEqual(v, tc.emptyInterfaceValue) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborData, v, v, tc.emptyInterfaceValue, tc.emptyInterfaceValue)
			}
		}
		// Test unmarshalling CBOR into RawMessage.
		var r RawMessage
		if err := Unmarshal(tc.cborData, &r); err != nil {
			t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
		} else if !bytes.Equal(r, tc.cborData) {
			t.Errorf("Unmarshal(0x%x) returned RawMessage %v, want %v", tc.cborData, r, tc.cborData)
		}
		// Test unmarshalling CBOR into compatible data types.
		for _, value := range tc.values {
			v := reflect.New(reflect.TypeOf(value))
			vPtr := v.Interface()
			if err := Unmarshal(tc.cborData, vPtr); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
			} else {
				if tm, ok := value.(time.Time); ok {
					if vt, ok := v.Elem().Interface().(time.Time); !ok || !tm.Equal(vt) {
						t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborData, v.Elem().Interface(), v.Elem().Interface(), value, value)
					}
				} else if !reflect.DeepEqual(v.Elem().Interface(), value) {
					t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborData, v.Elem().Interface(), v.Elem().Interface(), value, value)
				}
			}
		}
		// Test unmarshalling CBOR into incompatible data types.
		for _, typ := range tc.wrongTypes {
			v := reflect.New(typ)
			vPtr := v.Interface()
			if err := Unmarshal(tc.cborData, vPtr); err == nil {
				t.Errorf("Unmarshal(0x%x, %s) didn't return an error", tc.cborData, typ.String())
			} else if _, ok := err.(*UnmarshalTypeError); !ok {
				t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", tc.cborData, err)
			} else if !strings.Contains(err.Error(), "cannot unmarshal") {
				t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", tc.cborData, err.Error(), "cannot unmarshal")
			}
		}
	}
}

func TestUnmarshalFloat(t *testing.T) {
	for _, tc := range unmarshalFloatTests {
		// Test unmarshalling CBOR into empty interface.
		var v interface{}
		if err := Unmarshal(tc.cborData, &v); err != nil {
			t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
		} else {
			testFloat(t, tc.cborData, v, tc.emptyInterfaceValue, tc.equalityThreshold)
		}
		// Test unmarshalling CBOR into RawMessage.
		var r RawMessage
		if err := Unmarshal(tc.cborData, &r); err != nil {
			t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
		} else if !bytes.Equal(r, tc.cborData) {
			t.Errorf("Unmarshal(0x%x) returned RawMessage %v, want %v", tc.cborData, r, tc.cborData)
		}
		// Test unmarshalling CBOR into compatible data types.
		for _, value := range tc.values {
			v := reflect.New(reflect.TypeOf(value))
			vPtr := v.Interface()
			if err := Unmarshal(tc.cborData, vPtr); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
			} else {
				testFloat(t, tc.cborData, v.Elem().Interface(), value, tc.equalityThreshold)
			}
		}
		// Test unmarshalling CBOR into incompatible data types.
		for _, typ := range tc.wrongTypes {
			v := reflect.New(typ)
			vPtr := v.Interface()
			if err := Unmarshal(tc.cborData, vPtr); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error", tc.cborData)
			} else if _, ok := err.(*UnmarshalTypeError); !ok {
				t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", tc.cborData, err)
			} else if !strings.Contains(err.Error(), "cannot unmarshal") {
				t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", tc.cborData, err.Error(), "cannot unmarshal")
			}
		}
	}
}

func testFloat(t *testing.T, cborData []byte, f interface{}, wantf interface{}, equalityThreshold float64) {
	switch wantf := wantf.(type) {
	case float32:
		f, ok := f.(float32)
		if !ok {
			t.Errorf("Unmarshal(0x%x) returned value of type %T, want float32", cborData, f)
			return
		}
		if math.IsNaN(float64(wantf)) {
			if !math.IsNaN(float64(f)) {
				t.Errorf("Unmarshal(0x%x) = %f, want NaN", cborData, f)
			}
		} else if math.IsInf(float64(wantf), 0) {
			if f != wantf {
				t.Errorf("Unmarshal(0x%x) = %f, want %f", cborData, f, wantf)
			}
		} else if math.Abs(float64(f-wantf)) > equalityThreshold {
			t.Errorf("Unmarshal(0x%x) = %.18f, want %.18f, diff %.18f > threshold %.18f", cborData, f, wantf, math.Abs(float64(f-wantf)), equalityThreshold)
		}
	case float64:
		f, ok := f.(float64)
		if !ok {
			t.Errorf("Unmarshal(0x%x) returned value of type %T, want float64", cborData, f)
			return
		}
		if math.IsNaN(wantf) {
			if !math.IsNaN(f) {
				t.Errorf("Unmarshal(0x%x) = %f, want NaN", cborData, f)
			}
		} else if math.IsInf(wantf, 0) {
			if f != wantf {
				t.Errorf("Unmarshal(0x%x) = %f, want %f", cborData, f, wantf)
			}
		} else if math.Abs(f-wantf) > equalityThreshold {
			t.Errorf("Unmarshal(0x%x) = %.18f, want %.18f, diff %.18f > threshold %.18f", cborData, f, wantf, math.Abs(f-wantf), equalityThreshold)
		}
	}
}

func TestNegIntOverflow(t *testing.T) {
	cborData := hexDecode("3bffffffffffffffff") // -18446744073709551616

	// Decode CBOR neg int that overflows Go int64 to empty interface
	var v1 interface{}
	wantObj := bigIntOrPanic("-18446744073709551616")
	if err := Unmarshal(cborData, &v1); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %+v", cborData, err)
	} else if !reflect.DeepEqual(v1, wantObj) {
		t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", cborData, v1, v1, wantObj, wantObj)
	}

	// Decode CBOR neg int that overflows Go int64 to big.Int
	var v2 big.Int
	if err := Unmarshal(cborData, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %+v", cborData, err)
	} else if !reflect.DeepEqual(v2, wantObj) {
		t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", cborData, v2, v2, wantObj, wantObj)
	}

	// Decode CBOR neg int that overflows Go int64 to int64
	var v3 int64
	if err := Unmarshal(cborData, &v3); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
	} else if !strings.Contains(err.Error(), "cannot unmarshal") {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), "cannot unmarshal")
	}
}

func TestUnmarshalIntoPtrPrimitives(t *testing.T) {
	cborDataInt := hexDecode("1818")                          // 24
	cborDataString := hexDecode("7f657374726561646d696e67ff") // "streaming"

	const wantInt = 24
	const wantString = "streaming"

	var p1 *int
	var p2 *string
	var p3 *RawMessage

	var i int
	pi := &i
	ppi := &pi

	var s string
	ps := &s
	pps := &ps

	var r RawMessage
	pr := &r
	ppr := &pr

	// Unmarshal CBOR integer into a non-nil pointer.
	if err := Unmarshal(cborDataInt, &ppi); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborDataInt, err)
	} else if i != wantInt {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %d", cborDataInt, i, i, wantInt)
	}
	// Unmarshal CBOR integer into a nil pointer.
	if err := Unmarshal(cborDataInt, &p1); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborDataInt, err)
	} else if *p1 != wantInt {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %d", cborDataInt, *pi, pi, wantInt)
	}

	// Unmarshal CBOR string into a non-nil pointer.
	if err := Unmarshal(cborDataString, &pps); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborDataString, err)
	} else if s != wantString {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v", cborDataString, s, s, wantString)
	}
	// Unmarshal CBOR string into a nil pointer.
	if err := Unmarshal(cborDataString, &p2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborDataString, err)
	} else if *p2 != wantString {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v", cborDataString, *p2, p2, wantString)
	}

	// Unmarshal CBOR string into a non-nil RawMessage.
	if err := Unmarshal(cborDataString, &ppr); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborDataString, err)
	} else if !bytes.Equal(r, cborDataString) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v", cborDataString, r, r, cborDataString)
	}
	// Unmarshal CBOR string into a nil pointer to RawMessage.
	if err := Unmarshal(cborDataString, &p3); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborDataString, err)
	} else if !bytes.Equal(*p3, cborDataString) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v", cborDataString, *p3, p3, cborDataString)
	}
}

func TestUnmarshalIntoPtrArrayPtrElem(t *testing.T) {
	cborData := hexDecode("83010203") // []int{1, 2, 3}

	n1, n2, n3 := 1, 2, 3

	wantArray := []*int{&n1, &n2, &n3}

	var p *[]*int

	var slc []*int
	pslc := &slc
	ppslc := &pslc

	// Unmarshal CBOR array into a non-nil pointer.
	if err := Unmarshal(cborData, &ppslc); err != nil {
		t.Errorf("Unmarshal(0x%x, %s) returned error %v", cborData, reflect.TypeOf(ppslc), err)
	} else if !reflect.DeepEqual(slc, wantArray) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v", cborData, slc, slc, wantArray)
	}
	// Unmarshal CBOR array into a nil pointer.
	if err := Unmarshal(cborData, &p); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if !reflect.DeepEqual(*p, wantArray) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v", cborData, *p, p, wantArray)
	}
}

func TestUnmarshalIntoPtrMapPtrElem(t *testing.T) {
	cborData := hexDecode("a201020304") // {1: 2, 3: 4}

	n1, n2, n3, n4 := 1, 2, 3, 4

	wantMap := map[int]*int{n1: &n2, n3: &n4}

	var p *map[int]*int

	var m map[int]*int
	pm := &m
	ppm := &pm

	// Unmarshal CBOR map into a non-nil pointer.
	if err := Unmarshal(cborData, &ppm); err != nil {
		t.Errorf("Unmarshal(0x%x, %s) returned error %v", cborData, reflect.TypeOf(ppm), err)
	} else if !reflect.DeepEqual(m, wantMap) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v", cborData, m, m, wantMap)
	}
	// Unmarshal CBOR map into a nil pointer.
	if err := Unmarshal(cborData, &p); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if !reflect.DeepEqual(*p, wantMap) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v", cborData, *p, p, wantMap)
	}
}

func TestUnmarshalIntoPtrStructPtrElem(t *testing.T) {
	type s1 struct {
		A *string `cbor:"a"`
		B *string `cbor:"b"`
		C *string `cbor:"c"`
		D *string `cbor:"d"`
		E *string `cbor:"e"`
	}

	cborData := hexDecode("a56161614161626142616361436164614461656145") // map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"}

	a, b, c, d, e := "A", "B", "C", "D", "E"
	wantObj := s1{A: &a, B: &b, C: &c, D: &d, E: &e}

	var p *s1

	var s s1
	ps := &s
	pps := &ps

	// Unmarshal CBOR map into a non-nil pointer.
	if err := Unmarshal(cborData, &pps); err != nil {
		t.Errorf("Unmarshal(0x%x, %s) returned error %v", cborData, reflect.TypeOf(pps), err)
	} else if !reflect.DeepEqual(s, wantObj) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v", cborData, s, s, wantObj)
	}
	// Unmarshal CBOR map into a nil pointer.
	if err := Unmarshal(cborData, &p); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if !reflect.DeepEqual(*p, wantObj) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v", cborData, *p, p, wantObj)
	}
}

func TestUnmarshalIntoArray(t *testing.T) {
	cborData := hexDecode("83010203") // []int{1, 2, 3}

	// Unmarshal CBOR array into Go array.
	var arr1 [3]int
	if err := Unmarshal(cborData, &arr1); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if arr1 != [3]int{1, 2, 3} {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want [3]int{1, 2, 3}", cborData, arr1, arr1)
	}

	// Unmarshal CBOR array into Go array with more elements.
	var arr2 [5]int
	if err := Unmarshal(cborData, &arr2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if arr2 != [5]int{1, 2, 3, 0, 0} {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want [5]int{1, 2, 3, 0, 0}", cborData, arr2, arr2)
	}

	// Unmarshal CBOR array into Go array with less elements.
	var arr3 [1]int
	if err := Unmarshal(cborData, &arr3); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if arr3 != [1]int{1} {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want [0]int{1}", cborData, arr3, arr3)
	}
}

type nilUnmarshaler string

func (s *nilUnmarshaler) UnmarshalCBOR(data []byte) error {
	if len(data) == 1 && (data[0] == 0xf6 || data[0] == 0xf7) {
		*s = "null"
	} else {
		*s = nilUnmarshaler(data)
	}
	return nil
}

func TestUnmarshalNil(t *testing.T) {
	type T struct {
		I int
	}

	cborData := [][]byte{hexDecode("f6"), hexDecode("f7")} // CBOR null and undefined values

	testCases := []struct {
		name      string
		value     interface{}
		wantValue interface{}
	}{
		// Unmarshalling CBOR null to the following types is a no-op.
		{"bool", true, true},
		{"int", int(-1), int(-1)},
		{"int8", int8(-2), int8(-2)},
		{"int16", int16(-3), int16(-3)},
		{"int32", int32(-4), int32(-4)},
		{"int64", int64(-5), int64(-5)},
		{"uint", uint(1), uint(1)},
		{"uint8", uint8(2), uint8(2)},
		{"uint16", uint16(3), uint16(3)},
		{"uint32", uint32(4), uint32(4)},
		{"uint64", uint64(5), uint64(5)},
		{"float32", float32(1.23), float32(1.23)},
		{"float64", float64(4.56), float64(4.56)},
		{"string", "hello", "hello"},
		{"array", [3]int{1, 2, 3}, [3]int{1, 2, 3}},

		// Unmarshalling CBOR null to slice/map sets Go values to nil.
		{"[]byte", []byte{1, 2, 3}, []byte(nil)},
		{"slice", []string{"hello", "world"}, []string(nil)},
		{"map", map[string]bool{"hello": true, "goodbye": false}, map[string]bool(nil)},

		// Unmarshalling CBOR null to time.Time is a no-op.
		{"time.Time", time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC), time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC)},

		// Unmarshalling CBOR null to big.Int is a no-op.
		{"big.Int", bigIntOrPanic("123"), bigIntOrPanic("123")},

		// Unmarshalling CBOR null to user defined struct types is a no-op.
		{"user defined struct", T{I: 123}, T{I: 123}},

		// Unmarshalling CBOR null to cbor.Tag and cbor.RawTag is a no-op.
		{"cbor.RawTag", RawTag{123, []byte{4, 5, 6}}, RawTag{123, []byte{4, 5, 6}}},
		{"cbor.Tag", Tag{123, "hello world"}, Tag{123, "hello world"}},

		// Unmarshalling to cbor.RawMessage sets cbor.RawMessage to raw CBOR bytes (0xf6 or 0xf7).
		// It's tested in TestUnmarshal().

		// Unmarshalling to types implementing cbor.BinaryUnmarshaler is a no-op.
		{"cbor.BinaryUnmarshaler", number(456), number(456)},

		// When unmarshalling to types implementing cbor.Unmarshaler,
		{"cbor.Unmarshaler", nilUnmarshaler("hello world"), nilUnmarshaler("null")},
	}

	// Unmarshalling to values of specified Go types.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, data := range cborData {
				v := reflect.New(reflect.TypeOf(tc.value))
				v.Elem().Set(reflect.ValueOf(tc.value))

				if err := Unmarshal(data, v.Interface()); err != nil {
					t.Errorf("Unmarshal(0x%x) to %T returned error %v", data, v.Elem().Interface(), err)
				} else if !reflect.DeepEqual(v.Elem().Interface(), tc.wantValue) {
					t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", data, v.Elem().Interface(), v.Elem().Interface(), tc.wantValue, tc.wantValue)
				}
			}
		})
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
			err := Unmarshal(cborData, tc.v)
			if err == nil {
				t.Errorf("Unmarshal(0x%x, %v) didn't return an error", cborData, tc.v)
			} else if _, ok := err.(*InvalidUnmarshalError); !ok {
				t.Errorf("Unmarshal(0x%x, %v) error type %T, want *InvalidUnmarshalError", cborData, tc.v, err)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Unmarshal(0x%x, %v) error %q, want %q", cborData, tc.v, err.Error(), tc.wantErrorMsg)
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
	{"Nil data", []byte(nil), "EOF", false},
	{"Empty data", []byte{}, "EOF", false},
	{"Tag number not followed by tag content", []byte{0xc0}, "unexpected EOF", false},
	{"Definite length strings with tagged chunk", hexDecode("5fc64401020304ff"), "cbor: wrong element type tag for indefinite-length byte string", false},
	{"Definite length strings with tagged chunk", hexDecode("7fc06161ff"), "cbor: wrong element type tag for indefinite-length UTF-8 text string", false},
	{"Indefinite length strings with invalid head", hexDecode("7f61"), "unexpected EOF", false},
	{"Invalid nested tag number", hexDecode("d864dc1a514b67b0"), "cbor: invalid additional information", true},
	// Data from 7049bis G.1
	// Premature end of the input
	{"End of input in a head", hexDecode("18"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("19"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("1a"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("1b"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("1901"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("1a0102"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("1b01020304050607"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("38"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("58"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("78"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("98"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("9a01ff00"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("b8"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("d8"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("f8"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("f900"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("fa0000"), "unexpected EOF", false},
	{"End of input in a head", hexDecode("fb000000"), "unexpected EOF", false},
	{"Definite length strings with short data", hexDecode("41"), "unexpected EOF", false},
	{"Definite length strings with short data", hexDecode("61"), "unexpected EOF", false},
	{"Definite length strings with short data", hexDecode("5affffffff00"), "unexpected EOF", false},
	{"Definite length strings with short data", hexDecode("5bffffffffffffffff010203"), "cbor: byte string length 18446744073709551615 is too large, causing integer overflow", false},
	{"Definite length strings with short data", hexDecode("7affffffff00"), "unexpected EOF", false},
	{"Definite length strings with short data", hexDecode("7b7fffffffffffffff010203"), "unexpected EOF", false},
	{"Definite length maps and arrays not closed with enough items", hexDecode("81"), "unexpected EOF", false},
	{"Definite length maps and arrays not closed with enough items", hexDecode("818181818181818181"), "unexpected EOF", false},
	{"Definite length maps and arrays not closed with enough items", hexDecode("8200"), "unexpected EOF", false},
	{"Definite length maps and arrays not closed with enough items", hexDecode("a1"), "unexpected EOF", false},
	{"Definite length maps and arrays not closed with enough items", hexDecode("a20102"), "unexpected EOF", false},
	{"Definite length maps and arrays not closed with enough items", hexDecode("a100"), "unexpected EOF", false},
	{"Definite length maps and arrays not closed with enough items", hexDecode("a2000000"), "unexpected EOF", false},
	{"Indefinite length strings not closed by a break stop code", hexDecode("5f4100"), "unexpected EOF", false},
	{"Indefinite length strings not closed by a break stop code", hexDecode("7f6100"), "unexpected EOF", false},
	{"Indefinite length maps and arrays not closed by a break stop code", hexDecode("9f"), "unexpected EOF", false},
	{"Indefinite length maps and arrays not closed by a break stop code", hexDecode("9f0102"), "unexpected EOF", false},
	{"Indefinite length maps and arrays not closed by a break stop code", hexDecode("bf"), "unexpected EOF", false},
	{"Indefinite length maps and arrays not closed by a break stop code", hexDecode("bf01020102"), "unexpected EOF", false},
	{"Indefinite length maps and arrays not closed by a break stop code", hexDecode("819f"), "unexpected EOF", false},
	{"Indefinite length maps and arrays not closed by a break stop code", hexDecode("9f8000"), "unexpected EOF", false},
	{"Indefinite length maps and arrays not closed by a break stop code", hexDecode("9f9f9f9f9fffffffff"), "unexpected EOF", false},
	{"Indefinite length maps and arrays not closed by a break stop code", hexDecode("9f819f819f9fffffff"), "unexpected EOF", false},
	// Five subkinds of well-formedness error kind 3 (syntax error)
	{"Reserved additional information values", hexDecode("3e"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("5c"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("5d"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("5e"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("7c"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("7d"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("7e"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("9c"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("9d"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("9e"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("bc"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("bd"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("be"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("dc"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("dd"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("de"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("fc"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("fd"), "cbor: invalid additional information", true},
	{"Reserved additional information values", hexDecode("fe"), "cbor: invalid additional information", true},
	{"Reserved two-byte encodings of simple types", hexDecode("f800"), "cbor: invalid simple value 0 for type primitives", true},
	{"Reserved two-byte encodings of simple types", hexDecode("f801"), "cbor: invalid simple value 1 for type primitives", true},
	{"Reserved two-byte encodings of simple types", hexDecode("f818"), "cbor: invalid simple value 24 for type primitives", true},
	{"Reserved two-byte encodings of simple types", hexDecode("f81f"), "cbor: invalid simple value 31 for type primitives", true},
	{"Indefinite length string chunks not of the correct type", hexDecode("5f00ff"), "cbor: wrong element type positive integer for indefinite-length byte string", false},
	{"Indefinite length string chunks not of the correct type", hexDecode("5f21ff"), "cbor: wrong element type negative integer for indefinite-length byte string", false},
	{"Indefinite length string chunks not of the correct type", hexDecode("5f6100ff"), "cbor: wrong element type UTF-8 text string for indefinite-length byte string", false},
	{"Indefinite length string chunks not of the correct type", hexDecode("5f80ff"), "cbor: wrong element type array for indefinite-length byte string", false},
	{"Indefinite length string chunks not of the correct type", hexDecode("5fa0ff"), "cbor: wrong element type map for indefinite-length byte string", false},
	{"Indefinite length string chunks not of the correct type", hexDecode("5fc000ff"), "cbor: wrong element type tag for indefinite-length byte string", false},
	{"Indefinite length string chunks not of the correct type", hexDecode("5fe0ff"), "cbor: wrong element type primitives for indefinite-length byte string", false},
	{"Indefinite length string chunks not of the correct type", hexDecode("7f4100ff"), "cbor: wrong element type byte string for indefinite-length UTF-8 text string", false},
	{"Indefinite length string chunks not definite length", hexDecode("5f5f4100ffff"), "cbor: indefinite-length byte string chunk is not definite-length", false},
	{"Indefinite length string chunks not definite length", hexDecode("7f7f6100ffff"), "cbor: indefinite-length UTF-8 text string chunk is not definite-length", false},
	{"Break occurring on its own outside of an indefinite length item", hexDecode("ff"), "cbor: unexpected \"break\" code", true},
	{"Break occurring in a definite length array or map or a tag", hexDecode("81ff"), "cbor: unexpected \"break\" code", true},
	{"Break occurring in a definite length array or map or a tag", hexDecode("8200ff"), "cbor: unexpected \"break\" code", true},
	{"Break occurring in a definite length array or map or a tag", hexDecode("a1ff"), "cbor: unexpected \"break\" code", true},
	{"Break occurring in a definite length array or map or a tag", hexDecode("a1ff00"), "cbor: unexpected \"break\" code", true},
	{"Break occurring in a definite length array or map or a tag", hexDecode("a100ff"), "cbor: unexpected \"break\" code", true},
	{"Break occurring in a definite length array or map or a tag", hexDecode("a20000ff"), "cbor: unexpected \"break\" code", true},
	{"Break occurring in a definite length array or map or a tag", hexDecode("9f81ff"), "cbor: unexpected \"break\" code", true},
	{"Break occurring in a definite length array or map or a tag", hexDecode("9f829f819f9fffffffff"), "cbor: unexpected \"break\" code", true},
	{"Break in indefinite length map would lead to odd number of items (break in a value position)", hexDecode("bf00ff"), "cbor: unexpected \"break\" code", true},
	{"Break in indefinite length map would lead to odd number of items (break in a value position)", hexDecode("bf000000ff"), "cbor: unexpected \"break\" code", true},
	{"Major type 0 with additional information 31", hexDecode("1f"), "cbor: invalid additional information 31 for type positive integer", true},
	{"Major type 1 with additional information 31", hexDecode("3f"), "cbor: invalid additional information 31 for type negative integer", true},
	{"Major type 6 with additional information 31", hexDecode("df"), "cbor: invalid additional information 31 for type tag", true},
}

func TestInvalidCBORUnmarshal(t *testing.T) {
	for _, tc := range invalidCBORUnmarshalTests {
		t.Run(tc.name, func(t *testing.T) {
			var i interface{}
			err := Unmarshal(tc.cborData, &i)
			if err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error", tc.cborData)
			} else if !tc.errorMsgPartialMatch && err.Error() != tc.wantErrorMsg {
				t.Errorf("Unmarshal(0x%x) error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
			} else if tc.errorMsgPartialMatch && !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestValidUTF8String(t *testing.T) {
	dmRejectInvalidUTF8, err := DecOptions{UTF8: UTF8RejectInvalid}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned an error %+v", err)
	}
	dmDecodeInvalidUTF8, err := DecOptions{UTF8: UTF8DecodeInvalid}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned an error %+v", err)
	}

	testCases := []struct {
		name     string
		cborData []byte
		dm       DecMode
		wantObj  interface{}
	}{
		{
			name:     "with UTF8RejectInvalid",
			cborData: hexDecode("6973747265616d696e67"),
			dm:       dmRejectInvalidUTF8,
			wantObj:  "streaming",
		},
		{
			name:     "with UTF8DecodeInvalid",
			cborData: hexDecode("6973747265616d696e67"),
			dm:       dmDecodeInvalidUTF8,
			wantObj:  "streaming",
		},
		{
			name:     "indef length with UTF8RejectInvalid",
			cborData: hexDecode("7f657374726561646d696e67ff"),
			dm:       dmRejectInvalidUTF8,
			wantObj:  "streaming",
		},
		{
			name:     "indef length with UTF8DecodeInvalid",
			cborData: hexDecode("7f657374726561646d696e67ff"),
			dm:       dmDecodeInvalidUTF8,
			wantObj:  "streaming",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Decode to empty interface
			var i interface{}
			err = tc.dm.Unmarshal(tc.cborData, &i)
			if err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %q", tc.cborData, err)
			}
			if !reflect.DeepEqual(i, tc.wantObj) {
				t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", tc.cborData, i, i, tc.wantObj, tc.wantObj)
			}

			// Decode to string
			var v string
			err = tc.dm.Unmarshal(tc.cborData, &v)
			if err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %q", tc.cborData, err)
			}
			if !reflect.DeepEqual(v, tc.wantObj) {
				t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", tc.cborData, v, v, tc.wantObj, tc.wantObj)
			}
		})
	}
}

func TestInvalidUTF8String(t *testing.T) {
	dmRejectInvalidUTF8, err := DecOptions{UTF8: UTF8RejectInvalid}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned an error %+v", err)
	}
	dmDecodeInvalidUTF8, err := DecOptions{UTF8: UTF8DecodeInvalid}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned an error %+v", err)
	}

	testCases := []struct {
		name         string
		cborData     []byte
		dm           DecMode
		wantObj      interface{}
		wantErrorMsg string
	}{
		{
			name:         "with UTF8RejectInvalid",
			cborData:     hexDecode("61fe"),
			dm:           dmRejectInvalidUTF8,
			wantErrorMsg: invalidUTF8ErrorMsg,
		},
		{
			name:     "with UTF8DecodeInvalid",
			cborData: hexDecode("61fe"),
			dm:       dmDecodeInvalidUTF8,
			wantObj:  string([]byte{0xfe}),
		},
		{
			name:         "indef length with UTF8RejectInvalid",
			cborData:     hexDecode("7f62e6b061b4ff"),
			dm:           dmRejectInvalidUTF8,
			wantErrorMsg: invalidUTF8ErrorMsg,
		},
		{
			name:     "indef length with UTF8DecodeInvalid",
			cborData: hexDecode("7f62e6b061b4ff"),
			dm:       dmDecodeInvalidUTF8,
			wantObj:  string([]byte{0xe6, 0xb0, 0xb4}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Decode to empty interface
			var v interface{}
			err = tc.dm.Unmarshal(tc.cborData, &v)
			if tc.wantErrorMsg != "" {
				if err == nil {
					t.Errorf("Unmarshal(0x%x) didn't return error", tc.cborData)
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("Unmarshal(0x%x) error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Unmarshal(0x%x) returned error %q", tc.cborData, err)
				}
				if !reflect.DeepEqual(v, tc.wantObj) {
					t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", tc.cborData, v, v, tc.wantObj, tc.wantObj)
				}
			}

			// Decode to string
			var s string
			err = tc.dm.Unmarshal(tc.cborData, &s)
			if tc.wantErrorMsg != "" {
				if err == nil {
					t.Errorf("Unmarshal(0x%x) didn't return error", tc.cborData)
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("Unmarshal(0x%x) error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Unmarshal(0x%x) returned error %q", tc.cborData, err)
				}
				if !reflect.DeepEqual(s, tc.wantObj) {
					t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", tc.cborData, s, s, tc.wantObj, tc.wantObj)
				}
			}
		})
	}

	// Test decoding of mixed invalid text string and valid text string
	// with UTF8RejectInvalid option (default)
	cborData := hexDecode("7f62e6b061b4ff7f657374726561646d696e67ff")
	dec := NewDecoder(bytes.NewReader(cborData))
	var s string
	if err := dec.Decode(&s); err == nil {
		t.Errorf("Decode() didn't return an error")
	} else if s != "" {
		t.Errorf("Decode() returned %q, want %q", s, "")
	}
	if err := dec.Decode(&s); err != nil {
		t.Errorf("Decode() returned error %v", err)
	} else if s != "streaming" {
		t.Errorf("Decode() returned %q, want %q", s, "streaming")
	}

	// Test decoding of mixed invalid text string and valid text string
	// with UTF8DecodeInvalid option
	dec = dmDecodeInvalidUTF8.NewDecoder(bytes.NewReader(cborData))
	if err := dec.Decode(&s); err != nil {
		t.Errorf("Decode() returned error %q", err)
	} else if s != string([]byte{0xe6, 0xb0, 0xb4}) {
		t.Errorf("Decode() returned %q, want %q", s, string([]byte{0xe6, 0xb0, 0xb4}))
	}
	if err := dec.Decode(&s); err != nil {
		t.Errorf("Decode() returned error %v", err)
	} else if s != "streaming" {
		t.Errorf("Decode() returned %q, want %q", s, "streaming")
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
			if err := Unmarshal(tc.cborData, &v); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
			} else if !reflect.DeepEqual(v, want) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborData, v, v, want, want)
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
	wantCBORType := "UTF-8 text string"
	wantGoType := "int"
	wantStructFieldName := "cbor.outer2.ArrayField"
	wantErrorMsg := "cannot unmarshal UTF-8 text string into Go struct field cbor.outer2.ArrayField of type int"

	var v outer2
	if err := Unmarshal(cborData, &v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else {
		if typeError, ok := err.(*UnmarshalTypeError); !ok {
			t.Errorf("Unmarshal(0x%x) returned wrong type of error %T, want (*UnmarshalTypeError)", cborData, err)
		} else {
			if typeError.CBORType != wantCBORType {
				t.Errorf("Unmarshal(0x%x) returned (*UnmarshalTypeError).CBORType %s, want %s", cborData, typeError.CBORType, wantCBORType)
			}
			if typeError.GoType != wantGoType {
				t.Errorf("Unmarshal(0x%x) returned (*UnmarshalTypeError).GoType %s, want %s", cborData, typeError.GoType, wantGoType)
			}
			if typeError.StructFieldName != wantStructFieldName {
				t.Errorf("Unmarshal(0x%x) returned (*UnmarshalTypeError).StructFieldName %s, want %s", cborData, typeError.StructFieldName, wantStructFieldName)
			}
			if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
			}
		}
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, want, want)
	}
}

func TestUnmarshalStructError2(t *testing.T) {
	// Unmarshal integer and invalid UTF8 string as field name into struct
	type strc struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
	}
	want := strc{
		A: "A",
	}

	// Unmarshal returns first error encountered, which is *UnmarshalTypeError (failed to unmarshal int into Go string)
	cborData := hexDecode("a3fa47c35000026161614161fe6142") // {100000.0:2, "a":"A", 0xfe: B}
	wantCBORType := "primitives"
	wantGoType := "string"
	wantErrorMsg := "cannot unmarshal primitives into Go value of type string"

	v := strc{}
	if err := Unmarshal(cborData, &v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else {
		if typeError, ok := err.(*UnmarshalTypeError); !ok {
			t.Errorf("Unmarshal(0x%x) returned wrong type of error %T, want (*UnmarshalTypeError)", cborData, err)
		} else {
			if typeError.CBORType != wantCBORType {
				t.Errorf("Unmarshal(0x%x) returned (*UnmarshalTypeError).CBORType %s, want %s", cborData, typeError.CBORType, wantCBORType)
			}
			if typeError.GoType != wantGoType {
				t.Errorf("Unmarshal(0x%x) returned (*UnmarshalTypeError).GoType %s, want %s", cborData, typeError.GoType, wantGoType)
			}
			if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
			}
		}
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, want, want)
	}

	// Unmarshal returns first error encountered, which is *cbor.SemanticError (invalid UTF8 string)
	cborData = hexDecode("a361fe6142010261616141") // {0xfe: B, 1:2, "a":"A"}
	v = strc{}
	if err := Unmarshal(cborData, &v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else {
		if _, ok := err.(*SemanticError); !ok {
			t.Errorf("Unmarshal(0x%x) returned wrong type of error %T, want (*SemanticError)", cborData, err)
		} else if err.Error() != invalidUTF8ErrorMsg {
			t.Errorf("Unmarshal(0x%x) returned error %q, want error %q", cborData, err.Error(), invalidUTF8ErrorMsg)
		}
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, want, want)
	}

	// Unmarshal returns first error encountered, which is *cbor.SemanticError (invalid UTF8 string)
	cborData = hexDecode("a3616261fe010261616141") // {"b": 0xfe, 1:2, "a":"A"}
	v = strc{}
	if err := Unmarshal(cborData, &v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else {
		if _, ok := err.(*SemanticError); !ok {
			t.Errorf("Unmarshal(0x%x) returned wrong type of error %T, want (*SemanticError)", cborData, err)
		} else if err.Error() != invalidUTF8ErrorMsg {
			t.Errorf("Unmarshal(0x%x) returned error %q, want error %q", cborData, err.Error(), invalidUTF8ErrorMsg)
		}
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, want, want)
	}
}

func TestUnmarshalPrefilledArray(t *testing.T) {
	prefilledArr := []int{1, 2, 3, 4, 5}
	want := []int{10, 11, 3, 4, 5}
	cborData := hexDecode("820a0b") // []int{10, 11}
	if err := Unmarshal(cborData, &prefilledArr); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if len(prefilledArr) != 2 || cap(prefilledArr) != 5 {
		t.Errorf("Unmarshal(0x%x) = %v (len %d, cap %d), want len == 2, cap == 5", cborData, prefilledArr, len(prefilledArr), cap(prefilledArr))
	}
	if !reflect.DeepEqual(prefilledArr[:cap(prefilledArr)], want) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, prefilledArr, prefilledArr, want, want)
	}

	cborData = hexDecode("80") // empty array
	if err := Unmarshal(cborData, &prefilledArr); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if len(prefilledArr) != 0 || cap(prefilledArr) != 0 {
		t.Errorf("Unmarshal(0x%x) = %v (len %d, cap %d), want len == 0, cap == 0", cborData, prefilledArr, len(prefilledArr), cap(prefilledArr))
	}
}

func TestUnmarshalPrefilledMap(t *testing.T) {
	prefilledMap := map[string]string{"key": "value", "a": "1"}
	want := map[string]string{"key": "value", "a": "A", "b": "B", "c": "C", "d": "D", "e": "E"}
	cborData := hexDecode("a56161614161626142616361436164614461656145") // map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"}
	if err := Unmarshal(cborData, &prefilledMap); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(prefilledMap, want) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, prefilledMap, prefilledMap, want, want)
	}

	prefilledMap = map[string]string{"key": "value"}
	want = map[string]string{"key": "value"}
	cborData = hexDecode("a0") // map[string]string{}
	if err := Unmarshal(cborData, &prefilledMap); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(prefilledMap, want) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, prefilledMap, prefilledMap, want, want)
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
	if err := Unmarshal(cborData, &prefilledStruct); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(prefilledStruct, want) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, prefilledStruct, prefilledStruct, want, want)
	}
	if len(prefilledStruct.B) != 2 || cap(prefilledStruct.B) != 4 {
		t.Errorf("Unmarshal(0x%x) = %v (len %d, cap %d), want len == 2, cap == 5", cborData, prefilledStruct.B, len(prefilledStruct.B), cap(prefilledStruct.B))
	}
	if !reflect.DeepEqual(prefilledStruct.B[:cap(prefilledStruct.B)], []int{2, 3, 400, 500}) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, prefilledStruct.B, prefilledStruct.B, []int{2, 3, 400, 500}, []int{2, 3, 400, 500})
	}
}

func TestStructFieldNil(t *testing.T) {
	type TestStruct struct {
		I   int
		PI  *int
		PPI **int
	}
	var struc TestStruct
	cborData, err := Marshal(struc)
	if err != nil {
		t.Fatalf("Marshal(%+v) returned error %v", struc, err)
	}
	var struc2 TestStruct
	err = Unmarshal(cborData, &struc2)
	if err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if !reflect.DeepEqual(struc, struc2) {
		t.Errorf("Unmarshal(0x%x) returned %+v, want %+v", cborData, struc2, struc)
	}
}

func TestLengthOverflowsInt(t *testing.T) {
	// Data is generating by go-fuzz.
	// string/slice/map length in uint64 cast to int causes integer overflow.
	cborData := [][]byte{
		hexDecode("bbcf30303030303030cfd697829782"),
		hexDecode("5bcf30303030303030cfd697829782"),
	}
	wantErrorMsg := "is too large"
	for _, data := range cborData {
		var intf interface{}
		if err := Unmarshal(data, &intf); err == nil {
			t.Errorf("Unmarshal(0x%x) didn't return an error, want error containing substring %q", data, wantErrorMsg)
		} else if !strings.Contains(err.Error(), wantErrorMsg) {
			t.Errorf("Unmarshal(0x%x) returned error %q, want error containing substring %q", data, err.Error(), wantErrorMsg)
		}
	}
}

func TestMapKeyUnhashable(t *testing.T) {
	testCases := []struct {
		name         string
		cborData     []byte
		wantErrorMsg string
	}{
		{"slice as map key", hexDecode("bf8030ff"), "cbor: invalid map key type: []interface {}"},                                                             // {[]: -17}
		{"slice as map key", hexDecode("a1813030"), "cbor: invalid map key type: []interface {}"},                                                             // {[-17]: -17}
		{"slice as map key", hexDecode("bfd1a388f730303030303030303030303030ff"), "cbor: invalid map key type: []interface {}"},                               // {17({[undefined, -17, -17, -17, -17, -17, -17, -17]: -17, -17: -17}): -17}}
		{"byte slice as map key", hexDecode("8f3030a730304430303030303030303030303030303030303030303030303030303030"), "cbor: invalid map key type: []uint8"}, // [-17, -17, {-17: -17, h'30303030': -17}, -17, -17, -17, -17, -17, -17, -17, -17, -17, -17, -17, -17]
		{"map as map key", hexDecode("bf30a1a030ff"), "cbor: invalid map key type: map"},                                                                      // {-17: {{}: -17}}, empty map as map key
		{"map as map key", hexDecode("bfb0303030303030303030303030303030303030303030303030303030303030303030ff"), "cbor: invalid map key type: map"},          // {{-17: -17}: -17}, map as key
		{"tagged slice as map key", hexDecode("a1c84c30303030303030303030303030"), "cbor: invalid map key type: cbor.Tag"},                                    // {8(h'303030303030303030303030'): -17}
		{"nested-tagged slice as map key", hexDecode("a33030306430303030d1cb4030"), "cbor: invalid map key type: cbor.Tag"},                                   // {-17: "0000", 17(11(h'')): -17}
		{"big.Int as map key", hexDecode("a13bbd3030303030303030"), "cbor: invalid map key type: big.Int"},                                                    // {-13632449055575519281: -17}
		{"tagged big.Int as map key", hexDecode("a1c24901000000000000000030"), "cbor: invalid map key type: big.Int"},                                         // {18446744073709551616: -17}
		{"tagged big.Int as map key", hexDecode("a1c34901000000000000000030"), "cbor: invalid map key type: big.Int"},                                         // {-18446744073709551617: -17}
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v interface{}
			if err := Unmarshal(tc.cborData, &v); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", tc.cborData, tc.wantErrorMsg)
			} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
			}
			if _, ok := v.(map[interface{}]interface{}); ok {
				var v map[interface{}]interface{}
				if err := Unmarshal(tc.cborData, &v); err == nil {
					t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", tc.cborData, tc.wantErrorMsg)
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
				}
			}
		})
	}
}

func TestMapKeyNaN(t *testing.T) {
	// Data is generating by go-fuzz.
	cborData := hexDecode("b0303030303030303030303030303030303038303030faffff30303030303030303030303030") // {-17: -17, NaN: -17}
	var intf interface{}
	if err := Unmarshal(cborData, &intf); err != nil {
		t.Fatalf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	em, err := EncOptions{Sort: SortCanonical}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned an error %v", err)
	}
	if _, err := em.Marshal(intf); err != nil {
		t.Errorf("Marshal(%v) returned error %v", intf, err)
	}
}

func TestUnmarshalUndefinedElement(t *testing.T) {
	// Data is generating by go-fuzz.
	cborData := hexDecode("bfd1a388f730303030303030303030303030ff") // {17({[undefined, -17, -17, -17, -17, -17, -17, -17]: -17, -17: -17}): -17}
	var intf interface{}
	wantErrorMsg := "invalid map key type"
	if err := Unmarshal(cborData, &intf); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want error containing substring %q", cborData, wantErrorMsg)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing substring %q", cborData, err.Error(), wantErrorMsg)
	}
}

func TestMapKeyNil(t *testing.T) {
	testData := [][]byte{
		hexDecode("a1f630"), // {null: -17}
	}
	want := map[interface{}]interface{}{nil: int64(-17)}
	for _, data := range testData {
		var intf interface{}
		if err := Unmarshal(data, &intf); err != nil {
			t.Fatalf("Unmarshal(0x%x) returned error %v", data, err)
		} else if !reflect.DeepEqual(intf, want) {
			t.Errorf("Unmarshal(0x%x) returned %+v, want %+v", data, intf, want)
		}
		if _, err := Marshal(intf); err != nil {
			t.Errorf("Marshal(%v) returned error %v", intf, err)
		}

		var v map[interface{}]interface{}
		if err := Unmarshal(data, &v); err != nil {
			t.Errorf("Unmarshal(0x%x) returned error %v", data, err)
		} else if !reflect.DeepEqual(v, want) {
			t.Errorf("Unmarshal(0x%x) returned %+v, want %+v", data, v, want)
		}
		if _, err := Marshal(v); err != nil {
			t.Errorf("Marshal(%v) returned error %v", v, err)
		}
	}
}

func TestDecodeTime(t *testing.T) {
	testCases := []struct {
		name            string
		cborRFC3339Time []byte
		cborUnixTime    []byte
		wantTime        time.Time
	}{
		// Decoding CBOR null/defined to time.Time is no-op.  See TestUnmarshalNil.
		{
			name:            "NaN",
			cborRFC3339Time: hexDecode("f97e00"),
			cborUnixTime:    hexDecode("f97e00"),
			wantTime:        time.Time{},
		},
		{
			name:            "positive infinity",
			cborRFC3339Time: hexDecode("f97c00"),
			cborUnixTime:    hexDecode("f97c00"),
			wantTime:        time.Time{},
		},
		{
			name:            "negative infinity",
			cborRFC3339Time: hexDecode("f9fc00"),
			cborUnixTime:    hexDecode("f9fc00"),
			wantTime:        time.Time{},
		},
		{
			name:            "time without fractional seconds", // positive integer
			cborRFC3339Time: hexDecode("74323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("1a514b67b0"),
			wantTime:        parseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "time with fractional seconds", // float
			cborRFC3339Time: hexDecode("7819313937302d30312d30315432313a34363a34302d30363a3030"),
			cborUnixTime:    hexDecode("fa47c35000"),
			wantTime:        parseTime(time.RFC3339Nano, "1970-01-01T21:46:40-06:00"),
		},
		{
			name:            "time with fractional seconds", // float
			cborRFC3339Time: hexDecode("76323031332d30332d32315432303a30343a30302e355a"),
			cborUnixTime:    hexDecode("fb41d452d9ec200000"),
			wantTime:        parseTime(time.RFC3339Nano, "2013-03-21T20:04:00.5Z"),
		},
		{
			name:            "time before January 1, 1970 UTC without fractional seconds", // negative integer
			cborRFC3339Time: hexDecode("74313936392d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("3a0177f2cf"),
			wantTime:        parseTime(time.RFC3339Nano, "1969-03-21T20:04:00Z"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm := time.Now()
			if err := Unmarshal(tc.cborRFC3339Time, &tm); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborRFC3339Time, err)
			} else if !tc.wantTime.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborRFC3339Time, tm, tm, tc.wantTime, tc.wantTime)
			}
			tm = time.Now()
			if err := Unmarshal(tc.cborUnixTime, &tm); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborUnixTime, err)
			} else if !tc.wantTime.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborUnixTime, tm, tm, tc.wantTime, tc.wantTime)
			}
		})
	}
}

func TestDecodeTimeWithTag(t *testing.T) {
	testCases := []struct {
		name            string
		cborRFC3339Time []byte
		cborUnixTime    []byte
		wantTime        time.Time
	}{
		{
			name:            "time without fractional seconds", // positive integer
			cborRFC3339Time: hexDecode("c074323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("c11a514b67b0"),
			wantTime:        parseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "time with fractional seconds", // float
			cborRFC3339Time: hexDecode("c076323031332d30332d32315432303a30343a30302e355a"),
			cborUnixTime:    hexDecode("c1fb41d452d9ec200000"),
			wantTime:        parseTime(time.RFC3339Nano, "2013-03-21T20:04:00.5Z"),
		},
		{
			name:            "time before January 1, 1970 UTC without fractional seconds", // negative integer
			cborRFC3339Time: hexDecode("c074313936392d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("c13a0177f2cf"),
			wantTime:        parseTime(time.RFC3339Nano, "1969-03-21T20:04:00Z"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm := time.Now()
			if err := Unmarshal(tc.cborRFC3339Time, &tm); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborRFC3339Time, err)
			} else if !tc.wantTime.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborRFC3339Time, tm, tm, tc.wantTime, tc.wantTime)
			}
			tm = time.Now()
			if err := Unmarshal(tc.cborUnixTime, &tm); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborUnixTime, err)
			} else if !tc.wantTime.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborUnixTime, tm, tm, tc.wantTime, tc.wantTime)
			}

			var v interface{}
			if err := Unmarshal(tc.cborRFC3339Time, &v); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborRFC3339Time, err)
			} else if tm, ok := v.(time.Time); !ok || !tc.wantTime.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborRFC3339Time, v, v, tc.wantTime, tc.wantTime)
			}
			v = nil
			if err := Unmarshal(tc.cborUnixTime, &v); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborUnixTime, err)
			} else if tm, ok := v.(time.Time); !ok || !tc.wantTime.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborUnixTime, v, v, tc.wantTime, tc.wantTime)
			}
		})
	}
}

func TestDecodeTimeError(t *testing.T) {
	testCases := []struct {
		name         string
		cborData     []byte
		wantErrorMsg string
	}{
		{
			name:         "invalid RFC3339 time string",
			cborData:     hexDecode("7f657374726561646d696e67ff"),
			wantErrorMsg: "cbor: cannot set streaming for time.Time",
		},
		{
			name:         "byte string data cannot be decoded into time.Time",
			cborData:     hexDecode("4f013030303030303030e03031ed3030"),
			wantErrorMsg: "cbor: cannot unmarshal byte string into Go value of type time.Time",
		},
		{
			name:         "bool cannot be decoded into time.Time",
			cborData:     hexDecode("f4"),
			wantErrorMsg: "cbor: cannot unmarshal primitives into Go value of type time.Time",
		},
		{
			name:         "invalid UTF-8 string",
			cborData:     hexDecode("7f62e6b061b4ff"),
			wantErrorMsg: "cbor: invalid UTF-8 string",
		},
		{
			name:         "negative integer overflow",
			cborData:     hexDecode("3bffffffffffffffff"),
			wantErrorMsg: "cbor: cannot unmarshal negative integer into Go value of type time.Time",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm := time.Now()
			if err := Unmarshal(tc.cborData, &tm); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", tc.cborData, tc.wantErrorMsg)
			} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecodeInvalidTagTime(t *testing.T) {
	typeTimeSlice := reflect.TypeOf([]time.Time{})

	testCases := []struct {
		name          string
		cborData      []byte
		decodeToTypes []reflect.Type
		wantErrorMsg  string
	}{
		{
			name:          "Tag 0 with invalid RFC3339 time string",
			cborData:      hexDecode("c07f657374726561646d696e67ff"),
			decodeToTypes: []reflect.Type{typeIntf, typeTime},
			wantErrorMsg:  "cbor: cannot set streaming for time.Time",
		},
		{
			name:          "Tag 0 with invalid UTF-8 string",
			cborData:      hexDecode("c07f62e6b061b4ff"),
			decodeToTypes: []reflect.Type{typeIntf, typeTime},
			wantErrorMsg:  "cbor: invalid UTF-8 string",
		},
		{
			name:          "Tag 0 with integer content",
			cborData:      hexDecode("c01a514b67b0"),
			decodeToTypes: []reflect.Type{typeIntf, typeTime},
			wantErrorMsg:  "cbor: tag number 0 must be followed by text string, got positive integer",
		},
		{
			name:          "Tag 0 with byte string content",
			cborData:      hexDecode("c04f013030303030303030e03031ed3030"),
			decodeToTypes: []reflect.Type{typeIntf, typeTime},
			wantErrorMsg:  "cbor: tag number 0 must be followed by text string, got byte string",
		},
		{
			name:          "Tag 0 with integer content as array element",
			cborData:      hexDecode("81c01a514b67b0"),
			decodeToTypes: []reflect.Type{typeIntf, typeTimeSlice},
			wantErrorMsg:  "cbor: tag number 0 must be followed by text string, got positive integer",
		},
		{
			name:          "Tag 1 with negative integer overflow",
			cborData:      hexDecode("c13bffffffffffffffff"),
			decodeToTypes: []reflect.Type{typeIntf, typeTime},
			wantErrorMsg:  "cbor: cannot unmarshal tag into Go value of type time.Time",
		},
		{
			name:          "Tag 1 with string content",
			cborData:      hexDecode("c174323031332d30332d32315432303a30343a30305a"),
			decodeToTypes: []reflect.Type{typeIntf, typeTime},
			wantErrorMsg:  "cbor: tag number 1 must be followed by integer or floating-point number, got UTF-8 text string",
		},
		{
			name:          "Tag 1 with simple value",
			cborData:      hexDecode("d801f6"), // 1(null)
			decodeToTypes: []reflect.Type{typeIntf, typeTime},
			wantErrorMsg:  "cbor: tag number 1 must be followed by integer or floating-point number, got primitive",
		},
		{
			name:          "Tag 1 with string content as array element",
			cborData:      hexDecode("81c174323031332d30332d32315432303a30343a30305a"),
			decodeToTypes: []reflect.Type{typeIntf, typeTimeSlice},
			wantErrorMsg:  "cbor: tag number 1 must be followed by integer or floating-point number, got UTF-8 text string",
		},
	}
	dm, _ := DecOptions{TimeTag: DecTagOptional}.DecMode()
	for _, tc := range testCases {
		for _, decodeToType := range tc.decodeToTypes {
			t.Run(tc.name+" decode to "+decodeToType.String(), func(t *testing.T) {
				v := reflect.New(decodeToType)
				if err := dm.Unmarshal(tc.cborData, v.Interface()); err == nil {
					t.Errorf("Unmarshal(0x%x) didn't return error, want error msg %q", tc.cborData, tc.wantErrorMsg)
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err, tc.wantErrorMsg)
				}
			})
		}
	}
}

func TestDecodeTag0Error(t *testing.T) {
	cborData := hexDecode("c01a514b67b0") // 0(1363896240)
	wantErrorMsg := "cbor: tag number 0 must be followed by text string, got positive integer"

	timeTagIgnoredDM, _ := DecOptions{TimeTag: DecTagIgnored}.DecMode()
	timeTagOptionalDM, _ := DecOptions{TimeTag: DecTagOptional}.DecMode()
	timeTagRequiredDM, _ := DecOptions{TimeTag: DecTagRequired}.DecMode()

	testCases := []struct {
		name string
		dm   DecMode
	}{
		{name: "DecTagIgnored", dm: timeTagIgnoredDM},
		{name: "DecTagOptional", dm: timeTagOptionalDM},
		{name: "DecTagRequired", dm: timeTagRequiredDM},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Decode to interface{}
			var v interface{}
			if err := tc.dm.Unmarshal(cborData, &v); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return error, want error msg %q", cborData, wantErrorMsg)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", cborData, err, wantErrorMsg)
			}

			// Decode to time.Time
			var tm time.Time
			if err := tc.dm.Unmarshal(cborData, &tm); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return error, want error msg %q", cborData, wantErrorMsg)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", cborData, err, wantErrorMsg)
			}

			// Decode to uint64
			var ui uint64
			if err := tc.dm.Unmarshal(cborData, &ui); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return error, want error msg %q", cborData, wantErrorMsg)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", cborData, err, wantErrorMsg)
			}
		})
	}
}

func TestDecodeTag1Error(t *testing.T) {
	cborData := hexDecode("c174323031332d30332d32315432303a30343a30305a") // 1("2013-03-21T20:04:00Z")
	wantErrorMsg := "cbor: tag number 1 must be followed by integer or floating-point number, got UTF-8 text string"

	timeTagIgnoredDM, _ := DecOptions{TimeTag: DecTagIgnored}.DecMode()
	timeTagOptionalDM, _ := DecOptions{TimeTag: DecTagOptional}.DecMode()
	timeTagRequiredDM, _ := DecOptions{TimeTag: DecTagRequired}.DecMode()

	testCases := []struct {
		name string
		dm   DecMode
	}{
		{name: "DecTagIgnored", dm: timeTagIgnoredDM},
		{name: "DecTagOptional", dm: timeTagOptionalDM},
		{name: "DecTagRequired", dm: timeTagRequiredDM},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Decode to interface{}
			var v interface{}
			if err := tc.dm.Unmarshal(cborData, &v); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return error, want error msg %q", cborData, wantErrorMsg)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", cborData, err, wantErrorMsg)
			}

			// Decode to time.Time
			var tm time.Time
			if err := tc.dm.Unmarshal(cborData, &tm); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return error, want error msg %q", cborData, wantErrorMsg)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", cborData, err, wantErrorMsg)
			}

			// Decode to string
			var s string
			if err := tc.dm.Unmarshal(cborData, &s); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return error, want error msg %q", cborData, wantErrorMsg)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", cborData, err, wantErrorMsg)
			}
		})
	}
}

func TestDecodeTimeStreaming(t *testing.T) {
	// Decoder decodes from mixed invalid and valid time.
	testCases := []struct {
		cborData     []byte
		wantErrorMsg string
		wantObj      time.Time
	}{
		{
			cborData:     hexDecode("c07f62e6b061b4ff"),
			wantErrorMsg: "cbor: invalid UTF-8 string",
		},
		{
			cborData: hexDecode("c074323031332d30332d32315432303a30343a30305a"),
			wantObj:  time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC),
		},
		{
			cborData:     hexDecode("c01a514b67b0"),
			wantErrorMsg: "cbor: tag number 0 must be followed by text string, got positive integer",
		},
		{
			cborData: hexDecode("c074323031332d30332d32315432303a30343a30305a"),
			wantObj:  time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC),
		},
		{
			cborData:     hexDecode("c13bffffffffffffffff"),
			wantErrorMsg: "cbor: cannot unmarshal tag into Go value of type time.Time",
		},
		{
			cborData: hexDecode("c11a514b67b0"),
			wantObj:  time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC),
		},
		{
			cborData:     hexDecode("c174323031332d30332d32315432303a30343a30305a"),
			wantErrorMsg: "tag number 1 must be followed by integer or floating-point number, got UTF-8 text string",
		},
		{
			cborData: hexDecode("c11a514b67b0"),
			wantObj:  time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC),
		},
	}
	// Data is a mixed stream of valid and invalid time data
	var cborData []byte
	for _, tc := range testCases {
		cborData = append(cborData, tc.cborData...)
	}
	dm, _ := DecOptions{TimeTag: DecTagOptional}.DecMode()
	dec := dm.NewDecoder(bytes.NewReader(cborData))
	for _, tc := range testCases {
		var v interface{}
		err := dec.Decode(&v)
		if tc.wantErrorMsg != "" {
			if err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return error, want error msg %q", tc.cborData, tc.wantErrorMsg)
			} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error msg %q, want %q", tc.cborData, err, tc.wantErrorMsg)
			}
		} else {
			tm, ok := v.(time.Time)
			if !ok {
				t.Errorf("Unmarshal(0x%x) returned %s (%T), want time.Time", tc.cborData, v, v)
			}
			if !tc.wantObj.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) returned %s, want %s", tc.cborData, tm, tc.wantObj)
			}
		}
	}
	dec = dm.NewDecoder(bytes.NewReader(cborData))
	for _, tc := range testCases {
		var tm time.Time
		err := dec.Decode(&tm)
		if tc.wantErrorMsg != "" {
			if err == nil {
				t.Errorf("Unmarshal(0x%x) did't return error, want error msg %q", tc.cborData, tc.wantErrorMsg)
			} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error msg %q, want %q", tc.cborData, err, tc.wantErrorMsg)
			}
		} else {
			if !tc.wantObj.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) returned %s, want %s", tc.cborData, tm, tc.wantObj)
			}
		}
	}
}

func TestDecTimeTagOption(t *testing.T) {
	timeTagIgnoredDecMode, _ := DecOptions{TimeTag: DecTagIgnored}.DecMode()
	timeTagOptionalDecMode, _ := DecOptions{TimeTag: DecTagOptional}.DecMode()
	timeTagRequiredDecMode, _ := DecOptions{TimeTag: DecTagRequired}.DecMode()

	testCases := []struct {
		name            string
		cborRFC3339Time []byte
		cborUnixTime    []byte
		decMode         DecMode
		wantTime        time.Time
		wantErrorMsg    string
	}{
		// not-tagged time CBOR data
		{
			name:            "not-tagged data with DecTagIgnored option",
			cborRFC3339Time: hexDecode("74323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("1a514b67b0"),
			decMode:         timeTagIgnoredDecMode,
			wantTime:        parseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "not-tagged data with timeTagOptionalDecMode option",
			cborRFC3339Time: hexDecode("74323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("1a514b67b0"),
			decMode:         timeTagOptionalDecMode,
			wantTime:        parseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "not-tagged data with timeTagRequiredDecMode option",
			cborRFC3339Time: hexDecode("74323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("1a514b67b0"),
			decMode:         timeTagRequiredDecMode,
			wantErrorMsg:    "expect CBOR tag value",
		},
		// tagged time CBOR data
		{
			name:            "tagged data with timeTagIgnoredDecMode option",
			cborRFC3339Time: hexDecode("c074323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("c11a514b67b0"),
			decMode:         timeTagIgnoredDecMode,
			wantTime:        parseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "tagged data with timeTagOptionalDecMode option",
			cborRFC3339Time: hexDecode("c074323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("c11a514b67b0"),
			decMode:         timeTagOptionalDecMode,
			wantTime:        parseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "tagged data with timeTagRequiredDecMode option",
			cborRFC3339Time: hexDecode("c074323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("c11a514b67b0"),
			decMode:         timeTagRequiredDecMode,
			wantTime:        parseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		// mis-tagged time CBOR data
		{
			name:            "mis-tagged data with timeTagIgnoredDecMode option",
			cborRFC3339Time: hexDecode("c8c974323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("c8c91a514b67b0"),
			decMode:         timeTagIgnoredDecMode,
			wantTime:        parseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "mis-tagged data with timeTagOptionalDecMode option",
			cborRFC3339Time: hexDecode("c8c974323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("c8c91a514b67b0"),
			decMode:         timeTagOptionalDecMode,
			wantErrorMsg:    "cbor: wrong tag number for time.Time, got 8, expect 0 or 1",
		},
		{
			name:            "mis-tagged data with timeTagRequiredDecMode option",
			cborRFC3339Time: hexDecode("c8c974323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    hexDecode("c8c91a514b67b0"),
			decMode:         timeTagRequiredDecMode,
			wantErrorMsg:    "cbor: wrong tag number for time.Time, got 8, expect 0 or 1",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm := time.Now()
			err := tc.decMode.Unmarshal(tc.cborRFC3339Time, &tm)
			if tc.wantErrorMsg != "" {
				if err == nil {
					t.Errorf("Unmarshal(0x%x) didn't return error", tc.cborRFC3339Time)
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", tc.cborRFC3339Time, err.Error(), tc.wantErrorMsg)
				}
			} else {
				if !tc.wantTime.Equal(tm) {
					t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborRFC3339Time, tm, tm, tc.wantTime, tc.wantTime)
				}
			}

			tm = time.Now()
			err = tc.decMode.Unmarshal(tc.cborUnixTime, &tm)
			if tc.wantErrorMsg != "" {
				if err == nil {
					t.Errorf("Unmarshal(0x%x) didn't return error", tc.cborRFC3339Time)
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", tc.cborRFC3339Time, err.Error(), tc.wantErrorMsg)
				}
			} else {
				if !tc.wantTime.Equal(tm) {
					t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborRFC3339Time, tm, tm, tc.wantTime, tc.wantTime)
				}
			}
		})
	}
}

func TestUnmarshalStructTag1(t *testing.T) {
	type strc struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
	}
	want := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	cborData := hexDecode("a3616161416162614261636143") // {"a":"A", "b":"B", "c":"C"}

	var v strc
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, want, want)
	}
}

func TestUnmarshalStructTag2(t *testing.T) {
	type strc struct {
		A string `json:"a"`
		B string `json:"b"`
		C string `json:"c"`
	}
	want := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	cborData := hexDecode("a3616161416162614261636143") // {"a":"A", "b":"B", "c":"C"}

	var v strc
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, want, want)
	}
}

func TestUnmarshalStructTag3(t *testing.T) {
	type strc struct {
		A string `json:"x" cbor:"a"`
		B string `json:"y" cbor:"b"`
		C string `json:"z"`
	}
	want := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	cborData := hexDecode("a36161614161626142617a6143") // {"a":"A", "b":"B", "z":"C"}

	var v strc
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, v, v, want, want)
	}
}

func TestUnmarshalStructTag4(t *testing.T) {
	type strc struct {
		A string `json:"x" cbor:"a"`
		B string `json:"y" cbor:"b"`
		C string `json:"-"`
	}
	want := strc{
		A: "A",
		B: "B",
	}
	cborData := hexDecode("a3616161416162614261636143") // {"a":"A", "b":"B", "c":"C"}

	var v strc
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, v, v, want, want)
	}
}

type number uint64

func (n number) MarshalBinary() (data []byte, err error) {
	if n == 0 {
		return []byte{}, nil
	}
	data = make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(n))
	return
}

func (n *number) UnmarshalBinary(data []byte) (err error) {
	if len(data) == 0 {
		*n = 0
		return nil
	}
	if len(data) != 8 {
		return errors.New("number:UnmarshalBinary: invalid length")
	}
	*n = number(binary.BigEndian.Uint64(data))
	return
}

type stru struct {
	a, b, c string
}

func (s *stru) MarshalBinary() ([]byte, error) {
	if s.a == "" && s.b == "" && s.c == "" {
		return []byte{}, nil
	}
	return []byte(fmt.Sprintf("%s,%s,%s", s.a, s.b, s.c)), nil
}

func (s *stru) UnmarshalBinary(data []byte) (err error) {
	if len(data) == 0 {
		s.a, s.b, s.c = "", "", ""
		return nil
	}
	ss := strings.Split(string(data), ",")
	if len(ss) != 3 {
		return errors.New("stru:UnmarshalBinary: invalid element count")
	}
	s.a, s.b, s.c = ss[0], ss[1], ss[2]
	return
}

type marshalBinaryError string

func (n marshalBinaryError) MarshalBinary() (data []byte, err error) {
	return nil, errors.New(string(n))
}

func TestBinaryMarshalerUnmarshaler(t *testing.T) {
	testCases := []roundTripTest{
		{
			name:         "primitive obj",
			obj:          number(1234567890),
			wantCborData: hexDecode("4800000000499602d2"),
		},
		{
			name:         "struct obj",
			obj:          stru{a: "a", b: "b", c: "c"},
			wantCborData: hexDecode("45612C622C63"),
		},
	}
	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, testCases, em, dm)
}

func TestBinaryUnmarshalerError(t *testing.T) { //nolint:dupl
	testCases := []struct {
		name         string
		typ          reflect.Type
		cborData     []byte
		wantErrorMsg string
	}{
		{
			name:         "primitive type",
			typ:          reflect.TypeOf(number(0)),
			cborData:     hexDecode("44499602d2"),
			wantErrorMsg: "number:UnmarshalBinary: invalid length",
		},
		{
			name:         "struct type",
			typ:          reflect.TypeOf(stru{}),
			cborData:     hexDecode("47612C622C632C64"),
			wantErrorMsg: "stru:UnmarshalBinary: invalid element count",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.New(tc.typ)
			if err := Unmarshal(tc.cborData, v.Interface()); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", tc.cborData, tc.wantErrorMsg)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestBinaryMarshalerError(t *testing.T) {
	wantErrorMsg := "MarshalBinary: error"
	v := marshalBinaryError(wantErrorMsg)
	if _, err := Marshal(v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", v, wantErrorMsg)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Unmarshal(0x%x) returned error %q, want %q", v, err.Error(), wantErrorMsg)
	}
}

type number2 uint64

func (n number2) MarshalCBOR() (data []byte, err error) {
	m := map[string]uint64{"num": uint64(n)}
	return Marshal(m)
}

func (n *number2) UnmarshalCBOR(data []byte) (err error) {
	var v map[string]uint64
	if err := Unmarshal(data, &v); err != nil {
		return err
	}
	*n = number2(v["num"])
	return nil
}

type stru2 struct {
	a, b, c string
}

func (s *stru2) MarshalCBOR() ([]byte, error) {
	v := []string{s.a, s.b, s.c}
	return Marshal(v)
}

func (s *stru2) UnmarshalCBOR(data []byte) (err error) {
	var v []string
	if err := Unmarshal(data, &v); err != nil {
		return err
	}
	if len(v) > 0 {
		s.a = v[0]
	}
	if len(v) > 1 {
		s.b = v[1]
	}
	if len(v) > 2 {
		s.c = v[2]
	}
	return nil
}

type marshalCBORError string

func (n marshalCBORError) MarshalCBOR() (data []byte, err error) {
	return nil, errors.New(string(n))
}

func TestMarshalerUnmarshaler(t *testing.T) {
	testCases := []roundTripTest{
		{
			name:         "primitive obj",
			obj:          number2(1),
			wantCborData: hexDecode("a1636e756d01"),
		},
		{
			name:         "struct obj",
			obj:          stru2{a: "a", b: "b", c: "c"},
			wantCborData: hexDecode("83616161626163"),
		},
	}
	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, testCases, em, dm)
}

func TestUnmarshalerError(t *testing.T) { //nolint:dupl
	testCases := []struct {
		name         string
		typ          reflect.Type
		cborData     []byte
		wantErrorMsg string
	}{
		{
			name:         "primitive type",
			typ:          reflect.TypeOf(number2(0)),
			cborData:     hexDecode("44499602d2"),
			wantErrorMsg: "cbor: cannot unmarshal byte string into Go value of type map[string]uint64",
		},
		{
			name:         "struct type",
			typ:          reflect.TypeOf(stru2{}),
			cborData:     hexDecode("47612C622C632C64"),
			wantErrorMsg: "cbor: cannot unmarshal byte string into Go value of type []string",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.New(tc.typ)
			if err := Unmarshal(tc.cborData, v.Interface()); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", tc.cborData, tc.wantErrorMsg)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestMarshalerError(t *testing.T) {
	wantErrorMsg := "MarshalCBOR: error"
	v := marshalCBORError(wantErrorMsg)
	if _, err := Marshal(v); err == nil {
		t.Errorf("Marshal(%+v) didn't return an error, want error msg %q", v, wantErrorMsg)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Marshal(%+v) returned error %q, want %q", v, err.Error(), wantErrorMsg)
	}
}

// Found at https://github.com/oasislabs/oasis-core/blob/master/go/common/cbor/cbor_test.go
func TestOutOfMem1(t *testing.T) {
	cborData := []byte("\x9b\x00\x00000000")
	var f []byte
	if err := Unmarshal(cborData, &f); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	}
}

// Found at https://github.com/oasislabs/oasis-core/blob/master/go/common/cbor/cbor_test.go
func TestOutOfMem2(t *testing.T) {
	cborData := []byte("\x9b\x00\x00\x81112233")
	var f []byte
	if err := Unmarshal(cborData, &f); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	}
}

// Found at https://github.com/cose-wg/Examples/tree/master/RFC8152
func TestCOSEExamples(t *testing.T) {
	cborData := [][]byte{
		hexDecode("D8608443A10101A1054C02D1F7E6F26C43D4868D87CE582464F84D913BA60A76070A9A48F26E97E863E2852948658F0811139868826E89218A75715B818440A101225818DBD43C4E9D719C27C6275C67D628D493F090593DB8218F11818344A1013818A220A401022001215820B2ADD44368EA6D641F9CA9AF308B4079AEB519F11E9B8A55A600B21233E86E6822F40458246D65726961646F632E6272616E64796275636B406275636B6C616E642E6578616D706C6540"),
		hexDecode("D8628440A054546869732069732074686520636F6E74656E742E818343A10126A1044231315840E2AEAFD40D69D19DFE6E52077C5D7FF4E408282CBEFB5D06CBF414AF2E19D982AC45AC98B8544C908B4507DE1E90B717C3D34816FE926A2B98F53AFD2FA0F30A"),
		hexDecode("D8628440A054546869732069732074686520636F6E74656E742E828343A10126A1044231315840E2AEAFD40D69D19DFE6E52077C5D7FF4E408282CBEFB5D06CBF414AF2E19D982AC45AC98B8544C908B4507DE1E90B717C3D34816FE926A2B98F53AFD2FA0F30A8344A1013823A104581E62696C626F2E62616767696E7340686F626269746F6E2E6578616D706C65588400A2D28A7C2BDB1587877420F65ADF7D0B9A06635DD1DE64BB62974C863F0B160DD2163734034E6AC003B01E8705524C5C4CA479A952F0247EE8CB0B4FB7397BA08D009E0C8BF482270CC5771AA143966E5A469A09F613488030C5B07EC6D722E3835ADB5B2D8C44E95FFB13877DD2582866883535DE3BB03D01753F83AB87BB4F7A0297"),
		hexDecode("D8628440A1078343A10126A10442313158405AC05E289D5D0E1B0A7F048A5D2B643813DED50BC9E49220F4F7278F85F19D4A77D655C9D3B51E805A74B099E1E085AACD97FC29D72F887E8802BB6650CCEB2C54546869732069732074686520636F6E74656E742E818343A10126A1044231315840E2AEAFD40D69D19DFE6E52077C5D7FF4E408282CBEFB5D06CBF414AF2E19D982AC45AC98B8544C908B4507DE1E90B717C3D34816FE926A2B98F53AFD2FA0F30A"),
		hexDecode("D8628456A2687265736572766564F40281687265736572766564A054546869732069732074686520636F6E74656E742E818343A10126A10442313158403FC54702AA56E1B2CB20284294C9106A63F91BAC658D69351210A031D8FC7C5FF3E4BE39445B1A3E83E1510D1ACA2F2E8A7C081C7645042B18ABA9D1FAD1BD9C"),
		hexDecode("D28443A10126A10442313154546869732069732074686520636F6E74656E742E58408EB33E4CA31D1C465AB05AAC34CC6B23D58FEF5C083106C4D25A91AEF0B0117E2AF9A291AA32E14AB834DC56ED2A223444547E01F11D3B0916E5A4C345CACB36"),
		hexDecode("D8608443A10101A1054CC9CF4DF2FE6C632BF788641358247ADBE2709CA818FB415F1E5DF66F4E1A51053BA6D65A1A0C52A357DA7A644B8070A151B0818344A1013818A220A40102200121582098F50A4FF6C05861C8860D13A638EA56C3F5AD7590BBFBF054E1C7B4D91D628022F50458246D65726961646F632E6272616E64796275636B406275636B6C616E642E6578616D706C6540"),
		hexDecode("D8608443A1010AA1054D89F52F65A1C580933B5261A76C581C753548A19B1307084CA7B2056924ED95F2E3B17006DFE931B687B847818343A10129A2335061616262636364646565666667676868044A6F75722D73656372657440"),
		hexDecode("D8608443A10101A2054CC9CF4DF2FE6C632BF7886413078344A1013823A104581E62696C626F2E62616767696E7340686F626269746F6E2E6578616D706C65588400929663C8789BB28177AE28467E66377DA12302D7F9594D2999AFA5DFA531294F8896F2B6CDF1740014F4C7F1A358E3A6CF57F4ED6FB02FCF8F7AA989F5DFD07F0700A3A7D8F3C604BA70FA9411BD10C2591B483E1D2C31DE003183E434D8FBA18F17A4C7E3DFA003AC1CF3D30D44D2533C4989D3AC38C38B71481CC3430C9D65E7DDFF58247ADBE2709CA818FB415F1E5DF66F4E1A51053BA6D65A1A0C52A357DA7A644B8070A151B0818344A1013818A220A40102200121582098F50A4FF6C05861C8860D13A638EA56C3F5AD7590BBFBF054E1C7B4D91D628022F50458246D65726961646F632E6272616E64796275636B406275636B6C616E642E6578616D706C6540"),
		hexDecode("D8608443A10101A1054C02D1F7E6F26C43D4868D87CE582464F84D913BA60A76070A9A48F26E97E863E28529D8F5335E5F0165EEE976B4A5F6C6F09D818344A101381FA3225821706572656772696E2E746F6F6B407475636B626F726F7567682E6578616D706C650458246D65726961646F632E6272616E64796275636B406275636B6C616E642E6578616D706C6535420101581841E0D76F579DBD0D936A662D54D8582037DE2E366FDE1C62"),
		hexDecode("D08343A1010AA1054D89F52F65A1C580933B5261A78C581C5974E1B99A3A4CC09A659AA2E9E7FFF161D38CE71CB45CE460FFB569"),
		hexDecode("D08343A1010AA1064261A7581C252A8911D465C125B6764739700F0141ED09192DE139E053BD09ABCA"),
		hexDecode("D8618543A1010FA054546869732069732074686520636F6E74656E742E489E1226BA1F81B848818340A20125044A6F75722D73656372657440"),
		hexDecode("D8618543A10105A054546869732069732074686520636F6E74656E742E582081A03448ACD3D305376EAA11FB3FE416A955BE2CBE7EC96F012C994BC3F16A41818344A101381AA3225821706572656772696E2E746F6F6B407475636B626F726F7567682E6578616D706C650458246D65726961646F632E6272616E64796275636B406275636B6C616E642E6578616D706C653558404D8553E7E74F3C6A3A9DD3EF286A8195CBF8A23D19558CCFEC7D34B824F42D92BD06BD2C7F0271F0214E141FB779AE2856ABF585A58368B017E7F2A9E5CE4DB540"),
		hexDecode("D8618543A1010EA054546869732069732074686520636F6E74656E742E4836F5AFAF0BAB5D43818340A2012404582430313863306165352D346439622D343731622D626664362D6565663331346263373033375818711AB0DC2FC4585DCE27EFFA6781C8093EBA906F227B6EB0"),
		hexDecode("D8618543A10105A054546869732069732074686520636F6E74656E742E5820BF48235E809B5C42E995F2B7D5FA13620E7ED834E337F6AA43DF161E49E9323E828344A101381CA220A4010220032158420043B12669ACAC3FD27898FFBA0BCD2E6C366D53BC4DB71F909A759304ACFB5E18CDC7BA0B13FF8C7636271A6924B1AC63C02688075B55EF2D613574E7DC242F79C322F504581E62696C626F2E62616767696E7340686F626269746F6E2E6578616D706C655828339BC4F79984CDC6B3E6CE5F315A4C7D2B0AC466FCEA69E8C07DFBCA5BB1F661BC5F8E0DF9E3EFF58340A2012404582430313863306165352D346439622D343731622D626664362D65656633313462633730333758280B2C7CFCE04E98276342D6476A7723C090DFDD15F9A518E7736549E998370695E6D6A83B4AE507BB"),
		hexDecode("D18443A1010FA054546869732069732074686520636F6E74656E742E48726043745027214F"),
	}
	for _, d := range cborData {
		var v interface{}
		if err := Unmarshal(d, &v); err != nil {
			t.Errorf("Unmarshal(0x%x) returned error %v", d, err)
		}
	}
}

func TestUnmarshalStructKeyAsIntError(t *testing.T) {
	type T1 struct {
		F1 int `cbor:"1,keyasint"`
	}
	cborData := hexDecode("a13bffffffffffffffff01") // {1: -18446744073709551616}
	var v T1
	if err := Unmarshal(cborData, &v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
	} else if !strings.Contains(err.Error(), "cannot unmarshal") {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), "cannot unmarshal")
	}
}

func TestUnmarshalArrayToStruct(t *testing.T) {
	type T struct {
		_ struct{} `cbor:",toarray"`
		A int
		B int
		C int
	}
	testCases := []struct {
		name     string
		cborData []byte
	}{
		{"definite length array", hexDecode("83010203")},
		{"indefinite length array", hexDecode("9f010203ff")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v T
			if err := Unmarshal(tc.cborData, &v); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
			}
		})
	}
}

func TestUnmarshalArrayToStructNoToArrayOptionError(t *testing.T) {
	type T struct {
		A int
		B int
		C int
	}
	cborData := hexDecode("8301020383010203")
	var v1 T
	wantT := T{}
	dec := NewDecoder(bytes.NewReader(cborData))
	if err := dec.Decode(&v1); err == nil {
		t.Errorf("Decode(%+v) didn't return an error", v1)
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Decode(%+v) returned wrong error type %T, want (*UnmarshalTypeError)", v1, err)
	} else if !strings.Contains(err.Error(), "cannot unmarshal") {
		t.Errorf("Decode(%+v) returned error %q, want error containing %q", err.Error(), v1, "cannot unmarshal")
	}
	if !reflect.DeepEqual(v1, wantT) {
		t.Errorf("Decode() = %+v (%T), want %+v (%T)", v1, v1, wantT, wantT)
	}
	var v2 []int
	want := []int{1, 2, 3}
	if err := dec.Decode(&v2); err != nil {
		t.Errorf("Decode() returned error %v", err)
	}
	if !reflect.DeepEqual(v2, want) {
		t.Errorf("Decode() = %+v (%T), want %+v (%T)", v2, v2, want, want)
	}
}

func TestUnmarshalNonArrayDataToStructToArray(t *testing.T) {
	type T struct {
		_ struct{} `cbor:",toarray"`
		A int
		B int
		C int
	}
	testCases := []struct {
		name     string
		cborData []byte
	}{
		{"CBOR positive int", hexDecode("00")},                        // 0
		{"CBOR negative int", hexDecode("20")},                        // -1
		{"CBOR byte string", hexDecode("4401020304")},                 // h`01020304`
		{"CBOR text string", hexDecode("7f657374726561646d696e67ff")}, // streaming
		{"CBOR map", hexDecode("a3614101614202614303")},               // {"A": 1, "B": 2, "C": 3}
		{"CBOR bool", hexDecode("f5")},                                // true
		{"CBOR float", hexDecode("fa7f7fffff")},                       // 3.4028234663852886e+38
	}
	wantT := T{}
	wantErrorMsg := "cannot unmarshal"
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v T
			if err := Unmarshal(tc.cborData, &v); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error", tc.cborData)
			} else if _, ok := err.(*UnmarshalTypeError); !ok {
				t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", tc.cborData, err)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", tc.cborData, err.Error(), wantErrorMsg)
			}
			if !reflect.DeepEqual(v, wantT) {
				t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", tc.cborData, v, v, wantT, wantT)
			}
		})
	}
}

func TestUnmarshalArrayToStructWrongSizeError(t *testing.T) {
	type T struct {
		_ struct{} `cbor:",toarray"`
		A int
		B int
	}
	cborData := hexDecode("8301020383010203")
	var v1 T
	wantT := T{}
	dec := NewDecoder(bytes.NewReader(cborData))
	if err := dec.Decode(&v1); err == nil {
		t.Errorf("Decode(%+v) didn't return an error", v1)
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Decode(%+v) returned wrong error type %T, want (*UnmarshalTypeError)", v1, err)
	} else if !strings.Contains(err.Error(), "cannot unmarshal") {
		t.Errorf("Decode(%+v) returned error %q, want error containing %q", v1, err.Error(), "cannot unmarshal")
	}
	if !reflect.DeepEqual(v1, wantT) {
		t.Errorf("Decode() = %+v (%T), want %+v (%T)", v1, v1, wantT, wantT)
	}
	var v2 []int
	want := []int{1, 2, 3}
	if err := dec.Decode(&v2); err != nil {
		t.Errorf("Decode() returned error %v", err)
	}
	if !reflect.DeepEqual(v2, want) {
		t.Errorf("Decode() = %+v (%T), want %+v (%T)", v2, v2, want, want)
	}
}

func TestUnmarshalArrayToStructWrongFieldTypeError(t *testing.T) {
	type T struct {
		_ struct{} `cbor:",toarray"`
		A int
		B string
		C int
	}
	testCases := []struct {
		name         string
		cborData     []byte
		wantErrorMsg string
		wantV        interface{}
	}{
		// [1, 2, 3]
		{"wrong field type", hexDecode("83010203"), "cannot unmarshal", T{A: 1, C: 3}},
		// [1, 0xfe, 3]
		{"invalid UTF-8 string", hexDecode("830161fe03"), invalidUTF8ErrorMsg, T{A: 1, C: 3}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v T
			if err := Unmarshal(tc.cborData, &v); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error", tc.cborData)
			} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", tc.cborData, err.Error(), tc.wantErrorMsg)
			}
			if !reflect.DeepEqual(v, tc.wantV) {
				t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", tc.cborData, v, v, tc.wantV, tc.wantV)
			}
		})
	}
}

func TestUnmarshalArrayToStructCannotSetEmbeddedPointerError(t *testing.T) {
	type (
		s1 struct {
			x int //nolint:unused,structcheck
			X int
		}
		S2 struct {
			y int //nolint:unused,structcheck
			Y int
		}
		S struct {
			_ struct{} `cbor:",toarray"`
			*s1
			*S2
		}
	)
	cborData := []byte{0x82, 0x02, 0x04} // [2, 4]
	const wantErrorMsg = "cannot set embedded pointer to unexported struct"
	wantV := S{S2: &S2{Y: 4}}
	var v S
	err := Unmarshal(cborData, &v)
	if err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want error %q", cborData, wantErrorMsg)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(v, wantV) {
		t.Errorf("Decode() = %+v (%T), want %+v (%T)", v, v, wantV, wantV)
	}
}

func TestUnmarshalIntoSliceError(t *testing.T) {
	cborData := []byte{0x83, 0x61, 0x61, 0x61, 0xfe, 0x61, 0x62} // ["a", 0xfe, "b"]
	wantErrorMsg := invalidUTF8ErrorMsg
	var want interface{}

	// Unmarshal CBOR array into Go empty interface.
	var v1 interface{}
	want = []interface{}{"a", interface{}(nil), "b"}
	if err := Unmarshal(cborData, &v1); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", cborData, wantErrorMsg)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Unmarshal(0x%x) returned error %q, want %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(v1, want) {
		t.Errorf("Unmarshal(0x%x) = %v, want %v", cborData, v1, want)
	}

	// Unmarshal CBOR array into Go slice.
	var v2 []string
	want = []string{"a", "", "b"}
	if err := Unmarshal(cborData, &v2); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", cborData, wantErrorMsg)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Unmarshal(0x%x) returned error %q, want %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(v2, want) {
		t.Errorf("Unmarshal(0x%x) = %v, want %v", cborData, v2, want)
	}

	// Unmarshal CBOR array into Go array.
	var v3 [3]string
	want = [3]string{"a", "", "b"}
	if err := Unmarshal(cborData, &v3); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", cborData, wantErrorMsg)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Unmarshal(0x%x) returned error %q, want %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(v3, want) {
		t.Errorf("Unmarshal(0x%x) = %v, want %v", cborData, v3, want)
	}

	// Unmarshal CBOR array into populated Go slice.
	v4 := []string{"hello", "to", "you"}
	want = []string{"a", "to", "b"}
	if err := Unmarshal(cborData, &v4); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", cborData, wantErrorMsg)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Unmarshal(0x%x) returned error %q, want %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(v4, want) {
		t.Errorf("Unmarshal(0x%x) = %v, want %v", cborData, v4, want)
	}
}

func TestUnmarshalIntoMapError(t *testing.T) {
	cborData := [][]byte{
		{0xa3, 0x61, 0x61, 0x61, 0x41, 0x61, 0xfe, 0x61, 0x43, 0x61, 0x62, 0x61, 0x42}, // {"a":"A", 0xfe: "C", "b":"B"}
		{0xa3, 0x61, 0x61, 0x61, 0x41, 0x61, 0x63, 0x61, 0xfe, 0x61, 0x62, 0x61, 0x42}, // {"a":"A", "c": 0xfe, "b":"B"}
	}
	wantErrorMsg := invalidUTF8ErrorMsg
	var want interface{}

	for _, data := range cborData {
		// Unmarshal CBOR map into Go empty interface.
		var v1 interface{}
		want = map[interface{}]interface{}{"a": "A", "b": "B"}
		if err := Unmarshal(data, &v1); err == nil {
			t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", data, wantErrorMsg)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("Unmarshal(0x%x) returned error %q, want %q", data, err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(v1, want) {
			t.Errorf("Unmarshal(0x%x) = %v, want %v", data, v1, want)
		}

		// Unmarshal CBOR map into Go map[interface{}]interface{}.
		var v2 map[interface{}]interface{}
		want = map[interface{}]interface{}{"a": "A", "b": "B"}
		if err := Unmarshal(data, &v2); err == nil {
			t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", data, wantErrorMsg)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("Unmarshal(0x%x) returned error %q, want %q", data, err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(v2, want) {
			t.Errorf("Unmarshal(0x%x) = %v, want %v", data, v2, want)
		}

		// Unmarshal CBOR array into Go map[string]string.
		var v3 map[string]string
		want = map[string]string{"a": "A", "b": "B"}
		if err := Unmarshal(data, &v3); err == nil {
			t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", data, wantErrorMsg)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("Unmarshal(0x%x) returned error %q, want %q", data, err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(v3, want) {
			t.Errorf("Unmarshal(0x%x) = %v, want %v", data, v3, want)
		}

		// Unmarshal CBOR array into populated Go map[string]string.
		v4 := map[string]string{"c": "D"}
		want = map[string]string{"a": "A", "b": "B", "c": "D"}
		if err := Unmarshal(data, &v4); err == nil {
			t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", data, wantErrorMsg)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("Unmarshal(0x%x) returned error %q, want %q", data, err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(v4, want) {
			t.Errorf("Unmarshal(0x%x) = %v, want %v", data, v4, want)
		}
	}
}

func TestUnmarshalDeepNesting(t *testing.T) {
	// Construct this object rather than embed such a large constant in the code
	type TestNode struct {
		Value int
		Child *TestNode
	}
	n := &TestNode{Value: 0}
	root := n
	for i := 0; i < 65534; i++ {
		child := &TestNode{Value: i}
		n.Child = child
		n = child
	}
	em, err := EncOptions{}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned error %v", err)
	}
	cborData, err := em.Marshal(root)
	if err != nil {
		t.Errorf("Marshal() deeply nested object returned error %v", err)
	}

	// Try unmarshal it
	dm, err := DecOptions{MaxNestedLevels: 65535}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned error %v", err)
	}
	var readback TestNode
	err = dm.Unmarshal(cborData, &readback)
	if err != nil {
		t.Errorf("Unmarshal() of deeply nested object returned error: %v", err)
	}
	if !reflect.DeepEqual(root, &readback) {
		t.Errorf("Unmarshal() of deeply nested object did not match\nGot: %#v\n Want: %#v\n",
			&readback, root)
	}
}

func TestStructToArrayError(t *testing.T) {
	type coseHeader struct {
		Alg int    `cbor:"1,keyasint,omitempty"`
		Kid []byte `cbor:"4,keyasint,omitempty"`
		IV  []byte `cbor:"5,keyasint,omitempty"`
	}
	type nestedCWT struct {
		_           struct{} `cbor:",toarray"`
		Protected   []byte
		Unprotected coseHeader
		Ciphertext  []byte
	}
	for _, tc := range []struct {
		cborData     []byte
		wantErrorMsg string
	}{
		// [-17, [-17, -17], -17]
		{hexDecode("9f3082303030ff"), "cbor: cannot unmarshal negative integer into Go struct field cbor.nestedCWT.Protected of type []uint8"},
		// [[], [], ["\x930000", -17]]
		{hexDecode("9f9fff9fff9f65933030303030ffff"), "cbor: cannot unmarshal array into Go struct field cbor.nestedCWT.Unprotected of type cbor.coseHeader (cannot decode CBOR array to struct without toarray option)"},
	} {
		var v nestedCWT
		if err := Unmarshal(tc.cborData, &v); err == nil {
			t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", tc.cborData, tc.wantErrorMsg)
		} else if err.Error() != tc.wantErrorMsg {
			t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
		}
	}
}

func TestStructKeyAsIntError(t *testing.T) {
	type claims struct {
		Iss string  `cbor:"1,keyasint"`
		Sub string  `cbor:"2,keyasint"`
		Aud string  `cbor:"3,keyasint"`
		Exp float64 `cbor:"4,keyasint"`
		Nbf float64 `cbor:"5,keyasint"`
		Iat float64 `cbor:"6,keyasint"`
		Cti []byte  `cbor:"7,keyasint"`
	}
	cborData := hexDecode("bf0783e662f03030ff") // {7: [simple(6), "\xF00", -17]}
	wantErrorMsg := invalidUTF8ErrorMsg
	wantV := claims{Cti: []byte{6, 0, 0}}
	var v claims
	if err := Unmarshal(cborData, &v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", cborData, wantErrorMsg)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Unmarshal(0x%x) returned error %q, want %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(v, wantV) {
		t.Errorf("Unmarshal(0x%x) = %v, want %v", cborData, v, wantV)
	}
}

func TestUnmarshalToNotNilInterface(t *testing.T) {
	cborData := hexDecode("83010203") // []uint64{1, 2, 3}
	s := "hello"                      //nolint:goconst
	var v interface{} = s             // Unmarshal() sees v as type interface{} and sets CBOR data as default Go type.  s is unmodified.  Same behavior as encoding/json.
	wantV := []interface{}{uint64(1), uint64(2), uint64(3)}
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if !reflect.DeepEqual(v, wantV) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, wantV, wantV)
	} else if s != "hello" {
		t.Errorf("Unmarshal(0x%x) modified s %q", cborData, s)
	}
}

func TestDecOptions(t *testing.T) {
	opts1 := DecOptions{
		DupMapKey:         DupMapKeyEnforcedAPF,
		TimeTag:           DecTagRequired,
		MaxNestedLevels:   100,
		MaxArrayElements:  102,
		MaxMapPairs:       101,
		IndefLength:       IndefLengthForbidden,
		TagsMd:            TagsForbidden,
		IntDec:            IntDecConvertSigned,
		ExtraReturnErrors: ExtraDecErrorUnknownField,
		UTF8:              UTF8DecodeInvalid,
	}
	dm, err := opts1.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned an error %v", err)
	} else {
		opts2 := dm.DecOptions()
		if !reflect.DeepEqual(opts1, opts2) {
			t.Errorf("DecOptions->DecMode->DecOptions returned different values: %v, %v", opts1, opts2)
		}
	}
}

type roundTripTest struct {
	name         string
	obj          interface{}
	wantCborData []byte
}

func testRoundTrip(t *testing.T, testCases []roundTripTest, em EncMode, dm DecMode) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := em.Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}
			v := reflect.New(reflect.TypeOf(tc.obj))
			if err := dm.Unmarshal(b, v.Interface()); err != nil {
				t.Errorf("Unmarshal() returned error %v", err)
			}
			if !reflect.DeepEqual(tc.obj, v.Elem().Interface()) {
				t.Errorf("Marshal-Unmarshal returned different values: %v, %v", tc.obj, v.Elem().Interface())
			}
		})
	}
}

func TestDecModeInvalidTimeTag(t *testing.T) {
	wantErrorMsg := "cbor: invalid TimeTag 101"
	_, err := DecOptions{TimeTag: 101}.DecMode()
	if err == nil {
		t.Errorf("DecMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("DecMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestDecModeInvalidDuplicateMapKey(t *testing.T) {
	wantErrorMsg := "cbor: invalid DupMapKey 101"
	_, err := DecOptions{DupMapKey: 101}.DecMode()
	if err == nil {
		t.Errorf("DecMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("DecMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestDecModeDefaultMaxNestedLevel(t *testing.T) {
	dm, err := DecOptions{}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned error %v", err)
	} else {
		maxNestedLevels := dm.DecOptions().MaxNestedLevels
		if maxNestedLevels != 32 {
			t.Errorf("DecOptions().MaxNestedLevels = %d, want %v", maxNestedLevels, 32)
		}
	}
}

func TestDecModeInvalidMaxNestedLevel(t *testing.T) {
	testCases := []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "MaxNestedLevels < 4",
			opts:         DecOptions{MaxNestedLevels: 1},
			wantErrorMsg: "cbor: invalid MaxNestedLevels 1 (range is [4, 65535])",
		},
		{
			name:         "MaxNestedLevels > 65535",
			opts:         DecOptions{MaxNestedLevels: 65536},
			wantErrorMsg: "cbor: invalid MaxNestedLevels 65536 (range is [4, 65535])",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeDefaultMaxMapPairs(t *testing.T) {
	dm, err := DecOptions{}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned error %v", err)
	} else {
		maxMapPairs := dm.DecOptions().MaxMapPairs
		if maxMapPairs != defaultMaxMapPairs {
			t.Errorf("DecOptions().MaxMapPairs = %d, want %v", maxMapPairs, defaultMaxMapPairs)
		}
	}
}

func TestDecModeInvalidMaxMapPairs(t *testing.T) {
	testCases := []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "MaxMapPairs < 16",
			opts:         DecOptions{MaxMapPairs: 1},
			wantErrorMsg: "cbor: invalid MaxMapPairs 1 (range is [16, 2147483647])",
		},
		{
			name:         "MaxMapPairs > 2147483647",
			opts:         DecOptions{MaxMapPairs: 2147483648},
			wantErrorMsg: "cbor: invalid MaxMapPairs 2147483648 (range is [16, 2147483647])",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeDefaultMaxArrayElements(t *testing.T) {
	dm, err := DecOptions{}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned error %v", err)
	} else {
		maxArrayElements := dm.DecOptions().MaxArrayElements
		if maxArrayElements != defaultMaxArrayElements {
			t.Errorf("DecOptions().MaxArrayElementsr = %d, want %v", maxArrayElements, defaultMaxArrayElements)
		}
	}
}

func TestDecModeInvalidMaxArrayElements(t *testing.T) {
	testCases := []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "MaxArrayElements < 16",
			opts:         DecOptions{MaxArrayElements: 1},
			wantErrorMsg: "cbor: invalid MaxArrayElements 1 (range is [16, 2147483647])",
		},
		{
			name:         "MaxArrayElements > 2147483647",
			opts:         DecOptions{MaxArrayElements: 2147483648},
			wantErrorMsg: "cbor: invalid MaxArrayElements 2147483648 (range is [16, 2147483647])",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidIndefiniteLengthMode(t *testing.T) {
	wantErrorMsg := "cbor: invalid IndefLength 101"
	_, err := DecOptions{IndefLength: 101}.DecMode()
	if err == nil {
		t.Errorf("DecMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("DecMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestDecModeInvalidTagsMode(t *testing.T) {
	wantErrorMsg := "cbor: invalid TagsMd 101"
	_, err := DecOptions{TagsMd: 101}.DecMode()
	if err == nil {
		t.Errorf("DecMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("DecMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestUnmarshalStructKeyAsIntNumError(t *testing.T) {
	type T1 struct {
		F1 int `cbor:"a,keyasint"`
	}
	type T2 struct {
		F1 int `cbor:"-18446744073709551616,keyasint"`
	}
	testCases := []struct {
		name         string
		cborData     []byte
		obj          interface{}
		wantErrorMsg string
	}{
		{
			name:         "string as key",
			cborData:     hexDecode("a1616101"),
			obj:          T1{},
			wantErrorMsg: "cbor: failed to parse field name \"a\" to int",
		},
		{
			name:         "out of range int as key",
			cborData:     hexDecode("a13bffffffffffffffff01"),
			obj:          T2{},
			wantErrorMsg: "cbor: failed to parse field name \"-18446744073709551616\" to int",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.New(reflect.TypeOf(tc.obj))
			err := Unmarshal(tc.cborData, v.Interface())
			if err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error %q", tc.cborData, tc.wantErrorMsg)
			} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) error %v, want %v", tc.cborData, err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestUnmarshalEmptyMapWithDupMapKeyOpt(t *testing.T) {
	testCases := []struct {
		name     string
		cborData []byte
		wantV    interface{}
	}{
		{
			name:     "empty map",
			cborData: hexDecode("a0"),
			wantV:    map[interface{}]interface{}{},
		},
		{
			name:     "indefinite empty map",
			cborData: hexDecode("bfff"),
			wantV:    map[interface{}]interface{}{},
		},
	}

	dm, err := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned error %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v interface{}
			if err := dm.Unmarshal(tc.cborData, &v); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
			}
			if !reflect.DeepEqual(v, tc.wantV) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborData, v, v, tc.wantV, tc.wantV)
			}
		})
	}
}

func TestUnmarshalDupMapKeyToEmptyInterface(t *testing.T) {
	cborData := hexDecode("a6616161416162614261636143616161466164614461656145") // {"a": "A", "b": "B", "c": "C", "a": "F", "d": "D", "e": "E"}

	// Duplicate key overwrites previous value (default).
	wantV := map[interface{}]interface{}{"a": "F", "b": "B", "c": "C", "d": "D", "e": "E"}
	var v interface{}
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(v, wantV) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, wantV, wantV)
	}

	// Duplicate key triggers error.
	wantV = map[interface{}]interface{}{"a": nil, "b": "B", "c": "C"}
	wantErrorMsg := "cbor: found duplicate map key \"a\" at map element index 3"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	var v2 interface{}
	if err := dm.Unmarshal(cborData, &v2); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else if _, ok := err.(*DupMapKeyError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*DupMapKeyError)", cborData, err)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(v2, wantV) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v2, v2, wantV, wantV)
	}
}

func TestStreamDupMapKeyToEmptyInterface(t *testing.T) {
	cborData := hexDecode("a6616161416162614261636143616161466164614461656145") // map with duplicate key "c": {"a": "A", "b": "B", "c": "C", "a": "F", "d": "D", "e": "E"}

	var b []byte
	for i := 0; i < 3; i++ {
		b = append(b, cborData...)
	}

	// Duplicate key overwrites previous value (default).
	wantV := map[interface{}]interface{}{"a": "F", "b": "B", "c": "C", "d": "D", "e": "E"}
	dec := NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var v1 interface{}
		if err := dec.Decode(&v1); err != nil {
			t.Errorf("Decode() returned error %v", err)
		}
		if !reflect.DeepEqual(v1, wantV) {
			t.Errorf("Decode() = %v (%T), want %v (%T)", v1, v1, wantV, wantV)
		}
	}
	var v interface{}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}

	// Duplicate key triggers error.
	wantV = map[interface{}]interface{}{"a": nil, "b": "B", "c": "C"}
	wantErrorMsg := "cbor: found duplicate map key \"a\" at map element index 3"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	dec = dm.NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var v2 interface{}
		if err := dec.Decode(&v2); err == nil {
			t.Errorf("Decode() didn't return an error")
		} else if _, ok := err.(*DupMapKeyError); !ok {
			t.Errorf("Decode() returned wrong error type %T, want (*DupMapKeyError)", err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("Decode() returned error %q, want error containing %q", err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(v2, wantV) {
			t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v2, v2, wantV, wantV)
		}
	}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}
}

func TestUnmarshalDupMapKeyToEmptyMap(t *testing.T) {
	cborData := hexDecode("a6616161416162614261636143616161466164614461656145") // {"a": "A", "b": "B", "c": "C", "a": "F", "d": "D", "e": "E"}

	// Duplicate key overwrites previous value (default).
	wantM := map[string]string{"a": "F", "b": "B", "c": "C", "d": "D", "e": "E"}
	var m map[string]string
	if err := Unmarshal(cborData, &m); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(m, wantM) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, m, m, wantM, wantM)
	}

	// Duplicate key triggers error.
	wantM = map[string]string{"a": "", "b": "B", "c": "C"}
	wantErrorMsg := "cbor: found duplicate map key \"a\" at map element index 3"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	var m2 map[string]string
	if err := dm.Unmarshal(cborData, &m2); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else if _, ok := err.(*DupMapKeyError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*DupMapKeyError)", cborData, err)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(m2, wantM) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, m2, m2, wantM, wantM)
	}
}

func TestStreamDupMapKeyToEmptyMap(t *testing.T) {
	cborData := hexDecode("a6616161416162614261636143616161466164614461656145") // {"a": "A", "b": "B", "c": "C", "a": "F", "d": "D", "e": "E"}

	var b []byte
	for i := 0; i < 3; i++ {
		b = append(b, cborData...)
	}

	// Duplicate key overwrites previous value (default).
	wantM := map[string]string{"a": "F", "b": "B", "c": "C", "d": "D", "e": "E"}
	dec := NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var m1 map[string]string
		if err := dec.Decode(&m1); err != nil {
			t.Errorf("Decode() returned error %v", err)
		}
		if !reflect.DeepEqual(m1, wantM) {
			t.Errorf("Decode() = %v (%T), want %v (%T)", m1, m1, wantM, wantM)
		}
	}
	var v interface{}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}

	// Duplicate key triggers error.
	wantM = map[string]string{"a": "", "b": "B", "c": "C"}
	wantErrorMsg := "cbor: found duplicate map key \"a\" at map element index 3"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	dec = dm.NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var m2 map[string]string
		if err := dec.Decode(&m2); err == nil {
			t.Errorf("Decode() didn't return an error")
		} else if _, ok := err.(*DupMapKeyError); !ok {
			t.Errorf("Decode() returned wrong error type %T, want (*DupMapKeyError)", err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("Decode() returned error %q, want error containing %q", err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(m2, wantM) {
			t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, m2, m2, wantM, wantM)
		}
	}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}
}

func TestUnmarshalDupMapKeyToNotEmptyMap(t *testing.T) {
	cborData := hexDecode("a6616161416162614261636143616161466164614461656145") // {"a": "A", "b": "B", "c": "C", "a": "F", "d": "D", "e": "E"}

	// Duplicate key overwrites previous value (default).
	m := map[string]string{"a": "Z", "b": "Z", "c": "Z", "d": "Z", "e": "Z", "f": "Z"}
	wantM := map[string]string{"a": "F", "b": "B", "c": "C", "d": "D", "e": "E", "f": "Z"}
	if err := Unmarshal(cborData, &m); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(m, wantM) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, m, m, wantM, wantM)
	}

	// Duplicate key triggers error.
	m2 := map[string]string{"a": "Z", "b": "Z", "c": "Z", "d": "Z", "e": "Z", "f": "Z"}
	wantM = map[string]string{"a": "", "b": "B", "c": "C", "d": "Z", "e": "Z", "f": "Z"}
	wantErrorMsg := "cbor: found duplicate map key \"a\" at map element index 3"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	if err := dm.Unmarshal(cborData, &m2); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else if _, ok := err.(*DupMapKeyError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*DupMapKeyError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(m2, wantM) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, m2, m2, wantM, wantM)
	}
}

func TestStreamDupMapKeyToNotEmptyMap(t *testing.T) {
	cborData := hexDecode("a6616161416162614261636143616161466164614461656145") // {"a": "A", "b": "B", "c": "C", "a": "F", "d": "D", "e": "E"}

	var b []byte
	for i := 0; i < 3; i++ {
		b = append(b, cborData...)
	}

	// Duplicate key overwrites previous value (default).
	wantM := map[string]string{"a": "F", "b": "B", "c": "C", "d": "D", "e": "E", "f": "Z"}
	dec := NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		m1 := map[string]string{"a": "Z", "b": "Z", "c": "Z", "d": "Z", "e": "Z", "f": "Z"}
		if err := dec.Decode(&m1); err != nil {
			t.Errorf("Decode() returned error %v", err)
		}
		if !reflect.DeepEqual(m1, wantM) {
			t.Errorf("Decode() = %v (%T), want %v (%T)", m1, m1, wantM, wantM)
		}
	}
	var v interface{}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}

	// Duplicate key triggers error.
	wantM = map[string]string{"a": "", "b": "B", "c": "C", "d": "Z", "e": "Z", "f": "Z"}
	wantErrorMsg := "cbor: found duplicate map key \"a\" at map element index 3"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	dec = dm.NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		m2 := map[string]string{"a": "Z", "b": "Z", "c": "Z", "d": "Z", "e": "Z", "f": "Z"}
		if err := dec.Decode(&m2); err == nil {
			t.Errorf("Decode() didn't return an error")
		} else if _, ok := err.(*DupMapKeyError); !ok {
			t.Errorf("Decode() returned wrong error type %T, want (*DupMapKeyError)", err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("Decode() returned error %q, want error containing %q", err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(m2, wantM) {
			t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, m2, m2, wantM, wantM)
		}
	}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}
}

func TestUnmarshalDupMapKeyToStruct(t *testing.T) {
	type s struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
		D string `cbor:"d"`
		E string `cbor:"e"`
	}
	cborData := hexDecode("a6616161416162614261636143616161466164614461656145") // {"a": "A", "b": "B", "c": "C", "a": "F", "d": "D", "e": "E"}

	// Duplicate key doesn't overwrite previous value (default).
	wantS := s{A: "A", B: "B", C: "C", D: "D", E: "E"}
	var s1 s
	if err := Unmarshal(cborData, &s1); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(s1, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s1, s1, wantS, wantS)
	}

	// Duplicate key triggers error.
	wantS = s{A: "A", B: "B", C: "C"}
	wantErrorMsg := "cbor: found duplicate map key \"a\" at map element index 3"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	var s2 s
	if err := dm.Unmarshal(cborData, &s2); err == nil {
		t.Errorf("Unmarshal(0x%x, %s) didn't return an error", cborData, reflect.TypeOf(s2))
	} else if _, ok := err.(*DupMapKeyError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*DupMapKeyError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s2, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
	}
}

func TestStreamDupMapKeyToStruct(t *testing.T) {
	type s struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
		D string `cbor:"d"`
		E string `cbor:"e"`
	}
	cborData := hexDecode("a6616161416162614261636143616161466164614461656145") // {"a": "A", "b": "B", "c": "C", "a": "F", "d": "D", "e": "E"}

	var b []byte
	for i := 0; i < 3; i++ {
		b = append(b, cborData...)
	}

	// Duplicate key overwrites previous value (default).
	wantS := s{A: "A", B: "B", C: "C", D: "D", E: "E"}
	dec := NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var s1 s
		if err := dec.Decode(&s1); err != nil {
			t.Errorf("Decode() returned error %v", err)
		}
		if !reflect.DeepEqual(s1, wantS) {
			t.Errorf("Decode() = %v (%T), want %v (%T)", s1, s1, wantS, wantS)
		}
	}
	var v interface{}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}

	// Duplicate key triggers error.
	wantS = s{A: "A", B: "B", C: "C"}
	wantErrorMsg := "cbor: found duplicate map key \"a\" at map element index 3"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	dec = dm.NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var s2 s
		if err := dec.Decode(&s2); err == nil {
			t.Errorf("Decode() didn't return an error")
		} else if _, ok := err.(*DupMapKeyError); !ok {
			t.Errorf("Decode() returned wrong error type %T, want (*DupMapKeyError)", err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("Decode() returned error %q, want error containing %q", err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(s2, wantS) {
			t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
		}
	}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}
}

// dupl map key is a struct field
func TestUnmarshalDupMapKeyToStructKeyAsInt(t *testing.T) {
	type s struct {
		A int `cbor:"1,keyasint"`
		B int `cbor:"3,keyasint"`
		C int `cbor:"5,keyasint"`
	}
	cborData := hexDecode("a40102030401030506") // {1:2, 3:4, 1:3, 5:6}

	// Duplicate key doesn't overwrite previous value (default).
	wantS := s{A: 2, B: 4, C: 6}
	var s1 s
	if err := Unmarshal(cborData, &s1); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(s1, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s1, s1, wantS, wantS)
	}

	// Duplicate key triggers error.
	wantS = s{A: 2, B: 4}
	wantErrorMsg := "cbor: found duplicate map key \"1\" at map element index 2"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	var s2 s
	if err := dm.Unmarshal(cborData, &s2); err == nil {
		t.Errorf("Unmarshal(0x%x, %s) didn't return an error", cborData, reflect.TypeOf(s2))
	} else if _, ok := err.(*DupMapKeyError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*DupMapKeyError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s2, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
	}
}

func TestStreamDupMapKeyToStructKeyAsInt(t *testing.T) {
	type s struct {
		A int `cbor:"1,keyasint"`
		B int `cbor:"3,keyasint"`
		C int `cbor:"5,keyasint"`
	}
	cborData := hexDecode("a40102030401030506") // {1:2, 3:4, 1:3, 5:6}

	var b []byte
	for i := 0; i < 3; i++ {
		b = append(b, cborData...)
	}

	// Duplicate key overwrites previous value (default).
	wantS := s{A: 2, B: 4, C: 6}
	dec := NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var s1 s
		if err := dec.Decode(&s1); err != nil {
			t.Errorf("Decode() returned error %v", err)
		}
		if !reflect.DeepEqual(s1, wantS) {
			t.Errorf("Decode() = %v (%T), want %v (%T)", s1, s1, wantS, wantS)
		}
	}
	var v interface{}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}

	// Duplicate key triggers error.
	wantS = s{A: 2, B: 4}
	wantErrorMsg := "cbor: found duplicate map key \"1\" at map element index 2"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	dec = dm.NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var s2 s
		if err := dec.Decode(&s2); err == nil {
			t.Errorf("Decode() didn't return an error")
		} else if _, ok := err.(*DupMapKeyError); !ok {
			t.Errorf("Decode() returned wrong error type %T, want (*DupMapKeyError)", err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("Decode() returned error %q, want error containing %q", err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(s2, wantS) {
			t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
		}
	}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}
}

func TestUnmarshalDupMapKeyToStructNoMatchingField(t *testing.T) {
	type s struct {
		B string `cbor:"b"`
		C string `cbor:"c"`
		D string `cbor:"d"`
		E string `cbor:"e"`
	}
	cborData := hexDecode("a6616161416162614261636143616161466164614461656145") // {"a": "A", "b": "B", "c": "C", "a": "F", "d": "D", "e": "E"}

	wantS := s{B: "B", C: "C", D: "D", E: "E"}
	var s1 s
	if err := Unmarshal(cborData, &s1); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(s1, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s1, s1, wantS, wantS)
	}

	// Duplicate key triggers error even though map key "a" doesn't have a corresponding struct field.
	wantS = s{B: "B", C: "C"}
	wantErrorMsg := "cbor: found duplicate map key \"a\" at map element index 3"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	var s2 s
	if err := dm.Unmarshal(cborData, &s2); err == nil {
		t.Errorf("Unmarshal(0x%x, %s) didn't return an error", cborData, reflect.TypeOf(s2))
	} else if _, ok := err.(*DupMapKeyError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*DupMapKeyError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s2, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
	}
}

func TestStreamDupMapKeyToStructNoMatchingField(t *testing.T) {
	type s struct {
		B string `cbor:"b"`
		C string `cbor:"c"`
		D string `cbor:"d"`
		E string `cbor:"e"`
	}
	cborData := hexDecode("a6616161416162614261636143616161466164614461656145") // {"a": "A", "b": "B", "c": "C", "a": "F", "d": "D", "e": "E"}

	var b []byte
	for i := 0; i < 3; i++ {
		b = append(b, cborData...)
	}

	// Duplicate key overwrites previous value (default).
	wantS := s{B: "B", C: "C", D: "D", E: "E"}
	dec := NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var s1 s
		if err := dec.Decode(&s1); err != nil {
			t.Errorf("Decode() returned error %v", err)
		}
		if !reflect.DeepEqual(s1, wantS) {
			t.Errorf("Decode() = %v (%T), want %v (%T)", s1, s1, wantS, wantS)
		}
	}
	var v interface{}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}

	// Duplicate key triggers error.
	wantS = s{B: "B", C: "C"}
	wantErrorMsg := "cbor: found duplicate map key \"a\" at map element index 3"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	dec = dm.NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var s2 s
		if err := dec.Decode(&s2); err == nil {
			t.Errorf("Decode() didn't return an error")
		} else if _, ok := err.(*DupMapKeyError); !ok {
			t.Errorf("Decode() returned wrong error type %T, want (*DupMapKeyError)", err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("Decode() returned error %q, want error containing %q", err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(s2, wantS) {
			t.Errorf("Decode() = %v (%T), want %v (%T)", s2, s2, wantS, wantS)
		}
	}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}
}

func TestUnmarshalDupMapKeyToStructKeyAsIntNoMatchingField(t *testing.T) {
	type s struct {
		B int `cbor:"3,keyasint"`
		C int `cbor:"5,keyasint"`
	}
	cborData := hexDecode("a40102030401030506") // {1:2, 3:4, 1:3, 5:6}

	wantS := s{B: 4, C: 6}
	var s1 s
	if err := Unmarshal(cborData, &s1); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	}
	if !reflect.DeepEqual(s1, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s1, s1, wantS, wantS)
	}

	// Duplicate key triggers error even though map key "a" doesn't have a corresponding struct field.
	wantS = s{B: 4}
	wantErrorMsg := "cbor: found duplicate map key \"1\" at map element index 2"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	var s2 s
	if err := dm.Unmarshal(cborData, &s2); err == nil {
		t.Errorf("Unmarshal(0x%x, %s) didn't return an error", cborData, reflect.TypeOf(s2))
	} else if _, ok := err.(*DupMapKeyError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*DupMapKeyError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s2, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
	}
}

func TestStreamDupMapKeyToStructKeyAsIntNoMatchingField(t *testing.T) {
	type s struct {
		B int `cbor:"3,keyasint"`
		C int `cbor:"5,keyasint"`
	}
	cborData := hexDecode("a40102030401030506") // {1:2, 3:4, 1:3, 5:6}

	var b []byte
	for i := 0; i < 3; i++ {
		b = append(b, cborData...)
	}

	// Duplicate key overwrites previous value (default).
	wantS := s{B: 4, C: 6}
	dec := NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var s1 s
		if err := dec.Decode(&s1); err != nil {
			t.Errorf("Decode() returned error %v", err)
		}
		if !reflect.DeepEqual(s1, wantS) {
			t.Errorf("Decode() = %v (%T), want %v (%T)", s1, s1, wantS, wantS)
		}
	}
	var v interface{}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}

	// Duplicate key triggers error.
	wantS = s{B: 4}
	wantErrorMsg := "cbor: found duplicate map key \"1\" at map element index 2"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	dec = dm.NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var s2 s
		if err := dec.Decode(&s2); err == nil {
			t.Errorf("Decode() didn't return an error")
		} else if _, ok := err.(*DupMapKeyError); !ok {
			t.Errorf("Decode() returned wrong error type %T, want (*DupMapKeyError)", err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("Decode() returned error %q, want error containing %q", err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(s2, wantS) {
			t.Errorf("Decode() = %v (%T), want %v (%T)", s2, s2, wantS, wantS)
		}
	}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}
}

func TestUnmarshalDupMapKeyToStructWrongType(t *testing.T) {
	type s struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
		D string `cbor:"d"`
		E string `cbor:"e"`
	}
	cborData := hexDecode("a861616141fa47c35000026162614261636143fa47c3500003616161466164614461656145") // {"a": "A", 100000.0:2, "b": "B", "c": "C", 100000.0:3, "a": "F", "d": "D", "e": "E"}

	var s1 s
	wantS := s{A: "A", B: "B", C: "C", D: "D", E: "E"}
	wantErrorMsg := "cbor: cannot unmarshal"
	if err := Unmarshal(cborData, &s1); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s1, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s1, s1, wantS, wantS)
	}

	wantS = s{A: "A", B: "B", C: "C"}
	wantErrorMsg = "cbor: found duplicate map key \"100000\" at map element index 4"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	var s2 s
	if err := dm.Unmarshal(cborData, &s2); err == nil {
		t.Errorf("Unmarshal(0x%x, %s) didn't return an error", cborData, reflect.TypeOf(s2))
	} else if _, ok := err.(*DupMapKeyError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*DupMapKeyError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s2, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
	}
}

func TestStreamDupMapKeyToStructWrongType(t *testing.T) {
	type s struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
		D string `cbor:"d"`
		E string `cbor:"e"`
	}
	cborData := hexDecode("a861616141fa47c35000026162614261636143fa47c3500003616161466164614461656145") // {"a": "A", 100000.0:2, "b": "B", "c": "C", 100000.0:3, "a": "F", "d": "D", "e": "E"}

	var b []byte
	for i := 0; i < 3; i++ {
		b = append(b, cborData...)
	}

	wantS := s{A: "A", B: "B", C: "C", D: "D", E: "E"}
	wantErrorMsg := "cbor: cannot unmarshal"
	dec := NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var s1 s
		if err := dec.Decode(&s1); err == nil {
			t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
		} else if _, ok := err.(*UnmarshalTypeError); !ok {
			t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
		} else if !strings.Contains(err.Error(), wantErrorMsg) {
			t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(s1, wantS) {
			t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s1, s1, wantS, wantS)
		}
	}
	var v interface{}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}

	// Duplicate key triggers error.
	wantS = s{A: "A", B: "B", C: "C"}
	wantErrorMsg = "cbor: found duplicate map key \"100000\" at map element index 4"
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	dec = dm.NewDecoder(bytes.NewReader(b))
	for i := 0; i < 3; i++ {
		var s2 s
		if err := dec.Decode(&s2); err == nil {
			t.Errorf("Decode() didn't return an error")
		} else if _, ok := err.(*DupMapKeyError); !ok {
			t.Errorf("Decode() returned wrong error type %T, want (*DupMapKeyError)", err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("Decode() returned error %q, want error containing %q", err.Error(), wantErrorMsg)
		}
		if !reflect.DeepEqual(s2, wantS) {
			t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
		}
	}
	if err := dec.Decode(&v); err != io.EOF {
		t.Errorf("Decode() returned error %v, want %v", err, io.EOF)
	}
}

func TestUnmarshalDupMapKeyToStructStringParseError(t *testing.T) {
	type s struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
		D string `cbor:"d"`
		E string `cbor:"e"`
	}
	cborData := hexDecode("a661fe6141616261426163614361fe61466164614461656145") // {"\xFE": "A", "b": "B", "c": "C", "\xFE": "F", "d": "D", "e": "E"}
	wantS := s{A: "", B: "B", C: "C", D: "D", E: "E"}
	wantErrorMsg := "cbor: invalid UTF-8 string"

	// Duplicate key doesn't overwrite previous value (default).
	var s1 s
	if err := Unmarshal(cborData, &s1); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else if _, ok := err.(*SemanticError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*SemanticError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s1, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s1, s1, wantS, wantS)
	}

	// Duplicate key triggers error.
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	var s2 s
	if err := dm.Unmarshal(cborData, &s2); err == nil {
		t.Errorf("Unmarshal(0x%x, %s) didn't return an error", cborData, reflect.TypeOf(s2))
	} else if _, ok := err.(*SemanticError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*SemanticError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s2, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
	}
}

func TestUnmarshalDupMapKeyToStructIntParseError(t *testing.T) {
	type s struct {
		A int `cbor:"1,keyasint"`
		B int `cbor:"3,keyasint"`
		C int `cbor:"5,keyasint"`
	}
	cborData := hexDecode("a43bffffffffffffffff0203043bffffffffffffffff030506") // {-18446744073709551616:2, 3:4, -18446744073709551616:3, 5:6}

	// Duplicate key doesn't overwrite previous value (default).
	wantS := s{B: 4, C: 6}
	wantErrorMsg := "cbor: cannot unmarshal"
	var s1 s
	if err := Unmarshal(cborData, &s1); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s1, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s1, s1, wantS, wantS)
	}

	// Duplicate key triggers error.
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	var s2 s
	if err := dm.Unmarshal(cborData, &s2); err == nil {
		t.Errorf("Unmarshal(0x%x, %s) didn't return an error", cborData, reflect.TypeOf(s2))
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s2, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
	}
}

func TestUnmarshalDupMapKeyToStructWrongTypeParseError(t *testing.T) {
	type s struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
		D string `cbor:"d"`
		E string `cbor:"e"`
	}
	cborData := hexDecode("a68161fe614161626142616361438161fe61466164614461656145") // {["\xFE"]: "A", "b": "B", "c": "C", ["\xFE"]: "F", "d": "D", "e": "E"}

	// Duplicate key doesn't overwrite previous value (default).
	wantS := s{A: "", B: "B", C: "C", D: "D", E: "E"}
	wantErrorMsg := "cbor: cannot unmarshal"
	var s1 s
	if err := Unmarshal(cborData, &s1); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s1, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s1, s1, wantS, wantS)
	}

	// Duplicate key triggers error.
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	var s2 s
	if err := dm.Unmarshal(cborData, &s2); err == nil {
		t.Errorf("Unmarshal(0x%x, %s) didn't return an error", cborData, reflect.TypeOf(s2))
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s2, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
	}
}

func TestUnmarshalDupMapKeyToStructWrongTypeUnhashableError(t *testing.T) {
	type s struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
		D string `cbor:"d"`
		E string `cbor:"e"`
	}
	cborData := hexDecode("a6810061416162614261636143810061466164614461656145") // {[0]: "A", "b": "B", "c": "C", [0]: "F", "d": "D", "e": "E"}
	wantS := s{A: "", B: "B", C: "C", D: "D", E: "E"}

	// Duplicate key doesn't overwrite previous value (default).
	wantErrorMsg := "cbor: cannot unmarshal"
	var s1 s
	if err := Unmarshal(cborData, &s1); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s1, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s1, s1, wantS, wantS)
	}

	// Duplicate key triggers error.
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	var s2 s
	if err := dm.Unmarshal(cborData, &s2); err == nil {
		t.Errorf("Unmarshal(0x%x, %s) didn't return an error", cborData, reflect.TypeOf(s2))
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s2, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
	}
}

func TestUnmarshalDupMapKeyToStructTagTypeError(t *testing.T) {
	type s struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
		D string `cbor:"d"`
		E string `cbor:"e"`
	}
	cborData := hexDecode("a6c24901000000000000000061416162614261636143c24901000000000000000061466164614461656145") // {bignum(18446744073709551616): "A", "b": "B", "c": "C", bignum(18446744073709551616): "F", "d": "D", "e": "E"}
	wantS := s{A: "", B: "B", C: "C", D: "D", E: "E"}

	// Duplicate key doesn't overwrite previous value (default).
	wantErrorMsg := "cbor: cannot unmarshal"
	var s1 s
	if err := Unmarshal(cborData, &s1); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s1, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s1, s1, wantS, wantS)
	}

	// Duplicate key triggers error.
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	var s2 s
	if err := dm.Unmarshal(cborData, &s2); err == nil {
		t.Errorf("Unmarshal(0x%x, %s) didn't return an error", cborData, reflect.TypeOf(s2))
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), wantErrorMsg)
	}
	if !reflect.DeepEqual(s2, wantS) {
		t.Errorf("Unmarshal(0x%x) = %+v (%T), want %+v (%T)", cborData, s2, s2, wantS, wantS)
	}
}

func TestIndefiniteLengthArrayToArray(t *testing.T) {
	testCases := []struct {
		name     string
		cborData []byte
		wantV    interface{}
	}{
		{
			name:     "CBOR empty array to Go 5 elem array",
			cborData: hexDecode("9fff"),
			wantV:    [5]byte{},
		},
		{
			name:     "CBOR 3 elem array to Go 5 elem array",
			cborData: hexDecode("9f010203ff"),
			wantV:    [5]byte{1, 2, 3, 0, 0},
		},
		{
			name:     "CBOR 10 elem array to Go 5 elem array",
			cborData: hexDecode("9f0102030405060708090aff"),
			wantV:    [5]byte{1, 2, 3, 4, 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.New(reflect.TypeOf(tc.wantV))
			if err := Unmarshal(tc.cborData, v.Interface()); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
			}
			if !reflect.DeepEqual(v.Elem().Interface(), tc.wantV) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborData, v.Elem().Interface(), v.Elem().Interface(), tc.wantV, tc.wantV)
			}
		})
	}
}

func TestExceedMaxArrayElements(t *testing.T) {
	testCases := []struct {
		name         string
		opts         DecOptions
		cborData     []byte
		wantErrorMsg string
	}{
		{
			name:         "array",
			opts:         DecOptions{MaxArrayElements: 16},
			cborData:     hexDecode("910101010101010101010101010101010101"),
			wantErrorMsg: "cbor: exceeded max number of elements 16 for CBOR array",
		},
		{
			name:         "indefinite length array",
			opts:         DecOptions{MaxArrayElements: 16},
			cborData:     hexDecode("9f0101010101010101010101010101010101ff"),
			wantErrorMsg: "cbor: exceeded max number of elements 16 for CBOR array",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dm, _ := tc.opts.DecMode()
			var v interface{}
			if err := dm.Unmarshal(tc.cborData, &v); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error", tc.cborData)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestExceedMaxMapPairs(t *testing.T) {
	testCases := []struct {
		name         string
		opts         DecOptions
		cborData     []byte
		wantErrorMsg string
	}{
		{
			name:         "array",
			opts:         DecOptions{MaxMapPairs: 16},
			cborData:     hexDecode("b101010101010101010101010101010101010101010101010101010101010101010101"),
			wantErrorMsg: "cbor: exceeded max number of key-value pairs 16 for CBOR map",
		},
		{
			name:         "indefinite length array",
			opts:         DecOptions{MaxMapPairs: 16},
			cborData:     hexDecode("bf01010101010101010101010101010101010101010101010101010101010101010101ff"),
			wantErrorMsg: "cbor: exceeded max number of key-value pairs 16 for CBOR map",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dm, _ := tc.opts.DecMode()
			var v interface{}
			if err := dm.Unmarshal(tc.cborData, &v); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error", tc.cborData)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecIndefiniteLengthOption(t *testing.T) {
	testCases := []struct {
		name         string
		opts         DecOptions
		cborData     []byte
		wantErrorMsg string
	}{
		{
			name:         "byte string",
			opts:         DecOptions{IndefLength: IndefLengthForbidden},
			cborData:     hexDecode("5fff"),
			wantErrorMsg: "cbor: indefinite-length byte string isn't allowed",
		},
		{
			name:         "text string",
			opts:         DecOptions{IndefLength: IndefLengthForbidden},
			cborData:     hexDecode("7fff"),
			wantErrorMsg: "cbor: indefinite-length UTF-8 text string isn't allowed",
		},
		{
			name:         "array",
			opts:         DecOptions{IndefLength: IndefLengthForbidden},
			cborData:     hexDecode("9fff"),
			wantErrorMsg: "cbor: indefinite-length array isn't allowed",
		},
		{
			name:         "indefinite length array",
			opts:         DecOptions{IndefLength: IndefLengthForbidden},
			cborData:     hexDecode("bfff"),
			wantErrorMsg: "cbor: indefinite-length map isn't allowed",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Default option allows indefinite length items
			var v interface{}
			if err := Unmarshal(tc.cborData, &v); err != nil {
				t.Errorf("Unmarshal(0x%x) returned an error %v", tc.cborData, err)
			}

			dm, _ := tc.opts.DecMode()
			if err := dm.Unmarshal(tc.cborData, &v); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error", tc.cborData)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecTagsMdOption(t *testing.T) {
	cborData := hexDecode("c074323031332d30332d32315432303a30343a30305a")
	wantErrorMsg := "cbor: CBOR tag isn't allowed"

	// Default option allows CBOR tags
	var v interface{}
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned an error %v", cborData, err)
	}

	// Decoding CBOR tags with TagsForbidden option returns error
	dm, _ := DecOptions{TagsMd: TagsForbidden}.DecMode()
	if err := dm.Unmarshal(cborData, &v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", cborData)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("Unmarshal(0x%x) returned error %q, want %q", cborData, err.Error(), wantErrorMsg)
	}

	// Create DecMode with TagSet and TagsForbidden option returns error
	wantErrorMsg = "cbor: cannot create DecMode with TagSet when TagsMd is TagsForbidden"
	tags := NewTagSet()
	_, err := DecOptions{TagsMd: TagsForbidden}.DecModeWithTags(tags)
	if err == nil {
		t.Errorf("DecModeWithTags() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("DecModeWithTags() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
	_, err = DecOptions{TagsMd: TagsForbidden}.DecModeWithSharedTags(tags)
	if err == nil {
		t.Errorf("DecModeWithSharedTags() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("DecModeWithSharedTags() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestDecModeInvalidIntDec(t *testing.T) {
	wantErrorMsg := "cbor: invalid IntDec 101"
	_, err := DecOptions{IntDec: 101}.DecMode()
	if err == nil {
		t.Errorf("DecMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("DecMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestIntDec(t *testing.T) {
	dm, err := DecOptions{IntDec: IntDecConvertSigned}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned an error %+v", err)
	}

	testCases := []struct {
		name         string
		cborData     []byte
		wantObj      interface{}
		wantErrorMsg string
	}{
		{
			name:     "CBOR pos int",
			cborData: hexDecode("1a000f4240"),
			wantObj:  int64(1000000),
		},
		{
			name:         "CBOR pos int overflows int64",
			cborData:     hexDecode("1bffffffffffffffff"),
			wantErrorMsg: "18446744073709551615 overflows Go's int64",
		},
		{
			name:     "CBOR neg int",
			cborData: hexDecode("3903e7"),
			wantObj:  int64(-1000),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v interface{}
			err := dm.Unmarshal(tc.cborData, &v)
			if err == nil {
				if tc.wantErrorMsg != "" {
					t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", tc.cborData, tc.wantErrorMsg)
				} else if !reflect.DeepEqual(v, tc.wantObj) {
					t.Errorf("Unmarshal(0x%x) return %v (%T), want %v (%T)", tc.cborData, v, v, tc.wantObj, tc.wantObj)
				}
			} else {
				if tc.wantErrorMsg == "" {
					t.Errorf("Unmarshal(0x%x) returned error %q", tc.cborData, err)
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
				}
			}
		})
	}
}

func TestDecModeInvalidExtraError(t *testing.T) {
	wantErrorMsg := "cbor: invalid ExtraReturnErrors 3"
	_, err := DecOptions{ExtraReturnErrors: 3}.DecMode()
	if err == nil {
		t.Errorf("DecMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("DecMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestExtraErrorCondUnknownField(t *testing.T) {
	type s struct {
		A string
		B string
		C string
	}

	dm, _ := DecOptions{}.DecMode()
	dmUnknownFieldError, _ := DecOptions{ExtraReturnErrors: ExtraDecErrorUnknownField}.DecMode()

	testCases := []struct {
		name         string
		cborData     []byte
		dm           DecMode
		wantObj      interface{}
		wantErrorMsg string
	}{
		{
			name:     "field by field match",
			cborData: hexDecode("a3614161616142616261436163"), // map[string]string{"A": "a", "B": "b", "C": "c"}
			dm:       dm,
			wantObj:  s{A: "a", B: "b", C: "c"},
		},
		{
			name:     "field by field match with ExtraDecErrorUnknownField",
			cborData: hexDecode("a3614161616142616261436163"), // map[string]string{"A": "a", "B": "b", "C": "c"}
			dm:       dmUnknownFieldError,
			wantObj:  s{A: "a", B: "b", C: "c"},
		},
		{
			name:     "CBOR map less field",
			cborData: hexDecode("a26141616161426162"), // map[string]string{"A": "a", "B": "b"}
			dm:       dm,
			wantObj:  s{A: "a", B: "b", C: ""},
		},
		{
			name:     "CBOR map less field with ExtraDecErrorUnknownField",
			cborData: hexDecode("a26141616161426162"), // map[string]string{"A": "a", "B": "b"}
			dm:       dmUnknownFieldError,
			wantObj:  s{A: "a", B: "b", C: ""},
		},
		{
			name:     "CBOR map unknown field",
			cborData: hexDecode("a461416161614261626143616361446164"), // map[string]string{"A": "a", "B": "b", "C": "c", "D": "d"}
			dm:       dm,
			wantObj:  s{A: "a", B: "b", C: "c"},
		},
		{
			name:         "CBOR map unknown field with ExtraDecErrorUnknownField",
			cborData:     hexDecode("a461416161614261626143616361446164"), // map[string]string{"A": "a", "B": "b", "C": "c", "D": "d"}
			dm:           dmUnknownFieldError,
			wantErrorMsg: "cbor: found unknown field at map element index 3",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v s
			err := tc.dm.Unmarshal(tc.cborData, &v)
			if err == nil {
				if tc.wantErrorMsg != "" {
					t.Errorf("Unmarshal(0x%x) didn't return an error, want %q", tc.cborData, tc.wantErrorMsg)
				} else if !reflect.DeepEqual(v, tc.wantObj) {
					t.Errorf("Unmarshal(0x%x) return %v (%T), want %v (%T)", tc.cborData, v, v, tc.wantObj, tc.wantObj)
				}
			} else {
				if tc.wantErrorMsg == "" {
					t.Errorf("Unmarshal(0x%x) returned error %q", tc.cborData, err)
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
				}
			}
		})
	}
}

func TestInvalidUTF8Mode(t *testing.T) {
	wantErrorMsg := "cbor: invalid UTF8 2"
	_, err := DecOptions{UTF8: 2}.DecMode()
	if err == nil {
		t.Errorf("DecMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("DecMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestStreamExtraErrorCondUnknownField(t *testing.T) {
	type s struct {
		A string
		B string
		C string
	}

	cborData := hexDecode("a461416161614461646142616261436163a3614161616142616261436163") // map[string]string{"A": "a", "D": "d", "B": "b", "C": "c"}, map[string]string{"A": "a", "B": "b", "C": "c"}
	wantErrorMsg := "cbor: found unknown field at map element index 1"
	wantObj := s{A: "a", B: "b", C: "c"}

	dmUnknownFieldError, _ := DecOptions{ExtraReturnErrors: ExtraDecErrorUnknownField}.DecMode()
	dec := dmUnknownFieldError.NewDecoder(bytes.NewReader(cborData))

	var v1 s
	err := dec.Decode(&v1)
	if err == nil {
		t.Errorf("Decode() didn't return an error, want %q", wantErrorMsg)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Decode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}

	var v2 s
	err = dec.Decode(&v2)
	if err != nil {
		t.Errorf("Decode() returned an error %v", err)
	} else if !reflect.DeepEqual(v2, wantObj) {
		t.Errorf("Decode() return %v (%T), want %v (%T)", v2, v2, wantObj, wantObj)
	}
}

// TestUnmarshalTagNum55799 is identical to TestUnmarshal,
// except that CBOR test data is prefixed with tag number 55799 (0xd9d9f7).
func TestUnmarshalTagNum55799(t *testing.T) {
	tagNum55799 := hexDecode("d9d9f7")

	for _, tc := range unmarshalTests {
		// Prefix tag number 55799 to CBOR test data
		cborData := make([]byte, len(tc.cborData)+6)
		copy(cborData, tagNum55799)
		copy(cborData[3:], tagNum55799)
		copy(cborData[6:], tc.cborData)

		// Test unmarshalling CBOR into empty interface.
		var v interface{}
		if err := Unmarshal(cborData, &v); err != nil {
			t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
		} else {
			if tm, ok := tc.emptyInterfaceValue.(time.Time); ok {
				if vt, ok := v.(time.Time); !ok || !tm.Equal(vt) {
					t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, tc.emptyInterfaceValue, tc.emptyInterfaceValue)
				}
			} else if !reflect.DeepEqual(v, tc.emptyInterfaceValue) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, tc.emptyInterfaceValue, tc.emptyInterfaceValue)
			}
		}

		// Test unmarshalling CBOR into RawMessage.
		var r RawMessage
		if err := Unmarshal(cborData, &r); err != nil {
			t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
		} else if !bytes.Equal(r, tc.cborData) {
			t.Errorf("Unmarshal(0x%x) returned RawMessage %v, want %v", cborData, r, tc.cborData)
		}

		// Test unmarshalling CBOR into compatible data types.
		for _, value := range tc.values {
			v := reflect.New(reflect.TypeOf(value))
			vPtr := v.Interface()
			if err := Unmarshal(cborData, vPtr); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
			} else {
				if tm, ok := value.(time.Time); ok {
					if vt, ok := v.Elem().Interface().(time.Time); !ok || !tm.Equal(vt) {
						t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v.Elem().Interface(), v.Elem().Interface(), value, value)
					}
				} else if !reflect.DeepEqual(v.Elem().Interface(), value) {
					t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v.Elem().Interface(), v.Elem().Interface(), value, value)
				}
			}
		}

		// Test unmarshalling CBOR into incompatible data types.
		for _, typ := range tc.wrongTypes {
			v := reflect.New(typ)
			vPtr := v.Interface()
			if err := Unmarshal(cborData, vPtr); err == nil {
				t.Errorf("Unmarshal(0x%x, %s) didn't return an error", cborData, typ.String())
			} else if _, ok := err.(*UnmarshalTypeError); !ok {
				t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", cborData, err)
			} else if !strings.Contains(err.Error(), "cannot unmarshal") {
				t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", cborData, err.Error(), "cannot unmarshal")
			}
		}
	}
}

// TestUnmarshalFloatWithTagNum55799 is identical to TestUnmarshalFloat,
// except that CBOR test data is prefixed with tag number 55799 (0xd9d9f7).
func TestUnmarshalFloatWithTagNum55799(t *testing.T) {
	tagNum55799 := hexDecode("d9d9f7")

	for _, tc := range unmarshalFloatTests {
		// Prefix tag number 55799 to CBOR test data
		cborData := make([]byte, len(tc.cborData)+3)
		copy(cborData, tagNum55799)
		copy(cborData[3:], tc.cborData)

		// Test unmarshalling CBOR into empty interface.
		var v interface{}
		if err := Unmarshal(tc.cborData, &v); err != nil {
			t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
		} else {
			testFloat(t, tc.cborData, v, tc.emptyInterfaceValue, tc.equalityThreshold)
		}

		// Test unmarshalling CBOR into RawMessage.
		var r RawMessage
		if err := Unmarshal(tc.cborData, &r); err != nil {
			t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
		} else if !bytes.Equal(r, tc.cborData) {
			t.Errorf("Unmarshal(0x%x) returned RawMessage %v, want %v", tc.cborData, r, tc.cborData)
		}

		// Test unmarshalling CBOR into compatible data types.
		for _, value := range tc.values {
			v := reflect.New(reflect.TypeOf(value))
			vPtr := v.Interface()
			if err := Unmarshal(tc.cborData, vPtr); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
			} else {
				testFloat(t, tc.cborData, v.Elem().Interface(), value, tc.equalityThreshold)
			}
		}

		// Test unmarshalling CBOR into incompatible data types.
		for _, typ := range tc.wrongTypes {
			v := reflect.New(typ)
			vPtr := v.Interface()
			if err := Unmarshal(tc.cborData, vPtr); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error", tc.cborData)
			} else if _, ok := err.(*UnmarshalTypeError); !ok {
				t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", tc.cborData, err)
			} else if !strings.Contains(err.Error(), "cannot unmarshal") {
				t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", tc.cborData, err.Error(), "cannot unmarshal")
			}
		}
	}
}

func TestUnmarshalTagNum55799AsElement(t *testing.T) {
	testCases := []struct {
		name                string
		cborData            []byte
		emptyInterfaceValue interface{}
		values              []interface{}
		wrongTypes          []reflect.Type
	}{
		{
			"array",
			hexDecode("d9d9f783d9d9f701d9d9f702d9d9f703"), // 55799([55799(1), 55799(2), 55799(3)])
			[]interface{}{uint64(1), uint64(2), uint64(3)},
			[]interface{}{[]interface{}{uint64(1), uint64(2), uint64(3)}, []byte{1, 2, 3}, []int{1, 2, 3}, []uint{1, 2, 3}, [0]int{}, [1]int{1}, [3]int{1, 2, 3}, [5]int{1, 2, 3, 0, 0}, []float32{1, 2, 3}, []float64{1, 2, 3}},
			[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeString, typeBool, typeStringSlice, typeMapStringInt, reflect.TypeOf([3]string{}), typeTag, typeRawTag},
		},
		{
			"map",
			hexDecode("d9d9f7a2d9d9f701d9d9f702d9d9f703d9d9f704"), // 55799({55799(1): 55799(2), 55799(3): 55799(4)})
			map[interface{}]interface{}{uint64(1): uint64(2), uint64(3): uint64(4)},
			[]interface{}{map[interface{}]interface{}{uint64(1): uint64(2), uint64(3): uint64(4)}, map[uint]int{1: 2, 3: 4}, map[int]uint{1: 2, 3: 4}},
			[]reflect.Type{typeUint8, typeUint16, typeUint32, typeUint64, typeInt8, typeInt16, typeInt32, typeInt64, typeFloat32, typeFloat64, typeByteSlice, typeByteArray, typeString, typeBool, typeIntSlice, typeMapStringInt, typeTag, typeRawTag},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test unmarshalling CBOR into empty interface.
			var v interface{}
			if err := Unmarshal(tc.cborData, &v); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
			} else {
				if tm, ok := tc.emptyInterfaceValue.(time.Time); ok {
					if vt, ok := v.(time.Time); !ok || !tm.Equal(vt) {
						t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborData, v, v, tc.emptyInterfaceValue, tc.emptyInterfaceValue)
					}
				} else if !reflect.DeepEqual(v, tc.emptyInterfaceValue) {
					t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborData, v, v, tc.emptyInterfaceValue, tc.emptyInterfaceValue)
				}
			}

			// Test unmarshalling CBOR into compatible data types.
			for _, value := range tc.values {
				v := reflect.New(reflect.TypeOf(value))
				vPtr := v.Interface()
				if err := Unmarshal(tc.cborData, vPtr); err != nil {
					t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborData, err)
				} else {
					if tm, ok := value.(time.Time); ok {
						if vt, ok := v.Elem().Interface().(time.Time); !ok || !tm.Equal(vt) {
							t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborData, v.Elem().Interface(), v.Elem().Interface(), value, value)
						}
					} else if !reflect.DeepEqual(v.Elem().Interface(), value) {
						t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborData, v.Elem().Interface(), v.Elem().Interface(), value, value)
					}
				}
			}

			// Test unmarshalling CBOR into incompatible data types.
			for _, typ := range tc.wrongTypes {
				v := reflect.New(typ)
				vPtr := v.Interface()
				if err := Unmarshal(tc.cborData, vPtr); err == nil {
					t.Errorf("Unmarshal(0x%x, %s) didn't return an error", tc.cborData, typ.String())
				} else if _, ok := err.(*UnmarshalTypeError); !ok {
					t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*UnmarshalTypeError)", tc.cborData, err)
				} else if !strings.Contains(err.Error(), "cannot unmarshal") {
					t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", tc.cborData, err.Error(), "cannot unmarshal")
				}
			}
		})
	}
}

func TestUnmarshalTagNum55799ToBinaryUnmarshaler(t *testing.T) {
	cborData := hexDecode("d9d9f74800000000499602d2") // 55799(h'00000000499602D2')
	wantObj := number(1234567890)

	var v number
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if !reflect.DeepEqual(v, wantObj) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, wantObj, wantObj)
	}
}

func TestUnmarshalTagNum55799ToUnmarshaler(t *testing.T) {
	cborData := hexDecode("d9d9f7d864a1636e756d01") // 55799(100({"num": 1}))
	wantObj := number3(1)

	var v number3
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if !reflect.DeepEqual(v, wantObj) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, wantObj, wantObj)
	}
}

func TestUnmarshalTagNum55799ToRegisteredGoType(t *testing.T) {
	type myInt int
	typ := reflect.TypeOf(myInt(0))

	tags := NewTagSet()
	if err := tags.Add(TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, typ, 125); err != nil {
		t.Fatalf("TagSet.Add(%s, %v) returned error %v", typ, 125, err)
	}

	dm, _ := DecOptions{}.DecModeWithTags(tags)

	cborData := hexDecode("d9d9f7d87d01") // 55799(125(1))
	wantObj := myInt(1)

	var v myInt
	if err := dm.Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if !reflect.DeepEqual(v, wantObj) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, wantObj, wantObj)
	}
}

// TODO: wait for clarification from 7049bis https://github.com/cbor-wg/CBORbis/issues/183
// Nested tag number 55799 may be stripeed as well depending on 7049bis clarification.
func TestUnmarshalNestedTagNum55799ToEmptyInterface(t *testing.T) {
	cborData := hexDecode("d864d9d9f701") // 100(55799(1))
	wantObj := Tag{100, Tag{55799, uint64(1)}}

	var v interface{}
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if !reflect.DeepEqual(v, wantObj) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, wantObj, wantObj)
	}
}

func TestUnmarshalNestedTagNum55799ToValue(t *testing.T) {
	cborData := hexDecode("d864d9d9f701") // 100(55799(1))
	wantObj := 1

	var v int
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if !reflect.DeepEqual(v, wantObj) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, wantObj, wantObj)
	}
}

func TestUnmarshalNestedTagNum55799ToTag(t *testing.T) {
	cborData := hexDecode("d864d9d9f701") // 100(55799(1))
	wantObj := Tag{100, Tag{55799, uint64(1)}}

	var v Tag
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if !reflect.DeepEqual(v, wantObj) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, wantObj, wantObj)
	}
}

func TestUnmarshalNestedTagNum55799ToTime(t *testing.T) {
	cborData := hexDecode("c0d9d9f774323031332d30332d32315432303a30343a30305a") // 0(55799("2013-03-21T20:04:00Z"))
	wantErrorMsg := "tag number 0 must be followed by text string, got tag"

	var v time.Time
	if err := Unmarshal(cborData, &v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return error", cborData)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %s, want %s", cborData, err.Error(), wantErrorMsg)
	}
}

func TestUnmarshalNestedTagNum55799ToBinaryUnmarshaler(t *testing.T) {
	cborData := hexDecode("d864d9d9f74800000000499602d2") // 100(55799(h'00000000499602D2'))
	wantObj := number(1234567890)

	var v number
	if err := Unmarshal(cborData, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", cborData, err)
	} else if !reflect.DeepEqual(v, wantObj) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", cborData, v, v, wantObj, wantObj)
	}
}

func TestUnmarshalNestedTagNum55799ToUnmarshaler(t *testing.T) {
	cborData := hexDecode("d864d9d9f7a1636e756d01") // 100(55799({"num": 1}))
	wantErrorMsg := "wrong tag content type"

	var v number3
	if err := Unmarshal(cborData, &v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return error", cborData)
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %s, want %s", cborData, err.Error(), wantErrorMsg)
	}
}

func TestUnmarshalNestedTagNum55799ToRegisteredGoType(t *testing.T) {
	type myInt int
	typ := reflect.TypeOf(myInt(0))

	tags := NewTagSet()
	if err := tags.Add(TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, typ, 125); err != nil {
		t.Fatalf("TagSet.Add(%s, %v) returned error %v", typ, 125, err)
	}

	dm, _ := DecOptions{}.DecModeWithTags(tags)

	cborData := hexDecode("d87dd9d9f701") // 125(55799(1))
	wantErrorMsg := "cbor: wrong tag number for cbor.myInt, got [125 55799], expected [125]"

	var v myInt
	if err := dm.Unmarshal(cborData, &v); err == nil {
		t.Errorf("Unmarshal() didn't return error")
	} else if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("Unmarshal(0x%x) returned error %s, want %s", cborData, err.Error(), wantErrorMsg)
	}
}

func TestUnmarshalPosIntToBigInt(t *testing.T) {
	cborData := hexDecode("1bffffffffffffffff") // 18446744073709551615
	wantEmptyInterfaceValue := uint64(18446744073709551615)
	wantBigIntValue := bigIntOrPanic("18446744073709551615")

	var v1 interface{}
	if err := Unmarshal(cborData, &v1); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %+v", cborData, err)
	} else if !reflect.DeepEqual(v1, wantEmptyInterfaceValue) {
		t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", cborData, v1, v1, wantEmptyInterfaceValue, wantEmptyInterfaceValue)
	}

	var v2 big.Int
	if err := Unmarshal(cborData, &v2); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %+v", cborData, err)
	} else if !reflect.DeepEqual(v2, wantBigIntValue) {
		t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", cborData, v2, v2, wantBigIntValue, wantBigIntValue)
	}
}

func TestUnmarshalNegIntToBigInt(t *testing.T) {
	testCases := []struct {
		name                    string
		cborData                []byte
		wantEmptyInterfaceValue interface{}
		wantBigIntValue         big.Int
	}{
		{
			name:                    "fit Go int64",
			cborData:                hexDecode("3b7fffffffffffffff"), // -9223372036854775808
			wantEmptyInterfaceValue: int64(-9223372036854775808),
			wantBigIntValue:         bigIntOrPanic("-9223372036854775808"),
		},
		{
			name:                    "overflow Go int64",
			cborData:                hexDecode("3b8000000000000000"), // -9223372036854775809
			wantEmptyInterfaceValue: bigIntOrPanic("-9223372036854775809"),
			wantBigIntValue:         bigIntOrPanic("-9223372036854775809"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v1 interface{}
			if err := Unmarshal(tc.cborData, &v1); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %+v", tc.cborData, err)
			} else if !reflect.DeepEqual(v1, tc.wantEmptyInterfaceValue) {
				t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", tc.cborData, v1, v1, tc.wantEmptyInterfaceValue, tc.wantEmptyInterfaceValue)
			}

			var v2 big.Int
			if err := Unmarshal(tc.cborData, &v2); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %+v", tc.cborData, err)
			} else if !reflect.DeepEqual(v2, tc.wantBigIntValue) {
				t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", tc.cborData, v2, v2, tc.wantBigIntValue, tc.wantBigIntValue)
			}
		})
	}
}

func TestUnmarshalTag2(t *testing.T) {
	testCases := []struct {
		name                    string
		cborData                []byte
		wantEmptyInterfaceValue interface{}
		wantValues              []interface{}
	}{
		{
			name:                    "fit Go int64",
			cborData:                hexDecode("c2430f4240"), // 2(1000000)
			wantEmptyInterfaceValue: bigIntOrPanic("1000000"),
			wantValues: []interface{}{
				int64(1000000),
				uint64(1000000),
				float32(1000000),
				float64(1000000),
				bigIntOrPanic("1000000"),
			},
		},
		{
			name:                    "fit Go uint64",
			cborData:                hexDecode("c248ffffffffffffffff"), // 2(18446744073709551615)
			wantEmptyInterfaceValue: bigIntOrPanic("18446744073709551615"),
			wantValues: []interface{}{
				uint64(18446744073709551615),
				float32(18446744073709551615),
				float64(18446744073709551615),
				bigIntOrPanic("18446744073709551615"),
			},
		},
		{
			name:                    "fit Go uint64 with leading zeros",
			cborData:                hexDecode("c24900ffffffffffffffff"), // 2(18446744073709551615)
			wantEmptyInterfaceValue: bigIntOrPanic("18446744073709551615"),
			wantValues: []interface{}{
				uint64(18446744073709551615),
				float32(18446744073709551615),
				float64(18446744073709551615),
				bigIntOrPanic("18446744073709551615"),
			},
		},
		{
			name:                    "overflow Go uint64",
			cborData:                hexDecode("c249010000000000000000"), // 2(18446744073709551616)
			wantEmptyInterfaceValue: bigIntOrPanic("18446744073709551616"),
			wantValues: []interface{}{
				bigIntOrPanic("18446744073709551616"),
			},
		},
		{
			name:                    "overflow Go uint64 with leading zeros",
			cborData:                hexDecode("c24b0000010000000000000000"), // 2(18446744073709551616)
			wantEmptyInterfaceValue: bigIntOrPanic("18446744073709551616"),
			wantValues: []interface{}{
				bigIntOrPanic("18446744073709551616"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v1 interface{}
			if err := Unmarshal(tc.cborData, &v1); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %+v", tc.cborData, err)
			} else if !reflect.DeepEqual(v1, tc.wantEmptyInterfaceValue) {
				t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", tc.cborData, v1, v1, tc.wantEmptyInterfaceValue, tc.wantEmptyInterfaceValue)
			}

			for _, wantValue := range tc.wantValues {
				v := reflect.New(reflect.TypeOf(wantValue))
				if err := Unmarshal(tc.cborData, v.Interface()); err != nil {
					t.Errorf("Unmarshal(0x%x) returned error %+v", tc.cborData, err)
				} else if !reflect.DeepEqual(v.Elem().Interface(), wantValue) {
					t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", tc.cborData, v.Elem().Interface(), v.Elem().Interface(), wantValue, wantValue)
				}
			}
		})
	}
}

func TestUnmarshalTag3(t *testing.T) {
	testCases := []struct {
		name                    string
		cborData                []byte
		wantEmptyInterfaceValue interface{}
		wantValues              []interface{}
	}{
		{
			name:                    "fit Go int64",
			cborData:                hexDecode("c3487fffffffffffffff"), // 3(-9223372036854775808)
			wantEmptyInterfaceValue: bigIntOrPanic("-9223372036854775808"),
			wantValues: []interface{}{
				int64(-9223372036854775808),
				float32(-9223372036854775808),
				float64(-9223372036854775808),
				bigIntOrPanic("-9223372036854775808"),
			},
		},
		{
			name:                    "fit Go int64 with leading zeros",
			cborData:                hexDecode("c349007fffffffffffffff"), // 3(-9223372036854775808)
			wantEmptyInterfaceValue: bigIntOrPanic("-9223372036854775808"),
			wantValues: []interface{}{
				int64(-9223372036854775808),
				float32(-9223372036854775808),
				float64(-9223372036854775808),
				bigIntOrPanic("-9223372036854775808"),
			},
		},
		{
			name:                    "overflow Go int64",
			cborData:                hexDecode("c349010000000000000000"), // 3(-18446744073709551617)
			wantEmptyInterfaceValue: bigIntOrPanic("-18446744073709551617"),
			wantValues: []interface{}{
				bigIntOrPanic("-18446744073709551617"),
			},
		},
		{
			name:                    "overflow Go int64 with leading zeros",
			cborData:                hexDecode("c34b0000010000000000000000"), // 3(-18446744073709551617)
			wantEmptyInterfaceValue: bigIntOrPanic("-18446744073709551617"),
			wantValues: []interface{}{
				bigIntOrPanic("-18446744073709551617"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v1 interface{}
			if err := Unmarshal(tc.cborData, &v1); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %+v", tc.cborData, err)
			} else if !reflect.DeepEqual(v1, tc.wantEmptyInterfaceValue) {
				t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", tc.cborData, v1, v1, tc.wantEmptyInterfaceValue, tc.wantEmptyInterfaceValue)
			}

			for _, wantValue := range tc.wantValues {
				v := reflect.New(reflect.TypeOf(wantValue))
				if err := Unmarshal(tc.cborData, v.Interface()); err != nil {
					t.Errorf("Unmarshal(0x%x) returned error %+v", tc.cborData, err)
				} else if !reflect.DeepEqual(v.Elem().Interface(), wantValue) {
					t.Errorf("Unmarshal(0x%x) returned %v (%T), want %v (%T)", tc.cborData, v.Elem().Interface(), v.Elem().Interface(), wantValue, wantValue)
				}
			}
		})
	}
}

func TestUnmarshalInvalidTagBignum(t *testing.T) {
	typeBigIntSlice := reflect.TypeOf([]big.Int{})

	testCases := []struct {
		name          string
		cborData      []byte
		decodeToTypes []reflect.Type
		wantErrorMsg  string
	}{
		{
			name:          "Tag 2 with string",
			cborData:      hexDecode("c27f657374726561646d696e67ff"),
			decodeToTypes: []reflect.Type{typeIntf, typeBigInt},
			wantErrorMsg:  "cbor: tag number 2 or 3 must be followed by byte string, got UTF-8 text string",
		},
		{
			name:          "Tag 3 with string",
			cborData:      hexDecode("c37f657374726561646d696e67ff"),
			decodeToTypes: []reflect.Type{typeIntf, typeBigInt},
			wantErrorMsg:  "cbor: tag number 2 or 3 must be followed by byte string, got UTF-8 text string",
		},
		{
			name:          "Tag 3 with negavtive int",
			cborData:      hexDecode("81C330"), // [3(-17)]
			decodeToTypes: []reflect.Type{typeIntf, typeBigIntSlice},
			wantErrorMsg:  "cbor: tag number 2 or 3 must be followed by byte string, got negative integer",
		},
	}
	for _, tc := range testCases {
		for _, decodeToType := range tc.decodeToTypes {
			t.Run(tc.name+" decode to "+decodeToType.String(), func(t *testing.T) {
				v := reflect.New(decodeToType)
				if err := Unmarshal(tc.cborData, v.Interface()); err == nil {
					t.Errorf("Unmarshal(0x%x) didn't return error, want error msg %q", tc.cborData, tc.wantErrorMsg)
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.cborData, err, tc.wantErrorMsg)
				}
			})
		}
	}
}

type Foo interface {
	Foo() string
}

type UintFoo uint

func (f *UintFoo) Foo() string {
	return fmt.Sprint(f)
}

type IntFoo int

func (f *IntFoo) Foo() string {
	return fmt.Sprint(*f)
}

type ByteFoo []byte

func (f *ByteFoo) Foo() string {
	return fmt.Sprint(*f)
}

type StringFoo string

func (f *StringFoo) Foo() string {
	return string(*f)
}

type ArrayFoo []int

func (f *ArrayFoo) Foo() string {
	return fmt.Sprint(*f)
}

type MapFoo map[int]int

func (f *MapFoo) Foo() string {
	return fmt.Sprint(*f)
}

type StructFoo struct {
	Value int `cbor:"1,keyasint"`
}

func (f *StructFoo) Foo() string {
	return fmt.Sprint(*f)
}

type TestExample struct {
	Message string `cbor:"1,keyasint"`
	Foo     Foo    `cbor:"2,keyasint"`
}

func TestUnmarshalToInterface(t *testing.T) {

	uintFoo, uintFoo123 := UintFoo(0), UintFoo(123)
	intFoo, intFooNeg1 := IntFoo(0), IntFoo(-1)
	byteFoo, byteFoo123 := ByteFoo(nil), ByteFoo([]byte{1, 2, 3})
	stringFoo, stringFoo123 := StringFoo(""), StringFoo("123")
	arrayFoo, arrayFoo123 := ArrayFoo(nil), ArrayFoo([]int{1, 2, 3})
	mapFoo, mapFoo123 := MapFoo(nil), MapFoo(map[int]int{1: 1, 2: 2, 3: 3})

	em, _ := EncOptions{Sort: SortCanonical}.EncMode()

	testCases := []struct {
		name           string
		data           []byte
		v              *TestExample
		unmarshalToObj *TestExample
	}{
		{
			name: "uint",
			data: hexDecode("a2016c736f6d65206d65737361676502187b"), // {1: "some message", 2: 123}
			v: &TestExample{
				Message: "some message",
				Foo:     &uintFoo123,
			},
			unmarshalToObj: &TestExample{Foo: &uintFoo},
		},
		{
			name: "int",
			data: hexDecode("a2016c736f6d65206d6573736167650220"), // {1: "some message", 2: -1}
			v: &TestExample{
				Message: "some message",
				Foo:     &intFooNeg1,
			},
			unmarshalToObj: &TestExample{Foo: &intFoo},
		},
		{
			name: "bytes",
			data: hexDecode("a2016c736f6d65206d6573736167650243010203"), // {1: "some message", 2: [1,2,3]}
			v: &TestExample{
				Message: "some message",
				Foo:     &byteFoo123,
			},
			unmarshalToObj: &TestExample{Foo: &byteFoo},
		},
		{
			name: "string",
			data: hexDecode("a2016c736f6d65206d6573736167650263313233"), // {1: "some message", 2: "123"}
			v: &TestExample{
				Message: "some message",
				Foo:     &stringFoo123,
			},
			unmarshalToObj: &TestExample{Foo: &stringFoo},
		},
		{
			name: "array",
			data: hexDecode("a2016c736f6d65206d6573736167650283010203"), // {1: "some message", 2: []int{1,2,3}}
			v: &TestExample{
				Message: "some message",
				Foo:     &arrayFoo123,
			},
			unmarshalToObj: &TestExample{Foo: &arrayFoo},
		},
		{
			name: "map",
			data: hexDecode("a2016c736f6d65206d65737361676502a3010102020303"), // {1: "some message", 2: map[int]int{1:1,2:2,3:3}}
			v: &TestExample{
				Message: "some message",
				Foo:     &mapFoo123,
			},
			unmarshalToObj: &TestExample{Foo: &mapFoo},
		},
		{
			name: "struct",
			data: hexDecode("a2016c736f6d65206d65737361676502a1011901c8"), // {1: "some message", 2: {1: 456}}
			v: &TestExample{
				Message: "some message",
				Foo:     &StructFoo{Value: 456},
			},
			unmarshalToObj: &TestExample{Foo: &StructFoo{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			data, err := em.Marshal(tc.v)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.v, err)
			} else if !bytes.Equal(data, tc.data) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.v, data, tc.v)
			}

			// Unmarshal to empty interface
			var einterface TestExample
			if err = Unmarshal(data, &einterface); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error (*UnmarshalTypeError)", data)
			} else if _, ok := err.(*UnmarshalTypeError); !ok {
				t.Errorf("Unmarshal(0x%x) returned wrong type of error %T, want (*UnmarshalTypeError)", data, err)
			}

			// Unmarshal to interface value
			err = Unmarshal(data, tc.unmarshalToObj)
			if err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", data, err)
			} else if !reflect.DeepEqual(tc.unmarshalToObj, tc.v) {
				t.Errorf("Unmarshal(0x%x) = %v, want %v", data, tc.unmarshalToObj, tc.v)
			}
		})
	}
}

type Bar struct {
	I int
}

func (b *Bar) Foo() string {
	return fmt.Sprint(*b)
}

type FooStruct struct {
	Foos []Foo
}

func TestUnmarshalTaggedDataToInterface(t *testing.T) {

	var tags = NewTagSet()
	err := tags.Add(
		TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired},
		reflect.TypeOf(&Bar{}),
		4,
	)
	if err != nil {
		t.Error(err)
	}

	v := &FooStruct{
		Foos: []Foo{&Bar{1}},
	}

	want := hexDecode("a164466f6f7381c4a1614901") // {"Foos": [4({"I": 1})]}

	em, _ := EncOptions{}.EncModeWithTags(tags)
	data, err := em.Marshal(v)
	if err != nil {
		t.Errorf("Marshal(%+v) returned error %v", v, err)
	} else if !bytes.Equal(data, want) {
		t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", v, data, want)
	}

	dm, _ := DecOptions{}.DecModeWithTags(tags)

	// Unmarshal to empty interface
	var v1 Bar
	if err = dm.Unmarshal(data, &v1); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error, want error (*UnmarshalTypeError)", data)
	} else if _, ok := err.(*UnmarshalTypeError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong type of error %T, want (*UnmarshalTypeError)", data, err)
	}

	// Unmarshal to interface value
	v2 := &FooStruct{
		Foos: []Foo{&Bar{}},
	}
	err = dm.Unmarshal(data, v2)
	if err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", data, err)
	} else if !reflect.DeepEqual(v2, v) {
		t.Errorf("Unmarshal(0x%x) = %v, want %v", data, v2, v)
	}
}

type B interface {
	Foo()
}

type C struct {
	Field int
}

func (c *C) Foo() {}

type D struct {
	Field string
}

func (d *D) Foo() {}

type A1 struct {
	Field B
}

type A2 struct {
	Fields []B
}

func TestUnmarshalRegisteredTagToInterface(t *testing.T) {
	var err error
	tags := NewTagSet()
	err = tags.Add(TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, reflect.TypeOf(C{}), 279)
	if err != nil {
		t.Error(err)
	}
	err = tags.Add(TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, reflect.TypeOf(D{}), 280)
	if err != nil {
		t.Error(err)
	}

	encMode, _ := PreferredUnsortedEncOptions().EncModeWithTags(tags)
	decMode, _ := DecOptions{}.DecModeWithTags(tags)

	v1 := A1{Field: &C{Field: 5}}
	data1, err := encMode.Marshal(v1)
	if err != nil {
		t.Fatalf("Marshal(%+v) returned error %v", v1, err)
	}

	v2 := A2{Fields: []B{&C{Field: 5}, &D{Field: "a"}}}
	data2, err := encMode.Marshal(v2)
	if err != nil {
		t.Fatalf("Marshal(%+v) returned error %v", v2, err)
	}

	testCases := []struct {
		name           string
		data           []byte
		unmarshalToObj interface{}
		wantValue      interface{}
	}{
		{
			name:           "interface type",
			data:           data1,
			unmarshalToObj: &A1{},
			wantValue:      &v1,
		},
		{
			name:           "concrete type",
			data:           data1,
			unmarshalToObj: &A1{Field: &C{}},
			wantValue:      &v1,
		},
		{
			name:           "slice of interface type",
			data:           data2,
			unmarshalToObj: &A2{},
			wantValue:      &v2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err = decMode.Unmarshal(tc.data, tc.unmarshalToObj)
			if err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.data, err)
			}
			if !reflect.DeepEqual(tc.unmarshalToObj, tc.wantValue) {
				t.Errorf("Unmarshal(0x%x) = %v, want %v", tc.data, tc.unmarshalToObj, tc.wantValue)
			}
		})
	}
}

func TestDecModeInvalidDefaultMapType(t *testing.T) {
	testCases := []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "byte slice",
			opts:         DecOptions{DefaultMapType: reflect.TypeOf([]byte(nil))},
			wantErrorMsg: "cbor: invalid DefaultMapType []uint8",
		},
		{
			name:         "int slice",
			opts:         DecOptions{DefaultMapType: reflect.TypeOf([]int(nil))},
			wantErrorMsg: "cbor: invalid DefaultMapType []int",
		},
		{
			name:         "string",
			opts:         DecOptions{DefaultMapType: reflect.TypeOf("")},
			wantErrorMsg: "cbor: invalid DefaultMapType string",
		},
		{
			name:         "unnamed struct type",
			opts:         DecOptions{DefaultMapType: reflect.TypeOf(struct{}{})},
			wantErrorMsg: "cbor: invalid DefaultMapType struct {}",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestUnmarshalToDefaultMapType(t *testing.T) {

	cborDataMapIntInt := hexDecode("a201020304")                                             // {1: 2, 3: 4}
	cborDataMapStringInt := hexDecode("a2616101616202")                                      // {"a": 1, "b": 2}
	cborDataArrayOfMapStringint := hexDecode("82a2616101616202a2616303616404")               // [{"a": 1, "b": 2}, {"c": 3, "d": 4}]
	cborDataNestedMap := hexDecode("a268496e744669656c6401684d61704669656c64a2616101616202") // {"IntField": 1, "MapField": {"a": 1, "b": 2}}

	decOptionsDefault := DecOptions{}
	decOptionsMapIntfIntfType := DecOptions{DefaultMapType: reflect.TypeOf(map[interface{}]interface{}(nil))}
	decOptionsMapStringIntType := DecOptions{DefaultMapType: reflect.TypeOf(map[string]int(nil))}
	decOptionsMapStringIntfType := DecOptions{DefaultMapType: reflect.TypeOf(map[string]interface{}(nil))}

	testCases := []struct {
		name         string
		opts         DecOptions
		cborData     []byte
		wantValue    interface{}
		wantErrorMsg string
	}{
		// Decode CBOR map to map[interface{}]interface{} using default options
		{
			name:      "decode CBOR map[int]int to Go map[interface{}]interface{} (default)",
			opts:      decOptionsDefault,
			cborData:  cborDataMapIntInt,
			wantValue: map[interface{}]interface{}{uint64(1): uint64(2), uint64(3): uint64(4)},
		},
		{
			name:      "decode CBOR map[string]int to Go map[interface{}]interface{} (default)",
			opts:      decOptionsDefault,
			cborData:  cborDataMapStringInt,
			wantValue: map[interface{}]interface{}{"a": uint64(1), "b": uint64(2)},
		},
		{
			name:     "decode CBOR array of map[string]int to Go []map[interface{}]interface{} (default)",
			opts:     decOptionsDefault,
			cborData: cborDataArrayOfMapStringint,
			wantValue: []interface{}{
				map[interface{}]interface{}{"a": uint64(1), "b": uint64(2)},
				map[interface{}]interface{}{"c": uint64(3), "d": uint64(4)},
			},
		},
		{
			name:     "decode CBOR nested map to Go map[interface{}]interface{} (default)",
			opts:     decOptionsDefault,
			cborData: cborDataNestedMap,
			wantValue: map[interface{}]interface{}{
				"IntField": uint64(1),
				"MapField": map[interface{}]interface{}{"a": uint64(1), "b": uint64(2)},
			},
		},
		// Decode CBOR map to map[interface{}]interface{} using default map type option
		{
			name:      "decode CBOR map[int]int to Go map[interface{}]interface{}",
			opts:      decOptionsMapIntfIntfType,
			cborData:  cborDataMapIntInt,
			wantValue: map[interface{}]interface{}{uint64(1): uint64(2), uint64(3): uint64(4)},
		},
		{
			name:      "decode CBOR map[string]int to Go map[interface{}]interface{}",
			opts:      decOptionsMapIntfIntfType,
			cborData:  cborDataMapStringInt,
			wantValue: map[interface{}]interface{}{"a": uint64(1), "b": uint64(2)},
		},
		{
			name:     "decode CBOR array of map[string]int to Go []map[interface{}]interface{}",
			opts:     decOptionsMapIntfIntfType,
			cborData: cborDataArrayOfMapStringint,
			wantValue: []interface{}{
				map[interface{}]interface{}{"a": uint64(1), "b": uint64(2)},
				map[interface{}]interface{}{"c": uint64(3), "d": uint64(4)},
			},
		},
		{
			name:     "decode CBOR nested map to Go map[interface{}]interface{}",
			opts:     decOptionsMapIntfIntfType,
			cborData: cborDataNestedMap,
			wantValue: map[interface{}]interface{}{
				"IntField": uint64(1),
				"MapField": map[interface{}]interface{}{"a": uint64(1), "b": uint64(2)},
			},
		},
		// Decode CBOR map to map[string]interface{} using default map type option
		{
			name:         "decode CBOR map[int]int to Go map[string]interface{}",
			opts:         decOptionsMapStringIntfType,
			cborData:     cborDataMapIntInt,
			wantErrorMsg: "cbor: cannot unmarshal positive integer into Go value of type string",
		},
		{
			name:      "decode CBOR map[string]int to Go map[string]interface{}",
			opts:      decOptionsMapStringIntfType,
			cborData:  cborDataMapStringInt,
			wantValue: map[string]interface{}{"a": uint64(1), "b": uint64(2)},
		},
		{
			name:     "decode CBOR array of map[string]int to Go []map[string]interface{}",
			opts:     decOptionsMapStringIntfType,
			cborData: cborDataArrayOfMapStringint,
			wantValue: []interface{}{
				map[string]interface{}{"a": uint64(1), "b": uint64(2)},
				map[string]interface{}{"c": uint64(3), "d": uint64(4)},
			},
		},
		{
			name:     "decode CBOR nested map to Go map[string]interface{}",
			opts:     decOptionsMapStringIntfType,
			cborData: cborDataNestedMap,
			wantValue: map[string]interface{}{
				"IntField": uint64(1),
				"MapField": map[string]interface{}{"a": uint64(1), "b": uint64(2)},
			},
		},
		// Decode CBOR map to map[string]int using default map type option
		{
			name:         "decode CBOR map[int]int to Go map[string]int",
			opts:         decOptionsMapStringIntType,
			cborData:     cborDataMapIntInt,
			wantErrorMsg: "cbor: cannot unmarshal positive integer into Go value of type string",
		},
		{
			name:      "decode CBOR map[string]int to Go map[string]int",
			opts:      decOptionsMapStringIntType,
			cborData:  cborDataMapStringInt,
			wantValue: map[string]int{"a": 1, "b": 2},
		},
		{
			name:     "decode CBOR array of map[string]int to Go []map[string]int",
			opts:     decOptionsMapStringIntType,
			cborData: cborDataArrayOfMapStringint,
			wantValue: []interface{}{
				map[string]int{"a": 1, "b": 2},
				map[string]int{"c": 3, "d": 4},
			},
		},
		{
			name:         "decode CBOR nested map to Go map[string]int",
			opts:         decOptionsMapStringIntType,
			cborData:     cborDataNestedMap,
			wantErrorMsg: "cbor: cannot unmarshal map into Go value of type int",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decMode, _ := tc.opts.DecMode()

			var v interface{}
			err := decMode.Unmarshal(tc.cborData, &v)
			if err != nil {
				if tc.wantErrorMsg == "" {
					t.Errorf("Unmarshal(0x%x) to empty interface returned error %v", tc.cborData, err)
				} else if tc.wantErrorMsg != err.Error() {
					t.Errorf("Unmarshal(0x%x) error %q, want %q", tc.cborData, err.Error(), tc.wantErrorMsg)
				}
			} else {
				if tc.wantValue == nil {
					t.Errorf("Unmarshal(0x%x) = %v (%T), want error %q", tc.cborData, v, v, tc.wantErrorMsg)
				} else if !reflect.DeepEqual(v, tc.wantValue) {
					t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborData, v, v, tc.wantValue, tc.wantValue)
				}
			}
		})
	}
}
