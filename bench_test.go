// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"
)

const rounds = 100

type claims struct {
	Iss string `cbor:"1,keyasint"`
	Sub string `cbor:"2,keyasint"`
	Aud string `cbor:"3,keyasint"`
	Exp int    `cbor:"4,keyasint"`
	Nbf int    `cbor:"5,keyasint"`
	Iat int    `cbor:"6,keyasint"`
	Cti []byte `cbor:"7,keyasint"`
}

type coseHeader struct {
	Alg int    `cbor:"1,keyasint,omitempty"`
	Kid []byte `cbor:"4,keyasint,omitempty"`
	IV  []byte `cbor:"5,keyasint,omitempty"`
}

type macedCOSE struct {
	_           struct{} `cbor:",toarray"`
	Protected   []byte
	Unprotected coseHeader
	Payload     []byte
	Tag         []byte
}

type coseKey struct {
	Kty       int        `cbor:"1,keyasint,omitempty"`
	Kid       []byte     `cbor:"2,keyasint,omitempty"`
	Alg       int        `cbor:"3,keyasint,omitempty"`
	KeyOpts   int        `cbor:"4,keyasint,omitempty"`
	IV        []byte     `cbor:"5,keyasint,omitempty"`
	CrvOrNOrK RawMessage `cbor:"-1,keyasint,omitempty"` // K for symmetric keys, Crv for elliptic curve keys, N for RSA modulus
	XOrE      RawMessage `cbor:"-2,keyasint,omitempty"` // X for curve x-coordinate, E for RSA public exponent
	Y         RawMessage `cbor:"-3,keyasint,omitempty"` // Y for curve y-coordinate
	D         []byte     `cbor:"-4,keyasint,omitempty"`
}

type attestationObject struct {
	AuthnData []byte     `cbor:"authData"`
	Fmt       string     `cbor:"fmt"`
	AttStmt   RawMessage `cbor:"attStmt"`
}

type SenMLRecord struct {
	BaseName    string  `cbor:"-2,keyasint,omitempty"`
	BaseTime    float64 `cbor:"-3,keyasint,omitempty"`
	BaseUnit    string  `cbor:"-4,keyasint,omitempty"`
	BaseValue   float64 `cbor:"-5,keyasint,omitempty"`
	BaseSum     float64 `cbor:"-6,keyasint,omitempty"`
	BaseVersion int     `cbor:"-1,keyasint,omitempty"`
	Name        string  `cbor:"0,keyasint,omitempty"`
	Unit        string  `cbor:"1,keyasint,omitempty"`
	Time        float64 `cbor:"6,keyasint,omitempty"`
	UpdateTime  float64 `cbor:"7,keyasint,omitempty"`
	Value       float64 `cbor:"2,keyasint,omitempty"`
	ValueS      string  `cbor:"3,keyasint,omitempty"`
	ValueB      bool    `cbor:"4,keyasint,omitempty"`
	ValueD      string  `cbor:"8,keyasint,omitempty"`
	Sum         float64 `cbor:"5,keyasint,omitempty"`
}

type T1 struct {
	T    bool
	UI   uint
	I    int
	F    float64
	B    []byte
	S    string
	Slci []int
	Mss  map[string]string
}

type T2 struct {
	T    bool              `cbor:"1,keyasint"`
	UI   uint              `cbor:"2,keyasint"`
	I    int               `cbor:"3,keyasint"`
	F    float64           `cbor:"4,keyasint"`
	B    []byte            `cbor:"5,keyasint"`
	S    string            `cbor:"6,keyasint"`
	Slci []int             `cbor:"7,keyasint"`
	Mss  map[string]string `cbor:"8,keyasint"`
}

type T3 struct {
	_    struct{} `cbor:",toarray"`
	T    bool
	UI   uint
	I    int
	F    float64
	B    []byte
	S    string
	Slci []int
	Mss  map[string]string
}

type ManyFieldsOneOmitEmpty struct {
	F01, F02, F03, F04, F05, F06, F07, F08, F09, F10, F11, F12, F13, F14, F15, F16 int
	F17, F18, F19, F20, F21, F22, F23, F24, F25, F26, F27, F28, F29, F30, F31      int

	F32 int `cbor:",omitempty"`
}

type SomeFieldsOneOmitEmpty struct {
	F01, F02, F03, F04, F05, F06, F07 int

	F08 int `cbor:",omitempty"`
}

type ManyFieldsAllOmitEmpty struct {
	F01 int `cbor:",omitempty"`
	F02 int `cbor:",omitempty"`
	F03 int `cbor:",omitempty"`
	F04 int `cbor:",omitempty"`
	F05 int `cbor:",omitempty"`
	F06 int `cbor:",omitempty"`
	F07 int `cbor:",omitempty"`
	F08 int `cbor:",omitempty"`
	F09 int `cbor:",omitempty"`
	F10 int `cbor:",omitempty"`
	F11 int `cbor:",omitempty"`
	F12 int `cbor:",omitempty"`
	F13 int `cbor:",omitempty"`
	F14 int `cbor:",omitempty"`
	F15 int `cbor:",omitempty"`
	F16 int `cbor:",omitempty"`
	F17 int `cbor:",omitempty"`
	F18 int `cbor:",omitempty"`
	F19 int `cbor:",omitempty"`
	F20 int `cbor:",omitempty"`
	F21 int `cbor:",omitempty"`
	F22 int `cbor:",omitempty"`
	F23 int `cbor:",omitempty"`
	F24 int `cbor:",omitempty"`
	F25 int `cbor:",omitempty"`
	F26 int `cbor:",omitempty"`
	F27 int `cbor:",omitempty"`
	F28 int `cbor:",omitempty"`
	F29 int `cbor:",omitempty"`
	F30 int `cbor:",omitempty"`
	F31 int `cbor:",omitempty"`
	F32 int `cbor:",omitempty"`
}

type SomeFieldsAllOmitEmpty struct {
	F01 int `cbor:",omitempty"`
	F02 int `cbor:",omitempty"`
	F03 int `cbor:",omitempty"`
	F04 int `cbor:",omitempty"`
	F05 int `cbor:",omitempty"`
	F06 int `cbor:",omitempty"`
	F07 int `cbor:",omitempty"`
	F08 int `cbor:",omitempty"`
}

// address is a defined type with underlying []byte.
// address has 8 methods and doesn't implement any Unmarshaler or Marshaler interfaces.
type address []byte

func (a address) Bytes() []byte         { return append([]byte(nil), a...) }
func (a address) Compare(x address) int { return bytes.Compare(a, x) }
func (a address) Equal(x address) bool  { return bytes.Equal(a, x) }
func (a address) Hex() string           { return hex.EncodeToString([]byte(a)) }
func (a address) HexWithPrefix() string { return fmt.Sprintf("0x%x", []byte(a)) }
func (a address) Len() int              { return len(a) }
func (a *address) Reset(x address)      { *a = append((*a)[:0], x...) }
func (a address) String() string        { return fmt.Sprintf("0x%x", []byte(a)) }

// valueMarshaler implements Marshaler with a value receiver.
type valueMarshaler [4]byte

func (m valueMarshaler) MarshalCBOR() ([]byte, error) {
	return []byte{0x44, m[0], m[1], m[2], m[3]}, nil
}

// ptrMarshaler implements Marshaler with a pointer receiver.
type ptrMarshaler [4]byte

func (m *ptrMarshaler) MarshalCBOR() ([]byte, error) {
	return []byte{0x44, m[0], m[1], m[2], m[3]}, nil
}

