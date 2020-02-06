// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

/*
Package cbor provides a fuzz-tested CBOR encoder and decoder with full support
for float16, Canonical CBOR, CTAP2 Canonical CBOR, and custom settings.

CBOR encoding options allow "preferred serialization" by encoding integers and floats
to their smallest forms (like float16) when values fit.

Struct tags like "keyasint", "toarray" and "omitempty" makes CBOR data smaller.

For example, "toarray" makes struct fields encode to array elements.  And "keyasint"
makes struct fields encode to elements of CBOR map with int keys.

Basics

Function signatures identical to encoding/json include:

    Marshal, Unmarshal, NewEncoder, NewDecoder, encoder.Encode, decoder.Decode.  

Codec functions are available at package-level (using defaults) or by creating modes 
from options at runtime.

"Mode" in this API means definite way of encoding or decoding. Specifically, EncMode or DecMode.

EncMode and DecMode interfaces are created from EncOptions or DecOptions structs.  For example,

    em := cbor.EncOptions{...}.EncMode()
    em := cbor.CanonicalEncOptions().EncMode()
    em := cbor.CTAP2EncOptions().EncMode()

Modes use immutable options to avoid side-effects and simplify concurrency. Behavior of modes 
won't accidentally change at runtime after they're created.  

Modes are intended to be reused and are safe for concurrent use.

EncMode and DecMode Interfaces

    // EncMode interface uses immutable options and is safe for concurrent use.
    type EncMode interface {
	Marshal(v interface{}) ([]byte, error)
	NewEncoder(w io.Writer) *Encoder
	EncOptions() EncOptions  // returns copy of options
    }

    // DecMode interface uses immutable options and is safe for concurrent use.
    type DecMode interface {
	Unmarshal(data []byte, v interface{}) error
	NewDecoder(r io.Reader) *Decoder
	DecOptions() DecOptions  // returns copy of options
    }

Using Default Encoding Mode

    b, err := cbor.Marshal(v)
    
    encoder := cbor.NewEncoder(w)
    err = encoder.Encode(v)
    
Using Default Decoding Mode

    err := cbor.Unmarshal(b, &v)
    
    decoder := cbor.NewDecoder(r)
    err = decoder.Decode(&v)

Creating and Using Encoding Modes

    // Create EncOptions using either struct literal or a function.
    opts := cbor.CanonicalEncOptions()

    // If needed, modify encoding options
    opts.Time = cbor.TimeUnix

    // Create reusable EncMode interface with immutable options, safe for concurrent use.
    em, err := opts.EncMode()   

    // Use EncMode like encoding/json, with same function signatures.
    b, err := em.Marshal(v)
    // or
    encoder := em.NewEncoder(w)
    err := encoder.Encode(v)

Default Options

Default encoding options are listed at https://github.com/fxamacker/cbor#api

Struct Tags  

Struct tags like `cbor:"name,omitempty"` and `json:"name,omitempty"` work as expected.
If both struct tags are specified then `cbor` is used.

Struct tags like "keyasint", "toarray", and "omitempty" make it easy to use
very compact formats like COSE and CWT (CBOR Web Tokens) with structs.

For example, "toarray" makes struct fields encode to array elements.  And "keyasint"
makes struct fields encode to elements of CBOR map with int keys.

https://raw.githubusercontent.com/fxamacker/images/master/cbor/v2.0.0/cbor_easy_api.png

Tests and Fuzzing

Over 375 tests are included in this package. Cover-guided fuzzing is handled by a separate package: 
fxamacker/cbor-fuzz.
*/
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
	decodingStructTypeCache sync.Map // map[reflect.Type]*decodingStructType
	encodingStructTypeCache sync.Map // map[reflect.Type]*encodingStructType
	encodeFuncCache         sync.Map // map[reflect.Type]encodeFunc
)

type decodingStructType struct {
	fields  fields
	toArray bool
}

func getDecodingStructType(t reflect.Type) *decodingStructType {
	if v, _ := decodingStructTypeCache.Load(t); v != nil {
		return v.(*decodingStructType)
	}

	flds, structOptions := getFields(t)

	toArray := hasToArrayOption(structOptions)

	for i := 0; i < len(flds); i++ {
		flds[i].isUnmarshaler = implementsUnmarshaler(flds[i].typ)
	}

	structType := &decodingStructType{fields: flds, toArray: toArray}
	decodingStructTypeCache.Store(t, structType)
	return structType
}

type encodingStructType struct {
	fields            fields
	bytewiseFields    fields
	lengthFirstFields fields
	err               error
	toArray           bool
	omitEmpty         bool
	hasAnonymousField bool
}

func (st *encodingStructType) getFields(em *encMode) fields {
	if em.sort == SortNone {
		return st.fields
	}
	if em.sort == SortLengthFirst {
		return st.lengthFirstFields
	}
	return st.bytewiseFields
}

type bytewiseFieldSorter struct {
	fields fields
}

func (x *bytewiseFieldSorter) Len() int {
	return len(x.fields)
}

func (x *bytewiseFieldSorter) Swap(i, j int) {
	x.fields[i], x.fields[j] = x.fields[j], x.fields[i]
}

func (x *bytewiseFieldSorter) Less(i, j int) bool {
	return bytes.Compare(x.fields[i].cborName, x.fields[j].cborName) <= 0
}

type lengthFirstFieldSorter struct {
	fields fields
}

func (x *lengthFirstFieldSorter) Len() int {
	return len(x.fields)
}

func (x *lengthFirstFieldSorter) Swap(i, j int) {
	x.fields[i], x.fields[j] = x.fields[j], x.fields[i]
}

func (x *lengthFirstFieldSorter) Less(i, j int) bool {
	if len(x.fields[i].cborName) != len(x.fields[j].cborName) {
		return len(x.fields[i].cborName) < len(x.fields[j].cborName)
	}
	return bytes.Compare(x.fields[i].cborName, x.fields[j].cborName) <= 0
}

func getEncodingStructType(t reflect.Type) *encodingStructType {
	if v, _ := encodingStructTypeCache.Load(t); v != nil {
		return v.(*encodingStructType)
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
		structType := &encodingStructType{err: err}
		encodingStructTypeCache.Store(t, structType)
		return structType
	}

	// Sort fields by canonical order
	bytewiseFields := make(fields, len(flds))
	copy(bytewiseFields, flds)
	sort.Sort(&bytewiseFieldSorter{bytewiseFields})

	lengthFirstFields := bytewiseFields
	if hasKeyAsInt && hasKeyAsStr {
		lengthFirstFields = make(fields, len(flds))
		copy(lengthFirstFields, flds)
		sort.Sort(&lengthFirstFieldSorter{lengthFirstFields})
	}

	structType := &encodingStructType{
		fields:            flds,
		bytewiseFields:    bytewiseFields,
		lengthFirstFields: lengthFirstFields,
		omitEmpty:         omitEmpty,
		hasAnonymousField: hasAnonymousField,
	}
	encodingStructTypeCache.Store(t, structType)
	return structType
}

func getEncodingStructToArrayType(t reflect.Type, flds fields) *encodingStructType {
	var hasAnonymousField bool
	for i := 0; i < len(flds); i++ {
		// Get field's encodeFunc
		flds[i].ef = getEncodeFunc(flds[i].typ)
		if flds[i].ef == nil {
			structType := &encodingStructType{err: &UnsupportedTypeError{t}}
			encodingStructTypeCache.Store(t, structType)
			return structType
		}

		// Check if field is from embedded struct
		if len(flds[i].idx) > 1 {
			hasAnonymousField = true
		}
	}

	structType := &encodingStructType{
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
