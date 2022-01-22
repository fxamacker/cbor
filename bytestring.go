// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"encoding/hex"
	"reflect"
)

var (
	typeByteString = reflect.TypeOf(NewByteString(nil))
)

type ByteString struct {
	// XXX: replace with interface{} storing fixed-length byte array?
	// We use a string because []byte isn't comparable
	data string
}

func NewByteString(data []byte) ByteString {
	bs := ByteString{
		data: string(data),
	}
	return bs
}

func (bs ByteString) Bytes() []byte {
	return []byte(bs.data)
}

func (bs ByteString) String() string {
	return hex.EncodeToString([]byte(bs.data))
}
