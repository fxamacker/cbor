package cbor

import "reflect"

// SimpleValue represents CBOR simple value [0, 255].
type SimpleValue uint8

var (
	typeSimpleValue = reflect.TypeOf(SimpleValue(0))
)
