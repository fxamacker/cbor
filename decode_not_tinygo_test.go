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
