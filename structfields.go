// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

package cbor

import (
	"errors"
	"reflect"
	"sort"
	"strings"
)

// fieldByIndex returns the nested field corresponding to the index.  It
// allocates pointer to struct field if it is nil and settable.
// reflect.Value.FieldByIndex() panics at nil pointer to unexported anonymous
// field.  This function returns error.
func fieldByIndex(v reflect.Value, index []int) (reflect.Value, error) {
	for _, i := range index {
		if v.Kind() == reflect.Ptr && v.Type().Elem().Kind() == reflect.Struct {
			if v.IsNil() {
				if !v.CanSet() {
					return reflect.Value{}, errors.New("cbor: cannot set embedded pointer to unexported struct: " + v.Type().String())
				}
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v, nil
}

type field struct {
	name          string
	cborName      []byte
	idx           []int
	typ           reflect.Type
	ef            encodeFunc
	isUnmarshaler bool
	tagged        bool // used to choose dominant field (at the same level tagged fields dominate untagged fields)
	omitempty     bool // used to skip empty field
	keyasint      bool // used to encode/decode field name as int
}

type fields []field

func (s fields) Len() int {
	return len(s)
}

func (s fields) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// byIndex sorts fields by field idx at each level, breaking ties with idx depth.
type byIndex struct {
	fields
}

func (s byIndex) Less(i, j int) bool {
	iidx := s.fields[i].idx
	jidx := s.fields[j].idx
	for k, d := range iidx {
		if k >= len(jidx) {
			// fields[j].idx is a subset of fields[i].idx.
			return false
		}
		if d != jidx[k] {
			// fields[i].idx and fields[j].idx are different.
			return d < jidx[k]
		}
	}
	// fields[i].idx is either the same as, or a subset of fields[j].idx.
	return true
}

// byNameLevelAndTag sorts fields by field name, idx depth, and presence of tag.
type byNameLevelAndTag struct {
	fields
}

func (s byNameLevelAndTag) Less(i, j int) bool {
	if s.fields[i].name != s.fields[j].name {
		return s.fields[i].name < s.fields[j].name
	}
	if len(s.fields[i].idx) != len(s.fields[j].idx) {
		return len(s.fields[i].idx) < len(s.fields[j].idx)
	}
	if s.fields[i].tagged != s.fields[j].tagged {
		return s.fields[i].tagged
	}
	return i < j // Field i and j have the same name, depth, and tagged status. Nothing else matters.
}

func getFieldNameAndOptionsFromTag(tag string) (name string, omitEmpty bool, keyAsInt bool) {
	if len(tag) == 0 {
		return
	}
	tokens := strings.Split(tag, ",")
	name = tokens[0]
	for _, s := range tokens[1:] {
		if s == "omitempty" {
			omitEmpty = true
		} else if s == "keyasint" {
			keyAsInt = true
		}
	}
	return
}

// getFields returns a list of visible fields of struct type typ following Go
// visibility rules for struct fields.
func getFields(typ reflect.Type) fields {
	// Inspired by Go JSON encoding package's typeFields() function in encoding/json/encode.go.

	var current map[reflect.Type][][]int // key: struct type, value: field index of this struct type at the same level
	next := map[reflect.Type][][]int{typ: nil}

	visited := map[reflect.Type]bool{} // Inspected struct type at less nested levels.

	var flds fields

	for len(next) > 0 {
		current, next = next, map[reflect.Type][][]int{}

		for structType, structIdxSlice := range current {
			if len(structIdxSlice) > 1 {
				continue // Fields of the same type at the same level are ignored.
			}

			if visited[structType] {
				continue
			}
			visited[structType] = true

			var fieldIdx []int
			if len(structIdxSlice) > 0 {
				fieldIdx = structIdxSlice[0]
			}

			for i := 0; i < structType.NumField(); i++ {
				f := structType.Field(i)
				ft := f.Type

				if ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}

				exportable := f.PkgPath == ""
				if f.Anonymous {
					if !exportable && ft.Kind() != reflect.Struct {
						// Nonexportable anonymous fields of non-struct type are ignored.
						continue
					}
					// Nonexportable anonymous field of struct type can contain exportable fields for serialization.
				} else if !exportable {
					// Nonexportable fields are ignored.
					continue
				}

				tag := f.Tag.Get("cbor")
				if tag == "-" {
					continue
				}
				if tag == "" {
					tag = f.Tag.Get("json")
					if tag == "-" {
						continue
					}
				}

				idx := make([]int, len(fieldIdx)+1)
				copy(idx, fieldIdx)
				idx[len(fieldIdx)] = i

				tagged := len(tag) > 0
				tagFieldName, omitempty, keyasint := getFieldNameAndOptionsFromTag(tag)

				fieldName := tagFieldName
				if tagFieldName == "" {
					fieldName = f.Name
				}

				if !f.Anonymous || ft.Kind() != reflect.Struct || len(tagFieldName) > 0 {
					flds = append(flds, field{name: fieldName, idx: idx, typ: f.Type, tagged: tagged, omitempty: omitempty, keyasint: keyasint})
					continue
				}

				// f is anonymous struct of type ft.
				next[ft] = append(next[ft], idx)
			}
		}
	}

	sort.Sort(byNameLevelAndTag{flds})

	// Keep visible fields.
	visibleFields := flds[:0]
	for i, j := 0, 0; i < len(flds); i = j {
		name := flds[i].name
		for j = i + 1; j < len(flds) && flds[j].name == name; j++ {
		}
		if j-i == 1 || len(flds[i].idx) < len(flds[i+1].idx) || (flds[i].tagged && !flds[i+1].tagged) {
			// Keep the field if the field name is unique, or if the first field
			// is at a less nested level, or if the first field is tagged and
			// the second field is not.
			visibleFields = append(visibleFields, flds[i])
		}
	}

	sort.Sort(byIndex{visibleFields})

	return visibleFields
}
