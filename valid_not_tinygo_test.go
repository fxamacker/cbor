// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

//go:build !tinygo

package cbor

import "testing"

func Test32Depth(t *testing.T) {
	testCases := []struct {
		name      string
		data      []byte
		wantDepth int
	}{
		{"32-level array", hexDecode("82018181818181818181818181818181818181818181818181818181818181818101"), 32},
		{"32-level indefinite length array", hexDecode("9f018181818181818181818181818181818181818181818181818181818181818101ff"), 32},
		{"32-level map", hexDecode("a1018181818181818181818181818181818181818181818181818181818181818101"), 32},
		{"32-level indefinite length map", hexDecode("bf018181818181818181818181818181818181818181818181818181818181818101ff"), 32},
		{"32-level tag", hexDecode("d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d864d86474323031332d30332d32315432303a30343a30305a"), 32}, // 100(100(...("2013-03-21T20:04:00Z")))
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := decoder{data: tc.data, dm: defaultDecMode}
			depth, err := d.wellformedInternal(0, false)
			if err != nil {
				t.Errorf("wellformed(0x%x) returned error %v", tc.data, err)
			}
			if depth != tc.wantDepth {
				t.Errorf("wellformed(0x%x) returned depth %d, want %d", tc.data, depth, tc.wantDepth)
			}
		})
	}
}
