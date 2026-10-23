// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestInvalidSort(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{Sort: -1},
			wantErrorMsg: "cbor: invalid Sort -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{Sort: 101},
			wantErrorMsg: "cbor: invalid Sort 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestInvalidShortestFloat(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{ShortestFloat: -1},
			wantErrorMsg: "cbor: invalid ShortestFloat -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{ShortestFloat: 101},
			wantErrorMsg: "cbor: invalid ShortestFloat 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestInvalidNaNConvert(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{NaNConvert: -1},
			wantErrorMsg: "cbor: invalid NaNConvert -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{NaNConvert: 101},
			wantErrorMsg: "cbor: invalid NaNConvert 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestInvalidInfConvert(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{InfConvert: -1},
			wantErrorMsg: "cbor: invalid InfConvert -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{InfConvert: 101},
			wantErrorMsg: "cbor: invalid InfConvert 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestEncModeInvalidBigIntConvertMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{BigIntConvert: -1},
			wantErrorMsg: "cbor: invalid BigIntConvert -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{BigIntConvert: 101},
			wantErrorMsg: "cbor: invalid BigIntConvert 101",
		},
		{
			name:         "above range of valid modes with TagsForbidden",
			opts:         EncOptions{BigIntConvert: 101, TagsMd: TagsForbidden},
			wantErrorMsg: "cbor: invalid BigIntConvert 101",
		},
		{
			name:         "BigIntConvertNone with TagsForbidden",
			opts:         EncOptions{BigIntConvert: BigIntConvertNone, TagsMd: TagsForbidden},
			wantErrorMsg: "cbor: cannot set TagsMd to TagsForbidden when BigIntConvert is BigIntConvertNone",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestInvalidTimeMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{Time: -1},
			wantErrorMsg: "cbor: invalid Time -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{Time: 101},
			wantErrorMsg: "cbor: invalid Time 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestEncModeInvalidTimeTag(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{TimeTag: -1},
			wantErrorMsg: "cbor: invalid TimeTag -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{TimeTag: 101},
			wantErrorMsg: "cbor: invalid TimeTag 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestEncModeInvalidIndefiniteLengthMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{IndefLength: -1},
			wantErrorMsg: "cbor: invalid IndefLength -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{IndefLength: 101},
			wantErrorMsg: "cbor: invalid IndefLength 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestInvalidNilContainers(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{NilContainers: -1},
			wantErrorMsg: "cbor: invalid NilContainers -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{NilContainers: 101},
			wantErrorMsg: "cbor: invalid NilContainers 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestEncModeInvalidTagsMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{TagsMd: -1},
			wantErrorMsg: "cbor: invalid TagsMd -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{TagsMd: 101},
			wantErrorMsg: "cbor: invalid TagsMd 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestInvalidOmitEmptyMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{OmitEmpty: -1},
			wantErrorMsg: "cbor: invalid OmitEmpty -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{OmitEmpty: 101},
			wantErrorMsg: "cbor: invalid OmitEmpty 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestEncModeStringType(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "invalid mode",
			opts:         EncOptions{String: -1},
			wantErrorMsg: "cbor: invalid String -1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestEncModeInvalidFieldNameMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{FieldName: -1},
			wantErrorMsg: "cbor: invalid FieldName -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{FieldName: 101},
			wantErrorMsg: "cbor: invalid FieldName 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestInvalidByteSliceExpectedFormat(t *testing.T) {
	tagsForbiddenErrorMsg := "cbor: cannot set TagsMd to TagsForbidden when ByteSliceLaterFormat is not ByteSliceLaterFormatNone"

	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{ByteSliceLaterFormat: -1},
			wantErrorMsg: "cbor: invalid ByteSliceLaterFormat -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{ByteSliceLaterFormat: 101},
			wantErrorMsg: "cbor: invalid ByteSliceLaterFormat 101",
		},
		{
			name:         "above range of valid modes with TagsForbidden",
			opts:         EncOptions{ByteSliceLaterFormat: 101, TagsMd: TagsForbidden},
			wantErrorMsg: "cbor: invalid ByteSliceLaterFormat 101",
		},
		{
			name:         "ByteSliceLaterFormatBase64URL with TagsForbidden",
			opts:         EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase64URL, TagsMd: TagsForbidden},
			wantErrorMsg: tagsForbiddenErrorMsg,
		},
		{
			name:         "ByteSliceLaterFormatBase64 with TagsForbidden",
			opts:         EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase64, TagsMd: TagsForbidden},
			wantErrorMsg: tagsForbiddenErrorMsg,
		},
		{
			name:         "ByteSliceLaterFormatBase16 with TagsForbidden",
			opts:         EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase16, TagsMd: TagsForbidden},
			wantErrorMsg: tagsForbiddenErrorMsg,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestInvalidByteArray(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{ByteArray: -1},
			wantErrorMsg: "cbor: invalid ByteArray -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{ByteArray: 101},
			wantErrorMsg: "cbor: invalid ByteArray 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestEncModeInvalidBinaryMarshalerMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{BinaryMarshaler: -1},
			wantErrorMsg: "cbor: invalid BinaryMarshaler -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{BinaryMarshaler: 101},
			wantErrorMsg: "cbor: invalid BinaryMarshaler 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestEncModeInvalidTextMarshalerMode(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "below range of valid modes",
			opts:         EncOptions{TextMarshaler: -1},
			wantErrorMsg: "cbor: invalid TextMarshaler -1",
		},
		{
			name:         "above range of valid modes",
			opts:         EncOptions{TextMarshaler: 101},
			wantErrorMsg: "cbor: invalid TextMarshaler 101",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			if err == nil {
				t.Errorf("EncMode() didn't return an error")
			} else if err.Error() != tc.wantErrorMsg {
				t.Errorf("EncMode() returned error %q, want %q", err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

// TestInvalidToIndefArrayStructTagMode verifies that EncOptions rejects
// out-of-range values for ToIndefArrayStructTagMode at mode construction time.
func TestInvalidToIndefArrayStructTagMode(t *testing.T) {
	for _, m := range []ToIndefArrayStructTagMode{-1, maxToIndefArrayStructTagMode, maxToIndefArrayStructTagMode + 1} {
		_, err := EncOptions{ToIndefArrayStructTag: m}.EncMode()
		if err == nil {
			t.Errorf("expected error for mode %d, got nil", m)
			continue
		}
		if !strings.Contains(err.Error(), "ToIndefArrayStructTag") {
			t.Errorf("error for mode %d should mention ToIndefArrayStructTag, got: %v", m, err)
		}
	}
}

// TestToIndefArrayStructTagRejectsIndefLengthForbidden verifies that
// EncOptions rejects the combination of IndefLengthForbidden and
// ToIndefArrayStructTagAllowed at mode construction time, since the two
// settings are contradictory.
func TestToIndefArrayStructTagRejectsIndefLengthForbidden(t *testing.T) {
	_, err := EncOptions{
		IndefLength:           IndefLengthForbidden,
		ToIndefArrayStructTag: ToIndefArrayStructTagAllowed,
	}.EncMode()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "ToIndefArrayStructTag") || !strings.Contains(err.Error(), "IndefLength") {
		t.Errorf("error should mention both options, got: %v", err)
	}
}

// TestToIndefArrayStructTagOptionsRoundTrip verifies that an EncOptions
// value containing the non-zero value of ToIndefArrayStructTag round-trips
// through EncOptions -> EncMode -> EncOptions with all fields preserved.
// TestEncOptions cannot cover this case because the non-zero values of
// IndefLength and ToIndefArrayStructTag are mutually exclusive.
func TestToIndefArrayStructTagOptionsRoundTrip(t *testing.T) {
	opts := EncOptions{ToIndefArrayStructTag: ToIndefArrayStructTagAllowed}
	em := mustEncMode(t, opts)
	got := em.EncOptions()
	if !reflect.DeepEqual(opts, got) {
		t.Errorf("EncOptions round-trip mismatch:\n got: %#v\nwant: %#v", got, opts)
	}
}

func TestEncOptionsTagsForbidden(t *testing.T) {
	// It's not valid to set both TagsMd and TimeTag to a non-zero value in the same EncOptions,
	// so this exercises the options-mode-options roundtrip for non-zero TagsMd.
	opts1 := EncOptions{
		TagsMd: TagsForbidden,
	}
	em := mustEncMode(t, opts1)
	opts2 := em.EncOptions()
	if !reflect.DeepEqual(opts1, opts2) {
		t.Errorf("EncOptions->EncMode->EncOptions returned different values: %#v, %#v", opts1, opts2)
	}
}

func TestCanonicalEncOptions(t *testing.T) { //nolint:dupl
	wantSortMode := SortCanonical
	wantShortestFloat := ShortestFloat16
	wantNaNConvert := NaNConvert7e00
	wantInfConvert := InfConvertFloat16
	wantErrorMsg := "cbor: indefinite-length array isn't allowed"
	em := mustEncMode(t, CanonicalEncOptions())
	opts := em.EncOptions()
	if opts.Sort != wantSortMode {
		t.Errorf("CanonicalEncOptions() returned EncOptions with Sort %d, want %d", opts.Sort, wantSortMode)
	}
	if opts.ShortestFloat != wantShortestFloat {
		t.Errorf("CanonicalEncOptions() returned EncOptions with ShortestFloat %d, want %d", opts.ShortestFloat, wantShortestFloat)
	}
	if opts.NaNConvert != wantNaNConvert {
		t.Errorf("CanonicalEncOptions() returned EncOptions with NaNConvert %d, want %d", opts.NaNConvert, wantNaNConvert)
	}
	if opts.InfConvert != wantInfConvert {
		t.Errorf("CanonicalEncOptions() returned EncOptions with InfConvert %d, want %d", opts.InfConvert, wantInfConvert)
	}
	enc := em.NewEncoder(io.Discard)
	if err := enc.StartIndefiniteArray(); err == nil {
		t.Errorf("StartIndefiniteArray() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("StartIndefiniteArray() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestCTAP2EncOptions(t *testing.T) { //nolint:dupl
	wantSortMode := SortCTAP2
	wantShortestFloat := ShortestFloatNone
	wantNaNConvert := NaNConvertNone
	wantInfConvert := InfConvertNone
	wantErrorMsg := "cbor: indefinite-length array isn't allowed"
	em := mustEncMode(t, CTAP2EncOptions())
	opts := em.EncOptions()
	if opts.Sort != wantSortMode {
		t.Errorf("CTAP2EncOptions() returned EncOptions with Sort %d, want %d", opts.Sort, wantSortMode)
	}
	if opts.ShortestFloat != wantShortestFloat {
		t.Errorf("CTAP2EncOptions() returned EncOptions with ShortestFloat %d, want %d", opts.ShortestFloat, wantShortestFloat)
	}
	if opts.NaNConvert != wantNaNConvert {
		t.Errorf("CTAP2EncOptions() returned EncOptions with NaNConvert %d, want %d", opts.NaNConvert, wantNaNConvert)
	}
	if opts.InfConvert != wantInfConvert {
		t.Errorf("CTAP2EncOptions() returned EncOptions with InfConvert %d, want %d", opts.InfConvert, wantInfConvert)
	}
	enc := em.NewEncoder(io.Discard)
	if err := enc.StartIndefiniteArray(); err == nil {
		t.Errorf("StartIndefiniteArray() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("StartIndefiniteArray() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestCoreDetEncOptions(t *testing.T) { //nolint:dupl
	wantSortMode := SortCoreDeterministic
	wantShortestFloat := ShortestFloat16
	wantNaNConvert := NaNConvert7e00
	wantInfConvert := InfConvertFloat16
	wantErrorMsg := "cbor: indefinite-length array isn't allowed"
	em := mustEncMode(t, CoreDetEncOptions())
	opts := em.EncOptions()
	if opts.Sort != wantSortMode {
		t.Errorf("CoreDetEncOptions() returned EncOptions with Sort %d, want %d", opts.Sort, wantSortMode)
	}
	if opts.ShortestFloat != wantShortestFloat {
		t.Errorf("CoreDetEncOptions() returned EncOptions with ShortestFloat %d, want %d", opts.ShortestFloat, wantShortestFloat)
	}
	if opts.NaNConvert != wantNaNConvert {
		t.Errorf("CoreDetEncOptions() returned EncOptions with NaNConvert %d, want %d", opts.NaNConvert, wantNaNConvert)
	}
	if opts.InfConvert != wantInfConvert {
		t.Errorf("CoreDetEncOptions() returned EncOptions with InfConvert %d, want %d", opts.InfConvert, wantInfConvert)
	}
	enc := em.NewEncoder(io.Discard)
	if err := enc.StartIndefiniteArray(); err == nil {
		t.Errorf("StartIndefiniteArray() didn't return an error")
	} else if err.Error() != wantErrorMsg {
		t.Errorf("StartIndefiniteArray() returned error %q, want %q", err.Error(), wantErrorMsg)
	}
}

func TestPreferredUnsortedEncOptions(t *testing.T) {
	wantSortMode := SortNone
	wantShortestFloat := ShortestFloat16
	wantNaNConvert := NaNConvert7e00
	wantInfConvert := InfConvertFloat16
	em := mustEncMode(t, PreferredUnsortedEncOptions())
	opts := em.EncOptions()
	if opts.Sort != wantSortMode {
		t.Errorf("PreferredUnsortedEncOptions() returned EncOptions with Sort %d, want %d", opts.Sort, wantSortMode)
	}
	if opts.ShortestFloat != wantShortestFloat {
		t.Errorf("PreferredUnsortedEncOptions() returned EncOptions with ShortestFloat %d, want %d", opts.ShortestFloat, wantShortestFloat)
	}
	if opts.NaNConvert != wantNaNConvert {
		t.Errorf("PreferredUnsortedEncOptions() returned EncOptions with NaNConvert %d, want %d", opts.NaNConvert, wantNaNConvert)
	}
	if opts.InfConvert != wantInfConvert {
		t.Errorf("PreferredUnsortedEncOptions() returned EncOptions with InfConvert %d, want %d", opts.InfConvert, wantInfConvert)
	}
	enc := em.NewEncoder(io.Discard)
	if err := enc.StartIndefiniteArray(); err != nil {
		t.Errorf("StartIndefiniteArray() returned error %v", err)
	}
}

type stubTranscoder struct{}

func (stubTranscoder) Transcode(io.Writer, io.Reader) error {
	return nil
}

func TestEncOptions(t *testing.T) {
	opts1 := EncOptions{
		Sort:                    SortBytewiseLexical,
		ShortestFloat:           ShortestFloat16,
		NaNConvert:              NaNConvertPreserveSignal,
		InfConvert:              InfConvertNone,
		BigIntConvert:           BigIntConvertNone,
		Time:                    TimeRFC3339Nano,
		TimeTag:                 EncTagRequired,
		IndefLength:             IndefLengthForbidden,
		NilContainers:           NilContainerAsEmpty,
		TagsMd:                  TagsAllowed,
		OmitEmpty:               OmitEmptyGoValue,
		String:                  StringToByteString,
		FieldName:               FieldNameToByteString,
		ByteSliceLaterFormat:    ByteSliceLaterFormatBase16,
		ByteArray:               ByteArrayToArray,
		BinaryMarshaler:         BinaryMarshalerNone,
		TextMarshaler:           TextMarshalerTextString,
		JSONMarshalerTranscoder: stubTranscoder{},
	}
	ov := reflect.ValueOf(opts1)
	for i := range ov.NumField() {
		fv := ov.Field(i)
		if fv.IsZero() {
			fn := ov.Type().Field(i).Name
			if fn == "TagsMd" {
				// Roundtripping non-zero values for TagsMd is tested separately
				// since the non-zero value (TagsForbidden) is incompatible with the
				// non-zero value for other options (e.g. TimeTag).
				continue
			}
			if fn == "ToIndefArrayStructTag" {
				// Roundtripping non-zero values for ToIndefArrayStructTag is tested
				// separately since the non-zero value (ToIndefArrayStructTagAllowed)
				// is incompatible with IndefLengthForbidden, which is the non-zero
				// value used by IndefLength above.
				continue
			}
			t.Errorf("options field %q is unset or set to the zero value for its type", fn)
		}
	}
	em := mustEncMode(t, opts1)
	opts2 := em.EncOptions()
	if !reflect.DeepEqual(opts1, opts2) {
		t.Errorf("EncOptions->EncMode->EncOptions returned different values: %#v, %#v", opts1, opts2)
	}
}
