# CBOR Codec in Go

<!-- [![](https://github.com/fxamacker/images/raw/master/cbor/v2.5.0/fxamacker_cbor_banner.png)](#cbor-library-in-go) -->

[fxamacker/cbor](https://github.com/fxamacker/cbor) is a library for encoding and decoding [CBOR](https://www.rfc-editor.org/info/std94) and [CBOR Sequences](https://www.rfc-editor.org/rfc/rfc8742.html).

CBOR is a [trusted alternative](https://www.rfc-editor.org/rfc/rfc8949.html#name-comparison-of-other-binary-) to JSON, MessagePack, Protocol Buffers, etc.&nbsp; CBOR is an Internet&nbsp;Standard defined by [IETF&nbsp;STD&nbsp;94 (RFC&nbsp;8949)](https://www.rfc-editor.org/info/std94) and is designed to be relevant for decades.

`fxamacker/cbor` is used in projects by Arm Ltd., Cisco, Dapper Labs, EdgeX&nbsp;Foundry, Fraunhofer&#8209;AISEC, Linux&nbsp;Foundation, Microsoft, Mozilla, Oasis&nbsp;Protocol, Tailscale, Teleport, [and&nbsp;others](https://github.com/fxamacker/cbor#who-uses-fxamackercbor).

See [Quick&nbsp;Start](#quick-start).

## fxamacker/cbor

[![](https://github.com/fxamacker/cbor/workflows/ci/badge.svg)](https://github.com/fxamacker/cbor/actions?query=workflow%3Aci)
[![](https://github.com/fxamacker/cbor/workflows/cover%20%E2%89%A596%25/badge.svg)](https://github.com/fxamacker/cbor/actions?query=workflow%3A%22cover+%E2%89%A596%25%22)
[![](https://github.com/fxamacker/cbor/workflows/linters/badge.svg)](https://github.com/fxamacker/cbor/actions?query=workflow%3Alinters)
[![CodeQL](https://github.com/fxamacker/cbor/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/fxamacker/cbor/actions/workflows/codeql-analysis.yml)
[![](https://img.shields.io/badge/fuzzing-3%2B%20billion%20execs-44c010)](#fuzzing-and-code-coverage)
[![Go Report Card](https://goreportcard.com/badge/github.com/fxamacker/cbor)](https://goreportcard.com/report/github.com/fxamacker/cbor)
[![](https://img.shields.io/badge/go-%3E%3D%201.12-blue)](#cbor-library-installation)

`fxamacker/cbor` is a CBOR codec in full conformance with [IETF STD&nbsp;94 (RFC&nbsp;8949)](https://www.rfc-editor.org/info/std94). It also supports CBOR Sequences ([RFC&nbsp;8742](https://www.rfc-editor.org/rfc/rfc8742.html)) and Extended Diagnostic Notation ([Appendix G of RFC&nbsp;8610](https://www.rfc-editor.org/rfc/rfc8610.html#appendix-G)).

Features include full support for CBOR tags, [Core Deterministic Encoding](https://www.rfc-editor.org/rfc/rfc8949.html#name-core-deterministic-encoding), duplicate map key detection, etc.

Struct tags (`toarray`, `keyasint`, `omitempty`) reduce encoded size of structs.

![alt text](https://github.com/fxamacker/images/raw/master/cbor/v2.3.0/cbor_struct_tags_api.svg?sanitize=1 "CBOR API and Go Struct Tags")

API is mostly same as `encoding/json`, plus interfaces that simplify concurrency for CBOR options.

Design balances tradeoffs between speed, security, memory, encoded data size, usability, etc.

<details><summary>Design and Feature Highlights</summary><p/>

__🚀&nbsp; Speed__

Encoding and decoding is fast without using Go's `unsafe` package.  Slower settings are opt-in.  Default limits allow very fast and memory efficient rejection of malformed CBOR data.

__🔒&nbsp; Security__

Decoder has configurable limits that defend against malicious inputs.  Duplicate map key detection is supported.  By contrast, `encoding/gob` is [not designed to be hardened against adversarial inputs](https://pkg.go.dev/encoding/gob#hdr-Security).

Codec passed multiple confidential security assessments in 2022.  No vulnerabilities found in subset of codec in a [nonconfidential security assessment](https://github.com/veraison/go-cose/blob/v1.0.0-rc.1/reports/NCC_Microsoft-go-cose-Report_2022-05-26_v1.0.pdf) prepared by NCC&nbsp;Group for Microsoft&nbsp;Corporation.

__🗜️&nbsp; Data Size__

Struct tags (`toarray`, `keyasint`, `omitempty`) automatically reduce size of encoded structs. Encoding optionally shrinks float64→32→16 when values fit.

__:jigsaw:&nbsp; Usability__

API is mostly same as `encoding/json` plus interfaces that simplify concurrency for CBOR options.  Encoding and decoding modes can be created at startup and reused by any goroutines.

Presets include Core Deterministic Encoding, Preferred Serialization, CTAP2 Canonical CBOR, etc.

__📆&nbsp;  Extensibility__

Features include CBOR [extension points](https://www.rfc-editor.org/rfc/rfc8949.html#section-7.1) (e.g. CBOR tags) and extensive settings.  API has interfaces that allow users to create custom encoding and decoding without modifying this library.

</details>

<details><summary>Efficient Rejection of Malicious CBOR Data</summary><p/>

Decoding 10 bytes of malicious data into `[]byte` doesn't exhaust memory. E.g.  
`[]byte{0x9B, 0x00, 0x00, 0x42, 0xFA, 0x42, 0xFA, 0x42, 0xFA, 0x42}`.

| Codec | Speed (ns/op) | Memory | Allocs |
| :---- | ------------: | -----: | -----: |
| fxamacker/cbor 2.5.0-beta2 | 44.33 ± 2% | 32 B/op | 2 allocs/op |
| fxamacker/cbor 0.1.0 - 2.4.0 | ~44.68 ± 6% | 32 B/op |  2 allocs/op |
| ugorji/go 1.2.10 | 5524792.50 ± 3% | 67110491 B/op |  12 allocs/op |
| ugorji/go 1.1.0 - 1.2.6 | 💥 runtime: | out of memory: | cannot allocate |

- go1.19.6, linux/amd64, i5-13600K (DDR4 not overclocked)
- go test -bench=. -benchmem -count=20


</details>

## Quick Start

__Install__: `go get github.com/fxamacker/cbor/v2` and `import "github.com/fxamacker/cbor/v2"`.

### Key Points

- Encoding and decoding modes are created from options (settings).
- Modes can be created at startup and reused.
- Modes are safe for concurrent use.

### Default Mode

Package level functions (default mode) use default settings.

```go
// API matches encoding/json.
b, err := cbor.Marshal(v)        // encode v to []byte b
err := cbor.Unmarshal(b, &v)     // decode []byte b to v
encoder := cbor.NewEncoder(w)    // create encoder with io.Writer w
decoder := cbor.NewDecoder(r)    // create decoder with io.Reader r
```

Some CBOR-based formats or protocols may require non-default settings.

For example, WebAuthn uses "CTAP2 Canonical CBOR" settings.  It is available as a preset.

### Presets

Presets can be used as-is or as a starting point to adjust settings.

```go
// EncOptions is a struct of encoder settings.
func CoreDetEncOptions() EncOptions              // RFC 8949 Core Deterministic Encoding
func PreferredUnsortedEncOptions() EncOptions    // RFC 8949 Preferred Serialization
func CanonicalEncOptions() EncOptions            // RFC 7049 Canonical CBOR
func CTAP2EncOptions() EncOptions                // FIDO2 CTAP2 Canonical CBOR
```

Presets are used to create custom modes.

### Custom Modes

Modes are created from settings. Once created, modes have immutable settings.

💡 Create the mode at startup and reuse it. It is safe for concurrent use.

```Go
// Create encoding mode.
opts := cbor.CoreDetEncOptions()   // use preset options as a starting point
opts.Time = cbor.TimeUnix          // change any settings if needed
em, err := opts.EncMode()          // create an immutable encoding mode

// Reuse the encoding mode. It is safe for concurrent use.

// API matches encoding/json.
b, err := em.Marshal(v)            // encode v to []byte b
encoder := em.NewEncoder(w)        // create encoder with io.Writer w
err := encoder.Encode(v)           // encode v to io.Writer w
```

### Struct Tags

Struct tags (`toarray`, `keyasint`, `omitempty`) reduce encoded size of structs.

<details><summary>Example using struct tags</summary><p/>
	
![alt text](https://github.com/fxamacker/images/raw/master/cbor/v2.3.0/cbor_struct_tags_api.svg?sanitize=1 "CBOR API and Go Struct Tags")

</details>

Struct tags simplify use of CBOR-based protocols that require CBOR arrays or maps with integer keys.

### CBOR Tags

CBOR tags are specified in a `TagSet`.

Custom modes can be created with a `TagSet` to handle CBOR tags.
 
```go
em, err := opts.EncMode()                  // no CBOR tags
em, err := opts.EncModeWithTags(ts)        // immutable CBOR tags
em, err := opts.EncModeWithSharedTags(ts)  // mutable shared CBOR tags
```

`TagSet` and modes using it are safe for concurrent use.  Equivalent API is available for `DecMode`.

<details><summary>Example using TagSet and TagOptions</summary><p/>

```go
// Use signedCWT struct defined in "Decoding CWT" example.

// Create TagSet (safe for concurrency).
tags := cbor.NewTagSet()
// Register tag COSE_Sign1 18 with signedCWT type.
tags.Add(	
	cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired}, 
	reflect.TypeOf(signedCWT{}), 
	18)

// Create DecMode with immutable tags.
dm, _ := cbor.DecOptions{}.DecModeWithTags(tags)

// Unmarshal to signedCWT with tag support.
var v signedCWT
if err := dm.Unmarshal(data, &v); err != nil {
	return err
}

// Create EncMode with immutable tags.
em, _ := cbor.EncOptions{}.EncModeWithTags(tags)

// Marshal signedCWT with tag number.
if data, err := cbor.Marshal(v); err != nil {
	return err
}
```

</details>

### Functions and Interfaces

<details><summary>Functions and interfaces at a glance</summary><p/>

Common functions with same API as `encoding/json`:  
- `Marshal`, `Unmarshal`
- `NewEncoder`, `(*Encoder).Encode`
- `NewDecoder`, `(*Decoder).Decode`

NOTE: `Unmarshal` will return `ExtraneousDataError` if there are remaining bytes
because RFC 8949 treats CBOR data item with remaining bytes as malformed.
- 💡 Use `UnmarshalFirst` to decode first CBOR data item and return any remaining bytes.

Other useful functions: 
- `Diagnose`, `DiagnoseFirst` produce human-readable [Extended Diagnostic Notation](https://www.rfc-editor.org/rfc/rfc8610.html#appendix-G) from CBOR data.
- `UnmarshalFirst` decodes first CBOR data item and return any remaining bytes.

Interfaces identical or comparable to Go `encoding` packages include:  
`Marshaler`, `Unmarshaler`, `BinaryMarshaler`, and `BinaryUnmarshaler`.

The `RawMessage` type can be used to delay CBOR decoding or precompute CBOR encoding.

</details>

### Security Tips

🔒 Use Go's `io.LimitReader` to limit size when decoding very large or indefinite size data.

Default limits may need to be increased for systems handling very large data (e.g. blockchains).

`DecOptions` can be used to modify default limits for `MaxArrayElements`, `MaxMapPairs`, and `MaxNestedLevels`.

## Status

v2.5.0-beta5 is fuzz tested and production quality.  However, docs need to be updated before v2.5.0 release.

IMPORTANT: Changes in [v2.5.0-beta](https://github.com/fxamacker/cbor/releases/tag/v2.5.0-beta) should be reviewed before upgrading because of bugfixes to error handling of extraneous data and other edge cases.

- TODO for v2.5.0: update docs and prepare release notes.  No more features planned for v2.5.0.
- [v2.5.0-beta5](https://github.com/fxamacker/cbor/releases/tag/v2.5.0-beta5) - Add `Decoder.Buffered` function which is same as `encoding/json`.
- [v2.5.0-beta4](https://github.com/fxamacker/cbor/releases/tag/v2.5.0-beta4) - Bugfix for `Diagnose` to return `io.EOF` on empty data like the others.
- [v2.5.0-beta3](https://github.com/fxamacker/cbor/releases/tag/v2.5.0-beta3) - Add functions: `Diagnose`, `DiagnoseFirst`, `UnmarshalFirst`, `Wellformed`.
- [v2.5.0-beta2](https://github.com/fxamacker/cbor/releases/tag/v2.5.0-beta2) - Bugfix to retry in `Decoder` if `io.Reader`'s `Read()` returns 0 bytes read with nil error.
- [v2.5.0-beta](https://github.com/fxamacker/cbor/releases/tag/v2.5.0-beta) - Notable improvements, optimizations, bugfixes, and 8 new contributors!


<details><summary>👉 Benchmark Comparison: v2.4.0 vs v2.5.0-beta2</summary><p/>

Comparison of v2.4.0 vs v2.5.0-beta2 provided by @448 (edited to fit width).

PR [#382](https://github.com/fxamacker/cbor/pull/382) returns buffer to pool in `Encode()`. It adds a bit of overhead to `Encode()` but `NewEncoder().Encode()` is a lot faster and uses less memory as shown here:

```
$ benchstat bench-v2.4.0.log bench-f9e6291.log 
goos: linux
goarch: amd64
pkg: github.com/fxamacker/cbor/v2
cpu: 12th Gen Intel(R) Core(TM) i7-12700H
                                                     │ bench-v2.4.0.log │  bench-f9e6291.log                  │
                                                     │      sec/op      │   sec/op     vs base                │
NewEncoderEncode/Go_bool_to_CBOR_bool-20                   236.70n ± 2%   58.04n ± 1%  -75.48% (p=0.000 n=10)
NewEncoderEncode/Go_uint64_to_CBOR_positive_int-20         238.00n ± 2%   63.93n ± 1%  -73.14% (p=0.000 n=10)
NewEncoderEncode/Go_int64_to_CBOR_negative_int-20          238.65n ± 2%   64.88n ± 1%  -72.81% (p=0.000 n=10)
NewEncoderEncode/Go_float64_to_CBOR_float-20               242.00n ± 2%   63.00n ± 1%  -73.97% (p=0.000 n=10)
NewEncoderEncode/Go_[]uint8_to_CBOR_bytes-20               245.60n ± 1%   68.55n ± 1%  -72.09% (p=0.000 n=10)
NewEncoderEncode/Go_string_to_CBOR_text-20                 243.20n ± 3%   68.39n ± 1%  -71.88% (p=0.000 n=10)
NewEncoderEncode/Go_[]int_to_CBOR_array-20                 563.0n ± 2%    378.3n ± 0%  -32.81% (p=0.000 n=10)
NewEncoderEncode/Go_map[string]string_to_CBOR_map-20       2.043µ ± 2%    1.906µ ± 2%   -6.75% (p=0.000 n=10)
geomean                                                    349.7n         122.7n       -64.92%

                                                     │ bench-v2.4.0.log │    bench-f9e6291.log                │
                                                     │       B/op       │    B/op     vs base                 │
NewEncoderEncode/Go_bool_to_CBOR_bool-20                     128.0 ± 0%     0.0 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_uint64_to_CBOR_positive_int-20           128.0 ± 0%     0.0 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_int64_to_CBOR_negative_int-20            128.0 ± 0%     0.0 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_float64_to_CBOR_float-20                 128.0 ± 0%     0.0 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_[]uint8_to_CBOR_bytes-20                 128.0 ± 0%     0.0 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_string_to_CBOR_text-20                   128.0 ± 0%     0.0 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_[]int_to_CBOR_array-20                   128.0 ± 0%     0.0 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_map[string]string_to_CBOR_map-20         544.0 ± 0%   416.0 ± 0%   -23.53% (p=0.000 n=10)
geomean                                                      153.4                    ?                       ¹ ²
¹ summaries must be >0 to compute geomean
² ratios must be >0 to compute geomean

                                                     │ bench-v2.4.0.log │    bench-f9e6291.log                │
                                                     │    allocs/op     │ allocs/op   vs base                 │
NewEncoderEncode/Go_bool_to_CBOR_bool-20                     2.000 ± 0%   0.000 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_uint64_to_CBOR_positive_int-20           2.000 ± 0%   0.000 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_int64_to_CBOR_negative_int-20            2.000 ± 0%   0.000 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_float64_to_CBOR_float-20                 2.000 ± 0%   0.000 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_[]uint8_to_CBOR_bytes-20                 2.000 ± 0%   0.000 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_string_to_CBOR_text-20                   2.000 ± 0%   0.000 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_[]int_to_CBOR_array-20                   2.000 ± 0%   0.000 ± 0%  -100.00% (p=0.000 n=10)
NewEncoderEncode/Go_map[string]string_to_CBOR_map-20         28.00 ± 0%   26.00 ± 0%    -7.14% (p=0.000 n=10)
geomean                                                      2.782                    ?                       ¹ ²
¹ summaries must be >0 to compute geomean
² ratios must be >0 to compute geomean
```

</details>

<!--
## What is CBOR?

[CBOR](https://tools.ietf.org/html/rfc8949) is a concise binary data format inspired by [JSON](https://www.json.org) and [MessagePack](https://msgpack.org).  CBOR is defined in [RFC 8949](https://tools.ietf.org/html/rfc8949) (December 2020) which obsoletes [RFC 7049](https://tools.ietf.org/html/rfc7049) (October 2013).  

CBOR is an [Internet Standard](https://en.wikipedia.org/wiki/Internet_Standard) by [IETF](https://www.ietf.org).  It's used in other standards like [WebAuthn](https://en.wikipedia.org/wiki/WebAuthn) by [W3C](https://www.w3.org), [COSE (RFC 8152)](https://tools.ietf.org/html/rfc8152), [CWT (RFC 8392)](https://tools.ietf.org/html/rfc8392), [CDDL (RFC 8610)](https://datatracker.ietf.org/doc/html/rfc8610) and [more](CBOR_GOLANG.md).

[Reasons for choosing CBOR](https://github.com/fxamacker/cbor/wiki/Why-CBOR) vary by project.  Some projects replaced protobuf, encoding/json, encoding/gob, etc. with CBOR.  For example, by replacing protobuf with CBOR in gRPC.
-->

## Who uses fxamacker/cbor

`fxamacker/cbor` is used in projects by Arm Ltd., Berlin Institute of Health at Charité, Chainlink, Cisco, Confidential Computing Consortium, ConsenSys, Dapper&nbsp;Labs, EdgeX&nbsp;Foundry, F5, Fraunhofer&#8209;AISEC, Linux&nbsp;Foundation, Microsoft, Mozilla, National&nbsp;Cybersecurity&nbsp;Agency&nbsp;of&nbsp;France (govt), Netherlands (govt), Oasis Protocol, Smallstep, Tailscale, Taurus SA, Teleport, TIBCO, and others.

Github reports [2000+ repositories](https://github.com/fxamacker/cbor/network/dependents?package_id=UGFja2FnZS0yMjcwNDY1OTQ4) depend on fxamacker/cbor/v2. Additional 190+ repos are using v1 (please upgrade to v2).

`fxamacker/cbor` passed multiple confidential security assessments.  A [nonconfidential security assessment](https://github.com/veraison/go-cose/blob/v1.0.0-rc.1/reports/NCC_Microsoft-go-cose-Report_2022-05-26_v1.0.pdf) (prepared by NCC Group for Microsoft Corporation) includes a subset of fxamacker/cbor v2.4.0 in its scope.

## CBOR Options

### Encoding Options

Integers always encode to the shortest form that preserves value.  By default, time values are encoded without tags.

Encoding of other data types and map key sort order are determined by encoder options.

| EncOptions | Available Settings (defaults listed first)
| :--- | :--- |
| Sort | **SortNone**, SortLengthFirst, SortBytewiseLexical <br/> Aliases: SortCanonical, SortCTAP2, SortCoreDeterministic |
| Time | **TimeUnix**, TimeUnixMicro, TimeUnixDynamic, TimeRFC3339, TimeRFC3339Nano |
| TimeTag | **EncTagNone**, EncTagRequired |
| ShortestFloat | **ShortestFloatNone**, ShortestFloat16  |
| BigIntConvert | **BigIntConvertShortest**, BigIntConvertNone |
| InfConvert | **InfConvertFloat16**, InfConvertNone |
| NaNConvert | **NaNConvert7e00**, NaNConvertNone, NaNConvertQuiet, NaNConvertPreserveSignal |
| IndefLength | **IndefLengthAllowed**, IndefLengthForbidden  |
| TagsMd | **TagsAllowed**, TagsForbidden |

See [Options](#options) section for details about each setting.

### Decoding Options

| DecOptions | Available Settings (defaults listed first)  |
| :--- | :--- |
| TimeTag | **DecTagIgnored**, DecTagOptional, DecTagRequired |
| DupMapKey | **DupMapKeyQuiet**, DupMapKeyEnforcedAPF |
| IntDec | **IntDecConvertNone**, IntDecConvertSigned |
| IndefLength | **IndefLengthAllowed**, IndefLengthForbidden |
| TagsMd | **TagsAllowed**, TagsForbidden |
| ExtraReturnErrors | **ExtraDecErrorNone**, ExtraDecErrorUnknownField |
| MaxNestedLevels | **32**, can be set to [4, 65535] |
| MaxArrayElements | **131072**, can be set to [16, 2147483647] |
| MaxMapPairs | **131072**, can be set to [16, 2147483647] |

See [Options](#options) section for details about each setting.

## Standards
This library is a full-featured generic CBOR [(RFC 8949)](https://tools.ietf.org/html/rfc8949) encoder and decoder.  Notable CBOR features include:

| CBOR Feature  | Description  |
| :--- | :--- |
| CBOR tags | API supports built-in and user-defined tags.  |
| Preferred serialization | Integers encode to fewest bytes. Optional float64 → float32 → float16. |
| Map key sorting | Unsorted, length-first (Canonical CBOR), and bytewise-lexicographic (CTAP2). |
| Duplicate map keys | Always forbid for encoding and option to allow/forbid for decoding.   |
| Indefinite length data | Option to allow/forbid for encoding and decoding. |
| Well-formedness | Always checked and enforced. |
| Basic validity checks | Check UTF-8 validity and optionally check duplicate map keys. |
| Security considerations | Prevent integer overflow and resource exhaustion (RFC 8949 Section 10). |

See the Features section for list of [Encoding Options](#encoding-options) and [Decoding Options](#decoding-options).

Known limitations are noted in the [Limitations section](#limitations). 

Go nil values for slices, maps, pointers, etc. are encoded as CBOR null.  Empty slices, maps, etc. are encoded as empty CBOR arrays and maps.

Decoder checks for all required well-formedness errors, including all "subkinds" of syntax errors and too little data.

After well-formedness is verified, basic validity errors are handled as follows:

* Invalid UTF-8 string: Decoder always checks and returns invalid UTF-8 string error.
* Duplicate keys in a map: Decoder has options to ignore or enforce rejection of duplicate map keys.

When decoding well-formed CBOR arrays and maps, decoder saves the first error it encounters and continues with the next item.  Options to handle this differently may be added in the future.

By default, decoder treats time values of floating-point NaN and Infinity as if they are CBOR Null or CBOR Undefined.

See [Options](#options) section for detailed settings or [Features](#features) section for a summary of options.

__Click to expand topic:__

<details>
 <summary>Duplicate Map Keys</summary><p>

This library provides options for fast detection and rejection of duplicate map keys based on applying a Go-specific data model to CBOR's extended generic data model in order to determine duplicate vs distinct map keys. Detection relies on whether the CBOR map key would be a duplicate "key" when decoded and applied to the user-provided Go map or struct. 

`DupMapKeyQuiet` turns off detection of duplicate map keys. It tries to use a "keep fastest" method by choosing either "keep first" or "keep last" depending on the Go data type.

`DupMapKeyEnforcedAPF` enforces detection and rejection of duplidate map keys. Decoding stops immediately and returns `DupMapKeyError` when the first duplicate key is detected. The error includes the duplicate map key and the index number. 

APF suffix means "Allow Partial Fill" so the destination map or struct can contain some decoded values at the time of error. It is the caller's responsibility to respond to the `DupMapKeyError` by discarding the partially filled result if that's required by their protocol.

</details>

<details>
 <summary>Tag Validity</summary><p>

This library checks tag validity for built-in tags (currently tag numbers 0, 1, 2, 3, and 55799):

* Inadmissible type for tag content 
* Inadmissible value for tag content

Unknown tag data items (not tag number 0, 1, 2, 3, or 55799) are handled in two ways:

* When decoding into an empty interface, unknown tag data item will be decoded into `cbor.Tag` data type, which contains tag number and tag content.  The tag content will be decoded into the default Go data type for the CBOR data type.
* When decoding into other Go types, unknown tag data item is decoded into the specified Go type.  If Go type is registered with a tag number, the tag number can optionally be verified.

Decoder also has an option to forbid tag data items (treat any tag data item as error) which is specified by protocols such as CTAP2 Canonical CBOR.  

For more information, see [decoding options](#decoding-options-1) and [tag options](#tag-options).

</details>

## Limitations

If any of these limitations prevent you from using this library, please open an issue along with a link to your project.

* CBOR `Undefined` (0xf7) value decodes to Go's `nil` value.  CBOR `Null` (0xf6) more closely matches Go's `nil`.
* CBOR `simple values` that are unassigned/reserved by IANA are not fully supported until v2.5.0.
* CBOR map keys with data types not supported by Go for map keys are ignored and an error is returned after continuing to decode remaining items.  
* When using io.Reader interface to read very large or indefinite length CBOR data, Go's `io.LimitReader` should be used to limit size.
* When decoding registered CBOR tag data to interface type, decoder creates a pointer to registered Go type matching CBOR tag number.  Requiring a pointer for this is a Go limitation. 

<hr>

## API
Many function signatures are identical to Go's encoding/json, such as:  
`Marshal`, `Unmarshal`, `NewEncoder`, `NewDecoder`, `(*Encoder).Encode`, and `(*Decoder).Decode`.

Interfaces identical or comparable to Go's encoding, encoding/json, or encoding/gob include:  
`Marshaler`, `Unmarshaler`, `BinaryMarshaler`, and `BinaryUnmarshaler`.

Like `encoding/json`, `RawMessage` can be used to delay CBOR decoding or precompute CBOR encoding.

"Mode" in this API means defined way of encoding or decoding -- it links the standard API to CBOR options and CBOR tags.

EncMode and DecMode are interfaces created from EncOptions or DecOptions structs.  
For example, `em, err := cbor.EncOptions{...}.EncMode()` or `em, err := cbor.CanonicalEncOptions().EncMode()`.

EncMode and DecMode use immutable options so their behavior won't accidentally change at runtime.  Modes are intended to be reused and are safe for concurrent use.

__API for Default Mode__

If default options are acceptable, then you don't need to create EncMode or DecMode.

```go
Marshal(v interface{}) ([]byte, error)
NewEncoder(w io.Writer) *Encoder

Unmarshal(data []byte, v interface{}) error
NewDecoder(r io.Reader) *Decoder
```

__API for Creating & Using Encoding Modes__

```go
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
func (opts EncOptions) EncMode() (EncMode, error) {}

// EncModeWithTags returns EncMode with options and tags that are both immutable. 
func (opts EncOptions) EncModeWithTags(tags TagSet) (EncMode, error) {}

// EncModeWithSharedTags returns EncMode with immutable options and mutable shared tags. 
func (opts EncOptions) EncModeWithSharedTags(tags TagSet) (EncMode, error) {}
```

The empty curly braces prevent a syntax highlighting bug, please ignore them.

__API for Predefined Encoding Options__

```go
func CoreDetEncOptions() EncOptions {}              // RFC 8949 Core Deterministic Encoding
func PreferredUnsortedEncOptions() EncOptions {}    // RFC 8949 Preferred Serialization
func CanonicalEncOptions() EncOptions {}            // RFC 7049 Canonical CBOR
func CTAP2EncOptions() EncOptions {}                // FIDO2 CTAP2 Canonical CBOR
```

__API for Creating & Using Decoding Modes__

```go
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
func (opts DecOptions) DecMode() (DecMode, error) {}

// DecModeWithTags returns DecMode with options and tags that are both immutable. 
func (opts DecOptions) DecModeWithTags(tags TagSet) (DecMode, error) {}

// DecModeWithSharedTags returns DecMode with immutable options and mutable shared tags. 
func (opts DecOptions) DecModeWithSharedTags(tags TagSet) (DecMode, error) {}
```

The empty curly braces prevent a syntax highlighting bug, please ignore them.

__API for Using CBOR Tags__

`TagSet` can be used to associate user-defined Go type(s) to tag number(s).  It's also used to create EncMode or DecMode. For example, `em := EncOptions{...}.EncModeWithTags(ts)` or `em := EncOptions{...}.EncModeWithSharedTags(ts)`. This allows every standard API exported by em (like `Marshal` and `NewEncoder`) to use the specified tags automatically.

`Tag` and `RawTag` can be used to encode/decode a tag number with a Go value, but `TagSet` is generally recommended.

```go
type TagSet interface {
    // Add adds given tag number(s), content type, and tag options to TagSet.
    Add(opts TagOptions, contentType reflect.Type, num uint64, nestedNum ...uint64) error

    // Remove removes given tag content type from TagSet.
    Remove(contentType reflect.Type)    
}
```

`Tag` and `RawTag` types can also be used to encode/decode tag number with Go value.

```go
type Tag struct {
    Number  uint64
    Content interface{}
}

type RawTag struct {
    Number  uint64
    Content RawMessage
}
```

See [API docs (godoc.org)](https://godoc.org/github.com/fxamacker/cbor/v2) for more details and more functions.  See [Usage section](#usage) for usage and code examples.

<hr>

## Go Struct Tags

This library supports both "cbor" and "json" key for some (not all) struct tags.  If "cbor" and "json" keys are both present for the same field, then "cbor" key will be used.

| Key | Format Str | Scope | Description |
| --- | ---------- | ----- | ------------|
| cbor or json | "myName" | field | Name of field to use such as "myName", etc. like encoding/json. |
| cbor or json | ",omitempty" | field | Omit (ignore) this field if value is empty, like encoding/json. |
| cbor or json | "-" | field | Omit (ignore) this field always, like encoding/json. |
| cbor | ",keyasint" | field | Treat field as an element of CBOR map with specified int as key. |
| cbor | ",toarray" | struct | Treat each field as an element of CBOR array. This automatically disables "omitempty" and "keyasint" for all fields in the struct. |

The "keyasint" struct tag requires an integer key to be specified:

```
type myStruct struct {
    MyField     int64    `cbor:"-1,keyasint,omitempty'`
    OurField    string   `cbor:"0,keyasint,omitempty"`
    FooField    Foo      `cbor:"5,keyasint,omitempty"`
    BarField    Bar      `cbor:"hello,omitempty"`
    ...
}
```

The "toarray" struct tag requires a special field "_" (underscore) to indicate "toarray" applies to the entire struct:

```
type myStruct struct {
    _           struct{}    `cbor:",toarray"`
    MyField     int64
    OurField    string
    ...
}
```

__Click to expand:__

<details>
  <summary>Example Using CBOR Web Tokens</summary><p>
   
![alt text](https://github.com/fxamacker/images/raw/master/cbor/v2.3.0/cbor_struct_tags_api.svg?sanitize=1 "CBOR API and Go Struct Tags")

</details>

## Options

Decoding options and encoding options.

### Decoding Options

| DecOptions.TimeTag | Description |
| ------------------ | ----------- |
| DecTagIgnored (default) | Tag numbers are ignored (if present) for time values. |
| DecTagOptional | Tag numbers are only checked for validity if present for time values. |
| DecTagRequired | Tag numbers must be provided for time values except for CBOR Null and CBOR Undefined. |

The following CBOR time values are decoded as Go's "zero time instant":

* CBOR Null
* CBOR Undefined
* CBOR floating-point NaN
* CBOR floating-point Infinity

Go's `time` package provides `IsZero` function, which reports whether t represents "zero time instant"  
(January 1, year 1, 00:00:00 UTC).

<br>

| DecOptions.DupMapKey | Description |
| -------------------- | ----------- |
| DupMapKeyQuiet (default) | turns off detection of duplicate map keys. It uses a "keep fastest" method by choosing either "keep first" or "keep last" depending on the Go data type. |
| DupMapKeyEnforcedAPF | enforces detection and rejection of duplidate map keys. Decoding stops immediately and returns `DupMapKeyError` when the first duplicate key is detected. The error includes the duplicate map key and the index number. |

`DupMapKeyEnforcedAPF` uses "Allow Partial Fill" so the destination map or struct can contain some decoded values at the time of error.  Users can respond to the `DupMapKeyError` by discarding the partially filled result if that's required by their protocol.

<br>

| DecOptions.IntDec | Description |
| ------------------ | ----------- |
| IntDecConvertNone (default) | When decoding to Go interface{}, CBOR positive int (major type 0) decode to uint64 value, and CBOR negative int (major type 1) decode to int64 value. |
| IntDecConvertSigned | When decoding to Go interface{}, CBOR positive/negative int (major type 0 and 1) decode to int64 value. |

If `IntDecConvertedSigned` is used and value overflows int64, UnmarshalTypeError is returned.

<br>

| DecOptions.IndefLength | Description |
| ---------------------- | ----------- |
|IndefLengthAllowed (default) | allow indefinite length data |
|IndefLengthForbidden | forbid indefinite length data |

<br>

| DecOptions.TagsMd | Description |
| ----------------- | ----------- |
|TagsAllowed (default) | allow CBOR tags (major type 6) |
|TagsForbidden | forbid CBOR tags (major type 6) |

<br>

| DecOptions.ExtraReturnErrors | Description |
| ----------------- | ----------- |
|ExtraDecErrorNone (default) | no extra decoding errors.  E.g. ignore unknown fields if encountered. |
|ExtraDecErrorUnknownField | return error if unknown field is encountered |

<br>

| DecOptions.MaxNestedLevels | Description |
| -------------------------- | ----------- |
| 32 (default) | allowed setting is [4, 65535] |

<br>

| DecOptions.MaxArrayElements | Description |
| --------------------------- | ----------- |
| 131072 (default) | allowed setting is [16, 2147483647] |

<br>

| DecOptions.MaxMapPairs | Description |
| ---------------------- | ----------- |
| 131072 (default) | allowed setting is [16, 2147483647] |

### Encoding Options

__Integers always encode to the shortest form that preserves value__.  Encoding of other data types and map key sort order are determined by encoding options.

These functions are provided to create and return a modifiable EncOptions struct with predefined settings.

| Predefined EncOptions | Description |
| --------------------- | ----------- |
| CanonicalEncOptions() |[Canonical CBOR (RFC 7049 Section 3.9)](https://tools.ietf.org/html/rfc7049#section-3.9). |
| CTAP2EncOptions() |[CTAP2 Canonical CBOR (FIDO2 CTAP2)](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form). |
| PreferredUnsortedEncOptions() |Unsorted, encode float64->float32->float16 when values fit, NaN values encoded as float16 0x7e00. |
| CoreDetEncOptions() |PreferredUnsortedEncOptions() + map keys are sorted bytewise lexicographic. |

<br>

| EncOptions.Sort | Description |
| --------------- | ----------- |
| SortNone (default) |No sorting for map keys. |
| SortLengthFirst |Length-first map key ordering. |
| SortBytewiseLexical |Bytewise lexicographic map key ordering [(RFC 8949 Section 4.2.1)](https://datatracker.ietf.org/doc/html/rfc8949#section-4.2.1).|
| SortCanonical |(alias) Same as SortLengthFirst [(RFC 7049 Section 3.9)](https://tools.ietf.org/html/rfc7049#section-3.9) |
| SortCTAP2 |(alias) Same as SortBytewiseLexical [(CTAP2 Canonical CBOR)](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form). |
| SortCoreDeterministic |(alias) Same as SortBytewiseLexical [(RFC 8949 Section 4.2.1)](https://datatracker.ietf.org/doc/html/rfc8949#section-4.2.1). |

<br>

| EncOptions.Time | Description |
| --------------- | ----------- |
| TimeUnix (default) | (seconds) Encode as integer. |
| TimeUnixMicro | (microseconds) Encode as floating-point.  ShortestFloat option determines size. |
| TimeUnixDynamic | (seconds or microseconds) Encode as integer if time doesn't have fractional seconds, otherwise encode as floating-point rounded to microseconds. |
| TimeRFC3339 | (seconds) Encode as RFC 3339 formatted string. |
| TimeRFC3339Nano | (nanoseconds) Encode as RFC3339 formatted string. |

<br>

| EncOptions.TimeTag | Description |
| ------------------ | ----------- |
| EncTagNone (default) | Tag number will not be encoded for time values. |
| EncTagRequired | Tag number (0 or 1) will be encoded unless time value is undefined/zero-instant. |

By default, undefined (zero instant) time values will encode as CBOR Null without tag number for both EncTagNone and EncTagRequired.  Although CBOR Undefined might be technically more correct for EncTagRequired, CBOR Undefined might not be supported by other generic decoders and it isn't supported by JSON.

Go's `time` package provides `IsZero` function, which reports whether t represents the zero time instant, January 1, year 1, 00:00:00 UTC. 

<br>

| EncOptions.BigIntConvert | Description |
| ------------------------ | ----------- |
| BigIntConvertShortest (default) | Encode big.Int as CBOR integer if value fits. |
| BigIntConvertNone | Encode big.Int as CBOR bignum (tag 2 or 3). |

<br>

__Floating-Point Options__

Encoder has 3 types of options for floating-point data: ShortestFloatMode, InfConvertMode, and NaNConvertMode.

| EncOptions.ShortestFloat | Description |
| ------------------------ | ----------- |
| ShortestFloatNone (default) | No size conversion. Encode float32 and float64 to CBOR floating-point of same bit-size. |
| ShortestFloat16 | Encode float64 -> float32 -> float16 ([IEEE 754 binary16](https://en.wikipedia.org/wiki/Half-precision_floating-point_format)) when values fit. |

Conversions for infinity and NaN use InfConvert and NaNConvert settings.

| EncOptions.InfConvert | Description |
| --------------------- | ----------- |
| InfConvertFloat16 (default) | Convert +- infinity to float16 since they always preserve value (recommended) |
| InfConvertNone |Don't convert +- infinity to other representations -- used by CTAP2 Canonical CBOR |

<br>

| EncOptions.NaNConvert | Description |
| --------------------- | ----------- |
| NaNConvert7e00 (default) | Encode to 0xf97e00 (CBOR float16 = 0x7e00) -- used by RFC 8949 Preferred Encoding, etc. |
| NaNConvertNone | Don't convert NaN to other representations -- used by CTAP2 Canonical CBOR. |
| NaNConvertQuiet | Force quiet bit = 1 and use shortest form that preserves NaN payload. |
| NaNConvertPreserveSignal | Convert to smallest form that preserves value (quit bit unmodified and NaN payload preserved). |

<br>

| EncOptions.IndefLength | Description |
| ---------------------- | ----------- |
|IndefLengthAllowed (default) | allow indefinite length data |
|IndefLengthForbidden | forbid indefinite length data |

<br>

| EncOptions.TagsMd | Description |
| ----------------- | ----------- |
|TagsAllowed (default) | allow CBOR tags (major type 6) |
|TagsForbidden | forbid CBOR tags (major type 6) |

### Tag Options

TagOptions specifies how encoder and decoder handle tag number registered with TagSet.

| TagOptions.DecTag | Description |
| ------------------ | ----------- |
| DecTagIgnored (default) | Tag numbers are ignored (if present). |
| DecTagOptional | Tag numbers are only checked for validity if present. |
| DecTagRequired | Tag numbers must be provided except for CBOR Null and CBOR Undefined. |

<br>

| TagOptions.EncTag | Description |
| ------------------ | ----------- |
| EncTagNone (default) | Tag number will not be encoded. |
| EncTagRequired | Tag number will be encoded. |
	
<hr>

## Fuzzing and Code Coverage

__Over 375 tests__ must pass on 4 architectures before tagging a release.  They include all RFC 7049 and RFC 8949 examples, bugs found by fuzzing, maliciously crafted CBOR data, and over 87 tests with malformed data.  There's some overlap in the tests but it isn't a high priority to trim tests.

__Code coverage__ must not fall below 95% when tagging a release.  Code coverage is above 96% (`go test -cover`) for cbor v2.5 which is among the highest for codecs written in Go.

__Coverage-guided fuzzing__ must pass billions of execs using a previously generated corpus before tagging a release.  Fuzzing is usually continued after the release is tagged and is manually stopped after reaching several billion execs.  Fuzzing is done using nonpublic code which may eventually get merged into this project.

To prevent delays to release schedules, fuzzing is not restarted for a release if changes are limited to ci, docs, and comments.

<hr>

## Versions and API Changes
This project uses [Semantic Versioning](https://semver.org), so the API is always backwards compatible unless the major version number changes.  

These functions have signatures identical to encoding/json and they will likely never change even after major new releases:  
`Marshal`, `Unmarshal`, `NewEncoder`, `NewDecoder`, `(*Encoder).Encode`, and `(*Decoder).Decode`.

Newly added API documented as "subject to change" are excluded from SemVer.

Newly added API in the master branch that has never been release tagged are excluded from SemVer.

Bug fixes like detecting an error that was missed in prior version are excluded from SemVer as long as function parameters, etc. are unchanged.

## Code of Conduct 
This project has adopted the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).  Contact [faye.github@gmail.com](mailto:faye.github@gmail.com) with any questions or comments.

## Contributing
Please refer to [How to Contribute](CONTRIBUTING.md).

## Security Policy
Security fixes are provided for the latest released version of fxamacker/cbor.

For the full text of the Security Policy, see [SECURITY.md](SECURITY.md).

## Disclaimers
Phrases like "no crashes", "doesn't crash", and "is secure" mean there are no known crash bugs in the latest version based on results of unit tests and coverage-guided fuzzing.  They don't imply the software is 100% bug-free or 100% invulnerable to all known and unknown attacks.

Please read the license for additional disclaimers and terms.

## Special Thanks

__Making this library better__  

* Stefan Tatschner for using this library in [sep](https://rumpelsepp.org/projects/sep), being the 1st to discover my CBOR library, requesting time.Time in issue #1, and submitting this library in a [PR to cbor.io](https://github.com/cbor/cbor.github.io/pull/56) on Aug 12, 2019.
* Yawning Angel for using this library to [oasis-core](https://github.com/oasislabs/oasis-core), and requesting BinaryMarshaler in issue #5.
* Jernej Kos for requesting RawMessage in issue #11 and offering feedback on v2.1 API for CBOR tags.
* ZenGround0 for using this library in [go-filecoin](https://github.com/filecoin-project/go-filecoin), filing "toarray" bug in issue #129, and requesting  
CBOR BSTR <--> Go array in #133.
* Keith Randall for [fixing Go bugs and providing workarounds](https://github.com/golang/go/issues/36400) so we don't have to wait for new versions of Go.

__Help clarifying CBOR RFC 7049 or 7049bis (7049bis is the draft of RFC 8949)__

* Carsten Bormann for RFC 7049 (CBOR), adding this library to cbor.io, his fast confirmation to my RFC 7049 errata, approving my pull request to 7049bis, and his patience when I misread a line in 7049bis.
* Laurence Lundblade for his help on the IETF mailing list for 7049bis and for pointing out on a CBORbis issue that CBOR Undefined might be problematic translating to JSON.
* Jeffrey Yasskin for his help on the IETF mailing list for 7049bis.

__Words of encouragement and support__

* Jakob Borg for his words of encouragement about this library at Go Forum.  This is especially appreciated in the early stages when there's a lot of rough edges.


## License 
Copyright © 2019-2023 [Faye Amacker](https://github.com/fxamacker).  

fxamacker/cbor is licensed under the MIT License.  See [LICENSE](LICENSE) for the full license text.  

<hr>
