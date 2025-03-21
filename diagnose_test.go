// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestDiagnosticNotationExamples(t *testing.T) {
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
			str, err := Diagnose(tc.cbor)
			if err != nil {
				t.Errorf("Diagnostic(0x%x) returned error %q", tc.cbor, err)
			} else if str != tc.diag {
				t.Errorf("Diagnostic(0x%x) returned `%s`, want `%s`", tc.cbor, str, tc.diag)
			}

			str, rest, err := DiagnoseFirst(tc.cbor)
			if err != nil {
				t.Errorf("Diagnostic(0x%x) returned error %q", tc.cbor, err)
			} else if str != tc.diag {
				t.Errorf("Diagnostic(0x%x) returned `%s`, want `%s`", tc.cbor, str, tc.diag)
			}

			if rest == nil {
				t.Errorf("Diagnostic(0x%x) returned nil rest", tc.cbor)
			} else if len(rest) != 0 {
				t.Errorf("Diagnostic(0x%x) returned non-empty rest '%x'", tc.cbor, rest)
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
				ByteStringEncoding: ByteStringBase16Encoding,
			},
		},
		{
			"base32",
			hexDecode("4412345678"),
			`b32'CI2FM6A'`,
			&DiagOptions{
				ByteStringEncoding: ByteStringBase32Encoding,
			},
		},
		{
			"base32hex",
			hexDecode("4412345678"),
			`h32'28Q5CU0'`,
			&DiagOptions{
				ByteStringEncoding: ByteStringBase32HexEncoding,
			},
		},
		{
			"base64",
			hexDecode("4412345678"),
			`b64'EjRWeA'`,
			&DiagOptions{
				ByteStringEncoding: ByteStringBase64Encoding,
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
			"without ByteStringEmbeddedCBOR",
			hexDecode("4101"),
			`h'01'`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: false,
			},
		},
		{
			"with ByteStringEmbeddedCBOR",
			hexDecode("4101"),
			`<<1>>`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: true,
			},
		},
		{
			"multi CBOR items without ByteStringEmbeddedCBOR",
			hexDecode("420102"),
			`h'0102'`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: false,
			},
		},
		{
			"multi CBOR items with ByteStringEmbeddedCBOR",
			hexDecode("420102"),
			`<<1, 2>>`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: true,
			},
		},
		{
			"multi CBOR items with ByteStringEmbeddedCBOR",
			hexDecode("4563666F6FF6"),
			`h'63666f6ff6'`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: false,
			},
		},
		{
			"multi CBOR items with ByteStringEmbeddedCBOR",
			hexDecode("4563666F6FF6"),
			`<<"foo", null>>`,
			&DiagOptions{
				ByteStringEmbeddedCBOR: true,
			},
		},
		{
			"indefinite length byte string with no chunks",
			hexDecode("5fff"),
			`''_`,
			&DiagOptions{},
		},
		{
			"indefinite length byte string with a empty byte string",
			hexDecode("5f40ff"),
			`(_ h'')`, // RFC 8949, Section 8.1 says `(_ '')` but it looks wrong and conflicts with Appendix A.
			&DiagOptions{},
		},
		{
			"indefinite length byte string with two empty byte string",
			hexDecode("5f4040ff"),
			`(_ h'', h'')`,
			&DiagOptions{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			dm, err := tc.opts.DiagMode()
			if err != nil {
				t.Errorf("DiagMode() for 0x%x returned error %q", tc.cbor, err)
			}

			str, err := dm.Diagnose(tc.cbor)
			if err != nil {
				t.Errorf("Diagnose(0x%x) returned error %q", tc.cbor, err)
			} else if str != tc.diag {
				t.Errorf("Diagnose(0x%x) returned `%s`, want %s", tc.cbor, str, tc.diag)
			}
		})
	}
}

