// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"reflect"
	"testing"
)

func TestDecModeInvalidOptions(t *testing.T) {
	type myUint8 uint8
	type myUint8Slice []myUint8

	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "DupMapKey below range of valid modes",
			opts:         DecOptions{DupMapKey: -1},
			wantErrorMsg: "cbor: invalid DupMapKey -1",
		},
		{
			name:         "DupMapKey above range of valid modes",
			opts:         DecOptions{DupMapKey: 101},
			wantErrorMsg: "cbor: invalid DupMapKey 101",
		},
		{
			name:         "TimeTag below range of valid modes",
			opts:         DecOptions{TimeTag: -1},
			wantErrorMsg: "cbor: invalid TimeTag -1",
		},
		{
			name:         "TimeTag above range of valid modes",
			opts:         DecOptions{TimeTag: 101},
			wantErrorMsg: "cbor: invalid TimeTag 101",
		},
		{
			name:         "MaxNestedLevels < 4",
			opts:         DecOptions{MaxNestedLevels: 1},
			wantErrorMsg: "cbor: invalid MaxNestedLevels 1 (range is [4, 65535])",
		},
		{
			name:         "MaxNestedLevels > 65535",
			opts:         DecOptions{MaxNestedLevels: 65536},
			wantErrorMsg: "cbor: invalid MaxNestedLevels 65536 (range is [4, 65535])",
		},
		{
			name:         "MaxArrayElements < 16",
			opts:         DecOptions{MaxArrayElements: 1},
			wantErrorMsg: "cbor: invalid MaxArrayElements 1 (range is [16, 2147483647])",
		},
		{
			name:         "MaxArrayElements > 2147483647",
			opts:         DecOptions{MaxArrayElements: 2147483648},
			wantErrorMsg: "cbor: invalid MaxArrayElements 2147483648 (range is [16, 2147483647])",
		},
		{
			name:         "MaxMapPairs < 16",
			opts:         DecOptions{MaxMapPairs: 1},
			wantErrorMsg: "cbor: invalid MaxMapPairs 1 (range is [16, 2147483647])",
		},
		{
			name:         "MaxMapPairs > 2147483647",
			opts:         DecOptions{MaxMapPairs: 2147483648},
			wantErrorMsg: "cbor: invalid MaxMapPairs 2147483648 (range is [16, 2147483647])",
		},
		{
			name:         "IndefLength below range of valid modes",
			opts:         DecOptions{IndefLength: -1},
			wantErrorMsg: "cbor: invalid IndefLength -1",
		},
		{
			name:         "IndefLength above range of valid modes",
			opts:         DecOptions{IndefLength: 101},
			wantErrorMsg: "cbor: invalid IndefLength 101",
		},
		{
			name:         "TagsMd below range of valid modes",
			opts:         DecOptions{TagsMd: -1},
			wantErrorMsg: "cbor: invalid TagsMd -1",
		},
		{
			name:         "TagsMd above range of valid modes",
			opts:         DecOptions{TagsMd: 101},
			wantErrorMsg: "cbor: invalid TagsMd 101",
		},
		{
			name:         "IntDec below range of valid modes",
			opts:         DecOptions{IntDec: -1},
			wantErrorMsg: "cbor: invalid IntDec -1",
		},
		{
			name:         "IntDec above range of valid modes",
			opts:         DecOptions{IntDec: 101},
			wantErrorMsg: "cbor: invalid IntDec 101",
		},
		{
			name:         "MapKeyByteString below range of valid modes",
			opts:         DecOptions{MapKeyByteString: -1},
			wantErrorMsg: "cbor: invalid MapKeyByteString -1",
		},
		{
			name:         "MapKeyByteString above range of valid modes",
			opts:         DecOptions{MapKeyByteString: 101},
			wantErrorMsg: "cbor: invalid MapKeyByteString 101",
		},
		{
			name:         "ExtraReturnErrors above range of valid values",
			opts:         DecOptions{ExtraReturnErrors: 3},
			wantErrorMsg: "cbor: invalid ExtraReturnErrors 3",
		},
		{
			name:         "DefaultMapType byte slice",
			opts:         DecOptions{DefaultMapType: reflect.TypeFor[[]byte]()},
			wantErrorMsg: "cbor: invalid DefaultMapType []uint8 (not of kind map)",
		},
		{
			name:         "DefaultMapType int slice",
			opts:         DecOptions{DefaultMapType: reflect.TypeFor[[]int]()},
			wantErrorMsg: "cbor: invalid DefaultMapType []int (not of kind map)",
		},
		{
			name:         "DefaultMapType string",
			opts:         DecOptions{DefaultMapType: reflect.TypeFor[string]()},
			wantErrorMsg: "cbor: invalid DefaultMapType string (not of kind map)",
		},
		{
			name:         "DefaultMapType unnamed struct type",
			opts:         DecOptions{DefaultMapType: reflect.TypeOf(struct{}{})},
			wantErrorMsg: "cbor: invalid DefaultMapType struct {} (not of kind map)",
		},
		{
			name:         "UTF8 below range of valid modes",
			opts:         DecOptions{UTF8: -1},
			wantErrorMsg: "cbor: invalid UTF8 -1",
		},
		{
			name:         "UTF8 above range of valid modes",
			opts:         DecOptions{UTF8: 101},
			wantErrorMsg: "cbor: invalid UTF8 101",
		},
		{
			name:         "FieldNameMatching below range of valid modes",
			opts:         DecOptions{FieldNameMatching: -1},
			wantErrorMsg: "cbor: invalid FieldNameMatching -1",
		},
		{
			name:         "FieldNameMatching above range of valid modes",
			opts:         DecOptions{FieldNameMatching: 101},
			wantErrorMsg: "cbor: invalid FieldNameMatching 101",
		},
		{
			name:         "BigIntDec below range of valid modes",
			opts:         DecOptions{BigIntDec: -1},
			wantErrorMsg: "cbor: invalid BigIntDec -1",
		},
		{
			name:         "BigIntDec above range of valid modes",
			opts:         DecOptions{BigIntDec: 101},
			wantErrorMsg: "cbor: invalid BigIntDec 101",
		},
		{
			name:         "DefaultByteStringType neither slice nor string",
			opts:         DecOptions{DefaultByteStringType: reflect.TypeFor[int]()},
			wantErrorMsg: "cbor: invalid DefaultByteStringType int (not of kind string or []uint8)",
		},
		{
			name:         "DefaultByteStringType slice of non-byte",
			opts:         DecOptions{DefaultByteStringType: reflect.TypeFor[[]int]()},
			wantErrorMsg: "cbor: invalid DefaultByteStringType []int (not of kind string or []uint8)",
		},
		{
			name:         "DefaultByteStringType pointer to byte array",
			opts:         DecOptions{DefaultByteStringType: reflect.TypeFor[*[42]byte]()},
			wantErrorMsg: "cbor: invalid DefaultByteStringType *[42]uint8 (not of kind string or []uint8)",
		},
		{
			name:         "DefaultByteStringType byte array",
			opts:         DecOptions{DefaultByteStringType: reflect.TypeFor[[42]byte]()},
			wantErrorMsg: "cbor: invalid DefaultByteStringType [42]uint8 (not of kind string or []uint8)",
		},
		{
			name:         "DefaultByteStringType slice with named uint8",
			opts:         DecOptions{DefaultByteStringType: reflect.TypeFor[[]myUint8]()},
			wantErrorMsg: "cbor: invalid DefaultByteStringType []cbor.myUint8 (not of kind string or []uint8)",
		},
		{
			name:         "DefaultByteStringType named slice with named uint8",
			opts:         DecOptions{DefaultByteStringType: reflect.TypeFor[myUint8Slice]()},
			wantErrorMsg: "cbor: invalid DefaultByteStringType cbor.myUint8Slice (not of kind string or []uint8)",
		},
		{
			name:         "ByteStringToString below range of valid modes",
			opts:         DecOptions{ByteStringToString: -1},
			wantErrorMsg: "cbor: invalid ByteStringToString -1",
		},
		{
			name:         "ByteStringToString above range of valid modes",
			opts:         DecOptions{ByteStringToString: 101},
			wantErrorMsg: "cbor: invalid ByteStringToString 101",
		},
		{
			name:         "FieldNameByteString below range of valid modes",
			opts:         DecOptions{FieldNameByteString: -1},
			wantErrorMsg: "cbor: invalid FieldNameByteString -1",
		},
		{
			name:         "FieldNameByteString above range of valid modes",
			opts:         DecOptions{FieldNameByteString: 101},
			wantErrorMsg: "cbor: invalid FieldNameByteString 101",
		},
		{
			name:         "UnrecognizedTagToAny below range of valid modes",
			opts:         DecOptions{UnrecognizedTagToAny: -1},
			wantErrorMsg: "cbor: invalid UnrecognizedTagToAny -1",
		},
		{
			name:         "UnrecognizedTagToAny above range of valid modes",
			opts:         DecOptions{UnrecognizedTagToAny: 101},
			wantErrorMsg: "cbor: invalid UnrecognizedTagToAny 101",
		},
		{
			name:         "TimeTagToAny below range of valid modes",
			opts:         DecOptions{TimeTagToAny: -1},
			wantErrorMsg: "cbor: invalid TimeTagToAny -1",
		},
		{
			name:         "TimeTagToAny above range of valid modes",
			opts:         DecOptions{TimeTagToAny: 4},
			wantErrorMsg: "cbor: invalid TimeTagToAny 4",
		},
		{
			name:         "NaN below range of valid modes",
			opts:         DecOptions{NaN: -1},
			wantErrorMsg: "cbor: invalid NaN -1",
		},
		{
			name:         "NaN above range of valid modes",
			opts:         DecOptions{NaN: 101},
			wantErrorMsg: "cbor: invalid NaN 101",
		},
		{
			name:         "Inf below range of valid modes",
			opts:         DecOptions{Inf: -1},
			wantErrorMsg: "cbor: invalid Inf -1",
		},
		{
			name:         "Inf above range of valid modes",
			opts:         DecOptions{Inf: 101},
			wantErrorMsg: "cbor: invalid Inf 101",
		},
		{
			name:         "ByteStringToTime below range of valid modes",
			opts:         DecOptions{ByteStringToTime: -1},
			wantErrorMsg: "cbor: invalid ByteStringToTime -1",
		},
		{
			name:         "ByteStringToTime above range of valid modes",
			opts:         DecOptions{ByteStringToTime: 4},
			wantErrorMsg: "cbor: invalid ByteStringToTime 4",
		},
		{
			name:         "ByteStringExpectedFormat below range of valid modes",
			opts:         DecOptions{ByteStringExpectedFormat: -1},
			wantErrorMsg: "cbor: invalid ByteStringExpectedFormat -1",
		},
		{
			name:         "ByteStringExpectedFormat above range of valid modes",
			opts:         DecOptions{ByteStringExpectedFormat: 101},
			wantErrorMsg: "cbor: invalid ByteStringExpectedFormat 101",
		},
		{
			name:         "BignumTag below range of valid modes",
			opts:         DecOptions{BignumTag: -1},
			wantErrorMsg: "cbor: invalid BignumTag -1",
		},
		{
			name:         "BignumTag above range of valid modes",
			opts:         DecOptions{BignumTag: 101},
			wantErrorMsg: "cbor: invalid BignumTag 101",
		},
		{
			name:         "BinaryUnmarshaler below range of valid modes",
			opts:         DecOptions{BinaryUnmarshaler: -1},
			wantErrorMsg: "cbor: invalid BinaryUnmarshaler -1",
		},
		{
			name:         "BinaryUnmarshaler above range of valid modes",
			opts:         DecOptions{BinaryUnmarshaler: 101},
			wantErrorMsg: "cbor: invalid BinaryUnmarshaler 101",
		},
		{
			name:         "TextUnmarshaler below range of valid modes",
			opts:         DecOptions{TextUnmarshaler: -1},
			wantErrorMsg: "cbor: invalid TextUnmarshaler -1",
		},
		{
			name:         "TextUnmarshaler above range of valid modes",
			opts:         DecOptions{TextUnmarshaler: 101},
			wantErrorMsg: "cbor: invalid TextUnmarshaler 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			checkError[error](t, "DecMode", err, tc.wantErrorMsg)
		})
	}
}

