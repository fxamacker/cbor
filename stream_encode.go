// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import "io"

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
	enc := defaultEncMode.NewEncoder(w)
	enc.stream = true
	return &StreamEncoder{enc}
}

// Flush writes streamed data to underlying io.Writer.
func (se *StreamEncoder) Flush() error {
	_, err := se.Encoder.e.WriteTo(se.Encoder.w)
	return err
}

// EncodeArrayHead encodes CBOR array head of specified size.
func (se *StreamEncoder) EncodeArrayHead(size uint64) {
	encodeHead(se.Encoder.e, byte(cborTypeArray), size)
}

// EncodeTagHead encodes CBOR tag head with num as tag number.
func (se *StreamEncoder) EncodeTagHead(num uint64) {
	encodeHead(se.Encoder.e, byte(cborTypeTag), num)
}
