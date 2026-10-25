// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestDecoder(t *testing.T) {
	var buf bytes.Buffer
	for range 5 {
		for _, tc := range unmarshalTestCases {
			buf.Write(tc.data)
		}
	}

	testCases := []struct {
		name   string
		reader io.Reader
	}{
		{
			name:   "bytes.Buffer",
			reader: &buf,
		},
		{
			name:   "1 byte reader",
			reader: newNBytesReader(buf.Bytes(), 1),
		},
		{
			name:   "toggled reader",
			reader: newToggledReader(buf.Bytes(), 1),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := NewDecoder(tc.reader)
			bytesRead := 0
			for range 5 {
				for _, tc := range unmarshalTestCases {
					var v any
					err := decoder.Decode(&v)
					checkUnmarshalFunc(t, "Decoder.Decode", v, err, tc.wantInterfaceValue, nonFloatEqual)
					bytesRead += len(tc.data)
					if decoder.NumBytesRead() != bytesRead {
						t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), bytesRead)
					}
				}
			}
			for range 2 {
				// no more data
				testDecoderEOF(t, decoder)
			}
		})
	}
}

func TestDecoderUnmarshalTypeError(t *testing.T) {
	var buf bytes.Buffer
	for range 5 {
		for _, tc := range unmarshalTestCases {
			for range len(tc.wrongTypes) * 2 {
				buf.Write(tc.data)
			}
		}
	}

	testCases := []struct {
		name   string
		reader io.Reader
	}{
		{
			name:   "bytes.Buffer",
			reader: &buf,
		},
		{
			name:   "1 byte reader",
			reader: newNBytesReader(buf.Bytes(), 1),
		},
		{
			name:   "toggled reader",
			reader: newToggledReader(buf.Bytes(), 1),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := NewDecoder(tc.reader)
			bytesRead := 0
			for range 5 {
				for _, tc := range unmarshalTestCases {
					for _, typ := range tc.wrongTypes {
						v := reflect.New(typ)
						err := decoder.Decode(v.Interface())
						checkError[*UnmarshalTypeError](t, "Decoder.Decode", err, "")
						bytesRead += len(tc.data)
						if decoder.NumBytesRead() != bytesRead {
							t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), bytesRead)
						}

						var vi any
						err = decoder.Decode(&vi)
						checkUnmarshalFunc(t, "Decoder.Decode", vi, err, tc.wantInterfaceValue, nonFloatEqual)
						bytesRead += len(tc.data)
						if decoder.NumBytesRead() != bytesRead {
							t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), bytesRead)
						}
					}
				}
			}
			for range 2 {
				// no more data
				testDecoderEOF(t, decoder)
			}
		})
	}
}

func TestDecoderUnexpectedEOFError(t *testing.T) {
	var buf bytes.Buffer
	for _, tc := range unmarshalTestCases {
		buf.Write(tc.data)
	}
	buf.Truncate(buf.Len() - 1)

	testCases := []struct {
		name   string
		reader io.Reader
	}{
		{
			name:   "bytes.Buffer",
			reader: &buf,
		},
		{
			name:   "1 byte reader",
			reader: newNBytesReader(buf.Bytes(), 1),
		},
		{
			name:   "toggled reader",
			reader: newToggledReader(buf.Bytes(), 1),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			decoder := NewDecoder(tc.reader)
			bytesRead := 0
			for i := range len(unmarshalTestCases) - 1 {
				tc := unmarshalTestCases[i]
				var v any
				err := decoder.Decode(&v)
				checkUnmarshalFunc(t, "Decoder.Decode", v, err, tc.wantInterfaceValue, nonFloatEqual)
				bytesRead += len(tc.data)
				if decoder.NumBytesRead() != bytesRead {
					t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), bytesRead)
				}
			}
			for range 2 {
				// truncated data
				var v any
				err := decoder.Decode(&v)
				if v != nil {
					t.Errorf("Decoder.Decode = %v (%T), want nil", v, v)
				}
				if err != io.ErrUnexpectedEOF {
					t.Errorf("Decoder.Decode error = %v, want io.ErrUnexpectedEOF", err)
				}
			}
		})
	}
}

func TestDecoderReadError(t *testing.T) {
	var buf bytes.Buffer
	for _, tc := range unmarshalTestCases {
		buf.Write(tc.data)
	}
	buf.Truncate(buf.Len() - 1)

	readerErr := errors.New("reader error")

	testCases := []struct {
		name   string
		reader io.Reader
	}{
		{
			name:   "byte reader",
			reader: newNBytesReaderWithError(buf.Bytes(), 512, readerErr),
		},
		{
			name:   "1 byte reader",
			reader: newNBytesReaderWithError(buf.Bytes(), 1, readerErr),
		},
		{
			name:   "toggled reader",
			reader: newToggledReaderWithError(buf.Bytes(), 1, readerErr),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := NewDecoder(tc.reader)
			bytesRead := 0
			for i := range len(unmarshalTestCases) - 1 {
				tc := unmarshalTestCases[i]
				var v any
				err := decoder.Decode(&v)
				checkUnmarshalFunc(t, "Decoder.Decode", v, err, tc.wantInterfaceValue, nonFloatEqual)
				bytesRead += len(tc.data)
				if decoder.NumBytesRead() != bytesRead {
					t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), bytesRead)
				}
			}
			for range 2 {
				// truncated data because Reader returned error
				var v any
				err := decoder.Decode(&v)
				if v != nil {
					t.Errorf("Decoder.Decode = %v (%T), want nil", v, v)
				}
				if err != readerErr {
					t.Errorf("Decoder.Decode error = %v, want %v", err, readerErr)
				}
			}
		})
	}
}

