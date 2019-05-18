// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

package cbor

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

type cborType uint8

const (
	cborTypePositiveInt cborType = 0x00
	cborTypeNegativeInt cborType = 0x20
	cborTypeByteString  cborType = 0x40
	cborTypeTextString  cborType = 0x60
	cborTypeArray       cborType = 0x80
	cborTypeMap         cborType = 0xA0
	cborTypeTag         cborType = 0xC0
	cborTypePrimitives  cborType = 0xE0
)

func (t cborType) String() string {
	switch t {
	case cborTypePositiveInt:
		return "positive integer"
	case cborTypeNegativeInt:
		return "negative integer"
	case cborTypeByteString:
		return "byte string"
	case cborTypeTextString:
		return "UTF-8 text string"
	case cborTypeArray:
		return "array"
	case cborTypeMap:
		return "map"
	case cborTypeTag:
		return "tag"
	case cborTypePrimitives:
		return "primitives"
	default:
		return "Invalid type " + strconv.Itoa(int(t))
	}
}

// InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "cbor: Unmarshal(nil)"
	}
	if e.Type.Kind() != reflect.Ptr {
		return "cbor: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "cbor: Unmarshal(nil " + e.Type.String() + ")"
}

// UnmarshalTypeError describes a CBOR value that was not appropriate for a Go type.
type UnmarshalTypeError struct {
	Value  string       // description of CBOR value
	Type   reflect.Type // type of Go value it could not be assigned to
	Struct string       // struct type containing the field
	Field  string       // name of the field holding the Go value
	errMsg string       // additional error message (optional)
}

func (e *UnmarshalTypeError) Error() string {
	var s string
	if e.Struct != "" || e.Field != "" {
		s = "cbor: cannot unmarshal " + e.Value + " into Go struct field " + e.Struct + "." + e.Field + " of type " + e.Type.String()
	} else {
		s = "cbor: cannot unmarshal " + e.Value + " into Go value of type " + e.Type.String()
	}
	if e.errMsg != "" {
		s += " (" + e.errMsg + ")"
	}
	return s
}

type decodeState struct {
	data   []byte
	offset int // next read offset in data
	err    error
}

func (d *decodeState) reset(data []byte) {
	d.data = data
	d.offset = 0
	d.err = nil
}

func (d *decodeState) value(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	if _, err := Valid(d.data[d.offset:]); err != nil {
		return err
	}

	return d.parse(rv.Elem())
}

// skip moves data offset to the next item.  skip assumes data is well-formed,
// and does not perform bounds checking.
func (d *decodeState) skip() {
	t := cborType(d.data[d.offset] & 0xE0)
	ai := d.data[d.offset] & 0x1F
	val := uint64(ai)
	d.offset++

	switch ai {
	case 24:
		val = uint64(d.data[d.offset])
		d.offset++
	case 25:
		val = uint64(binary.BigEndian.Uint16(d.data[d.offset : d.offset+2]))
		d.offset += 2
	case 26:
		val = uint64(binary.BigEndian.Uint32(d.data[d.offset : d.offset+4]))
		d.offset += 4
	case 27:
		val = binary.BigEndian.Uint64(d.data[d.offset : d.offset+8])
		d.offset += 8
	}

	if ai == 31 {
		switch t {
		case cborTypeByteString, cborTypeTextString, cborTypeArray, cborTypeMap:
			for true {
				if d.data[d.offset] == 0xFF {
					d.offset++
					return
				}
				d.skip()
			}
		}
	}

	switch t {
	case cborTypeByteString, cborTypeTextString:
		d.offset += int(val)
	case cborTypeArray:
		for i := 0; i < int(val); i++ {
			d.skip()
		}
	case cborTypeMap:
		for i := 0; i < int(val)*2; i++ {
			d.skip()
		}
	case cborTypeTag:
		d.skip()
	}
}

