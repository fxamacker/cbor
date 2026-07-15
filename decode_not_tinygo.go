// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

//go:build !tinygo

package cbor

const (
	defaultMaxNestedLevels = 32
	minMaxNestedLevels     = 4
	maxMaxNestedLevels     = 65535
)