func TestDecoderNoData(t *testing.T) {
	readerErr := errors.New("reader error")

	testCases := []struct {
		name    string
		reader  io.Reader
		wantErr error
	}{
		{
			name:    "bytes.Buffer",
			reader:  new(bytes.Buffer),
			wantErr: io.EOF,
		},
		{
			name:    "1 byte reader",
			reader:  newNBytesReaderWithError(nil, 0, readerErr),
			wantErr: readerErr,
		},
		{
			name:    "toggled reader",
			reader:  newToggledReaderWithError(nil, 0, readerErr),
			wantErr: readerErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := NewDecoder(tc.reader)
			for range 2 {
				var v any
				err := decoder.Decode(&v)
				if v != nil {
					t.Errorf("Decoder.Decode = %v (%T), want nil", v, v)
				}
				if err != tc.wantErr {
					t.Errorf("Decoder.Decode error = %v, want %v", err, tc.wantErr)
				}
			}
		})
	}
}

func TestDecoderRecoverableReadError(t *testing.T) {
	data := mustHexDecode("83010203") // [1,2,3]
	wantValue := []any{uint64(1), uint64(2), uint64(3)}
	recoverableReaderErr := errors.New("recoverable reader error")

	decoder := NewDecoder(newRecoverableReader(data, 1, recoverableReaderErr))

	var v any
	err := decoder.Decode(&v)
	if err != recoverableReaderErr {
		t.Fatalf("Decoder.Decode error = %v, want %v", err, recoverableReaderErr)
	}

	err = decoder.Decode(&v)
	checkUnmarshal(t, "Decoder.Decode", v, err, wantValue)
	if decoder.NumBytesRead() != len(data) {
		t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), len(data))
	}

	// no more data
	testDecoderEOF(t, decoder)
}

func TestDecoderInvalidData(t *testing.T) {
	data := []byte{0x01, 0x3e}
	decoder := NewDecoder(bytes.NewReader(data))

	var v1 any
	err := decoder.Decode(&v1)
	checkNoError(t, "Decoder.Decode", err)

	var v2 any
	err = decoder.Decode(&v2)
	checkError[error](t, "Decoder.Decode", err, "cbor: invalid additional information 30 for type negative integer")
}

func TestDecoderSkip(t *testing.T) {
	var buf bytes.Buffer
	for range 5 {
		for _, tc := range unmarshalTestCases {
			buf.Write(tc.data)
		}
	}

	testCases := []struct {
		name   string
		reader io.Reader
	}{
		{
			name:   "bytes.Buffer",
			reader: &buf,
		},
		{
			name:   "1 byte reader",
			reader: newNBytesReader(buf.Bytes(), 1),
		},
		{
			name:   "toggled reader",
			reader: newToggledReader(buf.Bytes(), 1),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := NewDecoder(tc.reader)
			bytesRead := 0
			for range 5 {
				for _, tc := range unmarshalTestCases {
					err := decoder.Skip()
					checkNoError(t, "Decoder.Skip", err)

					bytesRead += len(tc.data)
					if decoder.NumBytesRead() != bytesRead {
						t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), bytesRead)
					}
				}
			}
			for range 2 {
				// no more data
				err := decoder.Skip()
				if err != io.EOF {
					t.Errorf("Decoder.Skip error = %v, want io.EOF", err)
				}
			}
		})
	}
}

func TestDecoderSkipInvalidDataError(t *testing.T) {
	var buf bytes.Buffer
	for _, tc := range unmarshalTestCases {
		buf.Write(tc.data)
	}
	buf.WriteByte(0x3e)

	testCases := []struct {
		name   string
		reader io.Reader
	}{
		{
			name:   "bytes.Buffer",
			reader: &buf,
		},
		{
			name:   "1 byte reader",
			reader: newNBytesReader(buf.Bytes(), 1),
		},
		{
			name:   "toggled reader",
			reader: newToggledReader(buf.Bytes(), 1),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := NewDecoder(tc.reader)
			bytesRead := 0
			for _, tc := range unmarshalTestCases {
				err := decoder.Skip()
				checkNoError(t, "Decoder.Skip", err)

				bytesRead += len(tc.data)
				if decoder.NumBytesRead() != bytesRead {
					t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), bytesRead)
				}
			}
			for range 2 {
				// last data item is invalid
				err := decoder.Skip()
				if err == nil {
					t.Fatalf("Decoder.Skip: expected error containing %q, got nil", "cbor: invalid additional information")
				} else if !strings.Contains(err.Error(), "cbor: invalid additional information") {
					t.Errorf("Decoder.Skip: got %q, want substring %q", err.Error(), "cbor: invalid additional information")
				}
			}
		})
	}
}

func TestDecoderSkipUnexpectedEOFError(t *testing.T) {
	var buf bytes.Buffer
	for _, tc := range unmarshalTestCases {
		buf.Write(tc.data)
	}
	buf.Truncate(buf.Len() - 1)

	testCases := []struct {
		name   string
		reader io.Reader
	}{
		{
			name:   "bytes.Buffer",
			reader: &buf,
		},
		{
			name:   "1 byte reader",
			reader: newNBytesReader(buf.Bytes(), 1),
		},
		{
			name:   "toggled reader",
			reader: newToggledReader(buf.Bytes(), 1),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := NewDecoder(tc.reader)
			bytesRead := 0
			for _, tc := range unmarshalTestCases[:len(unmarshalTestCases)-1] {
				err := decoder.Skip()
				checkNoError(t, "Decoder.Skip", err)

				bytesRead += len(tc.data)
				if decoder.NumBytesRead() != bytesRead {
					t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), bytesRead)
				}
			}
			for range 2 {
				// last data item is invalid
				err := decoder.Skip()
				if err != io.ErrUnexpectedEOF {
					t.Errorf("Decoder.Skip error = %v, want io.ErrUnexpectedEOF", err)
				}
			}
		})
	}
}

