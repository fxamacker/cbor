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

func mustEncMode(t *testing.T, opts EncOptions) EncMode { //nolint:gocritic // ignore hugeParam
	t.Helper()
	em, err := opts.EncMode()
	if err != nil {
		t.Fatalf("EncMode: unexpected error: %v", err)
	}
	return em
}

func mustEncModeWithTags(t *testing.T, opts EncOptions, tags TagSet) EncMode { //nolint:gocritic // ignore hugeParam
	t.Helper()
	em, err := opts.EncModeWithTags(tags)
	if err != nil {
		t.Fatalf("EncModeWithTags: unexpected error: %v", err)
	}
	return em
}

func mustUserBufferEncMode(t *testing.T, opts EncOptions) UserBufferEncMode { //nolint:gocritic // ignore hugeParam
	t.Helper()
	em, err := opts.UserBufferEncMode()
	if err != nil {
		t.Fatalf("UserBufferEncMode: unexpected error: %v", err)
	}
	return em
}

func mustDecMode(t *testing.T, opts DecOptions) DecMode { //nolint:gocritic // ignore hugeParam
	t.Helper()
	dm, err := opts.DecMode()
	if err != nil {
		t.Fatalf("DecMode: unexpected error: %v", err)
	}
	return dm
}

func mustDecModeWithTags(t *testing.T, opts DecOptions, tags TagSet) DecMode { //nolint:gocritic // ignore hugeParam
	t.Helper()
	dm, err := opts.DecModeWithTags(tags)
	if err != nil {
		t.Fatalf("DecModeWithTags: unexpected error: %v", err)
	}
	return dm
}

func mustAddTag(t *testing.T, tags TagSet, opts TagOptions, contentType reflect.Type, num uint64, nestedNum ...uint64) {
	t.Helper()
	err := tags.Add(opts, contentType, num, nestedNum...)
	if err != nil {
		t.Fatalf("TagSet.Add: unexpected error: %v", err)
	}
}

// Other helper functions

// checkMarshal checks whether a marshal call was successful.
func checkMarshal(t *testing.T, call string, b []byte, err error, want []byte) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", call, err)
	}
	if !bytes.Equal(b, want) {
		t.Errorf("%s = 0x%x, want 0x%x", call, b, want)
	}
}

// checkMarshalErr checks whether a marshal call failed with an error
// that errors.As matches to E, and no data.  The error message is checked
// if wantErrMsg isn't empty.  Use E = error to accept any error type.
func checkMarshalErr[E error](t *testing.T, call string, b []byte, err error, wantErrMsg string) {
	t.Helper()
	checkError[E](t, call, err, wantErrMsg)
	if b != nil {
		t.Errorf("%s = 0x%x, want nil", call, b)
	}
}

// checkMarshalCase calls checkMarshalErr[E] if wantErr is true, otherwise
// checkMarshal.  wantErrMsg is ignored if wantErr is false.
func checkMarshalCase[E error](t *testing.T, call string, b []byte, err error, want []byte, wantErr bool, wantErrMsg string) {
	t.Helper()
	if wantErr {
		checkMarshalErr[E](t, call, b, err, wantErrMsg)
	} else {
		checkMarshal(t, call, b, err, want)
	}
}

// checkUnmarshal checks whether an unmarshal call was successful and
// compares got value with want value.
func checkUnmarshal(t *testing.T, call string, got any, err error, want any) {
	t.Helper()
	checkUnmarshalFunc(t, call, got, err, want, reflect.DeepEqual)
}

// checkUnmarshalFunc is like checkUnmarshal but compares values with equal.
func checkUnmarshalFunc(t *testing.T, call string, got any, err error, want any, equal func(got, want any) bool) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", call, err)
	}
	if !equal(got, want) {
		t.Errorf("%s = %+v (%T), want %+v (%T)", call, got, got, want, want)
	}
}

// checkUnmarshalErr checks whether an unmarshal call failed with an error
// matching E (see checkError), and compares got value with want value.
func checkUnmarshalErr[E error](t *testing.T, call string, got any, err error, want any, wantErrMsg string) {
	t.Helper()
	checkUnmarshalErrFunc[E](t, call, got, err, want, wantErrMsg, reflect.DeepEqual)
}

// checkUnmarshalErrFunc is like checkUnmarshalErr but compares values with equal.
func checkUnmarshalErrFunc[E error](t *testing.T, call string, got any, err error, want any, wantErrMsg string, equal func(got, want any) bool) {
	t.Helper()
	checkError[E](t, call, err, wantErrMsg)
	if !equal(got, want) {
		t.Errorf("%s = %+v (%T), want %+v (%T)", call, got, got, want, want)
	}
}

