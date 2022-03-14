// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"errors"
	"io"
	"math/big"
	"strings"
	"testing"
	"testing/iotest"
)

func TestStreamDecodeBool(t *testing.T) {

	expectedType := BoolType

	testCases := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{"false", []byte{0xf4}, false},
		{"true", []byte{0xf5}, true},
	}

	t.Parallel()

	for _, tc := range testCases {

		// For each test case, test 2 StreamDecoders.

		decoders := []struct {
			name string
			sd   *StreamDecoder
		}{
			{"byte_decoder", NewByteStreamDecoder(tc.data)},
			{"reader_decoder", NewStreamDecoder(bytes.NewReader(tc.data))},
		}

		for _, sd := range decoders {

			t.Run(sd.name+" "+tc.name, func(t *testing.T) {

				// NextType() peeks at next CBOR data type (data offset is not moved)
				nt, err := sd.sd.NextType()
				if err != nil {
					t.Errorf("NextType() returned error %v", err)
				}
				if nt != expectedType {
					t.Errorf("NextType() returned %s, want %s", nt, expectedType)
				}

				wantErrorMsg := "cbor: cannot decode CBOR boolean type to string"

				// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
				_, err = sd.sd.DecodeString()
				if err == nil {
					t.Errorf("DecodeString() didn't return error")
				} else if _, ok := err.(*WrongTypeError); !ok {
					t.Errorf("DecodeString() returned error %v (%T), want WrongTypeError", err, err)
				} else if err.Error() != wantErrorMsg {
					t.Errorf("DecodeString() returned error %q, want %q", err.Error(), wantErrorMsg)
				}

				// DecodeBool() should return boolean value (data offset is moved)
				v, err := sd.sd.DecodeBool()
				if err != nil {
					t.Errorf("DecodeBool() returned error %v", err)
				}
				if v != tc.expected {
					t.Errorf("DecodeBool() returned %v, want %v", v, tc.expected)
				}

				// NextType() should return io.EOF
				_, err = sd.sd.NextType()
				if err != io.EOF {
					t.Errorf("NextType() returned error %v, want io.EOF", err)
				}

				// DecodeBool() should return io.EOF
				_, err = sd.sd.DecodeBool()
				if err != io.EOF {
					t.Errorf("DecodeBool() returned error %v, want io.EOF", err)
				}
			})
		}
	}
}

