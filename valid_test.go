// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestValid1(t *testing.T) {
	for _, tc := range marshalTestCases {
		err := Wellformed(tc.wantData)
		checkNoError(t, "Wellformed", err)
	}
}

func TestValid2(t *testing.T) {
	for _, tc := range marshalTestCases {
		dm := mustDecMode(t, DecOptions{DupMapKey: DupMapKeyEnforcedAPF})
		err := dm.Wellformed(tc.wantData)
		checkNoError(t, "dm.Wellformed", err)
		err = dm.(*decMode).wellformedSyntaxMode.Wellformed(tc.wantData)
		checkNoError(t, "wellformedSyntaxMode.Wellformed", err)
	}
}

func TestValidExtraneousData(t *testing.T) {
	testCases := []struct {
		name                     string
		data                     []byte
		extraneousDataNumOfBytes int
		extraneousDataIndex      int
	}{
		{
			name:                     "two numbers",
			data:                     []byte{0x00, 0x01},
			extraneousDataNumOfBytes: 1,
			extraneousDataIndex:      1,
		}, // 0, 1
		{
			name:                     "bytestring and int",
			data:                     []byte{0x44, 0x01, 0x02, 0x03, 0x04, 0x00},
			extraneousDataNumOfBytes: 1,
			extraneousDataIndex:      5,
		}, // h'01020304', 0
		{
			name:                     "int and partial array",
			data:                     []byte{0x00, 0x83, 0x01, 0x02},
			extraneousDataNumOfBytes: 3,
			extraneousDataIndex:      1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Wellformed(tc.data)
			checkError[*ExtraneousDataError](t, "Wellformed", err, (&ExtraneousDataError{numOfBytes: tc.extraneousDataNumOfBytes, index: tc.extraneousDataIndex}).Error())
		})
	}
}

func TestValidOnStreamingData(t *testing.T) {
	var buf bytes.Buffer
	for _, tc := range marshalTestCases {
		buf.Write(tc.wantData)
	}
	d := decoder{data: buf.Bytes(), dm: defaultDecMode}
	for range marshalTestCases {
		err := d.wellformed(true, false)
		checkNoError(t, "d.wellformed", err)
	}
}

