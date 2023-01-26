// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"fmt"
	"strings"
	"testing"
)

func TestDiagnoseExamples(t *testing.T) {
	// https://www.rfc-editor.org/rfc/rfc8949.html#name-examples-of-encoded-cbor-da
	testCases := []struct {
		cbor []byte
		diag string
	}{
		{
			hexDecode("00"),
			`0`,
		},
		{
			hexDecode("01"),
			`1`,
		},
		{
			hexDecode("0a"),
			`10`,
		},
		{
			hexDecode("17"),
			`23`,
		},
		{
			hexDecode("1818"),
			`24`,
		},
		{
			hexDecode("1819"),
			`25`,
		},
		{
			hexDecode("1864"),
			`100`,
		},
		{
			hexDecode("1903e8"),
			`1000`,
		},
		{
			hexDecode("1a000f4240"),
			`1000000`,
		},
		{
			hexDecode("1b000000e8d4a51000"),
			`1000000000000`,
		},
		{
			hexDecode("1bffffffffffffffff"),
			`18446744073709551615`,
		},
		{
			hexDecode("c249010000000000000000"),
			`18446744073709551616`,
		},
		{
			hexDecode("3bffffffffffffffff"),
			`-18446744073709551616`,
		},
		{
			hexDecode("c349010000000000000000"),
			`-18446744073709551617`,
		},
		{
			hexDecode("20"),
			`-1`,
		},
		{
			hexDecode("29"),
			`-10`,
		},
		{
			hexDecode("3863"),
			`-100`,
		},
		{
			hexDecode("3903e7"),
			`-1000`,
		},
		{
			hexDecode("f90000"),
			`0.0`,
		},
		{
			hexDecode("f98000"),
			`-0.0`,
		},
		{
			hexDecode("f93c00"),
			`1.0`,
		},
		{
			hexDecode("fb3ff199999999999a"),
			`1.1`,
		},
		{
			hexDecode("f93e00"),
			`1.5`,
		},
		{
			hexDecode("f97bff"),
			`65504.0`,
		},
		{
			hexDecode("fa47c35000"),
			`100000.0`,
		},
		{
			hexDecode("fa7f7fffff"),
			`3.4028234663852886e+38`,
		},
		{
			hexDecode("fb7e37e43c8800759c"),
			`1.0e+300`,
		},
		{
			hexDecode("f90001"),
			`5.960464477539063e-8`,
		},
		{
			hexDecode("f90400"),
			`0.00006103515625`,
		},
		{
			hexDecode("f9c400"),
			`-4.0`,
		},
		{
			hexDecode("fbc010666666666666"),
			`-4.1`,
		},
		{
			hexDecode("f97c00"),
			`Infinity`,
		},
		{
			hexDecode("f97e00"),
			`NaN`,
		},
		{
			hexDecode("f9fc00"),
			`-Infinity`,
		},
		{
			hexDecode("fa7f800000"),
			`Infinity`,
		},
		{
			hexDecode("fa7fc00000"),
			`NaN`,
		},
		{
			hexDecode("faff800000"),
			`-Infinity`,
		},
		{
			hexDecode("fb7ff0000000000000"),
			`Infinity`,
		},
		{
			hexDecode("fb7ff8000000000000"),
			`NaN`,
		},
		{
			hexDecode("fbfff0000000000000"),
			`-Infinity`,
		},
		{
			hexDecode("f4"),
			`false`,
		},
		{
			hexDecode("f5"),
			`true`,
		},
		{
			hexDecode("f6"),
			`null`,
		},
		{
			hexDecode("f7"),
			`undefined`,
		},
		{
			hexDecode("f0"),
			`simple(16)`,
		},
		{
			hexDecode("f8ff"),
			`simple(255)`,
		},
		{
			hexDecode("c074323031332d30332d32315432303a30343a30305a"),
			`0("2013-03-21T20:04:00Z")`,
		},
		{
			hexDecode("c11a514b67b0"),
			`1(1363896240)`,
		},
		{
			hexDecode("c1fb41d452d9ec200000"),
			`1(1363896240.5)`,
		},
		{
			hexDecode("d74401020304"),
			`23(h'01020304')`,
		},
		{
			hexDecode("d818456449455446"),
			`24(h'6449455446')`,
		},
		{
			hexDecode("d82076687474703a2f2f7777772e6578616d706c652e636f6d"),
			`32("http://www.example.com")`,
		},
		{
			hexDecode("40"),
			`h''`,
		},
		{
			hexDecode("4401020304"),
			`h'01020304'`,
		},
		{
			hexDecode("60"),
			`""`,
		},
		{
			hexDecode("6161"),
			`"a"`,
		},
		{
			hexDecode("6449455446"),
			`"IETF"`,
		},
		{
			hexDecode("62225c"),
			`"\"\\"`,
		},
		{
			hexDecode("62c3bc"),
			`"\u00fc"`,
		},
		{
			hexDecode("63e6b0b4"),
			`"\u6c34"`,
		},
		{
			hexDecode("64f0908591"),
			`"\ud800\udd51"`,
		},
		{
			hexDecode("80"),
			`[]`,
		},
		{
			hexDecode("83010203"),
			`[1, 2, 3]`,
		},
		{
			hexDecode("8301820203820405"),
			`[1, [2, 3], [4, 5]]`,
		},
		{
			hexDecode("98190102030405060708090a0b0c0d0e0f101112131415161718181819"),
			`[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25]`,
		},
		{
			hexDecode("a0"),
			`{}`,
		},
		{
			hexDecode("a201020304"),
			`{1: 2, 3: 4}`,
		},
		{
			hexDecode("a26161016162820203"),
			`{"a": 1, "b": [2, 3]}`,
		},
		{
			hexDecode("826161a161626163"),
			`["a", {"b": "c"}]`,
		},
		{
			hexDecode("a56161614161626142616361436164614461656145"),
			`{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"}`,
		},
		{
			hexDecode("5f42010243030405ff"),
			`(_ h'0102', h'030405')`,
		},
		{
			hexDecode("7f657374726561646d696e67ff"),
			`(_ "strea", "ming")`,
		},
		{
			hexDecode("9fff"),
			`[_ ]`,
		},
		{
			hexDecode("9f018202039f0405ffff"),
			`[_ 1, [2, 3], [_ 4, 5]]`,
		},
		{
			hexDecode("9f01820203820405ff"),
			`[_ 1, [2, 3], [4, 5]]`,
		},
		{
			hexDecode("83018202039f0405ff"),
			`[1, [2, 3], [_ 4, 5]]`,
		},
		{
			hexDecode("83019f0203ff820405"),
			`[1, [_ 2, 3], [4, 5]]`,
		},
		{
			hexDecode("9f0102030405060708090a0b0c0d0e0f101112131415161718181819ff"),
			`[_ 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25]`,
		},
		{
			hexDecode("bf61610161629f0203ffff"),
			`{_ "a": 1, "b": [_ 2, 3]}`,
		},
		{
			hexDecode("826161bf61626163ff"),
			`["a", {_ "b": "c"}]`,
		},
		{
			hexDecode("bf6346756ef563416d7421ff"),
			`{_ "Fun": true, "Amt": -2}`,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Diagnostic %d", i), func(t *testing.T) {
			data, err := Diag(tc.cbor, nil)
			if err != nil {
				t.Errorf("Diag(0x%x) returned error %q", tc.cbor, err)
			} else if string(data) != tc.diag {
				t.Errorf("Diag(0x%x) returned `%s`, want `%s`", tc.cbor, string(data), tc.diag)
			}
		})
	}
}