func (d *decodeState) parse(v reflect.Value) (err error) {
	if len(d.data) == d.offset {
		return io.EOF
	}

	if d.data[0] == 0xf6 || d.data[0] == 0xf7 { // CBOR null and CBOR undefined
		d.offset++
		return fillNil(cborTypePrimitives, v)
	}

	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			if !v.CanSet() {
				return errors.New("cbor: cannot set new value for " + v.Type().String())
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	// Process byte/text string.
	t := cborType(d.data[d.offset] & 0xE0)
	if t == cborTypeByteString {
		b, err := d.parseByteString()
		if err != nil {
			return err
		}
		return fillByteString(t, b, v)
	} else if t == cborTypeTextString {
		b, err := d.parseTextString()
		if err != nil {
			return err
		}
		return fillTextString(t, b, v)
	}

	t, ai, val, err := d.getHeader()
	if err != nil {
		return err
	}

	// Process other types.
	switch t {
	case cborTypePositiveInt:
		return fillPositiveInt(t, val, v)
	case cborTypeNegativeInt:
		nValue := int64(-1) ^ int64(val)
		return fillNegativeInt(t, nValue, v)
	case cborTypeTag:
		return d.parse(v)
	case cborTypePrimitives:
		if ai < 20 {
			return fillPositiveInt(t, uint64(ai), v)
		}
		switch ai {
		case 20, 21:
			return fillBool(t, ai == 21, v)
		case 24:
			return fillPositiveInt(t, uint64(val), v)
		case 25:
			f := uint16ToFloat64(uint16(val))
			return fillFloat(t, f, v)
		case 26:
			f := float64(math.Float32frombits(uint32(val)))
			return fillFloat(t, f, v)
		case 27:
			f := math.Float64frombits(val)
			return fillFloat(t, f, v)
		}
	case cborTypeArray:
		valInt := int(val)
		if valInt < 0 || uint64(valInt) != val {
			// Detect integer overflow
			return errors.New("cbor: " + t.String() + " length " + strconv.FormatUint(val, 10) + " is too large, causing integer overflow")
		}
		count := valInt
		if ai == 31 {
			count = -1
		}
		if isNilInterface(v) {
			// >100% improvement in ns/op and less allocs/op with parseArrayInterface(), compared with parseArray().
			arr, err := d.parseArrayInterface(t, count)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(arr))
			return nil
		} else if v.Kind() == reflect.Slice {
			return d.parseSlice(t, count, v)
		} else if v.Kind() == reflect.Array {
			return d.parseArray(t, count, v)
		} else {
			hasSize := count >= 0
			for i := 0; (hasSize && i < count) || (!hasSize && !d.foundBreak()); i++ {
				d.skip()
			}
			return &UnmarshalTypeError{Value: t.String(), Type: v.Type()}
		}
	case cborTypeMap:
		valInt := int(val)
		if valInt < 0 || uint64(valInt) != val {
			// Detect integer overflow
			return errors.New("cbor: " + t.String() + " length " + strconv.FormatUint(val, 10) + " is too large, causing integer overflow")
		}
		count := valInt
		if ai == 31 {
			count = -1
		}
		if isNilInterface(v) {
			// >100% improvement in ns/op and less allocs/op with parseMapInterface(), compared with parseMap().
			m, err := d.parseMapInterface(t, count)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(m))
			return nil
		} else if v.Kind() == reflect.Struct {
			return d.parseStruct(t, count, v)
		} else if v.Kind() == reflect.Map {
			return d.parseMap(t, count, v)
		} else {
			hasSize := count >= 0
			for i := 0; (hasSize && i < count*2) || (!hasSize && !d.foundBreak()); i++ {
				d.skip()
			}
			return &UnmarshalTypeError{Value: t.String(), Type: v.Type()}
		}
	}
	return nil
}

