// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

//go:build !tinygo

package cbor

import (
	"testing"
)

// unsupportedFuncTypeErrMsg is the error message for marshaling func(int) int.
const unsupportedFuncTypeErrMsg = "cbor: unsupported type: func(int) int"

func TestUnmarshalDeepNesting(t *testing.T) {
	// Construct this object rather than embed such a large constant in the code
	type TestNode struct {
		Value int
		Child *TestNode
	}
	n := &TestNode{Value: 0}
	root := n
	for i := range 65534 {
		child := &TestNode{Value: i}
		n.Child = child
		n = child
	}
	em := mustEncMode(t, EncOptions{})
	data, err := em.Marshal(root)
	checkNoError(t, "em.Marshal", err)

	// Try unmarshal it
	dm := mustDecMode(t, DecOptions{MaxNestedLevels: 65535})
	var readback TestNode
	err = unmarshalWithDM(t, dm, data, &readback)
	checkUnmarshal(t, "dm.Unmarshal", readback, err, *root)
}