func TestDiagnoseTextString(t *testing.T) {
	testCases := []struct {
		title string
		cbor  []byte
		diag  string
		opts  *DiagOptions
	}{
		{
			"\t",
			hexDecode("6109"),
			`"\t"`,
			&DiagOptions{},
		},
		{
			"\r",
			hexDecode("610d"),
			`"\r"`,
			&DiagOptions{},
		},
		{
			"other ascii",
			hexDecode("611b"),
			`"\u001b"`,
			&DiagOptions{},
		},
		{
			"valid UTF-8 text in byte string",
			hexDecode("4d68656c6c6f2c20e4bda0e5a5bd"),
			`'hello, \u4f60\u597d'`,
			&DiagOptions{
				ByteStringText: true,
			},
		},
		{
			"valid UTF-8 text in text string",
			hexDecode("6d68656c6c6f2c20e4bda0e5a5bd"),
			`"hello, \u4f60\u597d"`, // "hello, 你好"
			&DiagOptions{
				ByteStringText: true,
			},
		},
		{
			"invalid UTF-8 text in byte string",
			hexDecode("4d68656c6c6fffeee4bda0e5a5bd"),
			`h'68656c6c6fffeee4bda0e5a5bd'`,
			&DiagOptions{
				ByteStringText: true,
			},
		},
		{
			"valid grapheme cluster text in byte string",
			hexDecode("583448656c6c6f2c2027e29da4efb88fe2808df09f94a5270ae4bda0e5a5bdefbc8c22f09fa791e2808df09fa49de2808df09fa79122"),
			`'Hello, \'\u2764\ufe0f\u200d\ud83d\udd25\'\n\u4f60\u597d\uff0c"\ud83e\uddd1\u200d\ud83e\udd1d\u200d\ud83e\uddd1"'`,
			&DiagOptions{
				ByteStringText: true,
			},
		},
		{
			"valid grapheme cluster text in text string",
			hexDecode("783448656c6c6f2c2027e29da4efb88fe2808df09f94a5270ae4bda0e5a5bdefbc8c22f09fa791e2808df09fa49de2808df09fa79122"),
			`"Hello, '\u2764\ufe0f\u200d\ud83d\udd25'\n\u4f60\u597d\uff0c\"\ud83e\uddd1\u200d\ud83e\udd1d\u200d\ud83e\uddd1\""`, // "Hello, '❤️‍🔥'\n你好，\"🧑‍🤝‍🧑\""
			&DiagOptions{
				ByteStringText: true,
			},
		},
		{
			"invalid grapheme cluster text in byte string",
			hexDecode("583448656c6c6feeff27e29da4efb88fe2808df09f94a5270de4bda0e5a5bdefbc8c22f09fa791e2808df09fa49de2808df09fa79122"),
			`h'48656c6c6feeff27e29da4efb88fe2808df09f94a5270de4bda0e5a5bdefbc8c22f09fa791e2808df09fa49de2808df09fa79122'`,
			&DiagOptions{
				ByteStringText: true,
			},
		},
		{
			"indefinite length text string with no chunks",
			hexDecode("7fff"),
			`""_`,
			&DiagOptions{},
		},
		{
			"indefinite length text string with a empty text string",
			hexDecode("7f60ff"),
			`(_ "")`,
			&DiagOptions{},
		},
		{
			"indefinite length text string with two empty text string",
			hexDecode("7f6060ff"),
			`(_ "", "")`,
			&DiagOptions{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			dm, err := tc.opts.DiagMode()
			if err != nil {
				t.Errorf("DiagMode() for 0x%x returned error %q", tc.cbor, err)
			}

			str, err := dm.Diagnose(tc.cbor)
			if err != nil {
				t.Errorf("Diagnose(0x%x) returned error %q", tc.cbor, err)
			} else if str != tc.diag {
				t.Errorf("Diagnose(0x%x) returned `%s`, want %s", tc.cbor, str, tc.diag)
			}
		})
	}
}

func TestDiagnoseInvalidTextString(t *testing.T) {
	testCases := []struct {
		title        string
		cbor         []byte
		wantErrorMsg string
		opts         *DiagOptions
	}{
		{
			"invalid UTF-8 text in text string",
			hexDecode("6d68656c6c6fffeee4bda0e5a5bd"),
			"invalid UTF-8 string",
			&DiagOptions{
				ByteStringText: true,
			},
		},
		{
			"invalid grapheme cluster text in text string",
			hexDecode("783448656c6c6feeff27e29da4efb88fe2808df09f94a5270de4bda0e5a5bdefbc8c22f09fa791e2808df09fa49de2808df09fa79122"),
			"invalid UTF-8 string",
			&DiagOptions{
				ByteStringText: true,
			},
		},
		{
			"invalid indefinite length text string",
			hexDecode("7f6040ff"),
			`wrong element type`,
			&DiagOptions{
				ByteStringText: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			dm, err := tc.opts.DiagMode()
			if err != nil {
				t.Errorf("DiagMode() for 0x%x returned error %q", tc.cbor, err)
			}

			_, err = dm.Diagnose(tc.cbor)
			if err == nil {
				t.Errorf("Diagnose(0x%x) didn't return error", tc.cbor)
			} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Diagnose(0x%x) returned error %q", tc.cbor, err)
			}
		})
	}
}

