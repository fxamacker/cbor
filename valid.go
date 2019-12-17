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
	if len(data) == 0 {
		return nil, io.EOF
	}
	offset, err := valid(data, 0)
	if err != nil {
		return nil, err
	}
	return data[offset:], nil
}

func valid(data []byte, off int) (int, error) {
	off, t, ai, val, err := validHead(data, off)
	if err != nil {
		return 0, err
	}
	if ai == 31 {
		return validIndefinite(data, off, t)
	}

	switch t {
	case cborTypeByteString, cborTypeTextString:
		valInt := int(val)
		if valInt < 0 {
			// Detect integer overflow
			return 0, errors.New("cbor: " + t.String() + " length " + strconv.FormatUint(val, 10) + " is too large, causing integer overflow")
		}
		if len(data)-off < valInt { // valInt+off may overflow integer
			return 0, io.ErrUnexpectedEOF
		}
		off += valInt
	case cborTypeArray, cborTypeMap:
		valInt := int(val)
		if valInt < 0 {
			// Detect integer overflow
			return 0, errors.New("cbor: " + t.String() + " length " + strconv.FormatUint(val, 10) + " is too large, causing integer overflow")
		}
		count := 1
		if t == cborTypeMap {
			count = 2
		}
		for j := 0; j < count; j++ {
			for i := 0; i < valInt; i++ {
				if off, err = valid(data, off); err != nil {
					return 0, err
				}
			}
		}
	case cborTypeTag:
		// Scan nested tag numbers to avoid recursion.
		for true {
			if len(data)-off < 1 { // Tag number must be followed by tag content.
				return 0, io.ErrUnexpectedEOF
			}
			if cborType(data[off]&0xE0) != cborTypeTag {
				break
			}
			if off, _, _, _, err = validHead(data, off); err != nil {
				return 0, err
			}
		}
		// Check tag content.
		if off, err = valid(data, off); err != nil {
			return 0, err
		}
	}
	return off, nil
}

func validIndefinite(data []byte, off int, t cborType) (_ int, err error) {
	isString := (t == cborTypeByteString) || (t == cborTypeTextString)
	for true {
		if len(data)-off < 1 {
			return 0, io.ErrUnexpectedEOF
		}
		if data[off] == 0xFF {
			off++
			break
		}
		if isString {
			// Peek ahead to get next type and indefinite length status.
			nextType := cborType(data[off] & 0xE0)
			if t != nextType {
				return 0, &SyntaxError{"cbor: wrong element type " + nextType.String() + " for indefinite-length " + t.String()}
			}
			if (data[off] & 0x1F) == 31 {
				return 0, &SyntaxError{"cbor: indefinite-length " + t.String() + " chunk is not definite-length"}
			}
		}
		if off, err = valid(data, off); err != nil {
			return 0, err
		}
		if t == cborTypeMap {
			if off, err = valid(data, off); err != nil {
				return 0, err
			}
		}
	}
	return off, nil
}

func validHead(data []byte, off int) (_ int, t cborType, ai byte, val uint64, err error) {
	dataLen := len(data) - off
	if dataLen < 1 {
		return 0, 0, 0, 0, io.ErrUnexpectedEOF
	}

	t = cborType(data[off] & 0xE0)
	ai = data[off] & 0x1F
	val = uint64(ai)
	off++

	switch ai {
	case 24:
		if dataLen < 2 {
			return 0, 0, 0, 0, io.ErrUnexpectedEOF
		}
		val = uint64(data[off])
		off++
	case 25:
		if dataLen < 3 {
			return 0, 0, 0, 0, io.ErrUnexpectedEOF
		}
		val = uint64(binary.BigEndian.Uint16(data[off : off+2]))
		off += 2
	case 26:
		if dataLen < 5 {
			return 0, 0, 0, 0, io.ErrUnexpectedEOF
		}
		val = uint64(binary.BigEndian.Uint32(data[off : off+4]))
		off += 4
	case 27:
		if dataLen < 9 {
			return 0, 0, 0, 0, io.ErrUnexpectedEOF
		}
		val = binary.BigEndian.Uint64(data[off : off+8])
		off += 8
	case 28, 29, 30:
		return 0, 0, 0, 0, &SyntaxError{"cbor: invalid additional information " + strconv.Itoa(int(ai)) + " for type " + t.String()}
	case 31:
		switch t {
		case cborTypePositiveInt, cborTypeNegativeInt, cborTypeTag:
			return 0, 0, 0, 0, &SyntaxError{"cbor: invalid additional information " + strconv.Itoa(int(ai)) + " for type " + t.String()}
		case cborTypePrimitives: // 0xFF (break code) should not be outside validIndefinite().
			return 0, 0, 0, 0, &SyntaxError{"cbor: unexpected \"break\" code"}
		}
	}
	if t == cborTypePrimitives && ai == 24 && val < 32 {
		return 0, 0, 0, 0, &SyntaxError{"cbor: invalid simple value " + strconv.Itoa(int(val)) + " for type " + t.String()}
	}
	return off, t, ai, val, nil
}
