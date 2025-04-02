// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"testing"
)

func TestValid1(t *testing.T) {
	for _, mt := range marshalTests {
		if err := Wellformed(mt.wantData); err != nil {
			t.Errorf("Wellformed() returned error %v", err)
		}
	}
}

func TestValid2(t *testing.T) {
	for _, mt := range marshalTests {
		dm, _ := DecOptions{DupMapKey: DupMapKeyEnforcedAPF}.DecMode()
		if err := dm.Wellformed(mt.wantData); err != nil {
			t.Errorf("Wellformed() returned error %v", err)
		}
	}
}

func TestValidExtraneousData(t *testing.T) {
	testCases := []struct {
		name                     string
		data                     []byte
		extraneousDataNumOfBytes int
		extraneousDataIndex      int
	}{
		{"two numbers", []byte{0x00, 0x01}, 1, 1},                                // 0, 1
		{"bytestring and int", []byte{0x44, 0x01, 0x02, 0x03, 0x04, 0x00}, 1, 5}, // h'01020304', 0
		{"int and partial array", []byte{0x00, 0x83, 0x01, 0x02}, 3, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Wellformed(tc.data)
			if err == nil {
				t.Errorf("Wellformed(0x%x) didn't return an error", tc.data)
			} else {
				ederr, ok := err.(*ExtraneousDataError)
				if !ok {
					t.Errorf("Wellformed(0x%x) error type %T, want *ExtraneousDataError", tc.data, err)
				} else if ederr.numOfBytes != tc.extraneousDataNumOfBytes {
					t.Errorf("Wellformed(0x%x) returned %d bytes of extraneous data, want %d", tc.data, ederr.numOfBytes, tc.extraneousDataNumOfBytes)
				} else if ederr.index != tc.extraneousDataIndex {
					t.Errorf("Wellformed(0x%x) returned extraneous data index %d, want %d", tc.data, ederr.index, tc.extraneousDataIndex)
				}
			}
		})
	}
}

func TestValidOnStreamingData(t *testing.T) {
	var buf bytes.Buffer
	for _, t := range marshalTests {
		buf.Write(t.wantData)
	}
	d := decoder{data: buf.Bytes(), dm: defaultDecMode}
	for i := 0; i < len(marshalTests); i++ {
		if err := d.wellformed(true, false); err != nil {
			t.Errorf("wellformed() returned error %v", err)
		}
	}
}