func TestDepth(t *testing.T) {
	testCases := []struct {
		name      string
		data      []byte
		wantDepth int
	}{
		{
			name:      "uint",
			data:      mustHexDecode("00"),
			wantDepth: 0,
		}, // 0
		{
			name:      "int",
			data:      mustHexDecode("20"),
			wantDepth: 0,
		}, // -1
		{
			name:      "bool",
			data:      mustHexDecode("f4"),
			wantDepth: 0,
		}, // false
		{
			name:      "nil",
			data:      mustHexDecode("f6"),
			wantDepth: 0,
		}, // nil
		{
			name:      "float",
			data:      mustHexDecode("fa47c35000"),
			wantDepth: 0,
		}, // 100000.0
		{
			name:      "byte string",
			data:      mustHexDecode("40"),
			wantDepth: 0,
		}, // []byte{}
		{
			name:      "indefinite-length byte string",
			data:      mustHexDecode("5f42010243030405ff"),
			wantDepth: 0,
		}, // []byte{1, 2, 3, 4, 5}
		{
			name:      "text string",
			data:      mustHexDecode("60"),
			wantDepth: 0,
		}, // ""
		{
			name:      "indefinite-length text string",
			data:      mustHexDecode("7f657374726561646d696e67ff"),
			wantDepth: 0,
		}, // "streaming"
		{
			name:      "empty array",
			data:      mustHexDecode("80"),
			wantDepth: 1,
		}, // []
		{
			name:      "indefinite-length empty array",
			data:      mustHexDecode("9fff"),
			wantDepth: 1,
		}, // []
		{
			name:      "array",
			data:      mustHexDecode("98190102030405060708090a0b0c0d0e0f101112131415161718181819"),
			wantDepth: 1,
		}, // [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25]
		{
			name:      "indefinite-length array",
			data:      mustHexDecode("9f0102030405060708090a0b0c0d0e0f101112131415161718181819ff"),
			wantDepth: 1,
		}, // [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25]
		{
			name:      "nested array",
			data:      mustHexDecode("8301820203820405"),
			wantDepth: 2,
		}, // [1,[2,3],[4,5]]
		{
			name:      "indefinite-length nested array",
			data:      mustHexDecode("83018202039f0405ff"),
			wantDepth: 2,
		}, // [1,[2,3],[4,5]]
		{
			name:      "array and map",
			data:      mustHexDecode("826161a161626163"),
			wantDepth: 2,
		}, // [a", {"b": "c"}]
		{
			name:      "indefinite-length array and map",
			data:      mustHexDecode("826161bf61626163ff"),
			wantDepth: 2,
		}, // [a", {"b": "c"}]
		{
			name:      "empty map",
			data:      mustHexDecode("a0"),
			wantDepth: 1,
		}, // {}
		{
			name:      "indefinite-length empty map",
			data:      mustHexDecode("bfff"),
			wantDepth: 1,
		}, // {}
		{
			name:      "map",
			data:      mustHexDecode("a201020304"),
			wantDepth: 1,
		}, // {1:2, 3:4}
		{
			name:      "nested map",
			data:      mustHexDecode("a26161016162820203"),
			wantDepth: 2,
		}, // {"a": 1, "b": [2, 3]}
		{
			name:      "indefinite-length nested map",
			data:      mustHexDecode("bf61610161629f0203ffff"),
			wantDepth: 2,
		}, // {"a": 1, "b": [2, 3]}
		{
			name:      "tag",
			data:      mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
			wantDepth: 0,
		}, // 0("2013-03-21T20:04:00Z")
		{
			name:      "tagged map",
			data:      mustHexDecode("d864a26161016162820203"),
			wantDepth: 2,
		}, // 100({"a": 1, "b": [2, 3]})
		{
			name:      "tagged map and array",
			data:      mustHexDecode("d864a26161016162d865820203"),
			wantDepth: 2,
		}, // 100({"a": 1, "b": 101([2, 3])})
		{
			name:      "tagged map and array",
			data:      mustHexDecode("d864a26161016162d865d866820203"),
			wantDepth: 3,
		}, // 100({"a": 1, "b": 101(102([2, 3]))})
		{
			name:      "nested tag",
			data:      mustHexDecode("d864d865d86674323031332d30332d32315432303a30343a30305a"),
			wantDepth: 2,
		}, // 100(101(102("2013-03-21T20:04:00Z")))
		{
			name:      "32-level array",
			data:      mustHexDecode("82018181818181818181818181818181818181818181818181818181818181818101"),
			wantDepth: 32,
		},
		{
			name:      "32-level indefinite-length array",
			data:      mustHexDecode("9f018181818181818181818181818181818181818181818181818181818181818101ff"),
			wantDepth: 32,
		},
		{
			name:      "32-level map",
			data:      mustHexDecode("a1018181818181818181818181818181818181818181818181818181818181818101"),
			wantDepth: 32,
		},
		{
			name:      "32-level indefinite-length map",
			data:      mustHexDecode("bf018181818181818181818181818181818181818181818181818181818181818101ff"),
			wantDepth: 32,
		},
		{
			name:      "32-level tag",
			data:      mustHexDecode("d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d86474323031332d30332d32315432303a30343a30305a"),
			wantDepth: 32,
		}, // 100(100(...("2013-03-21T20:04:00Z")))
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := decoder{data: tc.data, dm: defaultDecMode}
			depth, err := d.wellformedInternal(0, false)
			checkNoError(t, "d.wellformedInternal", err)
			if depth != tc.wantDepth {
				t.Errorf("d.wellformedInternal(0x%x) depth = %d, want %d", tc.data, depth, tc.wantDepth)
			}
		})
	}
}

func TestDepthError(t *testing.T) {
	testCases := []struct {
		name         string
		data         []byte
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "33-level array",
			data:         mustHexDecode("82018181818181818181818181818181818181818181818181818181818181818101"),
			opts:         DecOptions{MaxNestedLevels: 4},
			wantErrorMsg: "cbor: exceeded max nested level 4",
		},
		{
			name:         "33-level array",
			data:         mustHexDecode("82018181818181818181818181818181818181818181818181818181818181818101"),
			opts:         DecOptions{MaxNestedLevels: 10},
			wantErrorMsg: "cbor: exceeded max nested level 10",
		},
		{
			name:         "33-level array",
			data:         mustHexDecode("8201818181818181818181818181818181818181818181818181818181818181818101"),
			opts:         DecOptions{},
			wantErrorMsg: "cbor: exceeded max nested level 32",
		},
		{
			name:         "33-level indefinite-length array",
			data:         mustHexDecode("9f01818181818181818181818181818181818181818181818181818181818181818101ff"),
			opts:         DecOptions{},
			wantErrorMsg: "cbor: exceeded max nested level 32",
		},
		{
			name:         "33-level map",
			data:         mustHexDecode("a101818181818181818181818181818181818181818181818181818181818181818101"),
			opts:         DecOptions{},
			wantErrorMsg: "cbor: exceeded max nested level 32",
		},
		{
			name:         "33-level indefinite-length map",
			data:         mustHexDecode("bf01818181818181818181818181818181818181818181818181818181818181818101ff"),
			opts:         DecOptions{},
			wantErrorMsg: "cbor: exceeded max nested level 32",
		},
		{
			name:         "33-level tag",
			data:         mustHexDecode("d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d86474323031332d30332d32315432303a30343a30305a"),
			opts:         DecOptions{},
			wantErrorMsg: "cbor: exceeded max nested level 32",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dm, _ := tc.opts.decMode()
			d := decoder{data: tc.data, dm: dm}
			_, err := d.wellformedInternal(0, false)
			checkError[*MaxNestedLevelError](t, "d.wellformedInternal", err, tc.wantErrorMsg)
		})
	}
}

