// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"reflect"
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

// TestEncodeStructToIndefArrayBasic verifies the encoded bytes for a
// typical struct and that the data roundtrips through the default decoder.
func TestEncodeStructToIndefArrayBasic(t *testing.T) {
	type S struct {
		_     struct{} `cbor:",toindefarray"`
		Data  []byte
		Count int
	}

	em := allowToIndefArrayMode(t)
	in := S{Data: []byte{0x73, 0xf7, 0x2d, 0xbe}, Count: 3}
	want := []byte{0x9f, 0x44, 0x73, 0xf7, 0x2d, 0xbe, 0x03, 0xff} // EDN: [_ h'73f72dbe', 3]

	got, err := em.Marshal(in)
	checkMarshal(t, "em.Marshal", got, err, want)

	// Roundtrip: decoder accepts both definite and indefinite arrays for
	// toindefarray-tagged structs (the tag treats them the same way as
	// `toarray` for decoding).
	var out S
	err = unmarshal(t, got, &out)
	checkUnmarshal(t, "Unmarshal", out, err, in)
}

// TestEncodeStructToIndefArrayEmpty verifies that an empty struct produces
// exactly the two bytes 0x9f 0xff (no inner fields).
func TestEncodeStructToIndefArrayEmpty(t *testing.T) {
	type Empty struct {
		_ struct{} `cbor:",toindefarray"`
	}

	em := allowToIndefArrayMode(t)
	in := Empty{}
	want := []byte{0x9f, 0xff} // EDN: [_]

	got, err := em.Marshal(in)
	checkMarshal(t, "em.Marshal", got, err, want)

	var out Empty
	err = unmarshal(t, got, &out)
	checkUnmarshal(t, "Unmarshal", out, err, in)
}

// TestEncodeStructToIndefArrayNested verifies the byte layout and roundtrip
// of a struct with toindefarray containing another struct with toindefarray.
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

	em := allowToIndefArrayMode(t)
	in := Outer{
		Inner: Inner{X: []byte{0xab, 0xcd}, Y: 42},
		Z:     7,
	}
	want := []byte{0x9f, 0x9f, 0x42, 0xab, 0xcd, 0x18, 0x2a, 0xff, 0x07, 0xff} // EDN: [_ [_ h'abcd', 42], 7]

	got, err := em.Marshal(in)
	checkMarshal(t, "em.Marshal", got, err, want)

	var out Outer
	err = unmarshal(t, got, &out)
	checkUnmarshal(t, "Unmarshal", out, err, in)
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

	b, err := Marshal(S{A: 1})
	checkMarshalErr[error](t, "Marshal", b, err, "cbor: cannot encode struct cbor.S with `toindefarray` when ToIndefArrayStructTag is not ToIndefArrayStructTagAllowed")
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
	b, err := em.Marshal(Bad{A: 1})
	checkMarshalErr[error](t, "em.Marshal", b, err, "cbor: struct cbor.Bad cannot have both \"toarray\" and \"toindefarray\" options")
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
	in := Outer{Inner: Inner{A: 1, B: 2}, C: 3}
	want := []byte{0x9f, 0x01, 0x02, 0x03, 0xff} // EDN: [_ 1, 2, 3]  (promoted A, B, then C)

	got, err := em.Marshal(in)
	checkMarshal(t, "em.Marshal", got, err, want)
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
	in := Outer{Inner: Inner{A: 1, B: 2}, C: 3}
	want := []byte{0x9f, 0x82, 0x01, 0x02, 0x03, 0xff} // EDN: [_ [1, 2], 3]

	got, err := em.Marshal(in)
	checkMarshal(t, "em.Marshal", got, err, want)
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
	in := Outer{Inner: Inner{A: 1, B: 2}, C: 3}
	want := []byte{0x82, 0x9f, 0x01, 0x02, 0xff, 0x03} // EDN: [[_ 1, 2], 3]

	got, err := em.Marshal(in)
	checkMarshal(t, "em.Marshal", got, err, want)
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
		reflect.TypeFor[S](),
		121,
	); err != nil {
		t.Fatalf("TagSet.Add returned error: %v", err)
	}

	em, err := EncOptions{ToIndefArrayStructTag: ToIndefArrayStructTagAllowed}.EncModeWithTags(tags)
	if err != nil {
		t.Fatalf("EncModeWithTags returned error: %v", err)
	}
	in := S{Field1: 1, Field2: 2}
	want := []byte{0xd8, 0x79, 0x9f, 0x01, 0x02, 0xff} // EDN: 121([_ 1, 2])

	got, err := em.Marshal(in)
	checkMarshal(t, "em.Marshal", got, err, want)
}
