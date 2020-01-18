// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"math"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/x448/float16"
)

// Marshal returns the CBOR encoding of v.
//
// Marshal uses the following type-dependent default encodings:
//
// Boolean values encode as CBOR booleans (type 7).
//
// Positive integer values encode as CBOR positive integers (type 0).
//
// Negative integer values encode as CBOR negative integers (type 1).
//
// Floating point values encode as CBOR floating points (type 7).
//
// String values encode as CBOR text strings (type 3).
//
// []byte values encode as CBOR byte strings (type 2).
//
// Array and slice values encode as CBOR arrays (type 4).
//
// Map values encode as CBOR maps (type 5).
//
// Struct values encode as CBOR maps (type 5).  Each exported struct field
// becomes a pair with field name encoded as CBOR text string (type 3) and
// field value encoded based on its type.
//
// Pointer values encode as the value pointed to.
//
// Nil slice/map/pointer/interface values encode as CBOR nulls (type 7).
//
// time.Time values encode as text strings specified in RFC3339 when
// EncOptions.TimeRFC3339 is true; otherwise, time.Time values encode as
// numerical representation of seconds since January 1, 1970 UTC.
//
// If value implements the Marshaler interface, Marshal calls its MarshalCBOR
// method.  If value implements encoding.BinaryMarshaler instead, Marhsal
// calls its MarshalBinary method and encode it as CBOR byte string.
//
// Marshal supports format string stored under the "cbor" key in the struct
// field's tag.  CBOR format string can specify the name of the field, "omitempty"
// and "keyasint" options, and special case "-" for field omission.  If "cbor"
// key is absent, Marshal uses "json" key.
//
// Struct field name is treated as integer if it has "keyasint" option in
// its format string.  The format string must specify an integer as its
// field name.
//
// Special struct field "_" is used to specify struct level options, such as
// "toarray". "toarray" option enables Go struct to be encoded as CBOR array.
// "omitempty" is disabled by "toarray" to ensure that the same number
// of elements are encoded every time.
//
// Anonymous struct fields are usually marshalled as if their exported fields
// were fields in the outer struct.  Marshal follows the same struct fields
// visibility rules used by JSON encoding package.  An anonymous struct field
// with a name given in its CBOR tag is treated as having that name, rather
// than being anonymous.  An anonymous struct field of interface type is
// treated the same as having that type as its name, rather than being anonymous.
//
// Interface values encode as the value contained in the interface.  A nil
// interface value encodes as the null CBOR value.
//
// Channel, complex, and functon values cannot be encoded in CBOR.  Attempting
// to encode such a value causes Marshal to return an UnsupportedTypeError.
func Marshal(v interface{}, encOpts EncOptions) ([]byte, error) {
	e := getEncodeState()

	err := e.marshal(v, encOpts)
	if err != nil {
		putEncodeState(e)
		return nil, err
	}

	buf := make([]byte, e.Len())
	copy(buf, e.Bytes())

	putEncodeState(e)
	return buf, nil
}

// Marshaler is the interface implemented by types that can marshal themselves
// into valid CBOR.
type Marshaler interface {
	MarshalCBOR() ([]byte, error)
}

// UnsupportedTypeError is returned by Marshal when attempting to encode an
// unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "cbor: unsupported type: " + e.Type.String()
}

// SortMode identifies supported sorting order.
type SortMode int

const (
	// SortNone means no sorting.
	SortNone SortMode = 0

	// SortLengthFirst causes map keys or struct fields to be sorted such that:
	//     - If two keys have different lengths, the shorter one sorts earlier;
	//     - If two keys have the same length, the one with the lower value in
	//       (byte-wise) lexical order sorts earlier.
	// It is used in "Canonical CBOR" encoding in RFC 7049 3.9.
	SortLengthFirst SortMode = 1

	// SortBytewiseLexical causes map keys or struct fields to be sorted in the
	// bytewise lexicographic order of their deterministic CBOR encodings.
	// It is used in "CTAP2 Canonical CBOR" and "Core Deterministic Encoding"
	// in RFC 7049bis.
	SortBytewiseLexical SortMode = 2

	// SortCanonical is used in "Canonical CBOR" encoding in RFC 7049 3.9.
	SortCanonical SortMode = SortLengthFirst

	// SortCTAP2 is used in "CTAP2 Canonical CBOR".
	SortCTAP2 SortMode = SortBytewiseLexical

	// SortCoreDeterministic is used in "Core Deterministic Encoding" in RFC 7049bis.
	SortCoreDeterministic SortMode = SortBytewiseLexical

	maxSortMode SortMode = 3
)