func TestDecoderSkipReadError(t *testing.T) {
	var buf bytes.Buffer
	for _, tc := range unmarshalTestCases {
		buf.Write(tc.data)
	}
	buf.Truncate(buf.Len() - 1)

	readerErr := errors.New("reader error")

	testCases := []struct {
		name   string
		reader io.Reader
	}{
		{
			name:   "byte reader",
			reader: newNBytesReaderWithError(buf.Bytes(), 512, readerErr),
		},
		{
			name:   "1 byte reader",
			reader: newNBytesReaderWithError(buf.Bytes(), 1, readerErr),
		},
		{
			name:   "toggled reader",
			reader: newToggledReaderWithError(buf.Bytes(), 1, readerErr),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := NewDecoder(tc.reader)
			bytesRead := 0
			for _, tc := range unmarshalTestCases[:len(unmarshalTestCases)-1] {
				err := decoder.Skip()
				checkNoError(t, "Decoder.Skip", err)

				bytesRead += len(tc.data)
				if decoder.NumBytesRead() != bytesRead {
					t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), bytesRead)
				}
			}
			for range 2 {
				// truncated data because Reader returned error
				err := decoder.Skip()
				if err != readerErr {
					t.Errorf("Decoder.Skip error = %v, want %v", err, readerErr)
				}
			}
		})
	}
}

func TestDecoderSkipNoData(t *testing.T) {
	readerErr := errors.New("reader error")

	testCases := []struct {
		name    string
		reader  io.Reader
		wantErr error
	}{
		{
			name:    "bytes.Buffer",
			reader:  new(bytes.Buffer),
			wantErr: io.EOF,
		},
		{
			name:    "1 byte reader",
			reader:  newNBytesReaderWithError(nil, 0, readerErr),
			wantErr: readerErr,
		},
		{
			name:    "toggled reader",
			reader:  newToggledReaderWithError(nil, 0, readerErr),
			wantErr: readerErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := NewDecoder(tc.reader)
			for range 2 {
				err := decoder.Skip()
				if err != tc.wantErr {
					t.Errorf("Decoder.Skip error = %v, want %v", err, tc.wantErr)
				}
			}
		})
	}
}

func TestDecoderSkipRecoverableReadError(t *testing.T) {
	data := mustHexDecode("83010203") // [1,2,3]
	recoverableReaderErr := errors.New("recoverable reader error")

	decoder := NewDecoder(newRecoverableReader(data, 1, recoverableReaderErr))

	err := decoder.Skip()
	if err != recoverableReaderErr {
		t.Fatalf("Decoder.Skip error = %v, want %v", err, recoverableReaderErr)
	}

	err = decoder.Skip()
	checkNoError(t, "Decoder.Skip", err)
	if decoder.NumBytesRead() != len(data) {
		t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), len(data))
	}

	// no more data
	err = decoder.Skip()
	if err != io.EOF {
		t.Errorf("Decoder.Skip error = %v, want io.EOF", err)
	}
}

func TestDecoderStructTag(t *testing.T) {
	type strc struct {
		A string `json:"x" cbor:"a"`
		B string `json:"y" cbor:"b"`
		C string `json:"z"`
	}
	want := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	data := mustHexDecode("a36161614161626142617a6143") // {"a":"A", "b":"B", "z":"C"}

	var v strc
	dec := NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&v)
	checkUnmarshal(t, "Decoder.Decode", v, err, want)
}

func TestDecoderBuffered(t *testing.T) {
	testCases := []struct {
		name      string
		data      []byte
		buffered  []byte
		decodeErr error
	}{
		{
			name:      "empty",
			data:      []byte{},
			buffered:  []byte{},
			decodeErr: io.EOF,
		},
		{
			name:      "malformed CBOR data item",
			data:      []byte{0xc0},
			buffered:  []byte{0xc0},
			decodeErr: io.ErrUnexpectedEOF,
		},
		{
			name:     "1 CBOR data item",
			data:     []byte{0xc2, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			buffered: []byte{},
		},
		{
			name: "2 CBOR data items",
			data: []byte{
				// First CBOR data item
				0xc2, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				// Second CBOR data item
				0xc3, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			buffered: []byte{
				// Second CBOR data item
				0xc3, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "1 CBOR data item followed by non-CBOR data",
			data: []byte{
				// CBOR data item
				0xc2, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				// Extraneous non-CBOR data ("abc")
				0x61, 0x62, 0x63,
			},
			buffered: []byte{
				// non-CBOR data ("abc")
				0x61, 0x62, 0x63,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader(tc.data)

			decoder := NewDecoder(r)

			// Decoder's buffer has no data yet.
			br := decoder.Buffered()
			buffered, err := io.ReadAll(br)
			checkNoError(t, "io.ReadAll", err)
			if len(buffered) > 0 {
				t.Errorf("Buffered() = 0x%x (%d bytes), want 0 bytes", buffered, len(buffered))
			}

			var v any
			err = decoder.Decode(&v)
			if err != tc.decodeErr {
				t.Errorf("Decoder.Decode error = %v, want %v", err, tc.decodeErr)
			}

			br = decoder.Buffered()
			buffered, err = io.ReadAll(br)
			checkNoError(t, "io.ReadAll", err)
			if !bytes.Equal(tc.buffered, buffered) {
				t.Errorf("Buffered() = 0x%x (%d bytes), want 0x%x (%d bytes)", buffered, len(buffered), tc.buffered, len(tc.buffered))
			}
		})
	}
}

func TestEncoder(t *testing.T) {
	var want bytes.Buffer
	var w bytes.Buffer
	em := mustEncMode(t, CanonicalEncOptions())
	encoder := em.NewEncoder(&w)
	for _, tc := range marshalTestCases {
		for _, value := range tc.values {
			want.Write(tc.wantData)

			err := encoder.Encode(value)
			checkNoError(t, "Encoder.Encode", err)
		}
	}
	if !bytes.Equal(w.Bytes(), want.Bytes()) {
		t.Errorf("Encoder output = 0x%x, want 0x%x", w.Bytes(), want.Bytes())
	}
}

func TestEncoderError(t *testing.T) {
	testCases := []struct {
		name         string
		value        any
		wantErrorMsg string
	}{
		{
			name:         "channel cannot be marshaled",
			value:        make(chan bool),
			wantErrorMsg: "cbor: unsupported type: chan bool",
		},
		{
			name:         "function cannot be marshaled",
			value:        func(i int) int { return i * i },
			wantErrorMsg: "cbor: unsupported type: func(int) int",
		},
		{
			name:         "complex cannot be marshaled",
			value:        complex(100, 8),
			wantErrorMsg: "cbor: unsupported type: complex128",
		},
	}
	var w bytes.Buffer
	encoder := NewEncoder(&w)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := tc.value
			err := encoder.Encode(&v)
			checkError[*UnsupportedTypeError](t, "Encoder.Encode", err, tc.wantErrorMsg)
		})
	}
	if w.Len() != 0 {
		t.Errorf("Encoder output length = %d, want 0", w.Len())
	}
}