func (d *decodeState) parseInterface() (_ interface{}, err error) {
	if len(d.data) == d.offset {
		return nil, io.EOF
	}

	if d.data[0] == 0xf6 || d.data[0] == 0xf7 { // CBOR null and CBOR undefined
		d.offset++
		return nil, nil
	}

	// Process byte/text string.
	t := cborType(d.data[d.offset] & 0xE0)
	if t == cborTypeByteString {
		return d.parseByteString()
	} else if t == cborTypeTextString {
		b, err := d.parseTextString()
		if err != nil {
			return nil, err
		}
		return string(b), nil
	}

	t, ai, val, err := d.getHeader()
	if err != nil {
		return nil, err
	}

	// Process other types.
	switch t {
	case cborTypePositiveInt:
		return val, nil
	case cborTypeNegativeInt:
		nValue := int64(-1) ^ int64(val)
		return nValue, nil
	case cborTypeTag:
		return d.parseInterface()
	case cborTypePrimitives:
		if ai < 20 {
			return uint64(ai), nil
		}
		switch ai {
		case 20, 21:
			return (ai == 21), nil
		case 24:
			return uint64(val), nil
		case 25:
			f := uint16ToFloat64(uint16(val))
			return f, nil
		case 26:
			f := float64(math.Float32frombits(uint32(val)))
			return f, nil
		case 27:
			f := math.Float64frombits(val)
			return f, nil
		}
	case cborTypeArray:
		valInt := int(val)
		if valInt < 0 || uint64(valInt) != val {
			// Detect integer overflow
			return nil, errors.New("cbor: " + t.String() + " length " + strconv.FormatUint(val, 10) + " is too large, causing integer overflow")
		}
		count := valInt
		if ai == 31 {
			count = -1
		}
		return d.parseArrayInterface(t, count)
	case cborTypeMap:
		valInt := int(val)
		if valInt < 0 || uint64(valInt) != val {
			// Detect integer overflow
			return nil, errors.New("cbor: " + t.String() + " length " + strconv.FormatUint(val, 10) + " is too large, causing integer overflow")
		}
		count := valInt
		if ai == 31 {
			count = -1
		}
		return d.parseMapInterface(t, count)
	}
	panic("cbor: parseInterface() of type " + t.String() + " is not implemented")
}

// parseByteString parses CBOR encoded byte string.  It returns a byte slice
// pointing to a copy of parsed data.
func (d *decodeState) parseByteString() ([]byte, error) {
	val, isCopy, err := d.parseStringBuf(nil)
	if err != nil {
		return nil, err
	}

	if !isCopy {
		// Make a copy of val so that GC can collect underlying data val points to.
		copyVal := make([]byte, len(val))
		copy(copyVal, val)
		return copyVal, nil
	}
	return val, nil
}

// parseTextString parses CBOR encoded text string.  It does not return a string
// to prevent creating an extra copy of string.  Caller should wrap returned
// byte slice as string when needed.
//
// parseStruct() uses parseTextString() to improve memory and performance,
// compared with using parse(reflect.Value).  parse(reflect.Value) sets
// reflect.Value with parsed string, while parseTextString() returns parsed string.
func (d *decodeState) parseTextString() ([]byte, error) {
	val, _, err := d.parseStringBuf(nil)
	if err != nil {
		return nil, err
	}

	if !utf8.Valid(val) {
		return nil, &SemanticError{"cbor: invalid UTF-8 string"}
	}

	return val, nil
}

func (d *decodeState) parseStringBuf(p []byte) (_ []byte, isCopy bool, err error) {
	t, ai, val, err := d.getHeader()
	if err != nil {
		return nil, false, err
	}

	if t != cborTypeByteString && t != cborTypeTextString {
		panic("cbor: expect byte/text string data, got " + t.String())
	}

	if ai == 31 {
		// Process indefinite length string.
		if p == nil {
			p = make([]byte, 0, 64)
		}
		for !d.foundBreak() {
			if p, _, err = d.parseStringBuf(p); err != nil {
				return nil, false, err
			}
		}
		return p, true, nil
	}

	// Process definite length string.
	valInt := int(val)
	if valInt < 0 || uint64(valInt) != val {
		// Detect integer overflow
		return nil, false, errors.New("cbor: " + t.String() + " length " + strconv.FormatUint(val, 10) + " is too large, causing integer overflow")
	}
	if len(d.data)-d.offset < valInt {
		return nil, false, io.ErrUnexpectedEOF
	}
	oldOff, newOff := d.offset, d.offset+valInt
	d.offset = newOff

	if p != nil {
		p = append(p, d.data[oldOff:newOff]...)
		return p, true, nil
	}
	return d.data[oldOff:newOff], false, nil
}

