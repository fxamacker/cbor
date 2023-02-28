// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"errors"
	"math/big"
	"testing"
)

func TestStreamEncodeMap(t *testing.T) {
	t.Parallel()

	t.Run("default mode", func(t *testing.T) {
		expected := []byte{
			// map, 2 items follow
			0xa2,
			// UTF-8 string, length 5
			0x65,
			// h, e, l, l, o
			0x68, 0x65, 0x6c, 0x6c, 0x6f,
			// array, 1 items follow
			0x81,
			// 1
			0x01,
			// UTF-8 string, length 5
			0x65,
			// w, o, r, l, d
			0x77, 0x6f, 0x72, 0x6c, 0x64,
			// array, 1 items follow
			0x81,
			// -1
			0x20,
		}

		var buf bytes.Buffer
		se := NewStreamEncoder(&buf)
		defer se.Close()

		err := se.EncodeMapHead(2)
		if err != nil {
			t.Errorf("EncodeMapHead() returned error %v", err)
		}

		err = se.Encode("hello")
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.EncodeArrayHead(1)
		if err != nil {
			t.Errorf("EncodeArrayHead() returned error %v", err)
		}

		err = se.Encode(big.NewInt(1))
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.Encode("world")
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.EncodeArrayHead(1)
		if err != nil {
			t.Errorf("EncodeArrayHead() returned error %v", err)
		}

		err = se.Encode(big.NewInt(-1))
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.Flush()
		if err != nil {
			t.Errorf("Flush() returned error %v", err)
		}

		if !bytes.Equal(buf.Bytes(), expected) {
			t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), expected)
		}
	})

	t.Run("BigIntConvertNone mode", func(t *testing.T) {
		expected := []byte{
			// map, 2 items follow
			0xa2,
			// UTF-8 string, length 5
			0x65,
			// h, e, l, l, o
			0x68, 0x65, 0x6c, 0x6c, 0x6f,
			// array, 1 items follow
			0x81,
			// positive bignum (1)
			0xc2,
			// byte string, length 1
			0x41,
			0x01,
			// UTF-8 string, length 5
			0x65,
			// w, o, r, l, d
			0x77, 0x6f, 0x72, 0x6c, 0x64,
			// array, 1 items follow
			0x81,
			// negative bignum (-1)
			0xc3,
			// byte string, length 1
			0x40,
		}

		opts := EncOptions{BigIntConvert: BigIntConvertNone}
		em, err := opts.encMode()
		if err != nil {
			panic(err)
		}

		var buf bytes.Buffer
		se := em.NewStreamEncoder(&buf)
		defer se.Close()

		err = se.EncodeMapHead(2)
		if err != nil {
			t.Errorf("EncodeMapHead() returned error %v", err)
		}

		err = se.Encode("hello")
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.EncodeArrayHead(1)
		if err != nil {
			t.Errorf("EncodeArrayHead() returned error %v", err)
		}

		err = se.Encode(big.NewInt(1))
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.Encode("world")
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.EncodeArrayHead(1)
		if err != nil {
			t.Errorf("EncodeArrayHead() returned error %v", err)
		}

		err = se.Encode(big.NewInt(-1))
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.Flush()
		if err != nil {
			t.Errorf("Flush() returned error %v", err)
		}

		if !bytes.Equal(buf.Bytes(), expected) {
			t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), expected)
		}
	})
}

