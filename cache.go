package cbor

import (
	"bytes"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type decodingStructType struct {
	fields  fields
	err     error
	toArray bool
}

type encodingStructType struct {
	fields          fields
	canonicalFields fields
	err             error
	toArray         bool
}

// byCanonicalRule sorts fields by field name length and field name.
type byCanonicalRule struct {
	fields
}

func (s byCanonicalRule) Less(i, j int) bool {
	return bytes.Compare(s.fields[i].cborName, s.fields[j].cborName) <= 0
}

var (
	decodingStructTypeCache sync.Map // map[reflect.Type]decodingStructType
	encodingStructTypeCache sync.Map // map[reflect.Type]encodingStructType
	encodingTypeCache       sync.Map // map[reflect.Type]encodeFunc
)

func structToArray(tag string) bool {
	s := ",toarray"
	idx := strings.Index(tag, s)
	return idx >= 0 && (len(tag) == idx+len(s) || tag[idx+len(s)] == ',')
}

func getDecodingStructType(t reflect.Type) decodingStructType {
	if v, _ := decodingStructTypeCache.Load(t); v != nil {
		return v.(decodingStructType)
	}

	flds, structOptions := getFields(t)

	toArray := structToArray(structOptions)

	for i := 0; i < len(flds); i++ {
		flds[i].isUnmarshaler = implementsUnmarshaler(flds[i].typ)
	}

	structType := decodingStructType{fields: flds, toArray: toArray}
	decodingStructTypeCache.Store(t, structType)
	return structType
}

func getEncodingStructType(t reflect.Type) encodingStructType {
	if v, _ := encodingStructTypeCache.Load(t); v != nil {
		return v.(encodingStructType)
	}

	flds, structOptions := getFields(t)

	toArray := structToArray(structOptions)

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

	structType := encodingStructType{fields: flds, canonicalFields: canonicalFields, toArray: toArray}
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