func (d *decodeState) parseArrayInterface(t cborType, count int) (_ []interface{}, err error) {
	hasSize := count >= 0
	if count == -1 {
		count = d.numOfItemsUntilBreak() // peek ahead to get array size to preallocate slice for better performance
	}
	v := make([]interface{}, count)
	var e interface{}
	for i := 0; (hasSize && i < count) || (!hasSize && !d.foundBreak()); i++ {
		if e, err = d.parseInterface(); err != nil {
			return nil, err
		}
		v[i] = e
	}
	return v, nil
}

func (d *decodeState) parseSlice(t cborType, count int, v reflect.Value) error {
	hasSize := count >= 0
	if count == -1 {
		count = d.numOfItemsUntilBreak() // peek ahead to get array size to preallocate slice for better performance
	}
	if count == 0 {
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}
	if v.IsNil() || v.Cap() < count {
		v.Set(reflect.MakeSlice(v.Type(), count, count))
	}
	v.SetLen(count)
	for i := 0; (hasSize && i < count) || (!hasSize && !d.foundBreak()); i++ {
		if err := d.parse(v.Index(i)); err != nil {
			if _, ok := err.(*UnmarshalTypeError); ok {
				for i++; (hasSize && i < count) || (!hasSize && !d.foundBreak()); i++ {
					d.skip()
				}
			}
			return err
		}
	}
	return nil
}

func (d *decodeState) parseArray(t cborType, count int, v reflect.Value) error {
	hasSize := count >= 0
	if count == -1 {
		count = d.numOfItemsUntilBreak()
	}
	i := 0
	for ; i < v.Len() && ((hasSize && i < count) || (!hasSize && !d.foundBreak())); i++ {
		if err := d.parse(v.Index(i)); err != nil {
			if _, ok := err.(*UnmarshalTypeError); ok {
				for i++; (hasSize && i < count) || (!hasSize && !d.foundBreak()); i++ {
					d.skip()
				}
			}
			return err
		}
	}
	// Set remaining Go array elements to zero values.
	if i < v.Len() {
		zeroV := reflect.Zero(v.Type().Elem())
		for ; i < v.Len(); i++ {
			v.Index(i).Set(zeroV)
		}
	}
	// Skip remaining CBOR array elements
	for ; (hasSize && i < count) || (!hasSize && !d.foundBreak()); i++ {
		d.skip()
	}
	return nil
}

func (d *decodeState) parseMapInterface(t cborType, count int) (_ map[interface{}]interface{}, err error) {
	m := make(map[interface{}]interface{})
	hasSize := count >= 0
	var k, e interface{}
	for i := 0; (hasSize && i < count) || (!hasSize && !d.foundBreak()); i++ {
		if k, err = d.parseInterface(); err != nil {
			return nil, err
		}
		if e, err = d.parseInterface(); err != nil {
			return nil, err
		}
		m[k] = e
	}
	return m, nil
}

