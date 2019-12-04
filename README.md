[![CBOR Library in Go/Golang](https://user-images.githubusercontent.com/57072051/69258148-c874b580-0b81-11ea-982d-e44b21f3a0fe.png)](https://github.com/fxamacker/cbor/releases)

# CBOR library in Go
This library encodes and decodes CBOR.  It's been fuzz tested since v0.1 and fast since v1.3.

[![Build Status](https://travis-ci.com/fxamacker/cbor.svg?branch=master)](https://travis-ci.com/fxamacker/cbor)
[![codecov](https://codecov.io/gh/fxamacker/cbor/branch/master/graph/badge.svg?v=4)](https://codecov.io/gh/fxamacker/cbor)
[![Go Report Card](https://goreportcard.com/badge/github.com/fxamacker/cbor)](https://goreportcard.com/report/github.com/fxamacker/cbor)
[![Release](https://img.shields.io/github/release/fxamacker/cbor.svg?style=flat-square)](https://github.com/fxamacker/cbor/releases)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/fxamacker/cbor/master/LICENSE)

__What is CBOR__?  [CBOR](CBOR.md) ([RFC 7049](https://tools.ietf.org/html/rfc7049)) is a binary data format inspired by JSON and MessagePack.  CBOR is used in [IETF](https://www.ietf.org) Internet Standards such as COSE ([RFC 8152](https://tools.ietf.org/html/rfc8152)) and CWT ([RFC 8392 CBOR Web Token](https://tools.ietf.org/html/rfc8392)). WebAuthn also uses CBOR.

__Why this CBOR library?__ It doesn't crash and it has well-balanced qualities: small, fast, reliable and easy. 

* __Small__ and self-contained.  It has no external dependencies and no code gen. Programs like senmlCat (cisco/senml) are 4 MB smaller by switching to this library. Some programs can be smaller by 9+ MB.  See [size comparisons](#comparisons).

* __Fast__. v1.3 became faster than the most popular library.  Faster libraries will always exist, but speed is only one factor.  Choose this library if you value your time, program size, and system reliability. See [speed comparisons](#comparisons).

* __Reliable__ and safe. It prevents crashes on malicious CBOR data by using extensive tests, coverage-guided fuzzing, data validation, and avoiding Go's [`unsafe`](https://golang.org/pkg/unsafe/) package. It solely uses safe optimizations.

* __Easy__ and saves time.  It has the same API as [Go](https://golang.org)'s [`encoding/json`](https://golang.org/pkg/encoding/json/) when possible.  Existing structs don't require changes.  Go struct tags like `` `cbor:"name,omitempty"` `` and `` `json:"name,omitempty"` `` work as expected.

New struct tags like __`keyasint`__ and __`toarray`__ make CBOR, COSE, CWT, and SenML very easy to use.

Install with ```go get github.com/fxamacker/cbor``` and use it like Go's ```encoding/json```.

&nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; • [Design Goals](#design-goals) • [Comparisons](#comparisons)  • [Features](#features) • [Standards](#standards) • [Fuzzing](#fuzzing-and-code-coverage) • [Usage](#usage) • [Security Policy](#security-policy) •

## Current Status
Version 1.x has:

* __Stable API__ – won't make breaking API changes.  
* __Stable requirements__ – will always support Go v1.12.  
* __Passed fuzzing__ – v1.3.2 passed 4+ billion execs in coverage-guided fuzzing.

Each commit passes hundreds of unit tests. Each release also passes fuzz tests. See [Fuzzing and Code Coverage](#fuzzing-and-code-coverage).

Recent activity:

* [x] [Release v1.2](https://github.com/fxamacker/cbor/releases) -- add RawMessage type, Marshaler and Unmarshaler interfaces.
* [x] [Release v1.3](https://github.com/fxamacker/cbor/releases) -- faster encoding and decoding.
* [x] [Release v1.3](https://github.com/fxamacker/cbor/releases) -- add struct to/from CBOR array (`toarray` struct tag) for more compact data.
* [x] [Release v1.3](https://github.com/fxamacker/cbor/releases) -- add struct to/from CBOR map with int keys (`keyasint` struct tag). Simplifies using COSE, etc.
* [x] [Release v1.3.2](https://github.com/fxamacker/cbor/releases) -- (latest) prevent potential problems by adding checks after well-formedness.
* [ ] [Milestone v1.4](https://github.com/fxamacker/cbor/milestone/3) -- 🎈 add support for CBOR tags (major type 6.)

## Design Goals 
This CBOR library was created for my [WebAuthn (FIDO2) server library](https://github.com/fxamacker/webauthn), because existing CBOR libraries didn't meet certain criteria.  This library became a good fit for many other projects.

This library is designed to be:

* __Easy__ – idiomatic API like `encoding/json` with identical API when possible.
* __Small and self-contained__ – no external dependencies and no code gen.  Programs in cisco/senml are 4 MB smaller by switching to this library. In extreme cases programs can be smaller by 8+ MB. See [comparisons](#comparisons).
* __Safe and reliable__ – no `unsafe` pkg, coverage >95%, coverage-guided fuzzing, and data validation to avoid crashes on malformed or malicious data.

Competing factors are balanced:

* __Speed__ vs __safety__ vs __size__ – to keep size small, avoid code generation. For safety, validate data and avoid Go's `unsafe` pkg.  For speed, use safe optimizations: cache struct metadata, bypass reflect when appropriate, use sync.Pool to reuse transient objects, and etc. v1.3 is faster than the most popular `unsafe`-using codec library.
* __Standards compliance__ – CBOR ([RFC 7049](https://tools.ietf.org/html/rfc7049)) with minor [limitations](#limitations).  Encoding modes include default (no sorting), [RFC 7049 canonical](https://tools.ietf.org/html/rfc7049#section-3.9), and [CTAP2 canonical](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form). Decoding also checks for all required malformed data mentioned in latest RFC 7049bis.  See [Standards](#standards) section.

Initial releases focus on features, testing, and fuzzing.  After that, new releases (like v1.3) will also improve speed.

All releases prioritize reliability to avoid crashes on decoding malformed CBOR data. See [Fuzzing and Coverage](#fuzzing-and-code-coverage).

## Comparisons

Program size and speed comparisons are between this new library and the most popular codec.  This library prioritizes reliability so the speed comparison was a nice surprise.

__This library makes compiled programs smaller__. Programs like senmlCat can be 4 MB smaller by switching to this library.  Programs using more complex CBOR data types can be 9.2 MB smaller.

![alt text](https://user-images.githubusercontent.com/57072051/69993581-b98eeb00-1511-11ea-98da-46d14ad0fdd0.png "CBOR library and program size comparison chart")

__This library is faster__ for encoding and decoding CBOR Web Token (CWT claims).  WebAuthn has more impressive speed advantages but its CBOR data (on servers) is used during registration (not login).  So CWT speed matters more.

![alt text](https://user-images.githubusercontent.com/57072051/69992105-9c0c5200-150e-11ea-9888-9c50a0f41de2.png "CBOR library speed comparison chart")

The resource intensive `codec.CborHandle` initialization (in the other library) was placed outside the benchmark loop to make sure their library wasn't penalized.

Doing your own comparisons is highly recommended.  Use your most common message sizes and data types.

Additional comparisons may be added here from time to time.

## Features

* Idiomatic API like `encoding/json`.
* Support "cbor" and "json" keys in Go's struct tags. If both are specified, then "cbor" is used.
* Encode using smallest CBOR integer sizes for more compact data serialization.
* Decode slices, maps, and structs in-place.
* Decode into struct with field name case-insensitive match.
* Support canonical CBOR encoding for map/struct.
* Encode anonymous struct fields by `encoding/json` package struct fields visibility rules.
* Encode and decode nil slice/map/pointer/interface values correctly.
* Encode and decode time.Time as RFC 3339 formatted text string or Unix time.
* Encode and decode indefinite length bytes/string/array/map (["streaming"](https://tools.ietf.org/html/rfc7049#section-2.2)).
* v1.1 -- Support `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler` interfaces.
* v1.2 -- `cbor.RawMessage` can delay CBOR decoding or precompute CBOR encoding.
* v1.2 -- User-defined types can have custom CBOR encoding and decoding by implementing `cbor.Marshaler` and `cbor.Unmarshaler` interfaces. 
* v1.3 -- add struct to/from CBOR array (`toarray` struct tag) for more compact data
* v1.3 -- add struct to/from CBOR map with int keys (`keyasint` struct tag). Simplifies using COSE, etc.
* [Milestone v1.4](https://github.com/fxamacker/cbor/milestone/3) -- 🎈 add support for CBOR tags (major type 6.)

## Fuzzing and Code Coverage

__Hundreds of unit tests__ must pass before tagging a release.  They include all RFC 7049 examples, bugs found by fuzzing, 2 maliciously crafted CBOR data, and over 87 tests with malformed data based on RFC 7049bis.

__Code coverage__ must not fall below 95% when tagging a release.  Code coverage is 97.8% (`go test -cover`) for cbor v1.3 which is among the highest for libraries (in Go) of this type.

__Coverage-guided fuzzing__ must pass before tagging a release.  E.g. v1.3.2 was tagged when it reached 364.9 million execs and continued fuzzing (4+ billion execs) with [fxamacker/cbor-fuzz](https://github.com/fxamacker/cbor-fuzz).  Default corpus has:

* 2 files related to WebAuthn (FIDO U2F key).
* 3 files with custom struct.
* 9 files with [CWT examples (RFC 8392 Appendix A)](https://tools.ietf.org/html/rfc8392#appendix-A)
* 17 files with [COSE examples (RFC 8152 Appendix B & C)](https://github.com/cose-wg/Examples/tree/master/RFC8152).
* 81 files with [CBOR examples (RFC 7049 Appendix A) ](https://tools.ietf.org/html/rfc7049#appendix-A). It excludes 1 errata first reported in [issue #46](https://github.com/fxamacker/cbor/issues/46).

Over 1,000 files (corpus) are used for fuzzing because it includes fuzz-generated corpus.

## Standards
This library implements CBOR as specified in [RFC 7049](https://tools.ietf.org/html/rfc7049) with minor [limitations](#limitations).

Decoding also checks for all required well-formedness errors described in the latest RFC 7049bis, including all "subkinds" of syntax errors and too little data.

Encoding has 3 modes:

* default: no sorting, so it's the fastest mode.
* Canonical: [(RFC 7049 Section 3.9)](https://tools.ietf.org/html/rfc7049#section-3.9) uses length-first map key ordering.
* CTAP2Canonical: [(CTAP2 Canonical CBOR)](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form) uses bytewise lexicographic order for sorting keys.

CTAP2 Canonical CBOR encoding is used by [CTAP](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html) and [WebAuthn](https://www.w3.org/TR/webauthn/) in [FIDO2](https://fidoalliance.org/fido2/) framework.

All three encoding modes in this library use smallest form of CBOR integer that preserves data.  A new encoding mode will be added to do the same for floating point numbers in [milestone v1.4](https://github.com/fxamacker/cbor/milestone/3).

## Limitations
🎈 CBOR tags (type 6) is being added in the next release ([milestone v1.4](https://github.com/fxamacker/cbor/milestone/3)).

Current limitations:

* CBOR tags (type 6) are ignored.  Decoder simply decodes tagged data after ignoring the tags.
* CBOR negative int (type 1) that cannot fit into Go's int64 are not supported, such as RFC 7049 example -18446744073709551616.  Decoding these values returns `cbor.UnmarshalTypeError` like Go's `encoding/json`.
* CBOR `Undefined` (0xf7) value decodes to Go's `nil` value.  Use CBOR `Null` (0xf6) to round-trip with Go's `nil`.

Like Go's `encoding/json`, data validation checks the entire message to prevent partially filled (corrupted) data. This library also prevents crashes and resource exhaustion attacks from malicious CBOR data. Use Go's `io.LimitReader` when decoding very large data to limit size.

## System Requirements

* Go 1.12 (or newer)
* Tested and fuzzed on linux_amd64, but it should work on other platforms.

## Versions and API Changes
This project uses [Semantic Versioning](https://semver.org), so the API is always backwards compatible unless the major version number changes.

## API 
The API is the same as `encoding/json` when possible.

In addition to the API, the `keyasint` and `toarray` struct tags are worth knowing.  They can reduce programming effort, improve system performance, and reduce the size of serialized data.  

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
See [API docs](https://godoc.org/github.com/fxamacker/cbor) for more details.

## Installation
```
go get github.com/fxamacker/cbor
```
[Released versions](https://github.com/fxamacker/cbor/releases) benefit from longer fuzz tests.

## Usage
👉 Like Go's `encoding/json`, data validation checks the entire message to prevent partially filled (corrupted) data. This library also prevents crashes and resource exhaustion attacks from malicious CBOR data. Use Go's `io.LimitReader` when decoding very large data to limit size.

Like `encoding/json`:

* cbor.Marshal uses []byte
* cbor.Unmarshal uses []byte
* cbor.Encoder uses io.Writer
* cbor.Decoder uses io.Reader

The `keyasint` and `toarray` struct tags can reduce programming effort, improve system performance, and reduce the size of serialized data.

__Decoding CWT (CBOR Web Token)__ using `keyasint` and `toarray` struct tags:
```
// Signed CWT is defined in RFC 8392
type signedCWT struct {
	_           struct{} `cbor:",toarray"`
	Protected   []byte
	Unprotected coseHeader
	Payload     []byte
	Signature   []byte
}

// Part of COSE header definition
type coseHeader struct {
	Alg int    `cbor:"1,keyasint,omitempty"`
	Kid []byte `cbor:"4,keyasint,omitempty"`
	IV  []byte `cbor:"5,keyasint,omitempty"`
}

// data is []byte containing signed CWT

var v signedCWT
if err := cbor.Unmarshal(data, &v); err != nil {
	return err
}
```

__Encoding CWT (CBOR Web Token)__ using `keyasint` and `toarray` struct tags:
```
// Use signedCWT struct defined in "Decoding CWT" example.

var v signedCWT
...
if data, err := cbor.Marshal(v); err != nil {
	return err
}
```

__Decoding SenML__ using `keyasint` struct tag:
```
// RFC 8428 says, "The data is structured as a single array that 
// contains a series of SenML Records that can each contain fields"

type SenMLRecord struct {
	BaseName    string  `cbor:"-2,keyasint,omitempty"`
	BaseTime    float64 `cbor:"-3,keyasint,omitempty"`
	BaseUnit    string  `cbor:"-4,keyasint,omitempty"`
	BaseValue   float64 `cbor:"-5,keyasint,omitempty"`
	BaseSum     float64 `cbor:"-6,keyasint,omitempty"`
	BaseVersion int     `cbor:"-1,keyasint,omitempty"`
	Name        string  `cbor:"0,keyasint,omitempty"`
	Unit        string  `cbor:"1,keyasint,omitempty"`
	Value       float64 `cbor:"2,keyasint,omitempty"`
	ValueS      string  `cbor:"3,keyasint,omitempty"`
	ValueB      bool    `cbor:"4,keyasint,omitempty"`
	ValueD      string  `cbor:"8,keyasint,omitempty"`
	Sum         float64 `cbor:"5,keyasint,omitempty"`
	Time        float64 `cbor:"6,keyasint,omitempty"`
	UpdateTime  float64 `cbor:"7,keyasint,omitempty"`
}

// data is a []byte containing SenML

var v []SenMLRecord
if err := cbor.Unmarshal(data, &v); err != nil {
	return err
}
```

__Encoding SenML__ using `keyasint` struct tag:
```
// use SenMLRecord struct defined in "Decoding SenML" example

var v []SenMLRecord
...
if data, err := cbor.Marshal(v); err != nil {
	return err
}
```

__Decoding__:

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

__Encoding__:

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

__Encoding indefinite length array__:

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

More [examples](example_test.go).

## Benchmarks

Go structs are faster than maps with string keys:
* decoding into struct is >31% faster than decoding into map.
* encoding struct is >33% faster than encoding map.

Go structs with `keyasint` struct tag are faster than maps with integer keys:
* decoding into struct is >25% faster than decoding into map.
* encoding struct is >31% faster than encoding map.

Go structs with `toarray` struct tag are faster than slice:
* decoding into struct is >15% faster than decoding into slice.
* encoding struct is >10% faster than encoding slice.

Doing your own benchmarks is highly recommended.  Use your most common message sizes and data types.

See [Benchmarks for fxamacker/cbor](BENCHMARKS.md).

## Code of Conduct 
This project has adopted the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).  Contact [faye.github@gmail.com](mailto:faye.github@gmail.com) with any questions or comments.

## Contributing
Please refer to [How to Contribute](CONTRIBUTING.md).

## Security Policy
For v1, security fixes are provided only for the latest released version since the API won't break compatibility.

To report security vulnerabilities, please email [faye.github@gmail.com](mailto:faye.github@gmail.com) and allow time for the problem to be resolved before reporting it to the public.

## Disclaimers
Phrases like "no crashes" or "doesn't crash" mean there are no known crash bugs in the latest version based on results of unit tests and coverage-guided fuzzing.  It doesn't imply the software is 100% bug-free or 100% invulnerable to all known and unknown attacks.

Please read the license for additional disclaimers and terms.

## License 
Copyright (c) 2019 [Faye Amacker](https://github.com/fxamacker)

Licensed under [MIT License](LICENSE)

<hr>

&nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; • [Design Goals](#design-goals) • [Comparisons](#comparisons)  • [Features](#features) • [Standards](#standards) • [Fuzzing](#fuzzing-and-code-coverage) • [Usage](#usage) • [Security Policy](#security-policy) •