func (sm SortMode) valid() bool {
	return sm < maxSortMode
}

// ShortestFloatMode specifies which floating-point format should
// be used as the shortest possible format for CBOR encoding.
// It is not used for encoding Infinity and NaN values.
type ShortestFloatMode int

const (
	// ShortestFloatNone makes float values encode without any conversion.
	// This is the default for ShortestFloatMode in v1.
	// E.g. a float32 in Go will encode to CBOR float32.  And
	// a float64 in Go will encode to CBOR float64.
	ShortestFloatNone ShortestFloatMode = iota

	// ShortestFloat16 specifies float16 as the shortest form that preserves value.
	// E.g. if float64 can convert to float32 while perserving value, then
	// encoding will also try to convert float32 to float16.  So a float64 might
	// encode as CBOR float64, float32 or float16 depending on the value.
	ShortestFloat16

	maxShortestFloat
)

func (sfm ShortestFloatMode) valid() bool {
	return sfm < maxShortestFloat
}

// NaNConvertMode specifies how to encode NaN and overrides ShortestFloatMode.
// ShortestFloatMode is not used for encoding Infinity and NaN values.
type NaNConvertMode int

const (
	// NaNConvert7e00 always encodes NaN to 0xf97e00 (CBOR float16 = 0x7e00).
	NaNConvert7e00 NaNConvertMode = iota

	// NaNConvertNone never modifies or converts NaN to other representations
	// (float64 NaN stays float64, etc. even if it can use float16 without losing
	// any bits).
	NaNConvertNone

	// NaNConvertPreserveSignal converts NaN to the smallest form that preserves
	// value (quiet bit + payload) as described in RFC 7049bis Draft 12.
	NaNConvertPreserveSignal

	// NaNConvertQuiet always forces quiet bit = 1 and shortest form that preserves
	// NaN payload.
	NaNConvertQuiet

	maxNaNConvert
)

func (ncm NaNConvertMode) valid() bool {
	return ncm < maxNaNConvert
}

// InfConvertMode specifies how to encode Infinity and overrides ShortestFloatMode.
// ShortestFloatMode is not used for encoding Infinity and NaN values.
type InfConvertMode int

const (
	// InfConvertFloat16 always converts Inf to lossless IEEE binary16 (float16).
	InfConvertFloat16 InfConvertMode = iota

	// InfConvertNone never converts (used by CTAP2 Canonical CBOR).
	InfConvertNone

	maxInfConvert
)

func (icm InfConvertMode) valid() bool {
	return icm < maxInfConvert
}

// EncOptions specifies encoding options.
type EncOptions struct {
	// Sort specifies sorting order.
	Sort SortMode

	// ShortestFloat specifies the shortest floating-point encoding that preserves
	// the value being encoded.
	ShortestFloat ShortestFloatMode

	// NaNConvert specifies how to encode NaN and it overrides ShortestFloatMode.
	NaNConvert NaNConvertMode

	// InfConvert specifies how to encode Inf and it overrides ShortestFloatMode.
	InfConvert InfConvertMode

	// Canonical causes map and struct to be encoded in a predictable sequence
	// of bytes by sorting map keys or struct fields according to canonical rules:
	//     - If two keys have different lengths, the shorter one sorts earlier;
	//     - If two keys have the same length, the one with the lower value in
	//       (byte-wise) lexical order sorts earlier.
	//
	// Deprecated: Set Sort to SortCanonical instead.
	Canonical bool

	// CTAP2Canonical uses bytewise lexicographic order of map keys encodings:
	//     - If the major types are different, the one with the lower value in
	//       numerical order sorts earlier.
	//     - If two keys have different lengths, the shorter one sorts earlier;
	//     - If two keys have the same length, the one with the lower value in
	//       (byte-wise) lexical order sorts earlier.
	// Please note that when maps keys have the same data type, "canonical CBOR"
	// AND "CTAP2 canonical CBOR" render the same sort order.
	//
	// Deprecated: Set Sort to SortCTAP2 instead.
	CTAP2Canonical bool

	// TimeRFC3339 causes time.Time to be encoded as string in RFC3339 format;
	// otherwise, time.Time is encoded as numerical representation of seconds
	// since January 1, 1970 UTC.
	TimeRFC3339 bool

	disableIndefiniteLength bool
}

