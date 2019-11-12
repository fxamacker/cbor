[![Build Status](https://travis-ci.com/fxamacker/cbor.svg?branch=master)](https://travis-ci.com/fxamacker/cbor)
[![codecov](https://codecov.io/gh/fxamacker/cbor/branch/master/graph/badge.svg?v=4)](https://codecov.io/gh/fxamacker/cbor)
[![Go Report Card](https://goreportcard.com/badge/github.com/fxamacker/cbor)](https://goreportcard.com/report/github.com/fxamacker/cbor)
[![Release](https://img.shields.io/github/release/fxamacker/cbor.svg?style=flat-square)](https://github.com/fxamacker/cbor/releases)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/fxamacker/cbor)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/fxamacker/cbor/master/LICENSE)

# fxamacker/cbor - CBOR library in Go    

<p align="center">
  <img src="https://user-images.githubusercontent.com/57072051/68631010-c1c5ae80-04ae-11ea-8124-dd007bd19800.png" alt="fxamacker/cbor text logo with checkboxes">
</p>

CBOR is a concise binary alternative to JSON, and is specified in [RFC 7049](https://tools.ietf.org/html/rfc7049).

This library makes encoding and decoding CBOR painless.  It's a great fit for a variety of projects using CBOR.

It’s small enough for IoT. And reliable enough for [WebAuthn (FIDO2) servers](https://github.com/fxamacker/webauthn).  It avoids crashes and exploits when decoding malicious CBOR data.

Install with ```go get github.com/fxamacker/cbor``` and use it like Go's ```encoding/json```.  It supports `` `json:"name"` `` keys!

## Design Goals 
This CBOR library is designed to be:
* __Easy__ – idiomatic API like `encoding/json` to reduce learning curve.
* __Small and self-contained__ – compiles to under 0.5 MB and has no external dependencies.
* __Safe and reliable__ – no `unsafe` pkg, coverage >95%, [fuzz tested](#fuzzing-and-code-coverage), data validation to prevent crashes on malformed or malicious data.

Competing factors are balanced:
* __Speed__ vs __safety__ vs __size__ – to keep size small, avoid code generation. For safety, validate data and avoid Go's unsafe package.  For speed, use  safe optimizations: cache struct metadata, bypass reflect when appropriate, use sync.Pool to reuse transient objects, and etc.
* __Standards compliance__ – support [CBOR](https://tools.ietf.org/html/rfc7049), including [canonical CBOR encodings](https://tools.ietf.org/html/rfc7049#section-3.9) (RFC 7049 and [CTAP2](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form)) with minor [limitations](#limitations). For example, negative numbers that can't fit into Go's int64 aren’t supported (like `encoding/json`.)

Faster CBOR libraries exist. Choose this one if you value your time, program size, and reliability. There's practically no learning curve if you know `encoding/json`.

Avoiding crashes caused by malformed or malicious CBOR data means fewer headaches.  See [Fuzzing and Coverage](#fuzzing-and-code-coverage).

## Current Status
Version 1.x has:
* __Stable API__ – won't make breaking API changes.  
* __Stable requirements__ – will always support Go v1.12.  
* __Passed fuzzing__ – v1.2 passed 42 hours of [cbor-fuzz](https://github.com/fxamacker/cbor-fuzz).  See [Fuzzing and Code Coverage](#fuzzing-and-code-coverage).

Nov 05, 2019: v1.2 adds RawMessage type, Marshaler and Unmarshaler interfaces.  Passed 42+ hrs of fuzzing.

[:rocket: Milestone v1.3](https://github.com/fxamacker/cbor/milestone/2) will improve speed, reduce memory use, and simplify decoding COSE data (RFC 8152).

## Size Comparisons
Libraries and programs were compiled for linux_amd64 using Go 1.12.
 
![alt text](https://user-images.githubusercontent.com/33205765/68306684-9c304380-006f-11ea-8661-c87592bcaa51.png "Library and program size comparison chart")

## Features
* Idiomatic API like `encoding/json`.
* Decode slices, maps, and structs in-place.
* Decode into struct with field name case-insensitive match.
* Support canonical CBOR encoding for map/struct.
* Support both "cbor" and "json" keys for struct field format tags.
* Encode anonymous struct fields by `encoding/json` package struct fields visibility rules.
* Encode and decode nil slice/map/pointer/interface values correctly.
* Encode and decode indefinite length bytes/string/array/map (["streaming"](https://tools.ietf.org/html/rfc7049#section-2.2)).
* Encode and decode time.Time as RFC 3339 formatted text string or Unix time.
* v1.1 -- Support `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler` interfaces.
* v1.2 -- `cbor.RawMessage` can delay CBOR decoding or precompute CBOR encoding.
* v1.2 -- User-defined types can have custom CBOR encoding and decoding by implementing `cbor.Marshaler` and `cbor.Unmarshaler` interfaces. 

## Fuzzing and Code Coverage
Each release passes coverage-guided fuzzing using [fxamacker/cbor-fuzz](https://github.com/fxamacker/cbor-fuzz).  Default corpus has:
* 2 files related to WebAuthn (FIDO U2F key).
* 17 files with [COSE examples (RFC 8152 Appendix B & C)](https://github.com/cose-wg/Examples/tree/master/RFC8152).
* 82 files with [CBOR examples (RFC 7049 Appendix A) ](https://tools.ietf.org/html/rfc7049#appendix-A).
* 340 files generated by fuzzing for 50 hours with 2 workers on AMD EPYC 7601 virtual machine.

Unit tests include all RFC 7049 examples, bugs found by fuzzing, 2 maliciously crafted CBOR data, and etc.

Minimum code coverage is 95%.  Minimum fuzzing is 10 hours for each release but often longer (v1.2 passed 42+ hours.)

Code coverage is 97.8% (`go test -cover`) for cbor v1.2 which is among the highest for libraries of this type.

## Standards 
This library implements CBOR as specified in [RFC 7049](https://tools.ietf.org/html/rfc7049), with minor [limitations](#limitations).

It also supports [canonical CBOR encodings](https://tools.ietf.org/html/rfc7049#section-3.9) (both RFC 7049 and [CTAP2](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form)).  CTAP2 canonical CBOR encoding is used by [CTAP](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html) and [WebAuthn](https://www.w3.org/TR/webauthn/) in [FIDO2](https://fidoalliance.org/fido2/) framework.

## Limitations
* CBOR tags (type 6) are ignored.  Decoder simply decodes tagged data after ignoring the tags.
* CBOR negative int (type 1) that cannot fit into Go's int64 are not supported, such as RFC 7049 example -18446744073709551616.  Decoding these values returns `cbor.UnmarshalTypeError` like Go's `encoding/json`.
* CBOR `Undefined` (0xf7) value decodes to Go's `nil` value.  Use CBOR `Null` (0xf6) to round-trip with Go's `nil`.

## System Requirements
* Go 1.12 (or newer)
* Tested and fuzzed on linux_amd64, but it should work on other platforms.

## Versions and API Changes
This project uses [Semantic Versioning](https://semver.org), so the API is always backwards compatible unless the major version number changes.

## API 
See [API docs](https://godoc.org/github.com/fxamacker/cbor) for more details.

```
package cbor // import "github.com/fxamacker/cbor"

func Marshal(v interface{}, encOpts EncOptions) ([]byte, error)
func Unmarshal(data []byte, v interface{}) error
func Valid(data []byte) (rest []byte, err error)
type Decoder struct{ ... }
    func NewDecoder(r io.Reader) *Decoder
    func (dec *Decoder) Decode(v interface{}) (err error)
    func (dec *Decoder) NumBytesRead() int
type EncOptions struct{ ... }
type Encoder struct{ ... }
    func NewEncoder(w io.Writer, encOpts EncOptions) *Encoder
    func (enc *Encoder) Encode(v interface{}) error
    func (enc *Encoder) StartIndefiniteByteString() error
    func (enc *Encoder) StartIndefiniteTextString() error
    func (enc *Encoder) StartIndefiniteArray() error
    func (enc *Encoder) StartIndefiniteMap() error
    func (enc *Encoder) EndIndefinite() error
type InvalidUnmarshalError struct{ ... }
type Marshaler interface{ ... }
type RawMessage []byte
type SemanticError struct{ ... }
type SyntaxError struct{ ... }
type UnmarshalTypeError struct{ ... }
type Unmarshaler interface{ ... }
type UnsupportedTypeError struct{ ... }
```

## Installation
```
go get github.com/fxamacker/cbor
```

## Usage
See [examples](example_test.go).

Decoding:

```
// create a decoder
dec := cbor.NewDecoder(reader)

// decode into empty interface
var i interface{}
err = dec.Decode(&i)

// decode into struct 
var stru ExampleStruct
err = dec.Decode(&stru)

// decode into map
var m map[string]string
err = dec.Decode(&m)

// decode into primitive
var f float32
err = dec.Decode(&f)
```

Encoding:

```
// create an encoder with canonical CBOR encoding enabled
enc := cbor.NewEncoder(writer, cbor.EncOptions{Canonical: true})

// encode struct
err = enc.Encode(stru)

// encode map
err = enc.Encode(m)

// encode primitive
err = enc.Encode(f)
```

Encoding indefinite length array:

```
enc := cbor.NewEncoder(writer, cbor.EncOptions{})

// start indefinite length array encoding
err = enc.StartIndefiniteArray()

// encode array element
err = enc.Encode(1)

// encode array element
err = enc.Encode([]int{2, 3})

// start nested indefinite length array as array element
err = enc.StartIndefiniteArray()

// encode nested array element
err = enc.Encode(4)

// encode nested array element
err = enc.Encode(5)

// end nested indefinite length array
err = enc.EndIndefinite()

// end indefinite length array
err = enc.EndIndefinite()
```

## Benchmarks
Go structs are faster than maps:
* decoding into struct is >66% faster than decoding into map.
* encoding struct is >63% faster than encoding map.

See [Benchmarks for fxamacker/cbor](BENCHMARKS.md).

## Code of Conduct 
This project has adopted the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).  Contact [faye.github@gmail.com](mailto:faye.github@gmail.com) with any questions or comments.

## Contributing
Please refer to [How to Contribute](CONTRIBUTING.md).

## Security Policy
For v1, security fixes are provided only for the latest released version since the API won't break compatibility.

To report security vulnerabilities, please email [faye.github@gmail.com](mailto:faye.github@gmail.com) and allow time for the problem to be resolved before reporting it to the public.

## Disclaimers
Phrases like "NO CRASHES" and "NO EXPLOITS" mean there are none known to the maintainer based on results of unit tests and fuzzing.  It doesn't imply the software is perfect or 100% invulnerable to all known and unknown attacks.

Please read the license for additional disclaimers and terms.

## License 
Copyright (c) 2019 [Faye Amacker](https://github.com/fxamacker)

Licensed under [MIT License](LICENSE)
