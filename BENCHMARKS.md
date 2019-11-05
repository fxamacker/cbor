# Benchmarks for fxamacker/cbor 

See [bench_test.go](bench_test.go).

Benchmarks use data representing the following values:

* Boolean: `true`
* Positive integer: `18446744073709551615`
* Negative integer: `-1000`
* Float: `-4.1`
* Byte string: `h'0102030405060708090a0b0c0d0e0f101112131415161718191a'`
* Text string: `"The quick brown fox jumps over the lazy dog"`
* Array: `[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26]`
* Map: `{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E", "f": "F", "g": "G", "h": "H", "i": "I", "j": "J", "l": "L", "m": "M", "n": "N"}}`

Benchmarks show that decoding into struct is >66% faster than decoding into map, and encoding struct is >64% faster than encoding map.  

Decoding Benchmark | Time | Memory | Allocs 
--- | ---: | ---: | ---:
BenchmarkUnmarshal/CBOR_bool_to_Go_interface_{}-2 | 119 ns/op | 16 B/op | 1 allocs/op
BenchmarkUnmarshal/CBOR_bool_to_Go_bool-2 | 66.8 ns/op | 1 B/op | 1 allocs/op
BenchmarkUnmarshal/CBOR_positive_int_to_Go_interface_{}-2 | 142 ns/op | 24 B/op | 2 allocs/op
BenchmarkUnmarshal/CBOR_positive_int_to_Go_uint64-2 | 71.4 ns/op | 8 B/op | 1 allocs/op
BenchmarkUnmarshal/CBOR_negative_int_to_Go_interface_{}-2 | 140 ns/op | 24 B/op | 2 allocs/op
BenchmarkUnmarshal/CBOR_negative_int_to_Go_int64-2 | 73.9 ns/op | 8 B/op | 1 allocs/op
BenchmarkUnmarshal/CBOR_float_to_Go_interface_{}-2 | 145 ns/op | 24 B/op | 2 allocs/op
BenchmarkUnmarshal/CBOR_float_to_Go_float64-2 | 71.9 ns/op | 8 B/op | 1 allocs/op
BenchmarkUnmarshal/CBOR_bytes_to_Go_interface_{}-2 | 187 ns/op | 80 B/op | 3 allocs/op
BenchmarkUnmarshal/CBOR_bytes_to_Go_[]uint8-2 | 158 ns/op | 64 B/op | 2 allocs/op
BenchmarkUnmarshal/CBOR_text_to_Go_interface_{}-2 | 215 ns/op | 80 B/op | 3 allocs/op
BenchmarkUnmarshal/CBOR_text_to_Go_string-2 | 159 ns/op | 64 B/op | 2 allocs/op
BenchmarkUnmarshal/CBOR_array_to_Go_interface_{}-2 |1122 ns/op | 672 B/op | 29 allocs/op
BenchmarkUnmarshal/CBOR_array_to_Go_[]int-2 | 1025 ns/op | 272 B/op | 3 allocs/op
BenchmarkUnmarshal/CBOR_map_to_Go_interface_{}-2 | 3043 ns/op | 1420 B/op | 30 allocs/op
BenchmarkUnmarshal/CBOR_map_to_Go_map[string]interface_{}-2 | 3857 ns/op | 964 B/op | 19 allocs/op
BenchmarkUnmarshal/CBOR_map_to_Go_map[string]string-2 | 2647 ns/op | 741 B/op | 5 allocs/op
BenchmarkUnmarshal/CBOR_map_to_Go_struct-2 | 1274 ns/op| 224 B/op | 2 allocs/op

Encoding Benchmark | Time | Memory | Allocs 
--- | ---: | ---: | ---:
BenchmarkMarshal/Go_bool_to_CBOR_bool-2 | 75.0 ns/op	| 1 B/op | 1 allocs/op
BenchmarkMarshal/Go_uint64_to_CBOR_positive_int-2 | 86.1 ns/op | 16 B/op | 1 allocs/op
BenchmarkMarshal/Go_int64_to_CBOR_negative_int-2 | 81.4 ns/op | 3 B/op | 1 allocs/op
BenchmarkMarshal/Go_float64_to_CBOR_float-2 | 85.4 ns/op	| 16 B/op | 1 allocs/op
BenchmarkMarshal/Go_[]uint8_to_CBOR_bytes-2 | 119 ns/op | 32 B/op	| 1 allocs/op
BenchmarkMarshal/Go_string_to_CBOR_text-2 | 105 ns/op | 48 B/op | 1 allocs/op
BenchmarkMarshal/Go_[]int_to_CBOR_array-2 | 561 ns/op | 32 B/op	| 1 allocs/op
BenchmarkMarshal/Go_map[string]string_to_CBOR_map-2 | 2181 ns/op | 576 B/op | 28 allocs/op
BenchmarkMarshal/Go_struct_to_CBOR_map-2 | 780 ns/op | 64 B/op | 1 allocs/op