func TestStreamEncodeArray(t *testing.T) {
	t.Parallel()

	t.Run("default mode", func(t *testing.T) {
		expected := []byte{
			// array, 2 items follow
			0x82,
			// UTF-8 string, length 5
			0x65,
			// h, e, l, l, o
			0x68, 0x65, 0x6c, 0x6c, 0x6f,
			// array, 1 items follow
			0x81,
			// 1
			0x01,
		}

		var buf bytes.Buffer
		se := NewStreamEncoder(&buf)
		defer se.Close()

		err := se.EncodeArrayHead(2)
		if err != nil {
			t.Errorf("EncodeArrayHead() returned error %v", err)
		}

		err = se.Encode("hello")
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.EncodeArrayHead(1)
		if err != nil {
			t.Errorf("EncodeArrayHead() returned error %v", err)
		}

		err = se.Encode(big.NewInt(1))
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.Flush()
		if err != nil {
			t.Errorf("Flush() returned error %v", err)
		}

		if !bytes.Equal(buf.Bytes(), expected) {
			t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), expected)
		}
	})

	t.Run("BigIntConvertNone mode", func(t *testing.T) {
		expected := []byte{
			// array, 2 items follow
			0x82,
			// UTF-8 string, length 5
			0x65,
			// h, e, l, l, o
			0x68, 0x65, 0x6c, 0x6c, 0x6f,
			// array, 1 items follow
			0x81,
			// positive bignum
			0xc2,
			// byte string, length 1
			0x41,
			0x01,
		}

		opts := EncOptions{BigIntConvert: BigIntConvertNone}
		em, err := opts.encMode()
		if err != nil {
			panic(err)
		}

		var buf bytes.Buffer
		se := em.NewStreamEncoder(&buf)
		defer se.Close()

		err = se.EncodeArrayHead(2)
		if err != nil {
			t.Errorf("EncodeArrayHead() returned error %v", err)
		}

		err = se.Encode("hello")
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.EncodeArrayHead(1)
		if err != nil {
			t.Errorf("EncodeArrayHead() returned error %v", err)
		}

		err = se.Encode(big.NewInt(1))
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.Flush()
		if err != nil {
			t.Errorf("Flush() returned error %v", err)
		}

		if !bytes.Equal(buf.Bytes(), expected) {
			t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), expected)
		}
	})
}

func TestStreamEncodeTag(t *testing.T) {
	t.Parallel()

	t.Run("default mode", func(t *testing.T) {
		expected := []byte{
			// tag 128
			0xd8, 0x80,
			// array, 2 items follow
			0x82,
			// UTF-8 string, length 5
			0x65,
			// h, e, l, l, o
			0x68, 0x65, 0x6c, 0x6c, 0x6f,
			// 1
			0x01,
		}

		var buf bytes.Buffer
		se := NewStreamEncoder(&buf)
		defer se.Close()

		err := se.EncodeTagHead(128)
		if err != nil {
			t.Errorf("EncodeTagHead() returned error %v", err)
		}

		err = se.EncodeArrayHead(2)
		if err != nil {
			t.Errorf("EncodeArrayHead() returned error %v", err)
		}

		err = se.Encode("hello")
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.Encode(big.NewInt(1))
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.Flush()
		if err != nil {
			t.Errorf("Flush() returned error %v", err)
		}

		if !bytes.Equal(buf.Bytes(), expected) {
			t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), expected)
		}
	})

	t.Run("BigIntConvertNone", func(t *testing.T) {
		expected := []byte{
			// tag 128
			0xd8, 0x80,
			// array, 2 items follow
			0x82,
			// UTF-8 string, length 5
			0x65,
			// h, e, l, l, o
			0x68, 0x65, 0x6c, 0x6c, 0x6f,
			// positive bignum
			0xc2,
			// byte string, length 1
			0x41,
			0x01,
		}

		opts := EncOptions{BigIntConvert: BigIntConvertNone}
		em, err := opts.encMode()
		if err != nil {
			panic(err)
		}

		var buf bytes.Buffer
		se := em.NewStreamEncoder(&buf)
		defer se.Close()

		err = se.EncodeTagHead(128)
		if err != nil {
			t.Errorf("EncodeTagHead() returned error %v", err)
		}

		err = se.EncodeArrayHead(2)
		if err != nil {
			t.Errorf("EncodeArrayHead() returned error %v", err)
		}

		err = se.Encode("hello")
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.Encode(big.NewInt(1))
		if err != nil {
			t.Errorf("Encode() returned error %v", err)
		}

		err = se.Flush()
		if err != nil {
			t.Errorf("Flush() returned error %v", err)
		}

		if !bytes.Equal(buf.Bytes(), expected) {
			t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), expected)
		}
	})
}

