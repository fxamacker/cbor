// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

/*
Package cbor provides a fuzz-tested CBOR encoder and decoder with full support
for float16, Canonical CBOR, CTAP2 Canonical CBOR, and custom settings.

THIS VERSION IS OUTDATED

V2 IS AVAILABLE

https://github.com/fxamacker/cbor/releases

Basics

Encoding options allow "preferred serialization" by encoding integers and floats
to their smallest forms (like float16) when values fit.

Go struct tags like `cbor:"name,omitempty"` and `json:"name,omitempty"` work as expected.
If both struct tags are specified then `cbor` is used.

Struct tags like "keyasint", "toarray", and "omitempty" make it easy to use
very compact formats like COSE and CWT (CBOR Web Tokens) with structs.

For example, the "toarray" struct tag encodes/decodes struct fields as array elements.
And "keyasint" struct tag encodes/decodes struct fields to values of maps with specified int keys.

fxamacker/cbor-fuzz provides coverage-guided fuzzing for this package.

For latest API docs, see: https://github.com/fxamacker/cbor#api
*/
package cbor
