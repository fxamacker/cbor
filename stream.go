// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

package cbor

import (
	"errors"
	"io"
	"reflect"
)

var cborIndefiniteHeader = map[cborType][]byte{
	cborTypeByteString: {0x5f},
	cborTypeTextString: {0x7f},
	cborTypeArray:      {0x9f},
	cborTypeMap:        {0xbf},
}

// Decoder reads and decodes CBOR values from an input stream.
type Decoder struct {
	r         io.Reader
	buf       []byte
	d         decodeState
	off       int // start of unread data in buf
	bytesRead int
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the next CBOR-encoded value from its input and stores it in
// the value pointed to by v.
func (dec *Decoder) Decode(v interface{}) (err error) {
	if len(dec.buf) == dec.off {
		n, err := dec.read()
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			return io.EOF
		}
	}

	dec.d.reset(dec.buf[dec.off:])
	err = dec.d.value(v)
	dec.off += dec.d.offset
	dec.bytesRead += dec.d.offset
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			// Need to read more data.
			n, err := dec.read()
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 {
				return io.ErrUnexpectedEOF
			}
			return dec.Decode(v)
		}
		return err
	}
	return nil
}

func (dec *Decoder) read() (int, error) {
	// Copy unread data over read data and reset off to 0.
	if dec.off > 0 {
		n := copy(dec.buf, dec.buf[dec.off:])
		dec.buf = dec.buf[:n]
		dec.off = 0
	}

	// Grow buf if needed.
	const minRead = 512
	if cap(dec.buf)-len(dec.buf) < minRead {
		newBuf := make([]byte, len(dec.buf), 2*cap(dec.buf)+minRead)
		copy(newBuf, dec.buf)
		dec.buf = newBuf
	}

	// Read from reader and reslice buf.
	n, err := dec.r.Read(dec.buf[len(dec.buf):cap(dec.buf)])
	dec.buf = dec.buf[0 : len(dec.buf)+n]
	return n, err
}

// NumBytesRead returns the number of bytes read.
func (dec *Decoder) NumBytesRead() int {
	return dec.bytesRead
}

// Encoder writes CBOR values to an output stream.
type Encoder struct {
	w               io.Writer
	opts            EncOptions
	e               encodeState
	indefiniteTypes []cborType
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer, encOpts EncOptions) *Encoder {
	return &Encoder{w: w, opts: encOpts, e: encodeState{}}
}

// Encode writes the CBOR encoding of v to the stream.
func (enc *Encoder) Encode(v interface{}) error {
	if len(enc.indefiniteTypes) > 0 && v != nil {
		indefiniteType := enc.indefiniteTypes[len(enc.indefiniteTypes)-1]
		if indefiniteType == cborTypeTextString {
			vKind := reflect.TypeOf(v).Kind()
			if vKind != reflect.String {
				return errors.New("cbor: cannot encode item type " + vKind.String() + " for indefinite-length text string")
			}
		} else if indefiniteType == cborTypeByteString {
			vType := reflect.TypeOf(v)
			vKind := vType.Kind()
			if (vKind != reflect.Array && vKind != reflect.Slice) || vType.Elem().Kind() != reflect.Uint8 {
				return errors.New("cbor: cannot encode item type " + vKind.String() + " for indefinite-length byte string")
			}
		}
	}

	if err := enc.e.marshal(v, enc.opts); err != nil {
		return err
	}
	if _, err := enc.e.WriteTo(enc.w); err != nil {
		return err
	}
	return nil
}

func (enc *Encoder) startIndefinite(typ cborType) error {
	header, ok := cborIndefiniteHeader[typ]
	if !ok {
		return errors.New("cbor: cannot encode indefinite length for " + typ.String())
	}
	enc.indefiniteTypes = append(enc.indefiniteTypes, typ)
	if _, err := enc.w.Write(header); err != nil {
		return err
	}
	return nil
}

// StartIndefiniteByteString starts byte string encoding of indefinite length.
// Subsequent calls of (*Encoder).Encode() encodes definite length byte strings
// ("chunks") as one continguous string until EndIndefinite is called.
func (enc *Encoder) StartIndefiniteByteString() error {
	return enc.startIndefinite(cborTypeByteString)
}

// StartIndefiniteTextString starts text string encoding of indefinite length.
// Subsequent calls of (*Encoder).Encode() encodes definite length text strings
// ("chunks") as one continguous string until EndIndefinite is called.
func (enc *Encoder) StartIndefiniteTextString() error {
	return enc.startIndefinite(cborTypeTextString)
}

// StartIndefiniteArray starts array encoding of indefinite length.
// Subsequent calls of (*Encoder).Encode() encodes elements of the array
// until EndIndefinite is called.
func (enc *Encoder) StartIndefiniteArray() error {
	return enc.startIndefinite(cborTypeArray)
}

// StartIndefiniteMap starts array encoding of indefinite length.
// Subsequent calls of (*Encoder).Encode() encodes elements of the map
// until EndIndefinite is called.
func (enc *Encoder) StartIndefiniteMap() error {
	return enc.startIndefinite(cborTypeMap)
}

// EndIndefinite closes last opened indefinite length value.
func (enc *Encoder) EndIndefinite() error {
	if len(enc.indefiniteTypes) == 0 {
		return errors.New("cbor: cannot encode \"break\" code outside indefinite length values")
	}
	enc.indefiniteTypes = enc.indefiniteTypes[:len(enc.indefiniteTypes)-1]
	if _, err := enc.w.Write([]byte{0xff}); err != nil {
		return err
	}
	return nil
}
