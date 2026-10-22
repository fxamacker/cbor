// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"reflect"
	"testing"
)

func TestDecModeInvalidDuplicateMapKey(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{DupMapKey: -1},
			wantErrorMsg: "cbor: invalid DupMapKey -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{DupMapKey: 101},
			wantErrorMsg: "cbor: invalid DupMapKey 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidTimeTag(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{TimeTag: -1},
			wantErrorMsg: "cbor: invalid TimeTag -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{TimeTag: 101},
			wantErrorMsg: "cbor: invalid TimeTag 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeDefaultMaxNestedLevel(t *testing.T) {
	dm, err := DecOptions{}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned error %v", err)
	} else {
		maxNestedLevels := dm.DecOptions().MaxNestedLevels
		if maxNestedLevels != 32 {
			t.Errorf("DecOptions().MaxNestedLevels = %d, want %v", maxNestedLevels, 32)
		}
	}
}

func TestDecModeInvalidMaxNestedLevel(t *testing.T) {
	testCases := []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeDefaultMaxArrayElements(t *testing.T) {
	dm, err := DecOptions{}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned error %v", err)
	} else {
		maxArrayElements := dm.DecOptions().MaxArrayElements
		if maxArrayElements != defaultMaxArrayElements {
			t.Errorf("DecOptions().MaxArrayElementsr = %d, want %v", maxArrayElements, defaultMaxArrayElements)
		}
	}
}

func TestDecModeInvalidMaxArrayElements(t *testing.T) {
	testCases := []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeDefaultMaxMapPairs(t *testing.T) {
	dm, err := DecOptions{}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned error %v", err)
	} else {
		maxMapPairs := dm.DecOptions().MaxMapPairs
		if maxMapPairs != defaultMaxMapPairs {
			t.Errorf("DecOptions().MaxMapPairs = %d, want %v", maxMapPairs, defaultMaxMapPairs)
		}
	}
}

func TestDecModeInvalidMaxMapPairs(t *testing.T) {
	testCases := []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidIndefiniteLengthMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{IndefLength: -1},
			wantErrorMsg: "cbor: invalid IndefLength -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{IndefLength: 101},
			wantErrorMsg: "cbor: invalid IndefLength 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidTagsMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{TagsMd: -1},
			wantErrorMsg: "cbor: invalid TagsMd -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{TagsMd: 101},
			wantErrorMsg: "cbor: invalid TagsMd 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidIntDec(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{IntDec: -1},
			wantErrorMsg: "cbor: invalid IntDec -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{IntDec: 101},
			wantErrorMsg: "cbor: invalid IntDec 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidMapKeyByteString(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{MapKeyByteString: -1},
			wantErrorMsg: "cbor: invalid MapKeyByteString -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{MapKeyByteString: 101},
			wantErrorMsg: "cbor: invalid MapKeyByteString 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidExtraError(t *testing.T) {
	wantErrorMsg := "cbor: invalid ExtraReturnErrors 3"
	_, err := DecOptions{ExtraReturnErrors: 3}.DecMode()
	if err == nil {
		t.Errorf("DecMode() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("DecMode() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestDecModeInvalidDefaultMapType(t *testing.T) {
	testCases := []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "byte slice",
			opts:         DecOptions{DefaultMapType: reflect.TypeFor[[]byte]()},
			wantErrorMsg: "cbor: invalid DefaultMapType []uint8",
		},
		{
			name:         "int slice",
			opts:         DecOptions{DefaultMapType: reflect.TypeFor[[]int]()},
			wantErrorMsg: "cbor: invalid DefaultMapType []int",
		},
		{
			name:         "string",
			opts:         DecOptions{DefaultMapType: reflect.TypeFor[string]()},
			wantErrorMsg: "cbor: invalid DefaultMapType string",
		},
		{
			name:         "unnamed struct type",
			opts:         DecOptions{DefaultMapType: reflect.TypeOf(struct{}{})},
			wantErrorMsg: "cbor: invalid DefaultMapType struct {}",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestInvalidUTF8Mode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{UTF8: -1},
			wantErrorMsg: "cbor: invalid UTF8 -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{UTF8: 101},
			wantErrorMsg: "cbor: invalid UTF8 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidFieldNameMatchingMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{FieldNameMatching: -1},
			wantErrorMsg: "cbor: invalid FieldNameMatching -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{FieldNameMatching: 101},
			wantErrorMsg: "cbor: invalid FieldNameMatching 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestInvalidBigIntDecMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{BigIntDec: -1},
			wantErrorMsg: "cbor: invalid BigIntDec -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{BigIntDec: 101},
			wantErrorMsg: "cbor: invalid BigIntDec 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidDefaultByteStringType(t *testing.T) {
	type myUint8 uint8
	type myUint8Slice []myUint8

	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "neither slice nor string",
			opts:         DecOptions{DefaultByteStringType: reflect.TypeFor[int]()},
			wantErrorMsg: "cbor: invalid DefaultByteStringType: int is not of kind string or []uint8",
		},
		{
			name:         "slice of non-byte",
			opts:         DecOptions{DefaultByteStringType: reflect.TypeFor[[]int]()},
			wantErrorMsg: "cbor: invalid DefaultByteStringType: []int is not of kind string or []uint8",
		},
		{
			name:         "pointer to byte array",
			opts:         DecOptions{DefaultByteStringType: reflect.TypeFor[*[42]byte]()},
			wantErrorMsg: "cbor: invalid DefaultByteStringType: *[42]uint8 is not of kind string or []uint8",
		},
		{
			name:         "byte array",
			opts:         DecOptions{DefaultByteStringType: reflect.TypeFor[[42]byte]()},
			wantErrorMsg: "cbor: invalid DefaultByteStringType: [42]uint8 is not of kind string or []uint8",
		},
		{
			name:         "slice with named uint8",
			opts:         DecOptions{DefaultByteStringType: reflect.TypeFor[[]myUint8]()},
			wantErrorMsg: "cbor: invalid DefaultByteStringType: []cbor.myUint8 is not of kind string or []uint8",
		},
		{
			name:         "named slice with named uint8",
			opts:         DecOptions{DefaultByteStringType: reflect.TypeFor[myUint8Slice]()},
			wantErrorMsg: "cbor: invalid DefaultByteStringType: cbor.myUint8Slice is not of kind string or []uint8",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidByteStringToStringMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{ByteStringToString: -1},
			wantErrorMsg: "cbor: invalid ByteStringToString -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{ByteStringToString: 101},
			wantErrorMsg: "cbor: invalid ByteStringToString 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidFieldNameByteStringMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{FieldNameByteString: -1},
			wantErrorMsg: "cbor: invalid FieldNameByteString -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{FieldNameByteString: 101},
			wantErrorMsg: "cbor: invalid FieldNameByteString 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidReturnTypeForEmptyInterface(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{UnrecognizedTagToAny: -1},
			wantErrorMsg: "cbor: invalid UnrecognizedTagToAnyMode -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{UnrecognizedTagToAny: 101},
			wantErrorMsg: "cbor: invalid UnrecognizedTagToAnyMode 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidTimeTagToAnyMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{TimeTagToAny: -1},
			wantErrorMsg: "cbor: invalid TimeTagToAny -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{TimeTagToAny: 4},
			wantErrorMsg: "cbor: invalid TimeTagToAny 4",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("Expected non nil error from DecMode()")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Expected error: %q, want: %q \n", tc.wantErrorMsg, err.Error())
			}
		})
	}
}

