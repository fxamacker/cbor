// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestUnmarshalSimpleValue(t *testing.T) {
	t.Run("0..23", func(t *testing.T) {
		for i := range 24 {
			data := []byte{byte(cborTypePrimitives) | byte(i)}
			want := SimpleValue(i)

			switch i {
			case 20: // false
				testUnmarshalSimpleValueToEmptyInterface(t, data, false)
			case 21: // true
				testUnmarshalSimpleValueToEmptyInterface(t, data, true)
			case 22: // null
				testUnmarshalSimpleValueToEmptyInterface(t, data, nil)
			case 23: // undefined
				testUnmarshalSimpleValueToEmptyInterface(t, data, nil)
			default:
				testUnmarshalSimpleValueToEmptyInterface(t, data, want)
			}

			testUnmarshalSimpleValue(t, data, want)
		}
	})

	t.Run("24..31", func(t *testing.T) {
		for i := 24; i <= 31; i++ {
			data := []byte{byte(cborTypePrimitives) | byte(24), byte(i)}

			testUnmarshalInvalidSimpleValueToEmptyInterface(t, data)
			testUnmarshalInvalidSimpleValue(t, data)
		}
	})

	t.Run("32..255", func(t *testing.T) {
		for i := 32; i <= 255; i++ {
			data := []byte{byte(cborTypePrimitives) | byte(24), byte(i)}
			want := SimpleValue(i)
			testUnmarshalSimpleValueToEmptyInterface(t, data, want)
			testUnmarshalSimpleValue(t, data, want)
		}
	})
}

func TestUnmarshalSimpleValueOnBadData(t *testing.T) {
	testCases := []struct {
		name         string
		data         []byte
		wantErrorMsg string
	}{
		// Empty data
		{
			name:         "nil data",
			data:         nil,
			wantErrorMsg: io.EOF.Error(),
		},
		{
			name:         "empty data",
			data:         []byte{},
			wantErrorMsg: io.EOF.Error(),
		},

		// Wrong CBOR types
		{
			name:         "uint type",
			data:         mustHexDecode("01"),
			wantErrorMsg: "cbor: cannot unmarshal positive integer into Go value of type SimpleValue",
		},
		{
			name:         "int type",
			data:         mustHexDecode("20"),
			wantErrorMsg: "cbor: cannot unmarshal negative integer into Go value of type SimpleValue",
		},
		{
			name:         "byte string type",
			data:         mustHexDecode("40"),
			wantErrorMsg: "cbor: cannot unmarshal byte string into Go value of type SimpleValue",
		},
		{
			name:         "string type",
			data:         mustHexDecode("60"),
			wantErrorMsg: "cbor: cannot unmarshal UTF-8 text string into Go value of type SimpleValue",
		},
		{
			name:         "array type",
			data:         mustHexDecode("80"),
			wantErrorMsg: "cbor: cannot unmarshal array into Go value of type SimpleValue",
		},
		{
			name:         "map type",
			data:         mustHexDecode("a0"),
			wantErrorMsg: "cbor: cannot unmarshal map into Go value of type SimpleValue",
		},
		{
			name:         "tag type",
			data:         mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
			wantErrorMsg: "cbor: cannot unmarshal tag into Go value of type SimpleValue",
		},
		{
			name:         "float type",
			data:         mustHexDecode("f90000"),
			wantErrorMsg: "cbor: cannot unmarshal primitives into Go value of type SimpleValue (not simple values)",
		},

		// Truncated CBOR data
		{
			name:         "truncated head",
			data:         mustHexDecode("18"),
			wantErrorMsg: io.ErrUnexpectedEOF.Error(),
		},

		// Truncated CBOR simple value
		{
			name:         "truncated simple value",
			data:         mustHexDecode("f8"),
			wantErrorMsg: io.ErrUnexpectedEOF.Error(),
		},

		// Invalid simple value
		{
			name:         "invalid simple value",
			data:         mustHexDecode("f800"),
			wantErrorMsg: "cbor: invalid simple value 0 for type primitives",
		},

		// Extraneous CBOR data
		{
			name:         "extraneous data",
			data:         mustHexDecode("f4f5"),
			wantErrorMsg: "cbor: 1 bytes of extraneous data starting at index 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test SimpleValue.UnmarshalCBOR(data)
			{
				var v SimpleValue

				err := v.UnmarshalCBOR(tc.data)
				if err == nil {
					t.Errorf("UnmarshalCBOR(%x) didn't return error", tc.data)
				}
				if !strings.HasPrefix(err.Error(), tc.wantErrorMsg) {
					t.Errorf("UnmarshalCBOR(%x) returned error %q, want %q", tc.data, err.Error(), tc.wantErrorMsg)
				}
			}
			// Test Unmarshal(data, *SimpleValue), which calls SimpleValue.unmarshalCBOR() under the hood
			{
				var v SimpleValue

				err := unmarshal(t, tc.data, &v)
				checkError[error](t, "Unmarshal", err, tc.wantErrorMsg)
			}
		})
	}
}

