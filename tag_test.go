// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestTagNewTypeWithBuiltinUnderlyingType(t *testing.T) {
	type myBool bool
	type myUint uint
	type myUint8 uint8
	type myUint16 uint16
	type myUint32 uint32
	type myUint64 uint64
	type myInt int
	type myInt8 int8
	type myInt16 int16
	type myInt32 int32
	type myInt64 int64
	type myFloat32 float32
	type myFloat64 float64
	type myString string
	type myByteSlice []byte
	type myIntSlice []int
	type myIntArray [4]int
	type myMapIntInt map[int]int

	types := []reflect.Type{
		reflect.TypeFor[myBool](),
		reflect.TypeFor[myUint](),
		reflect.TypeFor[myUint8](),
		reflect.TypeFor[myUint16](),
		reflect.TypeFor[myUint32](),
		reflect.TypeFor[myUint64](),
		reflect.TypeFor[myInt](),
		reflect.TypeFor[myInt8](),
		reflect.TypeFor[myInt16](),
		reflect.TypeFor[myInt32](),
		reflect.TypeFor[myInt64](),
		reflect.TypeFor[myFloat32](),
		reflect.TypeFor[myFloat64](),
		reflect.TypeFor[myString](),
		reflect.TypeFor[myByteSlice](),
		reflect.TypeFor[myIntSlice](),
		reflect.TypeFor[myIntArray](),
		reflect.TypeFor[myMapIntInt](),
	}

	tags := NewTagSet()
	for i, typ := range types {
		tagNum := uint64(100 + i)
		mustAddTag(t, tags, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, typ, tagNum)
	}

	em := mustEncModeWithTags(t, EncOptions{Sort: SortCanonical}, tags)
	dm := mustDecModeWithTags(t, DecOptions{}, tags)

	testCases := []roundTripTestCase{
		{
			name: "bool",
			v:    myBool(true),
			data: mustHexDecode("d864f5"),
		},
		{
			name: "uint",
			v:    myUint(0),
			data: mustHexDecode("d86500"),
		},
		{
			name: "uint8",
			v:    myUint8(0),
			data: mustHexDecode("d86600"),
		},
		{
			name: "uint16",
			v:    myUint16(1000),
			data: mustHexDecode("d8671903e8"),
		},
		{
			name: "uint32",
			v:    myUint32(1000000),
			data: mustHexDecode("d8681a000f4240"),
		},
		{
			name: "uint64",
			v:    myUint64(1000000000000),
			data: mustHexDecode("d8691b000000e8d4a51000"),
		},
		{
			name: "int",
			v:    myInt(-1),
			data: mustHexDecode("d86a20"),
		},
		{
			name: "int8",
			v:    myInt8(-1),
			data: mustHexDecode("d86b20"),
		},
		{
			name: "int16",
			v:    myInt16(-1000),
			data: mustHexDecode("d86c3903e7"),
		},
		{
			name: "int32",
			v:    myInt32(-1000),
			data: mustHexDecode("d86d3903e7"),
		},
		{
			name: "int64",
			v:    myInt64(-1000),
			data: mustHexDecode("d86e3903e7"),
		},
		{
			name: "float32",
			v:    myFloat32(100000.0),
			data: mustHexDecode("d86ffa47c35000"),
		},
		{
			name: "float64",
			v:    myFloat64(1.1),
			data: mustHexDecode("d870fb3ff199999999999a"),
		},
		{
			name: "string",
			v:    myString("a"),
			data: mustHexDecode("d8716161"),
		},
		{
			name: "[]byte",
			v:    myByteSlice([]byte{1, 2, 3, 4}),
			data: mustHexDecode("d8724401020304"),
		},
		{
			name: "[]int",
			v:    myIntSlice([]int{1, 2, 3, 4}),
			data: mustHexDecode("d8738401020304"),
		},
		{
			name: "[4]int",
			v:    myIntArray([...]int{1, 2, 3, 4}),
			data: mustHexDecode("d8748401020304"),
		},
		{
			name: "map[int]int",
			v:    myMapIntInt(map[int]int{1: 2, 3: 4}),
			data: mustHexDecode("d875a201020304"),
		},
	}

	testRoundTripCases(t, testCases, em, dm)
}

func TestTagBinaryMarshalerUnmarshaler(t *testing.T) {
	t1 := reflect.TypeFor[*number]() // Use *number for testing purpose
	t2 := reflect.TypeFor[stru]()

	tags := NewTagSet()
	mustAddTag(t, tags, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, t1, 123)
	mustAddTag(t, tags, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, t2, 124)

	em := mustEncModeWithTags(t, EncOptions{}, tags)
	dm := mustDecModeWithTags(t, DecOptions{}, tags)

	testCases := []roundTripTestCase{
		{
			name: "primitive obj",
			v:    number(1234567890),
			data: mustHexDecode("d87b4800000000499602d2"),
		},
		{
			name: "struct obj",
			v:    stru{a: "a", b: "b", c: "c"},
			data: mustHexDecode("d87c45612C622C63"),
		},
	}

	testRoundTripCases(t, testCases, em, dm)
}

func TestTagStruct(t *testing.T) {
	type T struct {
		S string `cbor:"s,omitempty"`
	}

	t1 := reflect.TypeFor[T]()

	tags := NewTagSet()
	mustAddTag(t, tags, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, t1, 100)

	em := mustEncModeWithTags(t, EncOptions{}, tags)
	dm := mustDecModeWithTags(t, DecOptions{}, tags)

	data := mustHexDecode("d864a0") // {}
	var v T
	if err := unmarshalWithDM(t, dm, data, &v); err != nil {
		t.Errorf("Unmarshal() returned error %v", err)
	}
	b, err := em.Marshal(v)
	checkMarshal(t, "em.Marshal", b, err, data)
}

func TestTagFixedLengthStruct(t *testing.T) {
	type T struct {
		S string `cbor:"s"`
	}

	t1 := reflect.TypeFor[T]()

	tags := NewTagSet()
	mustAddTag(t, tags, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, t1, 100)

	em := mustEncModeWithTags(t, EncOptions{}, tags)
	dm := mustDecModeWithTags(t, DecOptions{}, tags)

	data := mustHexDecode("d864a1617360") // {"s":""}
	var v T
	if err := unmarshalWithDM(t, dm, data, &v); err != nil {
		t.Errorf("Unmarshal() returned error %v", err)
	}
	b, err := em.Marshal(v)
	checkMarshal(t, "em.Marshal", b, err, data)
}

