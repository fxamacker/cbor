// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	decodingStructTypeCache sync.Map // map[reflect.Type]decodingStructType
	encodingStructTypeCache sync.Map // map[reflect.Type]encodingStructType
	encodeFuncCache         sync.Map // map[reflect.Type]encodeFunc
)

type decodingStructType struct {
	fields  fields
	toArray bool
}

func getDecodingStructType(t reflect.Type) decodingStructType {
	if v, _ := decodingStructTypeCache.Load(t); v != nil {
		return v.(decodingStructType)
	}

	flds, structOptions := getFields(t)

	toArray := hasToArrayOption(structOptions)

	for i := 0; i < len(flds); i++ {
		flds[i].isUnmarshaler = implementsUnmarshaler(flds[i].typ)
	}

	structType := decodingStructType{fields: flds, toArray: toArray}
	decodingStructTypeCache.Store(t, structType)
	return structType
}

type encodingStructType struct {
	fields                  fields
	bytewiseCanonicalFields fields
	lenFirstCanonicalFields fields
	err                     error
	toArray                 bool
	omitEmpty               bool
	hasAnonymousField       bool
}

func (st encodingStructType) getFields(opts EncOptions) fields {
	switch opts.Sort {
	case SortLengthFirst:
		return st.lenFirstCanonicalFields
	case SortBytewiseLexical:
		return st.bytewiseCanonicalFields
	}
	return st.fields
}

type byBytewiseFields struct {
	fields
}

func (x byBytewiseFields) Less(i, j int) bool {
	return bytes.Compare(x.fields[i].cborName, x.fields[j].cborName) <= 0
}

type byLengthFirstFields struct {
	fields
}

func (x byLengthFirstFields) Less(i, j int) bool {
	if len(x.fields[i].cborName) != len(x.fields[j].cborName) {
		return len(x.fields[i].cborName) < len(x.fields[j].cborName)
	}
	return bytes.Compare(x.fields[i].cborName, x.fields[j].cborName) <= 0
}

func getEncodingStructType(t reflect.Type) encodingStructType {
	if v, _ := encodingStructTypeCache.Load(t); v != nil {
		return v.(encodingStructType)
	}

	flds, structOptions := getFields(t)

	if hasToArrayOption(structOptions) {
		return getEncodingStructToArrayType(t, flds)
	}

	var err error
	var omitEmpty bool
	var hasAnonymousField bool
	var hasKeyAsInt bool
	var hasKeyAsStr bool
	e := getEncodeState()
	for i := 0; i < len(flds); i++ {
		// Get field's encodeFunc
		flds[i].ef = getEncodeFunc(flds[i].typ)
		if flds[i].ef == nil {
			err = &UnsupportedTypeError{t}
			break
		}

		// Encode field name
		if flds[i].keyAsInt {
			nameAsInt, numErr := strconv.Atoi(flds[i].name)
			if numErr != nil {
				err = numErr
				break
			}
			if nameAsInt >= 0 {
				encodeHead(e, byte(cborTypePositiveInt), uint64(nameAsInt))
			} else {
				n := nameAsInt*(-1) - 1
				encodeHead(e, byte(cborTypeNegativeInt), uint64(n))
			}
			flds[i].cborName = make([]byte, e.Len())
			copy(flds[i].cborName, e.Bytes())
			e.Reset()

			hasKeyAsInt = true
		} else {
			encodeHead(e, byte(cborTypeTextString), uint64(len(flds[i].name)))
			flds[i].cborName = make([]byte, e.Len()+len(flds[i].name))
			n := copy(flds[i].cborName, e.Bytes())
			copy(flds[i].cborName[n:], flds[i].name)
			e.Reset()

			hasKeyAsStr = true
		}

		// Check if field is from embedded struct
		if len(flds[i].idx) > 1 {
			hasAnonymousField = true
		}

		// Check if field can be omitted when empty
		if flds[i].omitEmpty {
			omitEmpty = true
		}
	}
	putEncodeState(e)

	if err != nil {
		structType := encodingStructType{err: err}
		encodingStructTypeCache.Store(t, structType)
		return structType
	}

	// Sort fields by canonical order
	bytewiseCanonicalFields := make(fields, len(flds))
	copy(bytewiseCanonicalFields, flds)
	sort.Sort(byBytewiseFields{bytewiseCanonicalFields})

	lenFirstCanonicalFields := bytewiseCanonicalFields
	if hasKeyAsInt && hasKeyAsStr {
		lenFirstCanonicalFields = make(fields, len(flds))
		copy(lenFirstCanonicalFields, flds)
		sort.Sort(byLengthFirstFields{lenFirstCanonicalFields})
	}

	structType := encodingStructType{
		fields:                  flds,
		bytewiseCanonicalFields: bytewiseCanonicalFields,
		lenFirstCanonicalFields: lenFirstCanonicalFields,
		omitEmpty:               omitEmpty,
		hasAnonymousField:       hasAnonymousField,
	}
	encodingStructTypeCache.Store(t, structType)
	return structType
}

func getEncodingStructToArrayType(t reflect.Type, flds fields) encodingStructType {
	var hasAnonymousField bool
	for i := 0; i < len(flds); i++ {
		// Get field's encodeFunc
		flds[i].ef = getEncodeFunc(flds[i].typ)
		if flds[i].ef == nil {
			structType := encodingStructType{err: &UnsupportedTypeError{t}}
			encodingStructTypeCache.Store(t, structType)
			return structType
		}

		// Check if field is from embedded struct
		if len(flds[i].idx) > 1 {
			hasAnonymousField = true
		}
	}

	structType := encodingStructType{
		fields:            flds,
		toArray:           true,
		hasAnonymousField: hasAnonymousField,
	}
	encodingStructTypeCache.Store(t, structType)
	return structType
}

func getEncodeFunc(t reflect.Type) encodeFunc {
	if v, _ := encodeFuncCache.Load(t); v != nil {
		return v.(encodeFunc)
	}
	f := getEncodeFuncInternal(t)
	encodeFuncCache.Store(t, f)
	return f
}

func hasToArrayOption(tag string) bool {
	s := ",toarray"
	idx := strings.Index(tag, s)
	return idx >= 0 && (len(tag) == idx+len(s) || tag[idx+len(s)] == ',')
}
