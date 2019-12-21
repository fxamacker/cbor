// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

/*
 Modified in 2019 by Montgomery Edwards⁴⁴⁸ (github.com/x448)
 Original file was CC0 1.0 Public Domain licensed on Dec 20, 2019 
 from https://github.com/dereklstinson/half
 
 go-float16 - IEEE 754 binary16 half precision format
 Written in 2013 by h2so5 <mail@h2so5.net>
*/

package half

import (
	"math"
	"strconv"
	"testing"
)

func getFloatTable() map[Float16]float32 {
	table := map[Float16]float32{
		0x3c00: 1,
		0x4000: 2,
		0xc000: -2,
		0x7bfe: 65472,
		0x7bff: 65504,
		0xfbff: -65504,
		0x0000: 0,
		0x8000: float32(math.Copysign(0, -1)),
		0x7c00: float32(math.Inf(1)),
		0xfc00: float32(math.Inf(-1)),
		0x5b8f: 241.875,
		0x48c8: 9.5625,
	}
	return table
}

func TestFloat32(t *testing.T) {
	for k, v := range getFloatTable() {
		f := k.Float32()
		if f != v {
			t.Errorf("ToFloat32(%d) = %f, want %f.", k, f, v)
		}
	}
}

func TestFloat16Print(t *testing.T) {
	for k, v := range getFloatTable() {
		if k.String() != strconv.FormatFloat(float64(v), 'f', -1, 32) {
			s1 := strconv.FormatFloat(float64(v), 'f', -1, 32)
			s2 := k.String()
			t.Errorf("K fmt is %s, v fmt is %s", s1, s2)
		}

	}
}

func TestNewFloat16(t *testing.T) {
	for k, v := range getFloatTable() {
		i := NewFloat16(v)
		if i != k {
			t.Errorf("FromFloat32(%f) = %d, want %d.", v, i, k)
		}
	}
}