func TestTagToArrayStruct(t *testing.T) {
	type coseHeader struct {
		Alg int    `cbor:"1,keyasint,omitempty"`
		Kid []byte `cbor:"4,keyasint,omitempty"`
		IV  []byte `cbor:"5,keyasint,omitempty"`
	}
	type signedCWT struct {
		_           struct{} `cbor:",toarray"`
		Protected   []byte
		Unprotected coseHeader
		Payload     []byte
		Signature   []byte
	}

	t1 := reflect.TypeFor[signedCWT]()

	tags := NewTagSet()
	mustAddTag(t, tags, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, t1, 18)

	em := mustEncModeWithTags(t, EncOptions{}, tags)
	dm := mustDecModeWithTags(t, DecOptions{}, tags)

	// Data from https://tools.ietf.org/html/rfc8392#appendix-A section A.3
	data := mustHexDecode("d28443a10126a104524173796d6d657472696345434453413235365850a70175636f61703a2f2f61732e6578616d706c652e636f6d02656572696b77037818636f61703a2f2f6c696768742e6578616d706c652e636f6d041a5612aeb0051a5610d9f0061a5610d9f007420b7158405427c1ff28d23fbad1f29c4c7c6a555e601d6fa29f9179bc3d7438bacaca5acd08c8d4d4f96131680c429a01f85951ecee743a52b9b63632c57209120e1c9e30")
	var v signedCWT
	if err := unmarshalWithDM(t, dm, data, &v); err != nil {
		t.Errorf("Unmarshal() returned error %v", err)
	}
	b, err := em.Marshal(v)
	checkMarshal(t, "em.Marshal", b, err, data)
}

func TestNestedTagStruct(t *testing.T) {
	type coseHeader struct {
		Alg int    `cbor:"1,keyasint,omitempty"`
		Kid []byte `cbor:"4,keyasint,omitempty"`
		IV  []byte `cbor:"5,keyasint,omitempty"`
	}
	type macedCOSE struct {
		_           struct{} `cbor:",toarray"`
		Protected   []byte
		Unprotected coseHeader
		Payload     []byte
		Tag         []byte
	}

	t1 := reflect.TypeFor[macedCOSE]()

	// Register tag CBOR Web Token (CWT) 61 and COSE_Mac0 17 with macedCOSE type
	tags := NewTagSet()
	mustAddTag(t, tags, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, t1, 61, 17)

	em := mustEncModeWithTags(t, EncOptions{}, tags)
	dm := mustDecModeWithTags(t, DecOptions{}, tags)

	// Data from https://tools.ietf.org/html/rfc8392#appendix-A section A.4
	data := mustHexDecode("d83dd18443a10104a1044c53796d6d65747269633235365850a70175636f61703a2f2f61732e6578616d706c652e636f6d02656572696b77037818636f61703a2f2f6c696768742e6578616d706c652e636f6d041a5612aeb0051a5610d9f0061a5610d9f007420b7148093101ef6d789200")
	var v macedCOSE
	if err := unmarshalWithDM(t, dm, data, &v); err != nil {
		t.Errorf("Unmarshal() returned error %v", err)
	}
	b, err := em.Marshal(v)
	checkMarshal(t, "em.Marshal", b, err, data)
}

func TestAddTagError(t *testing.T) {
	type myInt int
	testCases := []struct {
		name         string
		typ          reflect.Type
		num          uint64
		opts         TagOptions
		wantErrorMsg string
	}{
		{
			name:         "nil type",
			typ:          nil,
			num:          100,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: cannot add nil content type to TagSet",
		},
		{
			name:         "DecTag is DecTagIgnored && EncTag is EncTagNone",
			typ:          reflect.TypeFor[myInt](),
			num:          100,
			opts:         TagOptions{DecTag: DecTagIgnored, EncTag: EncTagNone},
			wantErrorMsg: "cbor: cannot add tag with DecTagIgnored and EncTagNone options to TagSet",
		},
		{
			name:         "time.Time",
			typ:          reflect.TypeFor[time.Time](),
			num:          101,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: cannot add time.Time to TagSet, use EncOptions.TimeTag and DecOptions.TimeTag instead",
		},
		{
			name:         "builtin type string",
			typ:          reflect.TypeFor[string](),
			num:          102,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: can only add named types to TagSet, got string",
		},
		{
			name:         "unnamed type struct{}",
			typ:          reflect.TypeOf(struct{}{}),
			num:          103,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: can only add named types to TagSet, got struct {}",
		},
		{
			name:         "interface",
			typ:          reflect.TypeFor[io.Reader](),
			num:          104,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: can only add named types to TagSet, got io.Reader",
		},
		{
			name:         "cbor.Tag",
			typ:          reflect.TypeFor[Tag](),
			num:          105,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: cannot add cbor.Tag to TagSet",
		},
		{
			name:         "cbor.RawTag",
			typ:          reflect.TypeFor[RawTag](),
			num:          106,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: cannot add cbor.RawTag to TagSet",
		},
		{
			name:         "big.Int",
			typ:          reflect.TypeFor[big.Int](),
			num:          107,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: cannot add big.Int to TagSet, it's built-in and supported automatically",
		},
		/*
			{
				name:         "cbor.Unmarshaler",
				typ:          reflect.TypeOf(number2(0)),
				num:          107,
				opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagNone},
				wantErrorMsg: "cbor: cannot add cbor.Unmarshaler to TagSet with DecTag != DecTagIgnored",
			},
			{
				name:         "cbor.Marshaler",
				typ:          reflect.TypeOf(number2(0)),
				num:          108,
				opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
				wantErrorMsg: "cbor: cannot add cbor.Marshaler to TagSet with EncTag != EncTagNone",
			},
		*/
		{
			name:         "tag number 0",
			typ:          reflect.TypeFor[myInt](),
			num:          0,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: cannot add tag number 0 or 1 to TagSet, use EncOptions.TimeTag and DecOptions.TimeTag instead",
		},
		{
			name:         "tag number 1",
			typ:          reflect.TypeFor[myInt](),
			num:          1,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: cannot add tag number 0 or 1 to TagSet, use EncOptions.TimeTag and DecOptions.TimeTag instead",
		},
		{
			name:         "tag number 2",
			typ:          reflect.TypeFor[myInt](),
			num:          2,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: cannot add tag number 2 or 3 to TagSet, it's built-in and supported automatically",
		},
		{
			name:         "tag number 3",
			typ:          reflect.TypeFor[myInt](),
			num:          3,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: cannot add tag number 2 or 3 to TagSet, it's built-in and supported automatically",
		},
		{
			name:         "tag number 55799",
			typ:          reflect.TypeFor[myInt](),
			num:          55799,
			opts:         TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
			wantErrorMsg: "cbor: cannot add tag number 55799 to TagSet, it's built-in and ignored automatically",
		},
	}
	tags := NewTagSet()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tags.Add(tc.opts, tc.typ, tc.num); err == nil {
				t.Errorf("TagSet.Add(%s, %d) didn't return an error", tc.typ.String(), tc.num)
			} else if err.Error() != tc.wantErrorMsg {
				var typeString string
				if tc.typ == nil {
					typeString = "nil"
				} else {
					typeString = tc.typ.String()
				}
				t.Errorf("TagSet.Add(%s, %d) returned error msg %q, want %q", typeString, tc.num, err, tc.wantErrorMsg)
			}
		})
	}
}

