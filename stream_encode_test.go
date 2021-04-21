// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"math/big"
	"testing"
)

func TestStreamEncodeArray(t *testing.T) {
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
