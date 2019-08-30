// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

package cbor

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"sort"
	"sync"
)

type encodeFunc func(e *encodeState, v reflect.Value, opts EncOptions) (int, error)

// InvalidValueError is returned by Marshal when attempting to encode an
// invalid value.
type InvalidValueError struct {
	value reflect.Value
}

func (e *InvalidValueError) Error() string {
	return "cbor: invalid value: " + e.value.String()
}

// UnsupportedTypeError is returned by Marshal when attempting to encode an
// unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "cbor: unsupported type: " + e.Type.String()
}

var (
	cborFalse            = []byte{0xf4}
	cborTrue             = []byte{0xf5}
	cborNil              = []byte{0xf6}
	cborNan              = []byte{0xf9, 0x7e, 0x00}
	cborPositiveInfinity = []byte{0xf9, 0x7c, 0x00}
	cborNegativeInfinity = []byte{0xf9, 0xfc, 0x00}
)

// EncOptions specifies encoding options.
type EncOptions struct {
	// Canonical causes map and struct to be encoded in a predictable sequence
	// of bytes by sorting map keys or struct fields according to canonical rules:
	//     - If two keys have different CBOR types, the one with lower value in
	//       numerical order sorts earlier.
	//     - If two keys have different lengths, the shorter one sorts earlier.
	//     - If two keys have the same CBOR type and same length, the one with the
	//       lower value in (byte-wise) lexical order sorts earlier.
	Canonical bool
}

// An encodeState encodes CBOR into a bytes.Buffer.
type encodeState struct {
	bytes.Buffer
	scratch [64]byte
}

// encodeStatePool caches unused encodeState objects for later reuse.
var encodeStatePool = sync.Pool{
	New: func() interface{} {
		return new(encodeState)
	},
}

func newEncodeState() *encodeState {
	return encodeStatePool.Get().(*encodeState)
}

func returnEncodeState(e *encodeState) {
	e.Reset()
	encodeStatePool.Put(e)
}

func (e *encodeState) marshal(v interface{}, opts EncOptions) error {
	if v == nil {
		e.Write(cborNil)
		return nil
	}

	_, err := encode(e, reflect.ValueOf(v), opts)
	return err
}

func encode(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if !v.IsValid() {
		// v is zero value
		e.Write(cborNil)
		return 1, nil
	}

	f := getEncodeFunc(v.Kind())
	if f == nil {
		return 0, &UnsupportedTypeError{v.Type()}
	}

	return f(e, v, opts)
}

func encodeTypeAndAdditionalValue(e *encodeState, t byte, n uint64) int {
	if n <= 23 {
		e.WriteByte(t | byte(n))
		return 1
	} else if n <= math.MaxUint8 {
		e.scratch[0] = t | byte(24)
		e.scratch[1] = byte(n)
		e.Write(e.scratch[:2])
		return 2
	} else if n <= math.MaxUint16 {
		e.scratch[0] = t | byte(25)
		binary.BigEndian.PutUint16(e.scratch[1:], uint16(n))
		e.Write(e.scratch[:3])
		return 3
	} else if n <= math.MaxUint32 {
		e.scratch[0] = t | byte(26)
		binary.BigEndian.PutUint32(e.scratch[1:], uint32(n))
		e.Write(e.scratch[:5])
		return 5
	} else {
		e.scratch[0] = t | byte(27)
		binary.BigEndian.PutUint64(e.scratch[1:], uint64(n))
		e.Write(e.scratch[:9])
		return 9
	}
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
		return encodeTypeAndAdditionalValue(e, byte(cborTypePositiveInt), uint64(i)), nil
	}
	n := v.Int()*(-1) - 1
	return encodeTypeAndAdditionalValue(e, byte(cborTypeNegativeInt), uint64(n)), nil
}

func encodeUint(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	return encodeTypeAndAdditionalValue(e, byte(cborTypePositiveInt), v.Uint()), nil
}