func TestStreamEncodeNil(t *testing.T) {
	expected := []byte{0xf6}

	var buf bytes.Buffer
	se := NewStreamEncoder(&buf)
	defer se.Close()

	err := se.EncodeNil()
	if err != nil {
		t.Errorf("EncodeNil() returned error %v", err)
	}

	err = se.Flush()
	if err != nil {
		t.Errorf("Flush() returned error %v", err)
	}

	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), expected)
	}
}

func TestStreamEncodeRawBytes(t *testing.T) {
	testCases := []struct {
		name     string
		value    []byte
		expected []byte
	}{
		{"nil", nil, []byte{}},
		{"empty", []byte{}, []byte{}},
		{"not empty", []byte{0x01, 0x02, 0x03}, []byte{0x01, 0x02, 0x03}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeRawBytes(tc.value)
			if err != nil {
				t.Errorf("EncodeRawBytes() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeBool(t *testing.T) {
	testCases := []struct {
		name     string
		value    bool
		expected []byte
	}{
		{"false", false, []byte{0xf4}},
		{"true", true, []byte{0xf5}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeBool(tc.value)
			if err != nil {
				t.Errorf("EncodeBool() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeUint(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint
		expected []byte
	}{
		{"0", 0, []byte{0x00}},
		{"1", 1, []byte{0x01}},
		{"255", 255, []byte{0x18, 0xff}},
		{"65535", 65535, []byte{0x19, 0xff, 0xff}},
		{"4294967295", 4294967295, []byte{0x1a, 0xff, 0xff, 0xff, 0xff}},
		{"18446744073709551615", 18446744073709551615, []byte{0x1b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeUint(tc.value)
			if err != nil {
				t.Errorf("EncodeUint() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeUint8(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint8
		expected []byte
	}{
		{"0", 0, []byte{0x00}},
		{"1", 1, []byte{0x01}},
		{"100", 100, []byte{0x18, 0x64}},
		{"255", 255, []byte{0x18, 0xff}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeUint8(tc.value)
			if err != nil {
				t.Errorf("EncodeUint8() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeUint16(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint16
		expected []byte
	}{
		{"0", 0, []byte{0x00}},
		{"1", 1, []byte{0x01}},
		{"255", 255, []byte{0x18, 0xff}},
		{"65535", 65535, []byte{0x19, 0xff, 0xff}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeUint16(tc.value)
			if err != nil {
				t.Errorf("EncodeUint16() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeUint32(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint32
		expected []byte
	}{
		{"0", 0, []byte{0x00}},
		{"1", 1, []byte{0x01}},
		{"255", 255, []byte{0x18, 0xff}},
		{"65535", 65535, []byte{0x19, 0xff, 0xff}},
		{"4294967295", 4294967295, []byte{0x1a, 0xff, 0xff, 0xff, 0xff}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeUint32(tc.value)
			if err != nil {
				t.Errorf("EncodeUint32() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeUint64(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint64
		expected []byte
	}{
		{"0", 0, []byte{0x00}},
		{"1", 1, []byte{0x01}},
		{"255", 255, []byte{0x18, 0xff}},
		{"65535", 65535, []byte{0x19, 0xff, 0xff}},
		{"4294967295", 4294967295, []byte{0x1a, 0xff, 0xff, 0xff, 0xff}},
		{"18446744073709551615", 18446744073709551615, []byte{0x1b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeUint64(tc.value)
			if err != nil {
				t.Errorf("EncodeUint64() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeInt(t *testing.T) {
	testCases := []struct {
		name     string
		value    int
		expected []byte
	}{
		{"0", 0, []byte{0x00}},

		{"-1", -1, []byte{0x20}},
		{"1", 1, []byte{0x01}},

		{"-128", -128, []byte{0x38, 0x7f}},
		{"127", 127, []byte{0x18, 0x7f}},

		{"-32768", -32768, []byte{0x39, 0x7f, 0xff}},
		{"32767", 32767, []byte{0x19, 0x7f, 0xff}},

		{"-2147483648", -2147483648, []byte{0x3a, 0x7f, 0xff, 0xff, 0xff}},
		{"2147483647", 2147483647, []byte{0x1a, 0x7f, 0xff, 0xff, 0xff}},

		{"-9223372036854775808", -9223372036854775808, []byte{0x3b, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{"9223372036854775807", 9223372036854775807, []byte{0x1b, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeInt(tc.value)
			if err != nil {
				t.Errorf("EncodeInt() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeInt8(t *testing.T) {
	testCases := []struct {
		name     string
		value    int8
		expected []byte
	}{
		{"0", 0, []byte{0x00}},

		{"-1", -1, []byte{0x20}},
		{"1", 1, []byte{0x01}},

		{"-128", -128, []byte{0x38, 0x7f}},
		{"127", 127, []byte{0x18, 0x7f}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeInt8(tc.value)
			if err != nil {
				t.Errorf("EncodeInt8() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeInt16(t *testing.T) {
	testCases := []struct {
		name     string
		value    int16
		expected []byte
	}{
		{"0", 0, []byte{0x00}},

		{"-1", -1, []byte{0x20}},
		{"1", 1, []byte{0x01}},

		{"-128", -128, []byte{0x38, 0x7f}},
		{"127", 127, []byte{0x18, 0x7f}},

		{"-32768", -32768, []byte{0x39, 0x7f, 0xff}},
		{"32767", 32767, []byte{0x19, 0x7f, 0xff}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeInt16(tc.value)
			if err != nil {
				t.Errorf("EncodeInt16() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeInt32(t *testing.T) {
	testCases := []struct {
		name     string
		value    int32
		expected []byte
	}{
		{"0", 0, []byte{0x00}},

		{"-1", -1, []byte{0x20}},
		{"1", 1, []byte{0x01}},

		{"-128", -128, []byte{0x38, 0x7f}},
		{"127", 127, []byte{0x18, 0x7f}},

		{"-32768", -32768, []byte{0x39, 0x7f, 0xff}},
		{"32767", 32767, []byte{0x19, 0x7f, 0xff}},

		{"-2147483648", -2147483648, []byte{0x3a, 0x7f, 0xff, 0xff, 0xff}},
		{"2147483647", 2147483647, []byte{0x1a, 0x7f, 0xff, 0xff, 0xff}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeInt32(tc.value)
			if err != nil {
				t.Errorf("EncodeInt32() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeInt64(t *testing.T) {
	testCases := []struct {
		name     string
		value    int64
		expected []byte
	}{
		{"0", 0, []byte{0x00}},

		{"-1", -1, []byte{0x20}},
		{"1", 1, []byte{0x01}},

		{"-128", -128, []byte{0x38, 0x7f}},
		{"127", 127, []byte{0x18, 0x7f}},

		{"-32768", -32768, []byte{0x39, 0x7f, 0xff}},
		{"32767", 32767, []byte{0x19, 0x7f, 0xff}},

		{"-2147483648", -2147483648, []byte{0x3a, 0x7f, 0xff, 0xff, 0xff}},
		{"2147483647", 2147483647, []byte{0x1a, 0x7f, 0xff, 0xff, 0xff}},

		{"-9223372036854775808", -9223372036854775808, []byte{0x3b, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{"9223372036854775807", 9223372036854775807, []byte{0x1b, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeInt64(tc.value)
			if err != nil {
				t.Errorf("EncodeInt64() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeBigIntConvertNone(t *testing.T) {
	big1, _ := new(big.Int).SetString("18446744073709551616", 10)  // overflows uint64
	big2, _ := new(big.Int).SetString("-18446744073709551617", 10) // overflows int64

	testCases := []struct {
		name     string
		value    *big.Int
		expected []byte
	}{
		{"0", big.NewInt(0), []byte{0xc2, 0x40}},
		{"1", big.NewInt(1), []byte{0xc2, 0x41, 0x01}},
		{"-1", big.NewInt(-1), []byte{0xc3, 0x40}},
		{"18446744073709551616", big1, []byte{0xc2, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"-18446744073709551617", big2, []byte{0xc3, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dm, _ := EncOptions{BigIntConvert: BigIntConvertNone}.EncMode()

			var buf bytes.Buffer
			se := dm.NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeBigInt(tc.value)
			if err != nil {
				t.Errorf("EncodeBigInt() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeBigIntConvertShortest(t *testing.T) {
	big1, _ := new(big.Int).SetString("18446744073709551616", 10)  // overflows uint64
	big2, _ := new(big.Int).SetString("-18446744073709551617", 10) // overflows int64

	testCases := []struct {
		name     string
		value    *big.Int
		expected []byte
	}{
		{"0", big.NewInt(0), []byte{0x00}},
		{"1", big.NewInt(1), []byte{0x01}},
		{"-1", big.NewInt(-1), []byte{0x20}},
		{"18446744073709551616", big1, []byte{0xc2, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"-18446744073709551617", big2, []byte{0xc3, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dm, _ := EncOptions{BigIntConvert: BigIntConvertShortest}.EncMode()

			var buf bytes.Buffer
			se := dm.NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeBigInt(tc.value)
			if err != nil {
				t.Errorf("EncodeBigInt() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeBytes(t *testing.T) {
	testCases := []struct {
		name     string
		value    []byte
		expected []byte
	}{
		{"nil", nil, []byte{0xf6}},
		{"empty", []byte{}, []byte{0x40}},
		{"not empty", []byte{0x01, 0x02, 0x03, 0x04, 0x05}, []byte{0x45, 0x01, 0x02, 0x03, 0x04, 0x05}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeBytes(tc.value)
			if err != nil {
				t.Errorf("EncodeBytes() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeString(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		expected []byte
	}{
		{"empty", "", []byte{0x60}},
		{"not empty", "hello", []byte{0x65, 0x68, 0x65, 0x6c, 0x6c, 0x6f}},
	}

	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			se := NewStreamEncoder(&buf)
			defer se.Close()

			err := se.EncodeString(tc.value)
			if err != nil {
				t.Errorf("EncodeString() returned error %v", err)
			}

			err = se.Flush()
			if err != nil {
				t.Errorf("Flush() returned error %v", err)
			}

			if !bytes.Equal(buf.Bytes(), tc.expected) {
				t.Errorf("StreamEncoder encoded 0x%x, want 0x%x", buf.Bytes(), tc.expected)
			}
		})
	}
}

func TestStreamEncodeCloseError(t *testing.T) {
	var buf bytes.Buffer
	se := NewStreamEncoder(&buf)

	se.Close()

	var err error

	err = se.EncodeMapHead(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeMapHead() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeArrayHead(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeArrayHead() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeTagHead(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeTagHead() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeRawBytes(nil)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeRawBytes() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeNil()
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeNil() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeBool(true)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeBool() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeUint(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeUint() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeUint8(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeUint8() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeUint16(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeUint16() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeUint32(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeUint32() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeUint64(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeUint64() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeInt(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeInt() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeInt8(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeInt8() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeInt16(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeInt16() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeInt32(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeInt32() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeInt64(0)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeInt64() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeBytes(nil)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeBytes() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeString("")
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeString() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.EncodeBigInt(bigOne)
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("EncodeBigInt() returned error %v, want %v", err, ErrStreamClosed)
	}

	err = se.Flush()
	if !errors.Is(err, ErrStreamClosed) {
		t.Errorf("Flush() returned error %v, want %v", err, ErrStreamClosed)
	}

	se.Close() // Close on closed stream is no-op.
}
