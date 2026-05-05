// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

// allowToIndefArrayMode returns an EncMode with the toindefarray struct tag
// option allowed. It is a small helper to keep tests focused on behavior.
func allowToIndefArrayMode(t *testing.T) EncMode {
	t.Helper()
	em, err := EncOptions{ToIndefArrayStructTag: ToIndefArrayStructTagAllowed}.EncMode()
	if err != nil {
		t.Fatalf("EncMode() returned error: %v", err)
	}
	return em
}

// TestEncodeStructToIndefArrayBasic verifies header (0x9f), break (0xff),
// the encoded fields between them, and roundtrip via the default decoder.
func TestEncodeStructToIndefArrayBasic(t *testing.T) {
	type S struct {
		_     struct{} `cbor:",toindefarray"`
		Data  []byte
		Count int
	}

	in := S{Data: []byte{0x73, 0xf7, 0x2d, 0xbe}, Count: 3}

	em := allowToIndefArrayMode(t)
	got, err := em.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	if got[0] != 0x9f {
		t.Fatalf("expected first byte 0x9f, got 0x%02x; full=%x", got[0], got)
	}
	if got[len(got)-1] != 0xff {
		t.Fatalf("expected last byte 0xff, got 0x%02x; full=%x", got[len(got)-1], got)
	}

	// Roundtrip: decoder accepts both definite and indefinite arrays for
	// toindefarray-tagged structs (the tag treats them the same way as
	// `toarray` for decoding).
	var out S
	if err := Unmarshal(got, &out); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if !bytes.Equal(out.Data, in.Data) {
		t.Errorf("Data: got %x, want %x", out.Data, in.Data)
	}
	if out.Count != in.Count {
		t.Errorf("Count: got %d, want %d", out.Count, in.Count)
	}
}

// TestEncodeStructToIndefArrayEmpty verifies that an empty struct produces
// exactly the two bytes 0x9f 0xff (no inner fields).
func TestEncodeStructToIndefArrayEmpty(t *testing.T) {
	type Empty struct {
		_ struct{} `cbor:",toindefarray"`
	}

	em := allowToIndefArrayMode(t)
	got, err := em.Marshal(Empty{})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	want := []byte{0x9f, 0xff}
	if !bytes.Equal(got, want) {
		t.Errorf("got %x, want %x", got, want)
	}
}

// TestEncodeStructToIndefArrayNested verifies that a struct with toindefarray
// can contain another struct with toindefarray, producing nested 0x9f...0xff.
func TestEncodeStructToIndefArrayNested(t *testing.T) {
	type Inner struct {
		_ struct{} `cbor:",toindefarray"`
		X []byte
		Y int
	}
	type Outer struct {
		_     struct{} `cbor:",toindefarray"`
		Inner Inner
		Z     int
	}

	in := Outer{
		Inner: Inner{X: []byte{0xab, 0xcd}, Y: 42},
		Z:     7,
	}

	em := allowToIndefArrayMode(t)
	got, err := em.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	if got[0] != 0x9f {
		t.Errorf("outer head: got 0x%02x, want 0x9f", got[0])
	}
	if got[len(got)-1] != 0xff {
		t.Errorf("outer break: got 0x%02x, want 0xff", got[len(got)-1])
	}

	// Verify the inner indefinite-length array also appears.
	// Outer layout: 9f <inner-encoded> <z-encoded> ff
	// Inner layout starts with 9f and ends with ff before z.
	// We assert there is at least one inner 0x9f and a paired 0xff before
	// the final break.
	if !bytes.Contains(got[1:len(got)-1], []byte{0x9f}) {
		t.Errorf("expected nested 0x9f inside outer; got %x", got)
	}

	var out Outer
	if err := Unmarshal(got, &out); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if out.Z != in.Z || out.Inner.Y != in.Inner.Y || !bytes.Equal(out.Inner.X, in.Inner.X) {
		t.Errorf("decoded mismatch: %+v", out)
	}
}

// TestEncodeStructToIndefArrayForbiddenByDefault verifies that the default
// EncOptions (zero value) refuses to encode a struct tagged with
// `toindefarray`. The error message references the struct type and the
// option name to give callers a clear remediation path.
func TestEncodeStructToIndefArrayForbiddenByDefault(t *testing.T) {
	type S struct {
		_ struct{} `cbor:",toindefarray"`
		A int
	}

	_, err := Marshal(S{A: 1})
	if err == nil {
		t.Fatal("expected error from default Marshal, got nil")
	}
	if !strings.Contains(err.Error(), "ToIndefArrayStructTag") {
		t.Errorf("error should mention ToIndefArrayStructTag, got: %v", err)
	}
}

// TestEncodeStructToIndefArrayMutuallyExclusive verifies that a struct
// declaring both `toarray` and `toindefarray` is rejected at encode time
// regardless of mode (the rejection comes from the type cache itself).
func TestEncodeStructToIndefArrayMutuallyExclusive(t *testing.T) {
	type Bad struct {
		_ struct{} `cbor:",toarray,toindefarray"`
		A int
	}

	em := allowToIndefArrayMode(t)
	_, err := em.Marshal(Bad{A: 1})
	if err == nil {
		t.Fatal("expected error for struct declaring both toarray and toindefarray, got nil")
	}
	if !strings.Contains(err.Error(), "toarray") || !strings.Contains(err.Error(), "toindefarray") {
		t.Errorf("error should mention both options, got: %v", err)
	}
}

