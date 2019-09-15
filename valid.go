// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

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

// Valid checks whether CBOR data is complete and well-formed.
func Valid(data []byte) (rest []byte, err error) {
	offset, _, err := checkValid(data, 0)
	if err != nil {
		return nil, err
	}
	return data[offset:], nil
}

func checkValid(data []byte, off int) (_ int, t cborType, err error) {
	if len(data)-off < 1 {
		return 0, 0, io.EOF
	}
	t = cborType(data[off] & 0xE0)
	ai := data[off] & 0x1F
	val := uint64(ai)
	off++

	// Check additional information.
	switch ai {
	case 24:
		if len(data)-off < 1 {
			return 0, 0, io.ErrUnexpectedEOF
		}
		val = uint64(data[off])
		off++
	case 25:
		if len(data)-off < 2 {
			return 0, 0, io.ErrUnexpectedEOF
		}
		val = uint64(binary.BigEndian.Uint16(data[off : off+2]))
		off += 2
	case 26:
		if len(data)-off < 4 {
			return 0, 0, io.ErrUnexpectedEOF
		}
		val = uint64(binary.BigEndian.Uint32(data[off : off+4]))
		off += 4
	case 27:
		if len(data)-off < 8 {
			return 0, 0, io.ErrUnexpectedEOF
		}
		val = binary.BigEndian.Uint64(data[off : off+8])
		off += 8
	case 28, 29, 30:
		return 0, 0, &SyntaxError{"cbor: invalid additional information " + strconv.Itoa(int(ai)) + " for type " + t.String()}
	case 31:
		switch t {
		case cborTypePositiveInt, cborTypeNegativeInt, cborTypeTag:
			return 0, 0, &SyntaxError{"cbor: invalid additional information " + strconv.Itoa(int(ai)) + " for type " + t.String()}
		}
	}

	// Check indefinite byte/text string, array, and map.
	if ai == 31 {
		switch t {
		case cborTypeByteString, cborTypeTextString, cborTypeArray, cborTypeMap:
			if off, err = checkValidIndefinite(data, off, t); err != nil {
				return 0, 0, err
			}
			return off, t, nil
		}
	}

	switch t {
	case cborTypeByteString, cborTypeTextString: // Check byte/text string payload.
		valInt := int(val)
		if valInt < 0 || uint64(valInt) != val {
			// Detect integer overflow
			return 0, 0, errors.New("cbor: " + t.String() + " length " + strconv.FormatUint(val, 10) + " is too large, causing integer overflow")
		}
		if len(data)-off < valInt {
			return 0, 0, io.ErrUnexpectedEOF
		}
		off += int(val)

	case cborTypeArray, cborTypeMap: // Check array and map payload.
		valInt := int(val)
		if valInt < 0 || uint64(valInt) != val {
			// Detect integer overflow
			return 0, 0, errors.New("cbor: " + t.String() + " length " + strconv.FormatUint(val, 10) + " is too large, causing integer overflow")
		}
		elementCount := 1
		if t == cborTypeMap {
			elementCount = 2
		}
		for i := 0; i < elementCount; i++ {
			for i := 0; i < valInt; i++ {
				if off, _, err = checkValid(data, off); err != nil {
					if err == io.EOF {
						err = io.ErrUnexpectedEOF
					}
					return 0, 0, err
				}
			}
		}
	case cborTypeTag: // Check tagged item following tag.
		if off, t, err = checkValid(data, off); err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return 0, 0, err
		}
	case cborTypePrimitives: // 0xFF (break code) should not be outside checkValidIndefinite().
		if ai == 31 {
			return 0, 0, &SyntaxError{"cbor: unexpected \"break\" code"}
		}
	}
	return off, t, nil
}

func checkValidIndefinite(data []byte, off int, t cborType) (_ int, err error) {
	for true {
		if len(data)-off < 1 {
			return 0, io.ErrUnexpectedEOF
		}
		if data[off] == 0xFF {
			off++
			break
		}
		var nextType cborType
		if off, nextType, err = checkValid(data, off); err != nil {
			return 0, err
		}
		switch t {
		case cborTypeByteString, cborTypeTextString:
			if t != nextType {
				return 0, &SemanticError{"cbor: wrong element type " + nextType.String() + " for indefinite-length " + t.String()}
			}
		case cborTypeMap:
			if off, _, err = checkValid(data, off); err != nil {
				if err == io.EOF {
					err = io.ErrUnexpectedEOF
				}
				return 0, err
			}
		}
	}
	return off, nil
}