func TestValidBuiltinTagTest(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "tag 0",
			data: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
		},
		{
			name: "tag 1",
			data: mustHexDecode("c11a514b67b0"),
		},
		{
			name: "tag 2",
			data: mustHexDecode("c249010000000000000000"),
		},
		{
			name: "tag 3",
			data: mustHexDecode("c349010000000000000000"),
		},
		{
			name: "nested tag 0",
			data: mustHexDecode("d9d9f7c074323031332d30332d32315432303a30343a30305a"),
		},
		{
			name: "nested tag 1",
			data: mustHexDecode("d9d9f7c11a514b67b0"),
		},
		{
			name: "nested tag 2",
			data: mustHexDecode("d9d9f7c249010000000000000000"),
		},
		{
			name: "nested tag 3",
			data: mustHexDecode("d9d9f7c349010000000000000000"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := decoder{data: tc.data, dm: defaultDecMode}
			err := d.wellformed(true, true)
			checkNoError(t, "d.wellformed", err)
		})
	}
}

func TestInvalidBuiltinTagTest(t *testing.T) {
	testCases := []struct {
		name         string
		data         []byte
		wantErrorMsg string
	}{
		{
			name:         "tag 0",
			data:         mustHexDecode("c01a514b67b0"),
			wantErrorMsg: "cbor: tag number 0 must be followed by text string, got positive integer",
		},
		{
			name:         "tag 1",
			data:         mustHexDecode("c174323031332d30332d32315432303a30343a30305a"),
			wantErrorMsg: "cbor: tag number 1 must be followed by integer or floating-point number, got UTF-8 text string",
		},
		{
			name:         "tag 2",
			data:         mustHexDecode("c269010000000000000000"),
			wantErrorMsg: "cbor: tag number 2 or 3 must be followed by byte string, got UTF-8 text string",
		},
		{
			name:         "tag 3",
			data:         mustHexDecode("c300"),
			wantErrorMsg: "cbor: tag number 2 or 3 must be followed by byte string, got positive integer",
		},
		{
			name:         "nested tag 0",
			data:         mustHexDecode("d9d9f7c01a514b67b0"),
			wantErrorMsg: "cbor: tag number 0 must be followed by text string, got positive integer",
		},
		{
			name:         "nested tag 1",
			data:         mustHexDecode("d9d9f7c174323031332d30332d32315432303a30343a30305a"),
			wantErrorMsg: "cbor: tag number 1 must be followed by integer or floating-point number, got UTF-8 text string",
		},
		{
			name:         "nested tag 2",
			data:         mustHexDecode("d9d9f7c269010000000000000000"),
			wantErrorMsg: "cbor: tag number 2 or 3 must be followed by byte string, got UTF-8 text string",
		},
		{
			name:         "nested tag 3",
			data:         mustHexDecode("d9d9f7c300"),
			wantErrorMsg: "cbor: tag number 2 or 3 must be followed by byte string, got positive integer",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := decoder{data: tc.data, dm: defaultDecMode}
			err := d.wellformed(true, true)
			checkError[error](t, "d.wellformed", err, tc.wantErrorMsg)
		})
	}
}