// CanonicalEncOptions returns EncOptions for "Canonical CBOR" encoding,
// defined in RFC 7049 Section 3.9 with the following rules:
//
//     1. "Integers must be as small as possible."
//     2. "The expression of lengths in major types 2 through 5 must be as short as possible."
//     3. The keys in every map must be sorted in length-first sorting order.
//        See SortLengthFirst for details.
//     4. "Indefinite-length items must be made into definite-length items."
//     5. "If a protocol allows for IEEE floats, then additional canonicalization rules might
//        need to be added.  One example rule might be to have all floats start as a 64-bit
//        float, then do a test conversion to a 32-bit float; if the result is the same numeric
//        value, use the shorter value and repeat the process with a test conversion to a
//        16-bit float.  (This rule selects 16-bit float for positive and negative Infinity
//        as well.)  Also, there are many representations for NaN.  If NaN is an allowed value,
//        it must always be represented as 0xf97e00."
//
func CanonicalEncOptions() EncOptions {
	return EncOptions{
		Sort:                    SortCanonical,
		ShortestFloat:           ShortestFloat16,
		NaNConvert:              NaNConvert7e00,
		InfConvert:              InfConvertFloat16,
		disableIndefiniteLength: true,
	}
}

// CTAP2EncOptions returns EncOptions for "CTAP2 Canonical CBOR" encoding,
// defined in CTAP specification, with the following rules:
//
//     1. "Integers must be encoded as small as possible."
//     2. "The representations of any floating-point values are not changed."
//     3. "The expression of lengths in major types 2 through 5 must be as short as possible."
//     4. "Indefinite-length items must be made into definite-length items.""
//     5. The keys in every map must be sorted in bytewise lexicographic order.
//        See SortBytewiseLexical for details.
//     6. "Tags as defined in Section 2.4 in [RFC7049] MUST NOT be present."
//
func CTAP2EncOptions() EncOptions {
	return EncOptions{
		Sort:                    SortCTAP2,
		ShortestFloat:           ShortestFloatNone,
		NaNConvert:              NaNConvertNone,
		InfConvert:              InfConvertNone,
		disableIndefiniteLength: true,
	}
}

// CoreDetEncOptions returns EncOptions for "Core Deterministic" encoding,
// defined in RFC 7049bis with the following rules:
//
//     1. "Preferred serialization MUST be used. In particular, this means that arguments
//        (see Section 3) for integers, lengths in major types 2 through 5, and tags MUST
//        be as short as possible"
//        "Floating point values also MUST use the shortest form that preserves the value"
//     2. "Indefinite-length items MUST NOT appear."
//     3. "The keys in every map MUST be sorted in the bytewise lexicographic order of
//        their deterministic encodings."
//
func CoreDetEncOptions() EncOptions {
	return EncOptions{
		Sort:                    SortCoreDeterministic,
		ShortestFloat:           ShortestFloat16,
		NaNConvert:              NaNConvert7e00,
		InfConvert:              InfConvertFloat16,
		disableIndefiniteLength: true,
	}
}

// PreferredUnsortedEncOptions returns EncOptions for "Preferred Serialization" encoding,
// defined in RFC 7049bis with the following rules:
//
//     1. "The preferred serialization always uses the shortest form of representing the argument
//        (Section 3);"
//     2. "it also uses the shortest floating-point encoding that preserves the value being
//        encoded (see Section 5.5)."
//        "The preferred encoding for a floating-point value is the shortest floating-point encoding
//        that preserves its value, e.g., 0xf94580 for the number 5.5, and 0xfa45ad9c00 for the
//        number 5555.5, unless the CBOR-based protocol specifically excludes the use of the shorter
//        floating-point encodings. For NaN values, a shorter encoding is preferred if zero-padding
//        the shorter significand towards the right reconstitutes the original NaN value (for many
//        applications, the single NaN encoding 0xf97e00 will suffice)."
//     3. "Definite length encoding is preferred whenever the length is known at the time the
//        serialization of the item starts."
//
func PreferredUnsortedEncOptions() EncOptions {
	return EncOptions{
		Sort:          SortNone,
		ShortestFloat: ShortestFloat16,
		NaNConvert:    NaNConvert7e00,
		InfConvert:    InfConvertFloat16,
	}
}

// An encodeState encodes CBOR into a bytes.Buffer.
type encodeState struct {
	bytes.Buffer
	scratch [16]byte
}