func (d *decodeState) parseMap(t cborType, count int, v reflect.Value) error {
	if v.IsNil() {
		mapsize := count
		if mapsize < 0 {
			mapsize = 0
		}
		v.Set(reflect.MakeMapWithSize(v.Type(), mapsize))
	}
	hasSize := count >= 0
	keyType, eleType := v.Type().Key(), v.Type().Elem()
	reuseKey, reuseEle := isImmutableKind(keyType.Kind()), isImmutableKind(eleType.Kind())
	var keyValue, eleValue, zeroKeyValue, zeroEleValue reflect.Value
	for i := 0; (hasSize && i < count) || (!hasSize && !d.foundBreak()); i++ {
		if !keyValue.IsValid() {
			keyValue = reflect.New(keyType).Elem()
		} else if !reuseKey {
			if !zeroKeyValue.IsValid() {
				zeroKeyValue = reflect.Zero(keyType)
			}
			keyValue.Set(zeroKeyValue)
		}
		if err := d.parse(keyValue); err != nil {
			if _, ok := err.(*UnmarshalTypeError); ok {
				for i = i*2 + 1; (hasSize && i < count*2) || (!hasSize && !d.foundBreak()); i++ {
					d.skip()
				}
			}
			return err
		}

		if !eleValue.IsValid() {
			eleValue = reflect.New(eleType).Elem()
		} else if !reuseEle {
			if !zeroEleValue.IsValid() {
				zeroEleValue = reflect.Zero(eleType)
			}
			eleValue.Set(zeroEleValue)
		}
		if err := d.parse(eleValue); err != nil {
			if _, ok := err.(*UnmarshalTypeError); ok {
				for i = i*2 + 2; (hasSize && i < count*2) || (!hasSize && !d.foundBreak()); i++ {
					d.skip()
				}
			}
			return err
		}

		v.SetMapIndex(keyValue, eleValue)
	}
	return nil
}

func (d *decodeState) parseStruct(t cborType, count int, v reflect.Value) error {
	flds := getStructFields(v.Type(), false)

	hasSize := count >= 0
	for i := 0; (hasSize && i < count) || (!hasSize && !d.foundBreak()); i++ {
		t := cborType(d.data[d.offset] & 0xE0)
		if t != cborTypeTextString {
			if d.err == nil {
				d.err = &UnmarshalTypeError{Value: t.String(), Type: reflect.TypeOf(""), errMsg: "map key is of type " + t.String() + " and cannot be used to match struct " + v.Type().String() + " field name"}
			}
			d.skip() // skip key
			d.skip() // skip value
			continue
		}
		keyBytes, err := d.parseTextString()
		if err != nil {
			return err
		}

		var f *field
		for i := 0; i < len(flds); i++ {
			// Find field with exact match
			if len(flds[i].name) == len(keyBytes) && flds[i].name == string(keyBytes) {
				f = &flds[i]
				break
			}
		}
		if f == nil {
			keyString := string(keyBytes)
			for i := 0; i < len(flds); i++ {
				// Find field with case-insensitive match
				if len(flds[i].name) == len(keyString) && strings.EqualFold(flds[i].name, keyString) {
					f = &flds[i]
					break
				}
			}
		}
		if f == nil {
			d.skip()
			continue
		}
		// reflect.Value.FieldByIndex() panics at nil pointer to unexported
		// anonymous field.  fieldByIndex() returns error.
		fv, err := fieldByIndex(v, f.idx)
		if err != nil {
			return err
		}
		if !fv.IsValid() || !fv.CanSet() {
			d.skip()
			continue
		}
		for fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				if !fv.CanSet() {
					d.skip()
					continue
				}
				fv.Set(reflect.New(fv.Type().Elem()))
			}
			fv = fv.Elem()
		}
		if err := d.parse(fv); err != nil {
			typeError, ok := err.(*UnmarshalTypeError)
			if !ok {
				return err
			}
			typeError.Struct = v.Type().String()
			typeError.Field = string(keyBytes)
			if d.err == nil {
				d.err = typeError
			}
		}
	}
	return d.err
}

func (d *decodeState) getHeader() (t cborType, ai byte, val uint64, err error) {
	if len(d.data)-d.offset < 1 {
		err = io.ErrUnexpectedEOF
		return
	}
	t = cborType(d.data[d.offset] & 0xE0)
	ai = d.data[d.offset] & 0x1F
	val = uint64(ai)
	d.offset++

	switch ai {
	case 24:
		if len(d.data)-d.offset < 1 {
			err = io.ErrUnexpectedEOF
			return
		}
		val = uint64(d.data[d.offset])
		d.offset++
	case 25:
		if len(d.data)-d.offset < 2 {
			err = io.ErrUnexpectedEOF
			return
		}
		val = uint64(binary.BigEndian.Uint16(d.data[d.offset : d.offset+2]))
		d.offset += 2
	case 26:
		if len(d.data)-d.offset < 4 {
			err = io.ErrUnexpectedEOF
			return
		}
		val = uint64(binary.BigEndian.Uint32(d.data[d.offset : d.offset+4]))
		d.offset += 4
	case 27:
		if len(d.data)-d.offset < 8 {
			err = io.ErrUnexpectedEOF
			return
		}
		val = binary.BigEndian.Uint64(d.data[d.offset : d.offset+8])
		d.offset += 8
	}
	return
}

