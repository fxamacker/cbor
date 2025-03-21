// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"reflect"
	"testing"
)

func TestUnmarshalSimpleValue(t *testing.T) {
	t.Run("0..23", func(t *testing.T) {
		for i := 0; i <= 23; i++ {
			data := []byte{byte(cborTypePrimitives) | byte(i)}
			want := SimpleValue(i)

			switch i {
			case 20: // false
				testUnmarshalSimpleValueToEmptyInterface(t, data, false)
			case 21: // true
				testUnmarshalSimpleValueToEmptyInterface(t, data, true)
			case 22: // null
				testUnmarshalSimpleValueToEmptyInterface(t, data, nil)
			case 23: // undefined
				testUnmarshalSimpleValueToEmptyInterface(t, data, nil)
			default:
				testUnmarshalSimpleValueToEmptyInterface(t, data, want)
			}

			testUnmarshalSimpleValue(t, data, want)
		}
	})

	t.Run("24..31", func(t *testing.T) {
		for i := 24; i <= 31; i++ {
			data := []byte{byte(cborTypePrimitives) | byte(24), byte(i)}

			testUnmarshalInvalidSimpleValueToEmptyInterface(t, data)
			testUnmarshalInvalidSimpleValue(t, data)
		}
	})

	t.Run("32..255", func(t *testing.T) {
		for i := 32; i <= 255; i++ {
			data := []byte{byte(cborTypePrimitives) | byte(24), byte(i)}
			want := SimpleValue(i)
			testUnmarshalSimpleValueToEmptyInterface(t, data, want)
			testUnmarshalSimpleValue(t, data, want)
		}
	})
}

func testUnmarshalInvalidSimpleValueToEmptyInterface(t *testing.T, data []byte) {
	var v interface{}
	if err := Unmarshal(data, v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", data)
	} else if _, ok := err.(*SyntaxError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*SyntaxError)", data, err)
	}
}

func testUnmarshalInvalidSimpleValue(t *testing.T, data []byte) {
	var v SimpleValue
	if err := Unmarshal(data, v); err == nil {
		t.Errorf("Unmarshal(0x%x) didn't return an error", data)
	} else if _, ok := err.(*SyntaxError); !ok {
		t.Errorf("Unmarshal(0x%x) returned wrong error type %T, want (*SyntaxError)", data, err)
	}
}

func testUnmarshalSimpleValueToEmptyInterface(t *testing.T, data []byte, want interface{}) {
	var v interface{}
	if err := Unmarshal(data, &v); err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", data, err)
		return
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", data, v, v, want, want)
	}
}

func testUnmarshalSimpleValue(t *testing.T, data []byte, want SimpleValue) {
	cborNil := isCBORNil(data)

	// Decode to SimpleValue
	var v SimpleValue
	err := Unmarshal(data, &v)
	if err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", data, err)
		return
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", data, v, v, want, want)
	}

	// Decode to uninitialized *SimpleValue
	var pv *SimpleValue
	err = Unmarshal(data, &pv)
	if err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", data, err)
		return
	}
	if cborNil {
		if pv != nil {
			t.Errorf("Unmarshal(0x%x) returned %v, want nil *SimpleValue", data, *pv)
		}
	} else {
		if !reflect.DeepEqual(*pv, want) {
			t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", data, *pv, *pv, want, want)
		}
	}

	// Decode to initialized *SimpleValue
	v = SimpleValue(0)
	pv = &v
	err = Unmarshal(data, &pv)
	if err != nil {
		t.Errorf("Unmarshal(0x%x) returned error %v", data, err)
		return
	}
	if cborNil {
		if pv != nil {
			t.Errorf("Unmarshal(0x%x) returned %v, want nil *SimpleValue", data, *pv)
		}
	} else {
		if !reflect.DeepEqual(v, want) {
			t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", data, v, v, want, want)
		}
	}
}

func TestMarshalSimpleValue(t *testing.T) {
	t.Run("0..23", func(t *testing.T) {
		for i := 0; i <= 23; i++ {
			wantData := []byte{byte(cborTypePrimitives) | byte(i)}
			v := SimpleValue(i)

			data, err := Marshal(v)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", v, err)
				continue
			}
			if !bytes.Equal(data, wantData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, data, wantData)
			}
		}
	})

	t.Run("24..31", func(t *testing.T) {
		for i := 24; i <= 31; i++ {
			v := SimpleValue(i)

			if data, err := Marshal(v); err == nil {
				t.Errorf("Marshal(%v) didn't return an error", data)
			} else if _, ok := err.(*UnsupportedValueError); !ok {
				t.Errorf("Marshal(%v) returned wrong error type %T, want (*UnsupportedValueError)", data, err)
			}
		}
	})

	t.Run("32..255", func(t *testing.T) {
		for i := 32; i <= 255; i++ {
			wantData := []byte{byte(cborTypePrimitives) | byte(24), byte(i)}
			v := SimpleValue(i)

			data, err := Marshal(v)
			if err != nil {
				t.Errorf("Marshal(%v) returned error %v", v, err)
				continue
			}
			if !bytes.Equal(data, wantData) {
				t.Errorf("Marshal(%v) = 0x%x, want 0x%x", v, data, wantData)
			}
		}
	})
}