// encodeStatePool caches unused encodeState objects for later reuse.
var encodeStatePool = sync.Pool{
	New: func() interface{} {
		e := new(encodeState)
		e.Grow(32) // TODO: make this configurable
		return e
	},
}

func getEncodeState() *encodeState {
	return encodeStatePool.Get().(*encodeState)
}

// putEncodeState returns e to encodeStatePool.
func putEncodeState(e *encodeState) {
	e.Reset()
	encodeStatePool.Put(e)
}

func (e *encodeState) marshal(v interface{}, opts EncOptions) error {
	if !opts.Sort.valid() {
		return errors.New("cbor: invalid SortMode " + strconv.Itoa(int(opts.Sort)))
	}
	if opts.Canonical {
		if opts.Sort != SortNone && opts.Sort != SortCanonical {
			return errors.New("cbor: conflicting sorting modes not allowed")
		}
		opts.Sort = SortCanonical
	}
	if opts.CTAP2Canonical {
		if opts.Sort != SortNone && opts.Sort != SortCTAP2 {
			return errors.New("cbor: conflicting sorting modes not allowed")
		}
		opts.Sort = SortCTAP2
	}
	if !opts.ShortestFloat.valid() {
		return errors.New("cbor: invalid ShortestFloatMode " + strconv.Itoa(int(opts.ShortestFloat)))
	}
	if !opts.NaNConvert.valid() {
		return errors.New("cbor: invalid NaNConvertMode " + strconv.Itoa(int(opts.NaNConvert)))
	}
	if !opts.InfConvert.valid() {
		return errors.New("cbor: invalid InfConvertMode " + strconv.Itoa(int(opts.InfConvert)))
	}
	_, err := encode(e, reflect.ValueOf(v), opts)
	return err
}

type encodeFunc func(e *encodeState, v reflect.Value, opts EncOptions) (int, error)

var (
	cborFalse            = []byte{0xf4}
	cborTrue             = []byte{0xf5}
	cborNil              = []byte{0xf6}
	cborNan              = []byte{0xf9, 0x7e, 0x00}
	cborPositiveInfinity = []byte{0xf9, 0x7c, 0x00}
	cborNegativeInfinity = []byte{0xf9, 0xfc, 0x00}
)

func encode(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if !v.IsValid() {
		// v is zero value
		e.Write(cborNil)
		return 1, nil
	}
	f := getEncodeFunc(v.Type())
	if f == nil {
		return 0, &UnsupportedTypeError{v.Type()}
	}

	return f(e, v, opts)
}

func encodeBool(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if v.Bool() {
		return e.Write(cborTrue)
	}
	return e.Write(cborFalse)
}

func encodeInt(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	i := v.Int()
	if i >= 0 {
		return encodeHead(e, byte(cborTypePositiveInt), uint64(i)), nil
	}
	n := v.Int()*(-1) - 1
	return encodeHead(e, byte(cborTypeNegativeInt), uint64(n)), nil
}

func encodeUint(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	return encodeHead(e, byte(cborTypePositiveInt), v.Uint()), nil
}

func encodeFloat(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	f64 := v.Float()
	if math.IsNaN(f64) {
		return encodeNaN(e, v, opts)
	}
	if math.IsInf(f64, 0) {
		return encodeInf(e, v, opts)
	}
	fopt := opts.ShortestFloat
	if v.Kind() == reflect.Float64 && (fopt == ShortestFloatNone || cannotFitFloat32(f64)) {
		// Encode float64
		// Don't use encodeFloat64() because it cannot be inlined.
		e.scratch[0] = byte(cborTypePrimitives) | byte(27)
		binary.BigEndian.PutUint64(e.scratch[1:], math.Float64bits(f64))
		return e.Write(e.scratch[:9])
	}

	f32 := float32(f64)
	if fopt == ShortestFloat16 {
		var f16 float16.Float16
		p := float16.PrecisionFromfloat32(f32)
		if p == float16.PrecisionExact {
			// Roundtrip float32->float16->float32 test isn't needed.
			f16 = float16.Fromfloat32(f32)
		} else if p == float16.PrecisionUnknown {
			// Try roundtrip float32->float16->float32 to determine if float32 can fit into float16.
			f16 = float16.Fromfloat32(f32)
			if f16.Float32() == f32 {
				p = float16.PrecisionExact
			}
		}
		if p == float16.PrecisionExact {
			// Encode float16
			// Don't use encodeFloat16() because it cannot be inlined.
			e.scratch[0] = byte(cborTypePrimitives) | byte(25)
			binary.BigEndian.PutUint16(e.scratch[1:], uint16(f16))
			return e.Write(e.scratch[:3])
		}
	}

	// Encode float32
	// Don't use encodeFloat32() because it cannot be inlined.
	e.scratch[0] = byte(cborTypePrimitives) | byte(26)
	binary.BigEndian.PutUint32(e.scratch[1:], math.Float32bits(f32))
	return e.Write(e.scratch[:5])
}