func (d *decodeState) numOfItemsUntilBreak() int {
	savedOff := d.offset
	i := 0
	for !d.foundBreak() {
		d.skip()
		i++
	}
	d.offset = savedOff
	return i
}

func (d *decodeState) foundBreak() bool {
	if len(d.data) == d.offset {
		panic("cbor: unexpected EOF while searching for \"break\" code")
	}
	if d.data[d.offset] == 0xFF {
		d.offset++
		return true
	}
	return false
}

func fillNil(t cborType, v reflect.Value) error {
	if isNilInterface(v) {
		return nil
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.Interface, reflect.Ptr:
		v.Set(reflect.Zero(v.Type()))
		return nil
	default:
		return nil
	}
}

func fillPositiveInt(t cborType, val uint64, v reflect.Value) error {
	if isNilInterface(v) {
		v.Set(reflect.ValueOf(val))
		return nil
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val > math.MaxInt64 {
			return &UnmarshalTypeError{Value: t.String(), Type: v.Type(), errMsg: strconv.FormatUint(val, 10) + " overflows " + v.Type().String()}
		}
		if v.OverflowInt(int64(val)) {
			return &UnmarshalTypeError{Value: t.String(), Type: v.Type(), errMsg: strconv.FormatUint(val, 10) + " overflows " + v.Type().String()}
		}
		v.SetInt(int64(val))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.OverflowUint(val) {
			return &UnmarshalTypeError{Value: t.String(), Type: v.Type(), errMsg: strconv.FormatUint(val, 10) + " overflows " + v.Type().String()}
		}
		v.SetUint(val)
		return nil
	default:
		return &UnmarshalTypeError{Value: t.String(), Type: v.Type()}
	}
}

func fillNegativeInt(t cborType, val int64, v reflect.Value) error {
	if isNilInterface(v) {
		v.Set(reflect.ValueOf(val))
		return nil
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.OverflowInt(val) {
			return &UnmarshalTypeError{Value: t.String(), Type: v.Type(), errMsg: strconv.FormatInt(val, 10) + " overflows " + v.Type().String()}
		}
		v.SetInt(val)
		return nil
	default:
		return &UnmarshalTypeError{Value: t.String(), Type: v.Type()}
	}
}

func fillBool(t cborType, val bool, v reflect.Value) error {
	if isNilInterface(v) {
		v.Set(reflect.ValueOf(val))
		return nil
	}
	if v.Kind() == reflect.Bool {
		v.SetBool(val)
		return nil
	}
	return &UnmarshalTypeError{Value: t.String(), Type: v.Type()}
}

func fillFloat(t cborType, val float64, v reflect.Value) error {
	if isNilInterface(v) {
		v.Set(reflect.ValueOf(val))
		return nil
	}
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		if v.OverflowFloat(val) {
			return &UnmarshalTypeError{Value: t.String(), Type: v.Type(), errMsg: strconv.FormatFloat(val, 'E', -1, 64) + " overflows " + v.Type().String()}
		}
		v.SetFloat(val)
		return nil
	default:
		return &UnmarshalTypeError{Value: t.String(), Type: v.Type()}
	}
}

func fillByteString(t cborType, val []byte, v reflect.Value) error {
	if isNilInterface(v) {
		v.Set(reflect.ValueOf(val))
		return nil
	}
	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
		v.SetBytes(val)
		return nil
	}
	return &UnmarshalTypeError{Value: t.String(), Type: v.Type()}
}

