// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

//go:build !tinygo

package cbor

import "reflect"

func implements(concreteType reflect.Type, interfaceType reflect.Type) bool {
	return concreteType.Implements(interfaceType) ||
		reflect.PointerTo(concreteType).Implements(interfaceType)
}