func TestWellformedSyntaxMode(t *testing.T) {
	rejectedSimpleValues, err := NewSimpleValueRegistryFromDefaults(WithRejectedSimpleValue(16))
	checkNoError(t, "NewSimpleValueRegistryFromDefaults", err)

	type taggedFloat float64
	tags := NewTagSet()
	mustAddTag(t, tags, TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired}, reflect.TypeFor[taggedFloat](), 100)

	testCases := []struct {
		name                    string
		opts                    DecOptions
		tags                    TagSet
		data                    []byte
		wantWellformedErr       error // nil, or typed nil pointer of wanted error type
		wantWellformedSyntaxErr error // nil, or typed nil pointer of wanted error type
	}{
		{
			name: "default options",
			opts: DecOptions{},
			data: []byte{0xf9, 0x7e, 0x00}, // NaN
		},
		{
			name:                    "extraneous data",
			opts:                    DecOptions{NaN: NaNDecodeForbidden},
			data:                    []byte{0x00, 0x00},
			wantWellformedErr:       (*ExtraneousDataError)(nil),
			wantWellformedSyntaxErr: (*ExtraneousDataError)(nil),
		},
		{
			name:                    "MaxNestedLevels exceeded",
			opts:                    DecOptions{NaN: NaNDecodeForbidden, MaxNestedLevels: 4},
			data:                    []byte{0x81, 0x81, 0x81, 0x81, 0x81, 0x01}, // [[[[[1]]]]]
			wantWellformedErr:       (*MaxNestedLevelError)(nil),
			wantWellformedSyntaxErr: (*MaxNestedLevelError)(nil),
		},
		{
			name:                    "MaxArrayElements exceeded",
			opts:                    DecOptions{NaN: NaNDecodeForbidden, MaxArrayElements: 16},
			data:                    append([]byte{0x91}, make([]byte, 17)...), // [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]
			wantWellformedErr:       (*MaxArrayElementsError)(nil),
			wantWellformedSyntaxErr: (*MaxArrayElementsError)(nil),
		},
		{
			name:                    "MaxMapPairs exceeded",
			opts:                    DecOptions{NaN: NaNDecodeForbidden, MaxMapPairs: 16},
			data:                    append([]byte{0xb1}, make([]byte, 17*2)...),
			wantWellformedErr:       (*MaxMapPairsError)(nil),
			wantWellformedSyntaxErr: (*MaxMapPairsError)(nil),
		},
		{
			name:              "NaN forbidden",
			opts:              DecOptions{NaN: NaNDecodeForbidden},
			data:              []byte{0xa1, 0x01, 0xf9, 0x7e, 0x00}, // {1: NaN}
			wantWellformedErr: (*UnacceptableDataItemError)(nil),
		},
		{
			name:              "Inf forbidden",
			opts:              DecOptions{Inf: InfDecodeForbidden},
			data:              []byte{0xf9, 0x7c, 0x00}, // Infinity
			wantWellformedErr: (*UnacceptableDataItemError)(nil),
		},
		{
			name:              "simple value rejected",
			opts:              DecOptions{SimpleValues: rejectedSimpleValues},
			data:              []byte{0xf0}, // simple(16)
			wantWellformedErr: (*UnacceptableDataItemError)(nil),
		},
		{
			name:              "bignum tag forbidden",
			opts:              DecOptions{BignumTag: BignumTagForbidden},
			data:              []byte{0xc2, 0x41, 0x01}, // 2('1')
			wantWellformedErr: (*UnacceptableDataItemError)(nil),
		},
		{
			name:              "tags forbidden",
			opts:              DecOptions{TagsMd: TagsForbidden},
			data:              []byte{0xc1, 0x01}, // 1(1)
			wantWellformedErr: (*TagsMdError)(nil),
		},
		{
			name:              "NaN forbidden with tags",
			opts:              DecOptions{NaN: NaNDecodeForbidden},
			tags:              tags,
			data:              []byte{0xd8, 0x64, 0xf9, 0x7e, 0x00}, // 100(NaN)
			wantWellformedErr: (*UnacceptableDataItemError)(nil),
		},
		{
			name:              "indefinite-length array forbidden",
			opts:              DecOptions{IndefLength: IndefLengthForbidden},
			data:              []byte{0x9f, 0x01, 0xff}, // [_ 1]
			wantWellformedErr: (*IndefiniteLengthError)(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var dm DecMode
			if tc.tags != nil {
				dm = mustDecModeWithTags(t, tc.opts, tc.tags)
			} else {
				dm = mustDecMode(t, tc.opts)
			}

			name := fmt.Sprintf("Wellformed(0x%x)", tc.data)
			err := dm.Wellformed(tc.data)
			checkErrorType(t, name, err, tc.wantWellformedErr)

			name = fmt.Sprintf("wellformedSyntaxMode.Wellformed(0x%x)", tc.data)
			err = dm.(*decMode).wellformedSyntaxMode.Wellformed(tc.data)
			checkErrorType(t, name, err, tc.wantWellformedSyntaxErr)
		})
	}
}