var decodeBenchmarks = []struct {
	name          string
	data          []byte
	decodeToTypes []reflect.Type
}{
	{
		name:          "bool",
		data:          mustHexDecode("f5"),
		decodeToTypes: []reflect.Type{typeAny, typeBool},
	}, // true
	{
		name:          "uint",
		data:          mustHexDecode("1bffffffffffffffff"),
		decodeToTypes: []reflect.Type{typeAny, typeUint64},
	}, // uint64(18446744073709551615)
	{
		name:          "nint",
		data:          mustHexDecode("3903e7"),
		decodeToTypes: []reflect.Type{typeAny, typeInt64},
	}, // int64(-1000)
	{
		name:          "float",
		data:          mustHexDecode("fbc010666666666666"),
		decodeToTypes: []reflect.Type{typeAny, typeFloat64},
	}, // float64(-4.1)
	{
		name:          "bstr",
		data:          mustHexDecode("581a0102030405060708090a0b0c0d0e0f101112131415161718191a"),
		decodeToTypes: []reflect.Type{typeAny, typeByteSlice, reflect.TypeFor[address]()},
	}, // []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}
	{
		name:          "indef bstr",
		data:          mustHexDecode("5f410141024103410441054106410741084109410a410b410c410d410e410f4110411141124113411441154116411741184119411aff"),
		decodeToTypes: []reflect.Type{typeAny, typeByteSlice},
	}, // []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}
	{
		name:          "tstr",
		data:          mustHexDecode("782b54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f67"),
		decodeToTypes: []reflect.Type{typeAny, typeString},
	}, // "The quick brown fox jumps over the lazy dog"
	{
		name:          "indef tstr",
		data:          mustHexDecode("7f61546168616561206171617561696163616b612061626172616f6177616e61206166616f61786120616a6175616d617061736120616f61766165617261206174616861656120616c6161617a617961206164616f6167ff"),
		decodeToTypes: []reflect.Type{typeAny, typeString},
	}, // "The quick brown fox jumps over the lazy dog"
	{
		name:          "array",
		data:          mustHexDecode("981a0102030405060708090a0b0c0d0e0f101112131415161718181819181a"),
		decodeToTypes: []reflect.Type{typeAny, typeIntSlice},
	}, // []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}
	{
		name:          "indef array",
		data:          mustHexDecode("9f0102030405060708090a0b0c0d0e0f101112131415161718181819181aff"),
		decodeToTypes: []reflect.Type{typeAny, typeIntSlice},
	}, // []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}
	{
		name:          "map",
		data:          mustHexDecode("ad616161416162614261636143616461446165614561666146616761476168614861696149616a614a616c614c616d614d616e614e"),
		decodeToTypes: []reflect.Type{typeAny, typeMapStringAny, typeMapStringString},
	}, // map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"}}
	{
		name:          "indef map",
		data:          mustHexDecode("bf616161416162614261636143616461446165614561666146616761476168614861696149616a614a616b614b616c614c616d614d616e614eff"),
		decodeToTypes: []reflect.Type{typeAny, typeMapStringAny, typeMapStringString},
	}, // map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"}}
}

var encodeBenchmarks = []struct {
	name   string
	data   []byte
	values []any
}{
	{
		name:   "bool",
		data:   mustHexDecode("f5"),
		values: []any{true},
	},
	{
		name:   "uint",
		data:   mustHexDecode("1bffffffffffffffff"),
		values: []any{uint64(18446744073709551615)},
	},
	{
		name:   "nint",
		data:   mustHexDecode("3903e7"),
		values: []any{int64(-1000)},
	},
	{
		name:   "float",
		data:   mustHexDecode("fbc010666666666666"),
		values: []any{float64(-4.1)},
	},
	{
		name:   "bstr",
		data:   mustHexDecode("581a0102030405060708090a0b0c0d0e0f101112131415161718191a"),
		values: []any{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}},
	},
	{
		name:   "tstr",
		data:   mustHexDecode("782b54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f67"),
		values: []any{"The quick brown fox jumps over the lazy dog"},
	},
	{
		name:   "array",
		data:   mustHexDecode("981a0102030405060708090a0b0c0d0e0f101112131415161718181819181a"),
		values: []any{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}},
	},
	{
		name:   "map",
		data:   mustHexDecode("ad616161416162614261636143616461446165614561666146616761476168614861696149616a614a616c614c616d614d616e614e"),
		values: []any{map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"}},
	},
}

func BenchmarkUnmarshal(b *testing.B) {
	for _, bm := range decodeBenchmarks {
		for _, t := range bm.decodeToTypes {
			// Name uses the format: "Unmarshal/CBOR type & encoding/Go type",
			// e.g. "Unmarshal/bstr/[]byte", "Unmarshal/indef_bstr/[]byte"
			name := bm.name + "/" + goTypeName(t)
			b.Run(name, func(b *testing.B) {
				for b.Loop() {
					vPtr := reflect.New(t).Interface()
					if err := Unmarshal(bm.data, vPtr); err != nil {
						b.Fatal("Unmarshal:", err)
					}
				}
			})
		}
	}
	var moreBenchmarks = []struct {
		name         string
		data         []byte
		decodeToType reflect.Type
	}{
		// Unmarshal CBOR map with string key to map[string]any.
		{
			name:         "mixed map/map[string]any",
			data:         mustHexDecode("a86154f56255691bffffffffffffffff61493903e76146fbc0106666666666666142581a0102030405060708090a0b0c0d0e0f101112131415161718191a6153782b54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f6764536c6369981a0102030405060708090a0b0c0d0e0f101112131415161718181819181a634d7373ad6163614361656145616661466167614761686148616e614e616d614d61616141616261426164614461696149616a614a616c614c"),
			decodeToType: reflect.TypeFor[map[string]any](),
		},
		// Unmarshal CBOR map with string key to struct.
		{
			name:         "mixed map/struct",
			data:         mustHexDecode("a86154f56255491bffffffffffffffff61493903e76146fbc0106666666666666142581a0102030405060708090a0b0c0d0e0f101112131415161718191a6153782b54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f6764536c6369981a0102030405060708090a0b0c0d0e0f101112131415161718181819181a634d7373ad6163614361656145616661466167614761686148616e614e616d614d61616141616261426164614461696149616a614a616c614c"),
			decodeToType: reflect.TypeFor[T1](),
		},
		// Unmarshal CBOR map with integer key, such as COSE Key and SenML, to map[int]any.
		{
			name:         "mixed map/map[int]any",
			data:         mustHexDecode("a801f5021bffffffffffffffff033903e704fbc01066666666666605581a0102030405060708090a0b0c0d0e0f101112131415161718191a06782b54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f6707981a0102030405060708090a0b0c0d0e0f101112131415161718181819181a08ad61646144616661466167614761686148616d614d616e614e6161614161626142616361436165614561696149616a614a616c614c"),
			decodeToType: reflect.TypeFor[map[int]any](),
		},
		// Unmarshal CBOR map with integer key, such as COSE Key and SenML, to struct.
		{
			name:         "mixed map/struct keyasint",
			data:         mustHexDecode("a801f5021bffffffffffffffff033903e704fbc01066666666666605581a0102030405060708090a0b0c0d0e0f101112131415161718191a06782b54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f6707981a0102030405060708090a0b0c0d0e0f101112131415161718181819181a08ad61646144616661466167614761686148616d614d616e614e6161614161626142616361436165614561696149616a614a616c614c"),
			decodeToType: reflect.TypeFor[T2](),
		},
		// Unmarshal CBOR array of known sequence of data types, such as signed/maced/encrypted CWT, to []any.
		{
			name:         "mixed array/[]any",
			data:         mustHexDecode("88f51bffffffffffffffff3903e7fbc010666666666666581a0102030405060708090a0b0c0d0e0f101112131415161718191a782b54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f67981a0102030405060708090a0b0c0d0e0f101112131415161718181819181aad616261426163614361646144616561456166614661696149616e614e616161416167614761686148616a614a616c614c616d614d"),
			decodeToType: reflect.TypeFor[[]any](),
		},
		// Unmarshal CBOR array of known sequence of data types, such as signed/maced/encrypted CWT, to struct.
		{
			name:         "mixed array/struct toarray",
			data:         mustHexDecode("88f51bffffffffffffffff3903e7fbc010666666666666581a0102030405060708090a0b0c0d0e0f101112131415161718191a782b54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f67981a0102030405060708090a0b0c0d0e0f101112131415161718181819181aad616261426163614361646144616561456166614661696149616e614e616161416167614761686148616a614a616c614c616d614d"),
			decodeToType: reflect.TypeFor[T3](),
		},
		// Unmarshal CBOR epoch-based date/time (tag 1) to time.Time and any.
		{
			name:         "epoch time/time.Time",
			data:         mustHexDecode("c11a514b67b0"),
			decodeToType: typeTime,
		},
		{
			name:         "epoch time/any",
			data:         mustHexDecode("c11a514b67b0"),
			decodeToType: typeAny,
		},
		// Unmarshal CBOR bignum (tag 2) to big.Int and any.
		{
			name:         "bignum/big.Int",
			data:         mustHexDecode("c249010000000000000000"),
			decodeToType: typeBigInt,
		},
		{
			name:         "bignum/any",
			data:         mustHexDecode("c249010000000000000000"),
			decodeToType: typeAny,
		},
		// Unmarshal CBOR tag not registered in TagSet to Tag and any.
		{
			name:         "tag/Tag",
			data:         mustHexDecode("d8641a514b67b0"),
			decodeToType: typeTag,
		},
		{
			name:         "tag/any",
			data:         mustHexDecode("d8641a514b67b0"),
			decodeToType: typeAny,
		},
	}
	for _, bm := range moreBenchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				vPtr := reflect.New(bm.decodeToType).Interface()
				if err := Unmarshal(bm.data, vPtr); err != nil {
					b.Fatal("Unmarshal:", err)
				}
			}
		})
	}
}