func encodeInf(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	f64 := v.Float()
	if opts.InfConvert == InfConvertFloat16 {
		if f64 > 0 {
			return e.Write(cborPositiveInfinity)
		}
		return e.Write(cborNegativeInfinity)
	}
	if v.Kind() == reflect.Float64 {
		return encodeFloat64(e, f64)
	}
	return encodeFloat32(e, float32(f64))
}

func encodeNaN(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	switch opts.NaNConvert {
	case NaNConvert7e00:
		return e.Write(cborNan)

	case NaNConvertNone:
		if v.Kind() == reflect.Float64 {
			return encodeFloat64(e, v.Float())
		}
		f32 := float32NaNFromReflectValue(v)
		return encodeFloat32(e, f32)

	default: // NaNConvertPreserveSignal, NaNConvertQuiet
		if v.Kind() == reflect.Float64 {
			f64 := v.Float()
			f64bits := math.Float64bits(f64)
			if opts.NaNConvert == NaNConvertQuiet && f64bits&(1<<51) == 0 {
				f64bits = f64bits | (1 << 51) // Set quiet bit = 1
				f64 = math.Float64frombits(f64bits)
			}
			// The lower 29 bits are dropped when converting from float64 to float32.
			if f64bits&0x1fffffff != 0 {
				// Encode NaN as float64 because dropped coef bits from float64 to float32 are not all 0s.
				return encodeFloat64(e, f64)
			}
			// Create float32 from float64 manually because float32(f64) always turns on NaN's quiet bits.
			sign := uint32(f64bits>>32) & (1 << 31)
			exp := uint32(0x7f800000)
			coef := uint32((f64bits & 0xfffffffffffff) >> 29)
			f32bits := sign | exp | coef
			f32 := math.Float32frombits(f32bits)
			// The lower 13 bits are dropped when converting from float32 to float16.
			if f32bits&0x1fff != 0 {
				// Encode NaN as float32 because dropped coef bits from float32 to float16 are not all 0s.
				return encodeFloat32(e, f32)
			}
			// Encode NaN as float16
			f16, _ := float16.FromNaN32ps(f32) // Ignore err because it only returns error when f32 is not a NaN.
			return encodeFloat16(e, f16)
		}

		f32 := float32NaNFromReflectValue(v)
		f32bits := math.Float32bits(f32)
		if opts.NaNConvert == NaNConvertQuiet && f32bits&(1<<22) == 0 {
			f32bits = f32bits | (1 << 22) // Set quiet bit = 1
			f32 = math.Float32frombits(f32bits)
		}
		// The lower 13 bits are dropped coef bits when converting from float32 to float16.
		if f32bits&0x1fff != 0 {
			// Encode NaN as float32 because dropped coef bits from float32 to float16 are not all 0s.
			return encodeFloat32(e, f32)
		}
		f16, _ := float16.FromNaN32ps(f32) // Ignore err because it only returns error when f32 is not a NaN.
		return encodeFloat16(e, f16)
	}
}

func encodeFloat16(e *encodeState, f16 float16.Float16) (int, error) {
	e.scratch[0] = byte(cborTypePrimitives) | byte(25)
	binary.BigEndian.PutUint16(e.scratch[1:], uint16(f16))
	return e.Write(e.scratch[:3])
}

func encodeFloat32(e *encodeState, f32 float32) (int, error) {
	e.scratch[0] = byte(cborTypePrimitives) | byte(26)
	binary.BigEndian.PutUint32(e.scratch[1:], math.Float32bits(f32))
	return e.Write(e.scratch[:5])
}

func encodeFloat64(e *encodeState, f64 float64) (int, error) {
	e.scratch[0] = byte(cborTypePrimitives) | byte(27)
	binary.BigEndian.PutUint64(e.scratch[1:], math.Float64bits(f64))
	return e.Write(e.scratch[:9])
}

