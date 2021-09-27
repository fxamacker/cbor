package cbor

import (
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"unicode/utf8"
)

// Type represents CBOR type.
type Type uint8

const (
	UndefinedType Type = iota

	// CBOR major types
	UintType       // CBOR major type 0
	IntType        // CBOR major type 1
	ByteStringType // CBOR major type 2
	TextStringType // CBOR major type 3
	ArrayType      // CBOR major type 4
	MapType        // CBOR major type 5
	TagType        // CBOR major type 6
	// OtherType is CBOR major type 7.  It is for two types of data:
	// floating-point numbers and "simple values" that do not need any content.
	OtherType

	// Non-major types
	BigNumType // BigNumType is specified as part of CBOR major type 6
	NilType    // NilType is specified as part of CBOR major type 7
	BoolType   // BoolType is specified as part of CBOR major type 7
)

func (t Type) String() string {
	switch t {
	case UintType:
		return "CBOR uint type"
	case IntType:
		return "CBOR int type"
	case ByteStringType:
		return "CBOR byte string type"
	case TextStringType:
		return "CBOR text string type"
	case ArrayType:
		return "CBOR array type"
	case MapType:
		return "CBOR map type"
	case TagType:
		return "CBOR tag type"
	case OtherType:
		return "CBOR other type"
	case NilType:
		return "CBOR nil type"
	case BoolType:
		return "CBOR boolean type"
	case BigNumType:
		return "CBOR bignum type"
	default:
		return "undefined CBOR type"
	}
}

type WrongTypeError struct {
	ActualType   Type
	ExpectedType string
}

func (e *WrongTypeError) Error() string {
	return fmt.Sprintf("cannot decode %s to %s", e.ActualType.String(), e.ExpectedType)
}

// StreamDecoder validates complete CBOR data and decodes it in chunks.
//
// If CBOR data is malformed or fails decoding checks (set by options),
// validation error is saved and all subseqent decoding functions
// return the saved error.
//
// If DecodeXXX() tries to decode CBOR data of mismatched type,
// WrongTypeError is returned.  Caller can retry to decode the same
// data with a different DecodeXXX().
//
// If DecodeXXX() function returns other types of error,
// CBOR data is skipped.  User can decode next CBOR data.
type StreamDecoder struct {
	dec             *Decoder
	err             error
	remainingBytes  int // remaining bytes of a complete and validated CBOR data
	decodedMsgBytes int // number of bytes of decoded CBOR messages
}

// NewStreamDecoder returns a new StreamDecoder that reads from r using default DecMode.
func NewStreamDecoder(r io.Reader) *StreamDecoder {
	return defaultDecMode.NewStreamDecoder(r)
}

// NewByteStreamDecoder returns a new StreamDecoder that reads from data using default DecMode.
func NewByteStreamDecoder(data []byte) *StreamDecoder {
	return defaultDecMode.NewByteStreamDecoder(data)
}

// NextType returns the next CBOR type.
func (sd *StreamDecoder) NextType() (Type, error) {
	if err := sd.prepareNext(); err != nil {
		return UndefinedType, err
	}

	b := sd.dec.d.data[sd.dec.d.off]
	switch b & 0xe0 {
	case 0x00:
		return UintType, nil
	case 0x20:
		return IntType, nil
	case 0x40:
		return ByteStringType, nil
	case 0x60:
		return TextStringType, nil
	case 0x80:
		return ArrayType, nil
	case 0xa0:
		return MapType, nil
	case 0xc0:
		if b == 0xc2 || b == 0xc3 {
			return BigNumType, nil
		}
		return TagType, nil
	case 0xe0:
		switch b {
		case 0xf4, 0xf5:
			return BoolType, nil
		case 0xf6:
			return NilType, nil
		}
		return OtherType, nil
	}
	return UndefinedType, errors.New("cbor: unrecognized type")
}

// Skip skips next CBOR data.
func (sd *StreamDecoder) Skip() error {
	if err := sd.prepareNext(); err != nil {
		return err
	}

	d := sd.dec.d

	start := d.off
	d.skip()
	end := d.off

	sd.updateState(end - start)

	return nil
}

// DecodeRawBytes returns a copy of next CBOR data as raw bytes.
func (sd *StreamDecoder) DecodeRawBytes() ([]byte, error) {
	if err := sd.prepareNext(); err != nil {
		return nil, err
	}

	d := sd.dec.d

	start := d.off
	d.skip()
	end := d.off

	b := make([]byte, end-start)
	copy(b, d.data[start:end])

	sd.updateState(end - start)

	return b, nil
}