func BenchmarkUnmarshalFirst(b *testing.B) {
	// Random trailing data
	trailingData := mustHexDecode("4a6b0f4718c73f391091ea1c")
	for _, bm := range decodeBenchmarks {
		for _, t := range bm.decodeToTypes {
			// Name uses the format: "UnmarshalFirst/CBOR type & encoding/Go type",
			// e.g. "UnmarshalFirst/bstr/[]byte", "UnmarshalFirst/indef_bstr/[]byte"
			name := bm.name + "/" + goTypeName(t)
			data := make([]byte, 0, len(bm.data)+len(trailingData))
			data = append(data, bm.data...)
			data = append(data, trailingData...)
			b.Run(name, func(b *testing.B) {
				for b.Loop() {
					vPtr := reflect.New(t).Interface()
					if _, err := UnmarshalFirst(data, vPtr); err != nil {
						b.Fatal("UnmarshalFirst:", err)
					}
				}
			})
		}
	}
}

func BenchmarkUnmarshalFirstViaDecoder(b *testing.B) {
	// Random trailing data
	trailingData := mustHexDecode("4a6b0f4718c73f391091ea1c")
	for _, bm := range decodeBenchmarks {
		for _, t := range bm.decodeToTypes {
			// Name uses the format: "UnmarshalFirstViaDecoder/CBOR type & encoding/Go type",
			// e.g. "UnmarshalFirstViaDecoder/bstr/[]byte", "UnmarshalFirstViaDecoder/indef_bstr/[]byte"
			name := bm.name + "/" + goTypeName(t)
			data := make([]byte, 0, len(bm.data)+len(trailingData))
			data = append(data, bm.data...)
			data = append(data, trailingData...)
			b.Run(name, func(b *testing.B) {
				for b.Loop() {
					vPtr := reflect.New(t).Interface()
					if err := NewDecoder(bytes.NewReader(data)).Decode(vPtr); err != nil {
						b.Fatal("UnmarshalDecoder:", err)
					}
				}
			})
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	for _, bm := range decodeBenchmarks {
		for _, t := range bm.decodeToTypes {
			// Name uses the format: "Decode/CBOR type & encoding/Go type",
			// e.g. "Decode/bstr/[]byte", "Decode/indef_bstr/[]byte"
			name := bm.name + "/" + goTypeName(t)
			buf := bytes.NewReader(bm.data)
			decoder := NewDecoder(buf)
			b.Run(name, func(b *testing.B) {
				for b.Loop() {
					vPtr := reflect.New(t).Interface()
					if err := decoder.Decode(vPtr); err != nil {
						b.Fatal("Decode:", err)
					}
					buf.Seek(0, 0) //nolint:errcheck
				}
			})
		}
	}
}

func BenchmarkDecodeStream(b *testing.B) {
	var data []byte
	for _, bm := range decodeBenchmarks {
		for range len(bm.decodeToTypes) {
			data = append(data, bm.data...)
		}
	}

	for b.Loop() {
		buf := bytes.NewReader(data)
		decoder := NewDecoder(buf)
		for range rounds {
			for _, bm := range decodeBenchmarks {
				for _, t := range bm.decodeToTypes {
					vPtr := reflect.New(t).Interface()
					if err := decoder.Decode(vPtr); err != nil {
						b.Fatal("Decode:", err)
					}
				}
			}
			buf.Seek(0, 0) //nolint:errcheck
		}
	}
}

func BenchmarkMarshal(b *testing.B) {
	for _, bm := range encodeBenchmarks {
		for _, v := range bm.values {
			// Name uses the format: "Marshal/Go type/CBOR type",
			// e.g. "Marshal/[]byte/bstr"
			name := goTypeName(reflect.TypeOf(v)) + "/" + bm.name
			b.Run(name, func(b *testing.B) {
				for b.Loop() {
					if _, err := Marshal(v); err != nil {
						b.Fatal("Marshal:", err)
					}
				}
			})
		}
	}
	// Marshal map[string]any to CBOR map
	m1 := map[string]any{
		"T":    true,
		"UI":   uint(18446744073709551615),
		"I":    -1000,
		"F":    -4.1,
		"B":    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		"S":    "The quick brown fox jumps over the lazy dog",
		"Slci": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		"Mss":  map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"},
	}
	// Marshal struct to CBOR map
	v1 := T1{ //nolint:dupl
		T:    true,
		UI:   18446744073709551615,
		I:    -1000,
		F:    -4.1,
		B:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		S:    "The quick brown fox jumps over the lazy dog",
		Slci: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		Mss:  map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"},
	}
	// Marshal map[int]any to CBOR map
	m2 := map[int]any{
		1: true,
		2: uint(18446744073709551615),
		3: -1000,
		4: -4.1,
		5: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		6: "The quick brown fox jumps over the lazy dog",
		7: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		8: map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"},
	}
	// Marshal struct keyasint, such as COSE Key and SenML
	v2 := T2{ //nolint:dupl
		T:    true,
		UI:   18446744073709551615,
		I:    -1000,
		F:    -4.1,
		B:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		S:    "The quick brown fox jumps over the lazy dog",
		Slci: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		Mss:  map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"},
	}
	// Marshal []any to CBOR array.
	slc := []any{
		true,
		uint(18446744073709551615),
		-1000,
		-4.1,
		[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		"The quick brown fox jumps over the lazy dog",
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"},
	}
	// Marshal struct toarray to CBOR array, such as signed/maced/encrypted CWT.
	v3 := T3{ //nolint:dupl
		T:    true,
		UI:   18446744073709551615,
		I:    -1000,
		F:    -4.1,
		B:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		S:    "The quick brown fox jumps over the lazy dog",
		Slci: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26},
		Mss:  map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"},
	}
	// Marshal nested map[string]any to CBOR map.
	m3 := map[string]any{
		"id":          "kb-2e8bc9402de2",
		"title":       "Lorem ipsum dolor sit amet elit",
		"slug":        "dolor-sit-amet",
		"locale":      "en-US",
		"state":       "published",
		"createdAt":   "2026-02-03T08:15:42Z",
		"updatedAt":   "2026-03-11T16:04:09Z",
		"publishedAt": "2026-03-11T16:05:30Z",
		"archivedAt":  nil,
		"revision":    int64(7),
		"authorId":    "usr-9121c4",
		"featured":    false,
		"tags":        []any{"lorem", "ipsum", "dolors"},
		"coverImage":  map[string]any{"src": "/media/sed-do-tempor.png", "alt": "Sed do eiusmod tempor magna", "width": int64(960), "height": int64(540)},
		"body": map[string]any{
			"type": "doc",
			"content": []any{
				map[string]any{
					"type":    "heading",
					"attrs":   map[string]any{"textAlign": "left", "level": int64(2), "id": "overview"},
					"content": []any{map[string]any{"type": "text", "text": "Overview"}},
				},
				map[string]any{
					"type":  "paragraph",
					"attrs": map[string]any{"textAlign": "left"},
					"content": []any{
						map[string]any{"type": "text", "text": "Ut enim ad minim veniam, quis nostrud "},
						map[string]any{"type": "text", "text": "laboris", "marks": []any{map[string]any{"type": "bold"}}},
						map[string]any{"type": "text", "text": ", nisi ut aliquip ex ea."},
					},
				},
				map[string]any{
					"type":    "heading",
					"attrs":   map[string]any{"textAlign": "left", "level": int64(2), "id": "steps"},
					"content": []any{map[string]any{"type": "text", "text": "Steps"}},
				},
				map[string]any{
					"type":  "orderedList",
					"attrs": map[string]any{"start": int64(1)},
					"content": []any{
						map[string]any{
							"type": "listItem",
							"content": []any{
								map[string]any{
									"type":    "paragraph",
									"attrs":   map[string]any{"textAlign": "left"},
									"content": []any{map[string]any{"type": "text", "text": "Duis aute irure dolor in veniam."}},
								},
							},
						},
						map[string]any{
							"type": "listItem",
							"content": []any{
								map[string]any{
									"type":    "paragraph",
									"attrs":   map[string]any{"textAlign": "left"},
									"content": []any{map[string]any{"type": "text", "text": "Excepteur sint, sunt in culpa id"}},
								},
							},
						},
						map[string]any{
							"type": "listItem",
							"content": []any{
								map[string]any{
									"type":    "paragraph",
									"attrs":   map[string]any{"textAlign": "left"},
									"content": []any{map[string]any{"type": "text", "text": "Cillum dolore eu fugiat n"}},
								},
							},
						},
					},
				},
				map[string]any{
					"type":  "callout",
					"attrs": map[string]any{"tone": "caution"},
					"content": []any{
						map[string]any{
							"type":    "paragraph",
							"attrs":   map[string]any{"textAlign": "left"},
							"content": []any{map[string]any{"type": "text", "text": "Nulla pariatur, excepteur sint occaecat"}},
						},
					},
				},
				map[string]any{
					"type":  "paragraph",
					"attrs": map[string]any{"textAlign": "left"},
					"content": []any{
						map[string]any{"type": "text", "text": "Anim esse? "},
						map[string]any{
							"type": "text",
							"text": "Id est laborum.",
							"marks": []any{
								map[string]any{
									"type":  "link",
									"attrs": map[string]any{"href": "https://example.com/lorem/dolore", "target": "_blank", "rel": "noopener", "class": nil},
								},
							},
						},
						map[string]any{"type": "text", "text": "."},
					},
				},
			},
		},
		"history": []any{
			map[string]any{
				"revision":  int64(7),
				"savedAt":   "2026-03-11T16:04:09Z",
				"editorId":  "usr-9121c4",
				"note":      "Commodo consequat",
				"minor":     false,
				"diffStats": map[string]any{"insertions": int64(18), "deletions": int64(9)},
			},
			map[string]any{
				"revision":  int64(6),
				"savedAt":   "2026-02-27T10:22:51Z",
				"editorId":  "usr-5c91e9",
				"note":      nil,
				"minor":     true,
				"diffStats": map[string]any{"insertions": int64(2), "deletions": int64(2)},
			},
			map[string]any{
				"revision":  int64(5),
				"savedAt":   "2026-02-19T14:48:16Z",
				"editorId":  "usr-9121c4",
				"note":      "Velit esse sed",
				"minor":     false,
				"diffStats": map[string]any{"insertions": int64(41), "deletions": int64(3)},
			},
			map[string]any{
				"revision":  int64(4),
				"savedAt":   "2026-02-04T13:05:27Z",
				"editorId":  "usr-5c91e9",
				"note":      "Quis aut",
				"minor":     true,
				"diffStats": map[string]any{"insertions": int64(1), "deletions": int64(1)},
			},
		},
		"review": map[string]any{
			"decision":    "approved",
			"completedAt": "2026-03-11T16:04:58Z",
			"reviewerId":  "usr-4d0f12",
			"checklist":   map[string]any{"linksChecked": true, "screenshotsCurrent": true, "styleGuide": true},
		},
		"settings": map[string]any{
			"visibility":    "public",
			"layout":        "wide",
			"comments":      map[string]any{"enabled": true, "moderation": "afterPost", "sortOrder": "newest"},
			"toc":           map[string]any{"enabled": true, "maxDepth": int64(3)},
			"feedback":      map[string]any{"enabled": true, "prompt": "Amet consectetur?"},
			"sharing":       map[string]any{"enabled": true, "allowPrint": true, "allowCopy": false},
			"notifications": map[string]any{"onComment": true, "onFeedback": false},
			"permissions":   map[string]any{"edit": "team", "comment": "everyone"},
			"relatedLimit":  int64(4),
			"customCss":     nil,
			"layoutOptions": map[string]any{},
			"overrides":     map[string]any{},
		},
		"seo": map[string]any{
			"metaTitle":       "Lorem ipsum dolor sit am",
			"metaDescription": "Lorem ipsum dolor sit amet, adipiscing.",
			"canonicalUrl":    "https://example.com/help/dolor-sit-amet",
			"noIndex":         false,
			"keywords":        []any{"dolor sit amet", "ipsum", "sed do tempor"},
			"openGraph":       map[string]any{"type": "article", "imageUrl": "https://example.com/media/sed-do-tempor.png"},
			"alternates":      map[string]any{"de": "/de/help/dolor-sit-amet", "fr": "/fr/help/dolor-sit-amet", "es": "/es/help/dolor-sit-amet"},
			"structuredData":  map[string]any{},
		},
		"translations": map[string]any{
			"de": map[string]any{"state": "current", "revision": int64(7), "translatedAt": "2026-03-12T09:10:33Z", "translatorId": "usr-3378ff"},
			"fr": map[string]any{"state": "outdated", "revision": int64(5), "translatedAt": "2026-02-20T11:02:18Z", "translatorId": "usr-3378ff"},
			"es": map[string]any{"state": "outdated", "revision": int64(6), "translatedAt": "2026-02-28T08:44:05Z", "translatorId": "usr-3378ff"},
			"ja": map[string]any{"state": "requested", "revision": int64(7), "translatedAt": nil, "translatorId": nil},
		},
		"renditions": map[string]any{
			"web": map[string]any{
				"format":     "html",
				"url":        "/help/dolor-sit-amet",
				"renderedAt": "2026-03-11T16:05:12Z",
				"sizeBytes":  int64(23817),
				"checksum":   "384cb5f21297367c",
				"cacheHit":   false,
				"attempts":   int64(1),
				"error":      nil,
				"timing":     map[string]any{"queuedMs": int64(12), "renderMs": int64(184), "uploadMs": int64(37)},
				"input":      map[string]any{"revision": int64(7), "theme": "standard"},
			},
			"print": map[string]any{
				"format":     "pdf",
				"url":        "/help/pdf/dolor-sit-amet.pdf",
				"renderedAt": "2026-03-11T16:05:19Z",
				"sizeBytes":  int64(88214),
				"checksum":   "6d573c5617454cf4",
				"cacheHit":   false,
				"pageCount":  int64(2),
				"attempts":   int64(2),
				"error":      nil,
				"timing":     map[string]any{"queuedMs": int64(15), "renderMs": int64(1342), "uploadMs": int64(96)},
				"input":      map[string]any{"revision": int64(7), "theme": "printable"},
			},
		},
		"metrics": map[string]any{
			"views":          int64(4736),
			"uniqueVisitors": int64(3377),
			"helpfulVotes":   int64(212),
			"unhelpfulVotes": int64(19),
			"byChannel":      map[string]any{"web": int64(4529), "app": int64(207)},
			"bySource":       map[string]any{"search": int64(2894), "related": int64(1087), "direct": int64(755)},
			"byLocale":       map[string]any{"en": int64(3641), "de": int64(598), "fr": int64(301), "es": int64(196)},
			"byDevice":       map[string]any{"desktop": int64(3052), "mobile": int64(1684)},
		},
		"customFields": map[string]any{},
	}
	// Marshal nested map[any]any to CBOR map, such as m3 decoded to any with default options.
	var m4 any
	if data, err := Marshal(m3); err != nil {
		b.Fatal("Marshal:", err)
	} else if err = Unmarshal(data, &m4); err != nil {
		b.Fatal("Unmarshal:", err)
	}
	// Marshal pointers, so values are addressable like struct fields, slice elements, and map values.
	t1 := time.Unix(1363896240, 0).UTC()
	bi := new(big.Int).Lsh(big.NewInt(1), 64)
	tag := Tag{Number: 100, Content: uint64(1363896240)}
	vm := valueMarshaler{1, 2, 3, 4}
	pm := ptrMarshaler{1, 2, 3, 4}
	var omitZeroTime struct {
		T time.Time `cbor:",omitzero"`
	}
	var moreBenchmarks = []struct {
		name  string
		value any
	}{
		{
			name:  "map[string]any/map",
			value: m1,
		},
		{
			name:  "nested map[string]any/map",
			value: m3,
		},
		{
			name:  "nested map[any]any/map",
			value: m4,
		},
		{
			name:  "struct/map",
			value: v1,
		},
		{
			name:  "struct many flds all omitempty all empty/map",
			value: ManyFieldsAllOmitEmpty{},
		},
		{
			name:  "struct some flds all omitempty all empty/map",
			value: SomeFieldsAllOmitEmpty{},
		},
		{
			name: "struct many flds all omitempty all nonempty/map",
			value: ManyFieldsAllOmitEmpty{
				F01: 1, F02: 1, F03: 1, F04: 1, F05: 1, F06: 1, F07: 1, F08: 1, F09: 1, F10: 1, F11: 1, F12: 1, F13: 1, F14: 1, F15: 1, F16: 1,
				F17: 1, F18: 1, F19: 1, F20: 1, F21: 1, F22: 1, F23: 1, F24: 1, F25: 1, F26: 1, F27: 1, F28: 1, F29: 1, F30: 1, F31: 1, F32: 1,
			},
		},
		{
			name: "struct some flds all omitempty all nonempty/map",
			value: SomeFieldsAllOmitEmpty{
				F01: 1, F02: 1, F03: 1, F04: 1, F05: 1, F06: 1, F07: 1, F08: 1,
			},
		},
		{
			name:  "struct many flds one omitempty/map",
			value: ManyFieldsOneOmitEmpty{},
		},
		{
			name:  "struct some flds one omitempty/map",
			value: SomeFieldsOneOmitEmpty{},
		},
		{
			name:  "map[int]any/map",
			value: m2,
		},
		{
			name:  "struct keyasint/map",
			value: v2,
		},
		{
			name:  "[]any/array",
			value: slc,
		},
		{
			name:  "struct toarray/array",
			value: v3,
		},
		{
			name:  "*time.Time/uint",
			value: &t1,
		},
		{
			name:  "*big.Int/bignum",
			value: bi,
		},
		{
			name:  "*Tag/tag",
			value: &tag,
		},
		{
			name:  "*value receiver marshaler/bstr",
			value: &vm,
		},
		{
			name:  "*pointer receiver marshaler/bstr",
			value: &pm,
		},
		{
			name:  "*struct omitzero time.Time empty/map",
			value: &omitZeroTime,
		},
	}
	for _, bm := range moreBenchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				if _, err := Marshal(bm.value); err != nil {
					b.Fatal("Marshal:", err)
				}
			}
		})
	}

	// Marshal with non-default encoding options.
	sortedEM, err := EncOptions{Sort: SortBytewiseLexical}.EncMode()
	if err != nil {
		b.Fatal("EncMode:", err)
	}
	timeUnixMicroEM, err := EncOptions{Time: TimeUnixMicro}.EncMode()
	if err != nil {
		b.Fatal("EncMode:", err)
	}
	for _, bm := range []struct {
		name  string
		em    EncMode
		value any
	}{
		{
			name:  "nested map[string]any/map/bytewise lexical",
			em:    sortedEM,
			value: m3,
		},
		{
			name:  "nested map[any]any/map/bytewise lexical",
			em:    sortedEM,
			value: m4,
		},
		{
			name:  "time.Time/float/unix micro",
			em:    timeUnixMicroEM,
			value: t1,
		},
	} {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				if _, err := bm.em.Marshal(bm.value); err != nil {
					b.Fatal("Marshal:", err)
				}
			}
		})
	}
}