func encodeByteString(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if v.Kind() == reflect.Slice && v.IsNil() {
		return e.Write(cborNil)
	}
	slen := v.Len()
	if slen == 0 {
		return 1, e.WriteByte(byte(cborTypeByteString))
	}
	n1 := encodeHead(e, byte(cborTypeByteString), uint64(slen))
	if v.Kind() == reflect.Array {
		for i := 0; i < slen; i++ {
			e.WriteByte(byte(v.Index(i).Uint()))
		}
		return n1 + slen, nil
	}
	n2, _ := e.Write(v.Bytes())
	return n1 + n2, nil
}

func encodeString(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	return encodeStringInternal(e, v.String(), opts)
}

func encodeStringInternal(e *encodeState, s string, opts EncOptions) (int, error) {
	n1 := encodeHead(e, byte(cborTypeTextString), uint64(len(s)))
	n2, _ := e.WriteString(s)
	return n1 + n2, nil
}

type arrayEncoder struct {
	f encodeFunc
}

func (ae arrayEncoder) encodeArray(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if ae.f == nil {
		return 0, &UnsupportedTypeError{v.Type()}
	}
	if v.Kind() == reflect.Slice && v.IsNil() {
		return e.Write(cborNil)
	}
	alen := v.Len()
	if alen == 0 {
		return 1, e.WriteByte(byte(cborTypeArray))
	}
	n := encodeHead(e, byte(cborTypeArray), uint64(alen))
	for i := 0; i < alen; i++ {
		n1, err := ae.f(e, v.Index(i), opts)
		if err != nil {
			return 0, err
		}
		n += n1
	}
	return n, nil
}

type mapEncoder struct {
	kf, ef encodeFunc
}

func (me mapEncoder) encodeMap(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if opts.Sort != SortNone {
		return me.encodeMapCanonical(e, v, opts)
	}
	if me.kf == nil || me.ef == nil {
		return 0, &UnsupportedTypeError{v.Type()}
	}
	if v.IsNil() {
		return e.Write(cborNil)
	}
	mlen := v.Len()
	if mlen == 0 {
		return 1, e.WriteByte(byte(cborTypeMap))
	}
	n := encodeHead(e, byte(cborTypeMap), uint64(mlen))
	iter := v.MapRange()
	for iter.Next() {
		n1, err := me.kf(e, iter.Key(), opts)
		if err != nil {
			return 0, err
		}
		n2, err := me.ef(e, iter.Value(), opts)
		if err != nil {
			return 0, err
		}
		n += n1 + n2
	}
	return n, nil
}

type keyValue struct {
	keyCBORData, keyValueCBORData []byte
	keyLen, keyValueLen           int
}

type pairs []keyValue

func (x pairs) Len() int {
	return len(x)
}

func (x pairs) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

type byBytewiseKeyValues struct {
	pairs
}

func (x byBytewiseKeyValues) Less(i, j int) bool {
	return bytes.Compare(x.pairs[i].keyCBORData, x.pairs[j].keyCBORData) <= 0
}

type byLengthFirstKeyValues struct {
	pairs
}

func (x byLengthFirstKeyValues) Less(i, j int) bool {
	if len(x.pairs[i].keyCBORData) != len(x.pairs[j].keyCBORData) {
		return len(x.pairs[i].keyCBORData) < len(x.pairs[j].keyCBORData)
	}
	return bytes.Compare(x.pairs[i].keyCBORData, x.pairs[j].keyCBORData) <= 0
}

var keyValuePool = sync.Pool{}

func getKeyValues(length int) []keyValue {
	v := keyValuePool.Get()
	if v == nil {
		return make([]keyValue, 0, length)
	}
	x := v.([]keyValue)
	if cap(x) < length {
		// []keyValue from the pool does not have enough capacity.
		// Return it back to the pool and create a new one.
		keyValuePool.Put(x)
		return make([]keyValue, 0, length)
	}
	return x
}

func putKeyValues(x []keyValue) {
	x = x[:0]
	keyValuePool.Put(x)
}