func TestStreamDecodeNil(t *testing.T) {

	data := []byte{0xf6}

	expectedType := NilType

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	t.Parallel()

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// NextType() peeks at next CBOR data type (data offset is not moved)
			nt, err := sd.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != expectedType {
				t.Errorf("NextType() returned %s, want %s", nt, expectedType)
			}

			wantErrorMsg := "cbor: cannot decode CBOR nil type to bool"

			// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
			_, err = sd.sd.DecodeBool()
			if err == nil {
				t.Errorf("DecodeBool() didn't return error")
			} else if _, ok := err.(*WrongTypeError); !ok {
				t.Errorf("DecodeBool() returned error %v (%T), want WrongTypeError", err, err)
			} else if err.Error() != wantErrorMsg {
				t.Errorf("DecodeBool() returned error %q, want %q", err.Error(), wantErrorMsg)
			}

			// DecodeNil() should return no error (data offset is moved)
			err = sd.sd.DecodeNil()
			if err != nil {
				t.Errorf("DecodeNil() returned error %v", err)
			}

			// NextType() should return io.EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeNil() should return io.EOF
			err = sd.sd.DecodeNil()
			if err != io.EOF {
				t.Errorf("DecodeNil() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeUint(t *testing.T) {

	expectedType := UintType

	testCases := []struct {
		name     string
		data     []byte
		expected uint64
	}{
		{"0", []byte{0x00}, 0},
		{"1", []byte{0x01}, 1},
		{"255", []byte{0x18, 0xff}, 255},
		{"65535", []byte{0x19, 0xff, 0xff}, 65535},
		{"4294967295", []byte{0x1a, 0xff, 0xff, 0xff, 0xff}, 4294967295},
		{"18446744073709551615", []byte{0x1b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 18446744073709551615},
	}

	t.Parallel()

	for _, tc := range testCases {

		// For each test case, test 2 StreamDecoders.

		decoders := []struct {
			name string
			sd   *StreamDecoder
		}{
			{"byte_decoder", NewByteStreamDecoder(tc.data)},
			{"reader_decoder", NewStreamDecoder(bytes.NewReader(tc.data))},
		}

		for _, sd := range decoders {

			t.Run(sd.name+" "+tc.name, func(t *testing.T) {

				// NextType() peeks at next CBOR data type (data offset is not moved)
				nt, err := sd.sd.NextType()
				if err != nil {
					t.Errorf("NextType() returned error %v", err)
				}
				if nt != expectedType {
					t.Errorf("NextType() returned %s, want %s", nt, expectedType)
				}

				wantErrorMsg := "cbor: cannot decode CBOR uint type to bytes"

				// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
				_, err = sd.sd.DecodeBytes()
				if err == nil {
					t.Errorf("DecodeBytes() didn't return error")
				} else if _, ok := err.(*WrongTypeError); !ok {
					t.Errorf("DecodeBytes() returned error %v (%T), want WrongTypeError", err, err)
				} else if err.Error() != wantErrorMsg {
					t.Errorf("DecodeBytes() returned error %q, want %q", err.Error(), wantErrorMsg)
				}

				// DecodeUint64() should return uint64 value (data offset is moved)
				v, err := sd.sd.DecodeUint64()
				if err != nil {
					t.Errorf("DecodeUint64() returned error %v", err)
				}
				if v != tc.expected {
					t.Errorf("DecodeUint64() returned %v, want %v", v, tc.expected)
				}

				// NextType() should return io.EOF
				_, err = sd.sd.NextType()
				if err != io.EOF {
					t.Errorf("NextType() returned error %v, want io.EOF", err)
				}

				// DecodeUint64() should return io.EOF
				_, err = sd.sd.DecodeUint64()
				if err != io.EOF {
					t.Errorf("DecodeUint64() returned error %v, want io.EOF", err)
				}
			})
		}
	}
}

func TestStreamDecodeInt(t *testing.T) {

	testCases := []struct {
		name     string
		data     []byte
		expected int64
	}{
		{"0", []byte{0x00}, 0},

		{"-1", []byte{0x20}, -1},
		{"1", []byte{0x01}, 1},

		{"-128", []byte{0x38, 0x7f}, -128},
		{"127", []byte{0x18, 0x7f}, 127},

		{"-32768", []byte{0x39, 0x7f, 0xff}, -32768},
		{"32767", []byte{0x19, 0x7f, 0xff}, 32767},

		{"-2147483648", []byte{0x3a, 0x7f, 0xff, 0xff, 0xff}, -2147483648},
		{"2147483647", []byte{0x1a, 0x7f, 0xff, 0xff, 0xff}, 2147483647},

		{"-9223372036854775808", []byte{0x3b, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, -9223372036854775808},
		{"9223372036854775807", []byte{0x1b, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 9223372036854775807},
	}

	t.Parallel()

	for _, tc := range testCases {

		// For each test case, test 2 StreamDecoders.

		decoders := []struct {
			name string
			sd   *StreamDecoder
		}{
			{"byte_decoder", NewByteStreamDecoder(tc.data)},
			{"reader_decoder", NewStreamDecoder(bytes.NewReader(tc.data))},
		}

		for _, sd := range decoders {

			t.Run(sd.name+" "+tc.name, func(t *testing.T) {

				// NextType() peeks at next CBOR data type (data offset is not moved)
				nt, err := sd.sd.NextType()
				if err != nil {
					t.Errorf("NextType() returned error %v", err)
				}
				if nt != UintType && nt != IntType {
					t.Errorf("NextType() returned %s, want UintType or IntType", nt)
				}

				var wantErrorMsg string
				if tc.expected >= 0 {
					wantErrorMsg = "cbor: cannot decode CBOR uint type to big.Int"
				} else {
					wantErrorMsg = "cbor: cannot decode CBOR int type to big.Int"
				}

				// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
				_, err = sd.sd.DecodeBigInt()
				if err == nil {
					t.Errorf("DecodeBigInt() didn't return error")
				} else if _, ok := err.(*WrongTypeError); !ok {
					t.Errorf("DecodeBigInt() returned error %v (%T), want WrongTypeError", err, err)
				} else if err.Error() != wantErrorMsg {
					t.Errorf("DecodeBigInt() returned error %q, want %q", err.Error(), wantErrorMsg)
				}

				// DecodeInt64() should return int64 value (data offset is moved)
				v, err := sd.sd.DecodeInt64()
				if err != nil {
					t.Errorf("DecodeInt64() returned error %v", err)
				}
				if v != tc.expected {
					t.Errorf("DecodeInt64() returned %v, want %v", v, tc.expected)
				}

				// NextType() should return io.EOF
				_, err = sd.sd.NextType()
				if err != io.EOF {
					t.Errorf("NextType() returned error %v, want io.EOF", err)
				}

				// DecodeInt64() should return io.EOF
				_, err = sd.sd.DecodeInt64()
				if err != io.EOF {
					t.Errorf("DecodeInt64() returned error %v, want io.EOF", err)
				}
			})
		}
	}
}

func TestStreamDecodeIntOverflow(t *testing.T) {

	data := []byte{0x3b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	t.Parallel()

	// For each test case, test 2 StreamDecoders.

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// DecodeInt64() should return error (data offset is moved)
			_, err := sd.sd.DecodeInt64()
			if err == nil {
				t.Errorf("DecodeInt64() didn't return error")
			}

			// NextType() should return io.EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeInt64() should return io.EOF
			_, err = sd.sd.DecodeInt64()
			if err != io.EOF {
				t.Errorf("DecodeInt64() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeBytes(t *testing.T) {

	expectedType := ByteStringType

	testCases := []struct {
		name     string
		expected []byte
		data     []byte
	}{
		{"empty", []byte{}, []byte{0x40}},
		{"not empty", []byte{0x01, 0x02, 0x03, 0x04, 0x05}, []byte{0x45, 0x01, 0x02, 0x03, 0x04, 0x05}},
	}

	t.Parallel()

	for _, tc := range testCases {

		// For each test case, test 2 StreamDecoders.

		decoders := []struct {
			name string
			sd   *StreamDecoder
		}{
			{"byte_decoder", NewByteStreamDecoder(tc.data)},
			{"reader_decoder", NewStreamDecoder(bytes.NewReader(tc.data))},
		}

		for _, sd := range decoders {

			t.Run(sd.name+" "+tc.name, func(t *testing.T) {

				// NextType() peeks at next CBOR data type (data offset is not moved)
				nt, err := sd.sd.NextType()
				if err != nil {
					t.Errorf("NextType() returned error %v", err)
				}
				if nt != expectedType {
					t.Errorf("NextType() returned %s, want %s", nt, expectedType)
				}

				wantErrorMsg := "cbor: cannot decode CBOR byte string type to int64"

				// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
				i, err := sd.sd.DecodeInt64()
				if err == nil {
					t.Errorf("DecodeInt64() returned %v", i)
				} else if _, ok := err.(*WrongTypeError); !ok {
					t.Errorf("DecodeInt64() returned error %v (%T), want WrongTypeError", err, err)
				} else if err.Error() != wantErrorMsg {
					t.Errorf("DecodeInt64() returned error %q, want %q", err.Error(), wantErrorMsg)
				}

				// DecodeBytes() should return byte slice value (data offset is moved)
				v, err := sd.sd.DecodeBytes()
				if err != nil {
					t.Errorf("DecodeBytes() returned error %v", err)
				}
				if !bytes.Equal(v, tc.expected) {
					t.Errorf("DecodeBytes() returned %v, want %v", v, tc.expected)
				}

				// NextType() should return io.EOF
				_, err = sd.sd.NextType()
				if err != io.EOF {
					t.Errorf("NextType() returned error %v, want io.EOF", err)
				}

				// DecodeBytes() should return io.EOF
				_, err = sd.sd.DecodeBytes()
				if err != io.EOF {
					t.Errorf("DecodeBytes() returned error %v, want io.EOF", err)
				}
			})
		}
	}
}

func TestStreamDecodeIndefiniteLengthBytes(t *testing.T) {
	expectedType := ByteStringType

	data := []byte{0x5f, 0x42, 0x01, 0x02, 0x043, 0x03, 0x04, 0x05, 0xff}

	t.Parallel()

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// NextType() peeks at next CBOR data type (data offset is not moved)
			nt, err := sd.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != expectedType {
				t.Errorf("NextType() returned %s, want %s", nt, expectedType)
			}

			// DecodeBytes() should return error and byte string is skipped (data offset is moved)
			_, err = sd.sd.DecodeBytes()
			if err == nil {
				t.Errorf("DecodeBytes() didn't return error")
			}

			// NextType() should return io.EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeBytes() should return io.EOF
			_, err = sd.sd.DecodeBytes()
			if err != io.EOF {
				t.Errorf("DecodeBytes() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeString(t *testing.T) {

	expectedType := TextStringType

	testCases := []struct {
		name     string
		expected string
		data     []byte
	}{
		{"empty", "", []byte{0x60}},
		{"not empty", "hello", []byte{0x65, 0x68, 0x65, 0x6c, 0x6c, 0x6f}},
	}

	t.Parallel()

	for _, tc := range testCases {

		// For each test case, test 2 StreamDecoders.

		decoders := []struct {
			name string
			sd   *StreamDecoder
		}{
			{"byte_decoder", NewByteStreamDecoder(tc.data)},
			{"reader_decoder", NewStreamDecoder(bytes.NewReader(tc.data))},
		}

		for _, sd := range decoders {

			t.Run(sd.name+" "+tc.name, func(t *testing.T) {

				// NextType() peeks at next CBOR data type (data offset is not moved)
				nt, err := sd.sd.NextType()
				if err != nil {
					t.Errorf("NextType() returned error %v", err)
				}
				if nt != expectedType {
					t.Errorf("NextType() returned %s, want %s", nt, expectedType)
				}

				wantErrorMsg := "cbor: cannot decode CBOR text string type to bytes"

				// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
				i, err := sd.sd.DecodeBytes()
				if err == nil {
					t.Errorf("DecodeBytes() returned %v", i)
				} else if _, ok := err.(*WrongTypeError); !ok {
					t.Errorf("DecodeBytes() returned error %v (%T), want WrongTypeError", err, err)
				} else if err.Error() != wantErrorMsg {
					t.Errorf("DecodeBytes() returned error %q, want %q", err.Error(), wantErrorMsg)
				}

				// DecodeString() should return string value (data offset is moved)
				v, err := sd.sd.DecodeString()
				if err != nil {
					t.Errorf("DecodeString() returned error %v", err)
				}
				if v != tc.expected {
					t.Errorf("DecodeString() returned %v, want %v", v, tc.expected)
				}

				// NextType() should return io.EOF
				_, err = sd.sd.NextType()
				if err != io.EOF {
					t.Errorf("NextType() returned error %v, want io.EOF", err)
				}

				// DecodeString() should return io.EOF
				_, err = sd.sd.DecodeString()
				if err != io.EOF {
					t.Errorf("DecodeString() returned error %v, want io.EOF", err)
				}
			})
		}
	}
}

func TestStreamDecodeIndefiniteLengthString(t *testing.T) {
	expectedType := TextStringType

	data := []byte{0x7f, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x67, 0xff}

	t.Parallel()

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// NextType() peeks at next CBOR data type (data offset is not moved)
			nt, err := sd.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != expectedType {
				t.Errorf("NextType() returned %s, want %s", nt, expectedType)
			}

			// DecodeString() should return error and string is skipped (data offset is moved)
			_, err = sd.sd.DecodeString()
			if err == nil {
				t.Errorf("DecodeString() didn't return error")
			}

			// NextType() should return io.EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeString() should return io.EOF
			_, err = sd.sd.DecodeString()
			if err != io.EOF {
				t.Errorf("DecodeString() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeInvalidUTF8String(t *testing.T) {

	expectedType := TextStringType

	dmDecodeInvalidUTF8, err := DecOptions{UTF8: UTF8DecodeInvalid}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned an error %+v", err)
	}

	data := []byte{0x61, 0xfe}

	t.Parallel()

	testCases := []struct {
		name         string
		sd           *StreamDecoder
		wantObj      string
		wantErrorMsg string
	}{
		{
			name:         "byte decoder",
			sd:           NewByteStreamDecoder(data),
			wantErrorMsg: invalidUTF8ErrorMsg,
		},
		{
			name:    "byte decoder with UTF8DecodeInvalid",
			sd:      dmDecodeInvalidUTF8.NewByteStreamDecoder(data),
			wantObj: string([]byte{0xfe}),
		},
		{
			name:         "reader decoder",
			sd:           NewStreamDecoder(bytes.NewReader(data)),
			wantErrorMsg: invalidUTF8ErrorMsg,
		},
		{
			name:    "reader decoder with UTF8DecodeInvalid",
			sd:      dmDecodeInvalidUTF8.NewStreamDecoder(bytes.NewReader(data)),
			wantObj: string([]byte{0xfe}),
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			// NextType() peeks at next CBOR data type (data offset is not moved)
			nt, err := tc.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != expectedType {
				t.Errorf("NextType() returned %s, want %s", nt, expectedType)
			}

			// If decode option UTF8 is UTF8DecodeInvalid (default),
			// DecodeString() returns error and string is skipped (data offset is moved).
			// Otherwise, DecodeString() returns invalid UTF-8 string.
			s, err := tc.sd.DecodeString()

			if tc.wantErrorMsg != "" {
				if err == nil {
					t.Errorf("DecodeString() didn't return error")
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("DecodeString() returned error %q, want error %q", err.Error(), tc.wantErrorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("DecodeString() returned error %q", err)
				}
				if s != tc.wantObj {
					t.Errorf("DecodeString() returned %s, want %s", s, tc.wantObj)
				}
			}

			// NextType() should return io.EOF
			_, err = tc.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeString() should return io.EOF
			_, err = tc.sd.DecodeString()
			if err != io.EOF {
				t.Errorf("DecodeString() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeBigInt(t *testing.T) {

	expectedType := BigNumType

	big1, _ := new(big.Int).SetString("18446744073709551616", 10)  // overflows uint64
	big2, _ := new(big.Int).SetString("-18446744073709551617", 10) // overflows int64

	testCases := []struct {
		name     string
		expected *big.Int
		data     []byte
	}{
		{"0", big.NewInt(0), []byte{0xc2, 0x40}},
		{"1", big.NewInt(1), []byte{0xc2, 0x41, 0x01}},
		{"-1", big.NewInt(-1), []byte{0xc3, 0x40}},
		{"18446744073709551616", big1, []byte{0xc2, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"-18446744073709551617", big2, []byte{0xc3, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}

	t.Parallel()

	for _, tc := range testCases {

		// For each test case, test 2 StreamDecoders.

		decoders := []struct {
			name string
			sd   *StreamDecoder
		}{
			{"byte_decoder", NewByteStreamDecoder(tc.data)},
			{"reader_decoder", NewStreamDecoder(bytes.NewReader(tc.data))},
		}

		for _, sd := range decoders {

			t.Run(sd.name+" "+tc.name, func(t *testing.T) {

				// NextType() peeks at next CBOR data type (data offset is not moved)
				nt, err := sd.sd.NextType()
				if err != nil {
					t.Errorf("NextType() returned error %v", err)
				}
				if nt != expectedType {
					t.Errorf("NextType() returned %s, want %s", nt, expectedType)
				}

				wantErrorMsg := "cbor: cannot decode CBOR bignum type to string"

				// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
				i, err := sd.sd.DecodeString()
				if err == nil {
					t.Errorf("DecodeString() returned %v", i)
				} else if _, ok := err.(*WrongTypeError); !ok {
					t.Errorf("DecodeString() returned error %v (%T), want WrongTypeError", err, err)
				} else if err.Error() != wantErrorMsg {
					t.Errorf("DecodeString() returned error %q, want %q", err.Error(), wantErrorMsg)
				}

				// DecodeBigInt() should return *big.Int value (data offset is moved)
				v, err := sd.sd.DecodeBigInt()
				if err != nil {
					t.Errorf("DecodeBigInt() returned error %v", err)
				}
				if v.Cmp(tc.expected) != 0 {
					t.Errorf("DecodeBigInt() returned %v, want %v", v, tc.expected)
				}

				// NextType() should return io.EOF
				_, err = sd.sd.NextType()
				if err != io.EOF {
					t.Errorf("NextType() returned error %v, want io.EOF", err)
				}

				// DecodeBigInt() should return io.EOF
				_, err = sd.sd.DecodeBigInt()
				if err != io.EOF {
					t.Errorf("DecodeBigInt() returned error %v, want io.EOF", err)
				}
			})
		}
	}
}

func TestStreamDecodeTag(t *testing.T) {

	// 128("hello")
	data := []byte{
		// tag 128
		0xd8, 0x80,
		// UTF-8 string, length 5
		0x65,
		// h, e, l, l, o
		0x68, 0x65, 0x6c, 0x6c, 0x6f,
	}

	expectedTagNumber := uint64(128)
	expectedTagContent := "hello"

	t.Parallel()

	// For each test case, test 2 StreamDecoders.

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// NextType() peeks at next CBOR data type (data offset is not moved)
			nt, err := sd.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != TagType {
				t.Errorf("NextType() returned %s, want %s", nt, TagType)
			}

			wantErrorMsg := "cbor: cannot decode CBOR tag type to array"

			// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
			i, err := sd.sd.DecodeArrayHead()
			if err == nil {
				t.Errorf("DecodeArrayHead() returned %v", i)
			} else if _, ok := err.(*WrongTypeError); !ok {
				t.Errorf("DecodeArrayHead() returned error %v (%T), want WrongTypeError", err, err)
			} else if err.Error() != wantErrorMsg {
				t.Errorf("DecodeArrayHead() returned error %q, want %q", err.Error(), wantErrorMsg)
			}

			// DecodeTagNumber() should return uint64 value (data offset is moved)
			v, err := sd.sd.DecodeTagNumber()
			if err != nil {
				t.Errorf("DecodeTagNumber() returned error %v", err)
			}
			if v != expectedTagNumber {
				t.Errorf("DecodeTagNumber() returned %v, want %v", v, expectedTagNumber)
			}

			// NextType() should return string
			nt, err = sd.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != TextStringType {
				t.Errorf("NextType() returned %s, want %s", nt, TextStringType)
			}

			wantErrorMsg = "cbor: cannot decode CBOR text string type to array"

			// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
			i, err = sd.sd.DecodeArrayHead()
			if err == nil {
				t.Errorf("DecodeArrayHead() returned %v", i)
			} else if _, ok := err.(*WrongTypeError); !ok {
				t.Errorf("DecodeArrayHead() returned error %v (%T), want WrongTypeError", err, err)
			} else if err.Error() != wantErrorMsg {
				t.Errorf("DecodeArrayHead() returned error %q, want %q", err.Error(), wantErrorMsg)
			}

			// DecodeString() should return string value (data offset is moved)
			s, err := sd.sd.DecodeString()
			if err != nil {
				t.Errorf("DecodeString() returned error %v", err)
			}
			if s != expectedTagContent {
				t.Errorf("DecodeString() returned %v, want %v", v, expectedTagContent)
			}

			// NextType() should return io.EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeTagNumber() should return io.EOF
			_, err = sd.sd.DecodeTagNumber()
			if err != io.EOF {
				t.Errorf("DecodeTagNumber() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeEmptyArray(t *testing.T) {

	data := []byte{0x80}

	t.Parallel()

	// For each test case, test 2 StreamDecoders.

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// NextType() peeks at next CBOR data type (data offset is not moved)
			nt, err := sd.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != ArrayType {
				t.Errorf("NextType() returned %s, want %s", nt, ArrayType)
			}

			wantErrorMsg := "cbor: cannot decode CBOR array type to tag"

			// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
			i, err := sd.sd.DecodeTagNumber()
			if err == nil {
				t.Errorf("DecodeTagNumber() returned %v", i)
			} else if _, ok := err.(*WrongTypeError); !ok {
				t.Errorf("DecodeTagNumber() returned error %v (%T), want WrongTypeError", err, err)
			} else if err.Error() != wantErrorMsg {
				t.Errorf("DecodeTagNumber() returned error %q, want %q", err.Error(), wantErrorMsg)
			}

			// DecodeArrayHead() should return uint64 value (data offset is moved)
			v, err := sd.sd.DecodeArrayHead()
			if err != nil {
				t.Errorf("DecodeArrayHead() returned error %v", err)
			}
			if v != 0 {
				t.Errorf("DecodeArrayHead() returned %v, want %v", v, 0)
			}

			// NextType() should return io.EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeArrayHead() should return io.EOF
			_, err = sd.sd.DecodeArrayHead()
			if err != io.EOF {
				t.Errorf("DecodeArrayHead() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeArray(t *testing.T) {

	data := []byte{0x83, 0x01, 0x02, 0x03}

	t.Parallel()

	// For each test case, test 2 StreamDecoders.

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// NextType() peeks at next CBOR data type (data offset is not moved)
			nt, err := sd.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != ArrayType {
				t.Errorf("NextType() returned %s, want %s", nt, ArrayType)
			}

			wantErrorMsg := "cbor: cannot decode CBOR array type to map"

			// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
			i, err := sd.sd.DecodeMapHead()
			if err == nil {
				t.Errorf("DecodeMapHead() returned %v", i)
			} else if _, ok := err.(*WrongTypeError); !ok {
				t.Errorf("DecodeMapHead() returned error %v (%T), want WrongTypeError", err, err)
			} else if err.Error() != wantErrorMsg {
				t.Errorf("DecodeMapHead() returned error %q, want %q", err.Error(), wantErrorMsg)
			}

			// DecodeArrayHead() should return uint64 value (data offset is moved)
			v, err := sd.sd.DecodeArrayHead()
			if err != nil {
				t.Errorf("DecodeArrayHead() returned error %v", err)
			}
			if v != 3 {
				t.Errorf("DecodeArrayHead() returned %v, want %v", v, 3)
			}

			e, err := sd.sd.DecodeInt64()
			if err != nil {
				t.Errorf("DecodeInt64() returned error %v", err)
			}
			if e != 1 {
				t.Errorf("DecodeInt64() returned %v, want %v", e, 1)
			}

			e, err = sd.sd.DecodeInt64()
			if err != nil {
				t.Errorf("DecodeInt64() returned error %v", err)
			}
			if e != 2 {
				t.Errorf("DecodeInt64() returned %v, want %v", e, 2)
			}

			e, err = sd.sd.DecodeInt64()
			if err != nil {
				t.Errorf("DecodeInt64() returned error %v", err)
			}
			if e != 3 {
				t.Errorf("DecodeInt64() returned %v, want %v", e, 3)
			}

			// NextType() should return io.EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeArrayHead() should return io.EOF
			_, err = sd.sd.DecodeArrayHead()
			if err != io.EOF {
				t.Errorf("DecodeArrayHead() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeIndefiniteLengthArray(t *testing.T) {
	data := []byte{0x9f, 0x01, 0x02, 0x03, 0xff}

	t.Parallel()

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// NextType() peeks at next CBOR data type (data offset is not moved)
			nt, err := sd.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != ArrayType {
				t.Errorf("NextType() returned %s, want %s", nt, ArrayType)
			}

			// DecodeArrayHead() should return error and array is skipped (data offset is moved)
			_, err = sd.sd.DecodeArrayHead()
			if err == nil {
				t.Errorf("DecodeArrayHead() didn't return error")
			}

			// NextType() should return io.EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeArrayHead() should return io.EOF
			_, err = sd.sd.DecodeArrayHead()
			if err != io.EOF {
				t.Errorf("DecodeArrayHead() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeEmptyMap(t *testing.T) {

	data := []byte{0xa0}

	t.Parallel()

	// For each test case, test 2 StreamDecoders.

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// NextType() peeks at next CBOR data type (data offset is not moved)
			nt, err := sd.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != MapType {
				t.Errorf("NextType() returned %s, want %s", nt, MapType)
			}

			wantErrorMsg := "cbor: cannot decode CBOR map type to tag"

			// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
			i, err := sd.sd.DecodeTagNumber()
			if err == nil {
				t.Errorf("DecodeTagNumber() returned %v", i)
			} else if _, ok := err.(*WrongTypeError); !ok {
				t.Errorf("DecodeTagNumber() returned error %v (%T), want WrongTypeError", err, err)
			} else if err.Error() != wantErrorMsg {
				t.Errorf("DecodeTagNumber() returned error %q, want %q", err.Error(), wantErrorMsg)
			}

			// DecodeMapHead() should return uint64 value (data offset is moved)
			v, err := sd.sd.DecodeMapHead()
			if err != nil {
				t.Errorf("DecodeMapHead() returned error %v", err)
			}
			if v != 0 {
				t.Errorf("DecodeMapHead() returned %v, want %v", v, 0)
			}

			// NextType() should return io.EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeMapHead() should return io.EOF
			_, err = sd.sd.DecodeMapHead()
			if err != io.EOF {
				t.Errorf("DecodeMapHead() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeMap(t *testing.T) {

	data := []byte{0xa2, 0x01, 0x02, 0x03, 0x04}

	t.Parallel()

	// For each test case, test 2 StreamDecoders.

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// NextType() peeks at next CBOR data type (data offset is not moved)
			nt, err := sd.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != MapType {
				t.Errorf("NextType() returned %s, want %s", nt, MapType)
			}

			wantErrorMsg := "cbor: cannot decode CBOR map type to tag"

			// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
			i, err := sd.sd.DecodeTagNumber()
			if err == nil {
				t.Errorf("DecodeTagNumber() returned %v", i)
			} else if _, ok := err.(*WrongTypeError); !ok {
				t.Errorf("DecodeTagNumber() returned error %v (%T), want WrongTypeError", err, err)
			} else if err.Error() != wantErrorMsg {
				t.Errorf("DecodeTagNumber() returned error %q, want %q", err.Error(), wantErrorMsg)
			}

			// DecodeMapHead() should return uint64 value (data offset is moved)
			v, err := sd.sd.DecodeMapHead()
			if err != nil {
				t.Errorf("DecodeMapHead() returned error %v", err)
			}
			if v != 2 {
				t.Errorf("DecodeMapHead() returned %v, want %v", v, 3)
			}

			e, err := sd.sd.DecodeInt64()
			if err != nil {
				t.Errorf("DecodeInt64() returned error %v", err)
			}
			if e != 1 {
				t.Errorf("DecodeInt64() returned %v, want %v", e, 1)
			}

			e, err = sd.sd.DecodeInt64()
			if err != nil {
				t.Errorf("DecodeInt64() returned error %v", err)
			}
			if e != 2 {
				t.Errorf("DecodeInt64() returned %v, want %v", e, 2)
			}

			e, err = sd.sd.DecodeInt64()
			if err != nil {
				t.Errorf("DecodeInt64() returned error %v", err)
			}
			if e != 3 {
				t.Errorf("DecodeInt64() returned %v, want %v", e, 3)
			}

			e, err = sd.sd.DecodeInt64()
			if err != nil {
				t.Errorf("DecodeInt64() returned error %v", err)
			}
			if e != 4 {
				t.Errorf("DecodeInt64() returned %v, want %v", e, 4)
			}

			// NextType() should return io.EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeMapHead() should return io.EOF
			_, err = sd.sd.DecodeMapHead()
			if err != io.EOF {
				t.Errorf("DecodeMapHead() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeIndefiniteLengthMap(t *testing.T) {
	data := []byte{0xbf, 0x01, 0x02, 0x03, 0x04, 0xff}

	t.Parallel()

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// NextType() peeks at next CBOR data type (data offset is not moved)
			nt, err := sd.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != MapType {
				t.Errorf("NextType() returned %s, want %s", nt, MapType)
			}

			// DecodeMapHead() should return error and array is skipped (data offset is moved)
			_, err = sd.sd.DecodeMapHead()
			if err == nil {
				t.Errorf("DecodeMapHead() didn't return error")
			}

			// NextType() should return io.EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeMapHead() should return io.EOF
			_, err = sd.sd.DecodeMapHead()
			if err != io.EOF {
				t.Errorf("DecodeMapHead() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeRawBytes(t *testing.T) {

	data := []byte{0x83, 0x01, 0x02, 0x03}

	t.Parallel()

	// For each test case, test 2 StreamDecoders.

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// NextType() peeks at next CBOR data type (data offset is not moved)
			nt, err := sd.sd.NextType()
			if err != nil {
				t.Errorf("NextType() returned error %v", err)
			}
			if nt != ArrayType {
				t.Errorf("NextType() returned %s, want %s", nt, ArrayType)
			}

			wantErrorMsg := "cbor: cannot decode CBOR array type to nil"

			// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
			err = sd.sd.DecodeNil()
			if err == nil {
				t.Errorf("DecodeNil() returned no error")
			} else if _, ok := err.(*WrongTypeError); !ok {
				t.Errorf("DecodeNil() returned error %v (%T), want WrongTypeError", err, err)
			} else if err.Error() != wantErrorMsg {
				t.Errorf("DecodeNil() returned error %q, want %q", err.Error(), wantErrorMsg)
			}

			// DecodeRawBytes() should return byte slice value (data offset is moved)
			v, err := sd.sd.DecodeRawBytes()
			if err != nil {
				t.Errorf("DecodeRawBytes() returned error %v", err)
			}
			if !bytes.Equal(v, data) {
				t.Errorf("DecodeRawBytes() returned %v, want %v", v, data)
			}

			// NextType() should return io.EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}

			// DecodeRawBytes() should return io.EOF
			_, err = sd.sd.DecodeRawBytes()
			if err != io.EOF {
				t.Errorf("DecodeRawBytes() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeRawBytesZeroCopy(t *testing.T) {

	expectedType := ArrayType

	data := []byte{0x83, 0x01, 0x02, 0x03}

	t.Parallel()

	t.Run("byte_decoder", func(t *testing.T) {

		sd := NewByteStreamDecoder(data)

		// NextType() peeks at next CBOR data type (data offset is not moved)
		nt, err := sd.NextType()
		if err != nil {
			t.Errorf("NextType() returned error %v", err)
		}
		if nt != expectedType {
			t.Errorf("NextType() returned %s, want %s", nt, expectedType)
		}

		wantErrorMsg := "cbor: cannot decode CBOR array type to nil"

		// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
		err = sd.DecodeNil()
		if err == nil {
			t.Errorf("DecodeNil() returned no error")
		} else if _, ok := err.(*WrongTypeError); !ok {
			t.Errorf("DecodeNil() returned error %v (%T), want WrongTypeError", err, err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("DecodeNil() returned error %q, want %q", err.Error(), wantErrorMsg)
		}

		// DecodeRawBytesZeroCopy() should return byte slice (data offset is moved)
		v, err := sd.DecodeRawBytesZeroCopy()
		if err != nil {
			t.Errorf("DecodeRawBytesZeroCopy() returned error %v", err)
		}
		if !bytes.Equal(v, data) {
			t.Errorf("DecodeRawBytesZeroCopy() returned %v, want %v", v, data)
		}

		// NextType() should return io.EOF
		_, err = sd.NextType()
		if err != io.EOF {
			t.Errorf("NextType() returned error %v, want io.EOF", err)
		}

		// DecodeRawBytesZeroCopy() should return io.EOF
		_, err = sd.DecodeRawBytesZeroCopy()
		if err != io.EOF {
			t.Errorf("DecodeRawBytesZeroCopy() returned error %v, want io.EOF", err)
		}
	})

	t.Run("reader_decoder", func(t *testing.T) {
		sd := NewStreamDecoder(bytes.NewReader(data))

		// NextType() peeks at next CBOR data type (data offset is not moved)
		nt, err := sd.NextType()
		if err != nil {
			t.Errorf("NextType() returned error %v", err)
		}
		if nt != expectedType {
			t.Errorf("NextType() returned %s, want %s", nt, expectedType)
		}

		wantErrorMsg := "cbor: cannot decode CBOR array type to uint64"

		// DecodeXXX() should return WrongTypeError with type mismatch (data offset is not moved)
		_, err = sd.DecodeUint64()
		if err == nil {
			t.Errorf("DecodeUint64() returned no error")
		} else if _, ok := err.(*WrongTypeError); !ok {
			t.Errorf("DecodeUint64() returned error %v (%T), want WrongTypeError", err, err)
		} else if err.Error() != wantErrorMsg {
			t.Errorf("DecodeUint64() returned error %q, want %q", err.Error(), wantErrorMsg)
		}

		// DecodeRawBytesZeroCopy() should return error (data offset is not moved)
		_, err = sd.DecodeRawBytesZeroCopy()
		if err == nil {
			t.Errorf("DecodeRawBytesZeroCopy() didn't return error")
		}

		// DecodeRawBytes() should return []byte
		v, err := sd.DecodeRawBytes()
		if err != nil {
			t.Errorf("DecodeRawBytes() returned error %v", err)
		}
		if !bytes.Equal(v, data) {
			t.Errorf("DecodeRawBytes() returned %v, want %v", v, data)
		}
	})
}

func TestStreamDecodeSkip(t *testing.T) {

	data := []byte{
		0x18,                               // 24
		0x18, 0x44, 0x01, 0x02, 0x03, 0x04, // []byte{1, 2, 3, 4}
		0x83, 0x01, 0x02, 0x03, // [1, 2, 3]
		0xa2, 0x01, 0x02, 0x03, 0x04, // {1:2, 3:4}
	}

	t.Parallel()

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name+" ", func(t *testing.T) {

			// Skip uint
			err := sd.sd.Skip()
			if err != nil {
				t.Errorf("Skip() returned err %v", err)
			}

			// Skip byte string
			err = sd.sd.Skip()
			if err != nil {
				t.Errorf("Skip() returned err %v", err)
			}

			// Skip array
			err = sd.sd.Skip()
			if err != nil {
				t.Errorf("Skip() returned err %v", err)
			}

			// Skip map
			err = sd.sd.Skip()
			if err != nil {
				t.Errorf("Skip() returned err %v", err)
			}

			// EOF
			err = sd.sd.Skip()
			if err != io.EOF {
				t.Errorf("Skip() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeMultiData(t *testing.T) {

	data := []byte{
		0x18,                               // 24
		0x18, 0x44, 0x01, 0x02, 0x03, 0x04, // []byte{1, 2, 3, 4}
		0x83, 0x01, 0x02, 0x03, // [1, 2, 3]
		0xa2, 0x01, 0x02, 0x03, 0x04, // {1:2, 3:4}
	}

	t.Parallel()

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder", NewByteStreamDecoder(data)},
		{"reader_decoder", NewStreamDecoder(bytes.NewReader(data))},
	}

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// Decode uint
			i, err := sd.sd.DecodeInt64()
			if err != nil {
				t.Errorf("DecodeInt64() returned error %v", err)
			}
			if i != 24 {
				t.Errorf("DecodeInt64() returned %v, want %v", i, 24)
			}

			// Decode byte string
			b, err := sd.sd.DecodeBytes()
			if err != nil {
				t.Errorf("DecodeBytes() returned error %v", err)
			}
			if !bytes.Equal(b, []byte{1, 2, 3, 4}) {
				t.Errorf("DecodeBytes() returned %v, want %v", b, []byte{1, 2, 3, 4})
			}

			// Decode array
			arrayCount, err := sd.sd.DecodeArrayHead()
			if err != nil {
				t.Errorf("DecodeArrayHead() returned error %v", err)
			}
			if arrayCount != 3 {
				t.Errorf("DecodeArrayHead() returned %v, want %v", arrayCount, 3)
			}
			for i := uint64(0); i < arrayCount; i++ {
				var elem uint64
				elem, err = sd.sd.DecodeUint64()
				if err != nil {
					t.Errorf("DecodeUint64() returned error %v", err)
				}
				if elem != i+1 {
					t.Errorf("DecodeUint64() returned %v, want %v", elem, i+1)
				}
			}

			// Decode map
			mapCount, err := sd.sd.DecodeMapHead()
			if err != nil {
				t.Errorf("DecodeMapHead() returned error %v", err)
			}
			if mapCount != 2 {
				t.Errorf("DecodeMapHead() returned %v, want %v", mapCount, 2)
			}
			for i := uint64(0); i < mapCount; i++ {
				var k, v uint64
				k, err = sd.sd.DecodeUint64()
				if err != nil {
					t.Errorf("DecodeUint64() returned error %v", err)
				}
				if k != i*2+1 {
					t.Errorf("DecodeUint64() returned %v, want %v", k, i*2+1)
				}
				v, err = sd.sd.DecodeUint64()
				if err != nil {
					t.Errorf("DecodeUint64() returned error %v", err)
				}
				if v != i*2+2 {
					t.Errorf("DecodeUint64() returned %v, want %v", v, i*2+2)
				}
			}

			// EOF
			_, err = sd.sd.NextType()
			if err != io.EOF {
				t.Errorf("NextType() returned error %v, want io.EOF", err)
			}
		})
	}
}

func TestStreamDecodeMalformedData(t *testing.T) {
	testCases := []struct {
		name                 string
		data                 []byte
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
		// more
		{"End of input in a head", hexDecode("59"), "unexpected EOF", false},
		{"End of input in a head", hexDecode("5b"), "unexpected EOF", false},
		{"End of input in a head", hexDecode("d8"), "unexpected EOF", false},
		{"End of input in a head", hexDecode("d9"), "unexpected EOF", false},
	}

	t.Parallel()

	for _, tc := range testCases {

		// For each test case, test 2 StreamDecoders.

		decoders := []struct {
			name string
			sd   *StreamDecoder
		}{
			{"byte_decoder", NewByteStreamDecoder(tc.data)},
			{"reader_decoder", NewStreamDecoder(bytes.NewReader(tc.data))},
			{"onebytereader_decoder", NewStreamDecoder(iotest.OneByteReader(bytes.NewReader(tc.data)))},
		}

		for _, sd := range decoders {

			t.Run(sd.name+" "+tc.name, func(t *testing.T) {

				// DecodeXXX() and NextType() return the same error

				_, err := sd.sd.NextType()
				if err == nil {
					t.Errorf("NextType() didn't return an error")
				} else if !tc.errorMsgPartialMatch && err.Error() != tc.wantErrorMsg {
					t.Errorf("NextType() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
				} else if tc.errorMsgPartialMatch && !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("NextType() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
				}

				_, err = sd.sd.DecodeInt64()
				if err == nil {
					t.Errorf("DecodeInt64() didn't return an error")
				} else if !tc.errorMsgPartialMatch && err.Error() != tc.wantErrorMsg {
					t.Errorf("DecodeInt64() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
				} else if tc.errorMsgPartialMatch && !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("DecodeInt64() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
				}
			})
		}
	}
}

func TestStreamDecodeWithDecOptions(t *testing.T) {

	data := []byte{0x5f, 0x42, 0x01, 0x02, 0x043, 0x03, 0x04, 0x05, 0xff}

	expectedErrorMsg := "cbor: indefinite-length byte string isn't allowed"

	decMode, _ := DecOptions{IndefLength: IndefLengthForbidden}.DecMode()

	decoders := []struct {
		name string
		sd   *StreamDecoder
	}{
		{"byte_decoder_decopt", decMode.NewByteStreamDecoder(data)},
		{"reader_decoder_decopt", decMode.NewStreamDecoder(bytes.NewReader(data))},
	}

	t.Parallel()

	for _, sd := range decoders {

		t.Run(sd.name, func(t *testing.T) {

			// NextType() and DecodeXXX() return the same error.

			_, err := sd.sd.NextType()
			if err == nil {
				t.Errorf("NextType() didn't return error")
			}
			if err.Error() != expectedErrorMsg {
				t.Errorf("NextType()) returned error %q, want %q", err.Error(), expectedErrorMsg)
			}

			_, err = sd.sd.DecodeBytes()
			if err == nil {
				t.Errorf("DecodeBytes() didn't return error")
			}
			if err.Error() != expectedErrorMsg {
				t.Errorf("DecodeBytes()) returned error %q, want %q", err.Error(), expectedErrorMsg)
			}
		})
	}
}

type alwaysErrorReader struct{}

func (r *alwaysErrorReader) Read(p []byte) (int, error) {
	return 0, errors.New("reader error")
}

func TestStreamDecodeReaderError(t *testing.T) {
	expectedErrorMsg := "reader error"

	sd := NewStreamDecoder(&alwaysErrorReader{})

	// NextType() and DecodeXXX() return the same error.

	_, err := sd.NextType()
	if err == nil {
		t.Errorf("NextType() didn't return error")
	}
	if err.Error() != expectedErrorMsg {
		t.Errorf("NextType()) returned error %q, want %q", err.Error(), expectedErrorMsg)
	}

	_, err = sd.DecodeBytes()
	if err == nil {
		t.Errorf("DecodeBytes() didn't return error")
	}
	if err.Error() != expectedErrorMsg {
		t.Errorf("DecodeBytes()) returned error %q, want %q", err.Error(), expectedErrorMsg)
	}
}

func TestStreamDecoderNumBytesDecoded(t *testing.T) { //nolint:gocyclo

	t.Run("uint", func(t *testing.T) {
		data := []byte{0x18, 0x18}

		dec := NewByteStreamDecoder(data)

		_, err := dec.DecodeUint64()
		if err != nil {
			t.Errorf("DecodeUint64() returned error %v", err)
		}

		bytesDecoded := dec.NumBytesDecoded()
		if bytesDecoded != len(data) {
			t.Errorf("NumBytesDecoded() returned %v, want %v", bytesDecoded, len(data))
		}

		// Decode after EOD
		_, err = dec.DecodeUint64()
		if err == nil {
			t.Errorf("DecodeUint64() didn't return an error")
		}

		// Num of bytes decoded should remain the same
		bytesDecoded = dec.NumBytesDecoded()
		if bytesDecoded != len(data) {
			t.Errorf("NumBytesDecoded() returned %v after EOD, want %v", bytesDecoded, len(data))
		}
	})

	t.Run("byte slice", func(t *testing.T) {
		data := []byte{0x44, 0x01, 0x02, 0x03, 0x04}

		dec := NewByteStreamDecoder(data)

		_, err := dec.DecodeBytes()
		if err != nil {
			t.Errorf("DecodeBytes() returned error %v", err)
		}

		bytesDecoded := dec.NumBytesDecoded()
		if bytesDecoded != len(data) {
			t.Errorf("NumBytesDecoded() returned %v, want %v", bytesDecoded, len(data))
		}

		// Decode after EOD
		_, err = dec.DecodeUint64()
		if err == nil {
			t.Errorf("DecodeUint64() didn't return an error")
		}

		// Num of bytes decoded should remain the same
		bytesDecoded = dec.NumBytesDecoded()
		if bytesDecoded != len(data) {
			t.Errorf("NumBytesDecoded() returned %v after EOD, want %v", bytesDecoded, len(data))
		}
	})

	t.Run("string", func(t *testing.T) {

		data := []byte{0x63, 0xe6, 0xb0, 0xb4}

		dec := NewByteStreamDecoder(data)

		_, err := dec.DecodeString()
		if err != nil {
			t.Errorf("DecodeString() returned error %v", err)
		}

		bytesDecoded := dec.NumBytesDecoded()
		if bytesDecoded != len(data) {
			t.Errorf("NumBytesDecoded() returned %v, want %v", bytesDecoded, len(data))
		}

		// Decode after EOD
		_, err = dec.DecodeUint64()
		if err == nil {
			t.Errorf("DecodeUint64() didn't return an error")
		}

		// Num of bytes decoded should remain the same
		bytesDecoded = dec.NumBytesDecoded()
		if bytesDecoded != len(data) {
			t.Errorf("NumBytesDecoded() returned %v after EOD, want %v", bytesDecoded, len(data))
		}
	})

	t.Run("array", func(t *testing.T) {
		data := []byte{0x83, 0x01, 0x02, 0x03}

		dec := NewByteStreamDecoder(data)

		count, err := dec.DecodeArrayHead()
		if err != nil {
			t.Errorf("DecodeArrayHead() returned error %v", err)
		}

		bytesDecoded := dec.NumBytesDecoded()
		if bytesDecoded != 1 {
			t.Errorf("NumBytesDecoded() returned %v after decoding array head, want %v", bytesDecoded, 1)
		}

		for i := uint64(0); i < count; i++ {
			_, err = dec.DecodeUint64()
			if err != nil {
				t.Errorf("DecodeUint64() returned error %v", err)
			}

			bytesDecoded++

			num := dec.NumBytesDecoded()
			if num != bytesDecoded {
				t.Errorf("NumBytesDecoded() returned %v after decoding element %d, want %v", num, i, bytesDecoded)
			}
		}

		// Decode after EOD
		_, err = dec.DecodeUint64()
		if err == nil {
			t.Errorf("DecodeUint64() didn't return an error")
		}

		// Num of bytes decoded should remain the same
		num := dec.NumBytesDecoded()
		if num != bytesDecoded {
			t.Errorf("NumBytesDecoded() returned %v after EOD, want %v", num, bytesDecoded)
		}
	})

	t.Run("map", func(t *testing.T) {
		data := []byte{0xa2, 0x01, 0x02, 0x03, 0x04} // {1:2, 3:4}

		dec := NewByteStreamDecoder(data)

		count, err := dec.DecodeMapHead()
		if err != nil {
			t.Errorf("DecodeMapHead() returned error %v", err)
		}

		bytesDecoded := dec.NumBytesDecoded()
		if bytesDecoded != 1 {
			t.Errorf("NumBytesDecoded() returned %v after decoding map head, want %v", bytesDecoded, 1)
		}

		for i := uint64(0); i < count; i++ {
			_, err = dec.DecodeUint64()
			if err != nil {
				t.Errorf("DecodeUint64() returned error %v", err)
			}

			bytesDecoded++

			num := dec.NumBytesDecoded()
			if num != bytesDecoded {
				t.Errorf("NumBytesDecoded() returned %v after decoding map key %d, want %v", num, i, bytesDecoded)
			}

			_, err = dec.DecodeUint64()
			if err != nil {
				t.Errorf("DecodeUint64() returned error %v", err)
			}

			bytesDecoded++

			num = dec.NumBytesDecoded()
			if num != bytesDecoded {
				t.Errorf("NumBytesDecoded() returned %v after decoding map value %d, want %v", num, i, bytesDecoded)
			}
		}

		// Decode after EOD
		_, err = dec.DecodeUint64()
		if err == nil {
			t.Errorf("DecodeUint64() didn't return an error")
		}

		// Num of bytes decoded should remain the same
		num := dec.NumBytesDecoded()
		if num != bytesDecoded {
			t.Errorf("NumBytesDecoded() returned %v after EOD, want %v", num, bytesDecoded)
		}
	})

	t.Run("concatenated messages", func(t *testing.T) {

		data := []byte{
			0x18, 0x18, // uint64(24)
			0x63, 0xe6, 0xb0, 0xb4, // "水"
			0x83, 0x01, 0x02, 0x03, // [1, 2, 3]
			0xa2, 0x01, 0x02, 0x03, 0x04, // {1:2, 3:4}
		}

		const (
			msg1Length = 2
			msg2Length = 4
			msg3Length = 4
			msg4Length = 5
		)

		dec := NewByteStreamDecoder(data)

		// Decode uint64
		_, err := dec.DecodeUint64()
		if err != nil {
			t.Errorf("DecodeUint64() returned error %v", err)
		}

		bytesDecoded := dec.NumBytesDecoded()
		if bytesDecoded != msg1Length {
			t.Errorf("NumBytesDecoded() returned %v, want %v", bytesDecoded, msg1Length)
		}

		// Decode string
		_, err = dec.DecodeString()
		if err != nil {
			t.Errorf("DecodeString() returned error %v", err)
		}

		bytesDecoded = dec.NumBytesDecoded()
		if bytesDecoded != msg1Length+msg2Length {
			t.Errorf("NumBytesDecoded() returned %v, want %v", bytesDecoded, msg1Length+msg2Length)
		}

		// Decode array
		count, err := dec.DecodeArrayHead()
		if err != nil {
			t.Errorf("DecodeArrayHead() returned error %v", err)
		}
		for i := uint64(0); i < count; i++ {
			_, err = dec.DecodeUint64()
			if err != nil {
				t.Errorf("DecodeUint64() returned error %v", err)
			}
		}

		bytesDecoded = dec.NumBytesDecoded()
		if bytesDecoded != msg1Length+msg2Length+msg3Length {
			t.Errorf("NumBytesDecoded() returned %v, want %v", bytesDecoded, msg1Length+msg2Length+msg3Length)
		}

		// Decode map
		count, err = dec.DecodeMapHead()
		if err != nil {
			t.Errorf("DecodeMapHead() returned error %v", err)
		}
		for i := uint64(0); i < count; i++ {
			_, err = dec.DecodeUint64()
			if err != nil {
				t.Errorf("DecodeUint64() returned error %v", err)
			}
			_, err = dec.DecodeUint64()
			if err != nil {
				t.Errorf("DecodeUint64() returned error %v", err)
			}
		}

		bytesDecoded = dec.NumBytesDecoded()
		if bytesDecoded != msg1Length+msg2Length+msg3Length+msg4Length {
			t.Errorf("NumBytesDecoded() returned %v, want %v", bytesDecoded, msg1Length+msg2Length+msg3Length+msg4Length)
		}

		// Decode after EOD
		_, err = dec.NextType()
		if err == nil {
			t.Errorf("NextType() didn't return error")
		}

		// Number of bytes decoded should remain the same
		bytesDecoded = dec.NumBytesDecoded()
		if bytesDecoded != len(data) {
			t.Errorf("NumBytesDecoded() returned %v, want %v", bytesDecoded, len(data))
		}
	})
}

func TestStreamDecodeNextSize(t *testing.T) {
	testCases := []struct {
		name           string
		data           []byte
		expectedSize   uint64
		expectedErrMsg string
	}{
		{"no data", []byte{}, 0, "EOF"},
		{"invalid data", []byte{0x1b, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, 0, "unexpected EOF"},
		{"invalid data", []byte{0x9d}, 0, "cbor: invalid additional information"},
		{"false", []byte{0xf4}, 0, "size operation is not supported"},
		{"true", []byte{0xf5}, 0, "size operation is not supported"},
		{"nil", []byte{0xf6}, 0, "size operation is not supported"},
		{"uint", []byte{0x01}, 0, "size operation is not supported"},                              // 1
		{"int", []byte{0x20}, 0, "size operation is not supported"},                               // -1
		{"tag", []byte{0xc1, 0x1a, 0x51, 0x4b, 0x67, 0xb0}, 0, "size operation is not supported"}, // epoch-based date/time
		{"empty byte string", []byte{0x40}, 0, ""},                                                // []
		{"byte string", []byte{0x45, 0x01, 0x02, 0x03, 0x04, 0x05}, 5, ""},                        // [1, 2, 3, 4, 5]
		{"indef length byte string", []byte{0x5f, 0x42, 0x01, 0x02, 0x043, 0x03, 0x04, 0x05, 0xff}, 0, "size is unavailable for indefinite length"},
		{"empty text string", []byte{0x60}, 0, ""},                                                 // ""
		{"text string", []byte{0x65, 0x68, 0x65, 0x6c, 0x6c, 0x6f}, 5, ""},                         // "hello"
		{"text string", []byte{0x78, 0x01, 0x61}, 1, ""},                                           // "a"
		{"text string", []byte{0x79, 0x00, 0x01, 0x61}, 1, ""},                                     // "a"
		{"text string", []byte{0x7a, 0x00, 0x00, 0x00, 0x01, 0x61}, 1, ""},                         // "a"
		{"text string", []byte{0x7b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x61}, 1, ""}, // "a"
		{"indef length text string", []byte{0x7f, 0x65, 0x73, 0x74, 0x72, 0x65, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x67, 0xff}, 0, "size is unavailable for indefinite length"},
		{"empty array", []byte{0x80}, 0, ""},                                                                               // []
		{"array", []byte{0x83, 0x01, 0x02, 0x03}, 3, ""},                                                                   // [1, 2, 3]
		{"indef length array", []byte{0x9f, 0x01, 0x02, 0x03, 0xff}, 0, "size is unavailable for indefinite length"},       // [1, 2, 3]
		{"empty map", []byte{0xa0}, 0, ""},                                                                                 // []
		{"map", []byte{0xa2, 0x01, 0x02, 0x03, 0x04}, 2, ""},                                                               // [1, 2, 3]
		{"indef length map", []byte{0xbf, 0x01, 0x02, 0x03, 0x04, 0xff}, 0, "size is unavailable for indefinite length"},   // [1, 2, 3]
		{"big int 0", []byte{0xc2, 0x40}, 0, ""},                                                                           // bignum: 0
		{"big int 1", []byte{0xc2, 0x41, 0x01}, 1, ""},                                                                     // bignum: 1
		{"big int 18446744073709551616", []byte{0xc2, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 9, ""},  // bignum: 18446744073709551616
		{"big int -1", []byte{0xc3, 0x40}, 0, ""},                                                                          // bignum: -1
		{"big int -18446744073709551617", []byte{0xc3, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 9, ""}, // bignum: -18446744073709551617
	}

	for _, tc := range testCases {

		// For each test case, test 2 StreamDecoders.

		decoders := []struct {
			name string
			sd   *StreamDecoder
		}{
			{"byte_decoder", NewByteStreamDecoder(tc.data)},
			{"reader_decoder", NewStreamDecoder(bytes.NewReader(tc.data))},
		}

		for _, sd := range decoders {

			t.Run(sd.name+" "+tc.name, func(t *testing.T) {
				// NextSize() peeks at next CBOR data size
				size, err := sd.sd.NextSize()
				if size != tc.expectedSize {
					t.Errorf("NextSize() returned %d, want %d", size, tc.expectedSize)
				}
				if err == nil {
					if tc.expectedErrMsg != "" {
						t.Errorf("NextSize() didn't return error, want %s", tc.expectedErrMsg)
					}
				} else {
					if tc.expectedErrMsg == "" {
						t.Errorf("NextSize() returned error %v", err)
					} else if !strings.Contains(err.Error(), tc.expectedErrMsg) {
						t.Errorf("NextSize() returned error %v, want %s", err, tc.expectedErrMsg)
					}
				}
			})
		}
	}
}
