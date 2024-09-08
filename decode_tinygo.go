// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

//go:build tinygo

package cbor

import "reflect"

const (
	defaultMaxNestedLevels = 16    // was 32 for non-tinygo (24+ for tinygo v0.33 panics tests)
	minMaxNestedLevels     = 4     // same as non-tinygo
	maxMaxNestedLevels     = 65535 // same as non-tinygo (to allow testing)
)

// tinygo v0.33 doesn't implement Type.AssignableTo() and it panics.
// Type.AssignableTo() is used under the hood for Type.Implements().
//
// More details in https://github.com/tinygo-org/tinygo/issues/4277.
//
// implements() always returns false until tinygo implements Type.AssignableTo().
func implements(concreteType reflect.Type, interfaceType reflect.Type) bool {
	return false
}