func encodeFloat(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	f64 := v.Float()
	if math.IsNaN(f64) {
		return e.Write(cborNan)
	}
	if math.IsInf(f64, 1) {
		return e.Write(cborPositiveInfinity)
	}
	if math.IsInf(f64, -1) {
		return e.Write(cborNegativeInfinity)
	}
	if v.Kind() == reflect.Float32 {
		f32 := v.Interface().(float32)
		e.scratch[0] = byte(cborTypePrimitives) | byte(26)
		binary.BigEndian.PutUint32(e.scratch[1:], math.Float32bits(f32))
		e.Write(e.scratch[:5])
		return 5, nil
	}
	e.scratch[0] = byte(cborTypePrimitives) | byte(27)
	binary.BigEndian.PutUint64(e.scratch[1:], math.Float64bits(f64))
	e.Write(e.scratch[:9])
	return 9, nil
}

func encodeByteString(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	n1 := encodeTypeAndAdditionalValue(e, byte(cborTypeByteString), uint64(v.Len()))
	if v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			e.WriteByte(byte(v.Index(i).Uint()))
		}
		return n1 + v.Len(), nil
	}
	n2, _ := e.Write(v.Bytes())
	return n1 + n2, nil
}

func encodeString(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	return encodeStringInternal(e, v.String(), opts)
}

func encodeStringInternal(e *encodeState, s string, opts EncOptions) (int, error) {
	n1 := encodeTypeAndAdditionalValue(e, byte(cborTypeTextString), uint64(len(s)))
	n2, _ := e.WriteString(s)
	return n1 + n2, nil
}

func encodeSlice(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if v.IsNil() {
		return e.Write(cborNil)
	}
	return encodeArray(e, v, opts)
}

func encodeArray(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if v.Type().Elem().Kind() == reflect.Uint8 {
		return encodeByteString(e, v, opts)
	}

	if v.Len() == 0 {
		return encodeTypeAndAdditionalValue(e, byte(cborTypeArray), uint64(0)), nil
	}

	n := encodeTypeAndAdditionalValue(e, byte(cborTypeArray), uint64(v.Len()))
	for i := 0; i < v.Len(); i++ {
		n1, err := encode(e, v.Index(i), opts)
		if err != nil {
			return 0, err
		}
		n += n1
	}
	return n, nil
}

func encodeMap(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if v.IsNil() {
		return e.Write(cborNil)
	}
	if v.Len() == 0 {
		return encodeTypeAndAdditionalValue(e, byte(cborTypeMap), uint64(0)), nil
	}
	if opts.Canonical {
		return encodeMapCanonical(e, v, opts)
	}
	n := encodeTypeAndAdditionalValue(e, byte(cborTypeMap), uint64(v.Len()))
	iter := v.MapRange()
	for iter.Next() {
		kn, err := encode(e, iter.Key(), opts)
		if err != nil {
			return 0, err
		}
		en, err := encode(e, iter.Value(), opts)
		if err != nil {
			return 0, err
		}
		n += kn + en
	}
	return n, nil
}

type pair struct {
	keyCBORData, pairCBORData []byte
	keyLen, pairLen           int
}

type byCanonical struct {
	pairs []pair
}

func (v byCanonical) Len() int {
	return len(v.pairs)
}

func (v byCanonical) Swap(i, j int) {
	v.pairs[i], v.pairs[j] = v.pairs[j], v.pairs[i]
}

func (v byCanonical) Less(i, j int) bool {
	return bytes.Compare(v.pairs[i].keyCBORData, v.pairs[j].keyCBORData) <= 0
}

var byCanonicalPool = sync.Pool{}

func newByCanonical(length int) *byCanonical {
	v := byCanonicalPool.Get()
	if v == nil {
		return &byCanonical{pairs: make([]pair, 0, length)}
	}
	s := v.(*byCanonical)
	if cap(s.pairs) < length {
		// byCanonical object from the pool does not have enough capacity.
		// Return it back to the pool and create a new one.
		byCanonicalPool.Put(s)
		return &byCanonical{pairs: make([]pair, 0, length)}
	}
	s.pairs = s.pairs[:0]
	return s
}

func returnByCanonical(s *byCanonical) {
	s.pairs = s.pairs[:0]
	byCanonicalPool.Put(s)
}