func TestDecModeDefaultMaxNestedLevel(t *testing.T) {
	dm := mustDecMode(t, DecOptions{})
	maxNestedLevels := dm.DecOptions().MaxNestedLevels
	if maxNestedLevels != 32 {
		t.Errorf("DecOptions().MaxNestedLevels = %d, want %v", maxNestedLevels, 32)
	}
}

func TestDecModeDefaultMaxArrayElements(t *testing.T) {
	dm := mustDecMode(t, DecOptions{})
	maxArrayElements := dm.DecOptions().MaxArrayElements
	if maxArrayElements != defaultMaxArrayElements {
		t.Errorf("DecOptions().MaxArrayElementsr = %d, want %v", maxArrayElements, defaultMaxArrayElements)
	}
}

func TestDecModeDefaultMaxMapPairs(t *testing.T) {
	dm := mustDecMode(t, DecOptions{})
	maxMapPairs := dm.DecOptions().MaxMapPairs
	if maxMapPairs != defaultMaxMapPairs {
		t.Errorf("DecOptions().MaxMapPairs = %d, want %v", maxMapPairs, defaultMaxMapPairs)
	}
}

func TestDecOptionsConflictWithRegisteredTags(t *testing.T) {
	type empty struct{}

	for _, tc := range []struct {
		name         string
		opts         DecOptions
		tagNum       uint64
		wantErrorMsg string
	}{
		{
			name:         "base64url encoding tag ignored by default",
			opts:         DecOptions{},
			tagNum:       21,
			wantErrorMsg: "",
		},
		{
			name:         "base64url encoding tag conflicts in ByteStringToStringAllowedWithExpectedLaterEncoding mode",
			opts:         DecOptions{ByteStringToString: ByteStringToStringAllowedWithExpectedLaterEncoding},
			tagNum:       21,
			wantErrorMsg: "cbor: DecMode with ByteStringToString set to ByteStringToStringAllowedWithExpectedLaterEncoding or non-default ByteStringExpectedFormat treats tag 21 as built-in and conflicts with the provided TagSet's registration of cbor.empty",
		},
		{
			name:         "base64url encoding tag conflicts with non-default ByteSliceExpectedEncoding option",
			opts:         DecOptions{ByteStringExpectedFormat: ByteStringExpectedBase16},
			tagNum:       21,
			wantErrorMsg: "cbor: DecMode with ByteStringToString set to ByteStringToStringAllowedWithExpectedLaterEncoding or non-default ByteStringExpectedFormat treats tag 21 as built-in and conflicts with the provided TagSet's registration of cbor.empty",
		},
		{
			name:         "base64 encoding tag ignored by default",
			opts:         DecOptions{},
			tagNum:       22,
			wantErrorMsg: "",
		},
		{
			name:         "base64 encoding tag conflicts in ByteStringToStringAllowedWithExpectedLaterEncoding mode",
			opts:         DecOptions{ByteStringToString: ByteStringToStringAllowedWithExpectedLaterEncoding},
			tagNum:       22,
			wantErrorMsg: "cbor: DecMode with ByteStringToString set to ByteStringToStringAllowedWithExpectedLaterEncoding or non-default ByteStringExpectedFormat treats tag 22 as built-in and conflicts with the provided TagSet's registration of cbor.empty",
		},
		{
			name:         "base64 encoding tag conflicts with non-default ByteSliceExpectedEncoding option",
			opts:         DecOptions{ByteStringExpectedFormat: ByteStringExpectedBase16},
			tagNum:       22,
			wantErrorMsg: "cbor: DecMode with ByteStringToString set to ByteStringToStringAllowedWithExpectedLaterEncoding or non-default ByteStringExpectedFormat treats tag 22 as built-in and conflicts with the provided TagSet's registration of cbor.empty",
		},
		{
			name:         "base16 encoding tag ignored by default",
			opts:         DecOptions{},
			tagNum:       23,
			wantErrorMsg: "",
		},
		{
			name:         "base16 encoding tag conflicts in ByteStringToStringAllowedWithExpectedLaterEncoding mode",
			opts:         DecOptions{ByteStringToString: ByteStringToStringAllowedWithExpectedLaterEncoding},
			tagNum:       23,
			wantErrorMsg: "cbor: DecMode with ByteStringToString set to ByteStringToStringAllowedWithExpectedLaterEncoding or non-default ByteStringExpectedFormat treats tag 23 as built-in and conflicts with the provided TagSet's registration of cbor.empty",
		},
		{
			name:         "base16 encoding tag conflicts with non-default ByteSliceExpectedEncoding option",
			opts:         DecOptions{ByteStringExpectedFormat: ByteStringExpectedBase16},
			tagNum:       23,
			wantErrorMsg: "cbor: DecMode with ByteStringToString set to ByteStringToStringAllowedWithExpectedLaterEncoding or non-default ByteStringExpectedFormat treats tag 23 as built-in and conflicts with the provided TagSet's registration of cbor.empty",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tags := NewTagSet()
			mustAddTag(t, tags, TagOptions{DecTag: DecTagOptional}, reflect.TypeFor[empty](), tc.tagNum)

			if _, err := tc.opts.DecModeWithTags(tags); err == nil {
				if tc.wantErrorMsg != "" {
					t.Errorf("got nil error from DecModeWithTags, want %q", tc.wantErrorMsg)
				}
			} else if got := err.Error(); got != tc.wantErrorMsg {
				if tc.wantErrorMsg != "" {
					t.Errorf("unexpected error from DecModeWithTags, got %q want %q", got, tc.wantErrorMsg)
				} else {
					t.Errorf("want nil error from DecModeWithTags, got %q", got)
				}
			}

			if _, err := tc.opts.DecModeWithSharedTags(tags); err == nil {
				if tc.wantErrorMsg != "" {
					t.Errorf("got nil error from DecModeWithSharedTags, want %q", tc.wantErrorMsg)
				}
			} else if got := err.Error(); got != tc.wantErrorMsg {
				if tc.wantErrorMsg != "" {
					t.Errorf("unexpected error from DecModeWithSharedTags, got %q want %q", got, tc.wantErrorMsg)
				} else {
					t.Errorf("want nil error from DecModeWithSharedTags, got %q", got)
				}
			}
		})
	}
}