func BenchmarkMarshalCanonical(b *testing.B) {
	type strc struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
		D string `cbor:"d"`
		E string `cbor:"e"`
		F string `cbor:"f"`
		G string `cbor:"g"`
		H string `cbor:"h"`
		I string `cbor:"i"`
		J string `cbor:"j"`
		L string `cbor:"l"`
		M string `cbor:"m"`
		N string `cbor:"n"`
	}
	for _, bm := range []struct {
		name   string
		data   []byte
		values []any
	}{
		{
			name: "map",
			data: mustHexDecode("ad616161416162614261636143616461446165614561666146616761476168614861696149616a614a616c614c616d614d616e614e"),
			values: []any{
				map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"},
				strc{A: "A", B: "B", C: "C", D: "D", E: "E", F: "F", G: "G", H: "H", I: "I", J: "J", L: "L", M: "M", N: "N"},
				map[int]int{0: 0}, /* single-entry map */
			},
		},
	} {
		for _, v := range bm.values {
			// Name uses the format: "MarshalCanonical/Go type/CBOR type".
			name := goTypeName(reflect.TypeOf(v)) + "/" + bm.name
			b.Run(name, func(b *testing.B) {
				for b.Loop() {
					if _, err := Marshal(v); err != nil {
						b.Fatal("Marshal:", err)
					}
				}
			})
			// Canonical encoding
			// Name uses the format: "MarshalCanonical/Go type/CBOR type/canonical".
			name = goTypeName(reflect.TypeOf(v)) + "/" + bm.name + "/canonical"
			em, _ := EncOptions{Sort: SortCanonical}.EncMode()
			b.Run(name, func(b *testing.B) {
				for b.Loop() {
					if _, err := em.Marshal(v); err != nil {
						b.Fatal("Marshal:", err)
					}
				}
			})
		}
	}
}