func TestDecModeInvalidNaNDec(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{NaN: -1},
			wantErrorMsg: "cbor: invalid NaNDec -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{NaN: 101},
			wantErrorMsg: "cbor: invalid NaNDec 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidInfDec(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{Inf: -1},
			wantErrorMsg: "cbor: invalid InfDec -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{Inf: 101},
			wantErrorMsg: "cbor: invalid InfDec 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidByteStringToTimeMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{ByteStringToTime: -1},
			wantErrorMsg: "cbor: invalid ByteStringToTime -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{ByteStringToTime: 4},
			wantErrorMsg: "cbor: invalid ByteStringToTime 4",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("Expected non nil error from DecMode()")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Expected error: %q, want: %q \n", tc.wantErrorMsg, err.Error())
			}
		})
	}
}

func TestInvalidByteSliceExpectedEncodingMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{ByteStringExpectedFormat: -1},
			wantErrorMsg: "cbor: invalid ByteStringExpectedFormat -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{ByteStringExpectedFormat: 101},
			wantErrorMsg: "cbor: invalid ByteStringExpectedFormat 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidBignumTag(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{BignumTag: -1},
			wantErrorMsg: "cbor: invalid BignumTag -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{BignumTag: 101},
			wantErrorMsg: "cbor: invalid BignumTag 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("Expected non nil error from DecMode()")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("Expected error: %q, want: %q \n", tc.wantErrorMsg, err.Error())
			}
		})
	}
}