func TestIndefiniteByteString(t *testing.T) {
	want := mustHexDecode("5f42010243030405ff")
	var w bytes.Buffer
	encoder := NewEncoder(&w)
	err := encoder.StartIndefiniteByteString()
	checkNoError(t, "Encoder.StartIndefiniteByteString", err)

	err = encoder.Encode([]byte{1, 2})
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.Encode([3]byte{3, 4, 5})
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.EndIndefinite()
	checkNoError(t, "Encoder.EndIndefinite", err)

	if !bytes.Equal(w.Bytes(), want) {
		t.Errorf("Encoder output = 0x%x, want 0x%x", w.Bytes(), want)
	}
}

func TestIndefiniteByteStringWithInvalidChunk(t *testing.T) { //nolint:gocyclo
	t.Run("array as chunk", func(t *testing.T) {
		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteByteString()
		checkNoError(t, "Encoder.StartIndefiniteByteString", err)
		err = encoder.Encode([]int{1, 2})
		checkError[error](t, "Encoder.Encode", err, "cbor: cannot encode item type slice for indefinite-length byte string")
	})

	t.Run("text string as chunk", func(t *testing.T) {
		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteByteString()
		checkNoError(t, "Encoder.StartIndefiniteByteString", err)
		err = encoder.Encode("hello")
		checkError[error](t, "Encoder.Encode", err, "cbor: cannot encode item type string for indefinite-length byte string")
	})

	t.Run("nil as chunk", func(t *testing.T) {
		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteByteString()
		checkNoError(t, "Encoder.StartIndefiniteByteString", err)
		err = encoder.Encode(nil)
		checkError[error](t, "Encoder.Encode", err, "cbor: cannot encode nil for indefinite-length byte string")
	})

	children := []struct {
		start func(*Encoder) error
		typ   string
	}{
		{(*Encoder).StartIndefiniteByteString, "byte string"},
		{(*Encoder).StartIndefiniteTextString, "UTF-8 text string"},
		{(*Encoder).StartIndefiniteArray, "array"},
		{(*Encoder).StartIndefiniteMap, "map"},
	}
	for _, child := range children {
		t.Run("indefinite-length "+child.typ+" as chunk", func(t *testing.T) {
			var w bytes.Buffer
			encoder := NewEncoder(&w)
			err := encoder.StartIndefiniteByteString()
			checkNoError(t, "Encoder.StartIndefiniteByteString", err)
			err = child.start(encoder)
			wantErrMsg := "cbor: cannot encode indefinite-length " + child.typ +
				" as chunk of indefinite-length byte string"
			checkError[error](t, "child.start", err, wantErrMsg)
		})
	}

	t.Run("byte slice with ByteSliceLaterFormat option", func(t *testing.T) {
		expectedErrorMsg := "cbor: cannot encode item type slice for indefinite-length byte string"

		em := mustEncMode(t, EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase64})

		var w bytes.Buffer
		encoder := em.NewEncoder(&w)
		err := encoder.StartIndefiniteByteString()
		checkNoError(t, "Encoder.StartIndefiniteByteString", err)
		err = encoder.Encode([]byte("foo"))
		checkError[error](t, "Encoder.Encode", err, expectedErrorMsg)
	})

	t.Run("nil byte slice with NilContainerAsNull (default) option", func(t *testing.T) {
		expectedErrorMsg := "cbor: cannot encode item type slice for indefinite-length byte string"

		em := mustEncMode(t, EncOptions{NilContainers: NilContainerAsNull}) // NilContainerAsNull is the default option

		var w bytes.Buffer
		encoder := em.NewEncoder(&w)
		err := encoder.StartIndefiniteByteString()
		checkNoError(t, "Encoder.StartIndefiniteByteString", err)
		var nilBytes []byte
		err = encoder.Encode(nilBytes)
		checkError[error](t, "Encoder.Encode", err, expectedErrorMsg)
	})

	t.Run("RawMessage with tag chunk", func(t *testing.T) {
		expectedErrorMsg := "cbor: cannot encode item type slice for indefinite-length byte string"

		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteByteString()
		checkNoError(t, "Encoder.StartIndefiniteByteString", err)
		raw := RawMessage{0xd8, 0x22, 0x43, 0x01, 0x02, 0x03} // 34(h'010203')
		err = encoder.Encode(raw)
		checkError[error](t, "Encoder.Encode", err, expectedErrorMsg)
	})

	t.Run("RawMessage with indefinite-length byte string chunk", func(t *testing.T) {
		expectedErrorMsg := "cbor: cannot encode indefinite-length byte string as chunk of indefinite-length byte string"
		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteByteString()
		checkNoError(t, "Encoder.StartIndefiniteByteString", err)
		raw := RawMessage{0x5f, 0x41, 0x31, 0xff} // (_ h'31')
		err = encoder.Encode(raw)
		checkError[error](t, "Encoder.Encode", err, expectedErrorMsg)
	})

	t.Run("CBORMarshaler with byte slice kind emits int", func(t *testing.T) {
		expectedErrorMsg := "cbor: cannot encode item type slice for indefinite-length byte string"

		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteByteString()
		checkNoError(t, "Encoder.StartIndefiniteByteString", err)
		cborData := []byte{0x18, 0x7b} // 123
		err = encoder.Encode(bytesMarshaler(cborData))
		checkError[error](t, "Encoder.Encode", err, expectedErrorMsg)
	})

	t.Run("CBORMarshaler with byte slice kind emits text string", func(t *testing.T) {
		expectedErrorMsg := "cbor: cannot encode item type slice for indefinite-length byte string"

		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteByteString()
		checkNoError(t, "Encoder.StartIndefiniteByteString", err)
		cborData := []byte{0x63, 'f', 'o', 'o'} // "foo"
		err = encoder.Encode(bytesMarshaler(cborData))
		checkError[error](t, "Encoder.Encode", err, expectedErrorMsg)
	})

	t.Run("CBORMarshaler with byte slice kind emits indefinite-length byte string", func(t *testing.T) {
		expectedErrorMsg := "cbor: cannot encode indefinite-length byte string as chunk of indefinite-length byte string"

		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteByteString()
		checkNoError(t, "Encoder.StartIndefiniteByteString", err)
		cborData := []byte{0x5f, 0x41, 0x61, 0xff} // (_ h'61')
		err = encoder.Encode(bytesMarshaler(cborData))
		checkError[error](t, "Encoder.Encode", err, expectedErrorMsg)
	})
}

