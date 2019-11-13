package cbor

import (
	"bytes"
	"reflect"
	"sort"
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
		ef := getEncodeFunc(flds[i].typ)
		if ef == nil {
			if err == nil {
				err = &UnsupportedTypeError{t}
			}
		}
		flds[i].ef = ef

		encodeTypeAndAdditionalValue(es, byte(cborTypeTextString), uint64(len(flds[i].name)))
		flds[i].cborName = make([]byte, es.Len()+len(flds[i].name))
		copy(flds[i].cborName, es.Bytes())
		copy(flds[i].cborName[es.Len():], flds[i].name)

		es.Reset()
	}
	putEncodeState(es)

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
