// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"io"
	"reflect"
	"testing"
)

func TestEncModeInvalidOptions(t *testing.T) {
	tagsForbiddenErrorMsg := "cbor: cannot set TagsMd to TagsForbidden when ByteSliceLaterFormat is not ByteSliceLaterFormatNone"

	for _, tc := range []struct {
		name         string
		opts         EncOptions
		wantErrorMsg string
	}{
		{
			name:         "Sort below range of valid modes",
			opts:         EncOptions{Sort: -1},
			wantErrorMsg: "cbor: invalid Sort -1",
		},
		{
			name:         "Sort above range of valid modes",
			opts:         EncOptions{Sort: 101},
			wantErrorMsg: "cbor: invalid Sort 101",
		},
		{
			name:         "ShortestFloat below range of valid modes",
			opts:         EncOptions{ShortestFloat: -1},
			wantErrorMsg: "cbor: invalid ShortestFloat -1",
		},
		{
			name:         "ShortestFloat above range of valid modes",
			opts:         EncOptions{ShortestFloat: 101},
			wantErrorMsg: "cbor: invalid ShortestFloat 101",
		},
		{
			name:         "NaNConvert below range of valid modes",
			opts:         EncOptions{NaNConvert: -1},
			wantErrorMsg: "cbor: invalid NaNConvert -1",
		},
		{
			name:         "NaNConvert above range of valid modes",
			opts:         EncOptions{NaNConvert: 101},
			wantErrorMsg: "cbor: invalid NaNConvert 101",
		},
		{
			name:         "InfConvert below range of valid modes",
			opts:         EncOptions{InfConvert: -1},
			wantErrorMsg: "cbor: invalid InfConvert -1",
		},
		{
			name:         "InfConvert above range of valid modes",
			opts:         EncOptions{InfConvert: 101},
			wantErrorMsg: "cbor: invalid InfConvert 101",
		},
		{
			name:         "BigIntConvert below range of valid modes",
			opts:         EncOptions{BigIntConvert: -1},
			wantErrorMsg: "cbor: invalid BigIntConvert -1",
		},
		{
			name:         "BigIntConvert above range of valid modes",
			opts:         EncOptions{BigIntConvert: 101},
			wantErrorMsg: "cbor: invalid BigIntConvert 101",
		},
		{
			name:         "BigIntConvert above range of valid modes with TagsForbidden",
			opts:         EncOptions{BigIntConvert: 101, TagsMd: TagsForbidden},
			wantErrorMsg: "cbor: invalid BigIntConvert 101",
		},
		{
			name:         "BigIntConvert BigIntConvertNone with TagsForbidden",
			opts:         EncOptions{BigIntConvert: BigIntConvertNone, TagsMd: TagsForbidden},
			wantErrorMsg: "cbor: cannot set TagsMd to TagsForbidden when BigIntConvert is BigIntConvertNone",
		},
		{
			name:         "Time below range of valid modes",
			opts:         EncOptions{Time: -1},
			wantErrorMsg: "cbor: invalid Time -1",
		},
		{
			name:         "Time above range of valid modes",
			opts:         EncOptions{Time: 101},
			wantErrorMsg: "cbor: invalid Time 101",
		},
		{
			name:         "TimeTag below range of valid modes",
			opts:         EncOptions{TimeTag: -1},
			wantErrorMsg: "cbor: invalid TimeTag -1",
		},
		{
			name:         "TimeTag above range of valid modes",
			opts:         EncOptions{TimeTag: 101},
			wantErrorMsg: "cbor: invalid TimeTag 101",
		},
		{
			name:         "IndefLength below range of valid modes",
			opts:         EncOptions{IndefLength: -1},
			wantErrorMsg: "cbor: invalid IndefLength -1",
		},
		{
			name:         "IndefLength above range of valid modes",
			opts:         EncOptions{IndefLength: 101},
			wantErrorMsg: "cbor: invalid IndefLength 101",
		},
		{
			name:         "NilContainers below range of valid modes",
			opts:         EncOptions{NilContainers: -1},
			wantErrorMsg: "cbor: invalid NilContainers -1",
		},
		{
			name:         "NilContainers above range of valid modes",
			opts:         EncOptions{NilContainers: 101},
			wantErrorMsg: "cbor: invalid NilContainers 101",
		},
		{
			name:         "TagsMd below range of valid modes",
			opts:         EncOptions{TagsMd: -1},
			wantErrorMsg: "cbor: invalid TagsMd -1",
		},
		{
			name:         "TagsMd above range of valid modes",
			opts:         EncOptions{TagsMd: 101},
			wantErrorMsg: "cbor: invalid TagsMd 101",
		},
		{
			name:         "OmitEmpty below range of valid modes",
			opts:         EncOptions{OmitEmpty: -1},
			wantErrorMsg: "cbor: invalid OmitEmpty -1",
		},
		{
			name:         "OmitEmpty above range of valid modes",
			opts:         EncOptions{OmitEmpty: 101},
			wantErrorMsg: "cbor: invalid OmitEmpty 101",
		},
		{
			name:         "String invalid mode",
			opts:         EncOptions{String: -1},
			wantErrorMsg: "cbor: invalid String -1",
		},
		{
			name:         "FieldName below range of valid modes",
			opts:         EncOptions{FieldName: -1},
			wantErrorMsg: "cbor: invalid FieldName -1",
		},
		{
			name:         "FieldName above range of valid modes",
			opts:         EncOptions{FieldName: 101},
			wantErrorMsg: "cbor: invalid FieldName 101",
		},
		{
			name:         "ByteSliceLaterFormat below range of valid modes",
			opts:         EncOptions{ByteSliceLaterFormat: -1},
			wantErrorMsg: "cbor: invalid ByteSliceLaterFormat -1",
		},
		{
			name:         "ByteSliceLaterFormat above range of valid modes",
			opts:         EncOptions{ByteSliceLaterFormat: 101},
			wantErrorMsg: "cbor: invalid ByteSliceLaterFormat 101",
		},
		{
			name:         "ByteSliceLaterFormat above range of valid modes with TagsForbidden",
			opts:         EncOptions{ByteSliceLaterFormat: 101, TagsMd: TagsForbidden},
			wantErrorMsg: "cbor: invalid ByteSliceLaterFormat 101",
		},
		{
			name:         "ByteSliceLaterFormat ByteSliceLaterFormatBase64URL with TagsForbidden",
			opts:         EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase64URL, TagsMd: TagsForbidden},
			wantErrorMsg: tagsForbiddenErrorMsg,
		},
		{
			name:         "ByteSliceLaterFormat ByteSliceLaterFormatBase64 with TagsForbidden",
			opts:         EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase64, TagsMd: TagsForbidden},
			wantErrorMsg: tagsForbiddenErrorMsg,
		},
		{
			name:         "ByteSliceLaterFormat ByteSliceLaterFormatBase16 with TagsForbidden",
			opts:         EncOptions{ByteSliceLaterFormat: ByteSliceLaterFormatBase16, TagsMd: TagsForbidden},
			wantErrorMsg: tagsForbiddenErrorMsg,
		},
		{
			name:         "ByteArray below range of valid modes",
			opts:         EncOptions{ByteArray: -1},
			wantErrorMsg: "cbor: invalid ByteArray -1",
		},
		{
			name:         "ByteArray above range of valid modes",
			opts:         EncOptions{ByteArray: 101},
			wantErrorMsg: "cbor: invalid ByteArray 101",
		},
		{
			name:         "BinaryMarshaler below range of valid modes",
			opts:         EncOptions{BinaryMarshaler: -1},
			wantErrorMsg: "cbor: invalid BinaryMarshaler -1",
		},
		{
			name:         "BinaryMarshaler above range of valid modes",
			opts:         EncOptions{BinaryMarshaler: 101},
			wantErrorMsg: "cbor: invalid BinaryMarshaler 101",
		},
		{
			name:         "TextMarshaler below range of valid modes",
			opts:         EncOptions{TextMarshaler: -1},
			wantErrorMsg: "cbor: invalid TextMarshaler -1",
		},
		{
			name:         "TextMarshaler above range of valid modes",
			opts:         EncOptions{TextMarshaler: 101},
			wantErrorMsg: "cbor: invalid TextMarshaler 101",
		},
		{
			name:         "ToIndefArrayStructTag below range of valid modes",
			opts:         EncOptions{ToIndefArrayStructTag: -1},
			wantErrorMsg: "cbor: invalid ToIndefArrayStructTag -1",
		},
		{
			name:         "ToIndefArrayStructTag above range of valid modes",
			opts:         EncOptions{ToIndefArrayStructTag: maxToIndefArrayStructTagMode},
			wantErrorMsg: "cbor: invalid ToIndefArrayStructTag 2",
		},
		{
			name:         "ToIndefArrayStructTagAllowed with IndefLengthForbidden",
			opts:         EncOptions{IndefLength: IndefLengthForbidden, ToIndefArrayStructTag: ToIndefArrayStructTagAllowed},
			wantErrorMsg: "cbor: cannot set ToIndefArrayStructTag to ToIndefArrayStructTagAllowed when IndefLength is IndefLengthForbidden",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.opts.EncMode()
			checkError[error](t, "EncMode", err, tc.wantErrorMsg)
		})
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
		t.Errorf("em.EncOptions() = %#v, want %#v", got, opts)
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
		t.Errorf("em.EncOptions() = %#v, want %#v", opts2, opts1)
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
		t.Errorf("em.EncOptions().Sort = %d, want %d", opts.Sort, wantSortMode)
	}
	if opts.ShortestFloat != wantShortestFloat {
		t.Errorf("em.EncOptions().ShortestFloat = %d, want %d", opts.ShortestFloat, wantShortestFloat)
	}
	if opts.NaNConvert != wantNaNConvert {
		t.Errorf("em.EncOptions().NaNConvert = %d, want %d", opts.NaNConvert, wantNaNConvert)
	}
	if opts.InfConvert != wantInfConvert {
		t.Errorf("em.EncOptions().InfConvert = %d, want %d", opts.InfConvert, wantInfConvert)
	}
	enc := em.NewEncoder(io.Discard)
	err := enc.StartIndefiniteArray()
	checkError[error](t, "Encoder.StartIndefiniteArray", err, wantErrorMsg)
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
		t.Errorf("em.EncOptions().Sort = %d, want %d", opts.Sort, wantSortMode)
	}
	if opts.ShortestFloat != wantShortestFloat {
		t.Errorf("em.EncOptions().ShortestFloat = %d, want %d", opts.ShortestFloat, wantShortestFloat)
	}
	if opts.NaNConvert != wantNaNConvert {
		t.Errorf("em.EncOptions().NaNConvert = %d, want %d", opts.NaNConvert, wantNaNConvert)
	}
	if opts.InfConvert != wantInfConvert {
		t.Errorf("em.EncOptions().InfConvert = %d, want %d", opts.InfConvert, wantInfConvert)
	}
	enc := em.NewEncoder(io.Discard)
	err := enc.StartIndefiniteArray()
	checkError[error](t, "Encoder.StartIndefiniteArray", err, wantErrorMsg)
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
		t.Errorf("em.EncOptions().Sort = %d, want %d", opts.Sort, wantSortMode)
	}
	if opts.ShortestFloat != wantShortestFloat {
		t.Errorf("em.EncOptions().ShortestFloat = %d, want %d", opts.ShortestFloat, wantShortestFloat)
	}
	if opts.NaNConvert != wantNaNConvert {
		t.Errorf("em.EncOptions().NaNConvert = %d, want %d", opts.NaNConvert, wantNaNConvert)
	}
	if opts.InfConvert != wantInfConvert {
		t.Errorf("em.EncOptions().InfConvert = %d, want %d", opts.InfConvert, wantInfConvert)
	}
	enc := em.NewEncoder(io.Discard)
	err := enc.StartIndefiniteArray()
	checkError[error](t, "Encoder.StartIndefiniteArray", err, wantErrorMsg)
}