func TestIndefiniteTextString(t *testing.T) {
	want := mustHexDecode("7f657374726561646d696e67ff")
	var w bytes.Buffer
	encoder := NewEncoder(&w)
	err := encoder.StartIndefiniteTextString()
	checkNoError(t, "Encoder.StartIndefiniteTextString", err)

	err = encoder.Encode("strea")
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.Encode("ming")
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.EndIndefinite()
	checkNoError(t, "Encoder.EndIndefinite", err)

	if !bytes.Equal(w.Bytes(), want) {
		t.Errorf("Encoder output = 0x%x, want 0x%x", w.Bytes(), want)
	}
}

func TestIndefiniteTextStringWithInvalidChunk(t *testing.T) {
	t.Run("array as chunk", func(t *testing.T) {
		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteTextString()
		checkNoError(t, "Encoder.StartIndefiniteTextString", err)
		err = encoder.Encode([]byte{1, 2})
		checkError[error](t, "Encoder.Encode", err, "cbor: cannot encode item type slice for indefinite-length UTF-8 text string")
	})

	t.Run("integer as chunk", func(t *testing.T) {
		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteTextString()
		checkNoError(t, "Encoder.StartIndefiniteTextString", err)
		err = encoder.Encode(123)
		checkError[error](t, "Encoder.Encode", err, "cbor: cannot encode item type int for indefinite-length UTF-8 text string")
	})

	t.Run("nil as chunk", func(t *testing.T) {
		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteTextString()
		checkNoError(t, "Encoder.StartIndefiniteTextString", err)
		err = encoder.Encode(nil)
		checkError[error](t, "Encoder.Encode", err, "cbor: cannot encode nil for indefinite-length UTF-8 text string")
	})

	children := []struct {
		start func(*Encoder) error
		typ   string
	}{
		{(*Encoder).StartIndefiniteByteString, "byte string"},
		{(*Encoder).StartIndefiniteTextString, "UTF-8 text string"},
		{(*Encoder).StartIndefiniteArray, "array"},
		{(*Encoder).StartIndefiniteMap, "map"},
	}
	for _, child := range children {
		t.Run("indefinite-length "+child.typ+" as chunk", func(t *testing.T) {
			var w bytes.Buffer
			encoder := NewEncoder(&w)
			err := encoder.StartIndefiniteTextString()
			checkNoError(t, "Encoder.StartIndefiniteTextString", err)
			err = child.start(encoder)
			wantErrMsg := "cbor: cannot encode indefinite-length " + child.typ +
				" as chunk of indefinite-length UTF-8 text string"
			checkError[error](t, "child.start", err, wantErrMsg)
		})
	}

	t.Run("string with StringToByteString option", func(t *testing.T) {
		expectedErrorMsg := "cbor: cannot encode item type string for indefinite-length UTF-8 text string"

		em := mustEncMode(t, EncOptions{String: StringToByteString})
		var w bytes.Buffer
		encoder := em.NewEncoder(&w)
		err := encoder.StartIndefiniteTextString()
		checkNoError(t, "Encoder.StartIndefiniteTextString", err)
		err = encoder.Encode("foo")
		checkError[error](t, "Encoder.Encode", err, expectedErrorMsg)
	})

	t.Run("RawMessage with indefinite-length text string chunk", func(t *testing.T) {
		expectedErrorMsg := "cbor: cannot encode indefinite-length UTF-8 text string as chunk of indefinite-length UTF-8 text string"

		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteTextString()
		checkNoError(t, "Encoder.StartIndefiniteTextString", err)
		raw := RawMessage{0x7f, 0x61, 0x61, 0xff} // (_ "a")
		err = encoder.Encode(raw)
		checkError[error](t, "Encoder.Encode", err, expectedErrorMsg)
	})

	t.Run("CBORMarshaler with string kind emits uint", func(t *testing.T) {
		expectedErrorMsg := "cbor: cannot encode item type string for indefinite-length UTF-8 text string"

		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteTextString()
		checkNoError(t, "Encoder.StartIndefiniteTextString", err)
		cborData := []byte{0x18, 0x7b} // 123
		err = encoder.Encode(stringMarshaler(cborData))
		checkError[error](t, "Encoder.Encode", err, expectedErrorMsg)
	})

	t.Run("CBORMarshaler with string kind emits byte string", func(t *testing.T) {
		expectedErrorMsg := "cbor: cannot encode item type string for indefinite-length UTF-8 text string"

		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteTextString()
		checkNoError(t, "Encoder.StartIndefiniteTextString", err)
		cborData := []byte{0x43, 0x01, 0x02, 0x03} // h'010203'
		err = encoder.Encode(stringMarshaler(cborData))
		checkError[error](t, "Encoder.Encode", err, expectedErrorMsg)
	})

	t.Run("CBORMarshaler with string kind emits indefinite-length text string", func(t *testing.T) {
		expectedErrorMsg := "cbor: cannot encode indefinite-length UTF-8 text string as chunk of indefinite-length UTF-8 text string"

		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteTextString()
		checkNoError(t, "Encoder.StartIndefiniteTextString", err)
		cborData := []byte{0x7f, 0x61, 0x61, 0xff} // (_ "a")
		err = encoder.Encode(stringMarshaler(cborData))
		checkError[error](t, "Encoder.Encode", err, expectedErrorMsg)
	})
}

