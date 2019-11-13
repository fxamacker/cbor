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

func BenchmarkMarshalCanonical(b *testing.B) {
	for _, bm := range []struct {
		name     string
		cborData []byte
		values   []interface{}
	}{
		{"map", hexDecode("ad616161416162614261636143616461446165614561666146616761476168614861696149616a614a616c614c616d614d616e614e"), []interface{}{map[string]string{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"}, strc{A: "A", B: "B", C: "C", D: "D", E: "E", F: "F", G: "G", H: "H", I: "I", J: "J", L: "L", M: "M", N: "N"}}},
	} {
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
			// Canonical encoding
			name = "Go " + reflect.TypeOf(v).String() + " to CBOR " + bm.name + " canonical"
			if reflect.TypeOf(v).Kind() == reflect.Struct {
				name = "Go " + reflect.TypeOf(v).Kind().String() + " to CBOR " + bm.name + " canonical"
			}
			b.Run(name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if _, err := cbor.Marshal(v, cbor.EncOptions{Canonical: true}); err != nil {
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

func BenchmarkDecodeCOSE(b *testing.B) {
	cborData := hexDecode("a50102032620012158205af8047e9085ef79ec321280c7b95844d707d7fe4d73cd648f044c619ee74f6b22582036bb8c00768e90858012dc3831e15a389072bbdbe7e2e19155db9e1197655edf")
	for _, ctor := range []func() interface{}{
		func() interface{} { return new(interface{}) },
		func() interface{} { return new(map[interface{}]interface{}) },
		func() interface{} { return new(map[int]interface{}) },
	} {
		name := "COSE to Go " + reflect.TypeOf(ctor()).Elem().String()
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				v := ctor()
				if err := cbor.Unmarshal(cborData, v); err != nil {
					b.Fatal("Unmarshal:", err)
				}
			}
		})
	}
}

func BenchmarkDecodeWebAuthn(b *testing.B) {
	type attestationObject struct {
		AuthnData []byte          `cbor:"authData"`
		Fmt       string          `cbor:"fmt"`
		AttStmt   cbor.RawMessage `cbor:"attStmt"`
	}
	cborData := hexDecode("a363666d74686669646f2d7532666761747453746d74a26373696758483046022100e7ab373cfbd99fcd55fd59b0f6f17fef5b77a20ddec3db7f7e4d55174e366236022100828336b4822125fb56541fb14a8a273876acd339395ec2dad95cf41c1dd2a9ae637835638159024e3082024a30820132a0030201020204124a72fe300d06092a864886f70d01010b0500302e312c302a0603550403132359756269636f2055324620526f6f742043412053657269616c203435373230303633313020170d3134303830313030303030305a180f32303530303930343030303030305a302c312a302806035504030c2159756269636f205532462045452053657269616c203234393431343937323135383059301306072a8648ce3d020106082a8648ce3d030107034200043d8b1bbd2fcbf6086e107471601468484153c1c6d3b4b68a5e855e6e40757ee22bcd8988bf3befd7cdf21cb0bf5d7a150d844afe98103c6c6607d9faae287c02a33b3039302206092b0601040182c40a020415312e332e362e312e342e312e34313438322e312e313013060b2b0601040182e51c020101040403020520300d06092a864886f70d01010b05000382010100a14f1eea0076f6b8476a10a2be72e60d0271bb465b2dfbfc7c1bd12d351989917032631d795d097fa30a26a325634e85721bc2d01a86303f6bc075e5997319e122148b0496eec8d1f4f94cf4110de626c289443d1f0f5bbb239ca13e81d1d5aa9df5af8e36126475bfc23af06283157252762ff68879bcf0ef578d55d67f951b4f32b63c8aea5b0f99c67d7d814a7ff5a6f52df83e894a3a5d9c8b82e7f8bc8daf4c80175ff8972fda79333ec465d806eacc948f1bab22045a95558a48c20226dac003d41fbc9e05ea28a6bb5e10a49de060a0a4f6a2676a34d68c4abe8c61874355b9027e828ca9e064b002d62e8d8cf0744921753d35e3c87c5d5779453e7768617574684461746158c449960de5880e8c687434170f6476605b8fe4aeb9a28632c7995cf3ba831d976341000000000000000000000000000000000000000000408903fd7dfd2c9770e98cae0123b13a2c27828a106349bc6277140e7290b7e9eb7976aa3c04ed347027caf7da3a2fa76304751c02208acfc4e7fc6c7ebbc375c8a5010203262001215820ad7f7992c335b90d882b2802061b97a4fabca7e2ee3e7a51e728b8055e4eb9c7225820e0966ba7005987fece6f0e0e13447aa98cec248e4000a594b01b74c1cb1d40b3")
	for _, ctor := range []func() interface{}{
		func() interface{} { return new(interface{}) },
		func() interface{} { return new(map[interface{}]interface{}) },
		func() interface{} { return new(map[string]interface{}) },
		func() interface{} { return new(attestationObject) },
	} {
		t := reflect.TypeOf(ctor()).Elem()
		name := "atten object to Go " + t.String()
		if t.Kind() == reflect.Struct {
			name = "atten object to Go " + t.Kind().String()
		}
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				v := ctor()
				if err := cbor.Unmarshal(cborData, v); err != nil {
					b.Fatal("Unmarshal:", err)
				}
			}
		})
	}
}
