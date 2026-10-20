// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"math/big"
	"reflect"
	"testing"
	"time"
)

// Go helper functions

func mustHexDecode(s string) []byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func mustBigInt(s string) big.Int {
	bi, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("failed to convert " + s + " to big.Int")
	}
	return *bi
}

func mustParseTime(layout string, value string) time.Time {
	tm, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return tm
}

// CBOR helper functions

func isCBORNil(data []byte) bool {
	return len(data) > 0 && (data[0] == 0xf6 || data[0] == 0xf7)
}

func encodeCborHeader(t cborType, n uint64) []byte {
	if n <= maxAdditionalInformationWithoutArgument {
		const headSize = 1
		var b [headSize]byte
		b[0] = byte(t) | byte(n)
		return b[:]
	}

	if n <= math.MaxUint8 {
		const argumentSize = 1
		const headSize = 1 + argumentSize
		var b [headSize]byte
		b[0] = byte(t) | additionalInformationWith1ByteArgument
		b[1] = byte(n)
		return b[:]
	}

	if n <= math.MaxUint16 {
		const argumentSize = 2
		const headSize = 1 + argumentSize
		var b [headSize]byte
		b[0] = byte(t) | additionalInformationWith2ByteArgument
		binary.BigEndian.PutUint16(b[1:], uint16(n))
		return b[:]
	}

	if n <= math.MaxUint32 {
		const argumentSize = 4
		const headSize = 1 + argumentSize
		var b [headSize]byte
		b[0] = byte(t) | additionalInformationWith4ByteArgument
		binary.BigEndian.PutUint32(b[1:], uint32(n))
		return b[:]
	}

	const argumentSize = 8
	const headSize = 1 + argumentSize
	var b [headSize]byte
	b[0] = byte(t) | additionalInformationWith8ByteArgument
	binary.BigEndian.PutUint64(b[1:], n)
	return b[:]
}

// Codec helper functions

func unmarshal(t *testing.T, data []byte, v any) error {
	t.Helper()
	return unmarshalWithDM(t, defaultDecMode, data, v)
}

// unmarshalWithDM wraps DecMode.Unmarshal() to test a decoder invariant:
// given a well-formed CBOR data item and a valid non-nil pointer destination value,
// Unmarshal advances the data offset to the next data item on success and on error.
func unmarshalWithDM(t *testing.T, dm DecMode, data []byte, v any) error {
	t.Helper()

	rv := reflect.ValueOf(v)
	isValidValue := rv.IsValid() && rv.Kind() == reflect.Pointer && !rv.IsNil()
	isZeroValue := isValidValue && rv.Elem().IsZero()

	// Call Unmarshal to retain code coverage
	err := dm.Unmarshal(data, v)

	// Duplicate Unmarshal code to test d.off
	d := decoder{data: data, dm: dm.(*decMode)}

	off := d.off
	if wellformedErr := d.wellformed(false, false); wellformedErr != nil {
		if !sameError(wellformedErr, err) {
			t.Errorf("wellformed(0x%x) error = %v, want %v", data, wellformedErr, err)
		}
		return err
	}

	if !isValidValue {
		if _, ok := err.(*InvalidUnmarshalError); !ok {
			t.Errorf("Unmarshal(0x%x, %T) error = %v, want *InvalidUnmarshalError", data, v, err)
		}
		return err
	}

	v2 := reflect.New(rv.Type().Elem())

	d.off = off
	parseErr := d.value(v2.Interface())
	if isZeroValue && !sameError(parseErr, err) {
		t.Errorf("value(0x%x, %T) error = %v, want %v", data, v, parseErr, err)
	}

	if d.off != len(data) {
		t.Errorf("value(0x%x, %T) offset = %d, want %d", data, v, d.off, len(data))
	}

	return err
}

func testMarshalWithFP(t *testing.T, f func(t *testing.T)) {
	testCases := []struct {
		name     string
		enableFP bool
	}{
		{name: "enable fast path", enableFP: true},
		{name: "disable fast path", enableFP: false},
	}
	t.Cleanup(func() {
		enableEncodingFastPath = true
		encodeFuncCache.Clear()
		encodingStructTypeCache.Clear()
	})
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enableEncodingFastPath = tc.enableFP
			encodeFuncCache.Clear()
			encodingStructTypeCache.Clear()
			f(t)
		})
	}
}

// Other helper functions

type roundTripTestCase struct {
	name         string
	obj          any
	wantCborData []byte
}

func testRoundTrip(t *testing.T, testCases []roundTripTestCase, em EncMode, dm DecMode) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := em.Marshal(tc.obj)
			if err != nil {
				t.Errorf("Marshal(%+v) returned error %v", tc.obj, err)
			}
			if !bytes.Equal(b, tc.wantCborData) {
				t.Errorf("Marshal(%+v) = 0x%x, want 0x%x", tc.obj, b, tc.wantCborData)
			}
			v := reflect.New(reflect.TypeOf(tc.obj))
			if err := unmarshalWithDM(t, dm, b, v.Interface()); err != nil {
				t.Errorf("Unmarshal() returned error %v", err)
			}
			if !reflect.DeepEqual(tc.obj, v.Elem().Interface()) {
				t.Errorf("Marshal-Unmarshal returned different values: %v, %v", tc.obj, v.Elem().Interface())
			}
		})
	}
}