type bytesMarshaler []byte

func (m bytesMarshaler) MarshalCBOR() ([]byte, error) {
	return []byte(m), nil
}

type stringMarshaler string

func (m stringMarshaler) MarshalCBOR() ([]byte, error) {
	return []byte(m), nil
}

func TestIndefiniteArray(t *testing.T) {
	want := mustHexDecode("9f018202039f0405ffff")
	var w bytes.Buffer
	encoder := NewEncoder(&w)
	err := encoder.StartIndefiniteArray()
	checkNoError(t, "Encoder.StartIndefiniteArray", err)

	err = encoder.Encode(1)
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.Encode([]int{2, 3})
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.StartIndefiniteArray()
	checkNoError(t, "Encoder.StartIndefiniteArray", err)

	err = encoder.Encode(4)
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.Encode(5)
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.EndIndefinite()
	checkNoError(t, "Encoder.EndIndefinite", err)

	err = encoder.EndIndefinite()
	checkNoError(t, "Encoder.EndIndefinite", err)

	if !bytes.Equal(w.Bytes(), want) {
		t.Errorf("Encoder output = 0x%x, want 0x%x", w.Bytes(), want)
	}
}

func TestIndefiniteArrayWithNilElement(t *testing.T) {
	want := mustHexDecode("9f01f6ff") // [1, null]
	var w bytes.Buffer
	encoder := NewEncoder(&w)
	err := encoder.StartIndefiniteArray()
	checkNoError(t, "Encoder.StartIndefiniteArray", err)

	err = encoder.Encode(1)
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.Encode(nil)
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.EndIndefinite()
	checkNoError(t, "Encoder.EndIndefinite", err)

	if !bytes.Equal(w.Bytes(), want) {
		t.Errorf("Encoder output = 0x%x, want 0x%x", w.Bytes(), want)
	}

	var decoded []any
	err = unmarshal(t, w.Bytes(), &decoded)
	checkUnmarshal(t, "Unmarshal", decoded, err, []any{uint64(1), nil})
}

func TestIndefiniteMap(t *testing.T) {
	want := mustHexDecode("bf61610161629f0203ffff")
	var w bytes.Buffer
	em := mustEncMode(t, EncOptions{Sort: SortCanonical})
	encoder := em.NewEncoder(&w)
	err := encoder.StartIndefiniteMap()
	checkNoError(t, "Encoder.StartIndefiniteMap", err)

	err = encoder.Encode("a")
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.Encode(1)
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.Encode("b")
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.StartIndefiniteArray()
	checkNoError(t, "Encoder.StartIndefiniteArray", err)

	err = encoder.Encode(2)
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.Encode(3)
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.EndIndefinite()
	checkNoError(t, "Encoder.EndIndefinite", err)

	err = encoder.EndIndefinite()
	checkNoError(t, "Encoder.EndIndefinite", err)

	if !bytes.Equal(w.Bytes(), want) {
		t.Errorf("Encoder output = 0x%x, want 0x%x", w.Bytes(), want)
	}
}

func TestIndefiniteMapWithNilElement(t *testing.T) {
	want := mustHexDecode("bf6161f6ff") // {"a": null}
	var w bytes.Buffer
	encoder := NewEncoder(&w)
	err := encoder.StartIndefiniteMap()
	checkNoError(t, "Encoder.StartIndefiniteMap", err)

	err = encoder.Encode("a")
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.Encode(nil)
	checkNoError(t, "Encoder.Encode", err)

	err = encoder.EndIndefinite()
	checkNoError(t, "Encoder.EndIndefinite", err)

	if !bytes.Equal(w.Bytes(), want) {
		t.Errorf("Encoder output = 0x%x, want 0x%x", w.Bytes(), want)
	}

	var decoded map[string]any
	err = unmarshal(t, w.Bytes(), &decoded)
	checkUnmarshal(t, "Unmarshal", decoded, err, map[string]any{"a": nil})
}

