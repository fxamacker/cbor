// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"io"
	"math/big"
)

// StreamEncoder provides low-level API for sequential encoding.
//
// Users of StreamEncoder should be familiar with CBOR data models.
// When in doubt, please use Encoder instead.
//
// IMPORTANT: Use Valid() to check whether generated CBOR data is
// complete and well-formed.
type StreamEncoder struct {
	*Encoder
}

// NewStreamEncoder returns a new StreamEncoder for sequential encoding.
func NewStreamEncoder(w io.Writer) *StreamEncoder {
	return &StreamEncoder{
		Encoder: &Encoder{
			w:      w,
			em:     defaultEncMode,
			e:      getEncoderBuffer(),
			stream: true,
		},
	}
}

// Flush writes streamed data to underlying io.Writer.
func (se *StreamEncoder) Flush() error {
	_, err := se.Encoder.e.WriteTo(se.Encoder.w)
	return err
}

// EncodeArrayHead encodes CBOR array head of specified size.
func (se *StreamEncoder) EncodeArrayHead(size uint64) error {
	encodeHead(se.Encoder.e, byte(cborTypeArray), size)
	return nil
}

// EncodeTagHead encodes CBOR tag head with num as tag number.
func (se *StreamEncoder) EncodeTagHead(num uint64) error {
	encodeHead(se.Encoder.e, byte(cborTypeTag), num)
	return nil
}

// EncodeRawBytes writes b to the underlying writer.
// If b is an empty or nil byte slice, it is a no-op.
func (se *StreamEncoder) EncodeRawBytes(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	se.e.Write(b)
	return nil
}

// EncodeNil encodes CBOR nil.
func (se *StreamEncoder) EncodeNil() error {
	se.e.Write(cborNil)
	return nil
}

// EncodeBool encodes bool as CBOR bool.
func (se *StreamEncoder) EncodeBool(b bool) error {
	bytes := cborTrue
	if !b {
		bytes = cborFalse
	}
	se.e.Write(bytes)
	return nil
}

// EncodeUint encodes uint as CBOR positive integer.
func (se *StreamEncoder) EncodeUint(i uint) error {
	encodeHead(se.e, byte(cborTypePositiveInt), uint64(i))
	return nil
}

// EncodeUint8 encodes uint8 as CBOR positive integer.
func (se *StreamEncoder) EncodeUint8(i uint8) error {
	encodeHead(se.e, byte(cborTypePositiveInt), uint64(i))
	return nil
}

// EncodeUint16 encodes uint16 as CBOR positive integer.
func (se *StreamEncoder) EncodeUint16(i uint16) error {
	encodeHead(se.e, byte(cborTypePositiveInt), uint64(i))
	return nil
}

// EncodeUint32 encodes uint32 as CBOR positive integer.
func (se *StreamEncoder) EncodeUint32(i uint32) error {
	encodeHead(se.e, byte(cborTypePositiveInt), uint64(i))
	return nil
}

// EncodeUint64 encodes uint64 as CBOR positive integer.
func (se *StreamEncoder) EncodeUint64(i uint64) error {
	encodeHead(se.e, byte(cborTypePositiveInt), i)
	return nil
}

// EncodeInt encodes int as CBOR positive or negtive integer.
func (se *StreamEncoder) EncodeInt(i int) error {
	return se.EncodeInt64(int64(i))
}

// EncodeInt8 encodes int8 as CBOR positive or negtive integer.
func (se *StreamEncoder) EncodeInt8(i int8) error {
	return se.EncodeInt64(int64(i))
}

// EncodeInt16 encodes int16 as CBOR positive or negtive integer.
func (se *StreamEncoder) EncodeInt16(i int16) error {
	return se.EncodeInt64(int64(i))
}

// EncodeInt32 encodes int32 as CBOR positive or negtive integer.
func (se *StreamEncoder) EncodeInt32(i int32) error {
	return se.EncodeInt64(int64(i))
}

// EncodeInt64 encodes int64 as CBOR positive or negtive integer.
func (se *StreamEncoder) EncodeInt64(i int64) error {
	t := cborTypePositiveInt
	if i < 0 {
		t = cborTypeNegativeInt
		i = i*(-1) - 1
	}
	encodeHead(se.e, byte(t), uint64(i))
	return nil
}

// EncodeBytes encodes byte slice as CBOR byte string.
func (se *StreamEncoder) EncodeBytes(b []byte) error {
	if b == nil {
		se.e.Write(cborNil)
		return nil
	}

	encodeHead(se.e, byte(cborTypeByteString), uint64(len(b)))
	se.e.Write(b)
	return nil
}

// EncodeString encodes string as CBOR string.
func (se *StreamEncoder) EncodeString(s string) error {
	encodeHead(se.e, byte(cborTypeTextString), uint64(len(s)))
	se.e.WriteString(s)
	return nil
}

var bigOne = big.NewInt(1)

// EncodeBigInt encodes big.Int as CBOR bignum (tag number 2 and 3).
func (se *StreamEncoder) EncodeBigInt(v *big.Int) error {
	if se.em.bigIntConvert == BigIntConvertShortest {
		if v.IsUint64() {
			encodeHead(se.e, byte(cborTypePositiveInt), v.Uint64())
			return nil
		}
		if v.IsInt64() {
			return se.EncodeInt64(v.Int64())
		}
	}

	tagNum := 2
	sign := v.Sign()
	if sign < 0 {
		tagNum = 3

		// Create a new big.Int with value of -1 - v
		v = new(big.Int).Abs(v)
		v.Sub(v, bigOne)
	}

	b := v.Bytes()

	// Write tag number
	encodeHead(se.e, byte(cborTypeTag), uint64(tagNum))
	// Write bignum byte string
	encodeHead(se.e, byte(cborTypeByteString), uint64(len(b)))
	se.e.Write(b)
	return nil
}