func TestDecModeInvalidBinaryUnmarshaler(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{BinaryUnmarshaler: -1},
			wantErrorMsg: "cbor: invalid BinaryUnmarshaler -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{BinaryUnmarshaler: 101},
			wantErrorMsg: "cbor: invalid BinaryUnmarshaler 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecModeInvalidTextUnmarshaler(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         DecOptions{TextUnmarshaler: -1},
			wantErrorMsg: "cbor: invalid TextUnmarshaler -1",
		},
		{
			name:         "above range of valid modes",
			opts:         DecOptions{TextUnmarshaler: 101},
			wantErrorMsg: "cbor: invalid TextUnmarshaler 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.DecMode()
			if err == nil {
				t.Errorf("DecMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("DecMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecOptionsConflictWithRegisteredTags(t *testing.T) {
	type empty struct{}

	for _, tc := range []struct {
		name         string
		opts         DecOptions
		tags         func(TagSet) error
		wantErrorMsg string
	}{
		{
			name: "base64url encoding tag ignored by default",
			opts: DecOptions{},
			tags: func(tags TagSet) error {
				return tags.Add(TagOptions{DecTag: DecTagOptional}, reflect.TypeFor[empty](), 21)
			},
			wantErrorMsg: "",
		},
		{
			name: "base64url encoding tag conflicts in ByteStringToStringAllowedWithExpectedLaterEncoding mode",
			opts: DecOptions{ByteStringToString: ByteStringToStringAllowedWithExpectedLaterEncoding},
			tags: func(tags TagSet) error {
				return tags.Add(TagOptions{DecTag: DecTagOptional}, reflect.TypeFor[empty](), 21)
			},
			wantErrorMsg: "cbor: DecMode with non-default StringExpectedEncoding or ByteSliceExpectedEncoding treats tag 21 as built-in and conflicts with the provided TagSet's registration of cbor.empty",
		},
		{
			name: "base64url encoding tag conflicts with non-default ByteSliceExpectedEncoding option",
			opts: DecOptions{ByteStringExpectedFormat: ByteStringExpectedBase16},
			tags: func(tags TagSet) error {
				return tags.Add(TagOptions{DecTag: DecTagOptional}, reflect.TypeFor[empty](), 21)
			},
			wantErrorMsg: "cbor: DecMode with non-default StringExpectedEncoding or ByteSliceExpectedEncoding treats tag 21 as built-in and conflicts with the provided TagSet's registration of cbor.empty",
		},
		{
			name: "base64 encoding tag ignored by default",
			opts: DecOptions{},
			tags: func(tags TagSet) error {
				return tags.Add(TagOptions{DecTag: DecTagOptional}, reflect.TypeFor[empty](), 22)
			},
			wantErrorMsg: "",
		},
		{
			name: "base64 encoding tag conflicts in ByteStringToStringAllowedWithExpectedLaterEncoding mode",
			opts: DecOptions{ByteStringToString: ByteStringToStringAllowedWithExpectedLaterEncoding},
			tags: func(tags TagSet) error {
				return tags.Add(TagOptions{DecTag: DecTagOptional}, reflect.TypeFor[empty](), 22)
			},
			wantErrorMsg: "cbor: DecMode with non-default StringExpectedEncoding or ByteSliceExpectedEncoding treats tag 22 as built-in and conflicts with the provided TagSet's registration of cbor.empty",
		},
		{
			name: "base64 encoding tag conflicts with non-default ByteSliceExpectedEncoding option",
			opts: DecOptions{ByteStringExpectedFormat: ByteStringExpectedBase16},
			tags: func(tags TagSet) error {
				return tags.Add(TagOptions{DecTag: DecTagOptional}, reflect.TypeFor[empty](), 22)
			},
			wantErrorMsg: "cbor: DecMode with non-default StringExpectedEncoding or ByteSliceExpectedEncoding treats tag 22 as built-in and conflicts with the provided TagSet's registration of cbor.empty",
		},
		{
			name: "base16 encoding tag ignored by default",
			opts: DecOptions{},
			tags: func(tags TagSet) error {
				return tags.Add(TagOptions{DecTag: DecTagOptional}, reflect.TypeFor[empty](), 23)
			},
			wantErrorMsg: "",
		},
		{
			name: "base16 encoding tag conflicts in ByteStringToStringAllowedWithExpectedLaterEncoding mode",
			opts: DecOptions{ByteStringToString: ByteStringToStringAllowedWithExpectedLaterEncoding},
			tags: func(tags TagSet) error {
				return tags.Add(TagOptions{DecTag: DecTagOptional}, reflect.TypeFor[empty](), 23)
			},
			wantErrorMsg: "cbor: DecMode with non-default StringExpectedEncoding or ByteSliceExpectedEncoding treats tag 23 as built-in and conflicts with the provided TagSet's registration of cbor.empty",
		},
		{
			name: "base16 encoding tag conflicts with non-default ByteSliceExpectedEncoding option",
			opts: DecOptions{ByteStringExpectedFormat: ByteStringExpectedBase16},
			tags: func(tags TagSet) error {
				return tags.Add(TagOptions{DecTag: DecTagOptional}, reflect.TypeFor[empty](), 23)
			},
			wantErrorMsg: "cbor: DecMode with non-default StringExpectedEncoding or ByteSliceExpectedEncoding treats tag 23 as built-in and conflicts with the provided TagSet's registration of cbor.empty",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tags := NewTagSet()
			if err := tc.tags(tags); err != nil {
				t.Fatal(err)
			}

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
	dm, err := opts1.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned an error %v", err)
	} else {
		opts2 := dm.DecOptions()
		if !reflect.DeepEqual(opts1, opts2) {
			t.Errorf("DecOptions->DecMode->DecOptions returned different values: %#v, %#v", opts1, opts2)
		}
	}
}