func TestIndefiniteLengthMapItemCount(t *testing.T) {
	t.Run("odd item count", func(t *testing.T) {
		var buf bytes.Buffer
		encoder := NewEncoder(&buf)
		err := encoder.StartIndefiniteMap()
		checkNoError(t, "Encoder.StartIndefiniteMap", err)

		err = encoder.Encode("a")
		checkNoError(t, "Encoder.Encode", err)

		err = encoder.EndIndefinite()
		wantErrorMsg := "cbor: cannot end indefinite-length map with 1 item(s)"
		checkError[*IndefiniteLengthMapOddItemCountError](t, "Encoder.EndIndefinite", err, wantErrorMsg)

		// Write the missing value.
		err = encoder.Encode(1)
		checkNoError(t, "Encoder.Encode", err)

		err = encoder.EndIndefinite()
		checkNoError(t, "Encoder.EndIndefinite", err)

		wantData := mustHexDecode("bf616101ff") // {_ "a": 1}
		if !bytes.Equal(buf.Bytes(), wantData) {
			t.Errorf("Encoder output = 0x%x, want 0x%x", buf.Bytes(), wantData)
		}
	})

	t.Run("even item count with nested container", func(t *testing.T) {
		var buf bytes.Buffer
		encoder := NewEncoder(&buf)
		err := encoder.StartIndefiniteMap()
		checkNoError(t, "Encoder.StartIndefiniteMap", err)

		err = encoder.Encode("a")
		checkNoError(t, "Encoder.Encode", err)

		err = encoder.StartIndefiniteMap()
		checkNoError(t, "Encoder.StartIndefiniteMap", err)

		err = encoder.Encode("b")
		checkNoError(t, "Encoder.Encode", err)

		err = encoder.Encode(1)
		checkNoError(t, "Encoder.Encode", err)

		err = encoder.EndIndefinite()
		checkNoError(t, "Encoder.EndIndefinite", err)

		err = encoder.EndIndefinite()
		checkNoError(t, "Encoder.EndIndefinite", err)

		wantData := mustHexDecode("bf6161bf616201ffff") // {_ "a": {_ "b": 1}}
		if !bytes.Equal(buf.Bytes(), wantData) {
			t.Errorf("Encoder output = 0x%x, want 0x%x", buf.Bytes(), wantData)
		}
	})

	t.Run("odd item count with nested container", func(t *testing.T) {
		var buf bytes.Buffer
		encoder := NewEncoder(&buf)
		err := encoder.StartIndefiniteMap()
		checkNoError(t, "Encoder.StartIndefiniteMap", err)

		err = encoder.Encode("a")
		checkNoError(t, "Encoder.Encode", err)

		err = encoder.StartIndefiniteMap()
		checkNoError(t, "Encoder.StartIndefiniteMap", err)

		err = encoder.Encode("b")
		checkNoError(t, "Encoder.Encode", err)

		err = encoder.Encode(1)
		checkNoError(t, "Encoder.Encode", err)

		err = encoder.EndIndefinite()
		checkNoError(t, "Encoder.EndIndefinite", err)

		err = encoder.Encode("c")
		checkNoError(t, "Encoder.Encode", err)

		err = encoder.EndIndefinite()
		wantErrorMsg := "cbor: cannot end indefinite-length map with 3 item(s)"
		checkError[*IndefiniteLengthMapOddItemCountError](t, "Encoder.EndIndefinite", err, wantErrorMsg)
	})
}

func TestIndefiniteLengthError(t *testing.T) {
	var w bytes.Buffer
	encoder := NewEncoder(&w)
	err := encoder.StartIndefiniteByteString()
	checkNoError(t, "Encoder.StartIndefiniteByteString", err)
	err = encoder.EndIndefinite()
	checkNoError(t, "Encoder.EndIndefinite", err)
	err = encoder.EndIndefinite()
	checkError[error](t, "Encoder.EndIndefinite", err, "")
}

func TestEncodeNestedIndefiniteStringInArrayAndMap(t *testing.T) {
	t.Run("indefinite-length byte string in indefinite-length array", func(t *testing.T) {
		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteArray()
		checkNoError(t, "Encoder.StartIndefiniteArray", err)
		err = encoder.StartIndefiniteByteString()
		checkNoError(t, "Encoder.StartIndefiniteByteString", err)
		err = encoder.EndIndefinite()
		checkNoError(t, "Encoder.EndIndefinite", err)
		err = encoder.EndIndefinite()
		checkNoError(t, "Encoder.EndIndefinite", err)
	})

	t.Run("indefinite-length text string in indefinite-length map", func(t *testing.T) {
		var w bytes.Buffer
		encoder := NewEncoder(&w)
		err := encoder.StartIndefiniteMap()
		checkNoError(t, "Encoder.StartIndefiniteMap", err)
		err = encoder.Encode("k")
		checkNoError(t, "Encoder.Encode", err)
		err = encoder.StartIndefiniteTextString()
		checkNoError(t, "Encoder.StartIndefiniteTextString", err)
		err = encoder.EndIndefinite()
		checkNoError(t, "Encoder.EndIndefinite", err)
		err = encoder.EndIndefinite()
		checkNoError(t, "Encoder.EndIndefinite", err)
	})
}

func TestEncoderStructTag(t *testing.T) {
	type strc struct {
		A string `json:"x" cbor:"a"`
		B string `json:"y" cbor:"b"`
		C string `json:"z"`
	}
	v := strc{
		A: "A",
		B: "B",
		C: "C",
	}
	want := mustHexDecode("a36161614161626142617a6143") // {"a":"A", "b":"B", "z":"C"}

	var w bytes.Buffer
	encoder := NewEncoder(&w)
	err := encoder.Encode(v)
	checkNoError(t, "Encoder.Encode", err)
	if !bytes.Equal(w.Bytes(), want) {
		t.Errorf("Encoder output = 0x%x, want 0x%x", w.Bytes(), want)
	}
}

func TestRawMessage(t *testing.T) {
	type strc struct {
		A RawMessage  `cbor:"a"`
		B *RawMessage `cbor:"b"`
		C *RawMessage `cbor:"c"`
	}
	data := mustHexDecode("a361610161628202036163f6") // {"a": 1, "b": [2, 3], "c": nil},
	r := RawMessage(mustHexDecode("820203"))
	want := strc{
		A: RawMessage([]byte{0x01}),
		B: &r,
	}
	var v strc
	err := unmarshal(t, data, &v)
	checkUnmarshal(t, "Unmarshal", v, err, want)
	b, err := Marshal(v)
	checkMarshal(t, "Marshal", b, err, data)

	addr := fmt.Sprintf("%p", *v.B)
	err = unmarshal(t, v.A, v.B)
	checkNoError(t, "Unmarshal", err)
	if addr != fmt.Sprintf("%p", *v.B) {
		t.Fatalf("Unmarshal allocated a new RawMessage array, want reuse (sufficient capacity)")
	}
	err = unmarshal(t, data, v.B)
	checkNoError(t, "Unmarshal", err)
	if addr == fmt.Sprintf("%p", *v.B) {
		t.Fatalf("Unmarshal reused the RawMessage array, want a new one (insufficient capacity)")
	}
}

func TestNullRawMessage(t *testing.T) {
	r := RawMessage(nil)
	wantCborData := []byte{0xf6}
	b, err := Marshal(r)
	checkMarshal(t, "Marshal", b, err, wantCborData)
}

func TestEmptyRawMessage(t *testing.T) {
	var r RawMessage
	wantCborData := []byte{0xf6}
	b, err := Marshal(r)
	checkMarshal(t, "Marshal", b, err, wantCborData)
}

