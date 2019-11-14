package cbor

import (
	"bytes"
	"reflect"
	"sort"
	"strconv"
	"sync"
)

type encodingStructType struct {
	fields          fields
	canonicalFields fields
	err             error
}

// byCanonicalRule sorts fields by field name length and field name.
type byCanonicalRule struct {
	fields
}

func (s byCanonicalRule) Less(i, j int) bool {
	return bytes.Compare(s.fields[i].cborName, s.fields[j].cborName) <= 0
}

var (
	decodingStructTypeCache sync.Map // map[reflect.Type]fields
	encodingStructTypeCache sync.Map // map[reflect.Type]encodingStructType
	encodingTypeCache       sync.Map // map[reflect.Type]encodeFunc
)

func getDecodingStructType(t reflect.Type) fields {
	if v, _ := decodingStructTypeCache.Load(t); v != nil {
		return v.(fields)
	}
	flds := getFields(t)
	for i := 0; i < len(flds); i++ {
		flds[i].isUnmarshaler = implementsUnmarshaler(flds[i].typ)
	}
	decodingStructTypeCache.Store(t, flds)
	return flds
}

func getEncodingStructType(t reflect.Type) encodingStructType {
	if v, _ := encodingStructTypeCache.Load(t); v != nil {
		return v.(encodingStructType)
	}

	flds := getFields(t)

	var err error
	es := getEncodeState()
	for i := 0; i < len(flds); i++ {
		// Get field's encodeFunc
		flds[i].ef = getEncodeFunc(flds[i].typ)
		if flds[i].ef == nil {
			err = &UnsupportedTypeError{t}
			break
		}

		// Encode field name
		if flds[i].keyasint {
			nameAsInt, numErr := strconv.Atoi(flds[i].name)
			if numErr != nil {
				err = numErr
				break
			}
			if nameAsInt >= 0 {
				encodeTypeAndAdditionalValue(es, byte(cborTypePositiveInt), uint64(nameAsInt))
			} else {
				n := nameAsInt*(-1) - 1
				encodeTypeAndAdditionalValue(es, byte(cborTypeNegativeInt), uint64(n))
			}
			flds[i].cborName = make([]byte, es.Len())
			copy(flds[i].cborName, es.Bytes())
			es.Reset()
		} else {
			encodeTypeAndAdditionalValue(es, byte(cborTypeTextString), uint64(len(flds[i].name)))
			flds[i].cborName = make([]byte, es.Len()+len(flds[i].name))
			n := copy(flds[i].cborName, es.Bytes())
			copy(flds[i].cborName[n:], flds[i].name)
			es.Reset()
		}
	}
	putEncodeState(es)

	if err != nil {
		structType := encodingStructType{err: err}
		encodingStructTypeCache.Store(t, structType)
		return structType
	}

	// Sort fields by canonical order
	canonicalFields := make(fields, len(flds))
	copy(canonicalFields, flds)
	sort.Sort(byCanonicalRule{canonicalFields})

	structType := encodingStructType{fields: flds, canonicalFields: canonicalFields, err: err}
	encodingStructTypeCache.Store(t, structType)
	return structType
}

func getEncodeFunc(t reflect.Type) encodeFunc {
	if v, _ := encodingTypeCache.Load(t); v != nil {
		return v.(encodeFunc)
	}
	f := getEncodeFuncInternal(t)
	encodingTypeCache.Store(t, f)
	return f
}