func TestPreferredUnsortedEncOptions(t *testing.T) {
	wantSortMode := SortNone
	wantShortestFloat := ShortestFloat16
	wantNaNConvert := NaNConvert7e00
	wantInfConvert := InfConvertFloat16
	em := mustEncMode(t, PreferredUnsortedEncOptions())
	opts := em.EncOptions()
	if opts.Sort != wantSortMode {
		t.Errorf("em.EncOptions().Sort = %d, want %d", opts.Sort, wantSortMode)
	}
	if opts.ShortestFloat != wantShortestFloat {
		t.Errorf("em.EncOptions().ShortestFloat = %d, want %d", opts.ShortestFloat, wantShortestFloat)
	}
	if opts.NaNConvert != wantNaNConvert {
		t.Errorf("em.EncOptions().NaNConvert = %d, want %d", opts.NaNConvert, wantNaNConvert)
	}
	if opts.InfConvert != wantInfConvert {
		t.Errorf("em.EncOptions().InfConvert = %d, want %d", opts.InfConvert, wantInfConvert)
	}
	enc := em.NewEncoder(io.Discard)
	err := enc.StartIndefiniteArray()
	checkNoError(t, "Encoder.StartIndefiniteArray", err)
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
		t.Errorf("em.EncOptions() = %#v, want %#v", opts2, opts1)
	}
}