// TestToIndefArrayStructTagOptionsRoundTrip verifies that the non-zero value
// of ToIndefArrayStructTag is preserved through EncOptions -> EncMode ->
// EncOptions. TestEncOptions cannot cover this because the non-zero values
// of IndefLength and ToIndefArrayStructTag are mutually exclusive.
func TestToIndefArrayStructTagOptionsRoundTrip(t *testing.T) {
	opts := EncOptions{ToIndefArrayStructTag: ToIndefArrayStructTagAllowed}
	em, err := opts.EncMode()
	if err != nil {
		t.Fatalf("EncMode() returned error: %v", err)
	}
	got := em.EncOptions().ToIndefArrayStructTag
	if got != ToIndefArrayStructTagAllowed {
		t.Errorf("ToIndefArrayStructTag round-trip: got %v, want %v", got, ToIndefArrayStructTagAllowed)
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

// TestEncodeStructToIndefArrayEmbeddedField verifies that fields promoted
// from an embedded struct are encoded as elements of the indefinite-length
// array, in the same order they would appear under toarray.
func TestEncodeStructToIndefArrayEmbeddedField(t *testing.T) {
	type Inner struct {
		A uint64
		B uint64
	}
	type Outer struct {
		_ struct{} `cbor:",toindefarray"`
		Inner
		C uint64
	}

	em := allowToIndefArrayMode(t)
	got, err := em.Marshal(Outer{Inner: Inner{A: 1, B: 2}, C: 3})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	// 9f 01 02 03 ff -- promoted A, B, then C, between head and break.
	want := []byte{0x9f, 0x01, 0x02, 0x03, 0xff}
	if !bytes.Equal(got, want) {
		t.Errorf("got %x, want %x", got, want)
	}
}

// TestEncodeStructToIndefArrayParentIndefChildArray verifies that a struct
// using toindefarray can contain a struct using toarray, and the inner
// struct still encodes as a definite-length array.
func TestEncodeStructToIndefArrayParentIndefChildArray(t *testing.T) {
	type Inner struct {
		_ struct{} `cbor:",toarray"`
		A uint64
		B uint64
	}
	type Outer struct {
		_     struct{} `cbor:",toindefarray"`
		Inner Inner
		C     uint64
	}

	em := allowToIndefArrayMode(t)
	got, err := em.Marshal(Outer{Inner: Inner{A: 1, B: 2}, C: 3})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	// 9f       => outer indefinite-length array head
	//   82     => inner definite-length array of 2 elements
	//     01   => uint 1
	//     02   => uint 2
	//   03     => uint 3
	// ff       => break
	want := []byte{0x9f, 0x82, 0x01, 0x02, 0x03, 0xff}
	if !bytes.Equal(got, want) {
		t.Errorf("got %x, want %x", got, want)
	}
}

// TestEncodeStructToIndefArrayParentArrayChildIndef verifies the symmetric
// case: a parent with toarray containing a child with toindefarray.
func TestEncodeStructToIndefArrayParentArrayChildIndef(t *testing.T) {
	type Inner struct {
		_ struct{} `cbor:",toindefarray"`
		A uint64
		B uint64
	}
	type Outer struct {
		_     struct{} `cbor:",toarray"`
		Inner Inner
		C     uint64
	}

	em := allowToIndefArrayMode(t)
	got, err := em.Marshal(Outer{Inner: Inner{A: 1, B: 2}, C: 3})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	// 82       => outer definite-length array of 2 elements
	//   9f     => inner indefinite-length array head
	//     01   => uint 1
	//     02   => uint 2
	//   ff     => inner break
	//   03     => uint 3
	want := []byte{0x82, 0x9f, 0x01, 0x02, 0xff, 0x03}
	if !bytes.Equal(got, want) {
		t.Errorf("got %x, want %x", got, want)
	}
}

// TestEncodeStructToIndefArrayWithCBORTag verifies that a struct registered
// with a CBOR tag through TagSet and tagged `toindefarray` produces the tag
// header followed by an indefinite-length array of its fields.
func TestEncodeStructToIndefArrayWithCBORTag(t *testing.T) {
	type S struct {
		_      struct{} `cbor:",toindefarray"`
		Field1 uint64
		Field2 uint64
	}

	tags := NewTagSet()
	if err := tags.Add(
		TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired},
		reflect.TypeOf(S{}),
		121,
	); err != nil {
		t.Fatalf("TagSet.Add returned error: %v", err)
	}

	em, err := EncOptions{ToIndefArrayStructTag: ToIndefArrayStructTagAllowed}.EncModeWithTags(tags)
	if err != nil {
		t.Fatalf("EncModeWithTags returned error: %v", err)
	}

	got, err := em.Marshal(S{Field1: 1, Field2: 2})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	// d8 79  => tag 121
	// 9f     => indefinite-length array head
	// 01     => uint 1
	// 02     => uint 2
	// ff     => break
	want := []byte{0xd8, 0x79, 0x9f, 0x01, 0x02, 0xff}
	if !bytes.Equal(got, want) {
		t.Errorf("got %x, want %x", got, want)
	}
}