func testUnmarshalInvalidSimpleValueToEmptyInterface(t *testing.T, data []byte) {
	t.Helper()

	var v any
	err := unmarshal(t, data, v)
	checkError[*SyntaxError](t, "Unmarshal", err, "")
}

func testUnmarshalInvalidSimpleValue(t *testing.T, data []byte) {
	t.Helper()

	var v SimpleValue
	err := unmarshal(t, data, v)
	checkError[*SyntaxError](t, "Unmarshal", err, "")
}

func testUnmarshalSimpleValueToEmptyInterface(t *testing.T, data []byte, want any) {
	t.Helper()

	var v any
	err := unmarshal(t, data, &v)
	checkUnmarshal(t, "Unmarshal", v, err, want)
}

func testUnmarshalSimpleValue(t *testing.T, data []byte, want SimpleValue) {
	t.Helper()

	cborNil := isCBORNil(data)

	// Decode to SimpleValue
	var v SimpleValue
	err := unmarshal(t, data, &v)
	checkUnmarshal(t, "Unmarshal", v, err, want)

	// Decode to uninitialized *SimpleValue
	var pv *SimpleValue
	err = unmarshal(t, data, &pv)
	if cborNil {
		checkUnmarshal(t, "Unmarshal", pv, err, (*SimpleValue)(nil))
	} else {
		checkUnmarshal(t, "Unmarshal", derefOrNil(pv), err, want)
	}

	// Decode to initialized *SimpleValue
	v = SimpleValue(0)
	pv = &v
	err = unmarshal(t, data, &pv)
	if cborNil {
		checkUnmarshal(t, "Unmarshal", pv, err, (*SimpleValue)(nil))
	} else {
		checkUnmarshal(t, "Unmarshal", v, err, want)
	}
}

func TestMarshalSimpleValue(t *testing.T) {
	t.Run("0..23", func(t *testing.T) {
		for i := range 24 {
			wantData := []byte{byte(cborTypePrimitives) | byte(i)}
			v := SimpleValue(i)
			t.Run(fmt.Sprintf("SimpleValue(%d)", i), func(t *testing.T) {
				data, err := Marshal(v)
				checkMarshal(t, "Marshal", data, err, wantData)
			})
		}
	})

	t.Run("24..31", func(t *testing.T) {
		for i := 24; i <= 31; i++ {
			v := SimpleValue(i)
			t.Run(fmt.Sprintf("SimpleValue(%d)", i), func(t *testing.T) {
				data, err := Marshal(v)
				checkMarshalErr[*UnsupportedValueError](t, "Marshal", data, err, "")
			})
		}
	})

	t.Run("32..255", func(t *testing.T) {
		for i := 32; i <= 255; i++ {
			wantData := []byte{byte(cborTypePrimitives) | byte(24), byte(i)}
			v := SimpleValue(i)
			t.Run(fmt.Sprintf("SimpleValue(%d)", i), func(t *testing.T) {
				data, err := Marshal(v)
				checkMarshal(t, "Marshal", data, err, wantData)
			})
		}
	})
}
