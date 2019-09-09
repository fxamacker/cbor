[![Build Status](https://travis-ci.com/fxamacker/cbor.svg?branch=master)](https://travis-ci.com/fxamacker/cbor)
[![codecov](https://codecov.io/gh/fxamacker/cbor/branch/master/graph/badge.svg?v=4)](https://codecov.io/gh/fxamacker/cbor)
[![Go Report Card](https://goreportcard.com/badge/github.com/fxamacker/cbor)](https://goreportcard.com/report/github.com/fxamacker/cbor)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/fxamacker/cbor)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/fxamacker/cbor/master/LICENSE)

# cbor - CBOR encoding and decoding in Go

This library is designed to be:
* __Easy__ -- idiomatic Go API (like encoding/json).
* __Safe and reliable__ -- no `unsafe` pkg, test coverage at ~90%, and 9+ hrs of fuzzing with RFC 7049 test vectors.
* __Standards-compliant__ -- supports [RFC 7049](https://tools.ietf.org/html/rfc7049) and canonical CBOR encodings (both [RFC 7049](https://tools.ietf.org/html/rfc7049#section-3.9) and [CTAP2](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form)).
* __Small and self-contained__ -- pkg compiles to under 0.5 MB with no external dependencies.

`cbor` balances speed, safety (no `unsafe` pkg) and compiled size.  To keep size small, it doesn't use code generation.  For speed, it caches struct field types, bypasses `reflect` when appropriate, and uses `sync.Pool` to reuse transient objects.  

## Current Status

Sept 9, 2019: Current version (0.3) is expected to be promoted to 1.0 this month unless changes are requested by the Go community.  It passed 9+ hours of fuzzing and appears to be ready for production use on linux_amd64.

## Size comparison

Program size comparison (linux_amd64, Go 1.12) doing the same CBOR encoding and decoding:
- 2.7 MB program using fxamacker/cbor
- 11.9 MB program using ugorji/go

Library size comparison (linux_amd64, Go 1.12):
- 0.45 MB pkg -- fxamacker/cbor
- 2.9 MB pkg -- ugorji/go without code generation (`go install --tags "notfastpath"`)
- 5.7 MB pkg -- ugorji/go with code generation (default build)

## Features

* Idiomatic API as in `json` package.
* No external dependencies.
* No use of `unsafe` package.
* Tested with [RFC 7049 test vectors](https://tools.ietf.org/html/rfc7049#appendix-A).
* Test coverage at ~90%, and fuzzed 9+ hours using [cbor-fuzz](https://github.com/fxamacker/cbor-fuzz).
* Decode slices, maps, and structs in-place.
* Decode into struct with field name case-insensitive match.
* Support canonical CBOR encoding for map/struct.
* Support struct field format tags under "cbor" key.
* Encode anonymous struct fields by `json` package struct fields visibility rules.
* Encode and decode nil slice/map/pointer/interface values correctly.
* Encode and decode indefinite length bytes/string/array/map (["streaming"](https://tools.ietf.org/html/rfc7049#section-2.2)).
* Encode and decode time.Time as RFC 3339 formatted text string or Unix time.

## Standards 

This library implements CBOR as specified in [RFC 7049](https://tools.ietf.org/html/rfc7049), with minor [limitations](#limitations).

It also supports canonical CBOR encodings (both [RFC 7049](https://tools.ietf.org/html/rfc7049#section-3.9) and [CTAP2](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form)).  CTAP2 canonical CBOR encoding is used by [CTAP](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html) and [WebAuthn](https://www.w3.org/TR/webauthn/) in [FIDO2](https://fidoalliance.org/fido2/) framework.

## Limitations

* This package doesn't support CBOR tag encoding.
* Decoder ignores CBOR tag and decodes tagged data following the tag.
* Signed integer values incompatible with Go's int64 are not supported.
* RFC 7049 test vectors with signed integer values incompatible with Go's int64 are skipped. For example, the signed integer result -18446744073709551616 is incompatible with Go's int64 data type (cannot be assigned without overflow).

## Versions and API changes

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
type InvalidValueError struct{ ... }
type SemanticError struct{ ... }
type SyntaxError struct{ ... }
type UnmarshalTypeError struct{ ... }
type UnsupportedTypeError struct{ ... }
```

## Installation 

```
go get github.com/fxamacker/cbor
```

## Usage

See [examples](example_test.go).

Using decoder:

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

// decode into primitives
var f float32
err = dec.Decode(&f)
```

Using encoder:

```
// create an encoder with canonical CBOR encoding enabled
enc := cbor.NewEncoder(writer, cbor.EncOptions{Canonical: true})

// encode struct
err = enc.Encode(stru)

// encode map
err = enc.Encode(m)

// encode primitives
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

// close nested indefinite length array
err = enc.EndIndefinite()

// close outer indefinite length array
err = enc.EndIndefinite()
```

## Benchmarks

See [bench_test.go](bench_test.go).

`Unmarshal` benchmarks are made on CBOR data representing the following values:

* Boolean: `true`
* Positive integer: `18446744073709551615`
* Negative integer: `-1000`
* Float: `-4.1`
* Byte string: `h'0102030405060708090a0b0c0d0e0f101112131415161718191a'`
* Text string: `"The quick brown fox jumps over the lazy dog"`
* Array: `[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26]`
* Map: `{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"}}`

`Marshal` benchmarks are made on Go values representing the same values.

Benchmarks shows that decoding into struct is >50% faster than decoding into map, and encoding struct is >70% faster than encoding map.  

```
BenchmarkUnmarshal/CBOR_boolean_to_Go_interface_{}-2         	                        10000000	       132 ns/op	      16 B/op	       1 allocs/op
BenchmarkUnmarshal/CBOR_boolean_to_Go_bool-2                 	                        20000000	      76.1 ns/op	       1 B/op	       1 allocs/op
BenchmarkUnmarshal/CBOR_positive_integer_to_Go_interface_{}-2         	                10000000	       159 ns/op	      24 B/op	       2 allocs/op
BenchmarkUnmarshal/CBOR_positive_integer_to_Go_uint64-2               	                20000000	      82.3 ns/op	       8 B/op	       1 allocs/op
BenchmarkUnmarshal/CBOR_negative_integer_to_Go_interface_{}-2         	                10000000	       160 ns/op	      24 B/op	       2 allocs/op
BenchmarkUnmarshal/CBOR_negative_integer_to_Go_int64-2                	                20000000	      83.8 ns/op	       8 B/op	       1 allocs/op
BenchmarkUnmarshal/CBOR_float_to_Go_interface_{}-2                    	                10000000	       159 ns/op	      24 B/op	       2 allocs/op
BenchmarkUnmarshal/CBOR_float_to_Go_float64-2                         	                20000000	      81.4 ns/op	       8 B/op	       1 allocs/op
BenchmarkUnmarshal/CBOR_byte_string_to_Go_interface_{}-2              	                10000000	       214 ns/op	      80 B/op	       3 allocs/op
BenchmarkUnmarshal/CBOR_byte_string_to_Go_[]uint8-2                   	                10000000	       159 ns/op	      64 B/op	       2 allocs/op
BenchmarkUnmarshal/CBOR_byte_string_indefinite_length_to_Go_interface_{}-2         	 2000000	       738 ns/op	     112 B/op	       3 allocs/op
BenchmarkUnmarshal/CBOR_byte_string_indefinite_length_to_Go_[]uint8-2              	 2000000	       687 ns/op	      96 B/op	       2 allocs/op
BenchmarkUnmarshal/CBOR_text_string_to_Go_interface_{}-2                           	 5000000	       248 ns/op	      80 B/op	       3 allocs/op
BenchmarkUnmarshal/CBOR_text_string_to_Go_string-2                                 	10000000	       176 ns/op	      64 B/op	       2 allocs/op
BenchmarkUnmarshal/CBOR_text_string_indefinite_length_to_Go_interface_{}-2         	 1000000	      1147 ns/op	     144 B/op	       4 allocs/op
BenchmarkUnmarshal/CBOR_text_string_indefinite_length_to_Go_string-2               	 1000000	      1075 ns/op	     128 B/op	       3 allocs/op
BenchmarkUnmarshal/CBOR_array_to_Go_interface_{}-2                                 	 1000000	      1156 ns/op	     672 B/op	      29 allocs/op
BenchmarkUnmarshal/CBOR_array_to_Go_[]int-2                                        	 1000000	      1075 ns/op	     272 B/op	       3 allocs/op
BenchmarkUnmarshal/CBOR_array_indefinite_length_to_Go_interface_{}-2               	 1000000	      1347 ns/op	     672 B/op	      29 allocs/op
BenchmarkUnmarshal/CBOR_array_indefinite_length_to_Go_[]int-2                      	 1000000	      1275 ns/op	     272 B/op	       3 allocs/op
BenchmarkUnmarshal/CBOR_map_to_Go_interface_{}-2                                   	  500000	      3094 ns/op	    1420 B/op	      30 allocs/op
BenchmarkUnmarshal/CBOR_map_to_Go_map[string]interface_{}-2                        	  300000	      4064 ns/op	     964 B/op	      19 allocs/op
BenchmarkUnmarshal/CBOR_map_to_Go_map[string]string-2                              	  500000	      2808 ns/op	     740 B/op	       5 allocs/op
BenchmarkUnmarshal/CBOR_map_to_Go_cbor_test.strc-2                                 	 1000000	      1624 ns/op	     208 B/op	       1 allocs/op
BenchmarkUnmarshal/CBOR_map_indefinite_length_to_Go_interface_{}-2                 	  300000	      4236 ns/op	    2607 B/op	      33 allocs/op
BenchmarkUnmarshal/CBOR_map_indefinite_length_to_Go_map[string]interface_{}-2      	  300000	      5819 ns/op	    2422 B/op	      22 allocs/op
BenchmarkUnmarshal/CBOR_map_indefinite_length_to_Go_map[string]string-2            	  300000	      4268 ns/op	    2183 B/op	       7 allocs/op
BenchmarkUnmarshal/CBOR_map_indefinite_length_to_Go_cbor_test.strc-2               	 1000000	      1860 ns/op	     208 B/op	       1 allocs/op
```

```
BenchmarkMarshal/Go_bool_to_CBOR_boolean-2                                         	20000000	      62.8 ns/op	       1 B/op	       1 allocs/op
BenchmarkMarshal/Go_uint64_to_CBOR_positive_integer-2                              	20000000	      73.4 ns/op	      16 B/op	       1 allocs/op
BenchmarkMarshal/Go_int64_to_CBOR_negative_integer-2                               	20000000	      64.7 ns/op	       3 B/op	       1 allocs/op
BenchmarkMarshal/Go_float64_to_CBOR_float-2                                        	20000000	      70.4 ns/op	      16 B/op	       1 allocs/op
BenchmarkMarshal/Go_[]uint8_to_CBOR_byte_string-2                                  	20000000	       105 ns/op	      32 B/op	       1 allocs/op
BenchmarkMarshal/Go_string_to_CBOR_text_string-2                                   	20000000	      95.0 ns/op	      48 B/op	       1 allocs/op
BenchmarkMarshal/Go_[]int_to_CBOR_array-2                                          	 3000000	       599 ns/op	      32 B/op	       1 allocs/op
BenchmarkMarshal/Go_map[string]string_to_CBOR_map-2                                	  500000	      3276 ns/op	     576 B/op	      28 allocs/op
BenchmarkMarshal/Go_cbor_test.strc_to_CBOR_map-2                                   	 2000000	       868 ns/op	      64 B/op	       1 allocs/op
```

## License 

Copyright (c) 2019 [Faye Amacker](https://github.com/fxamacker)

Licensed under [MIT License](LICENSE)
