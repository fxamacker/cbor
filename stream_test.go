// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

package cbor_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/fxamacker/cbor"
)

func TestDecoder(t *testing.T) {
	var buf bytes.Buffer
	for i := 0; i < 5; i++ {
		for _, tc := range unmarshalTests {
			buf.Write(tc.cborData)
		}
	}
	decoder := cbor.NewDecoder(&buf)
	bytesRead := 0
	for i := 0; i < 5; i++ {
		for _, tc := range unmarshalTests {
			var v interface{}
			if err := decoder.Decode(&v); err != nil {
				t.Fatalf("Decode() returns error %v", err)
			}
			if !reflect.DeepEqual(v, tc.emptyInterfaceValue) {
				t.Errorf("Decode() = %v (%T), want %v (%T)", v, v, tc.emptyInterfaceValue, tc.emptyInterfaceValue)
			}
			bytesRead += len(tc.cborData)
			if decoder.NumBytesRead() != bytesRead {
				t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), bytesRead)
			}
		}
	}
	// no more data
	var v interface{}
	err := decoder.Decode(&v)
	if v != nil {
		t.Errorf("Decode() = %v (%T), want nil (no more data)", v, v)
	}
	if err != io.EOF {
		t.Errorf("Decode() returns error %v, want io.EOF (no more data)", err)
	}
}

func TestDecoderUnmarshalTypeError(t *testing.T) {
	var buf bytes.Buffer
	for i := 0; i < 5; i++ {
		for _, tc := range unmarshalTests {
			buf.Write(tc.cborData)
		}
	}
	decoder := cbor.NewDecoder(&buf)
	bytesRead := 0
	wrongType := true
	for i := 0; i < 5; i++ {
		for _, tc := range unmarshalTests {
			if wrongType && len(tc.wrongTypes) > 0 {
				wrongType = !wrongType
				typ := tc.wrongTypes[0]
				v := reflect.New(typ)
				vPtr := v.Interface()
				err := decoder.Decode(vPtr)
				if err == nil {
					t.Errorf("Unmarshal(0x%0x) returns %v (%T), want UnmarshalTypeError", tc.cborData, v.Elem().Interface(), v.Elem().Interface())
				} else if _, ok := err.(*cbor.UnmarshalTypeError); !ok {
					t.Errorf("Unmarshal(0x%0x) returns wrong error %s, want UnmarshalTypeError", tc.cborData, err.Error())
				}
			} else {
				wrongType = !wrongType
				var v interface{}
				if err := decoder.Decode(&v); err != nil {
					t.Errorf("Decode() returns error %v", err)
				}
				if !reflect.DeepEqual(v, tc.emptyInterfaceValue) {
					t.Errorf("Decode() = %v (%T), want %v (%T)", v, v, tc.emptyInterfaceValue, tc.emptyInterfaceValue)
				}
			}

			bytesRead += len(tc.cborData)
			if decoder.NumBytesRead() != bytesRead {
				t.Errorf("NumBytesRead() = %v, want %v", decoder.NumBytesRead(), bytesRead)
			}
		}
	}
	// no more data
	var v interface{}
	err := decoder.Decode(&v)
	if v != nil {
		t.Errorf("Decode() = %v (%T), want nil (no more data)", v, v)
	}
	if err != io.EOF {
		t.Errorf("Decode() returns error %v, want io.EOF (no more data)", err)
	}
}

func TestEncoder(t *testing.T) {
	var want bytes.Buffer
	var w bytes.Buffer
	encoder := cbor.NewEncoder(&w, cbor.EncOptions{Canonical: true})
	for _, tc := range marshalTests {
		for _, value := range tc.values {
			want.Write(tc.cborData)

			if err := encoder.Encode(value); err != nil {
				t.Fatalf("Encode() returns error %v", err)
			}
		}
	}
	if !bytes.Equal(w.Bytes(), want.Bytes()) {
		t.Errorf("Encoding mismatch: got %v, want %v", w.Bytes(), want.Bytes())
	}
}
