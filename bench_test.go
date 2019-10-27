// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

package cbor_test

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/fxamacker/cbor"
)

const rounds = 100

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

var decodeBenchmarks = []struct {
	name          string
	cborData      []byte
	decodeToTypes []reflect.Type
}{
	{"bool", hexDecode("f5"), []reflect.Type{typeIntf, typeBool}},                                                                                                                                                                                           // true
	{"positive int", hexDecode("1bffffffffffffffff"), []reflect.Type{typeIntf, typeUint64}},                                                                                                                                                                 // uint64(18446744073709551615)
	{"negative int", hexDecode("3903e7"), []reflect.Type{typeIntf, typeInt64}},                                                                                                                                                                              // int64(-1000)
	{"float", hexDecode("fbc010666666666666"), []reflect.Type{typeIntf, typeFloat64}},                                                                                                                                                                       // float64(-4.1)
	{"bytes", hexDecode("581a0102030405060708090a0b0c0d0e0f101112131415161718191a"), []reflect.Type{typeIntf, typeByteSlice}},                                                                                                                               // []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}
	{"bytes indef len", hexDecode("5f410141024103410441054106410741084109410a410b410c410d410e410f4110411141124113411441154116411741184119411aff"), []reflect.Type{typeIntf, typeByteSlice}},                                                                 // []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}
	{"text", hexDecode("782b54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f67"), []reflect.Type{typeIntf, typeString}},                                                                                                 // "The quick brown fox jumps over the lazy dog"
	{"text indef len", hexDecode("7f61546168616561206171617561696163616b612061626172616f6177616e61206166616f61786120616a6175616d617061736120616f61766165617261206174616861656120616c6161617a617961206164616f6167ff"), []reflect.Type{typeIntf, typeString}}, // "The quick brown fox jumps over the lazy dog"
	{"array", hexDecode("981a0102030405060708090a0b0c0d0e0f101112131415161718181819181a"), []reflect.Type{typeIntf, typeIntSlice}},                                                                                                                          // []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}
	{"array indef len", hexDecode("9f0102030405060708090a0b0c0d0e0f101112131415161718181819181aff"), []reflect.Type{typeIntf, typeIntSlice}},                                                                                                                // []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}
	{"map", hexDecode("ad616161416162614261636143616461446165614561666146616761476168614861696149616a614a616c614c616d614d616e614e"), []reflect.Type{typeIntf, typeMapStringIntf, typeMapStringString, reflect.TypeOf(strc{})}},                              // map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"}}
	{"map indef len", hexDecode("bf616161416162614261636143616461446165614561666146616761476168614861696149616a614a616b614b616c614c616d614d616e614eff"), []reflect.Type{typeIntf, typeMapStringIntf, typeMapStringString, reflect.TypeOf(strc{})}},          // map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"}}
}

var encodeBenchmarks = []struct {
	name     string
	cborData []byte
	values   []interface{}
}{
	{"bool", hexDecode("f5"), []interface{}{true}},
	{"positive int", hexDecode("1bffffffffffffffff"), []interface{}{uint64(18446744073709551615)}},
	{"negative int", hexDecode("3903e7"), []interface{}{int64(-1000)}},
	{"float", hexDecode("fbc010666666666666"), []interface{}{float64(-4.1)}},
	{"bytes", hexDecode("581a0102030405060708090a0b0c0d0e0f101112131415161718191a"), []interface{}{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}}},
	{"text", hexDecode("782b54686520717569636b2062726f776e20666f78206a756d7073206f76657220746865206c617a7920646f67"), []interface{}{"The quick brown fox jumps over the lazy dog"}},
	{"array", hexDecode("981a0102030405060708090a0b0c0d0e0f101112131415161718181819181a"), []interface{}{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}}},
	{"map", hexDecode("ad616161416162614261636143616461446165614561666146616761476168614861696149616a614a616c614c616d614d616e614e"), []interface{}{map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"}, strc{A: "A", B: "B", C: "C", D: "D", E: "E", F: "F", G: "G", H: "H", I: "I", J: "J", L: "L", M: "M", N: "N"}}},
}

func BenchmarkUnmarshal(b *testing.B) {
	for _, bm := range decodeBenchmarks {
		for _, t := range bm.decodeToTypes {
			name := "CBOR " + bm.name + " to Go " + t.String()
			if t.Kind() == reflect.Struct {
				name = "CBOR " + bm.name + " to Go " + t.Kind().String()
			}
			b.Run(name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					vPtr := reflect.New(t).Interface()
					if err := cbor.Unmarshal(bm.cborData, vPtr); err != nil {
						b.Fatal("Unmarshal:", err)
					}
				}
			})
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	for _, bm := range decodeBenchmarks {
		for _, t := range bm.decodeToTypes {
			name := "CBOR " + bm.name + " to Go " + t.String()
			if t.Kind() == reflect.Struct {
				name = "CBOR " + bm.name + " to Go " + t.Kind().String()
			}
			buf := bytes.NewReader(bm.cborData)
			decoder := cbor.NewDecoder(buf)
			b.Run(name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					vPtr := reflect.New(t).Interface()
					if err := decoder.Decode(vPtr); err != nil {
						b.Fatal("Decode:", err)
					}
					buf.Seek(0, 0)
				}
			})
		}
	}
}

func BenchmarkDecodeStream(b *testing.B) {
	var cborData []byte
	for _, bm := range decodeBenchmarks {
		for i := 0; i < len(bm.decodeToTypes); i++ {
			cborData = append(cborData, bm.cborData...)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := bytes.NewReader(cborData)
		decoder := cbor.NewDecoder(buf)
		for j := 0; j < rounds; j++ {
			for _, bm := range decodeBenchmarks {
				for _, t := range bm.decodeToTypes {
					vPtr := reflect.New(t).Interface()
					if err := decoder.Decode(vPtr); err != nil {
						b.Fatal("Decode:", err)
					}
				}
			}
			buf.Seek(0, 0)
		}
	}
}

func BenchmarkMarshal(b *testing.B) {
	for _, bm := range encodeBenchmarks {
		for _, v := range bm.values {
			name := "Go " + reflect.TypeOf(v).String() + " to CBOR " + bm.name
			if reflect.TypeOf(v).Kind() == reflect.Struct {
				name = "Go " + reflect.TypeOf(v).Kind().String() + " to CBOR " + bm.name
			}
			b.Run(name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if _, err := cbor.Marshal(v, cbor.EncOptions{}); err != nil {
						b.Fatal("Marshal:", err)
					}
				}
			})
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	for _, bm := range encodeBenchmarks {
		for _, v := range bm.values {
			name := "Go " + reflect.TypeOf(v).String() + " to CBOR " + bm.name
			if reflect.TypeOf(v).Kind() == reflect.Struct {
				name = "Go " + reflect.TypeOf(v).Kind().String() + " to CBOR " + bm.name
			}
			b.Run(name, func(b *testing.B) {
				encoder := cbor.NewEncoder(ioutil.Discard, cbor.EncOptions{})
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := encoder.Encode(v); err != nil {
						b.Fatal("Encode:", err)
					}
				}
			})
		}
	}
}

func BenchmarkEncodeStream(b *testing.B) {
	for i := 0; i < b.N; i++ {
		encoder := cbor.NewEncoder(ioutil.Discard, cbor.EncOptions{})
		for i := 0; i < rounds; i++ {
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
