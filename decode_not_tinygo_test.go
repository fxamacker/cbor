// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

//go:build !tinygo

package cbor

import (
	"reflect"
	"testing"
)

func TestUnmarshalDeepNesting(t *testing.T) {
	// Construct this object rather than embed such a large constant in the code
	type TestNode struct {
		Value int
		Child *TestNode
	}
	n := &TestNode{Value: 0}
	root := n
	for i := 0; i < 65534; i++ {
		child := &TestNode{Value: i}
		n.Child = child
		n = child
	}
	em, err := EncOptions{}.EncMode()
	if err != nil {
		t.Errorf("EncMode() returned error %v", err)
	}
	data, err := em.Marshal(root)
	if err != nil {
		t.Errorf("Marshal() deeply nested object returned error %v", err)
	}

	// Try unmarshal it
	dm, err := DecOptions{MaxNestedLevels: 65535}.DecMode()
	if err != nil {
		t.Errorf("DecMode() returned error %v", err)
	}
	var readback TestNode
	err = dm.Unmarshal(data, &readback)
	if err != nil {
		t.Errorf("Unmarshal() of deeply nested object returned error: %v", err)
	}
	if !reflect.DeepEqual(root, &readback) {
		t.Errorf("Unmarshal() of deeply nested object did not match\nGot: %#v\n Want: %#v\n",
			&readback, root)
	}
}

func TestUnmarshalRegisteredTagToInterface(t *testing.T) {
	var err error
	tags := NewTagSet()
	err = tags.Add(TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, reflect.TypeOf(C{}), 279)
	if err != nil {
		t.Error(err)
	}
	err = tags.Add(TagOptions{EncTag: EncTagRequired, DecTag: DecTagRequired}, reflect.TypeOf(D{}), 280)
	if err != nil {
		t.Error(err)
	}

	encMode, _ := PreferredUnsortedEncOptions().EncModeWithTags(tags)
	decMode, _ := DecOptions{}.DecModeWithTags(tags)

	v1 := A1{Field: &C{Field: 5}}
	data1, err := encMode.Marshal(v1)
	if err != nil {
		t.Fatalf("Marshal(%+v) returned error %v", v1, err)
	}

	v2 := A2{Fields: []B{&C{Field: 5}, &D{Field: "a"}}}
	data2, err := encMode.Marshal(v2)
	if err != nil {
		t.Fatalf("Marshal(%+v) returned error %v", v2, err)
	}

	testCases := []struct {
		name           string
		data           []byte
		unmarshalToObj interface{}
		wantValue      interface{}
	}{
		{
			name:           "interface type",
			data:           data1,
			unmarshalToObj: &A1{},
			wantValue:      &v1,
		},
		{
			name:           "slice of interface type",
			data:           data2,
			unmarshalToObj: &A2{},
			wantValue:      &v2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err = decMode.Unmarshal(tc.data, tc.unmarshalToObj)
			if err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.data, err)
			}
			if !reflect.DeepEqual(tc.unmarshalToObj, tc.wantValue) {
				t.Errorf("Unmarshal(0x%x) = %v, want %v", tc.data, tc.unmarshalToObj, tc.wantValue)
			}
		})
	}
}