func TestDecOptions(t *testing.T) {
	simpleValues, err := NewSimpleValueRegistryFromDefaults(WithRejectedSimpleValue(255))
	if err != nil {
		t.Fatal(err)
	}

	opts1 := DecOptions{
		DupMapKey:                 DupMapKeyEnforcedAPF,
		TimeTag:                   DecTagRequired,
		MaxNestedLevels:           100,
		MaxArrayElements:          102,
		MaxMapPairs:               101,
		IndefLength:               IndefLengthForbidden,
		TagsMd:                    TagsForbidden,
		IntDec:                    IntDecConvertSigned,
		MapKeyByteString:          MapKeyByteStringForbidden,
		ExtraReturnErrors:         ExtraDecErrorUnknownField,
		DefaultMapType:            reflect.TypeFor[map[string]any](),
		UTF8:                      UTF8DecodeInvalid,
		FieldNameMatching:         FieldNameMatchingCaseSensitive,
		BigIntDec:                 BigIntDecodePointer,
		DefaultByteStringType:     reflect.TypeFor[string](),
		ByteStringToString:        ByteStringToStringAllowed,
		FieldNameByteString:       FieldNameByteStringAllowed,
		UnrecognizedTagToAny:      UnrecognizedTagContentToAny,
		TimeTagToAny:              TimeTagToRFC3339,
		SimpleValues:              simpleValues,
		NaN:                       NaNDecodeForbidden,
		Inf:                       InfDecodeForbidden,
		ByteStringToTime:          ByteStringToTimeAllowed,
		ByteStringExpectedFormat:  ByteStringExpectedBase64URL,
		BignumTag:                 BignumTagForbidden,
		BinaryUnmarshaler:         BinaryUnmarshalerNone,
		TextUnmarshaler:           TextUnmarshalerTextString,
		JSONUnmarshalerTranscoder: stubTranscoder{},
	}
	ov := reflect.ValueOf(opts1)
	for i := range ov.NumField() {
		fv := ov.Field(i)
		if fv.IsZero() {
			t.Errorf("options field %q is unset or set to the zero value for its type", ov.Type().Field(i).Name)
		}
	}
	dm := mustDecMode(t, opts1)
	opts2 := dm.DecOptions()
	if !reflect.DeepEqual(opts1, opts2) {
		t.Errorf("DecOptions->DecMode->DecOptions returned different values: %#v, %#v", opts1, opts2)
	}
}
