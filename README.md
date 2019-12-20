[![CBOR Library - Slideshow and Latest Docs.](https://github.com/x448/images/raw/master/cbor/cbor_slides.gif)](https://github.com/fxamacker/cbor/blob/master/README.md)

# CBOR library in Go
This library is a CBOR encoder and decoder.  There are many CBOR libraries.  Choose this CBOR library if you value your time, program size, and system reliability.

[![Build Status](https://travis-ci.com/fxamacker/cbor.svg?branch=master)](https://travis-ci.com/fxamacker/cbor)
[![codecov](https://codecov.io/gh/fxamacker/cbor/branch/master/graph/badge.svg?v=4)](https://codecov.io/gh/fxamacker/cbor)
[![Go Report Card](https://goreportcard.com/badge/github.com/fxamacker/cbor)](https://goreportcard.com/report/github.com/fxamacker/cbor)
[![Release](https://img.shields.io/github/release/fxamacker/cbor.svg?style=flat-square)](https://github.com/fxamacker/cbor/releases)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/fxamacker/cbor/master/LICENSE)

__What is CBOR__?  [CBOR](CBOR_GOLANG.md) ([RFC 7049](https://tools.ietf.org/html/rfc7049)) is a binary data format inspired by JSON and MessagePack.  CBOR is used in [IETF](https://www.ietf.org) Internet Standards such as COSE ([RFC 8152](https://tools.ietf.org/html/rfc8152)) and CWT ([RFC 8392 CBOR Web Token](https://tools.ietf.org/html/rfc8392)). Even WebAuthn uses CBOR.

__Why this CBOR library?__ It doesn't crash and it has well-balanced qualities: small, fast, safe and easy. 

* __Small__ and self-contained.  Same programs are 4-9 MB smaller by switching to this library.  There are no external dependencies and no code gen.   

* __Fast__. v1.3 became faster than a well-known library that uses `unsafe` optimizations and code gen.  Faster libraries will always exist, but speed is only one factor.  This library doesn't use `unsafe` optimizations or code gen.  

* __Safe__ and reliable. It prevents crashes on malicious CBOR data by using extensive tests, coverage-guided fuzzing, data validation, and avoiding Go's [`unsafe`](https://golang.org/pkg/unsafe/) package. 

* __Easy__ and saves time.  It has the same API as [Go](https://golang.org)'s [`encoding/json`](https://golang.org/pkg/encoding/json/) when possible.  Existing structs don't require changes.  Go struct tags like `` `cbor:"name,omitempty"` `` and `` `json:"name,omitempty"` `` work as expected.

New struct tags like __`keyasint`__ and __`toarray`__ make compact CBOR data like COSE, CWT, and SenML easier to use.

<hr>

[![CBOR API](https://github.com/fxamacker/images/raw/master/cbor/cbor_easy_api.png)](#usage)

<hr>

👉  [Comparisons](#comparisons) • [Status](#current-status) • [Design Goals](#design-goals) • [Features](#features) • [Standards](#standards) • [Fuzzing](#fuzzing-and-code-coverage) • [Usage](#usage) • [Security Policy](#security-policy) • [License](#license)

## Comparisons

Comparisons are between this newer library and a well-known library that had 1,000+ stars before this library was created.  Default build settings for each library were used for all comparisons.

__This library is safer__.  Small malicious CBOR messages are rejected quickly before they exhaust system resources.

![alt text](https://github.com/fxamacker/images/raw/master/cbor/v1.3.4/cbor_safety_comparison.png "CBOR library safety comparison")

__This library is smaller__. Programs like senmlCat can be 4 MB smaller by switching to this library.  Programs using more complex CBOR data types can be 9.2 MB smaller.

![alt text](https://github.com/fxamacker/images/raw/master/cbor/v1.3.4/cbor_size_comparison.png "CBOR library and program size comparison chart")

__This library is faster__ for encoding and decoding CBOR Web Token (CWT).  However, speed is only one factor and it can vary depending on data types and sizes.  Unlike the other library, this one doesn't use Go's ```unsafe``` package or code gen.

![alt text](https://github.com/fxamacker/images/raw/master/cbor/v1.3.4/cbor_speed_comparison.png "CBOR library speed comparison chart")

The resource intensive `codec.CborHandle` initialization (in the other library) was placed outside the benchmark loop to make sure their library wasn't penalized.

Doing your own comparisons is highly recommended.  Use your most common message sizes and data types.

## Current Status
Version 1.x has:

* __Stable API__ – won't make breaking API changes.  
* __Stable requirements__ – will always support Go v1.12 (unless there's compelling reason).
* __Passed fuzzing__ – v1.3.4 passed 370+ million execs in coverage-guided fuzzing when it was released.

Each commit passes 145+ unit tests (plus many more subtests). Each release also passes 250+ million execs in coverage-guided fuzzing using 1,000+ CBOR files (corpus). See [Fuzzing and Code Coverage](#fuzzing-and-code-coverage).

Recent activity:

* [x] [Release v1.2](https://github.com/fxamacker/cbor/releases) -- add RawMessage type, Marshaler and Unmarshaler interfaces.
* [x] [Release v1.3](https://github.com/fxamacker/cbor/releases) -- faster encoding and decoding.
* [x] [Release v1.3](https://github.com/fxamacker/cbor/releases) -- add `toarray` struct tag to simplify using CBOR arrays.
* [x] [Release v1.3](https://github.com/fxamacker/cbor/releases) -- add `keyasint` struct tag to simplify using CBOR maps with int keys.
* [x] [Release v1.3.4](https://github.com/fxamacker/cbor/releases) -- (latest) bugfixes and refactoring.  Limit nested levels to 32 for arrays, maps, tags.

Coming soon: support for CBOR tags (major type 6) and new encoding option to shrink floats. After that, options for handling duplicate map keys.

## Design Goals 
This library is designed to be a generic CBOR encoder and decoder.  It was initially created for a [WebAuthn (FIDO2) server library](https://github.com/fxamacker/webauthn), because existing CBOR libraries (in Go) didn't meet certain criteria in 2019.

This library is designed to be:

* __Easy__ – API is like `encoding/json` plus `keyasint` and `toarray` struct tags.
* __Small and self-contained__ – no external dependencies and no code gen.  Programs in cisco/senml are 4 MB smaller by switching to this library. In extreme cases programs can be smaller by 9+ MB. See [comparisons](#comparisons).
* __Safe and reliable__ – no `unsafe` pkg, coverage >95%, coverage-guided fuzzing, and data validation to avoid crashes on malformed or malicious data.  See [comparisons](#comparisons).

Competing factors are balanced:

* __Speed__ vs __safety__ vs __size__ – to keep size small, avoid code generation. For safety, validate data and avoid Go's `unsafe` pkg.  For speed, use safe optimizations such as caching struct metadata. v1.3 is faster than a well-known library that uses `unsafe` and code gen.
* __Standards compliance__ – CBOR ([RFC 7049](https://tools.ietf.org/html/rfc7049)) with minor [limitations](#limitations).  Encoding modes include default (no sorting), [RFC 7049 canonical](https://tools.ietf.org/html/rfc7049#section-3.9), and [CTAP2 canonical](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form). Decoding also checks for all required malformed data mentioned in latest RFC 7049bis.  See [Standards](#standards) section.

All releases prioritize reliability to avoid crashes on decoding malformed CBOR data. See [Fuzzing and Coverage](#fuzzing-and-code-coverage).

Features not in Go's standard library are usually not added.  However, the __`toarray`__ struct tag in ugorji/go was too useful to ignore. It was added in v1.3 when a project mentioned they were using it with CBOR to save disk space.

## Features

* API is like `encoding/json` plus extra struct tags:
  * `cbor.Encoder` writes CBOR to `io.Writer`.  
  * `cbor.Decoder` reads CBOR from `io.Reader`.
  * `cbor.Marshal` writes CBOR to `[]byte`.  
  * `cbor.Unmarshal` reads CBOR from `[]byte`.  
* Support "cbor" and "json" keys in Go's struct tags. If both are specified, then "cbor" is used.
* `toarray` struct tag allows named struct fields for elements of CBOR arrays.
* `keyasint` struct tag allows named struct fields for elements of CBOR maps with int keys.
* Support 3 encoding modes: default (unsorted), Canonical, CTAP2Canonical.
* Support `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler` interfaces.
* Support `cbor.RawMessage` which can delay CBOR decoding or precompute CBOR encoding.
* Support `cbor.Marshaler` and `cbor.Unmarshaler` interfaces to allow user-defined types to have custom CBOR encoding and decoding.
* Support `time.Time` as RFC 3339 formatted text string or Unix time.
* Support indefinite length CBOR data (["streaming"](https://tools.ietf.org/html/rfc7049#section-2.2)).  For decoding, each indefinite length "container" must fit into memory to perform well-formedness checks that prevent exploits. Go's `io.LimitReader` can be used to limit sizes.
* Encoder uses struct field visibility rules for anonymous struct fields (same rules as `encoding/json`.)
* Encoder always uses smallest CBOR integer sizes for more compact data serialization.
* Decoder always checks for invalid UTF-8 string errors.
* Decoder always decodes in-place to slices, maps, and structs.
* Decoder uses case-insensitive field name match when decoding to structs. 
* Both encoder and decoder correctly handles nil slice, map, pointer, and interface values.

Coming soon: support for CBOR tags (major type 6) and new encoding option to shrink floats.  After that, options for handling duplicate map keys.

## Standards
This library implements CBOR as specified in [RFC 7049](https://tools.ietf.org/html/rfc7049) with minor [limitations](#limitations).

Encoder has 3 modes:

* default: no sorting, so it's the fastest mode.
* Canonical: [(RFC 7049 Section 3.9)](https://tools.ietf.org/html/rfc7049#section-3.9) uses length-first map key ordering.
* CTAP2Canonical: [(CTAP2 Canonical CBOR)](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form) uses bytewise lexicographic order for sorting keys.

CTAP2 Canonical CBOR encoding is used by [CTAP](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html) and [WebAuthn](https://www.w3.org/TR/webauthn/) in [FIDO2](https://fidoalliance.org/fido2/) framework.

All three encoding modes in this library use smallest form of CBOR integer that preserves data.  New encoding mode(s) will be added to do the same for floating point numbers in [milestone v2.0](https://github.com/fxamacker/cbor/milestone/3).

Decoder checks for all required well-formedness errors described in the latest RFC 7049bis, including all "subkinds" of syntax errors and too little data.

After well-formedness is verified, basic validity errors are handled as follows:

- Invalid UTF-8 string: Decoder always checks and returns invalid UTF-8 string error.
- Duplicate keys in a map: By default, decoder decodes to a map with duplicate keys by overwriting previous value with the same key.  Options to handle duplicate map keys in different ways may be added as a feature.

When decoding well-formed CBOR arrays and maps, decoder saves the first error it encounters and continues with the next item.  Options to handle this differently may be added in the future.

## Limitations
CBOR tags (type 6) is being added in the next release ([milestone v2.0](https://github.com/fxamacker/cbor/milestone/3)) and is coming soon.

Current limitations:

* CBOR tag numbers are ignored.  Decoder simply decodes tag content.
* CBOR negative int (type 1) that cannot fit into Go's int64 are not supported, such as RFC 7049 example -18446744073709551616.  Decoding these values returns `cbor.UnmarshalTypeError` like Go's `encoding/json`.
* CBOR `Undefined` (0xf7) value decodes to Go's `nil` value.  Use CBOR `Null` (0xf6) to round-trip with Go's `nil`.
* Duplicate map keys are not checked during decoding.  Option to handle duplicate map keys in different ways may be added as a feature.
* Nested levels for CBOR arrays, maps, and tags are limited to 32 to prevent exploits.  If you need a larger limit, please open an issue describing your data format and this limit will be reconsidered.

Like Go's `encoding/json`, data validation checks the entire message to prevent partially filled (corrupted) data. This library also prevents crashes and resource exhaustion attacks from malicious CBOR data. Use Go's `io.LimitReader` when decoding very large data to limit size.

## Fuzzing and Code Coverage

__Over 145 unit tests (plus many more subtests)__ must pass before tagging a release.  They include all RFC 7049 examples, bugs found by fuzzing, 2 maliciously crafted CBOR data, and over 87 tests with malformed data based on RFC 7049bis.

__Code coverage__ must not fall below 95% when tagging a release.  Code coverage is 97.8% (`go test -cover`) for cbor v1.3 which is among the highest for libraries (in Go) of this type.

__Coverage-guided fuzzing__ must pass 250+ million execs before tagging a release.  E.g. v1.3.2 was tagged when it reached 364.9 million execs and continued fuzzing (4+ billion execs) with [fxamacker/cbor-fuzz](https://github.com/fxamacker/cbor-fuzz).  Default corpus has:

* 2 files related to WebAuthn (FIDO U2F key).
* 3 files with custom struct.
* 9 files with [CWT examples (RFC 8392 Appendix A)](https://tools.ietf.org/html/rfc8392#appendix-A)
* 17 files with [COSE examples (RFC 8152 Appendix B & C)](https://github.com/cose-wg/Examples/tree/master/RFC8152).
* 81 files with [CBOR examples (RFC 7049 Appendix A) ](https://tools.ietf.org/html/rfc7049#appendix-A). It excludes 1 errata first reported in [issue #46](https://github.com/fxamacker/cbor/issues/46).

Over 1,000 files (corpus) are used for fuzzing because it includes fuzz-generated corpus.

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
👉 Use Go's `io.LimitReader` when decoding very large data to limit size.

The API is the same as `encoding/json` when possible:

* cbor.Marshal writes CBOR to []byte
* cbor.Unmarshal reads CBOR from []byte
* cbor.Encoder writes CBOR to io.Writer
* cbor.Decoder reads CBOR from io.Reader

The `keyasint` and `toarray` struct tags make it easy to use compact CBOR message formats.  Internet standards often use CBOR arrays and CBOR maps with int keys to save space.

Using named struct fields instead of array elements or maps with int keys makes code more readable and less error prone.

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

## Special Thanks
* Carsten Bormann for RFC 7049 (CBOR), his fast confirmation to my RFC 7049 errata, approving my pull request to 7049bis, and his patience when I misread a line in 7049bis.
* Montgomery Edwards⁴⁴⁸ for updating the README.md, creating comparison charts, and filing many helpful issues.
* Stefan Tatschner for being the 1st to discover my CBOR library, filing issues #1 and #2, and recommending this library.
* Yawning Angel for replacing a library with this one in a big project after an external security audit, and filing issue #5.
* Jernej Kos for filing issue #11 (add feature similar to json.RawMessage) and his kind words about this library.
* Jeffrey Yasskin and Laurence Lundblade for their help clarifying 7049bis on the IETF mailing list.
* Jakob Borg for his words of encouragement about this library at Go Forum.

## License 
Copyright (c) 2019 [Faye Amacker](https://github.com/fxamacker)

Licensed under [MIT License](LICENSE)

<hr>

👉  [Comparisons](#comparisons) • [Status](#current-status) • [Design Goals](#design-goals) • [Features](#features) • [Standards](#standards) • [Fuzzing](#fuzzing-and-code-coverage) • [Usage](#usage) • [Security Policy](#security-policy) • [License](#license)