func encodeMapCanonical(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	pairEncodeState := newEncodeState() // accumulated cbor encoded map key-value pairs
	pairs := newByCanonical(v.Len())    // for sorting keys

	iter := v.MapRange()
	for iter.Next() {
		n1, err := encode(pairEncodeState, iter.Key(), opts)
		if err != nil {
			returnEncodeState(pairEncodeState)
			returnByCanonical(pairs)
			return 0, err
		}
		n2, err := encode(pairEncodeState, iter.Value(), opts)
		if err != nil {
			returnEncodeState(pairEncodeState)
			returnByCanonical(pairs)
			return 0, err
		}
		pairs.pairs = append(pairs.pairs, pair{keyLen: n1, pairLen: n1 + n2})
	}
	b := pairEncodeState.Bytes()
	for i, offset := 0, 0; i < len(pairs.pairs); i++ {
		pairs.pairs[i].keyCBORData = b[offset : offset+pairs.pairs[i].keyLen]
		pairs.pairs[i].pairCBORData = b[offset : offset+pairs.pairs[i].pairLen]
		offset += pairs.pairs[i].pairLen
	}

	sort.Sort(pairs)

	n := encodeTypeAndAdditionalValue(e, byte(cborTypeMap), uint64(len(pairs.pairs)))
	for i := 0; i < len(pairs.pairs); i++ {
		n1, _ := e.Write(pairs.pairs[i].pairCBORData)
		n += n1
	}

	returnEncodeState(pairEncodeState)
	returnByCanonical(pairs)
	return n, nil
}

func encodeStruct(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	flds := getStructFields(v.Type(), opts.Canonical)

	kve := newEncodeState() // encode key-value pairs based on struct field tag options
	kvcount := 0
	for _, f := range flds {
		fv, err := fieldByIndex(v, f.idx)
		if err != nil {
			returnEncodeState(kve)
			return 0, err
		}
		if !fv.IsValid() || (f.omitempty && isEmptyValue(fv)) {
			continue
		}
		encodeStringInternal(kve, f.name, opts)
		_, err = encode(kve, fv, opts)
		if err != nil {
			returnEncodeState(kve)
			return 0, err
		}
		kvcount++
	}

	n := encodeTypeAndAdditionalValue(e, byte(cborTypeMap), uint64(kvcount))
	n1, _ := e.Write(kve.Bytes())

	returnEncodeState(kve)
	return n + n1, nil
}

func encodePtr(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if v.IsNil() {
		return e.Write(cborNil)
	}
	return encode(e, v.Elem(), opts)
}

func encodeIntf(e *encodeState, v reflect.Value, opts EncOptions) (int, error) {
	if v.IsNil() {
		return e.Write(cborNil)
	}
	return encode(e, v.Elem(), opts)
}

func getEncodeFunc(k reflect.Kind) encodeFunc {
	switch k {
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
	case reflect.Array:
		return encodeArray
	case reflect.Slice:
		return encodeSlice
	case reflect.Map:
		return encodeMap
	case reflect.Struct:
		return encodeStruct
	case reflect.Ptr:
		return encodePtr
	case reflect.Interface:
		return encodeIntf
	default:
		return nil
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
// Marshal supports format string stored under the "cbor" key in the struct
// field's tag.  CBOR format string can specify the name of the field, "omitempty"
// option, and special case "-" for field omission.
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
//
// CBOR cannot represent cyclic data structures and Marshal does not handle them.
//
// CTAP2 canonical CBOR encoding uses the following rules:
//
//     1. Integers must be encoded as small as possible.
//     2. The representations of any floating-point values are not changed.
//     3. The expression of lengths in major types 2 through 5 must be as short as possible.
//     4. Indefinite-length items must be made into definite-length items.
//     5. The keys in every map must be sorted lowest value to highest.
//     6. Tags must not be present.
//
// Canonical CBOR encoding specified in RFC 7049 section 3.9 consists of CTAP2
// canonical CBOR encoding rules 1, 3, 4, and 5.
//
// Marshal supports RFC 7049 and CTAP2 canonical CBOR encoding.
func Marshal(v interface{}, encOpts EncOptions) ([]byte, error) {
	e := newEncodeState()

	err := e.marshal(v, encOpts)
	if err != nil {
		returnEncodeState(e)
		return nil, err
	}

	buf := make([]byte, e.Len())
	copy(buf, e.Bytes())

	returnEncodeState(e)
	return buf, nil
}