// BenchmarkNewEncoderEncode benchmarks NewEncoder() and Encode().
func BenchmarkNewEncoderEncode(b *testing.B) {
	for _, bm := range encodeBenchmarks {
		for _, v := range bm.values {
			// Name uses the format: "NewEncoderEncode/Go type/CBOR type",
			// e.g. "NewEncoderEncode/[]byte/bstr"
			name := goTypeName(reflect.TypeOf(v)) + "/" + bm.name
			b.Run(name, func(b *testing.B) {
				for b.Loop() {
					encoder := NewEncoder(io.Discard)
					if err := encoder.Encode(v); err != nil {
						b.Fatal("Encode:", err)
					}
				}
			})
		}
	}
}

// BenchmarkEncode benchmarks Encode(). It reuses same Encoder to exclude NewEncoder()
// from the benchmark.
func BenchmarkEncode(b *testing.B) {
	for _, bm := range encodeBenchmarks {
		for _, v := range bm.values {
			// Name uses the format: "Encode/Go type/CBOR type",
			// e.g. "Encode/[]byte/bstr"
			name := goTypeName(reflect.TypeOf(v)) + "/" + bm.name
			b.Run(name, func(b *testing.B) {
				encoder := NewEncoder(io.Discard)
				for b.Loop() {
					if err := encoder.Encode(v); err != nil {
						b.Fatal("Encode:", err)
					}
				}
			})
		}
	}
}