func TestDepth(t *testing.T) {
	testCases := []struct {
		name      string
		data      []byte
		wantDepth int
	}{
		{"uint", hexDecode("00"), 0},                                                          // 0
		{"int", hexDecode("20"), 0},                                                           // -1
		{"bool", hexDecode("f4"), 0},                                                          // false
		{"nil", hexDecode("f6"), 0},                                                           // nil
		{"float", hexDecode("fa47c35000"), 0},                                                 // 100000.0
		{"byte string", hexDecode("40"), 0},                                                   // []byte{}
		{"indefinite length byte string", hexDecode("5f42010243030405ff"), 0},                 // []byte{1, 2, 3, 4, 5}
		{"text string", hexDecode("60"), 0},                                                   // ""
		{"indefinite length text string", hexDecode("7f657374726561646d696e67ff"), 0},         // "streaming"
		{"empty array", hexDecode("80"), 1},                                                   // []
		{"indefinite length empty array", hexDecode("9fff"), 1},                               // []
		{"array", hexDecode("98190102030405060708090a0b0c0d0e0f101112131415161718181819"), 1}, // [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25]
		{"indefinite length array", hexDecode("9f0102030405060708090a0b0c0d0e0f101112131415161718181819ff"), 1}, // [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25]
		{"nested array", hexDecode("8301820203820405"), 2},                                                      // [1,[2,3],[4,5]]
		{"indefinite length nested array", hexDecode("83018202039f0405ff"), 2},                                  // [1,[2,3],[4,5]]
		{"array and map", hexDecode("826161a161626163"), 2},                                                     // [a", {"b": "c"}]
		{"indefinite length array and map", hexDecode("826161bf61626163ff"), 2},                                 // [a", {"b": "c"}]
		{"empty map", hexDecode("a0"), 1},                                                                       // {}
		{"indefinite length empty map", hexDecode("bfff"), 1},                                                   // {}
		{"map", hexDecode("a201020304"), 1},                                                                     // {1:2, 3:4}
		{"nested map", hexDecode("a26161016162820203"), 2},                                                      // {"a": 1, "b": [2, 3]}
		{"indefinite length nested map", hexDecode("bf61610161629f0203ffff"), 2},                                // {"a": 1, "b": [2, 3]}
		{"tag", hexDecode("c074323031332d30332d32315432303a30343a30305a"), 0},                                   // 0("2013-03-21T20:04:00Z")
		{"tagged map", hexDecode("d864a26161016162820203"), 2},                                                  // 100({"a": 1, "b": [2, 3]})
		{"tagged map and array", hexDecode("d864a26161016162d865820203"), 2},                                    // 100({"a": 1, "b": 101([2, 3])})
		{"tagged map and array", hexDecode("d864a26161016162d865d866820203"), 3},                                // 100({"a": 1, "b": 101(102([2, 3]))})
		{"nested tag", hexDecode("d864d865d86674323031332d30332d32315432303a30343a30305a"), 2},                  // 100(101(102("2013-03-21T20:04:00Z")))
		{"32-level array", hexDecode("82018181818181818181818181818181818181818181818181818181818181818101"), 32},
		{"32-level indefinite length array", hexDecode("9f018181818181818181818181818181818181818181818181818181818181818101ff"), 32},
		{"32-level map", hexDecode("a1018181818181818181818181818181818181818181818181818181818181818101"), 32},
		{"32-level indefinite length map", hexDecode("bf018181818181818181818181818181818181818181818181818181818181818101ff"), 32},
		{"32-level tag", hexDecode("d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d86474323031332d30332d32315432303a30343a30305a"), 32}, // 100(100(...("2013-03-21T20:04:00Z")))
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := decoder{data: tc.data, dm: defaultDecMode}
			depth, err := d.wellformedInternal(0, false)
			if err != nil {
				t.Errorf("wellformed(0x%x) returned error %v", tc.data, err)
			}
			if depth != tc.wantDepth {
				t.Errorf("wellformed(0x%x) returned depth %d, want %d", tc.data, depth, tc.wantDepth)
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
			data:         hexDecode("82018181818181818181818181818181818181818181818181818181818181818101"),
			opts:         DecOptions{MaxNestedLevels: 4},
			wantErrorMsg: "cbor: exceeded max nested level 4",
		},
		{
			name:         "33-level array",
			data:         hexDecode("82018181818181818181818181818181818181818181818181818181818181818101"),
			opts:         DecOptions{MaxNestedLevels: 10},
			wantErrorMsg: "cbor: exceeded max nested level 10",
		},
		{
			name:         "33-level array",
			data:         hexDecode("8201818181818181818181818181818181818181818181818181818181818181818101"),
			opts:         DecOptions{},
			wantErrorMsg: "cbor: exceeded max nested level 32",
		},
		{
			name:         "33-level indefinite length array",
			data:         hexDecode("9f01818181818181818181818181818181818181818181818181818181818181818101ff"),
			opts:         DecOptions{},
			wantErrorMsg: "cbor: exceeded max nested level 32",
		},
		{
			name:         "33-level map",
			data:         hexDecode("a101818181818181818181818181818181818181818181818181818181818181818101"),
			opts:         DecOptions{},
			wantErrorMsg: "cbor: exceeded max nested level 32",
		},
		{
			name:         "33-level indefinite length map",
			data:         hexDecode("bf01818181818181818181818181818181818181818181818181818181818181818101ff"),
			opts:         DecOptions{},
			wantErrorMsg: "cbor: exceeded max nested level 32",
		},
		{
			name:         "33-level tag",
			data:         hexDecode("d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d86474323031332d30332d32315432303a30343a30305a"),
			opts:         DecOptions{},
			wantErrorMsg: "cbor: exceeded max nested level 32",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dm, _ := tc.opts.decMode()
			d := decoder{data: tc.data, dm: dm}
			if _, err := d.wellformedInternal(0, false); err == nil {
				t.Errorf("wellformed(0x%x) didn't return an error", tc.data)
			} else if _, ok := err.(*MaxNestedLevelError); !ok {
				t.Errorf("wellformed(0x%x) returned wrong error type %T, want (*MaxNestedLevelError)", tc.data, err)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("wellformed(0x%x) returned error %q, want error %q", tc.data, err.Error(), tc.wantErrorMsg)
			}
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
			data: hexDecode("c074323031332d30332d32315432303a30343a30305a"),
		},
		{
			name: "tag 1",
			data: hexDecode("c11a514b67b0"),
		},
		{
			name: "tag 2",
			data: hexDecode("c249010000000000000000"),
		},
		{
			name: "tag 3",
			data: hexDecode("c349010000000000000000"),
		},
		{
			name: "nested tag 0",
			data: hexDecode("d9d9f7c074323031332d30332d32315432303a30343a30305a"),
		},
		{
			name: "nested tag 1",
			data: hexDecode("d9d9f7c11a514b67b0"),
		},
		{
			name: "nested tag 2",
			data: hexDecode("d9d9f7c249010000000000000000"),
		},
		{
			name: "nested tag 3",
			data: hexDecode("d9d9f7c349010000000000000000"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := decoder{data: tc.data, dm: defaultDecMode}
			if err := d.wellformed(true, true); err != nil {
				t.Errorf("wellformed(0x%x) returned error %v", tc.data, err)
			}
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
			data:         hexDecode("c01a514b67b0"),
			wantErrorMsg: "cbor: tag number 0 must be followed by text string, got positive integer",
		},
		{
			name:         "tag 1",
			data:         hexDecode("c174323031332d30332d32315432303a30343a30305a"),
			wantErrorMsg: "cbor: tag number 1 must be followed by integer or floating-point number, got UTF-8 text string",
		},
		{
			name:         "tag 2",
			data:         hexDecode("c269010000000000000000"),
			wantErrorMsg: "cbor: tag number 2 or 3 must be followed by byte string, got UTF-8 text string",
		},
		{
			name:         "tag 3",
			data:         hexDecode("c300"),
			wantErrorMsg: "cbor: tag number 2 or 3 must be followed by byte string, got positive integer",
		},
		{
			name:         "nested tag 0",
			data:         hexDecode("d9d9f7c01a514b67b0"),
			wantErrorMsg: "cbor: tag number 0 must be followed by text string, got positive integer",
		},
		{
			name:         "nested tag 1",
			data:         hexDecode("d9d9f7c174323031332d30332d32315432303a30343a30305a"),
			wantErrorMsg: "cbor: tag number 1 must be followed by integer or floating-point number, got UTF-8 text string",
		},
		{
			name:         "nested tag 2",
			data:         hexDecode("d9d9f7c269010000000000000000"),
			wantErrorMsg: "cbor: tag number 2 or 3 must be followed by byte string, got UTF-8 text string",
		},
		{
			name:         "nested tag 3",
			data:         hexDecode("d9d9f7c300"),
			wantErrorMsg: "cbor: tag number 2 or 3 must be followed by byte string, got positive integer",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := decoder{data: tc.data, dm: defaultDecMode}
			err := d.wellformed(true, true)
			if err == nil {
				t.Errorf("wellformed(0x%x) didn't return an error", tc.data)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("wellformed(0x%x) error %q, want %q", tc.data, err.Error(), tc.wantErrorMsg)
			}
		})
	}
}
