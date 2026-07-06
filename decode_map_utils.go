// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import "strings"

// mapAction represents the next action during decoding a CBOR map to a Go struct.
type mapAction int

const (
	mapActionParseValueAndContinue mapAction = iota // The caller should process the map value.
	mapActionSkipValueAndContinue                   // The caller should skip the map value.
	mapActionSkipAllAndReturnError                  // The caller should skip the rest of the map and return an error.
)

// checkDupField checks if a struct field at index i has already been matched and returns the next action.
// If not matched, it marks the field as matched and returns mapActionParseValueAndContinue.
// If matched and DupMapKeyEnforcedAPF is specified in the given dm, it returns mapActionSkipAllAndReturnError.
// If matched and DupMapKeyEnforcedAPF is not specified in the given dm, it returns mapActionSkipValueAndContinue.
func checkDupField(dm *decMode, foundFldIdx []bool, i int) mapAction {
	if !foundFldIdx[i] {
		foundFldIdx[i] = true
		return mapActionParseValueAndContinue
	}
	if dm.dupMapKey == DupMapKeyEnforcedAPF {
		return mapActionSkipAllAndReturnError
	}
	return mapActionSkipValueAndContinue
}

// findStructFieldByKey finds a struct field matching keyBytes by name.
// It tries an exact match first. If no exact match is found and
// caseInsensitive is true, it falls back to a case-insensitive search.
// findStructFieldByKey returns the field index and true, or -1 and false.
func findStructFieldByKey(
	structType *decodingStructType,
	keyBytes []byte,
	caseInsensitive bool,
) (int, bool) {
	if fldIdx, ok := structType.fieldIndicesByName[string(keyBytes)]; ok {
		return fldIdx, true
	}
	if caseInsensitive {
		return findFieldCaseInsensitiveBytes(structType.fields, keyBytes)
	}
	return -1, false
}

// findFieldCaseInsensitiveBytes returns the index of the first field whose name
// case-insensitively matches keyBytes, or -1 and false if no field matches.
//
// It keeps the common ASCII path allocation-free and avoids strings.EqualFold's
// Unicode machinery unless either side contains non-ASCII bytes.
func findFieldCaseInsensitiveBytes(flds decodingFields, keyBytes []byte) (int, bool) {
	keyLen := len(keyBytes)
	var key string
	var keySet bool

	for i, f := range flds {
		if f.keyAsInt || len(f.name) != keyLen {
			continue
		}

		if match, ascii := equalFoldASCIIStringBytes(f.name, keyBytes); ascii {
			if match {
				return i, true
			}
			continue
		}

		if !keySet {
			key = string(keyBytes)
			keySet = true
		}
		if strings.EqualFold(f.name, key) {
			return i, true
		}
	}
	return -1, false
}

// equalFoldASCIIStringBytes compares s and b using ASCII case folding.
// ascii is false if either input contains non-ASCII bytes, in which case the
// caller must use strings.EqualFold to preserve Unicode case-folding semantics.
func equalFoldASCIIStringBytes(s string, b []byte) (match bool, ascii bool) {
	if len(s) != len(b) {
		return false, true
	}

	for i := range b {
		c := s[i]
		d := b[i]
		if c|d >= 0x80 {
			return false, false
		}

		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		if 'A' <= d && d <= 'Z' {
			d += 'a' - 'A'
		}
		if c != d {
			return false, true
		}
	}
	return true, true
}

// handleUnmatchedMapKey handles a map entry whose key does not match any struct
// field. It can return UnknownFieldError or DupMapKeyError.
// handleUnmatchedMapKey consumes the CBOR value, so the caller doesn't need to skip any values.
// If an error is returned, the caller should abort parsing the map and return the error.
// If no error is returned, the caller should continue to process the next map pair.
func handleUnmatchedMapKey(
	d *decoder,
	key any,
	i int,
	count int,
	hasSize bool,
	// *map[any]struct{} is used here because we use lazy initialization for uks
	uks *map[any]struct{}, //nolint:gocritic
) error {
	errOnUnknownField := (d.dm.extraReturnErrors & ExtraDecErrorUnknownField) > 0

	if errOnUnknownField {
		return d.skipMapForUnknownField(i, count, hasSize)
	}

	if d.dm.dupMapKey == DupMapKeyEnforcedAPF {
		if *uks == nil {
			*uks = make(map[any]struct{})
		}
		if _, dup := (*uks)[key]; dup {
			return d.skipMapForDupKey(key, i, count, hasSize)
		}
		(*uks)[key] = struct{}{}
	}

	// Skip value.
	d.skip()
	return nil
}