func TestDiagnoseByteString(t *testing.T) {
	testCases := []struct {
		title string
		cbor  []byte
		diag  string
		opts  *DiagOptions
	}{
		{
			"base16",
			hexDecode("4412345678"),
			`h'12345678'`,
			&DiagOptions{
				ByteStringEncoding: "base16",
			},
		},
		{
			"base32",
			hexDecode("4412345678"),
			`b32'CI2FM6A'`,
			&DiagOptions{
				ByteStringEncoding: "base32",
			},
		},
		{
			"base32hex",
			hexDecode("4412345678"),
			`h32'28Q5CU0'`,
			&DiagOptions{
				ByteStringEncoding: "base32hex",
			},
		},
		{
			"base64",
			hexDecode("4412345678"),
			`b64'EjRWeA'`,
			&DiagOptions{
				ByteStringEncoding: "base64",
			},
		},
		{
			"without ByteStringHexWhitespace option",
			hexDecode("4b48656c6c6f20776f726c64"),
			`h'48656c6c6f20776f726c64'`,
			&DiagOptions{
				ByteStringHexWhitespace: false,
			},
		},
		{
			"with ByteStringHexWhitespace option",
			hexDecode("4b48656c6c6f20776f726c64"),
			`h'48 65 6c 6c 6f 20 77 6f 72 6c 64'`,
			&DiagOptions{
				ByteStringHexWhitespace: true,
			},
		},
		{
			"without ByteStringText option",
			hexDecode("4b68656c6c6f20776f726c64"),
			`h'68656c6c6f20776f726c64'`,
			&DiagOptions{
				ByteStringText: false,
			},
		},
		{
			"with ByteStringText option",
			hexDecode("4b68656c6c6f20776f726c64"),
			`'hello world'`,
			&DiagOptions{
				ByteStringText: true,
			},
		},
		{
			"without ByteStringText option and with ByteStringHexWhitespace option",
			hexDecode("4b68656c6c6f20776f726c64"),
			`h'68 65 6c 6c 6f 20 77 6f 72 6c 64'`,
			&DiagOptions{
				ByteStringText:          false,
				ByteStringHexWhitespace: true,
			},
		},
		{
			"without ByteStringEmbeddedCBOR and CBORSequence option",
			hexDecode("4101"),
			`h'01'`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: false,
				CBORSequence:           false,
			},
		},
		{
			"with ByteStringEmbeddedCBOR and CBORSequence option",
			hexDecode("4101"),
			`<<1>>`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: true,
				CBORSequence:           true,
			},
		},
		{
			"without ByteStringEmbeddedCBOR and CBORSequence option",
			hexDecode("420102"),
			`h'0102'`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: false,
				CBORSequence:           false,
			},
		},
		{
			"with ByteStringEmbeddedCBOR and CBORSequence option",
			hexDecode("420102"),
			`<<1, 2>>`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: true,
				CBORSequence:           true,
			},
		},
		{
			"with CBORSequence option",
			hexDecode("0102"),
			`1, 2`,
			&DiagOptions{
				CBORSequence: true,
			},
		},
		{
			"with ByteStringEmbeddedCBOR and CBORSequence option",
			hexDecode("4563666F6FF6"),
			`h'63666f6ff6'`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: false,
				CBORSequence:           false,
			},
		},
		{
			"with ByteStringEmbeddedCBOR and CBORSequence option",
			hexDecode("4563666F6FF6"),
			`<<"foo", null>>`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: true,
				CBORSequence:           true,
			},
		},
		{
			"with ByteStringEmbeddedCBOR and without CBORSequence option",
			hexDecode("4563666F6FF6"),
			`h'63666f6ff6'`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: true,
				CBORSequence:           false,
			},
		},
		{
			"with CBORSequence option",
			hexDecode("63666F6FF6"),
			`"foo", null`,
			&DiagOptions{
				CBORSequence: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {

			data, err := Diag(tc.cbor, tc.opts)
			if err != nil {
				t.Errorf("Diag(0x%x) returned error %q", tc.cbor, err)
			} else if string(data) != tc.diag {
				t.Errorf("Diag(0x%x) returned `%s`, want %s", tc.cbor, string(data), tc.diag)
			}
		})
	}

	t.Run("invalid encoding", func(t *testing.T) {
		cborData := hexDecode("4b48656c6c6f20776f726c64")
		_, err := Diag(cborData, &DiagOptions{
			ByteStringEncoding: "base58",
		})
		if err == nil {
			t.Errorf("Diag(0x%x) didn't return error", cborData)
		} else if !strings.Contains(err.Error(), `base58`) {
			t.Errorf("Diag(0x%x) returned error %q", cborData, err)
		}
	})

	t.Run("without CBORSequence option", func(t *testing.T) {
		cborData := hexDecode("63666F6FF6")
		_, err := Diag(cborData, nil)
		if err == nil {
			t.Errorf("Diag(0x%x) didn't return error", cborData)
		} else if !strings.Contains(err.Error(), `extraneous data`) {
			t.Errorf("Diag(0x%x) returned error %q", cborData, err)
		}
	})
}