func BenchmarkEncodeStream(b *testing.B) {
	for b.Loop() {
		encoder := NewEncoder(io.Discard)
		for range rounds {
			for _, bm := range encodeBenchmarks {
				for _, v := range bm.values {
					if err := encoder.Encode(v); err != nil {
						b.Fatal("Encode:", err)
					}
				}
			}
		}
	}
}

func BenchmarkUnmarshalCOSE(b *testing.B) {
	// Data from https://tools.ietf.org/html/rfc8392#appendix-A section A.2
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "128-Bit Symmetric Key",
			data: mustHexDecode("a42050231f4c4d4d3051fdc2ec0a3851d5b3830104024c53796d6d6574726963313238030a"),
		},
		{
			name: "256-Bit Symmetric Key",
			data: mustHexDecode("a4205820403697de87af64611c1d32a05dab0fe1fcb715a86ab435f1ec99192d795693880104024c53796d6d6574726963323536030a"),
		},
		{
			name: "ECDSA P256 256-Bit Key",
			data: mustHexDecode("a72358206c1382765aec5358f117733d281c1c7bdc39884d04a45a1e6c67c858bc206c1922582060f7f1a780d8a783bfb7a2dd6b2796e8128dbbcef9d3d168db9529971a36e7b9215820143329cce7868e416927599cf65a34f3ce2ffda55a7eca69ed8919a394d42f0f2001010202524173796d6d657472696345434453413235360326"),
		},
	}
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				var v coseKey
				if err := Unmarshal(tc.data, &v); err != nil {
					b.Fatal("Unmarshal:", err)
				}
			}
		})
	}
}

func BenchmarkMarshalCOSE(b *testing.B) {
	// Data from https://tools.ietf.org/html/rfc8392#appendix-A section A.2
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "128-Bit Symmetric Key",
			data: mustHexDecode("a42050231f4c4d4d3051fdc2ec0a3851d5b3830104024c53796d6d6574726963313238030a"),
		},
		{
			name: "256-Bit Symmetric Key",
			data: mustHexDecode("a4205820403697de87af64611c1d32a05dab0fe1fcb715a86ab435f1ec99192d795693880104024c53796d6d6574726963323536030a"),
		},
		{
			name: "ECDSA P256 256-Bit Key",
			data: mustHexDecode("a72358206c1382765aec5358f117733d281c1c7bdc39884d04a45a1e6c67c858bc206c1922582060f7f1a780d8a783bfb7a2dd6b2796e8128dbbcef9d3d168db9529971a36e7b9215820143329cce7868e416927599cf65a34f3ce2ffda55a7eca69ed8919a394d42f0f2001010202524173796d6d657472696345434453413235360326"),
		},
	}
	for _, tc := range testCases {
		var v coseKey
		if err := Unmarshal(tc.data, &v); err != nil {
			b.Fatal("Unmarshal:", err)
		}
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				if _, err := Marshal(v); err != nil {
					b.Fatal("Marshal:", err)
				}
			}
		})
	}
}

func BenchmarkUnmarshalCWTClaims(b *testing.B) {
	// Data from https://tools.ietf.org/html/rfc8392#appendix-A section A.1
	data := mustHexDecode("a70175636f61703a2f2f61732e6578616d706c652e636f6d02656572696b77037818636f61703a2f2f6c696768742e6578616d706c652e636f6d041a5612aeb0051a5610d9f0061a5610d9f007420b71")
	for b.Loop() {
		var v claims
		if err := Unmarshal(data, &v); err != nil {
			b.Fatal("Unmarshal:", err)
		}
	}
}

func BenchmarkUnmarshalCWTClaimsWithDupMapKeyOpt(b *testing.B) {
	// Data from https://tools.ietf.org/html/rfc8392#appendix-A section A.1
	data := mustHexDecode("a70175636f61703a2f2f61732e6578616d706c652e636f6d02656572696b77037818636f61703a2f2f6c696768742e6578616d706c652e636f6d041a5612aeb0051a5610d9f0061a5610d9f007420b71")
	dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
	for b.Loop() {
		var v claims
		if err := dm.Unmarshal(data, &v); err != nil {
			b.Fatal("Unmarshal:", err)
		}
	}
}

func BenchmarkMarshalCWTClaims(b *testing.B) {
	// Data from https://tools.ietf.org/html/rfc8392#appendix-A section A.1
	data := mustHexDecode("a70175636f61703a2f2f61732e6578616d706c652e636f6d02656572696b77037818636f61703a2f2f6c696768742e6578616d706c652e636f6d041a5612aeb0051a5610d9f0061a5610d9f007420b71")
	var v claims
	if err := Unmarshal(data, &v); err != nil {
		b.Fatal("Unmarshal:", err)
	}
	for b.Loop() {
		if _, err := Marshal(v); err != nil {
			b.Fatal("Unmarshal:", err)
		}
	}
}

func BenchmarkUnmarshalSenML(b *testing.B) {
	// Data from https://tools.ietf.org/html/rfc8428#section-6
	data := mustHexDecode("87a721781b75726e3a6465763a6f773a3130653230373361303130383030363a22fb41d303a15b00106223614120050067766f6c7461676501615602fb405e066666666666a3006763757272656e74062402fb3ff3333333333333a3006763757272656e74062302fb3ff4cccccccccccda3006763757272656e74062202fb3ff6666666666666a3006763757272656e74062102f93e00a3006763757272656e74062002fb3ff999999999999aa3006763757272656e74060002fb3ffb333333333333")
	for b.Loop() {
		var v []SenMLRecord
		if err := Unmarshal(data, &v); err != nil {
			b.Fatal("Unmarshal:", err)
		}
	}
}

func BenchmarkMarshalSenML(b *testing.B) {
	// Data from https://tools.ietf.org/html/rfc8428#section-6
	data := mustHexDecode("87a721781b75726e3a6465763a6f773a3130653230373361303130383030363a22fb41d303a15b00106223614120050067766f6c7461676501615602fb405e066666666666a3006763757272656e74062402fb3ff3333333333333a3006763757272656e74062302fb3ff4cccccccccccda3006763757272656e74062202fb3ff6666666666666a3006763757272656e74062102f93e00a3006763757272656e74062002fb3ff999999999999aa3006763757272656e74060002fb3ffb333333333333")
	var v []SenMLRecord
	if err := Unmarshal(data, &v); err != nil {
		b.Fatal("Unmarshal:", err)
	}
	for b.Loop() {
		if _, err := Marshal(v); err != nil {
			b.Fatal("Unmarshal:", err)
		}
	}
}

func BenchmarkMarshalSenMLShortestFloat16(b *testing.B) {
	// Data from https://tools.ietf.org/html/rfc8428#section-6
	data := mustHexDecode("87a721781b75726e3a6465763a6f773a3130653230373361303130383030363a22fb41d303a15b00106223614120050067766f6c7461676501615602fb405e066666666666a3006763757272656e74062402fb3ff3333333333333a3006763757272656e74062302fb3ff4cccccccccccda3006763757272656e74062202fb3ff6666666666666a3006763757272656e74062102f93e00a3006763757272656e74062002fb3ff999999999999aa3006763757272656e74060002fb3ffb333333333333")
	var v []SenMLRecord
	if err := Unmarshal(data, &v); err != nil {
		b.Fatal("Unmarshal:", err)
	}
	em, _ := EncOptions{ShortestFloat: ShortestFloat16}.EncMode()
	for b.Loop() {
		if _, err := em.Marshal(v); err != nil {
			b.Fatal("Unmarshal:", err)
		}
	}
}