func fillTextString(t cborType, val []byte, v reflect.Value) error {
	if isNilInterface(v) {
		v.Set(reflect.ValueOf(string(val)))
		return nil
	}
	if v.Kind() == reflect.String {
		v.SetString(string(val))
		return nil
	}
	return &UnmarshalTypeError{Value: t.String(), Type: v.Type()}
}

func uint16ToFloat64(num uint16) float64 {
	bits := uint32(num)

	sign := bits >> 15
	exp := bits >> 10 & 0x1F
	frac := bits & 0x3FF

	switch exp {
	case 0:
	case 0x1F:
		exp = 0xFF
	default:
		exp = exp - 15 + 127
	}
	bits = sign<<31 | exp<<23 | frac<<13

	f := math.Float32frombits(bits)
	return float64(f)
}

func isNilInterface(v reflect.Value) bool {
	return v.Kind() == reflect.Interface && v.NumMethod() == 0 && v.IsNil()
}

func isImmutableKind(k reflect.Kind) bool {
	switch k {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}

// Unmarshal parses the CBOR-encoded data and stores the result in the value
// pointed to by v.  If v is nil or not a pointer, Unmarshal returns an error.
//
// Unmarshal uses the inverse of the encodings that Marshal uses, allocating
// maps, slices, and pointers as necessary, with the following additional rules:
//
// To unmarshal CBOR into a pointer, Unmarshal first handles the case of the
// CBOR being the CBOR literal null.  In that case, Unmarshal sets the pointer
// to nil.  Otherwise, Unmarshal unmarshals the CBOR into the value pointed at
// by the pointer.  If the pointer is nil, Unmarshal allocates a new value for
// it to point to.
//
// To unmarshal CBOR into an interface value, Unmarshal stores one of these in
// the interface value:
//
//     bool, for CBOR booleans
//     uint64, for CBOR positive integers
//     int64, for CBOR negative integers
//     float64, for CBOR floating points
//     []byte, for CBOR byte strings
//     string, for CBOR text strings
//     []interface{}, for CBOR arrays
//     map[interface{}]interface{}, for CBOR maps
//     nil, for CBOR null
//
// To unmarshal a CBOR array into a slice, Unmarshal allocates a new slice only
// if the CBOR array is empty or slice capacity is less than CBOR array length.
// Otherwise Unmarshal reuses the existing slice, overwriting existing elements.
// Unmarshal sets the slice length to CBOR array length.
//
// To ummarshal a CBOR array into a Go array, Unmarshal decodes CBOR array
// elements into corresponding Go array elements.  If the Go array is smaller
// than the CBOR array, the additional CBOR array elements are discarded.  If
// the CBOR array is smaller than the Go array, the additional Go array elements
// are set to zero values.
//
// To unmarshal a CBOR map into a map, Unmarshal allocates a new map only if the
// map is nil.  Otherwise Unmarshal reuses the existing map, keeping existing
// entries.  Unmarshal stores key-value pairs from the CBOR map into Go map.
//
// To unmarshal a CBOR map into a struct, Unmarshal matches CBOR map keys to the
// keys used by Marshal (either the struct field name or its tag), preferring an
// exact match but also accepting a case-insensitive match.  Map keys which
// don't have a corresponding struct field are ignored.
//
// If a CBOR value is not appropriate for a given Go type, or if a CBOR number
// overflows the Go type, Unmarshal skips that field and completes the
// unmarshalling as best as it can.  If no more serious errors are encountered,
// unmarshal returns an UnmarshalTypeError describing the earliest such error.
// In any case, it's not guaranteed that all the remaining fields following the
// problematic one will be unmarshaled into the target object.
//
// The CBOR null value unmarshals into a slice/map/pointer/interface by setting
// that Go value to nil.  Because null is often used to mean "not present",
// unmarshalling a CBOR null into any other Go type has no effect on the value
// produces no error.
//
// Unmarshal ignores CBOR tag data and parses tagged data following CBOR tag.
func Unmarshal(data []byte, v interface{}) error {
	d := decodeState{data: data}
	return d.value(v)
}