// DecodeRawBytesZeroCopy returns next CBOR data as raw bytes pointing to underlying data.
// It is only available for StreamDecoder created with NewByteStreamDecoder.
func (sd *StreamDecoder) DecodeRawBytesZeroCopy() ([]byte, error) {
	if sd.dec.r != nil {
		return nil, errors.New("cbor: DecodeRawBytesZeroCopy is only supported for StreamDecoder created with NewByteStreamDecoder")
	}

	if err := sd.prepareNext(); err != nil {
		return nil, err
	}

	d := sd.dec.d

	start := d.off
	d.skip()
	end := d.off

	b := d.data[start:end]

	sd.updateState(end - start)

	return b, nil
}

// DecodeNil decodes next CBOR data as nil.  WrongType error is returned if type is mismatched.
func (sd *StreamDecoder) DecodeNil() error {
	if err := sd.prepareNext(); err != nil {
		return err
	}

	d := sd.dec.d

	if d.data[d.off] == 0xf6 {
		d.off++
		sd.updateState(1)
		return nil
	}

	t, _ := sd.NextType()
	return &WrongTypeError{t, "nil"}
}

// DecodeBool decodes next CBOR data as bool.  WrongType error is returned if type is mismatched.
func (sd *StreamDecoder) DecodeBool() (bool, error) {
	if err := sd.prepareNext(); err != nil {
		return false, err
	}

	d := sd.dec.d

	b := d.data[d.off]
	switch b {

	case 0xf4, 0xf5:
		d.off++
		sd.updateState(1)
		return b == 0xf5, nil

	default:
		t, _ := sd.NextType()
		return false, &WrongTypeError{t, "bool"}
	}
}

// DecodeUint64 decodes next CBOR data as uint64.  WrongType error is returned if type is mismatched.
func (sd *StreamDecoder) DecodeUint64() (uint64, error) {
	if err := sd.prepareNext(); err != nil {
		return 0, err
	}

	d := sd.dec.d

	if d.nextCBORType() != cborTypePositiveInt {
		t, _ := sd.NextType()
		return 0, &WrongTypeError{t, "uint64"}
	}

	start := d.off
	_, _, val := d.getHead()
	end := d.off

	sd.updateState(end - start)

	return val, nil
}

// DecodeInt64 decodes next CBOR data as int64.  WrongType error is returned if type is mismatched.
func (sd *StreamDecoder) DecodeInt64() (int64, error) {
	if err := sd.prepareNext(); err != nil {
		return 0, err
	}

	d := sd.dec.d

	switch nt := d.nextCBORType(); nt {

	case cborTypePositiveInt, cborTypeNegativeInt:
		start := d.off
		_, _, val := d.getHead()
		end := d.off

		sd.updateState(end - start)

		if val > math.MaxInt64 {
			return 0, fmt.Errorf("cbor: %d overflow Go's int64", val)
		}

		if nt == cborTypePositiveInt {
			return int64(val), nil
		}

		nValue := int64(-1) ^ int64(val)
		return nValue, nil

	default:
		t, _ := sd.NextType()
		return 0, &WrongTypeError{t, "int64"}
	}
}

// DecodeBytes decodes next CBOR data as []byte.  WrongType error is returned if type is mismatched.
func (sd *StreamDecoder) DecodeBytes() ([]byte, error) {
	if err := sd.prepareNext(); err != nil {
		return nil, err
	}

	d := sd.dec.d

	if d.nextCBORType() != cborTypeByteString {
		t, _ := sd.NextType()
		return nil, &WrongTypeError{t, "bytes"}
	}

	start := d.off
	_, ai, val := d.getHead()

	if ai == 31 {
		// Indefinite length byte string isn't supported in StreamDecoder.  Skip it.
		d.off = start
		_ = sd.Skip()
		return nil, errors.New("cbor: indefinite length byte string isn't supported")
	}

	b := make([]byte, int(val))
	copy(b, d.data[d.off:d.off+int(val)])
	d.off += int(val)

	end := d.off

	sd.updateState(end - start)

	return b, nil
}

// DecodeString decodes next CBOR data as string.  WrongType error is returned if type is mismatched.
func (sd *StreamDecoder) DecodeString() (string, error) {
	if err := sd.prepareNext(); err != nil {
		return "", err
	}

	d := sd.dec.d

	if d.nextCBORType() != cborTypeTextString {
		t, _ := sd.NextType()
		return "", &WrongTypeError{t, "string"}
	}

	start := d.off
	_, ai, val := d.getHead()

	if ai == 31 {
		// Indefinite length text string isn't supported in StreamDecoder.  Skip it.
		d.off = start
		_ = sd.Skip()
		return "", errors.New("cbor: indefinite length text string isn't supported")
	}

	b := d.data[d.off : d.off+int(val)]

	d.off += int(val)
	end := d.off

	sd.updateState(end - start)

	if !utf8.Valid(b) {
		return "", &SemanticError{"cbor: invalid UTF-8 string"}
	}

	return string(b), nil
}