func (me mapEncoder) encodeMapCanonical(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if me.kf == nil || me.ef == nil {
		return 0, &UnsupportedTypeError{v.Type()}
	}
	if v.IsNil() {
		return e.Write(cborNil)
	}
	if v.Len() == 0 {
		return 1, e.WriteByte(byte(cborTypeMap))
	}

	kve := getEncodeState()      // accumulated cbor encoded key-values
	kvs := getKeyValues(v.Len()) // for sorting keys
	iter := v.MapRange()
	for iter.Next() {
		n1, err := me.kf(kve, iter.Key(), opts)
		if err != nil {
			putEncodeState(kve)
			putKeyValues(kvs)
			return 0, err
		}
		n2, err := me.ef(kve, iter.Value(), opts)
		if err != nil {
			putEncodeState(kve)
			putKeyValues(kvs)
			return 0, err
		}
		kvs = append(kvs, keyValue{keyLen: n1, keyValueLen: n1 + n2})
	}

	b := kve.Bytes()
	for i, off := 0, 0; i < len(kvs); i++ {
		kvs[i].keyCBORData = b[off : off+kvs[i].keyLen]
		kvs[i].keyValueCBORData = b[off : off+kvs[i].keyValueLen]
		off += kvs[i].keyValueLen
	}

	if opts.Sort == SortBytewiseLexical {
		sort.Sort(byBytewiseKeyValues{kvs})
	} else {
		sort.Sort(byLengthFirstKeyValues{kvs})
	}

	n := encodeHead(e, byte(cborTypeMap), uint64(len(kvs)))
	for i := 0; i < len(kvs); i++ {
		n1, _ := e.Write(kvs[i].keyValueCBORData)
		n += n1
	}

	putEncodeState(kve)
	putKeyValues(kvs)
	return n, nil
}

func encodeStructToArray(e *encodeState, v reflect.Value, opts EncOptions, flds fields) (int, error) {
	n := encodeHead(e, byte(cborTypeArray), uint64(len(flds)))
FieldLoop:
	for i := 0; i < len(flds); i++ {
		fv := v
		for k, n := range flds[i].idx {
			if k > 0 {
				if fv.Kind() == reflect.Ptr && fv.Type().Elem().Kind() == reflect.Struct {
					if fv.IsNil() {
						// Write nil for null pointer to embedded struct
						e.Write(cborNil)
						continue FieldLoop
					}
					fv = fv.Elem()
				}
			}
			fv = fv.Field(n)
		}
		n1, err := flds[i].ef(e, fv, opts)
		if err != nil {
			return 0, err
		}
		n += n1
	}
	return n, nil
}

func encodeFixedLengthStruct(e *encodeState, v reflect.Value, opts EncOptions, flds fields) (int, error) {
	n := encodeHead(e, byte(cborTypeMap), uint64(len(flds)))

	for i := 0; i < len(flds); i++ {
		n1, _ := e.Write(flds[i].cborName)

		fv := v.Field(flds[i].idx[0])
		n2, err := flds[i].ef(e, fv, opts)
		if err != nil {
			return 0, err
		}

		n += n1 + n2
	}

	return n, nil
}

func encodeStruct(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	structType := getEncodingStructType(v.Type())
	if structType.err != nil {
		return 0, structType.err
	}

	if structType.toArray {
		return encodeStructToArray(e, v, opts, structType.fields)
	}

	flds := structType.getFields(opts)

	if !structType.hasAnonymousField && !structType.omitEmpty {
		return encodeFixedLengthStruct(e, v, opts, flds)
	}

	kve := getEncodeState() // encode key-value pairs based on struct field tag options
	kvcount := 0
FieldLoop:
	for i := 0; i < len(flds); i++ {
		fv := v
		for k, n := range flds[i].idx {
			if k > 0 {
				if fv.Kind() == reflect.Ptr && fv.Type().Elem().Kind() == reflect.Struct {
					if fv.IsNil() {
						// Null pointer to embedded struct
						continue FieldLoop
					}
					fv = fv.Elem()
				}
			}
			fv = fv.Field(n)
		}
		if flds[i].omitEmpty && isEmptyValue(fv) {
			continue
		}

		kve.Write(flds[i].cborName)

		if _, err := flds[i].ef(kve, fv, opts); err != nil {
			putEncodeState(kve)
			return 0, err
		}
		kvcount++
	}

	n1 := encodeHead(e, byte(cborTypeMap), uint64(kvcount))
	n2, err := e.Write(kve.Bytes())

	putEncodeState(kve)
	return n1 + n2, err
}

func encodeIntf(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if v.IsNil() {
		return e.Write(cborNil)
	}
	return encode(e, v.Elem(), opts)
}

