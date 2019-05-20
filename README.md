[![Build Status](https://travis-ci.com/fxamacker/cbor.svg?branch=master)](https://travis-ci.com/fxamacker/cbor)
[![codecov](https://codecov.io/gh/fxamacker/cbor/branch/master/graph/badge.svg?v=4)](https://codecov.io/gh/fxamacker/cbor)
[![Go Report Card](https://goreportcard.com/badge/github.com/fxamacker/cbor)](https://goreportcard.com/report/github.com/fxamacker/cbor)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/fxamacker/cbor)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/fxamacker/cbor/master/LICENSE)

# cbor  

`cbor` is a [CBOR](http://tools.ietf.org/html/rfc7049) encoding and decoding package written in Go.  

The goals of this package are: lightweight, idiomatic, and reasonably fast.  

This package adds less than 400KB to the size of your binaries with no external dependencies.  

`cbor` adopts `json` package API, supports struct field format tags under "cbor" key, and follows `json` struct fields visibility rules.  If you are productive with `json` package, it is very easy to use this package.  

`cbor` strives to balance between fast performance and small binary.  It does not use `unsafe` package to avoid possible incompatibility with future Go releases.  It does not use code generation to keep binary small.  Instead, this package caches struct field types to improve struct encoding and decoding performance.  It bypasses `reflect` when decoding CBOR array/map into empty interface value.  It also uses `sync.Pool` to reuse transient objects.  See [benchmarks](#benchmarks).

## Canonical CBOR Support

This package supports [RFC 7049 canonical CBOR encoding](https://tools.ietf.org/html/rfc7049#section-3.9), and [CTAP2 canonical CBOR encoding](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html#ctap2-canonical-cbor-encoding-form).  CTAP2 canonical CBOR encoding is used by [CTAP](https://fidoalliance.org/specs/fido-v2.0-id-20180227/fido-client-to-authenticator-protocol-v2.0-id-20180227.html) and [WebAuthn](https://www.w3.org/TR/webauthn/) in [FIDO2](https://fidoalliance.org/fido2/) framework.

## Features

* Idiomatic API as in `json` package.
* No external dependencies.
* No use of `unsafe` package.
* Tested with [RFC 7049 test examples](https://tools.ietf.org/html/rfc7049#appendix-A).
* ~90% code coverage.
* Fuzz tested using [cbor-fuzz](https://gitHub.com/fxamacker/cbor-fuzz).
* Decode indefinite-length bytes/string/array/map.
* Decode slices, maps, and structs in-place.
* Decode into struct with field name case-insensitive match.
* Support struct field format tags under "cbor" key.
* Encode anonymous struct fields by `json` package struct fields visibility rules.
* Support [canonical CBOR encoding](#canonical-cbor-support) for map/struct.
* Encode and decode nil slice/map/pointer/interface values correctly.

## Installation 

```
go get github.com/fxamacker/cbor
```

## Usage

See [examples](example_test.go).

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

// create an encoder with canonical CBOR encoding enabled
enc := cbor.NewEncoder(writer, cbor.EncOptions{Canonical: true})

// encode struct
err = enc.Encode(stru)

// encode map
err = enc.Encode(m)

// encode primitives
err = enc.Encode(f)
```

## API 

See [API docs](https://godoc.org/github.com/fxamacker/cbor).

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
type InvalidUnmarshalError struct{ ... }
type InvalidValueError struct{ ... }
type SemanticError struct{ ... }
type SyntaxError struct{ ... }
type UnmarshalTypeError struct{ ... }
type UnsupportedTypeError struct{ ... }
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
BenchmarkMarshal/Go_map[string]string_to_CBOR_map-2                                	  500000	      3856 ns/op	     896 B/op	      29 allocs/op
BenchmarkMarshal/Go_cbor_test.strc_to_CBOR_map-2                                   	 2000000	       868 ns/op	      64 B/op	       1 allocs/op
```

## Limitations

* This package doesn't support CBOR tag encoding.
* Decoder ignores CBOR tag and decodes tagged data following the tag.

## License 

Copyright (c) 2019 [Faye Amacker](https://github.com/fxamacker)

Licensed under [MIT License](LICENSE)