// DecodeBigInt decodes next CBOR data as *big.Int.  WrongType error is returned if type is mismatched.
func (sd *StreamDecoder) DecodeBigInt() (*big.Int, error) {
	if err := sd.prepareNext(); err != nil {
		return nil, err
	}

	d := sd.dec.d

	t := d.data[d.off]
	if t != 0xc2 && t != 0xc3 {
		t, _ := sd.NextType()
		return nil, &WrongTypeError{t, "big.Int"}
	}

	d.off++
	sd.updateState(1)

	b, err := sd.DecodeBytes()
	if err != nil {
		return nil, err
	}

	bi := new(big.Int).SetBytes(b)

	if t == 0xc3 {
		bi.Add(bi, bigOne)
		bi.Neg(bi)
	}

	return bi, nil
}

// DecodeTagNumber decodes next CBOR data as tag number.  WrongType error is returned if type is mismatched.
func (sd *StreamDecoder) DecodeTagNumber() (uint64, error) {
	if err := sd.prepareNext(); err != nil {
		return 0, err
	}

	d := sd.dec.d

	if d.nextCBORType() != cborTypeTag {
		t, _ := sd.NextType()
		return 0, &WrongTypeError{t, "tag"}
	}

	start := d.off
	_, _, val := d.getHead()
	end := d.off

	sd.updateState(end - start)

	return val, nil
}

// DecodeArrayHead decodes next CBOR data as array size.  WrongType error is returned if type is mismatched.
func (sd *StreamDecoder) DecodeArrayHead() (uint64, error) {
	if err := sd.prepareNext(); err != nil {
		return 0, err
	}

	d := sd.dec.d

	if d.nextCBORType() != cborTypeArray {
		t, _ := sd.NextType()
		return 0, &WrongTypeError{t, "array"}
	}

	start := d.off
	_, ai, val := d.getHead()
	end := d.off

	if ai == 31 {
		// Indefinite length array isn't supported in StreamDecoder.  Skip it.
		d.off = start
		_ = sd.Skip()
		return 0, errors.New("cbor: indefinite length array isn't supported")
	}

	sd.updateState(end - start)

	return val, nil
}

// prepareNext reads and validates next CBOR data.
// prepareNext can return io error or CBOR validation error.
func (sd *StreamDecoder) prepareNext() error {
	if sd.err != nil {
		return sd.err
	}

	if sd.remainingBytes > 0 {
		return nil
	}

	return sd.readAndValidateNext()
}

func (sd *StreamDecoder) readAndValidateNext() error {

	lastMsgBytes := sd.dec.d.off

	length, err := sd._readAndValidateNext()
	if err != nil {
		sd.err = err
		return err
	}

	sd.decodedMsgBytes += lastMsgBytes

	sd.remainingBytes = length

	sd.dec.off += length
	sd.dec.bytesRead += length

	return nil
}

func (sd *StreamDecoder) _readAndValidateNext() (int, error) {
	if len(sd.dec.buf) == sd.dec.off {
		n, err := sd.dec.read()
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n == 0 {
			return 0, io.EOF
		}
	}

	for {
		sd.dec.d.reset(sd.dec.buf[sd.dec.off:])

		start := sd.dec.d.off
		err := sd.dec.d.valid()
		end := sd.dec.d.off

		// Restore decoder offset after validation
		sd.dec.d.off = start

		if err == nil {
			// Next complete data is read and validated
			return end - start, nil
		}

		if err != io.ErrUnexpectedEOF {
			return 0, err
		}

		// valid() returned io.ErrUnexpectedEOF.
		// Read more data and try again.
		n, err := sd.dec.read()
		if n == 0 {
			// No more data, it is incomplete CBOR.
			return 0, io.ErrUnexpectedEOF
		}
		if err != nil {
			return 0, err
		}
	}
}

func (sd *StreamDecoder) updateState(bytesRead int) {
	sd.remainingBytes -= bytesRead
	if sd.remainingBytes < 0 {
		sd.err = errors.New("remaining bytes are out of sync")
	}
}

// NumBytesDecoded returns the accumulated number of bytes decoded using "DecodeXXX()"
func (sd *StreamDecoder) NumBytesDecoded() int {
	return sd.decodedMsgBytes + sd.dec.d.off
}
