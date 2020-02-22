// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"encoding/binary"
	"errors"
	"io"
	"strconv"
)

// SyntaxError is a description of a CBOR syntax error.
type SyntaxError struct {
	msg string
}

func (e *SyntaxError) Error() string { return e.msg }

// SemanticError is a description of a CBOR semantic error.
type SemanticError struct {
	msg string
}

func (e *SemanticError) Error() string { return e.msg }

// MaxNestedLevelError indicates exceeded max nested level of any combination of CBOR arrays/maps/tags.
type MaxNestedLevelError struct {
	maxNestedLevel int
}

func (e *MaxNestedLevelError) Error() string {
	return "cbor: reached max nested level " + strconv.Itoa(e.maxNestedLevel)
}

// valid checks whether CBOR data is complete and well-formed.
func (d *decodeState) valid() error {
	if len(d.data) == d.off {
		return io.EOF
	}
	_, err := d.validInternal(0)
	return err
}

// validInternal checks data's well-formedness and returns max depth and error.
func (d *decodeState) validInternal(depth int) (int, error) {
	t, ai, val, err := d.validHead()
	if err != nil {
		return 0, err
	}

	switch t {
	case cborTypeByteString, cborTypeTextString:
		if ai == 31 {
			return d.validIndefiniteString(t, depth)
		}
		valInt := int(val)
		if valInt < 0 {
			// Detect integer overflow
			return 0, errors.New("cbor: " + t.String() + " length " + strconv.FormatUint(val, 10) + " is too large, causing integer overflow")
		}
		if len(d.data)-d.off < valInt { // valInt+off may overflow integer
			return 0, io.ErrUnexpectedEOF
		}
		d.off += valInt
	case cborTypeArray, cborTypeMap:
		depth++
		if depth > d.dm.maxNestedLevel {
			return 0, &MaxNestedLevelError{d.dm.maxNestedLevel}
		}

		if ai == 31 {
			return d.validIndefiniteArrOrMap(t, depth)
		}

		valInt := int(val)
		if valInt < 0 {
			// Detect integer overflow
			return 0, errors.New("cbor: " + t.String() + " length " + strconv.FormatUint(val, 10) + " is too large, causing integer overflow")
		}
		count := 1
		if t == cborTypeMap {
			count = 2
		}
		maxDepth := depth
		for j := 0; j < count; j++ {
			for i := 0; i < valInt; i++ {
				var dpt int
				if dpt, err = d.validInternal(depth); err != nil {
					return 0, err
				}
				if dpt > maxDepth {
					maxDepth = dpt // Save max depth
				}
			}
		}
		depth = maxDepth
	case cborTypeTag:
		// Scan nested tag numbers to avoid recursion.
		for {
			if len(d.data) == d.off { // Tag number must be followed by tag content.
				return 0, io.ErrUnexpectedEOF
			}
			if cborType(d.data[d.off]&0xe0) != cborTypeTag {
				break
			}
			if _, _, _, err = d.validHead(); err != nil {
				return 0, err
			}
			depth++
			if depth > d.dm.maxNestedLevel {
				return 0, &MaxNestedLevelError{d.dm.maxNestedLevel}
			}
		}
		// Check tag content.
		return d.validInternal(depth)
	}
	return depth, nil
}

// validIndefiniteString checks indefinite length byte/text string's well-formedness and returns max depth and error.
func (d *decodeState) validIndefiniteString(t cborType, depth int) (int, error) {
	var err error
	for {
		if len(d.data) == d.off {
			return 0, io.ErrUnexpectedEOF
		}
		if d.data[d.off] == 0xff {
			d.off++
			break
		}
		// Peek ahead to get next type and indefinite length status.
		nt := cborType(d.data[d.off] & 0xe0)
		if t != nt {
			return 0, &SyntaxError{"cbor: wrong element type " + nt.String() + " for indefinite-length " + t.String()}
		}
		if (d.data[d.off] & 0x1f) == 31 {
			return 0, &SyntaxError{"cbor: indefinite-length " + t.String() + " chunk is not definite-length"}
		}
		if depth, err = d.validInternal(depth); err != nil {
			return 0, err
		}
	}
	return depth, nil
}

// validIndefiniteArrOrMap checks indefinite length array/map's well-formedness and returns max depth and error.
func (d *decodeState) validIndefiniteArrOrMap(t cborType, depth int) (int, error) {
	var err error
	maxDepth := depth
	i := 0
	for {
		if len(d.data) == d.off {
			return 0, io.ErrUnexpectedEOF
		}
		if d.data[d.off] == 0xff {
			d.off++
			break
		}
		var dpt int
		if dpt, err = d.validInternal(depth); err != nil {
			return 0, err
		}
		if dpt > maxDepth {
			maxDepth = dpt
		}
		i++
	}
	if t == cborTypeMap && i%2 == 1 {
		return 0, &SyntaxError{"cbor: unexpected \"break\" code"}
	}
	return maxDepth, nil
}

func (d *decodeState) validHead() (t cborType, ai byte, val uint64, err error) {
	dataLen := len(d.data) - d.off
	if dataLen == 0 {
		return 0, 0, 0, io.ErrUnexpectedEOF
	}

	t = cborType(d.data[d.off] & 0xe0)
	ai = d.data[d.off] & 0x1f
	val = uint64(ai)
	d.off++

	if ai < 24 {
		return t, ai, val, nil
	}
	if ai == 24 {
		if dataLen < 2 {
			return 0, 0, 0, io.ErrUnexpectedEOF
		}
		val = uint64(d.data[d.off])
		d.off++
		if t == cborTypePrimitives && val < 32 {
			return 0, 0, 0, &SyntaxError{"cbor: invalid simple value " + strconv.Itoa(int(val)) + " for type " + t.String()}
		}
		return t, ai, val, nil
	}
	if ai == 25 {
		if dataLen < 3 {
			return 0, 0, 0, io.ErrUnexpectedEOF
		}
		val = uint64(binary.BigEndian.Uint16(d.data[d.off : d.off+2]))
		d.off += 2
		return t, ai, val, nil
	}
	if ai == 26 {
		if dataLen < 5 {
			return 0, 0, 0, io.ErrUnexpectedEOF
		}
		val = uint64(binary.BigEndian.Uint32(d.data[d.off : d.off+4]))
		d.off += 4
		return t, ai, val, nil
	}
	if ai == 27 {
		if dataLen < 9 {
			return 0, 0, 0, io.ErrUnexpectedEOF
		}
		val = binary.BigEndian.Uint64(d.data[d.off : d.off+8])
		d.off += 8
		return t, ai, val, nil
	}
	if ai == 31 {
		switch t {
		case cborTypePositiveInt, cborTypeNegativeInt, cborTypeTag:
			return 0, 0, 0, &SyntaxError{"cbor: invalid additional information " + strconv.Itoa(int(ai)) + " for type " + t.String()}
		case cborTypePrimitives: // 0xff (break code) should not be outside validIndefinite().
			return 0, 0, 0, &SyntaxError{"cbor: unexpected \"break\" code"}
		}
		return t, ai, val, nil
	}
	// ai == 28, 29, 30
	return 0, 0, 0, &SyntaxError{"cbor: invalid additional information " + strconv.Itoa(int(ai)) + " for type " + t.String()}
}