// checkUnmarshalCase calls checkError[E] if wantErr is true, otherwise
// checkUnmarshal.  got and want are not compared on error, and
// wantErrMsg is ignored if wantErr is false.
func checkUnmarshalCase[E error](t *testing.T, call string, got any, err error, want any, wantErr bool, wantErrMsg string) {
	t.Helper()
	checkUnmarshalCaseFunc[E](t, call, got, err, want, wantErr, wantErrMsg, reflect.DeepEqual)
}

// checkUnmarshalCaseFunc is like checkUnmarshalCase but compares values with equal.
func checkUnmarshalCaseFunc[E error](t *testing.T, call string, got any, err error, want any, wantErr bool, wantErrMsg string, equal func(got, want any) bool) {
	t.Helper()
	if wantErr {
		checkError[E](t, call, err, wantErrMsg)
	} else {
		checkUnmarshalFunc(t, call, got, err, want, equal)
	}
}

// testDecoderEOF calls dec.Decode and checks whether it returns io.EOF and no value.
func testDecoderEOF(t *testing.T, dec *Decoder) {
	t.Helper()
	var v any
	err := dec.Decode(&v)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Decoder.Decode: got error %v, want io.EOF", err)
	}
	if v != nil {
		t.Errorf("Decoder.Decode = %v (%T), want nil (no more data)", v, v)
	}
}

type roundTripTestCase struct {
	name string
	v    any
	data []byte
}

func testRoundTripCases(t *testing.T, testCases []roundTripTestCase, em EncMode, dm DecMode) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testRoundTripCase(t, tc.v, tc.data, em, dm)
		})
	}
}

func testRoundTripCase(t *testing.T, v any, data []byte, em EncMode, dm DecMode) {
	t.Helper()
	b, err := em.Marshal(v)
	checkMarshal(t, "em.Marshal", b, err, data)

	rv := reflect.New(reflect.TypeOf(v))
	err = unmarshalWithDM(t, dm, b, rv.Interface())
	checkUnmarshal(t, "dm.Unmarshal", rv.Elem().Interface(), err, v)
}

func nonFloatEqual(got any, want any) bool {
	if wt, ok := want.(time.Time); ok {
		gt, ok := got.(time.Time)
		return ok && wt.Equal(gt)
	}

	return reflect.DeepEqual(got, want)
}

func floatEqual(threshold float64) func(got, want any) bool {
	return func(got, want any) bool {
		var gotFloat64, wantFloat64 float64
		switch want := want.(type) {
		case float32:
			f, ok := got.(float32)
			if !ok {
				return false
			}
			gotFloat64, wantFloat64 = float64(f), float64(want)
		case float64:
			f, ok := got.(float64)
			if !ok {
				return false
			}
			gotFloat64, wantFloat64 = f, want
		default:
			return false
		}

		switch {
		case math.IsNaN(wantFloat64):
			return math.IsNaN(gotFloat64)
		case math.IsInf(wantFloat64, 0):
			return gotFloat64 == wantFloat64
		default:
			return math.Abs(gotFloat64-wantFloat64) <= threshold
		}
	}
}

func sameError(a, b error) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return reflect.TypeOf(a) == reflect.TypeOf(b) && a.Error() == b.Error()
}

// checkErrorType checks if got error matches want error.
func checkErrorType(t *testing.T, call string, got, want error) {
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

// checkError checks whether a call failed with an error that errors.As
// matches to E.  The error message is checked if wantErrMsg isn't empty.
// Use E = error to accept any error type.
func checkError[E error](t *testing.T, call string, err error, wantErrMsg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected %s, got nil", call, reflect.TypeFor[E]())
	}
	var e E
	if !errors.As(err, &e) {
		t.Errorf("%s returned %T, want %s", call, err, reflect.TypeFor[E]())
	} else if wantErrMsg != "" && err.Error() != wantErrMsg {
		t.Errorf("%s: got %q, want %q", call, err.Error(), wantErrMsg)
	}
}

// checkNoError checks whether a call succeeded.
func checkNoError(t *testing.T, call string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", call, err)
	}
}

// checkErrorCase checks a call in a test case that either fails or succeeds.
// If wantErr is true, it checks err like checkError; otherwise, like
// checkNoError, and wantErrMsg is ignored.
func checkErrorCase[E error](t *testing.T, call string, err error, wantErr bool, wantErrMsg string) {
	t.Helper()
	if wantErr {
		checkError[E](t, call, err, wantErrMsg)
	} else {
		checkNoError(t, call, err)
	}
}

// derefOrNil returns the value pointed to by p, or nil if p is a nil pointer.
// derefOrNil panics if p's dynamic type isn't a pointer.
func derefOrNil(p any) any {
	rv := reflect.ValueOf(p)
	if rv.IsNil() {
		return nil
	}
	return rv.Elem().Interface()
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
