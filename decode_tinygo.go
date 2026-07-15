// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

//go:build tinygo

package cbor

const (
	defaultMaxNestedLevels = 16    // was 32 for non-tinygo (24+ for tinygo v0.33 panics tests)
	minMaxNestedLevels     = 4     // same as non-tinygo
	maxMaxNestedLevels     = 65535 // same as non-tinygo (to allow testing)
)