func TestNilRawMessageUnmarshalCBORError(t *testing.T) {
	wantErrorMsg := "cbor.RawMessage: UnmarshalCBOR on nil pointer"
	var r *RawMessage
	data := mustHexDecode("01")
	err := r.UnmarshalCBOR(data)
	checkError[error](t, "r.UnmarshalCBOR", err, wantErrorMsg)
}

var streamTestReaders = []struct {
	name      string
	newReader func(data []byte) io.Reader
}{
	{"bytes.Reader", func(data []byte) io.Reader { return bytes.NewReader(data) }},
	{"1 byte reader", func(data []byte) io.Reader { return newNBytesReader(data, 1) }},
	{"toggled reader", func(data []byte) io.Reader { return newToggledReader(data, 1) }},
}

func TestDecoderRejectedDataItem(t *testing.T) {
	testCases := []struct {
		name    string
		opts    DecOptions
		data    []byte
		wantErr error
	}{
		{
			name:    "tags forbidden",
			opts:    DecOptions{TagsMd: TagsForbidden},
			data:    []byte{0xc1, 0x01}, // 1(1)
			wantErr: (*TagsMdError)(nil),
		},
		{
			name:    "indefinite-length array forbidden",
			opts:    DecOptions{IndefLength: IndefLengthForbidden},
			data:    []byte{0x9f, 0x01, 0xff}, // [_ 1]
			wantErr: (*IndefiniteLengthError)(nil),
		},
		{
			name:    "NaN forbidden in array",
			opts:    DecOptions{NaN: NaNDecodeForbidden},
			data:    []byte{0x82, 0xf9, 0x7e, 0x00, 0x01}, // [NaN, 1]
			wantErr: (*UnacceptableDataItemError)(nil),
		},
	}

	for _, tc := range testCases {
		dm := mustDecMode(t, tc.opts)

		data := append(append([]byte{}, tc.data...), 0x61, 0x61) // concatenate CBOR data item "a" after tc.data
		want := "a"

		for _, r := range streamTestReaders {
			t.Run(tc.name+"/"+r.name, func(t *testing.T) {
				decoder := dm.NewDecoder(r.newReader(data))

				// Decoding the first data item returns an error and offset is advanced to the next data item.
				var v1 any
				name := fmt.Sprintf("Decode(0x%x)", tc.data)
				err := decoder.Decode(&v1)
				checkErrorType(t, name, err, tc.wantErr)

				if got := decoder.NumBytesRead(); got != len(tc.data) {
					t.Errorf("NumBytesRead() = %d, want %d", got, len(tc.data))
				}

				// Decoding the second data item doesn't return error and offset is advanced to the end.
				var v2 any
				err = decoder.Decode(&v2)
				checkUnmarshal(t, "Decoder.Decode", v2, err, want)

				if got := decoder.NumBytesRead(); got != len(data) {
					t.Errorf("NumBytesRead() = %d, want %d", got, len(data))
				}

				// Skip the first data item.
				decoder = dm.NewDecoder(r.newReader(data))
				err = decoder.Skip()
				checkNoError(t, "Decoder.Skip", err)
				if got := decoder.NumBytesRead(); got != len(tc.data) {
					t.Errorf("NumBytesRead() = %d, want %d", got, len(tc.data))
				}

				// Skip the second data item.
				err = decoder.Skip()
				checkNoError(t, "Decoder.Skip", err)
				if got := decoder.NumBytesRead(); got != len(data) {
					t.Errorf("NumBytesRead() = %d, want %d", got, len(data))
				}
			})
		}
	}
}

func TestDecoderRejectedDataItemNoAdvance(t *testing.T) {
	testCases := []struct {
		name          string
		opts          DecOptions
		data          []byte
		wantDecodeErr error
		wantSkipErr   error
	}{
		{
			name:          "truncated data after forbidden NaN",
			opts:          DecOptions{NaN: NaNDecodeForbidden},
			data:          []byte{0x82, 0xf9, 0x7e, 0x00}, // truncated data (missing second array element after NaN)
			wantDecodeErr: (*UnacceptableDataItemError)(nil),
			wantSkipErr:   io.ErrUnexpectedEOF,
		},
		{
			name:          "malformed data after forbidden NaN",
			opts:          DecOptions{NaN: NaNDecodeForbidden},
			data:          []byte{0x82, 0xf9, 0x7e, 0x00, 0xff},
			wantDecodeErr: (*UnacceptableDataItemError)(nil),
			wantSkipErr:   (*SyntaxError)(nil),
		},
		{
			name:          "MaxNestedLevels exceeded after rejected value",
			opts:          DecOptions{NaN: NaNDecodeForbidden, MaxNestedLevels: 4},
			data:          []byte{0x82, 0xf9, 0x7e, 0x00, 0x81, 0x81, 0x81, 0x81, 0x01}, // [NaN, [[[[1]]]]]
			wantDecodeErr: (*UnacceptableDataItemError)(nil),
			wantSkipErr:   (*MaxNestedLevelError)(nil),
		},
	}

	for _, tc := range testCases {
		dm := mustDecMode(t, tc.opts)

		for _, r := range streamTestReaders {
			t.Run(tc.name+"/"+r.name, func(t *testing.T) {
				decoder := dm.NewDecoder(r.newReader(tc.data))
				for range 2 {
					var v any
					name := fmt.Sprintf("Decode(0x%x)", tc.data)
					err := decoder.Decode(&v)
					checkErrorType(t, name, err, tc.wantDecodeErr)
					if got := decoder.NumBytesRead(); got != 0 {
						t.Errorf("NumBytesRead() = %d, want 0", got)
					}
				}

				decoder = dm.NewDecoder(r.newReader(tc.data))
				for range 2 {
					name := fmt.Sprintf("Skip(0x%x)", tc.data)
					err := decoder.Skip()
					checkErrorType(t, name, err, tc.wantSkipErr)
					if got := decoder.NumBytesRead(); got != 0 {
						t.Errorf("NumBytesRead() = %d, want 0", got)
					}
				}
			})
		}
	}
}