func TestAddDuplicateTagContentTypeError(t *testing.T) {
	type myInt int
	myIntType := reflect.TypeFor[myInt]()
	wantErrorMsg := "cbor: content type cbor.myInt already exists in TagSet"

	tags := NewTagSet()
	// Add myIntType and 100 to tags
	mustAddTag(t, tags, TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired}, myIntType, 100)
	// Add myIntType and 101 to tags
	if err := tags.Add(TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired}, myIntType, 101); err == nil {
		t.Errorf("TagSet.Add(%s, %d) didn't return an error", myIntType.String(), 101)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("TagSet.Add(%s, %d) returned error msg %q, want %q", myIntType, 101, err, wantErrorMsg)
	}
}

func TestAddDuplicateTagNumError(t *testing.T) {
	type myBool bool
	type myInt int
	myBoolType := reflect.TypeFor[myBool]()
	myIntType := reflect.TypeFor[myInt]()
	wantErrorMsg := "cbor: tag number [100] already exists in TagSet"

	tags := NewTagSet()

	// Add myIntType and 100 to tags
	mustAddTag(t, tags, TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired}, myIntType, 100)
	// Add myBoolType and 100 to tags
	if err := tags.Add(TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired}, myBoolType, 100); err == nil {
		t.Errorf("TagSet.Add(%s, %d) didn't return an error", myBoolType.String(), 100)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("TagSet.Add(%s, %d) returned error msg %q, want %q", myBoolType, 100, err, wantErrorMsg)
	}
}

func TestAddDuplicateTagNumsError(t *testing.T) {
	type myBool bool
	type myInt int
	myBoolType := reflect.TypeFor[myBool]()
	myIntType := reflect.TypeFor[myInt]()
	wantErrorMsg := "cbor: tag number [100 101] already exists in TagSet"

	tags := NewTagSet()

	// Add myIntType and [100, 101] to tags
	mustAddTag(t, tags, TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired}, myIntType, 100, 101)
	// Add myBoolType and [100, 101] to tags
	if err := tags.Add(TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired}, myBoolType, 100, 101); err == nil {
		t.Errorf("TagSet.Add(%s, %d, %d) didn't return an error", myBoolType.String(), 100, 101)
	} else if err.Error() != wantErrorMsg {
		t.Errorf("TagSet.Add(%s, %d, %d) returned error msg %q, want %q", myBoolType, 100, 101, err, wantErrorMsg)
	}
}

func TestAddRemoveTag(t *testing.T) {
	type myInt int
	type myFloat float64
	myIntType := reflect.TypeFor[myInt]()
	myFloatType := reflect.TypeFor[myFloat]()
	pMyIntType := reflect.TypeFor[*myInt]()
	pMyFloatType := reflect.TypeFor[*myFloat]()

	tags := NewTagSet()
	stags := tags.(*syncTagSet)
	mustAddTag(t, tags, TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired}, myIntType, 100)
	mustAddTag(t, tags, TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired}, myFloatType, 101)
	if err := tags.Add(TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired}, pMyIntType, 102); err == nil {
		t.Errorf("TagSet.Add(%s, %d) didn't return an error", pMyIntType.String(), 102)
	}
	if err := tags.Add(TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired}, pMyFloatType, 103); err == nil {
		t.Errorf("TagSet.Add(%s, %d) didn't return an error", pMyFloatType.String(), 103)
	}
	if len(stags.t) != 2 {
		t.Errorf("TagSet len is %d, want %d", len(stags.t), 2)
	}
	tags.Remove(pMyIntType)
	if len(stags.t) != 1 {
		t.Errorf("TagSet len is %d, want %d", len(stags.t), 1)
	}
	tags.Remove(pMyFloatType)
	if len(stags.t) != 0 {
		t.Errorf("TagSet len is %d, want %d", len(stags.t), 0)
	}
	tags.Remove(myIntType)
	tags.Remove(myFloatType)
}

func TestTagSetRemoveNil(t *testing.T) {
	type myInt int
	myIntType := reflect.TypeFor[myInt]()

	tags := NewTagSet()
	stags := tags.(*syncTagSet)

	// Remove nil contentType from an empty TagSet
	tags.Remove(nil)
	if len(stags.t) != 0 {
		t.Errorf("TagSet len is %d, want %d", len(stags.t), 0)
	}

	mustAddTag(t, tags, TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired}, myIntType, 100)
	if len(stags.t) != 1 {
		t.Errorf("TagSet len is %d, want %d", len(stags.t), 1)
	}

	// Remove nil contentType from non-empty TagSet
	tags.Remove(nil)
	if len(stags.t) != 1 {
		t.Errorf("TagSet len is %d, want %d", len(stags.t), 1)
	}

	expectedTI := &tagItem{
		num:         []uint64{100},
		cborTagNum:  []byte{0xd8, 0x64},
		contentType: myIntType,
		opts:        TagOptions{DecTag: DecTagRequired, EncTag: EncTagRequired},
	}
	ti := tags.getTagItemFromType(myIntType)
	if !reflect.DeepEqual(ti, expectedTI) {
		t.Errorf("tagItem is %+v, want %+v", ti, expectedTI)
	}
}

