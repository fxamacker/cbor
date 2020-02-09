[![CBOR Library - Slideshow and Latest Docs.](https://github.com/fxamacker/images/raw/master/cbor/v2.0.0/cbor_slides.gif)](https://github.com/fxamacker/cbor/blob/master/README.md)

# CBOR library in Go
[__`fxamacker/cbor`__](https://github.com/fxamacker/cbor) is a CBOR encoder and decoder.  CBOR codec functions have identical signatures to `json/encoding`.  Each release passes 375+ tests and 250+ million execs fuzzing.

[![](https://github.com/fxamacker/cbor/workflows/ci/badge.svg)](https://github.com/fxamacker/cbor/actions?query=workflow%3Aci)
[![](https://github.com/fxamacker/cbor/workflows/cover%20%E2%89%A597%25/badge.svg)](https://github.com/fxamacker/cbor/actions?query=workflow%3A%22cover+%E2%89%A597%25%22)
[![](https://github.com/fxamacker/cbor/workflows/linters/badge.svg)](https://github.com/fxamacker/cbor/actions?query=workflow%3Alinters)
[![Go Report Card](https://goreportcard.com/badge/github.com/fxamacker/cbor)](https://goreportcard.com/report/github.com/fxamacker/cbor)
[![Release](https://img.shields.io/github/release/fxamacker/cbor.svg?style=flat-square)](https://github.com/fxamacker/cbor/releases)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/fxamacker/cbor/master/LICENSE)

__What is CBOR__?  [CBOR](CBOR_GOLANG.md) ([RFC 7049](https://tools.ietf.org/html/rfc7049)) is a binary data format inspired by JSON and MessagePack.  CBOR is used in [IETF](https://www.ietf.org) Internet Standards such as COSE ([RFC 8152](https://tools.ietf.org/html/rfc8152)) and CWT ([RFC 8392 CBOR Web Token](https://tools.ietf.org/html/rfc8392)). WebAuthn also uses CBOR.

__Why this CBOR library?__ It doesn't crash and it has well-balanced qualities: small, fast, safe and easy. It also supports "preferred serialization" by encoding integers and floats to their smallest forms when values fit.

* __Small apps__.  Same programs are 4-9 MB smaller by switching to this library.  No code gen and the only imported pkg is [x448/float16](https://github.com/x448/float16) which is maintained by the same team as this library.

* __Small data__.  The `toarray`, `keyasint`, and `omitempty` struct tags shrink size of Go structs encoded to CBOR.  Integers encode to smallest form that fits.  Floats can shrink from float64 -> float32 -> float16 if values can round-trip.

* __Fast__. v1.3 became faster than a well-known library that uses `unsafe` optimizations and code gen.  Faster libraries will always exist, but speed is only one factor.  This library doesn't use `unsafe` optimizations or code gen.  

* __Safe__ and reliable. It prevents crashes on malicious CBOR data by using extensive tests, coverage-guided fuzzing, data validation, and avoiding Go's [`unsafe`](https://golang.org/pkg/unsafe/) pkg. Nested levels for CBOR arrays, maps, and tags are limited to 32.

* __Easy__ and saves time.  It has the same API as [Go](https://golang.org)'s [`encoding/json`](https://golang.org/pkg/encoding/json/) when possible.  Existing structs don't require changes.  Go struct tags like `` `cbor:"name,omitempty"` `` and `` `json:"name,omitempty"` `` work as expected.  

* __Stable API__. Six codec functions have signatures identical to encoding/json and they will never change:  
`Marshal`, `Unmarshal`, `NewEncoder`, `NewDecoder`, `encoder.Encode`, and `decoder.Decode`.

Codec functions at package-level use default options. Same codec functions are exported by EncMode and DecMode.  EncMode and DecMode are interfaces created from EncOptions or DecOptions structs.

Predefined options make it easier to comply with standards like Canonical CBOR, CTAP2 Canonical CBOR, etc.

Customized options are easy.  For example, EncOptions.Time can be set to:  
`TimeUnix`, `TimeUnixMicro`, `TimeUnixDynamic`, `TimeRFC3339`, or `TimeRFC3339Nano`.

Struct tags like __`keyasint`__ and __`toarray`__ make compact CBOR data such as COSE, CWT, and SenML easier to use.

<hr>

[![CBOR API](https://github.com/fxamacker/images/raw/master/cbor/v2.0.0/cbor_easy_api.png)](#usage)

<hr>

👉  [Status](#current-status) • [Design Goals](#design-goals) • [Features](#features) • [Standards](#standards) • [Install](#installation) • [API](#api) • [Usage](#usage) • [Fuzzing](#fuzzing-and-code-coverage) • [Security Policy](#security-policy) • [License](#license)

## Comparisons

Comparisons are between this newer library and a well-known library that had 1,000+ stars before this library was created.  Default build settings for each library were used for all comparisons.

__This library is safer__.  Small malicious CBOR messages are rejected quickly before they exhaust system resources.

![alt text](https://github.com/fxamacker/images/raw/master/cbor/v2.0.0/cbor_safety_comparison.png "CBOR library safety comparison")

__This library is smaller__. Programs like senmlCat can be 4 MB smaller by switching to this library.  Programs using more complex CBOR data types can be 9.2 MB smaller.

![alt text](https://github.com/fxamacker/images/raw/master/cbor/v2.0.0/cbor_size_comparison.png "CBOR library and program size comparison chart")

__This library is faster__ for encoding and decoding CBOR Web Token (CWT).  However, speed is only one factor and it can vary depending on data types and sizes.  Unlike the other library, this one doesn't use Go's ```unsafe``` package or code gen.

![alt text](https://github.com/fxamacker/images/raw/master/cbor/v2.0.0/cbor_speed_comparison.png "CBOR library speed comparison chart")

The resource intensive `codec.CborHandle` initialization (in the other library) was placed outside the benchmark loop to make sure their library wasn't penalized.

Doing your own comparisons is highly recommended.  Use your most common message sizes and data types.

## Current Status
Latest version is v2.0, which has:

* __Stable API__ –  Six codec function signatures will never change.  No breaking API changes for other funcs in same major version.    These two functions are subject to change until the draft RFC is approved by IETF (est. "later in 2020"):
  * CoreDetEncOptions() is subject to change because it uses draft standard not yet approved by IETF.
  * PreferredUnsortedEncOptions() is subject to change because it uses draft standard not yet approved by IETF.
* __Passed all tests__ – v2.0 passed all 375+ tests on amd64, arm64, ppc64le and s390x with linux.
* __Passed fuzzing__ – v2.0 passed 1+ billion execs in coverage-guided fuzzing on Feb. 2, 2020.

Recent activity:

* [x] [Release v1.5](https://github.com/fxamacker/cbor/releases) -- Add option to shrink floating-point values to smaller sizes like float16 (if they preserve value).
* [x] [Release v1.5](https://github.com/fxamacker/cbor/releases) -- Add options for encoding floating-point NaN values: NaNConvertNone, NaNConvert7e00, NaNConvertQuiet, or NaNConvertPreserveSignal.
* [x] [Release v2.0](https://github.com/fxamacker/cbor/releases) -- Improved API, faster performance, less memory allocs, five options for encoding time and cleaner code. Decode NaN or Infinity time values as if they were NULL (to Go's "zero time".)

__Why v2.0?__:

v1 required breaking API changes to support upcoming new features like CBOR tags.  v2.0 has other benefits like reduced memory and including more functions with signatures identical to `encoding/json`. See [Design Goals](#design-goals) for more.

__Roadmap__:

 * v2.1 (est. Feb. 9-12, 2020) support for CBOR tags (major type 6) and some decoder optimizations.
 * v2.2 (Feb. 2020) options for handling duplicate map keys.

## Installation
```
$ GO111MODULE=on go get github.com/fxamacker/cbor/v2
```

```
import (
	"github.com/fxamacker/cbor/v2" // imports as package "cbor"
)
```

[Released versions](https://github.com/fxamacker/cbor/releases) benefit from longer fuzz tests.

## System Requirements

* Go 1.12 (or newer)
* amd64, arm64, ppc64le and s390x. Other architectures may also work but they are not tested as frequently. 

<hr>

👉  [Status](#current-status) • [Design Goals](#design-goals) • [Features](#features) • [Standards](#standards) • [Install](#installation) • [API](#api) • [Usage](#usage) • [Fuzzing](#fuzzing-and-code-coverage) • [Security Policy](#security-policy) • [License](#license)

## Design Goals 
This library is designed to be a generic CBOR encoder and decoder.  It was initially created for a [WebAuthn (FIDO2) server library](https://github.com/fxamacker/webauthn), because existing CBOR libraries (in Go) didn't meet certain criteria in 2019.

This library is designed to be:

* __Easy__ – API is like `encoding/json` plus `keyasint` and `toarray` struct tags.
* __Small__ – Programs in cisco/senml are 4 MB smaller by switching to this library. In extreme cases programs can be smaller by 9+ MB. No code gen and the only imported pkg is x448/float16 which is maintained by the same team.
* __Safe and reliable__ – No `unsafe` pkg, coverage >95%, coverage-guided fuzzing, and data validation to avoid crashes on malformed or malicious data.

Competing factors are balanced:

* __Speed__ vs __safety__ vs __size__ – to keep size small, avoid code generation. For safety, validate data and avoid Go's `unsafe` pkg.  For speed, use safe optimizations such as caching struct metadata. v1.4 is faster than a well-known library that uses `unsafe` and code gen.
* __Standards compliance__ – CBOR ([RFC 7049](https://tools.ietf.org/html/rfc7049)) with minor [limitations](#limitations).  Encoder supports options for sorting, floating-point conversions, and more.  Predefined configurations are also available so you can use "CTAP2 Canonical CBOR", etc. without knowing individual options.  Decoder checks for well-formedness, validates data, and limits nested levels to defend against attacks.  See [Standards](#standards).

v2.0 decoupled options from CBOR encoding & decoding functions:

* More encoding/decoding function signatures are identical to encoding/json.
* More function signatures can remain stable forever.
* More flexibility for evolving internal data types, optimizations, and concurrency.
* Features like CBOR tags can be added without more breaking API changes.
* Options to handle duplicate map keys can be added without more breaking API changes.

Avoiding `unsafe` package has benefits.  The `unsafe` package [warns](https://golang.org/pkg/unsafe/):

> Packages that import unsafe may be non-portable and are not protected by the Go 1 compatibility guidelines.

All releases prioritize reliability to avoid crashes on decoding malformed CBOR data. See [Fuzzing and Coverage](#fuzzing-and-code-coverage).

Features not in Go's standard library are usually not added.  However, the __`toarray`__ struct tag in __ugorji/go__ was too useful to ignore. It was added in v1.3 when a project mentioned they were using it with CBOR to save disk space.

## Features

* In v2, many function signatures are identical to encoding/json, including:  
`Marshal`, `Unmarshal`, `NewEncoder`, `NewDecoder`, `encoder.Encode`, `decoder.Decode`.
* Support customizable options to create encoding and decoding modes (interfaces) that are safe for concurrent use.  Once created, encoding and decoding modes cannot accidentally change their behavior at runtime.
* Support "cbor" and "json" keys in Go's struct tags. If both are specified, then "cbor" is used.
* `toarray` struct tag allows named struct fields for elements of CBOR arrays.
* `keyasint` struct tag allows named struct fields for elements of CBOR maps with int keys.
* Encoder has easy functions that return predefined configurations:  
  `CanonicalEncOptions`, `CTAP2EncOptions`, `CoreDetEncOptions`, `PreferredUnsortedEncOptions`
* Integers always encode to the smallest number of bytes.
* Encoder floating-point options:
  * ShortestFloatMode: `ShortestFloatNone` or `ShortestFloat16` (IEEE 754 binary16, etc. if value fits).
  * InfConvertMode: `InfConvertNone` or `InfConvertFloat16`.
  * NaNConvertMode: `NaNConvertNone`, `NaNConvert7e00`, `NaNConvertQuiet`, or `NaNConvertPreserveSignal`.
* Encoder sort options: `SortNone`, `SortBytewiseLexical`, `SortCanonical`, `SortCTAP2`, `SortCoreDeterministic`.
* Encoder time options: `TimeUnix`, `TimeUnixMicro`, `TimeUnixDynamic`, `TimeRFC3339`, `TimeRFC3339Nano`.
* Support `cbor.RawMessage` which can delay CBOR decoding or precompute CBOR encoding.
* Support Go's `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler` interfaces.
* Support `cbor.Marshaler` and `cbor.Unmarshaler` interfaces to allow user-defined types to have custom CBOR encoding and decoding.
* Support indefinite length CBOR data (["streaming"](https://tools.ietf.org/html/rfc7049#section-2.2)).  For decoding, each indefinite length "container" must fit into memory to perform well-formedness checks that prevent exploits. Go's `io.LimitReader` can be used to limit sizes.
* Decoder always checks for invalid UTF-8 string errors.
* Decoder always decodes in-place to slices, maps, and structs.
* Decoder uses case-insensitive field name match when decoding to structs. 
* Both encoder and decoder correctly handles nil slice, map, pointer, and interface values.

See [milestones](https://github.com/fxamacker/cbor/milestones) for upcoming features.

<hr>

👉  [Status](#current-status) • [Design Goals](#design-goals) • [Features](#features) • [Standards](#standards) • [Install](#installation) • [API](#api) • [Usage](#usage) • [Fuzzing](#fuzzing-and-code-coverage) • [Security Policy](#security-policy) • [License](#license)

## Standards
This library implements CBOR as specified in [RFC 7049](https://tools.ietf.org/html/rfc7049) with minor [limitations](#limitations).

Integers always encode to the shortest form that preserves value.  By default, time values are encoded without tags.

Encoding of other data types and map key sort order are determined by encoder options.

|Option                       |Available Settings (defaults in bold, aliases in italics)                                |
|-----------------------------|---------------------------------------------------------------------------------------|
|EncOptions.Sort |__`SortNone`__, `SortLengthFirst`, `SortBytewiseLexical`, _`SortCanonical`_, _`SortCTAP2Canonical`_, _`SortCoreDeterministic`_ |
|EncOptions.Time |__`TimeUnix`__, `TimeUnixMicro`, `TimeUnixDynamic`, `TimeRFC3339`, `TimeRFC3339Nano` |
|EncOptions.ShortestFloat |__`ShortestFloatNone`__, `ShortestFloat16` |
|EncOptions.InfConvert |__`InfConvertFloat16`__, `InfConvertNone` |
|EncOptions.NaNConvert |__`NaNConvert7e00`__, `NaNConvertNone`, `NaNConvertQuiet`, `NaNConvertPreserveSignal` |

See the end of [API section](#api) for more info about each encoding option.

Float16 conversions use [x448/float16](https://github.com/x448/float16) maintained by the same team as this library.  All 4+ billion possible conversions are verified to be correct in that library on amd64, arm64, ppc64le and s390x.

Go nil values for slices, maps, pointers, etc. are encoded as CBOR null.  Empty slices, maps, etc. are encoded as empty CBOR arrays and maps.

Decoder checks for all required well-formedness errors, including all "subkinds" of syntax errors and too little data.

After well-formedness is verified, basic validity errors are handled as follows:

* Invalid UTF-8 string: Decoder always checks and returns invalid UTF-8 string error.
* Duplicate keys in a map: By default, decoder decodes to a map with duplicate keys by overwriting previous value with the same key.  Options to handle duplicate map keys in different ways may be added as a feature.

When decoding well-formed CBOR arrays and maps, decoder saves the first error it encounters and continues with the next item.  Options to handle this differently may be added in the future.

## Limitations
CBOR tags (type 6) is being added in the next release ([milestone v2.1](https://github.com/fxamacker/cbor/milestone/11)) and is coming soon.

Known limitations:

* Currently, CBOR tag numbers are ignored.  Decoder simply decodes tag content. Work is in progress to add support.
* Currently, duplicate map keys are not checked during decoding.  Option to handle duplicate map keys in different ways will be added.
* Nested levels for CBOR arrays, maps, and tags are limited to 32 to quickly reject potentially malicious data.  This limit will be reconsidered upon request.
* CBOR negative int (type 1) that cannot fit into Go's int64 are not supported, such as RFC 7049 example -18446744073709551616.  Decoding these values returns `cbor.UnmarshalTypeError` like Go's `encoding/json`.
* CBOR `Undefined` (0xf7) value decodes to Go's `nil` value.  CBOR `Null` (0xf6) more closely matches Go's `nil`.

Like Go's `encoding/json`, data validation checks the entire message to prevent partially filled (corrupted) data. This library also prevents crashes and resource exhaustion attacks from malicious CBOR data. Use Go's `io.LimitReader` when decoding very large data to limit size.

## API 
In v2, many function signatures are identical to encoding/json, including:  
`Marshal`, `Unmarshal`, `NewEncoder`, `NewDecoder`, `encoder.Encode`, `decoder.Decode`.

"Mode" in this API means definite way of encoding or decoding.

EncMode and DecMode are interfaces created from EncOptions or DecOptions structs.  
For example, `em := cbor.EncOptions{...}.EncMode()` or `em := cbor.CanonicalEncOptions().EncMode()`.

EncMode and DecMode use immutable options so their behavior won't accidentally change at runtime.  Modes are intended to be reused and are safe for concurrent use.

__API for Default Mode of Encoding & Decoding__

If default options are acceptable, then you don't need to create EncMode or DecMode.
```
Marshal(v interface{}) ([]byte, error)
NewEncoder(w io.Writer) *Encoder

Unmarshal(data []byte, v interface{}) error
NewDecoder(r io.Reader) *Decoder
```

__API for Creating & Using Encoding Modes__
```
// EncMode interface uses immutable options and is safe for concurrent use.
type EncMode interface {
	Marshal(v interface{}) ([]byte, error)
	NewEncoder(w io.Writer) *Encoder
	EncOptions() EncOptions  // returns copy of options
}

// EncOptions specifies encoding options.
type EncOptions struct {
...
}

// EncMode returns an EncMode interface created from EncOptions.
func (opts EncOptions) EncMode() (EncMode, error) 
```

__API for Predefined Encoding Options__
```
func CanonicalEncOptions() EncOptions     // settings for RFC 7049 Canonical CBOR
func CTAP2EncOptions() EncOptions         // settings for FIDO2 CTAP2 Canonical CBOR
func CoreDetEncOptions() EncOptions       // modern settings from a draft RFC (subject to change)
func PreferredUnsortedEncOptions() EncOptions  // modern settings from a draft RFC (subject to change)
```

__API for Creating & Using Decoding Modes__
```
// DecMode interface uses immutable options and is safe for concurrent use.
type DecMode interface {
	Unmarshal(data []byte, v interface{}) error
	NewDecoder(r io.Reader) *Decoder
	DecOptions() DecOptions  // returns copy of options
}

// DecOptions specifies decoding options.
type DecOptions struct {
...
}

// DecMode returns a DecMode interface created from DecOptions.
func (opts DecOptions) DecMode() (DecMode, error)
```

The `keyasint` and `toarray` struct tags are worth knowing.  They can reduce programming effort, improve system performance, and reduce the size of serialized data.  

See [API docs (godoc.org)](https://godoc.org/github.com/fxamacker/cbor) for more details and more functions.  See [Usage section](#usage) for usage and code examples.  Options for the encoder are listed here.

__Integers always encode to the shortest form that preserves value__.  Encoding of other data types and map key sort order are determined by encoding options.

__Encoding Options__:

Encoder has options that can be set individually. These functions are provided to create and return a modifiable EncOptions struct with predefined settings.

|Predefined EncOptions        |Description                                                                            |
|-----------------------------|---------------------------------------------------------------------------------------|
|CanonicalEncOptions() |[Canonical CBOR (RFC 7049 Section 3.9)](https://tools.ietf.org/html/rfc7049#section-3.9).|
|CTAP2EncOptions() |[CTAP2 Canonical CBOR (FIDO2 CTAP2)](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form).|
|PreferredUnsortedEncOptions() |Unsorted, encode float64->float32->float16 when values fit, NaN values encoded as float16 0x7e00.|
|CoreDetEncOptions() |PreferredUnsortedEncOptions() + map keys are sorted bytewise lexicographic.|

🌱 CoreDetEncOptions() and PreferredUnsortedEncOptions() are subject to change until the draft RFC they used is approved by IETF.

|EncOptions.Sort              |Description                                                                            |
|-----------------------------|---------------------------------------------------------------------------------------|
|SortNone (default) |No sorting for map keys.|
|SortLengthFirst |Length-first map key ordering.|
|SortBytewiseLexical |Bytewise lexicographic map key ordering|
|SortCanonical |(alias) Same as SortLengthFirst [(RFC 7049 Section 3.9)](https://tools.ietf.org/html/rfc7049#section-3.9)|
|SortCTAP2Canonical |(alias) Same as SortBytewiseLexical [(CTAP2 Canonical CBOR)](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form).|
|SortCoreDeterministic |(alias) Same as SortBytewiseLexical.|


|EncOptions.Time              |Description                                                                            |
|-----------------------------|---------------------------------------------------------------------------------------|
|TimeUnix (default) |(seconds) Encode as integer.|
|TimeUnixMicro |(microseconds) Encode as floating-point.  ShortestFloat option determines size.|
|TimeUnixDynamic |(seconds or microseconds) Encode as integer if time doesn't have fractional seconds, otherwise encode as floating-point rounded to microseconds.|
|TimeRFC3339 |(seconds) Encode as RFC 3339 formatted string.|
|TimeRFC3339Nano |(nanoseconds) Encode as RFC3339 formatted string.|

By default, time values are encoded without tags.  Option to override this will be added in v2.1.

Encoder has 3 types of options for floating-point data: ShortestFloatMode, InfConvertMode, and NaNConvertMode.

|EncOptions.ShortestFloat     |Description                                                                            |
|-----------------------------|---------------------------------------------------------------------------------------|
|ShortestFloatNone (default) |No size conversion. Encode float32 and float64 to CBOR floating-point of same bit-size.|
|ShortestFloat16 |Encode float64 -> float32 -> float16 ([IEEE 754 binary16](https://en.wikipedia.org/wiki/Half-precision_floating-point_format)) when values fit.|

Conversions for infinity and NaN use InfConvert and NaNConvert settings.

|EncOptions.InfConvert        |Description                                                                            |
|-----------------------------|---------------------------------------------------------------------------------------|
|InfConvertFloat16 (default) |Convert +- infinity to float16 since they always preserve value (recommended)|
|InfConvertNone |Don't convert +- infinity to other representations -- used by CTAP2 Canonical CBOR|


|EncOptions.NaNConvert        |Description                                                                            |
|-----------------------------|---------------------------------------------------------------------------------------|
|NaNConvert7e00 (default) |Encode to 0xf97e00 (CBOR float16 = 0x7e00) -- used by RFC 7049 Canonical CBOR.|
|NaNConvertNone |Don't convert NaN to other representations -- used by CTAP2 Canonical CBOR.|
|NaNConvertQuiet |Force quiet bit = 1 and use shortest form that preserves NaN payload.|
|NaNConvertPreserveSignal |Convert to smallest form that preserves value (quit bit unmodified and NaN payload preserved).|

<hr>

👉  [Status](#current-status) • [Design Goals](#design-goals) • [Features](#features) • [Standards](#standards) • [Install](#installation) • [API](#api) • [Usage](#usage) • [Fuzzing](#fuzzing-and-code-coverage) • [Security Policy](#security-policy) • [License](#license)

## Usage
🛡️ Use Go's `io.LimitReader` to limit size when decoding very large data.

Functions with identical signatures to encoding/json include:  
`Marshal`, `Unmarshal`, `NewEncoder`, `NewDecoder`, `encoder.Encode`, `decoder.Decode`.

__Default Mode of Encoding__  

If default options are acceptable, package level functions can be used for encoding and decoding.
```
b, err := cbor.Marshal(v)
err := cbor.Unmarshal(b, &v)
encoder := cbor.NewEncoder(w)
decoder := cbor.NewDecoder(r)
```

__Creating and Using Encoding Modes__

EncMode is an interface ([API](#api)) created from EncOptions struct.  EncMode uses immutable options after being created and is safe for concurrent use.  For best performance, EncMode should be reused.

```
// Create EncOptions using either struct literal or a function.
opts := cbor.CanonicalEncOptions()

// If needed, modify encoding options
opts.Time = cbor.TimeUnix

// Create reusable EncMode interface with immutable options, safe for concurrent use.
em, err := opts.EncMode()   

// Use EncMode like encoding/json, with same function signatures.
b, err := em.Marshal(v)
// or
encoder := em.NewEncoder(w)
err := encoder.Encode(v)
```

__Struct Tags (keyasint, toarray, omitempty)__

The `keyasint`, `toarray`, and `omitempty` struct tags make it easy to use compact CBOR message formats.  Internet standards often use CBOR arrays and CBOR maps with int keys to save space.

<hr>

[![CBOR API](https://github.com/fxamacker/images/raw/master/cbor/v2.0.0/cbor_easy_api.png)](#usage)

<hr>

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

var v []*SenMLRecord
if err := cbor.Unmarshal(data, &v); err != nil {
	return err
}
```

__Encoding SenML__ using `keyasint` struct tag and `ShortestFloat16` encoding option:
```
// use SenMLRecord struct defined in "Decoding SenML" example

var v []*SenMLRecord
...
em, err := cbor.EncOptions{ShortestFloat: cbor.ShortestFloat16}.EncMode()
if data, err := em.Marshal(v); err != nil {
	return err
}
```

For more examples, see [examples_test.go](example_test.go).

<hr>

👉  [Status](#current-status) • [Design Goals](#design-goals) • [Features](#features) • [Standards](#standards) • [Install](#installation) • [API](#api) • [Usage](#usage) • [Fuzzing](#fuzzing-and-code-coverage) • [Security Policy](#security-policy) • [License](#license)

## Benchmarks

Go structs are faster than maps with string keys:

* decoding into struct is >27% faster than decoding into map.
* encoding struct is >36% faster than encoding map.

Go structs with `keyasint` struct tag are faster than maps with integer keys:

* decoding into struct is >23% faster than decoding into map.
* encoding struct is >35% faster than encoding map.

Go structs with `toarray` struct tag are faster than slice:

* decoding into struct is >14% faster than decoding into slice.
* encoding struct is >12% faster than encoding slice.

Doing your own benchmarks is highly recommended.  Use your most common message sizes and data types.

See [Benchmarks for fxamacker/cbor](CBOR_BENCHMARKS.md).

## Fuzzing and Code Coverage

__Over 375 tests__ must pass on 4 architectures before tagging a release.  They include all RFC 7049 examples, bugs found by fuzzing, 2 maliciously crafted CBOR data, and over 87 tests with malformed data.

__Code coverage__ must not fall below 95% when tagging a release.  Code coverage is 97.9% (`go test -cover`) for cbor v2.0 which is among the highest for libraries (in Go) of this type.

__Coverage-guided fuzzing__ must pass 250+ million execs before tagging a release.  E.g. v1.4 passed 532+ million execs in coverage-guided fuzzing at the time of release and reached 4+ billion execs 18 days later. Fuzzing uses [fxamacker/cbor-fuzz](https://github.com/fxamacker/cbor-fuzz).  Default corpus has:

* 2 files related to WebAuthn (FIDO U2F key).
* 3 files with custom struct.
* 9 files with [CWT examples (RFC 8392 Appendix A)](https://tools.ietf.org/html/rfc8392#appendix-A)
* 17 files with [COSE examples (RFC 8152 Appendix B & C)](https://github.com/cose-wg/Examples/tree/master/RFC8152).
* 81 files with [CBOR examples (RFC 7049 Appendix A) ](https://tools.ietf.org/html/rfc7049#appendix-A). It excludes 1 errata first reported in [issue #46](https://github.com/fxamacker/cbor/issues/46).

Over 1,100 files (corpus) are used for fuzzing because it includes fuzz-generated corpus.

## Versions and API Changes
This project uses [Semantic Versioning](https://semver.org), so the API is always backwards compatible unless the major version number changes.  Newly added API documented as "subject to change" are excluded from this rule.

Six codec functions have signatures identical to encoding/json and they will never change even after major new releases:  `Marshal`, `Unmarshal`, `NewEncoder`, `NewDecoder`, `encoder.Encode`, and `decoder.Decode`.

## Code of Conduct 
This project has adopted the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).  Contact [faye.github@gmail.com](mailto:faye.github@gmail.com) with any questions or comments.

## Contributing
Please refer to [How to Contribute](CONTRIBUTING.md).

## Security Policy
Security fixes are provided for the latest released version.

To report security vulnerabilities, please email [faye.github@gmail.com](mailto:faye.github@gmail.com) and allow time for the problem to be resolved before reporting it to the public.

## Disclaimers
Phrases like "no crashes" or "doesn't crash" mean there are no known crash bugs in the latest version based on results of unit tests and coverage-guided fuzzing.  It doesn't imply the software is 100% bug-free or 100% invulnerable to all known and unknown attacks.

Please read the license for additional disclaimers and terms.

## Special Thanks

__Making this library better__  

* Montgomery Edwards⁴⁴⁸ for contributing [float16 conversion code](https://github.com/x448/float16), updating the docs, creating comparison charts & slideshow, filing issues, nudging me to ask for feedback from users, and helping with design of v2.0 and v2.1 API.
* Stefan Tatschner for using this library in [sep](https://git.sr.ht/~rumpelsepp/sep), being the 1st to discover my CBOR library, requesting time.Time in issue #1, and submitting this library in a [PR to cbor.io](https://github.com/cbor/cbor.github.io/pull/56) on Aug 12, 2019.
* Yawning Angel for using this library to [oasis-core](https://github.com/oasislabs/oasis-core), and requesting BinaryMarshaler in issue #5.
* Jernej Kos for requesting RawMessage in issue #11 and offering feedback on v2.1 API for CBOR tags.
* ZenGround0 for using this library in [go-filecoin](https://github.com/filecoin-project/go-filecoin), filing "toarray" bug in issue #129, and requesting CBOR BSTR <--> Go array in #133.
* Keith Randall for [fixing Go bugs and providing workarounds](https://github.com/golang/go/issues/36400) so we don't have to wait for new versions of Go.

__Help clarifying CBOR RFC 7049 or 7049bis__

* Carsten Bormann for RFC 7049 (CBOR), his fast confirmation to my RFC 7049 errata, approving my pull request to 7049bis, and his patience when I misread a line in 7049bis.
* Jeffrey Yasskin for his help on the IETF mailing list for 7049bis.
* Laurence Lundblade for his help on the IETF mailing list for 7049bis.

__Words of encouragement and support__

* Jakob Borg for his words of encouragement about this library at Go Forum.  This is especially appreciated in the early stages when there's a lot of rough edges.


## License 
Copyright © 2019-present [Faye Amacker](https://github.com/fxamacker).

fxamacker/cbor is licensed under the MIT License.  See [LICENSE](LICENSE) for the full license text.

<hr>

👉  [Status](#current-status) • [Design Goals](#design-goals) • [Features](#features) • [Standards](#standards) • [Install](#installation) • [API](#api) • [Usage](#usage) • [Fuzzing](#fuzzing-and-code-coverage) • [Security Policy](#security-policy) • [License](#license)