func TestDiagnoseFloatingPointNumber(t *testing.T) {
	testCases := []struct {
		title string
		cbor  []byte
		diag  string
		opts  *DiagOptions
	}{
		{
			"float16 without FloatPrecisionIndicator option",
			hexDecode("f93e00"),
			`1.5`,
			&DiagOptions{
				FloatPrecisionIndicator: false,
			},
		},
		{
			"float16 with FloatPrecisionIndicator option",
			hexDecode("f93e00"),
			`1.5_1`,
			&DiagOptions{
				FloatPrecisionIndicator: true,
			},
		},
		{
			"float32 without FloatPrecisionIndicator option",
			hexDecode("fa47c35000"),
			`100000.0`,
			&DiagOptions{
				FloatPrecisionIndicator: false,
			},
		},
		{
			"float32 with FloatPrecisionIndicator option",
			hexDecode("fa47c35000"),
			`100000.0_2`,
			&DiagOptions{
				FloatPrecisionIndicator: true,
			},
		},
		{
			"float64 without FloatPrecisionIndicator option",
			hexDecode("fbc010666666666666"),
			`-4.1`,
			&DiagOptions{
				FloatPrecisionIndicator: false,
			},
		},
		{
			"float64 with FloatPrecisionIndicator option",
			hexDecode("fbc010666666666666"),
			`-4.1_3`,
			&DiagOptions{
				FloatPrecisionIndicator: true,
			},
		},
		{
			"with FloatPrecisionIndicator option",
			hexDecode("c1fb41d452d9ec200000"),
			`1(1363896240.5_3)`,
			&DiagOptions{
				FloatPrecisionIndicator: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			dm, err := tc.opts.DiagMode()
			if err != nil {
				t.Errorf("DiagMode() for 0x%x returned error %q", tc.cbor, err)
			}

			str, err := dm.Diagnose(tc.cbor)
			if err != nil {
				t.Errorf("Diagnose(0x%x) returned error %q", tc.cbor, err)
			} else if str != tc.diag {
				t.Errorf("Diagnose(0x%x) returned `%s`, want %s", tc.cbor, str, tc.diag)
			}
		})
	}
}

func TestDiagnoseFirst(t *testing.T) {
	testCases := []struct {
		title        string
		cbor         []byte
		diag         string
		wantRest     []byte
		wantErrorMsg string
	}{
		{
			"with no trailing data",
			hexDecode("f93e00"),
			`1.5`,
			[]byte{},
			"",
		},
		{
			"with CBOR Sequences",
			hexDecode("f93e0064494554464401020304"),
			`1.5`,
			hexDecode("64494554464401020304"),
			"",
		},
		{
			"with invalid CBOR trailing data",
			hexDecode("f93e00ff494554464401020304"),
			`1.5`,
			hexDecode("ff494554464401020304"),
			"",
		},
		{
			"with invalid CBOR data",
			hexDecode("f93e"),
			``,
			nil,
			"unexpected EOF",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			str, rest, err := DiagnoseFirst(tc.cbor)
			if str != tc.diag {
				t.Errorf("DiagnoseFirst(0x%x) returned `%s`, want %s", tc.cbor, str, tc.diag)
			}

			if bytes.Equal(rest, tc.wantRest) == false {
				if str != tc.diag {
					t.Errorf("DiagnoseFirst(0x%x) returned rest `%x`, want rest %x", tc.cbor, rest, tc.wantRest)
				}
			}

			switch {
			case tc.wantErrorMsg == "" && err != nil:
				t.Errorf("DiagnoseFirst(0x%x) returned error %q", tc.cbor, err)
			case tc.wantErrorMsg != "" && err == nil:
				t.Errorf("DiagnoseFirst(0x%x) returned nil error, want error %q", tc.cbor, err)
			case tc.wantErrorMsg != "" && !strings.Contains(err.Error(), tc.wantErrorMsg):
				t.Errorf("DiagnoseFirst(0x%x) returned error %q, want error %q", tc.cbor, err, tc.wantErrorMsg)
			}
		})
	}
}