func compareNonFloats(t *testing.T, data []byte, got any, want any) {
	t.Helper()

	switch tm := want.(type) {
	case time.Time:
		if vt, ok := got.(time.Time); !ok || !tm.Equal(vt) {
			t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", data, got, got, want, want)
		}

	default:
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", data, got, got, want, want)
		}
	}
}

func compareFloats(t *testing.T, data []byte, got any, want any, equalityThreshold float64) {
	t.Helper()

	var gotFloat64, wantFloat64 float64

	switch want := want.(type) {
	case float32:
		f, ok := got.(float32)
		if !ok {
			t.Errorf("Unmarshal(0x%x) returned value of type %T, want float32", data, f)
			return
		}
		gotFloat64, wantFloat64 = float64(f), float64(want)

	case float64:
		f, ok := got.(float64)
		if !ok {
			t.Errorf("Unmarshal(0x%x) returned value of type %T, want float64", data, f)
			return
		}
		gotFloat64, wantFloat64 = f, want
	}

	switch {
	case math.IsNaN(wantFloat64):
		if !math.IsNaN(gotFloat64) {
			t.Errorf("Unmarshal(0x%x) = %f, want NaN", data, gotFloat64)
		}

	case math.IsInf(wantFloat64, 0):
		if gotFloat64 != wantFloat64 {
			t.Errorf("Unmarshal(0x%x) = %f, want %f", data, gotFloat64, wantFloat64)
		}

	default:
		if math.Abs(gotFloat64-wantFloat64) > equalityThreshold {
			t.Errorf("Unmarshal(0x%x) = %.18f, want %.18f, diff %.18f > threshold %.18f", data, gotFloat64, wantFloat64, math.Abs(gotFloat64-wantFloat64), equalityThreshold)
		}
	}
}

func sameError(a, b error) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return reflect.TypeOf(a) == reflect.TypeOf(b) && a.Error() == b.Error()
}

// testErrorType tests if got error matches want error.
func testErrorType(t *testing.T, call string, got, want error) {
	t.Helper()
	if got == nil && want == nil {
		return
	}
	if want == nil {
		t.Errorf("%s: unexpected error: %v", call, got)
	} else if got == nil {
		t.Errorf("%s: expected %T, got nil", call, want)
	} else if rv := reflect.ValueOf(want); rv.Kind() != reflect.Pointer || !rv.IsNil() {
		if !errors.Is(got, want) {
			t.Errorf("%s error = %v, want %v", call, got, want)
		}
	} else if !errors.As(got, reflect.New(reflect.TypeOf(want)).Interface()) {
		t.Errorf("%s returned %T, want %T", call, got, want)
	}
}

// nBytesReader reads at most maxBytesPerRead into b.  It also returns error at the last read.
type nBytesReader struct {
	data            []byte
	maxBytesPerRead int
	off             int
	err             error
}

func newNBytesReader(data []byte, maxBytesPerRead int) *nBytesReader {
	return &nBytesReader{
		data:            append([]byte{}, data...),
		maxBytesPerRead: maxBytesPerRead,
		err:             io.EOF,
	}
}

func newNBytesReaderWithError(data []byte, maxBytesPerRead int, err error) *nBytesReader {
	return &nBytesReader{
		data:            append([]byte{}, data...),
		maxBytesPerRead: maxBytesPerRead,
		err:             err,
	}
}

func (r *nBytesReader) Read(b []byte) (int, error) {
	var n int
	if r.off < len(r.data) {
		numOfBytesToRead := min(len(r.data)-r.off, r.maxBytesPerRead)
		n = copy(b, r.data[r.off:r.off+numOfBytesToRead])
		r.off += n
	}
	if r.off == len(r.data) {
		return n, r.err
	}
	return n, nil
}

// toggledReader returns (0, nil) for every other read to mimic non-blocking read for stream reader.
type toggledReader struct {
	nBytesReader
	toggle bool
}

func newToggledReader(data []byte, maxBytesPerRead int) *toggledReader {
	return &toggledReader{
		nBytesReader: nBytesReader{
			data:            append([]byte{}, data...),
			maxBytesPerRead: maxBytesPerRead,
			err:             io.EOF,
		},
		toggle: true, // first read returns (0, nil)
	}
}

func newToggledReaderWithError(data []byte, maxBytesPerRead int, err error) *toggledReader {
	return &toggledReader{
		nBytesReader: nBytesReader{
			data:            append([]byte{}, data...),
			maxBytesPerRead: maxBytesPerRead,
			err:             err,
		},
		toggle: true, // first read returns (0, nil)
	}
}

func (r *toggledReader) Read(b []byte) (int, error) {
	defer func() {
		r.toggle = !r.toggle
	}()
	if r.toggle {
		return 0, nil
	}
	return r.nBytesReader.Read(b)
}

// recoverableReader returns a recoverable error at first read operation.
type recoverableReader struct {
	nBytesReader
	recoverableErr error
	first          bool
}

func newRecoverableReader(data []byte, maxBytesPerRead int, err error) *recoverableReader {
	return &recoverableReader{
		nBytesReader: nBytesReader{
			data:            append([]byte{}, data...),
			maxBytesPerRead: maxBytesPerRead,
			err:             io.EOF,
		},
		recoverableErr: err,
		first:          true,
	}
}

func (r *recoverableReader) Read(b []byte) (int, error) {
	if r.first {
		r.first = false
		return 0, r.recoverableErr
	}
	return r.nBytesReader.Read(b)
}