func encodeTime(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	t := v.Interface().(time.Time)
	if t.IsZero() {
		return e.Write(cborNil)
	} else if opts.TimeRFC3339 {
		return encodeStringInternal(e, t.Format(time.RFC3339Nano), opts)
	} else {
		t = t.UTC().Round(time.Microsecond)
		secs, nsecs := t.Unix(), uint64(t.Nanosecond())
		if nsecs == 0 {
			return encodeInt(e, reflect.ValueOf(secs), opts)
		}
		f := float64(secs) + float64(nsecs)/1e9
		return encodeFloat(e, reflect.ValueOf(f), opts)
	}
}

func encodeBinaryMarshalerType(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	m, ok := v.Interface().(encoding.BinaryMarshaler)
	if !ok {
		pv := reflect.New(v.Type())
		pv.Elem().Set(v)
		m = pv.Interface().(encoding.BinaryMarshaler)
	}
	data, err := m.MarshalBinary()
	if err != nil {
		return 0, err
	}
	n1 := encodeHead(e, byte(cborTypeByteString), uint64(len(data)))
	n2, _ := e.Write(data)
	return n1 + n2, nil
}

func encodeMarshalerType(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	m, ok := v.Interface().(Marshaler)
	if !ok {
		pv := reflect.New(v.Type())
		pv.Elem().Set(v)
		m = pv.Interface().(Marshaler)
	}
	data, err := m.MarshalCBOR()
	if err != nil {
		return 0, err
	}
	return e.Write(data)
}

func encodeHead(e *encodeState, t byte, n uint64) int {
	if n <= 23 {
		e.WriteByte(t | byte(n))
		return 1
	}
	if n <= math.MaxUint8 {
		e.scratch[0] = t | byte(24)
		e.scratch[1] = byte(n)
		e.Write(e.scratch[:2])
		return 2
	}
	if n <= math.MaxUint16 {
		e.scratch[0] = t | byte(25)
		binary.BigEndian.PutUint16(e.scratch[1:], uint16(n))
		e.Write(e.scratch[:3])
		return 3
	}
	if n <= math.MaxUint32 {
		e.scratch[0] = t | byte(26)
		binary.BigEndian.PutUint32(e.scratch[1:], uint32(n))
		e.Write(e.scratch[:5])
		return 5
	}
	e.scratch[0] = t | byte(27)
	binary.BigEndian.PutUint64(e.scratch[1:], uint64(n))
	e.Write(e.scratch[:9])
	return 9
}

var (
	typeMarshaler       = reflect.TypeOf((*Marshaler)(nil)).Elem()
	typeBinaryMarshaler = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
)

func getEncodeFuncInternal(t reflect.Type) encodeFunc {
	if t.Kind() == reflect.Ptr {
		return getEncodeIndirectValueFunc(t)
	}
	if reflect.PtrTo(t).Implements(typeMarshaler) {
		return encodeMarshalerType
	}
	if reflect.PtrTo(t).Implements(typeBinaryMarshaler) {
		if t == typeTime {
			return encodeTime
		}
		return encodeBinaryMarshalerType
	}
	switch t.Kind() {
	case reflect.Bool:
		return encodeBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return encodeInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return encodeUint
	case reflect.Float32, reflect.Float64:
		return encodeFloat
	case reflect.String:
		return encodeString
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return encodeByteString
		}
		return arrayEncoder{f: getEncodeFunc(t.Elem())}.encodeArray
	case reflect.Map:
		return mapEncoder{kf: getEncodeFunc(t.Key()), ef: getEncodeFunc(t.Elem())}.encodeMap
	case reflect.Struct:
		return encodeStruct
	case reflect.Interface:
		return encodeIntf
	}
	return nil
}

func getEncodeIndirectValueFunc(t reflect.Type) encodeFunc {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	f := getEncodeFunc(t)
	if f == nil {
		return f
	}
	return func(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
		for v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return e.Write(cborNil)
		}
		return f(e, v, opts)
	}
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func cannotFitFloat32(f64 float64) bool {
	f32 := float32(f64)
	return float64(f32) != f64
}

// float32NaNFromReflectValue extracts float32 NaN from reflect.Value while preserving NaN's quiet bit.
func float32NaNFromReflectValue(v reflect.Value) float32 {
	// Keith Randall's workaround for issue https://github.com/golang/go/issues/36400
	p := reflect.New(v.Type())
	p.Elem().Set(v)
	f32 := p.Convert(reflect.TypeOf((*float32)(nil))).Elem().Interface().(float32)
	return f32
}
