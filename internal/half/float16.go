// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

/*
 Modified in 2019 by Montgomery Edwards⁴⁴⁸ (github.com/x448)

 Original file was CC0 1.0 Public Domain licensed on Dec 20, 2019 
 from https://github.com/dereklstinson/half

 go-float16 - IEEE 754 binary16 half precision format
 Written in 2013 by h2so5 <mail@h2so5.net>
*/

// Package half is an IEEE 754 binary16 half precision format.
package half

import (
	"math"
	"strconv"
)

// A Float16 represents a 16-bit floating point number.
type Float16 uint16

//String satisfies the fmt Stringer interface
func (f Float16) String() string {
	return strconv.FormatFloat(float64(f.Float32()), 'f', -1, 32)
}

// NewFloat16 allocates and returns a new Float16 set to f.
func NewFloat16(f float32) Float16 {
	i := math.Float32bits(f)
	sign := uint16((i >> 31) & 0x1)
	exp := (i >> 23) & 0xff
	exp16 := int16(exp) - 112
	frac := uint16(i>>13) & 0x3ff
	switch exp {
	case 0:
		exp16 = 0
	case 0xff:
		exp16 = 0x1f
	default:
		if exp16 > 0x1e {
			exp16 = 0x1f
			frac = 0
		} else if exp16 < 0x01 {
			exp16 = 0
			frac = 0
		}
	}

	return (Float16)((sign << 15) | uint16(exp16<<10) | frac)
}

// Float32 returns the float32 representation of f.
func (f Float16) Float32() float32 {
	sign := uint32((f >> 15) & 0x1)
	exp := (f >> 10) & 0x1f
	exp32 := uint32(exp) + 127 - 15
	if exp == 0 {
		exp32 = 0
	} else if exp == 0x1f {
		exp32 = 0xff
	}
	return math.Float32frombits((sign << 31) | (exp32 << 23) | ((uint32)(f&0x3ff) << 13))
}