func TestAddTagTypeAliasError(t *testing.T) {
	type myBool = bool
	type myUint = uint
	type myUint8 = uint8
	type myUint16 = uint16
	type myUint32 = uint32
	type myUint64 = uint64
	type myInt = int
	type myInt8 = int8
	type myInt16 = int16
	type myInt32 = int32
	type myInt64 = int64
	type myFloat32 = float32
	type myFloat64 = float64
	type myString = string
	type myByteSlice = []byte
	type myIntSlice = []int
	type myIntArray = [4]int
	type myMapIntInt = map[int]int

	testCases := []struct {
		name         string
		typ          reflect.Type
		wantErrorMsg string
	}{
		{
			name:         "bool",
			typ:          reflect.TypeFor[myBool](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got bool",
		},
		{
			name:         "uint",
			typ:          reflect.TypeFor[myUint](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got uint",
		},
		{
			name:         "uint8",
			typ:          reflect.TypeFor[myUint8](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got uint8",
		},
		{
			name:         "uint16",
			typ:          reflect.TypeFor[myUint16](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got uint16",
		},
		{
			name:         "uint32",
			typ:          reflect.TypeFor[myUint32](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got uint32",
		},
		{
			name:         "uint64",
			typ:          reflect.TypeFor[myUint64](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got uint64",
		},
		{
			name:         "int",
			typ:          reflect.TypeFor[myInt](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got int",
		},
		{
			name:         "int8",
			typ:          reflect.TypeFor[myInt8](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got int8",
		},
		{
			name:         "int16",
			typ:          reflect.TypeFor[myInt16](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got int16",
		},
		{
			name:         "int32",
			typ:          reflect.TypeFor[myInt32](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got int32",
		},
		{
			name:         "int64",
			typ:          reflect.TypeFor[myInt64](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got int64",
		},
		{
			name:         "float32",
			typ:          reflect.TypeFor[myFloat32](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got float32",
		},
		{
			name:         "float64",
			typ:          reflect.TypeFor[myFloat64](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got float64",
		},
		{
			name:         "string",
			typ:          reflect.TypeFor[myString](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got string",
		},
		{
			name:         "[]byte",
			typ:          reflect.TypeFor[myByteSlice](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got []uint8",
		},
		{
			name:         "[]int",
			typ:          reflect.TypeFor[myIntSlice](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got []int",
		},
		{
			name:         "[4]int",
			typ:          reflect.TypeFor[myIntArray](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got [4]int",
		},
		{
			name:         "map[int]int",
			typ:          reflect.TypeFor[myMapIntInt](),
			wantErrorMsg: "cbor: can only add named types to TagSet, got map[int]int",
		},
	}

	tags := NewTagSet()
	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tags.Add(TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, tc.typ, uint64(100+i)); err == nil { //nolint:gosec
				t.Errorf("TagSet.Add(%s, %d) didn't return an error", tc.typ.String(), 0)
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("TagSet.Add(%s, %d) returned error msg %q, want %q", tc.typ.String(), 0, err, tc.wantErrorMsg)
			}
		})
	}
}

// TestDecodeTag decodes tag data with DecTagRequired/EncTagOptional/EncTagNone options.
func TestDecodeTagData(t *testing.T) {
	type myInt int
	type s struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
	}

	type tagInfo struct {
		t reflect.Type
		n []uint64
	}
	tagInfos := []tagInfo{
		{reflect.TypeFor[*number](), []uint64{123}}, // BinaryMarshaler *number
		{reflect.TypeFor[stru](), []uint64{124}},    // BinaryMarshaler stru
		{reflect.TypeFor[myInt](), []uint64{125}},   // non-struct type
		{reflect.TypeFor[s](), []uint64{126}},       // struct type
	}

	tagsDecRequired := NewTagSet()
	tagsDecOptional := NewTagSet()
	tagsDecIgnored := NewTagSet()
	for _, tag := range tagInfos {
		mustAddTag(t, tagsDecRequired, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, tag.t, tag.n[0], tag.n[1:]...)
		mustAddTag(t, tagsDecOptional, TagOptions{EncTag: EncTagRequired, DecTag: DecTagOptional}, tag.t, tag.n[0], tag.n[1:]...)
		mustAddTag(t, tagsDecIgnored, TagOptions{EncTag: EncTagRequired, DecTag: DecTagIgnored}, tag.t, tag.n[0], tag.n[1:]...)
	}

	type tag struct {
		name   string
		tagSet TagSet
	}
	tags := []tag{
		{"EncTagRequired_DecTagRequired", tagsDecRequired},
		{"EncTagRequired_DecTagOptional", tagsDecOptional},
		{"EncTagRequired_DecTagIgnored", tagsDecIgnored},
	}

	testCases := []roundTripTestCase{
		{
			name: "BinaryMarshaler non-struct",
			v:    number(1234567890),
			data: mustHexDecode("d87b4800000000499602d2"),
		},
		{
			name: "BinaryMarshaler struct",
			v:    stru{a: "a", b: "b", c: "c"},
			data: mustHexDecode("d87c45612C622C63"),
		},
		{
			name: "non-struct",
			v:    myInt(1),
			data: mustHexDecode("d87d01"),
		},
		{
			name: "struct",
			v:    s{A: "A", B: "B", C: "C"},
			data: mustHexDecode("d87ea3616161416162614261636143"), // {"a":"A", "b":"B", "c":"C"}
		},
	}
	for _, tag := range tags {
		t.Run(tag.name, func(t *testing.T) {
			em := mustEncModeWithTags(t, EncOptions{}, tag.tagSet)
			dm := mustDecModeWithTags(t, DecOptions{}, tag.tagSet)
			testRoundTripCases(t, testCases, em, dm)
		})
	}
}

// TestDecodeNoTag decodes no-tag data with DecTagRequired/EncTagOptional/EncTagNone options
func TestDecodeNoTagData(t *testing.T) {
	type myInt int
	type s struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
	}

	type tagInfo struct {
		t reflect.Type
		n []uint64
	}
	tagInfos := []tagInfo{
		{reflect.TypeFor[*number](), []uint64{123}}, // BinaryMarshaler *number
		{reflect.TypeFor[stru](), []uint64{124}},    // BinaryMarshaler stru
		{reflect.TypeFor[myInt](), []uint64{125}},   // non-struct type
		{reflect.TypeFor[s](), []uint64{126}},       // struct type
	}

	tagsDecRequired := NewTagSet()
	tagsDecOptional := NewTagSet()
	for _, tag := range tagInfos {
		mustAddTag(t, tagsDecRequired, TagOptions{EncTag: EncTagNone, DecTag: DecTagRequired}, tag.t, tag.n[0], tag.n[1:]...)
		mustAddTag(t, tagsDecOptional, TagOptions{EncTag: EncTagNone, DecTag: DecTagOptional}, tag.t, tag.n[0], tag.n[1:]...)
	}

	type tag struct {
		name   string
		tagSet TagSet
	}
	tags := []tag{
		{"EncTagIgnored_DecTagOptional", tagsDecOptional},
	}

	testCases := []roundTripTestCase{
		{
			name: "BinaryMarshaler non-struct",
			v:    number(1234567890),
			data: mustHexDecode("4800000000499602d2"),
		},
		{
			name: "BinaryMarshaler struct",
			v:    stru{a: "a", b: "b", c: "c"},
			data: mustHexDecode("45612C622C63"),
		},
		{
			name: "non-struct",
			v:    myInt(1),
			data: mustHexDecode("01"),
		},
		{
			name: "struct",
			v:    s{A: "A", B: "B", C: "C"},
			data: mustHexDecode("a3616161416162614261636143"), // {"a":"A", "b":"B", "c":"C"}
		},
	}

	for _, tag := range tags {
		t.Run(tag.name, func(t *testing.T) {
			em := mustEncModeWithTags(t, EncOptions{}, tag.tagSet)
			dm := mustDecModeWithTags(t, DecOptions{}, tag.tagSet)
			testRoundTripCases(t, testCases, em, dm)
		})
	}

	// Decode non-tag data with DecTagRequired option returns UnmarshalTypeError
	for _, tc := range testCases {
		name := "EncTagIgnored_DecTagRequired " + tc.name
		t.Run(name, func(t *testing.T) {
			dm := mustDecModeWithTags(t, DecOptions{}, tagsDecRequired)
			rv := reflect.New(reflect.TypeOf(tc.v))
			err := unmarshalWithDM(t, dm, tc.data, rv.Interface())
			checkError[*UnmarshalTypeError](t, "dm.Unmarshal", err, "")
			if !strings.Contains(err.Error(), "expect CBOR tag value") {
				t.Errorf("dm.Unmarshal: got %q, want substring %q", err.Error(), "expect CBOR tag value")
			}
		})
	}
}

// TestDecodeWrongTag decodes wrong tag data with DecTagRequired/EncTagOptional/EncTagNone options
func TestDecodeWrongTag(t *testing.T) {
	type myInt int
	type s struct {
		A string `cbor:"a"`
		B string `cbor:"b"`
		C string `cbor:"c"`
	}

	type tagInfo struct {
		t reflect.Type
		n []uint64
	}
	tagInfos := []tagInfo{
		{reflect.TypeFor[*number](), []uint64{123}}, // BinaryMarshaler *number
		{reflect.TypeFor[stru](), []uint64{124}},    // BinaryMarshaler stru
		{reflect.TypeFor[myInt](), []uint64{100}},   // non-struct type
		{reflect.TypeFor[s](), []uint64{101, 102}},  // struct type
	}

	tagsDecRequired := NewTagSet()
	tagsDecOptional := NewTagSet()
	tagsDecIgnored := NewTagSet()
	for _, tag := range tagInfos {
		mustAddTag(t, tagsDecRequired, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, tag.t, tag.n[0], tag.n[1:]...)
		mustAddTag(t, tagsDecOptional, TagOptions{EncTag: EncTagRequired, DecTag: DecTagOptional}, tag.t, tag.n[0], tag.n[1:]...)
		mustAddTag(t, tagsDecIgnored, TagOptions{EncTag: EncTagRequired, DecTag: DecTagIgnored}, tag.t, tag.n[0], tag.n[1:]...)
	}
	type tag struct {
		name   string
		tagSet TagSet
	}
	tags := []tag{
		{"EncTagRequired_DecTagRequired", tagsDecRequired},
		{"EncTagRequired_DecTagOptional", tagsDecOptional},
	}

	testCases := []struct {
		name         string
		obj          any
		data         []byte
		wantErrorMsg string
	}{
		{
			name:         "BinaryMarshaler non-struct",
			obj:          number(1234567890),
			data:         mustHexDecode("d87d4800000000499602d2"),
			wantErrorMsg: "cbor: wrong tag number for cbor.number, got [125], expected [123]",
		},
		{
			name:         "BinaryMarshaler struct",
			obj:          stru{a: "a", b: "b", c: "c"},
			data:         mustHexDecode("d87d45612C622C63"),
			wantErrorMsg: "cbor: wrong tag number for cbor.stru, got [125], expected [124]",
		},
		{
			name:         "non-struct",
			obj:          myInt(1),
			data:         mustHexDecode("d87d01"),
			wantErrorMsg: "cbor: wrong tag number for cbor.myInt, got [125], expected [100]",
		},
		{
			name:         "struct",
			obj:          s{A: "A", B: "B", C: "C"},
			data:         mustHexDecode("d87ea3616161416162614261636143"), // {"a":"A", "b":"B", "c":"C"}
			wantErrorMsg: "cbor: wrong tag number for cbor.s, got [126], expected [101 102]",
		},
	}

	for _, tc := range testCases {
		for _, tag := range tags {
			name := tag.name + " " + tc.name
			t.Run(name, func(t *testing.T) {
				dm := mustDecModeWithTags(t, DecOptions{}, tag.tagSet)
				v := reflect.New(reflect.TypeOf(tc.obj))
				err := unmarshalWithDM(t, dm, tc.data, v.Interface())
				checkError[*WrongTagError](t, "dm.Unmarshal", err, tc.wantErrorMsg)
			})
		}
	}

	// Decode wrong tag data with DecTagIgnored option returns no error.
	for _, tc := range testCases {
		name := "EncTagRequired_DecTagIgnored " + tc.name
		t.Run(name, func(t *testing.T) {
			dm := mustDecModeWithTags(t, DecOptions{}, tagsDecIgnored)
			v := reflect.New(reflect.TypeOf(tc.obj))
			err := unmarshalWithDM(t, dm, tc.data, v.Interface())
			checkUnmarshal(t, "dm.Unmarshal", v.Elem().Interface(), err, tc.obj)
		})
	}
}

func TestEncodeSharedTag(t *testing.T) {
	type myInt int

	myIntType := reflect.TypeFor[myInt]()

	sharedTagSet := NewTagSet()

	em, err := EncOptions{}.EncModeWithSharedTags(sharedTagSet)
	if err != nil {
		t.Errorf("EncModeWithSharedTags() returned error %v", err)
	}

	// Register myInt type with tag number 123
	mustAddTag(t, sharedTagSet, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, myIntType, 123)

	// Encode myInt with tag number 123
	v := myInt(1)
	wantCborData := mustHexDecode("d87b01")
	b, err := em.Marshal(v)
	checkMarshal(t, "em.Marshal", b, err, wantCborData)

	// Unregister myInt type
	sharedTagSet.Remove(myIntType)

	// Encode myInt without tag number 123
	v = myInt(2)
	wantCborData = mustHexDecode("02")
	b, err = em.Marshal(v)
	checkMarshal(t, "em.Marshal", b, err, wantCborData)

	// Register myInt type with tag number 234
	mustAddTag(t, sharedTagSet, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, myIntType, 234)

	// Encode myInt with tag number 234
	v = myInt(3)
	wantCborData = mustHexDecode("d8ea03")
	b, err = em.Marshal(v)
	checkMarshal(t, "em.Marshal", b, err, wantCborData)
}

func TestDecodeSharedTag(t *testing.T) {
	type myInt int

	myIntType := reflect.TypeFor[myInt]()

	sharedTagSet := NewTagSet()

	dm, err := DecOptions{}.DecModeWithSharedTags(sharedTagSet)
	if err != nil {
		t.Errorf("DecModeWithSharedTags() returned error %v", err)
	}

	// Register myInt type with tag number 123
	mustAddTag(t, sharedTagSet, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, myIntType, 123)

	// Decode myInt with tag number 123
	var v myInt
	wantV := myInt(1)
	data := mustHexDecode("d87b01")
	err = unmarshalWithDM(t, dm, data, &v)
	checkUnmarshal(t, "dm.Unmarshal", v, err, wantV)

	// Unregister myInt type
	sharedTagSet.Remove(myIntType)

	// Decode myInt without tag number
	wantV = myInt(2)
	data = mustHexDecode("02")
	err = unmarshalWithDM(t, dm, data, &v)
	checkUnmarshal(t, "dm.Unmarshal", v, err, wantV)

	// Register myInt type with tag number 234
	mustAddTag(t, sharedTagSet, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, myIntType, 234)

	// Decode myInt with tag number 234
	wantV = myInt(3)
	data = mustHexDecode("d8ea03")
	err = unmarshalWithDM(t, dm, data, &v)
	checkUnmarshal(t, "dm.Unmarshal", v, err, wantV)
}

func TestDecModeWithTagsError(t *testing.T) {
	// Create DecMode with nil as TagSet
	wantErrorMsg := "cbor: cannot create DecMode with nil value as TagSet"
	dm, err := DecOptions{}.DecModeWithTags(nil)
	if dm != nil {
		t.Errorf("DecModeWithTags(nil) returned %v", dm)
	}
	if err.Error() != wantErrorMsg {
		t.Errorf("DecModeWithTags(nil) returned error %q, want %q", err.Error(), wantErrorMsg)
	}
	dm, err = DecOptions{}.DecModeWithSharedTags(nil)
	if dm != nil {
		t.Errorf("DecModeWithSharedTags(nil) returned %v", dm)
	}
	if err.Error() != wantErrorMsg {
		t.Errorf("DecModeWithSharedTags(nil) returned error %q, want %q", err.Error(), wantErrorMsg)
	}

	// Create DecMode with invalid EncOptions
	wantErrorMsg = "cbor: invalid TimeTag 100"
	dm, err = DecOptions{TimeTag: 100}.DecModeWithTags(NewTagSet())
	if dm != nil {
		t.Errorf("DecModeWithTags() returned %v", dm)
	}
	if err.Error() != wantErrorMsg {
		t.Errorf("DecModeWithTags() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
	dm, err = DecOptions{TimeTag: 100}.DecModeWithSharedTags(NewTagSet())
	if dm != nil {
		t.Errorf("DecModeWithSharedTags() returned %v", dm)
	}
	if err.Error() != wantErrorMsg {
		t.Errorf("DecModeWithSharedTags() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestEncModeWithTagsError(t *testing.T) {
	// Create EncMode with nil as TagSet
	wantErrorMsg := "cbor: cannot create EncMode with nil value as TagSet"
	em, err := EncOptions{}.EncModeWithTags(nil)
	if em != nil {
		t.Errorf("EncModeWithTags(nil) returned %v", em)
	}
	if err.Error() != wantErrorMsg {
		t.Errorf("EncModeWithTags(nil) returned error %q, want %q", err.Error(), wantErrorMsg)
	}
	em, err = EncOptions{}.EncModeWithSharedTags(nil)
	if em != nil {
		t.Errorf("EncModeWithSharedTags(nil) returned %v", em)
	}
	if err.Error() != wantErrorMsg {
		t.Errorf("EncModeWithSharedTags(nil) returned error %q, want %q", err.Error(), wantErrorMsg)
	}

	// Create EncMode with invalid EncOptions
	wantErrorMsg = "cbor: invalid TimeTag 100"
	em, err = EncOptions{TimeTag: 100}.EncModeWithTags(NewTagSet())
	if em != nil {
		t.Errorf("EncModeWithTags() returned %v", em)
	}
	if err.Error() != wantErrorMsg {
		t.Errorf("EncModeWithTags() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
	em, err = EncOptions{TimeTag: 100}.EncModeWithSharedTags(NewTagSet())
	if em != nil {
		t.Errorf("EncModeWithSharedTags() returned %v", em)
	}
	if err.Error() != wantErrorMsg {
		t.Errorf("EncModeWithSharedTags() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestNilRawTagUnmarshalCBORError(t *testing.T) {
	wantErrorMsg := "cbor.RawTag: UnmarshalCBOR on nil pointer"
	var tag *RawTag
	data := mustHexDecode("c249010000000000000000")
	if err := tag.UnmarshalCBOR(data); err == nil {
		t.Errorf("UnmarshalCBOR() didn't return error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("UnmarshalCBOR() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestTagUnmarshalError(t *testing.T) {
	data := mustHexDecode("d87b61fe") // invalid UTF-8 string
	var tag Tag
	err := unmarshal(t, data, &tag)
	checkError[error](t, "Unmarshal", err, invalidUTF8ErrorMsg)
}

func TestTagMarshalError(t *testing.T) {
	wantErrorMsg := "cbor: unsupported type: chan bool"
	tag := Tag{
		Number:  123,
		Content: make(chan bool),
	}
	b, err := Marshal(tag)
	checkMarshalErr[error](t, "Marshal", b, err, wantErrorMsg)
}

func TestMarshalUninitializedTag(t *testing.T) {
	var v Tag
	b, err := Marshal(v)
	checkMarshal(t, "Marshal", b, err, []byte{cborNil})
}

func TestMarshalUninitializedRawTag(t *testing.T) {
	var v RawTag
	b, err := Marshal(v)
	checkMarshal(t, "Marshal", b, err, []byte{cborNil})
}

func TestMarshalTagWithEmptyContent(t *testing.T) {
	v := Tag{Number: 100}           // Tag.Content is empty
	want := mustHexDecode("d864f6") // 100(null)
	b, err := Marshal(v)
	checkMarshal(t, "Marshal", b, err, want)
}

func TestMarshalRawTagWithEmptyContent(t *testing.T) {
	v := RawTag{Number: 100}        // RawTag.Content is empty
	want := mustHexDecode("d864f6") // 100(null)
	b, err := Marshal(v)
	checkMarshal(t, "Marshal", b, err, want)
}

func TestMarshalAddressableTag(t *testing.T) {
	type s struct{ T Tag }

	v := s{T: Tag{Number: 100, Content: 1}}
	want := []byte{0xa1, 0x61, 0x54, 0xd8, 0x64, 0x01} // {"T": 100(1)}
	b, err := Marshal(&v)
	checkMarshal(t, "Marshal", b, err, want)
}

func TestEncodeTag(t *testing.T) {
	m := make(map[any]bool)
	m[10] = true
	m[100] = true
	m[-1] = true
	m["z"] = true
	m["aa"] = true
	m[[1]int{100}] = true
	m[[1]int{-1}] = true
	m[false] = true

	v := Tag{100, m}

	lenFirstSortedCborData := mustHexDecode("d864a80af520f5f4f51864f5617af58120f5626161f5811864f5") // tag number: 100, value: map with sorted keys: 10, -1, false, 100, "z", [-1], "aa", [100]
	bytewiseSortedCborData := mustHexDecode("d864a80af51864f520f5617af5626161f5811864f58120f5f4f5") // tag number: 100, value: map with sorted keys: 10, 100, -1, "z", "aa", [100], [-1], false

	em := mustEncMode(t, EncOptions{Sort: SortLengthFirst})
	b, err := em.Marshal(v)
	checkMarshal(t, "em.Marshal", b, err, lenFirstSortedCborData)

	em = mustEncMode(t, EncOptions{Sort: SortBytewiseLexical})
	b, err = em.Marshal(v)
	checkMarshal(t, "em.Marshal", b, err, bytewiseSortedCborData)
}

func TestDecodeTagToEmptyIface(t *testing.T) {
	type myBool bool
	type myUint uint

	typeMyBool := reflect.TypeFor[myBool]()
	typeMyUint := reflect.TypeFor[myUint]()

	tags := NewTagSet()
	mustAddTag(t, tags, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, typeMyBool, 100)
	mustAddTag(t, tags, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, typeMyUint, 101, 102)

	dm := mustDecModeWithTags(t, DecOptions{}, tags)
	dmSharedTags, _ := DecOptions{}.DecModeWithSharedTags(tags)

	testCases := []struct {
		name    string
		data    []byte
		wantObj any
	}{
		{
			name:    "registered myBool",
			data:    mustHexDecode("d864f5"), // 100(true)
			wantObj: myBool(true),
		},
		{
			name:    "registered myUint",
			data:    mustHexDecode("d865d86600"), // 101(102(0))
			wantObj: myUint(0),
		},
		{
			name:    "not registered bool",
			data:    mustHexDecode("d865f5"), // 101(true)
			wantObj: Tag{101, true},
		},
		{
			name:    "not registered uint",
			data:    mustHexDecode("d865d86700"), // 101(103(0))
			wantObj: Tag{101, Tag{103, uint64(0)}},
		},
		{
			name:    "not registered uint",
			data:    mustHexDecode("d865d866d86700"), // 101(102(103(0)))
			wantObj: Tag{101, Tag{102, Tag{103, uint64(0)}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v1 any
			err := unmarshalWithDM(t, dm, tc.data, &v1)
			checkUnmarshal(t, "dm.Unmarshal", v1, err, tc.wantObj)

			var v2 any
			err = unmarshalWithDM(t, dmSharedTags, tc.data, &v2)
			checkUnmarshal(t, "dm.Unmarshal", v2, err, tc.wantObj)
		})
	}
}

func TestDecodeRegisteredTagToEmptyIfaceError(t *testing.T) {
	type myInt int

	typeMyInt := reflect.TypeFor[myInt]()

	tags := NewTagSet()
	mustAddTag(t, tags, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, typeMyInt, 101, 102)

	dm := mustDecModeWithTags(t, DecOptions{}, tags)

	data := mustHexDecode("d865d8663bffffffffffffffff") // 101(102(-18446744073709551616))

	var v any
	err := unmarshalWithDM(t, dm, data, &v)
	checkError[*UnmarshalTypeError](t, "dm.Unmarshal", err, "")
}

type number3 uint64

// MarshalCBOR marshals number3 to CBOR tagged map (tag number 100)
func (n number3) MarshalCBOR() (data []byte, err error) {
	m := map[string]uint64{"num": uint64(n)}
	return Marshal(Tag{100, m})
}

// UnmarshalCBOR unmarshals CBOR tagged map to number3
func (n *number3) UnmarshalCBOR(data []byte) (err error) {
	var rawTag RawTag
	if err := Unmarshal(data, &rawTag); err != nil {
		return err
	}

	if rawTag.Number != 100 {
		return fmt.Errorf("wrong tag number %d, want %d", rawTag.Number, 100)
	}

	if getType(rawTag.Content[0]) != cborTypeMap {
		return errors.New("wrong tag content type, want map")
	}

	var v map[string]uint64
	if err := Unmarshal(rawTag.Content, &v); err != nil {
		return err
	}
	*n = number3(v["num"])
	return nil
}

func TestDecodeRegisterTagForUnmarshaler(t *testing.T) {
	typ := reflect.TypeFor[number3]()

	tags := NewTagSet()
	mustAddTag(t, tags, TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, typ, 100)

	data := mustHexDecode("d864a1636e756d01") // 100({"num": 1})
	wantObj := number3(1)

	em := mustEncModeWithTags(t, EncOptions{}, tags)
	dm := mustDecModeWithTags(t, DecOptions{}, tags)

	// Decode to a value of type any.  Unmarshal() should return object of registered type.
	var v1 any
	err := unmarshalWithDM(t, dm, data, &v1)
	checkUnmarshal(t, "dm.Unmarshal", v1, err, wantObj)
	b, err := em.Marshal(v1)
	checkMarshal(t, "em.Marshal", b, err, data)

	// Decode to registered type.
	var v2 number3
	err = unmarshalWithDM(t, dm, data, &v2)
	checkUnmarshal(t, "dm.Unmarshal", v2, err, wantObj)
	b, err = em.Marshal(v2)
	checkMarshal(t, "em.Marshal", b, err, data)
}

func TestMarshalRawTagContainingMalformedCBORData(t *testing.T) {
	testCases := []struct {
		name         string
		value        any
		wantErrorMsg string
	}{
		// Nil RawMessage and empty RawMessage are encoded as CBOR nil.
		{
			name:         "truncated data",
			value:        RawTag{Number: 100, Content: RawMessage{0xa6}},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.RawTag: unexpected EOF",
		},
		{
			name:         "malformed data",
			value:        RawTag{Number: 100, Content: RawMessage{0x1f}},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.RawTag: cbor: invalid additional information 31 for type positive integer",
		},
		{
			name:         "extraneous data",
			value:        RawTag{Number: 100, Content: RawMessage{0x01, 0x01}},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.RawTag: cbor: 1 bytes of extraneous data starting at index 3",
		},
		{
			name:         "invalid builtin tag",
			value:        RawTag{Number: 0, Content: RawMessage{0x01}},
			wantErrorMsg: "cbor: error calling MarshalCBOR for type cbor.RawTag: cbor: tag number 0 must be followed by text string, got positive integer",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.value)
			checkMarshalErr[*MarshalerError](t, "Marshal", b, err, tc.wantErrorMsg)
		})
	}
}

// TestEncodeBuiltinTag tests that marshaling a value of type Tag "does the right thing" when
// marshaling the enclosed data item of a built-in tag number.
func TestEncodeBuiltinTag(t *testing.T) {
	testCases := []struct {
		name string
		tag  Tag
		opts EncOptions
		want []byte
	}{
		{
			name: "unsigned bignum content not enclosed in expected encoding tag",
			tag:  Tag{Number: tagNumUnsignedBignum, Content: []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			opts: EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase16},
			want: []byte{0xc2, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "negative bignum content not enclosed in expected encoding tag",
			tag:  Tag{Number: tagNumNegativeBignum, Content: []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			opts: EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase16},
			want: []byte{0xc3, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "rfc 3339 content is not encoded as byte string",
			tag:  Tag{Number: tagNumRFC3339Time, Content: "2013-03-21T20:04:00Z"},
			opts: EncOptions{String: StringToByteString},
			want: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			em := mustEncMode(t, tc.opts)
			got, err := em.Marshal(tc.tag)
			checkMarshal(t, "em.Marshal", got, err, tc.want)
		})
	}
}

func TestUnmarshalRawTagOnBadData(t *testing.T) {
	testCases := []struct {
		name         string
		data         []byte
		wantErrorMsg string
	}{
		// Empty data
		{
			name:         "nil data",
			data:         nil,
			wantErrorMsg: io.EOF.Error(),
		},
		{
			name:         "empty data",
			data:         []byte{},
			wantErrorMsg: io.EOF.Error(),
		},

		// Wrong CBOR types
		{
			name:         "uint type",
			data:         mustHexDecode("01"),
			wantErrorMsg: "cbor: cannot unmarshal positive integer into Go value of type cbor.RawTag",
		},
		{
			name:         "int type",
			data:         mustHexDecode("20"),
			wantErrorMsg: "cbor: cannot unmarshal negative integer into Go value of type cbor.RawTag",
		},
		{
			name:         "byte string type",
			data:         mustHexDecode("40"),
			wantErrorMsg: "cbor: cannot unmarshal byte string into Go value of type cbor.RawTag",
		},
		{
			name:         "string type",
			data:         mustHexDecode("60"),
			wantErrorMsg: "cbor: cannot unmarshal UTF-8 text string into Go value of type cbor.RawTag",
		},
		{
			name:         "array type",
			data:         mustHexDecode("80"),
			wantErrorMsg: "cbor: cannot unmarshal array into Go value of type cbor.RawTag",
		},
		{
			name:         "map type",
			data:         mustHexDecode("a0"),
			wantErrorMsg: "cbor: cannot unmarshal map into Go value of type cbor.RawTag",
		},
		{
			name:         "primitive type",
			data:         mustHexDecode("f4"),
			wantErrorMsg: "cbor: cannot unmarshal primitives into Go value of type cbor.RawTag",
		},
		{
			name:         "float type",
			data:         mustHexDecode("f90000"),
			wantErrorMsg: "cbor: cannot unmarshal primitives into Go value of type cbor.RawTag",
		},

		// Truncated CBOR data
		{
			name:         "truncated head",
			data:         mustHexDecode("18"),
			wantErrorMsg: io.ErrUnexpectedEOF.Error(),
		},

		// Truncated CBOR tag data
		{
			name:         "truncated tag number",
			data:         mustHexDecode("d8"),
			wantErrorMsg: io.ErrUnexpectedEOF.Error(),
		},
		{
			name:         "tag number not followed by tag content",
			data:         mustHexDecode("da"),
			wantErrorMsg: io.ErrUnexpectedEOF.Error(),
		},
		{
			name:         "truncated tag content",
			data:         mustHexDecode("c074323031332d30332d32315432303a30343a3030"),
			wantErrorMsg: io.ErrUnexpectedEOF.Error(),
		},

		// Extraneous CBOR data
		{
			name:         "extraneous data",
			data:         mustHexDecode("c074323031332d30332d32315432303a30343a30305a00"),
			wantErrorMsg: "cbor: 1 bytes of extraneous data starting at index 22",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test RawTag.UnmarshalCBOR(data)
			{
				var v RawTag

				err := v.UnmarshalCBOR(tc.data)
				if err == nil {
					t.Errorf("UnmarshalCBOR(%x) didn't return error", tc.data)
				}
				if !strings.HasPrefix(err.Error(), tc.wantErrorMsg) {
					t.Errorf("UnmarshalCBOR(%x) returned error %q, want %q", tc.data, err.Error(), tc.wantErrorMsg)
				}
			}
			// Test Unmarshal(data, *RawTag), which calls RawTag.unmarshalCBOR() under the hood
			{
				var v RawTag
				err := unmarshal(t, tc.data, &v)
				checkError[error](t, "Unmarshal", err, tc.wantErrorMsg)
			}
		})
	}
}
