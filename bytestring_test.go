// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import "testing"

func TestByteString(t *testing.T) {
	type s1 struct {
		A ByteString `cbor:"a"`
	}
	type s2 struct {
		A *ByteString `cbor:"a"`
	}
	type s3 struct {
		A ByteString `cbor:"a,omitempty"`
	}
	type s4 struct {
		A *ByteString `cbor:"a,omitempty"`
	}

	emptybs := ByteString("")
	bs := ByteString("\x01\x02\x03\x04")

	testCases := []roundTripTest{
		{
			name:         "empty",
			obj:          emptybs,
			wantCborData: hexDecode("40"),
		},
		{
			name:         "not empty",
			obj:          bs,
			wantCborData: hexDecode("4401020304"),
		},
		{
			name:         "array",
			obj:          []ByteString{bs},
			wantCborData: hexDecode("814401020304"),
		},
		{
			name:         "map with ByteString key",
			obj:          map[ByteString]bool{bs: true},
			wantCborData: hexDecode("a14401020304f5"),
		},
		{
			name:         "empty ByteString field",
			obj:          s1{},
			wantCborData: hexDecode("a1616140"),
		},
		{
			name:         "not empty ByteString field",
			obj:          s1{A: bs},
			wantCborData: hexDecode("a161614401020304"),
		},
		{
			name:         "nil *ByteString field",
			obj:          s2{},
			wantCborData: hexDecode("a16161f6"),
		},
		{
			name:         "empty *ByteString field",
			obj:          s2{A: &emptybs},
			wantCborData: hexDecode("a1616140"),
		},
		{
			name:         "not empty *ByteString field",
			obj:          s2{A: &bs},
			wantCborData: hexDecode("a161614401020304"),
		},
		{
			name:         "empty ByteString field with omitempty option",
			obj:          s3{},
			wantCborData: hexDecode("a0"),
		},
		{
			name:         "not empty ByteString field with omitempty option",
			obj:          s3{A: bs},
			wantCborData: hexDecode("a161614401020304"),
		},
		{
			name:         "nil *ByteString field with omitempty option",
			obj:          s4{},
			wantCborData: hexDecode("a0"),
		},
		{
			name:         "empty *ByteString field with omitempty option",
			obj:          s4{A: &emptybs},
			wantCborData: hexDecode("a1616140"),
		},
		{
			name:         "not empty *ByteString field with omitempty option",
			obj:          s4{A: &bs},
			wantCborData: hexDecode("a161614401020304"),
		},
	}

	em, _ := EncOptions{}.EncMode()
	dm, _ := DecOptions{}.DecMode()
	testRoundTrip(t, testCases, em, dm)
}