func TestDiagnoseCBORSequences(t *testing.T) {
	testCases := []struct {
		title       string
		cbor        []byte
		diag        string
		opts        *DiagOptions
		returnError bool
	}{
		{
			"CBOR Sequences without CBORSequence option",
			hexDecode("f93e0064494554464401020304"),
			``,
			&DiagOptions{
				CBORSequence: false,
			},
			true,
		},
		{
			"CBOR Sequences with CBORSequence option",
			hexDecode("f93e0064494554464401020304"),
			`1.5, "IETF", h'01020304'`,
			&DiagOptions{
				CBORSequence: true,
			},
			false,
		},
		{
			"CBOR Sequences with CBORSequence option",
			hexDecode("0102"),
			`1, 2`,
			&DiagOptions{
				CBORSequence: true,
			},
			false,
		},
		{
			"CBOR Sequences with CBORSequence option",
			hexDecode("63666F6FF6"),
			`"foo", null`,
			&DiagOptions{
				CBORSequence: true,
			},
			false,
		},
		{
			"partial/incomplete CBOR Sequences",
			hexDecode("f93e00644945544644010203"),
			`1.5, "IETF"`,
			&DiagOptions{
				CBORSequence: true,
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			dm, err := tc.opts.DiagMode()
			if err != nil {
				t.Errorf("DiagMode() for 0x%x returned error %q", tc.cbor, err)
			}

			str, err := dm.Diagnose(tc.cbor)
			if tc.returnError && err == nil {
				t.Errorf("Diagnose(0x%x) returned error %q", tc.cbor, err)
			} else if !tc.returnError && err != nil {
				t.Errorf("Diagnose(0x%x) returned error %q", tc.cbor, err)
			}

			if str != tc.diag {
				t.Errorf("Diagnose(0x%x) returned `%s`, want %s", tc.cbor, str, tc.diag)
			}
		})
	}
}

func TestDiagnoseTag(t *testing.T) {
	testCases := []struct {
		title       string
		cbor        []byte
		diag        string
		opts        *DiagOptions
		returnError bool
	}{
		{
			"CBOR tag number 2 with not well-formed encoded CBOR data item",
			hexDecode("c201"),
			``,
			&DiagOptions{},
			true,
		},
		{
			"CBOR tag number 3 with not well-formed encoded CBOR data item",
			hexDecode("c301"),
			``,
			&DiagOptions{},
			true,
		},
		{
			"CBOR tag number 2 with well-formed encoded CBOR data item",
			hexDecode("c240"),
			`0`,
			&DiagOptions{},
			false,
		},
		{
			"CBOR tag number 3 with well-formed encoded CBOR data item",
			hexDecode("c340"),
			`-1`, // -1 - n
			&DiagOptions{},
			false,
		},
		{
			"CBOR tag number 2 with well-formed encoded CBOR data item",
			hexDecode("c249010000000000000000"),
			`18446744073709551616`,
			&DiagOptions{},
			false,
		},
		{
			"CBOR tag number 3 with well-formed encoded CBOR data item",
			hexDecode("c349010000000000000000"),
			`-18446744073709551617`, // -1 - n
			&DiagOptions{},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			dm, err := tc.opts.DiagMode()
			if err != nil {
				t.Errorf("DiagMode() for 0x%x returned error %q", tc.cbor, err)
			}

			str, err := dm.Diagnose(tc.cbor)
			if tc.returnError && err == nil {
				t.Errorf("Diagnose(0x%x) returned error %q", tc.cbor, err)
			} else if !tc.returnError && err != nil {
				t.Errorf("Diagnose(0x%x) returned error %q", tc.cbor, err)
			}

			if str != tc.diag {
				t.Errorf("Diagnose(0x%x) returned `%s`, want %s", tc.cbor, str, tc.diag)
			}
		})
	}
}