func TestDiagnoseFloatingPointNumber(t *testing.T) {
	testCases := []struct {
		title string
		cbor  []byte
		diag  string
		opts  *DiagOptions
	}{
		{
			"float16 without IndicateFloatPrecision option",
			hexDecode("f93e00"),
			`1.5`,
			&DiagOptions{
				IndicateFloatPrecision: false,
			},
		},
		{
			"float16 with IndicateFloatPrecision option",
			hexDecode("f93e00"),
			`1.5_1`,
			&DiagOptions{
				IndicateFloatPrecision: true,
			},
		},
		{
			"float32 without IndicateFloatPrecision option",
			hexDecode("fa47c35000"),
			`100000.0`,
			&DiagOptions{
				IndicateFloatPrecision: false,
			},
		},
		{
			"float32 with IndicateFloatPrecision option",
			hexDecode("fa47c35000"),
			`100000.0_2`,
			&DiagOptions{
				IndicateFloatPrecision: true,
			},
		},
		{
			"float64 without IndicateFloatPrecision option",
			hexDecode("fbc010666666666666"),
			`-4.1`,
			&DiagOptions{
				IndicateFloatPrecision: false,
			},
		},
		{
			"float64 with IndicateFloatPrecision option",
			hexDecode("fbc010666666666666"),
			`-4.1_3`,
			&DiagOptions{
				IndicateFloatPrecision: true,
			},
		},
		{
			"with IndicateFloatPrecision option",
			hexDecode("c1fb41d452d9ec200000"),
			`1(1363896240.5_3)`,
			&DiagOptions{
				IndicateFloatPrecision: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {

			data, err := Diag(tc.cbor, tc.opts)
			if err != nil {
				t.Errorf("Diag(0x%x) returned error %q", tc.cbor, err)
			} else if string(data) != tc.diag {
				t.Errorf("Diag(0x%x) returned `%s`, want %s", tc.cbor, string(data), tc.diag)
			}
		})
	}
}