func BenchmarkUnmarshalWebAuthn(b *testing.B) {
	// Data generated from Yubico security key
	data := mustHexDecode("a363666d74686669646f2d7532666761747453746d74a26373696758483046022100e7ab373cfbd99fcd55fd59b0f6f17fef5b77a20ddec3db7f7e4d55174e366236022100828336b4822125fb56541fb14a8a273876acd339395ec2dad95cf41c1dd2a9ae637835638159024e3082024a30820132a0030201020204124a72fe300d06092a864886f70d01010b0500302e312c302a0603550403132359756269636f2055324620526f6f742043412053657269616c203435373230303633313020170d3134303830313030303030305a180f32303530303930343030303030305a302c312a302806035504030c2159756269636f205532462045452053657269616c203234393431343937323135383059301306072a8648ce3d020106082a8648ce3d030107034200043d8b1bbd2fcbf6086e107471601468484153c1c6d3b4b68a5e855e6e40757ee22bcd8988bf3befd7cdf21cb0bf5d7a150d844afe98103c6c6607d9faae287c02a33b3039302206092b0601040182c40a020415312e332e362e312e342e312e34313438322e312e313013060b2b0601040182e51c020101040403020520300d06092a864886f70d01010b05000382010100a14f1eea0076f6b8476a10a2be72e60d0271bb465b2dfbfc7c1bd12d351989917032631d795d097fa30a26a325634e85721bc2d01a86303f6bc075e5997319e122148b0496eec8d1f4f94cf4110de626c289443d1f0f5bbb239ca13e81d1d5aa9df5af8e36126475bfc23af06283157252762ff68879bcf0ef578d55d67f951b4f32b63c8aea5b0f99c67d7d814a7ff5a6f52df83e894a3a5d9c8b82e7f8bc8daf4c80175ff8972fda79333ec465d806eacc948f1bab22045a95558a48c20226dac003d41fbc9e05ea28a6bb5e10a49de060a0a4f6a2676a34d68c4abe8c61874355b9027e828ca9e064b002d62e8d8cf0744921753d35e3c87c5d5779453e7768617574684461746158c449960de5880e8c687434170f6476605b8fe4aeb9a28632c7995cf3ba831d976341000000000000000000000000000000000000000000408903fd7dfd2c9770e98cae0123b13a2c27828a106349bc6277140e7290b7e9eb7976aa3c04ed347027caf7da3a2fa76304751c02208acfc4e7fc6c7ebbc375c8a5010203262001215820ad7f7992c335b90d882b2802061b97a4fabca7e2ee3e7a51e728b8055e4eb9c7225820e0966ba7005987fece6f0e0e13447aa98cec248e4000a594b01b74c1cb1d40b3")
	for b.Loop() {
		var v attestationObject
		if err := Unmarshal(data, &v); err != nil {
			b.Fatal("Unmarshal:", err)
		}
	}
}

func BenchmarkMarshalWebAuthn(b *testing.B) {
	// Data generated from Yubico security key
	data := mustHexDecode("a363666d74686669646f2d7532666761747453746d74a26373696758483046022100e7ab373cfbd99fcd55fd59b0f6f17fef5b77a20ddec3db7f7e4d55174e366236022100828336b4822125fb56541fb14a8a273876acd339395ec2dad95cf41c1dd2a9ae637835638159024e3082024a30820132a0030201020204124a72fe300d06092a864886f70d01010b0500302e312c302a0603550403132359756269636f2055324620526f6f742043412053657269616c203435373230303633313020170d3134303830313030303030305a180f32303530303930343030303030305a302c312a302806035504030c2159756269636f205532462045452053657269616c203234393431343937323135383059301306072a8648ce3d020106082a8648ce3d030107034200043d8b1bbd2fcbf6086e107471601468484153c1c6d3b4b68a5e855e6e40757ee22bcd8988bf3befd7cdf21cb0bf5d7a150d844afe98103c6c6607d9faae287c02a33b3039302206092b0601040182c40a020415312e332e362e312e342e312e34313438322e312e313013060b2b0601040182e51c020101040403020520300d06092a864886f70d01010b05000382010100a14f1eea0076f6b8476a10a2be72e60d0271bb465b2dfbfc7c1bd12d351989917032631d795d097fa30a26a325634e85721bc2d01a86303f6bc075e5997319e122148b0496eec8d1f4f94cf4110de626c289443d1f0f5bbb239ca13e81d1d5aa9df5af8e36126475bfc23af06283157252762ff68879bcf0ef578d55d67f951b4f32b63c8aea5b0f99c67d7d814a7ff5a6f52df83e894a3a5d9c8b82e7f8bc8daf4c80175ff8972fda79333ec465d806eacc948f1bab22045a95558a48c20226dac003d41fbc9e05ea28a6bb5e10a49de060a0a4f6a2676a34d68c4abe8c61874355b9027e828ca9e064b002d62e8d8cf0744921753d35e3c87c5d5779453e7768617574684461746158c449960de5880e8c687434170f6476605b8fe4aeb9a28632c7995cf3ba831d976341000000000000000000000000000000000000000000408903fd7dfd2c9770e98cae0123b13a2c27828a106349bc6277140e7290b7e9eb7976aa3c04ed347027caf7da3a2fa76304751c02208acfc4e7fc6c7ebbc375c8a5010203262001215820ad7f7992c335b90d882b2802061b97a4fabca7e2ee3e7a51e728b8055e4eb9c7225820e0966ba7005987fece6f0e0e13447aa98cec248e4000a594b01b74c1cb1d40b3")
	var v attestationObject
	if err := Unmarshal(data, &v); err != nil {
		b.Fatal("Unmarshal:", err)
	}
	for b.Loop() {
		if _, err := Marshal(v); err != nil {
			b.Fatal("Marshal:", err)
		}
	}
}

func BenchmarkUnmarshalCOSEMAC(b *testing.B) {
	// Data from https://tools.ietf.org/html/rfc8392#appendix-A section A.4
	data := mustHexDecode("d83dd18443a10104a1044c53796d6d65747269633235365850a70175636f61703a2f2f61732e6578616d706c652e636f6d02656572696b77037818636f61703a2f2f6c696768742e6578616d706c652e636f6d041a5612aeb0051a5610d9f0061a5610d9f007420b7148093101ef6d789200")

	for b.Loop() {
		var v macedCOSE
		if err := Unmarshal(data, &v); err != nil {
			b.Fatal("Unmarshal:", err)
		}
	}
}

func BenchmarkUnmarshalCOSEMACWithTag(b *testing.B) {
	// Data from https://tools.ietf.org/html/rfc8392#appendix-A section A.4
	data := mustHexDecode("d83dd18443a10104a1044c53796d6d65747269633235365850a70175636f61703a2f2f61732e6578616d706c652e636f6d02656572696b77037818636f61703a2f2f6c696768742e6578616d706c652e636f6d041a5612aeb0051a5610d9f0061a5610d9f007420b7148093101ef6d789200")

	// Register tag CBOR Web Token (CWT) 61 and COSE_Mac0 17 with macedCOSE type
	tags := NewTagSet()
	if err := tags.Add(TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, reflect.TypeFor[macedCOSE](), 61, 17); err != nil {
		b.Fatal("TagSet.Add:", err)
	}

	// Create DecMode with tags
	dm, _ := DecOptions{}.DecModeWithTags(tags)

	for b.Loop() {
		var v macedCOSE
		if err := dm.Unmarshal(data, &v); err != nil {
			b.Fatal("Unmarshal:", err)
		}
	}
}
func BenchmarkMarshalCOSEMAC(b *testing.B) {
	// Data from https://tools.ietf.org/html/rfc8392#appendix-A section A.4
	data := mustHexDecode("d83dd18443a10104a1044c53796d6d65747269633235365850a70175636f61703a2f2f61732e6578616d706c652e636f6d02656572696b77037818636f61703a2f2f6c696768742e6578616d706c652e636f6d041a5612aeb0051a5610d9f0061a5610d9f007420b7148093101ef6d789200")

	var v macedCOSE
	if err := Unmarshal(data, &v); err != nil {
		b.Fatal("Unmarshal():", err)
	}

	for b.Loop() {
		if _, err := Marshal(v); err != nil {
			b.Fatal("Marshal():", v, err)
		}
	}
}