func TestDiagnoseOptions(t *testing.T) {
	opts := DiagOptions{
		ByteStringEncoding:      ByteStringBase32Encoding,
		ByteStringHexWhitespace: true,
		ByteStringText:          false,
		ByteStringEmbeddedCBOR:  true,
		CBORSequence:            false,
		FloatPrecisionIndicator: true,
		MaxNestedLevels:         100,
		MaxArrayElements:        101,
		MaxMapPairs:             102,
	}
	dm, err := opts.DiagMode()
	if err != nil {
		t.Errorf("DiagMode() returned an error %v", err)
	}
	opts2 := dm.DiagOptions()
	if !reflect.DeepEqual(opts, opts2) {
		t.Errorf("DiagOptions() returned wrong options %v, want %v", opts2, opts)
	}

	opts = DiagOptions{
		ByteStringEncoding:      ByteStringBase64Encoding,
		ByteStringHexWhitespace: false,
		ByteStringText:          true,
		ByteStringEmbeddedCBOR:  false,
		CBORSequence:            true,
		FloatPrecisionIndicator: false,
		MaxNestedLevels:         100,
		MaxArrayElements:        101,
		MaxMapPairs:             102,
	}
	dm, err = opts.DiagMode()
	if err != nil {
		t.Errorf("DiagMode() returned an error %v", err)
	}
	opts2 = dm.DiagOptions()
	if !reflect.DeepEqual(opts, opts2) {
		t.Errorf("DiagOptions() returned wrong options %v, want %v", opts2, opts)
	}
}

func TestInvalidDiagnoseOptions(t *testing.T) {
	opts := &DiagOptions{
		ByteStringEncoding: ByteStringBase64Encoding + 1,
	}
	_, err := opts.DiagMode()
	if err == nil {
		t.Errorf("DiagMode() with invalid ByteStringEncoding option didn't return error")
	}
}

func TestDiagnoseExtraneousData(t *testing.T) {
	data := hexDecode("63666F6FF6")
	_, err := Diagnose(data)
	if err == nil {
		t.Errorf("Diagnose(0x%x) didn't return error", data)
	} else if !strings.Contains(err.Error(), `extraneous data`) {
		t.Errorf("Diagnose(0x%x) returned error %q", data, err)
	}

	_, _, err = DiagnoseFirst(data)
	if err != nil {
		t.Errorf("DiagnoseFirst(0x%x) returned error %v", data, err)
	}
}

func TestDiagnoseNotwellformedData(t *testing.T) {
	data := hexDecode("5f4060ff")
	_, err := Diagnose(data)
	if err == nil {
		t.Errorf("Diagnose(0x%x) didn't return error", data)
	} else if !strings.Contains(err.Error(), `wrong element type`) {
		t.Errorf("Diagnose(0x%x) returned error %q", data, err)
	}
}

func TestDiagnoseEmptyData(t *testing.T) {
	var emptyData []byte

	defaultMode, _ := DiagOptions{}.DiagMode()
	sequenceMode, _ := DiagOptions{CBORSequence: true}.DiagMode()

	testCases := []struct {
		name string
		dm   DiagMode
	}{
		{"default", defaultMode},
		{"sequence", sequenceMode},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := tc.dm.Diagnose(emptyData)
			if s != "" {
				t.Errorf("Diagnose() didn't return empty notation for empty data")
			}
			if err != io.EOF {
				t.Errorf("Diagnose() didn't return io.EOF for empty data")
			}

			s, rest, err := tc.dm.DiagnoseFirst(emptyData)
			if s != "" {
				t.Errorf("DiagnoseFirst() didn't return empty notation for empty data")
			}
			if len(rest) != 0 {
				t.Errorf("DiagnoseFirst() didn't return empty rest for empty data")
			}
			if err != io.EOF {
				t.Errorf("DiagnoseFirst() didn't return io.EOF for empty data")
			}
		})
	}
}

func BenchmarkDiagnose(b *testing.B) {
	for _, tc := range []struct {
		name  string
		opts  DiagOptions
		input []byte
	}{
		{
			name:  "escaped character in text string",
			opts:  DiagOptions{},
			input: hexDecode("62c3bc"), // "\u00fc"
		},
		{
			name:  "byte string base16 encoding",
			opts:  DiagOptions{ByteStringEncoding: ByteStringBase16Encoding},
			input: []byte("\x45hello"),
		},
		{
			name:  "byte string base32 encoding",
			opts:  DiagOptions{ByteStringEncoding: ByteStringBase32Encoding},
			input: []byte("\x45hello"),
		},
		{
			name:  "byte string base32hex encoding",
			opts:  DiagOptions{ByteStringEncoding: ByteStringBase32HexEncoding},
			input: []byte("\x45hello"),
		},
		{
			name:  "byte string base64url encoding",
			opts:  DiagOptions{ByteStringEncoding: ByteStringBase64Encoding},
			input: []byte("\x45hello"),
		},
	} {
		b.Run(tc.name, func(b *testing.B) {
			dm, err := tc.opts.DiagMode()
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = dm.Diagnose(tc.input)
			}
		})
	}
}
