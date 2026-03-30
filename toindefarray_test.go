package cbor

import (
	"encoding/hex"
	"testing"
)

func TestEncodeStructToIndefArray(t *testing.T) {
	type TestStruct struct {
		_     struct{} `cbor:",toindefarray"`
		Data  []byte
		Count int
	}

	ts := TestStruct{
		Data:  []byte{0x73, 0xf7, 0x2d, 0xbe},
		Count: 3,
	}

	b, err := Marshal(ts)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	h := hex.EncodeToString(b)
	t.Logf("Encoded: %s", h)

	// Check indefinite array header
	if b[0] != 0x9f {
		t.Errorf("Expected first byte 0x9f (indef array), got 0x%02x", b[0])
	}

	// Check break code at end
	if b[len(b)-1] != 0xff {
		t.Errorf("Expected last byte 0xff (break), got 0x%02x", b[len(b)-1])
	}

	// Decode back
	var ts2 TestStruct
	err = Unmarshal(b, &ts2)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if hex.EncodeToString(ts2.Data) != hex.EncodeToString(ts.Data) {
		t.Errorf("Data mismatch: got %x, want %x", ts2.Data, ts.Data)
	}
	if ts2.Count != ts.Count {
		t.Errorf("Count mismatch: got %d, want %d", ts2.Count, ts.Count)
	}
}

func TestEncodeStructToIndefArrayNested(t *testing.T) {
	type Inner struct {
		_    struct{} `cbor:",toindefarray"`
		X    []byte
		Y    int
	}

	type Outer struct {
		_     struct{} `cbor:",toindefarray"`
		Inner Inner
		Z     int
	}

	o := Outer{
		Inner: Inner{X: []byte{0xab, 0xcd}, Y: 42},
		Z:     7,
	}

	b, err := Marshal(o)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	t.Logf("Encoded: %s", hex.EncodeToString(b))

	if b[0] != 0x9f {
		t.Errorf("Expected outer 0x9f, got 0x%02x", b[0])
	}
	if b[len(b)-1] != 0xff {
		t.Errorf("Expected outer 0xff, got 0x%02x", b[len(b)-1])
	}

	var o2 Outer
	if err := Unmarshal(b, &o2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if o2.Z != 7 || o2.Inner.Y != 42 {
		t.Errorf("Decoded mismatch: %+v", o2)
	}
}