func BenchmarkMarshalCOSEMACWithTag(b *testing.B) {
	// Data from https://tools.ietf.org/html/rfc8392#appendix-A section A.4
	data := mustHexDecode("d83dd18443a10104a1044c53796d6d65747269633235365850a70175636f61703a2f2f61732e6578616d706c652e636f6d02656572696b77037818636f61703a2f2f6c696768742e6578616d706c652e636f6d041a5612aeb0051a5610d9f0061a5610d9f007420b7148093101ef6d789200")

	// Register tag CBOR Web Token (CWT) 61 and COSE_Mac0 17 with macedCOSE type
	tags := NewTagSet()
	if err := tags.Add(TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, reflect.TypeFor[macedCOSE](), 61, 17); err != nil {
		b.Fatal("TagSet.Add:", err)
	}

	// Create EncMode with tags.
	dm, _ := DecOptions{}.DecModeWithTags(tags)
	em, _ := EncOptions{}.EncModeWithTags(tags)

	var v macedCOSE
	if err := dm.Unmarshal(data, &v); err != nil {
		b.Fatal("Unmarshal():", err)
	}

	for b.Loop() {
		if _, err := em.Marshal(v); err != nil {
			b.Fatal("Marshal():", v, err)
		}
	}
}

func BenchmarkUnmarshalMapToStruct(b *testing.B) {
	type S struct {
		A, B, C, D, E, F, G, H, I, J, K, L, M bool
	}

	var (
		allKnownFields            = mustHexDecode("ad6141f56142f56143f56144f56145f56146f56147f56148f56149f5614af5614bf5614cf5614df5") // {"A": true, ... "M": true }
		allKnownDuplicateFields   = mustHexDecode("ad6141f56141f56141f56141f56141f56141f56141f56141f56141f56141f56141f56141f56141f5") // {"A": true, "A": true, "A": true, ...}
		allUnknownFields          = mustHexDecode("ad614ef5614ff56150f56151f56152f56153f56154f56155f56156f56157f56158f56159f5615af5") // {"N": true, ... "Z": true }
		allUnknownDuplicateFields = mustHexDecode("ad614ef5614ef5614ef5614ef5614ef5614ef5614ef5614ef5614ef5614ef5614ef5614ef5614ef5") // {"N": true, "N": true, "N": true, ...}
	)

	type ManyFields struct {
		AA, AB, AC, AD, AE, AF, AG, AH, AI, AJ, AK, AL, AM, AN, AO, AP, AQ, AR, AS, AT, AU, AV, AW, AX, AY, AZ bool
		BA, BB, BC, BD, BE, BF, BG, BH, BI, BJ, BK, BL, BM, BN, BO, BP, BQ, BR, BS, BT, BU, BV, BW, BX, BY, BZ bool
		CA, CB, CC, CD, CE, CF, CG, CH, CI, CJ, CK, CL, CM, CN, CO, CP, CQ, CR, CS, CT, CU, CV, CW, CX, CY, CZ bool
		DA, DB, DC, DD, DE, DF, DG, DH, DI, DJ, DK, DL, DM, DN, DO, DP, DQ, DR, DS, DT, DU, DV, DW, DX, DY, DZ bool
	}
	var manyFieldsOneKeyPerField []byte
	{
		// An EncOption that accepts a function to sort or shuffle keys might be useful for
		// cases like this. Here we are manually encoding the fields in reverse order to
		// target worst-case key-to-field matching.
		rt := reflect.TypeFor[ManyFields]()
		var buf bytes.Buffer
		if rt.NumField() > 255 {
			b.Fatalf("invalid test assumption: ManyFields expected to have no more than 255 fields, has %d", rt.NumField())
		}
		buf.WriteByte(0xb8)
		buf.WriteByte(byte(rt.NumField()))        //nolint:gosec
		for i := rt.NumField() - 1; i >= 0; i-- { // backwards
			f := rt.Field(i)
			if len(f.Name) > 23 {
				b.Fatalf("invalid test assumption: field name %q longer than 23 bytes", f.Name)
			}
			buf.WriteByte(byte(0x60 + len(f.Name))) //nolint:gosec
			buf.WriteString(f.Name)
			buf.WriteByte(0xf5) // true
		}
		manyFieldsOneKeyPerField = buf.Bytes()
	}

	type input struct {
		name   string
		data   []byte
		into   any
		reject bool
	}

	for _, tc := range []*struct {
		name   string
		opts   DecOptions
		inputs []input
	}{
		{
			name: "default",
			opts: DecOptions{},
			inputs: []input{
				{
					name:   "all known flds",
					data:   allKnownFields,
					into:   S{},
					reject: false,
				},
				{
					name:   "all known dup flds",
					data:   allKnownDuplicateFields,
					into:   S{},
					reject: false,
				},
				{
					name:   "all unk flds",
					data:   allUnknownFields,
					into:   S{},
					reject: false,
				},
				{
					name:   "all unk dup flds",
					data:   allUnknownDuplicateFields,
					into:   S{},
					reject: false,
				},
				{
					name:   "many flds one key per fld",
					data:   manyFieldsOneKeyPerField,
					into:   ManyFields{},
					reject: false,
				},
			},
		},
		{
			name: "rej unk",
			opts: DecOptions{ExtraReturnErrors: ExtraDecErrorUnknownField},
			inputs: []input{
				{
					name:   "all known flds",
					data:   allKnownFields,
					into:   S{},
					reject: false,
				},
				{
					name:   "all known dup flds",
					data:   allKnownDuplicateFields,
					into:   S{},
					reject: false,
				},
				{
					name:   "all unk flds",
					data:   allUnknownFields,
					into:   S{},
					reject: true,
				},
				{
					name:   "all unk dup flds",
					data:   allUnknownDuplicateFields,
					into:   S{},
					reject: true,
				},
			},
		},
		{
			name: "rej dup",
			opts: DecOptions{DupMapKey: DupMapKeyEnforcedAPF},
			inputs: []input{
				{
					name:   "all known flds",
					data:   allKnownFields,
					into:   S{},
					reject: false,
				},
				{
					name:   "all known dup flds",
					data:   allKnownDuplicateFields,
					into:   S{},
					reject: true,
				},
				{
					name:   "all unk flds",
					data:   allUnknownFields,
					into:   S{},
					reject: false,
				},
				{
					name:   "all unk dup flds",
					data:   allUnknownDuplicateFields,
					into:   S{},
					reject: true,
				},
			},
		},
		{
			name: "rej unk and dup",
			opts: DecOptions{
				DupMapKey:         DupMapKeyEnforcedAPF,
				ExtraReturnErrors: ExtraDecErrorUnknownField,
			},
			inputs: []input{
				{
					name:   "all known flds",
					data:   allKnownFields,
					into:   S{},
					reject: false,
				},
				{
					name:   "all known dup flds",
					data:   allKnownDuplicateFields,
					into:   S{},
					reject: true,
				},
				{
					name:   "all unk flds",
					data:   allUnknownFields,
					into:   S{},
					reject: true,
				},
				{
					name:   "all unk dup flds",
					data:   allUnknownDuplicateFields,
					into:   S{},
					reject: true,
				},
			},
		},
	} {
		for _, in := range tc.inputs {
			b.Run(fmt.Sprintf("%s/%s", tc.name, in.name), func(b *testing.B) {
				dm, err := tc.opts.DecMode()
				if err != nil {
					b.Fatal(err)
				}

				dst := reflect.New(reflect.TypeOf(in.into)).Interface()

				for b.Loop() {
					if err := dm.Unmarshal(in.data, dst); !in.reject && err != nil {
						b.Fatalf("unexpected error: %v", err)
					} else if in.reject && err == nil {
						b.Fatal("expected non-nil error")
					}
				}
			})
		}
	}
}

func goTypeName(t reflect.Type) string {
	if t.Kind() == reflect.Struct {
		return t.Kind().String()
	}
	s := t.String()
	s = strings.ReplaceAll(s, "interface {}", "any")
	s = strings.ReplaceAll(s, "uint8", "byte")
	return s
}
