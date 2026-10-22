// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Tests to unmarshal date/time

func TestDecodeTime(t *testing.T) {
	unmodified := time.Now()

	testCases := []struct {
		name            string
		cborRFC3339Time []byte
		cborUnixTime    []byte
		wantTime        time.Time
	}{
		// Decoding untagged CBOR null/defined to time.Time is no-op.  See TestUnmarshalNull.
		{
			name:            "null within unrecognized tag", // no-op in DecTagIgnored
			cborRFC3339Time: mustHexDecode("dadeadbeeff6"),
			cborUnixTime:    mustHexDecode("dadeadbeeff6"),
			wantTime:        unmodified,
		},
		{
			name:            "undefined within unrecognized tag", // no-op in DecTagIgnored
			cborRFC3339Time: mustHexDecode("dadeadbeeff7"),
			cborUnixTime:    mustHexDecode("dadeadbeeff7"),
			wantTime:        unmodified,
		},
		{
			name:            "NaN",
			cborRFC3339Time: mustHexDecode("f97e00"),
			cborUnixTime:    mustHexDecode("f97e00"),
			wantTime:        time.Time{},
		},
		{
			name:            "positive infinity",
			cborRFC3339Time: mustHexDecode("f97c00"),
			cborUnixTime:    mustHexDecode("f97c00"),
			wantTime:        time.Time{},
		},
		{
			name:            "negative infinity",
			cborRFC3339Time: mustHexDecode("f9fc00"),
			cborUnixTime:    mustHexDecode("f9fc00"),
			wantTime:        time.Time{},
		},
		{
			name:            "time without fractional seconds", // positive integer
			cborRFC3339Time: mustHexDecode("74323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("1a514b67b0"),
			wantTime:        mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "time with fractional seconds", // float
			cborRFC3339Time: mustHexDecode("7819313937302d30312d30315432313a34363a34302d30363a3030"),
			cborUnixTime:    mustHexDecode("fa47c35000"),
			wantTime:        mustParseTime(time.RFC3339Nano, "1970-01-01T21:46:40-06:00"),
		},
		{
			name:            "time with fractional seconds", // float
			cborRFC3339Time: mustHexDecode("76323031332d30332d32315432303a30343a30302e355a"),
			cborUnixTime:    mustHexDecode("fb41d452d9ec200000"),
			wantTime:        mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00.5Z"),
		},
		{
			name:            "time before January 1, 1970 UTC without fractional seconds", // negative integer
			cborRFC3339Time: mustHexDecode("74313936392d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("3a0177f2cf"),
			wantTime:        mustParseTime(time.RFC3339Nano, "1969-03-21T20:04:00Z"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm := unmodified
			if err := unmarshal(t, tc.cborRFC3339Time, &tm); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborRFC3339Time, err)
			} else if !tc.wantTime.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborRFC3339Time, tm, tm, tc.wantTime, tc.wantTime)
			}
			tm = unmodified
			if err := unmarshal(t, tc.cborUnixTime, &tm); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborUnixTime, err)
			} else if !tc.wantTime.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborUnixTime, tm, tm, tc.wantTime, tc.wantTime)
			}
		})
	}
}

func TestDecodeTimeWithTag(t *testing.T) {
	testCases := []struct {
		name            string
		cborRFC3339Time []byte
		cborUnixTime    []byte
		wantTime        time.Time
	}{
		{
			name:            "time without fractional seconds", // positive integer
			cborRFC3339Time: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("c11a514b67b0"),
			wantTime:        mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "time with fractional seconds", // float
			cborRFC3339Time: mustHexDecode("c076323031332d30332d32315432303a30343a30302e355a"),
			cborUnixTime:    mustHexDecode("c1fb41d452d9ec200000"),
			wantTime:        mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00.5Z"),
		},
		{
			name:            "time before January 1, 1970 UTC without fractional seconds", // negative integer
			cborRFC3339Time: mustHexDecode("c074313936392d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("c13a0177f2cf"),
			wantTime:        mustParseTime(time.RFC3339Nano, "1969-03-21T20:04:00Z"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm := time.Now()
			if err := unmarshal(t, tc.cborRFC3339Time, &tm); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborRFC3339Time, err)
			} else if !tc.wantTime.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborRFC3339Time, tm, tm, tc.wantTime, tc.wantTime)
			}
			tm = time.Now()
			if err := unmarshal(t, tc.cborUnixTime, &tm); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborUnixTime, err)
			} else if !tc.wantTime.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborUnixTime, tm, tm, tc.wantTime, tc.wantTime)
			}

			var v any
			if err := unmarshal(t, tc.cborRFC3339Time, &v); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborRFC3339Time, err)
			} else if tm, ok := v.(time.Time); !ok || !tc.wantTime.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborRFC3339Time, v, v, tc.wantTime, tc.wantTime)
			}
			v = nil
			if err := unmarshal(t, tc.cborUnixTime, &v); err != nil {
				t.Errorf("Unmarshal(0x%x) returned error %v", tc.cborUnixTime, err)
			} else if tm, ok := v.(time.Time); !ok || !tc.wantTime.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborUnixTime, v, v, tc.wantTime, tc.wantTime)
			}
		})
	}
}

func TestDecodeTimeError(t *testing.T) {
	testCases := []struct {
		name         string
		opts         DecOptions
		data         []byte
		wantErrorMsg string
	}{
		{
			name:         "invalid RFC3339 time string",
			data:         mustHexDecode("7f657374726561646d696e67ff"),
			wantErrorMsg: "cbor: cannot set streaming for time.Time",
		},
		{
			name:         "byte string data cannot be decoded into time.Time",
			data:         mustHexDecode("4f013030303030303030e03031ed3030"),
			wantErrorMsg: "cbor: cannot unmarshal byte string into Go value of type time.Time",
		},
		{
			name:         "bool cannot be decoded into time.Time",
			data:         mustHexDecode("f4"),
			wantErrorMsg: "cbor: cannot unmarshal primitives into Go value of type time.Time",
		},
		{
			name:         "invalid UTF-8 string",
			data:         mustHexDecode("7f62e6b061b4ff"),
			wantErrorMsg: "cbor: invalid UTF-8 string",
		},
		{
			name:         "negative integer overflow",
			data:         mustHexDecode("3bffffffffffffffff"),
			wantErrorMsg: "cbor: cannot unmarshal negative integer into Go value of type time.Time",
		},
		{
			name: "untagged byte string content cannot be decoded into time.Time with DefaultByteStringType string",
			opts: DecOptions{
				TimeTag:               DecTagOptional,
				DefaultByteStringType: reflect.TypeFor[string](),
			},
			data:         mustHexDecode("54323031332d30332d32315432303a30343a30305a"),
			wantErrorMsg: "cbor: cannot unmarshal byte string into Go value of type time.Time",
		},
		{
			name:         "time tag is validated when enclosed in unrecognized tag",
			data:         mustHexDecode("dadeadbeefc001"),
			wantErrorMsg: "cbor: tag number 0 must be followed by text string, got positive integer",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dm, err := tc.opts.DecMode()
			if err != nil {
				t.Fatal(err)
			}
			tm := time.Now()
			if err := unmarshalWithDM(t, dm, tc.data, &tm); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", tc.data, tc.wantErrorMsg)
			} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.data, err.Error(), tc.wantErrorMsg)
			}
		})
	}
}

func TestDecodeInvalidTagTime(t *testing.T) {
	typeTimeSlice := reflect.TypeFor[[]time.Time]()

	testCases := []struct {
		name          string
		data          []byte
		decodeToTypes []reflect.Type
		wantErrorMsg  string
	}{
		{
			name:          "tag 0 with invalid RFC3339 time string",
			data:          mustHexDecode("c07f657374726561646d696e67ff"),
			decodeToTypes: []reflect.Type{typeAny, typeTime},
			wantErrorMsg:  "cbor: cannot set streaming for time.Time",
		},
		{
			name:          "tag 0 with invalid UTF-8 string",
			data:          mustHexDecode("c07f62e6b061b4ff"),
			decodeToTypes: []reflect.Type{typeAny, typeTime},
			wantErrorMsg:  "cbor: invalid UTF-8 string",
		},
		{
			name:          "tag 0 with integer content",
			data:          mustHexDecode("c01a514b67b0"),
			decodeToTypes: []reflect.Type{typeAny, typeTime},
			wantErrorMsg:  "cbor: tag number 0 must be followed by text string, got positive integer",
		},
		{
			name:          "tag 0 with byte string content",
			data:          mustHexDecode("c04f013030303030303030e03031ed3030"),
			decodeToTypes: []reflect.Type{typeAny, typeTime},
			wantErrorMsg:  "cbor: tag number 0 must be followed by text string, got byte string",
		},
		{
			name:          "tag 0 with integer content as array element",
			data:          mustHexDecode("81c01a514b67b0"),
			decodeToTypes: []reflect.Type{typeAny, typeTimeSlice},
			wantErrorMsg:  "cbor: tag number 0 must be followed by text string, got positive integer",
		},
		{
			name:          "tag 1 with negative integer overflow",
			data:          mustHexDecode("c13bffffffffffffffff"),
			decodeToTypes: []reflect.Type{typeAny, typeTime},
			wantErrorMsg:  "cbor: cannot unmarshal negative integer into Go value of type time.Time (-18446744073709551616 overflows Go's int64)",
		},
		{
			name:          "tag 1 with positive float64 overflow",
			data:          mustHexDecode("c1fb4415af1d78b58c40"), // 1(1e+20)
			decodeToTypes: []reflect.Type{typeAny, typeTime},
			wantErrorMsg:  "overflows Go's int64",
		},
		{
			name:          "tag 1 with negative float64 overflow",
			data:          mustHexDecode("c1fbc415af1d78b58c40"), // 1(-1e+20)
			decodeToTypes: []reflect.Type{typeAny, typeTime},
			wantErrorMsg:  "overflows Go's int64",
		},
		{
			name:          "tag 1 with string content",
			data:          mustHexDecode("c174323031332d30332d32315432303a30343a30305a"),
			decodeToTypes: []reflect.Type{typeAny, typeTime},
			wantErrorMsg:  "cbor: tag number 1 must be followed by integer or floating-point number, got UTF-8 text string",
		},
		{
			name:          "tag 1 with simple value",
			data:          mustHexDecode("d801f6"), // 1(null)
			decodeToTypes: []reflect.Type{typeAny, typeTime},
			wantErrorMsg:  "cbor: tag number 1 must be followed by integer or floating-point number, got primitive",
		},
		{
			name:          "tag 1 with string content as array element",
			data:          mustHexDecode("81c174323031332d30332d32315432303a30343a30305a"),
			decodeToTypes: []reflect.Type{typeAny, typeTimeSlice},
			wantErrorMsg:  "cbor: tag number 1 must be followed by integer or floating-point number, got UTF-8 text string",
		},
	}
	dm, _ := DecOptions{TimeTag: DecTagOptional}.DecMode()
	for _, tc := range testCases {
		for _, decodeToType := range tc.decodeToTypes {
			t.Run(tc.name+" decode to "+decodeToType.String(), func(t *testing.T) {
				v := reflect.New(decodeToType)
				if err := unmarshalWithDM(t, dm, tc.data, v.Interface()); err == nil {
					t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", tc.data, tc.wantErrorMsg)
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("Unmarshal(0x%x) returned error %q, want %q", tc.data, err, tc.wantErrorMsg)
				}
			})
		}
	}
}

func TestDecodeTag0Error(t *testing.T) {
	data := mustHexDecode("c01a514b67b0") // 0(1363896240)
	wantErrorMsg := "cbor: tag number 0 must be followed by text string, got positive integer"

	timeTagIgnoredDM, _ := DecOptions{TimeTag: DecTagIgnored}.DecMode()
	timeTagOptionalDM, _ := DecOptions{TimeTag: DecTagOptional}.DecMode()
	timeTagRequiredDM, _ := DecOptions{TimeTag: DecTagRequired}.DecMode()

	testCases := []struct {
		name string
		dm   DecMode
	}{
		{name: "DecTagIgnored", dm: timeTagIgnoredDM},
		{name: "DecTagOptional", dm: timeTagOptionalDM},
		{name: "DecTagRequired", dm: timeTagRequiredDM},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Decode to a value of type any.
			var v any
			if err := unmarshalWithDM(t, tc.dm, data, &v); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", data, wantErrorMsg)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", data, err, wantErrorMsg)
			}

			// Decode to time.Time
			var tm time.Time
			if err := unmarshalWithDM(t, tc.dm, data, &tm); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", data, wantErrorMsg)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", data, err, wantErrorMsg)
			}

			// Decode to uint64
			var ui uint64
			if err := unmarshalWithDM(t, tc.dm, data, &ui); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", data, wantErrorMsg)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", data, err, wantErrorMsg)
			}
		})
	}
}

func TestDecodeTag1Error(t *testing.T) {
	data := mustHexDecode("c174323031332d30332d32315432303a30343a30305a") // 1("2013-03-21T20:04:00Z")
	wantErrorMsg := "cbor: tag number 1 must be followed by integer or floating-point number, got UTF-8 text string"

	timeTagIgnoredDM, _ := DecOptions{TimeTag: DecTagIgnored}.DecMode()
	timeTagOptionalDM, _ := DecOptions{TimeTag: DecTagOptional}.DecMode()
	timeTagRequiredDM, _ := DecOptions{TimeTag: DecTagRequired}.DecMode()

	testCases := []struct {
		name string
		dm   DecMode
	}{
		{name: "DecTagIgnored", dm: timeTagIgnoredDM},
		{name: "DecTagOptional", dm: timeTagOptionalDM},
		{name: "DecTagRequired", dm: timeTagRequiredDM},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Decode to a value of type any.
			var v any
			if err := unmarshalWithDM(t, tc.dm, data, &v); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", data, wantErrorMsg)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", data, err, wantErrorMsg)
			}

			// Decode to time.Time
			var tm time.Time
			if err := unmarshalWithDM(t, tc.dm, data, &tm); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", data, wantErrorMsg)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", data, err, wantErrorMsg)
			}

			// Decode to string
			var s string
			if err := unmarshalWithDM(t, tc.dm, data, &s); err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", data, wantErrorMsg)
			} else if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error %q, want %q", data, err, wantErrorMsg)
			}
		})
	}
}

func TestDecodeTimeStreaming(t *testing.T) {
	// Decoder decodes from mixed invalid and valid time.
	testCases := []struct {
		data         []byte
		wantErrorMsg string
		wantObj      time.Time
	}{
		{
			data:         mustHexDecode("c07f62e6b061b4ff"),
			wantErrorMsg: "cbor: invalid UTF-8 string",
		},
		{
			data:    mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
			wantObj: time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC),
		},
		{
			data:         mustHexDecode("c01a514b67b0"),
			wantErrorMsg: "cbor: tag number 0 must be followed by text string, got positive integer",
		},
		{
			data:    mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
			wantObj: time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC),
		},
		{
			data:         mustHexDecode("c13bffffffffffffffff"),
			wantErrorMsg: "cbor: cannot unmarshal negative integer into Go value of type time.Time (-18446744073709551616 overflows Go's int64)",
		},
		{
			data:    mustHexDecode("c11a514b67b0"),
			wantObj: time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC),
		},
		{
			data:         mustHexDecode("c174323031332d30332d32315432303a30343a30305a"),
			wantErrorMsg: "tag number 1 must be followed by integer or floating-point number, got UTF-8 text string",
		},
		{
			data:    mustHexDecode("c11a514b67b0"),
			wantObj: time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC),
		},
	}
	// Data is a mixed stream of valid and invalid time data
	var data []byte
	for _, tc := range testCases {
		data = append(data, tc.data...)
	}
	dm, _ := DecOptions{TimeTag: DecTagOptional}.DecMode()
	dec := dm.NewDecoder(bytes.NewReader(data))
	for _, tc := range testCases {
		var v any
		err := dec.Decode(&v)
		if tc.wantErrorMsg != "" {
			if err == nil {
				t.Errorf("Unmarshal(0x%x) didn't return an error, want error msg %q", tc.data, tc.wantErrorMsg)
			} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error msg %q, want %q", tc.data, err, tc.wantErrorMsg)
			}
		} else {
			tm, ok := v.(time.Time)
			if !ok {
				t.Errorf("Unmarshal(0x%x) returned %s (%T), want time.Time", tc.data, v, v)
			}
			if !tc.wantObj.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) returned %s, want %s", tc.data, tm, tc.wantObj)
			}
		}
	}
	dec = dm.NewDecoder(bytes.NewReader(data))
	for _, tc := range testCases {
		var tm time.Time
		err := dec.Decode(&tm)
		if tc.wantErrorMsg != "" {
			if err == nil {
				t.Errorf("Unmarshal(0x%x) did't return error, want error msg %q", tc.data, tc.wantErrorMsg)
			} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
				t.Errorf("Unmarshal(0x%x) returned error msg %q, want %q", tc.data, err, tc.wantErrorMsg)
			}
		} else {
			if !tc.wantObj.Equal(tm) {
				t.Errorf("Unmarshal(0x%x) returned %s, want %s", tc.data, tm, tc.wantObj)
			}
		}
	}
}

func TestDecTimeTagOption(t *testing.T) {
	timeTagIgnoredDecMode, _ := DecOptions{TimeTag: DecTagIgnored}.DecMode()
	timeTagOptionalDecMode, _ := DecOptions{TimeTag: DecTagOptional}.DecMode()
	timeTagRequiredDecMode, _ := DecOptions{TimeTag: DecTagRequired}.DecMode()

	testCases := []struct {
		name            string
		cborRFC3339Time []byte
		cborUnixTime    []byte
		decMode         DecMode
		wantTime        time.Time
		wantErrorMsg    string
	}{
		// not-tagged time CBOR data
		{
			name:            "not-tagged data with DecTagIgnored option",
			cborRFC3339Time: mustHexDecode("74323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("1a514b67b0"),
			decMode:         timeTagIgnoredDecMode,
			wantTime:        mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "not-tagged data with timeTagOptionalDecMode option",
			cborRFC3339Time: mustHexDecode("74323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("1a514b67b0"),
			decMode:         timeTagOptionalDecMode,
			wantTime:        mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "not-tagged data with timeTagRequiredDecMode option",
			cborRFC3339Time: mustHexDecode("74323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("1a514b67b0"),
			decMode:         timeTagRequiredDecMode,
			wantErrorMsg:    "expect CBOR tag value",
		},
		// tagged time CBOR data
		{
			name:            "tagged data with timeTagIgnoredDecMode option",
			cborRFC3339Time: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("c11a514b67b0"),
			decMode:         timeTagIgnoredDecMode,
			wantTime:        mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "tagged data with timeTagOptionalDecMode option",
			cborRFC3339Time: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("c11a514b67b0"),
			decMode:         timeTagOptionalDecMode,
			wantTime:        mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "tagged data with timeTagRequiredDecMode option",
			cborRFC3339Time: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("c11a514b67b0"),
			decMode:         timeTagRequiredDecMode,
			wantTime:        mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		// mis-tagged time CBOR data
		{
			name:            "mis-tagged data with timeTagIgnoredDecMode option",
			cborRFC3339Time: mustHexDecode("c8c974323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("c8c91a514b67b0"),
			decMode:         timeTagIgnoredDecMode,
			wantTime:        mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
		},
		{
			name:            "mis-tagged data with timeTagOptionalDecMode option",
			cborRFC3339Time: mustHexDecode("c8c974323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("c8c91a514b67b0"),
			decMode:         timeTagOptionalDecMode,
			wantErrorMsg:    "cbor: wrong tag number for time.Time, got 8, expect 0 or 1",
		},
		{
			name:            "mis-tagged data with timeTagRequiredDecMode option",
			cborRFC3339Time: mustHexDecode("c8c974323031332d30332d32315432303a30343a30305a"),
			cborUnixTime:    mustHexDecode("c8c91a514b67b0"),
			decMode:         timeTagRequiredDecMode,
			wantErrorMsg:    "cbor: wrong tag number for time.Time, got 8, expect 0 or 1",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm := time.Now()
			err := unmarshalWithDM(t, tc.decMode, tc.cborRFC3339Time, &tm)
			if tc.wantErrorMsg != "" {
				if err == nil {
					t.Errorf("Unmarshal(0x%x) didn't return an error", tc.cborRFC3339Time)
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", tc.cborRFC3339Time, err.Error(), tc.wantErrorMsg)
				}
			} else {
				if !tc.wantTime.Equal(tm) {
					t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborRFC3339Time, tm, tm, tc.wantTime, tc.wantTime)
				}
			}

			tm = time.Now()
			err = unmarshalWithDM(t, tc.decMode, tc.cborUnixTime, &tm)
			if tc.wantErrorMsg != "" {
				if err == nil {
					t.Errorf("Unmarshal(0x%x) didn't return an error", tc.cborRFC3339Time)
				} else if !strings.Contains(err.Error(), tc.wantErrorMsg) {
					t.Errorf("Unmarshal(0x%x) returned error %q, want error containing %q", tc.cborRFC3339Time, err.Error(), tc.wantErrorMsg)
				}
			} else {
				if !tc.wantTime.Equal(tm) {
					t.Errorf("Unmarshal(0x%x) = %v (%T), want %v (%T)", tc.cborRFC3339Time, tm, tm, tc.wantTime, tc.wantTime)
				}
			}
		})
	}
}

func TestDecModeTimeTagToAny(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		data         []byte
		want         any
		wantErrorMsg string
	}{
		{
			name: "Unmarshal tag 0 data to time.Time when TimeTagToAny is not set",
			opts: DecOptions{},
			data: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
			want: time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC),
		},
		{
			name: "Unmarshal tag 1 data to time.Time when TimeTagToAny is not set",
			opts: DecOptions{},
			data: mustHexDecode("c11a514b67b0"),
			want: time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC),
		},
		{
			name: "Unmarshal tag 0 data to RFC3339 string when TimeTagToAny is set",
			opts: DecOptions{TimeTagToAny: TimeTagToRFC3339},
			data: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"),
			want: "2013-03-21T20:04:00Z",
		},
		{
			name: "Unmarshal tag 1 data to RFC3339 string when TimeTagToAny is set",
			opts: DecOptions{TimeTagToAny: TimeTagToRFC3339},
			data: mustHexDecode("c11a514b67b0"),
			want: "2013-03-21T20:04:00Z",
		},
		{
			name: "Unmarshal tag 0 data to RFC3339Nano string when TimeTagToAny is set",
			opts: DecOptions{TimeTagToAny: TimeTagToRFC3339Nano},
			data: mustHexDecode("c076323031332d30332d32315432303a30343a30302e355a"),
			want: "2013-03-21T20:04:00.5Z",
		},
		{
			name: "Unmarshal tag 1 data to RFC3339Nano string when TimeTagToAny is set",
			opts: DecOptions{TimeTagToAny: TimeTagToRFC3339Nano},
			data: mustHexDecode("c1fb41d452d9ec200000"),
			want: "2013-03-21T20:04:00.5Z",
		},
		{
			name:         "error under TimeTagToRFC3339 when tag 0 contains an invalid RFC3339 timestamp",
			opts:         DecOptions{TimeTagToAny: TimeTagToRFC3339},
			data:         mustHexDecode("c07731303030302D30332D32315432303A30343A30302E355A"), // 0("10000-03-21T20:04:00.5Z")
			wantErrorMsg: `cbor: cannot set 10000-03-21T20:04:00.5Z for time.Time: parsing time "10000-03-21T20:04:00.5Z" as "2006-01-02T15:04:05Z07:00": cannot parse "0-03-21T20:04:00.5Z" as "-"`,
		},
		{
			name:         "error under TimeTagToRFC3339Nano when tag 0 contains an invalid RFC3339 timestamp",
			opts:         DecOptions{TimeTagToAny: TimeTagToRFC3339Nano},
			data:         mustHexDecode("c07731303030302D30332D32315432303A30343A30302E355A"), // 0("10000-03-21T20:04:00.5Z")
			wantErrorMsg: `cbor: cannot set 10000-03-21T20:04:00.5Z for time.Time: parsing time "10000-03-21T20:04:00.5Z" as "2006-01-02T15:04:05Z07:00": cannot parse "0-03-21T20:04:00.5Z" as "-"`,
		},
		{
			name:         "error under TimeTagToRFC3339 when tag 1 represents a time that can't be represented by valid RFC3339",
			opts:         DecOptions{TimeTagToAny: TimeTagToRFC3339},
			data:         mustHexDecode("c11b0000003afff44181"), // 1(253402300801)
			wantErrorMsg: "cbor: decoded time cannot be represented in RFC3339 format: Time.MarshalText: year outside of range [0,9999]",
		},
		{
			name:         "error under TimeTagToRFC3339Nano when tag 1 represents a time that can't be represented by valid RFC3339",
			opts:         DecOptions{TimeTagToAny: TimeTagToRFC3339Nano},
			data:         mustHexDecode("c11b0000003afff44181"), // 1(253402300801)
			wantErrorMsg: "cbor: decoded time cannot be represented in RFC3339 format with sub-second precision: Time.MarshalText: year outside of range [0,9999]",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dm, err := tc.opts.DecMode()
			if err != nil {
				t.Fatal(err)
			}

			var got any
			if err := unmarshalWithDM(t, dm, tc.data, &got); err != nil {
				if tc.wantErrorMsg == "" {
					t.Fatalf("unexpected error: %v", err)
				} else if gotErrMessage := err.Error(); tc.wantErrorMsg != gotErrMessage {
					t.Fatalf("want error %q, got %q", tc.wantErrorMsg, gotErrMessage)
				}
			} else if tc.wantErrorMsg != "" {
				t.Fatalf("got nil error, want %q", tc.wantErrorMsg)
			}

			compareNonFloats(t, tc.data, got, tc.want)

		})
	}
}

func TestDecModeByteStringToTime(t *testing.T) {
	for _, tc := range []struct {
		name         string
		opts         DecOptions
		data         []byte
		want         time.Time
		wantErrorMsg string
	}{
		{
			name:         "Unmarshal byte string to time.Time when ByteStringToTime is not set",
			opts:         DecOptions{},
			data:         mustHexDecode("54323031332D30332D32315432303A30343A30305A"), // '2013-03-21T20:04:00Z'
			wantErrorMsg: "cbor: cannot unmarshal byte string into Go value of type time.Time",
		},
		{
			name: "Unmarshal byte string to time.Time when ByteStringToTime is set to ByteStringToTimeAllowed",
			opts: DecOptions{ByteStringToTime: ByteStringToTimeAllowed},
			data: mustHexDecode("54323031332D30332D32315432303A30343A30305A"), // '2013-03-21T20:04:00Z'
			want: time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC),
		},
		{
			name: "Unmarshal byte string to time.Time with nano when ByteStringToTime is set to ByteStringToTimeAllowed",
			opts: DecOptions{ByteStringToTime: ByteStringToTimeAllowed},
			data: mustHexDecode("56323031332D30332D32315432303A30343A30302E355A"), // '2013-03-21T20:04:00.5Z'
			want: time.Date(2013, 3, 21, 20, 4, 0, 500000000, time.UTC),
		},
		{
			name:         "Unmarshal a byte string that is not a valid RFC3339 timestamp to time.Time when ByteStringToTime is set to ByteStringToTimeAllowed",
			opts:         DecOptions{ByteStringToTime: ByteStringToTimeAllowed},
			data:         mustHexDecode("4B696E76616C696454657874"), // 'invalidText'
			wantErrorMsg: `cbor: cannot set "invalidText" for time.Time: parsing time "invalidText" as "2006-01-02T15:04:05Z07:00": cannot parse "invalidText" as "2006"`,
		},
		{
			name:         "Unmarshal a byte string that is not a valid utf8 sequence to time.Time when ByteStringToTime is set to ByteStringToTimeAllowed",
			opts:         DecOptions{ByteStringToTime: ByteStringToTimeAllowed},
			data:         mustHexDecode("54323031338030332D32315432303A30343A30305A"), // "2013\x8003-21T20:04:00Z" -- the first hyphen of a valid RFC3339 string is replaced by a continuation byte
			wantErrorMsg: `cbor: cannot set "2013\x8003-21T20:04:00Z" for time.Time: parsing time "2013\x8003-21T20:04:00Z" as "2006-01-02T15:04:05Z07:00": cannot parse "\x8003-21T20:04:00Z" as "-"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dm, err := tc.opts.DecMode()
			if err != nil {
				t.Fatal(err)
			}

			var got time.Time
			if err := unmarshalWithDM(t, dm, tc.data, &got); err != nil {
				if tc.wantErrorMsg != err.Error() {
					t.Errorf("unexpected error: got %v want %v", err, tc.wantErrorMsg)
				}
			} else {
				compareNonFloats(t, tc.data, got, tc.want)
			}
		})
	}
}

func TestParseToTimeConsumeDataOnError(t *testing.T) {
	elementTestCases := []struct {
		name string
		b    []byte
	}{
		{name: "byte string", b: mustHexDecode("4101")},                           // h'01'
		{name: "indefinite-length byte string", b: mustHexDecode("5f4101ff")},     // (_ h'01')
		{name: "array", b: mustHexDecode("8101")},                                 // [1]
		{name: "indefinite-length array", b: mustHexDecode("9f01ff")},             // [_ 1]
		{name: "map", b: mustHexDecode("a1616101")},                               // {"a": 1}
		{name: "indefinite-length map", b: mustHexDecode("bf616101ff")},           // {_ "a": 1}
		{name: "unregistered tag with map content", b: mustHexDecode("d903e8a0")}, // 1000({})
	}

	type structWithTime struct {
		Time time.Time `cbor:"d"`
		Z    int       `cbor:"z"`
	}

	for _, element := range elementTestCases {
		// Definite-length map: {"d": <value>, "z": 7}
		def := append([]byte{0xa2, 0x61, 'd'}, element.b...)
		def = append(def, 0x61, 'z', 0x07)

		// Indefinite-length map: {_ "d": <value>, "z": 7}
		indef := append([]byte{0xbf, 0x61, 'd'}, element.b...)
		indef = append(indef, 0x61, 'z', 0x07, 0xff)

		for _, container := range []struct {
			name string
			data []byte
		}{
			{name: "definite-length map", data: def},
			{name: "indefinite-length map", data: indef},
		} {
			t.Run(container.name+"/"+element.name, func(t *testing.T) {
				data := container.data

				var v structWithTime
				err := unmarshal(t, data, &v)

				var typeError *UnmarshalTypeError
				if !errors.As(err, &typeError) {
					t.Errorf("Unmarshal(0x%x) error = %v (%T), want *cbor.UnmarshalTypeError", data, err, err)
				}
				if v.Z != 7 {
					t.Errorf("Unmarshal(0x%x) Z = %d, want 7", data, v.Z)
				}
			})
		}
	}

	type structWithTimeToArray struct {
		_    struct{} `cbor:",toarray"`
		Time time.Time
		Z    int
	}

	for _, element := range elementTestCases {
		// Definite-length array: [<value>, 7]
		def := append([]byte{0x82}, element.b...)
		def = append(def, 0x07)

		// Indefinite-length array : [_ <value>, 7]
		indef := append([]byte{0x9f}, element.b...)
		indef = append(indef, 0x07, 0xff)

		for _, container := range []struct {
			name string
			data []byte
		}{
			{name: "definite-length array", data: def},
			{name: "indefinite-length array", data: indef},
		} {
			t.Run(container.name+"/"+element.name, func(t *testing.T) {
				data := container.data

				var v structWithTimeToArray
				err := unmarshal(t, data, &v)

				var typeError *UnmarshalTypeError
				if !errors.As(err, &typeError) {
					t.Errorf("Unmarshal(0x%x) error = %v (%T), want *cbor.UnmarshalTypeError", data, err, err)
				}
				if v.Z != 7 {
					t.Errorf("Unmarshal(0x%x) Z = %d, want 7", data, v.Z)
				}
			})
		}
	}
}

func TestStreamDecodeToTimeAndInt(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{name: "byte string + int", data: mustHexDecode("410107")}, // h'01', 7
		{name: "array + int", data: mustHexDecode("810107")},       // [1], 7
		{name: "map + int", data: mustHexDecode("a161610107")},     // {"a": 1}, 7
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dec := NewDecoder(bytes.NewReader(tc.data))

			var tm time.Time
			err := dec.Decode(&tm)

			var typeError *UnmarshalTypeError
			if !errors.As(err, &typeError) {
				t.Errorf("Decode(0x%x) first item error = %v (%T), want *cbor.UnmarshalTypeError", tc.data, err, err)
			}

			var v int
			if err := dec.Decode(&v); err != nil {
				t.Errorf("Decode(0x%x) second item error = %v, want nil", tc.data, err)
			}
			if v != 7 {
				t.Errorf("Decode(0x%x) second item = %d, want 7", tc.data, v)
			}
		})
	}
}

// Tests to marshal time.Time

type encTimeOption struct {
	name           string
	opt            EncOptions
	hasLocalOffset bool
}

var (
	timeUnixOpt = encTimeOption{
		name:           "TimeUnix",
		opt:            EncOptions{Time: TimeUnix},
		hasLocalOffset: false,
	}

	timeUnixWithTagOpt = encTimeOption{
		name:           "TimeUnixWithTag",
		opt:            EncOptions{Time: TimeUnix, TimeTag: EncTagRequired},
		hasLocalOffset: false,
	}

	timeUnixMicroOpt = encTimeOption{
		name:           "TimeUnixMicro",
		opt:            EncOptions{Time: TimeUnixMicro},
		hasLocalOffset: false,
	}

	timeUnixMicroWithTagOpt = encTimeOption{
		name:           "TimeUnixMicroWithTag",
		opt:            EncOptions{Time: TimeUnixMicro, TimeTag: EncTagRequired},
		hasLocalOffset: false,
	}

	timeUnixDynamicOpt = encTimeOption{
		name:           "TimeUnixDynamic",
		opt:            EncOptions{Time: TimeUnixDynamic},
		hasLocalOffset: false,
	}

	timeUnixDynamicWithTagOpt = encTimeOption{
		name:           "TimeUnixDynamicWithTag",
		opt:            EncOptions{Time: TimeUnixDynamic, TimeTag: EncTagRequired},
		hasLocalOffset: false,
	}

	timeRFC3339Opt = encTimeOption{
		name:           "TimeRFC3339",
		opt:            EncOptions{Time: TimeRFC3339},
		hasLocalOffset: true,
	}

	timeRFC3339WithTagOpt = encTimeOption{
		name:           "TimeRFC3339WithTag",
		opt:            EncOptions{Time: TimeRFC3339, TimeTag: EncTagRequired},
		hasLocalOffset: true,
	}

	timeRFC3339NanoOpt = encTimeOption{
		name:           "TimeRFC3339Nano",
		opt:            EncOptions{Time: TimeRFC3339Nano},
		hasLocalOffset: true,
	}

	timeRFC3339NanoWithTagOpt = encTimeOption{
		name:           "TimeRFC3339NanoWithTag",
		opt:            EncOptions{Time: TimeRFC3339Nano, TimeTag: EncTagRequired},
		hasLocalOffset: true,
	}

	timeRFC3339NanoUTCOpt = encTimeOption{
		name:           "TimeRFC3339NanoUTC",
		opt:            EncOptions{Time: TimeRFC3339NanoUTC},
		hasLocalOffset: false,
	}

	timeRFC3339NanoUTCWithTagOpt = encTimeOption{
		name:           "TimeRFC3339NanoUTCWithTag",
		opt:            EncOptions{Time: TimeRFC3339NanoUTC, TimeTag: EncTagRequired},
		hasLocalOffset: false,
	}
)

func TestEncodeTime(t *testing.T) {
	testMarshalWithFP(t, testEncodeTime)
}

func testEncodeTime(t *testing.T) {
	type conversion struct {
		opt          encTimeOption
		wantCborData []byte
		roundtrip    bool
	}
	testCases := []struct {
		name    string
		tm      time.Time
		convert []conversion
	}{
		{
			name: "zero time",
			tm:   time.Time{},
			convert: []conversion{
				{
					opt:          timeUnixOpt,
					wantCborData: mustHexDecode("f6"), // encode as CBOR null
					roundtrip:    true,
				},
				{
					opt:          timeUnixWithTagOpt,
					wantCborData: mustHexDecode("f6"), // encode as CBOR null
					roundtrip:    true,
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: mustHexDecode("f6"), // encode as CBOR null
					roundtrip:    true,
				},
				{
					opt:          timeUnixMicroWithTagOpt,
					wantCborData: mustHexDecode("f6"), // encode as CBOR null
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: mustHexDecode("f6"), // encode as CBOR null
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicWithTagOpt,
					wantCborData: mustHexDecode("f6"), // encode as CBOR null
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: mustHexDecode("f6"), // encode as CBOR null
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339WithTagOpt,
					wantCborData: mustHexDecode("f6"), // encode as CBOR null
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: mustHexDecode("f6"), // encode as CBOR null
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoWithTagOpt,
					wantCborData: mustHexDecode("f6"), // encode as CBOR null
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCOpt,
					wantCborData: mustHexDecode("f6"), // encode as CBOR null
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCWithTagOpt,
					wantCborData: mustHexDecode("f6"), // encode as CBOR null
					roundtrip:    true,
				},
			},
		},
		{
			name: "UTC time without fractional seconds",
			tm:   mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00Z"),
			convert: []conversion{
				{
					opt:          timeUnixOpt,
					wantCborData: mustHexDecode("1a514b67b0"), // 1363896240
					roundtrip:    true,
				},
				{
					opt:          timeUnixWithTagOpt,
					wantCborData: mustHexDecode("c11a514b67b0"), // 1(1363896240)
					roundtrip:    true,
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: mustHexDecode("fb41d452d9ec000000"), // 1363896240.0
					roundtrip:    true,
				},
				{
					opt:          timeUnixMicroWithTagOpt,
					wantCborData: mustHexDecode("c1fb41d452d9ec000000"), // 1(1363896240.0)
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: mustHexDecode("1a514b67b0"), // 1363896240
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicWithTagOpt,
					wantCborData: mustHexDecode("c11a514b67b0"), // 1(1363896240)
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: mustHexDecode("74323031332d30332d32315432303a30343a30305a"), // "2013-03-21T20:04:00Z"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339WithTagOpt,
					wantCborData: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"), // 0("2013-03-21T20:04:00Z")
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: mustHexDecode("74323031332d30332d32315432303a30343a30305a"), // "2013-03-21T20:04:00Z"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoWithTagOpt,
					wantCborData: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"), // 0("2013-03-21T20:04:00Z")
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCOpt,
					wantCborData: mustHexDecode("74323031332d30332d32315432303a30343a30305a"), // "2013-03-21T20:04:00Z"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCWithTagOpt,
					wantCborData: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"), // 0("2013-03-21T20:04:00Z")
					roundtrip:    true,
				},
			},
		},
		{
			name: "local time without fractional seconds",
			tm:   mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00-05:00"),
			convert: []conversion{
				{
					opt:          timeUnixOpt,
					wantCborData: mustHexDecode("1a514bae00"), // 1363914240
					roundtrip:    true,
				},
				{
					opt:          timeUnixWithTagOpt,
					wantCborData: mustHexDecode("c11a514bae00"), // 1(1363914240)
					roundtrip:    true,
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: mustHexDecode("fb41d452eb80000000"), // 1363914240.0
					roundtrip:    true,
				},
				{
					opt:          timeUnixMicroWithTagOpt,
					wantCborData: mustHexDecode("c1fb41d452eb80000000"), // 1(1363914240.0)
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: mustHexDecode("1a514bae00"), // 1363914240
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicWithTagOpt,
					wantCborData: mustHexDecode("c11a514bae00"), // 1(1363914240)
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: mustHexDecode("7819323031332d30332d32315432303a30343a30302d30353a3030"), // "2013-03-21T20:04:00-05:00"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339WithTagOpt,
					wantCborData: mustHexDecode("c07819323031332d30332d32315432303a30343a30302d30353a3030"), // 0("2013-03-21T20:04:00-05:00")
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: mustHexDecode("7819323031332d30332d32315432303a30343a30302d30353a3030"), // "2013-03-21T20:04:00-05:00"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoWithTagOpt,
					wantCborData: mustHexDecode("c07819323031332d30332d32315432303a30343a30302d30353a3030"), // 0("2013-03-21T20:04:00-05:00")
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCOpt,
					wantCborData: mustHexDecode("74323031332d30332d32325430313a30343a30305a"), // "2013-03-22T01:04:00Z"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCWithTagOpt,
					wantCborData: mustHexDecode("c074323031332d30332d32325430313a30343a30305a"), // 0("2013-03-22T01:04:00Z")
					roundtrip:    true,
				},
			},
		},
		{
			name: "UTC time with fractional seconds",
			tm:   mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00.5Z"),
			convert: []conversion{
				{
					opt:          timeUnixOpt,
					wantCborData: mustHexDecode("1a514b67b0"), // 1363896240
					roundtrip:    false,
				},
				{
					opt:          timeUnixWithTagOpt,
					wantCborData: mustHexDecode("c11a514b67b0"), // 1(1363896240)
					roundtrip:    false,
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: mustHexDecode("fb41d452d9ec200000"), // 1363896240.5
					roundtrip:    true,
				},
				{
					opt:          timeUnixMicroWithTagOpt,
					wantCborData: mustHexDecode("c1fb41d452d9ec200000"), // 1(1363896240.5)
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: mustHexDecode("fb41d452d9ec200000"), // 1363896240.5
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicWithTagOpt,
					wantCborData: mustHexDecode("c1fb41d452d9ec200000"), // 1(1363896240.5)
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: mustHexDecode("74323031332d30332d32315432303a30343a30305a"), // "2013-03-21T20:04:00Z"
					roundtrip:    false,
				},
				{
					opt:          timeRFC3339WithTagOpt,
					wantCborData: mustHexDecode("c074323031332d30332d32315432303a30343a30305a"), // 0("2013-03-21T20:04:00Z")
					roundtrip:    false,
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: mustHexDecode("76323031332d30332d32315432303a30343a30302e355a"), // "2013-03-21T20:04:00.5Z"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoWithTagOpt,
					wantCborData: mustHexDecode("c076323031332d30332d32315432303a30343a30302e355a"), // 0("2013-03-21T20:04:00.5Z")
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCOpt,
					wantCborData: mustHexDecode("76323031332d30332d32315432303a30343a30302e355a"), // "2013-03-21T20:04:00.5Z"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCWithTagOpt,
					wantCborData: mustHexDecode("c076323031332d30332d32315432303a30343a30302e355a"), // 0("2013-03-21T20:04:00.5Z")
					roundtrip:    true,
				},
			},
		},
		{
			name: "local time with fractional seconds",
			tm:   mustParseTime(time.RFC3339Nano, "2013-03-21T20:04:00.5-05:00"),
			convert: []conversion{
				{
					opt:          timeUnixOpt,
					wantCborData: mustHexDecode("1a514bae00"), // 1363914240
					roundtrip:    false,
				},
				{
					opt:          timeUnixWithTagOpt,
					wantCborData: mustHexDecode("c11a514bae00"), // 1(1363914240)
					roundtrip:    false,
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: mustHexDecode("fb41d452eb80200000"), // 1363914240.5
					roundtrip:    true,
				},
				{
					opt:          timeUnixMicroWithTagOpt,
					wantCborData: mustHexDecode("c1fb41d452eb80200000"), // 1(1363914240.5)
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: mustHexDecode("fb41d452eb80200000"), // 1363914240.5
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicWithTagOpt,
					wantCborData: mustHexDecode("c1fb41d452eb80200000"), // 1(1363914240.5)
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: mustHexDecode("7819323031332d30332d32315432303a30343a30302d30353a3030"), // "2013-03-21T20:04:00-05:00"
					roundtrip:    false,
				},
				{
					opt:          timeRFC3339WithTagOpt,
					wantCborData: mustHexDecode("c07819323031332d30332d32315432303a30343a30302d30353a3030"), // 0("2013-03-21T20:04:00-05:00")
					roundtrip:    false,
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: mustHexDecode("781b323031332d30332d32315432303a30343a30302e352d30353a3030"), // "2013-03-21T20:04:00.5-05:00"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoWithTagOpt,
					wantCborData: mustHexDecode("c0781b323031332d30332d32315432303a30343a30302e352d30353a3030"), // 0("2013-03-21T20:04:00.5-05:00")
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCOpt,
					wantCborData: mustHexDecode("76323031332d30332d32325430313a30343a30302e355a"), // "2013-03-22T01:04:00.5Z"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCWithTagOpt,
					wantCborData: mustHexDecode("c076323031332d30332d32325430313a30343a30302e355a"), // 0("2013-03-22T01:04:00.5Z")
					roundtrip:    true,
				},
			},
		},
		{
			name: "UTC time before January 1, 1970 UTC without fractional seconds",
			tm:   mustParseTime(time.RFC3339Nano, "1969-03-21T20:04:00Z"),
			convert: []conversion{
				{
					opt:          timeUnixOpt,
					wantCborData: mustHexDecode("3a0177f2cf"), // -24638160
					roundtrip:    true,
				},
				{
					opt:          timeUnixWithTagOpt,
					wantCborData: mustHexDecode("c13a0177f2cf"), // 1(-24638160)
					roundtrip:    true,
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: mustHexDecode("fbc1777f2d00000000"), // -24638160.0
					roundtrip:    true,
				},
				{
					opt:          timeUnixMicroWithTagOpt,
					wantCborData: mustHexDecode("c1fbc1777f2d00000000"), // 1(-24638160.0)
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: mustHexDecode("3a0177f2cf"), // -24638160
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicWithTagOpt,
					wantCborData: mustHexDecode("c13a0177f2cf"), // 1(-24638160)
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: mustHexDecode("74313936392d30332d32315432303a30343a30305a"), // "1969-03-21T20:04:00Z"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339WithTagOpt,
					wantCborData: mustHexDecode("c074313936392d30332d32315432303a30343a30305a"), // 0("1969-03-21T20:04:00Z")
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: mustHexDecode("74313936392d30332d32315432303a30343a30305a"), // "1969-03-21T20:04:00Z"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoWithTagOpt,
					wantCborData: mustHexDecode("c074313936392d30332d32315432303a30343a30305a"), // 0("1969-03-21T20:04:00Z")
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCOpt,
					wantCborData: mustHexDecode("74313936392d30332d32315432303a30343a30305a"), // "1969-03-21T20:04:00Z"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCWithTagOpt,
					wantCborData: mustHexDecode("c074313936392d30332d32315432303a30343a30305a"), // 0("1969-03-21T20:04:00Z")
					roundtrip:    true,
				},
			},
		},
		{
			name: "local time before January 1, 1970 UTC without fractional seconds",
			tm:   mustParseTime(time.RFC3339Nano, "1969-03-21T20:04:00-05:00"),
			convert: []conversion{
				{
					opt:          timeUnixOpt,
					wantCborData: mustHexDecode("3a0177ac7f"), // -24620160
					roundtrip:    true,
				},
				{
					opt:          timeUnixWithTagOpt,
					wantCborData: mustHexDecode("c13a0177ac7f"), // 1(-24620160)
					roundtrip:    true,
				},
				{
					opt:          timeUnixMicroOpt,
					wantCborData: mustHexDecode("fbc1777ac800000000"), // -24620160.0
					roundtrip:    true,
				},
				{
					opt:          timeUnixMicroWithTagOpt,
					wantCborData: mustHexDecode("c1fbc1777ac800000000"), // 1(-24620160.0)
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicOpt,
					wantCborData: mustHexDecode("3a0177ac7f"), // -24620160
					roundtrip:    true,
				},
				{
					opt:          timeUnixDynamicWithTagOpt,
					wantCborData: mustHexDecode("c13a0177ac7f"), // 1(-24620160)
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339Opt,
					wantCborData: mustHexDecode("7819313936392d30332d32315432303a30343a30302d30353a3030"), // "1969-03-21T20:04:00-05:00"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339WithTagOpt,
					wantCborData: mustHexDecode("c07819313936392d30332d32315432303a30343a30302d30353a3030"), // 0("1969-03-21T20:04:00-05:00")
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoOpt,
					wantCborData: mustHexDecode("7819313936392d30332d32315432303a30343a30302d30353a3030"), // "1969-03-21T20:04:00-05:00"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoWithTagOpt,
					wantCborData: mustHexDecode("c07819313936392d30332d32315432303a30343a30302d30353a3030"), // 0("1969-03-21T20:04:00-05:00")
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCOpt,
					wantCborData: mustHexDecode("74313936392d30332d32325430313a30343a30305a"), // "1969-03-22T01:04:00Z"
					roundtrip:    true,
				},
				{
					opt:          timeRFC3339NanoUTCWithTagOpt,
					wantCborData: mustHexDecode("c074313936392d30332d32325430313a30343a30305a"), // 0("1969-03-22T01:04:00Z")
					roundtrip:    true,
				},
			},
		},
	}

	type s struct{ T time.Time }
	wantStructPrefix := []byte{0xa1, 0x61, 0x54}

	for _, tc := range testCases {
		for _, convert := range tc.convert {
			name := tc.name + " with " + convert.opt.name + " option"
			t.Run(name, func(t *testing.T) {
				em, err := convert.opt.opt.EncMode()
				if err != nil {
					t.Errorf("EncMode() returned error %v", err)
				}

				// Marshal time.Time
				b, err := em.Marshal(tc.tm)
				checkMarshal(t, "em.Marshal", b, err, convert.wantCborData)

				// Marshal *struct{time.Time}
				v := s{T: tc.tm}
				want := append([]byte(nil), wantStructPrefix...)
				want = append(want, convert.wantCborData...)
				b, err = em.Marshal(&v)
				checkMarshal(t, "em.Marshal", b, err, want)

				var tm time.Time
				err = unmarshal(t, convert.wantCborData, &tm)
				if err != nil {
					t.Errorf("Unmarshal(0x%x) to time.Time returned error %v", convert.wantCborData, err)
				} else if convert.roundtrip {
					if convert.opt.hasLocalOffset && tc.tm.Compare(tm) != 0 {
						t.Errorf("failed to rountrip local time: want %s, got %s", tc.tm, tm)
					}
					if !convert.opt.hasLocalOffset && tc.tm.UTC().Compare(tm.UTC()) != 0 {
						t.Errorf("failed to roundtrip UTC time: want %s, got %s", tc.tm, tm)
					}
				}
			})
		}
	}
}

func TestEncodeTimeUnixMicro(t *testing.T) {
	testMarshalWithFP(t, testEncodeTimeUnixMicro)
}

func testEncodeTimeUnixMicro(t *testing.T) {
	em, err := EncOptions{Time: TimeUnixMicro}.EncMode()
	if err != nil {
		t.Fatalf("EncMode() returned error %v", err)
	}

	testCases := []struct {
		name          string
		tm            time.Time
		wantData      []byte
		skipRoundTrip bool
	}{
		// Unix seconds within [minUnixNanoSecs, maxUnixNanoSecs].
		{
			name:     "1677-09-21T00:12:44Z",
			tm:       time.Unix(math.MinInt64/1_000_000_000, 0),
			wantData: mustHexDecode("fbc2012e0be8200000"),
		},
		{
			name:          "1677-09-21T00:12:44.000014Z",
			tm:            time.Date(1677, 9, 21, 0, 12, 44, 14000, time.UTC),
			wantData:      mustHexDecode("fbc2012e0be81ffff8"),
			skipRoundTrip: true,
		},
		{
			name:     "1677-09-21T00:12:44.5Z",
			tm:       time.Date(1677, 9, 21, 0, 12, 44, 500000000, time.UTC),
			wantData: mustHexDecode("fbc2012e0be81c0000"),
		},
		{
			name:     "1677-09-21T00:12:45.25Z",
			tm:       time.Date(1677, 9, 21, 0, 12, 45, 250000000, time.UTC),
			wantData: mustHexDecode("fbc2012e0be8160000"),
		},
		{
			name:     "1900-01-01T00:00:00.25Z",
			tm:       time.Date(1900, 1, 1, 0, 0, 0, 250000000, time.UTC),
			wantData: mustHexDecode("fbc1e0754fcff80000"),
		},
		{
			name:     "1969-12-31T23:59:59.002137Z",
			tm:       time.Date(1969, 12, 31, 23, 59, 59, 2137000, time.UTC),
			wantData: mustHexDecode("fbbfefee7e62dc6e2b"),
		},
		{
			name:     "1969-12-31T23:59:59.5Z",
			tm:       time.Date(1969, 12, 31, 23, 59, 59, 500000000, time.UTC),
			wantData: mustHexDecode("fbbfe0000000000000"),
		},
		{
			name:          "2013-03-21T20:04:00.25Z",
			tm:            time.Date(2013, 3, 21, 20, 4, 0, 250000000, time.UTC),
			wantData:      mustHexDecode("fb41d452d9ec0fffff"),
			skipRoundTrip: true,
		},
		{
			name:     "2026-09-24T12:00:00.5Z",
			tm:       time.Date(2026, 9, 24, 12, 0, 0, 500000000, time.UTC),
			wantData: mustHexDecode("fb41daad4430200000"),
		},
		{
			name:     "2026-11-01T00:04:58.5Z",
			tm:       time.Date(2026, 11, 1, 0, 4, 58, 500000000, time.UTC),
			wantData: mustHexDecode("fb41dab9a08aa00000"),
		},
		{
			name:     "2262-04-11T23:47:15Z",
			tm:       time.Unix(math.MaxInt64/1_000_000_000-1, 0),
			wantData: mustHexDecode("fb42012e0be8180000"),
		},
		{
			name:          "2262-04-11T23:47:15.000001Z",
			tm:            time.Date(2262, 4, 11, 23, 47, 15, 1000, time.UTC),
			wantData:      mustHexDecode("fb42012e0be8180000"),
			skipRoundTrip: true,
		},

		// Unix seconds outside [minUnixNanoSecs, maxUnixNanoSecs].
		{
			name:     "0001-01-01T00:00:01Z",
			tm:       time.Date(1, 1, 1, 0, 0, 1, 0, time.UTC),
			wantData: mustHexDecode("fbc22cef23edfe0000"),
		},
		{
			name:     "1000-01-01T00:00:00.25Z",
			tm:       time.Date(1000, 1, 1, 0, 0, 0, 250000000, time.UTC),
			wantData: mustHexDecode("fbc21c820bcdff0000"),
		},
		{
			name:     "1677-09-21T00:12:43Z",
			tm:       time.Date(1677, 9, 21, 0, 12, 43, 0, time.UTC),
			wantData: mustHexDecode("fbc2012e0be8280000"),
		},
		{
			name:     "2262-04-11T23:47:16.5Z",
			tm:       time.Date(2262, 4, 11, 23, 47, 16, 500000000, time.UTC),
			wantData: mustHexDecode("fb42012e0be8240000"),
		},
		{
			name:     "2262-04-11T23:47:16.875Z",
			tm:       time.Date(2262, 4, 11, 23, 47, 16, 875000000, time.UTC),
			wantData: mustHexDecode("fb42012e0be8270000"),
		},
		{
			name:     "2262-04-11T23:47:17Z",
			tm:       time.Date(2262, 4, 11, 23, 47, 17, 0, time.UTC),
			wantData: mustHexDecode("fb42012e0be8280000"),
		},
		{
			name:     "9999-12-31T23:59:59.5Z",
			tm:       time.Date(9999, 12, 31, 23, 59, 59, 500000000, time.UTC),
			wantData: mustHexDecode("fb424d7ffa20bfc000"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := em.Marshal(tc.tm)
			checkMarshal(t, "em.Marshal", b, err, tc.wantData)

			if !tc.skipRoundTrip {
				var got time.Time
				if err := Unmarshal(b, &got); err != nil {
					t.Errorf("Unmarshal(0x%x) returned error %v", b, err)
				} else if !got.Equal(tc.tm) {
					t.Errorf("Unmarshal(0x%x) = %v, want %v", b, got.UTC(), tc.tm)
				}
			}
		})
	}

	// Encoded values keep time order across the int64 Unix nanoseconds range limits
	// and across the boundaries where the encoding formula changes.
	for _, center := range []time.Time{
		time.Unix(0, math.MinInt64).UTC(),
		time.Unix(0, math.MaxInt64).UTC(),
		time.Unix(minUnixNanoSecs, 0).UTC(),
		time.Unix(maxUnixNanoSecs+1, 0).UTC(),
	} {
		prev := math.Inf(-1)
		for d := -2 * time.Millisecond; d <= 2*time.Millisecond; d += time.Microsecond {
			tm := center.Add(d)
			b, err := em.Marshal(tm)
			if err != nil {
				t.Fatalf("Marshal(%v) returned error %v", tm, err)
			}
			var f float64
			if err := Unmarshal(b, &f); err != nil {
				t.Fatalf("Unmarshal(0x%x) returned error %v", b, err)
			}
			if f < prev {
				t.Fatalf("Marshal(%v) encodes %v, want >= %v (encoded value of the previous time)", tm, f, prev)
			}
			prev = f
		}
	}
}

func TestEncodeTimeRFC3339Limits(t *testing.T) {
	testMarshalWithFP(t, testEncodeTimeRFC3339Limits)
}

func testEncodeTimeRFC3339Limits(t *testing.T) {
	timeModeNames := map[TimeMode]string{
		TimeRFC3339:        "TimeRFC3339",
		TimeRFC3339Nano:    "TimeRFC3339Nano",
		TimeRFC3339NanoUTC: "TimeRFC3339NanoUTC",
	}

	encTagModeNames := map[EncTagMode]string{
		EncTagNone:     "EncTagNone",
		EncTagRequired: "EncTagRequired",
	}

	const yearRangeErrorMsg = "cbor: unsupported value: time.Time year outside of range [0,9999] for RFC 3339"
	const offsetRangeErrorMsg = "cbor: unsupported value: time.Time UTC offset outside of range [-23:59,+23:59] for RFC 3339"

	type want struct {
		opt          TimeMode
		wantData     []byte
		wantErrorMsg string
	}

	testCases := []struct {
		name  string
		tm    time.Time
		wants []want
	}{
		// Years
		{
			name: "0000-01-01T00:00:00Z",
			tm:   time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC),
			wants: []want{
				{opt: TimeRFC3339, wantData: []byte("\x740000-01-01T00:00:00Z")},
				{opt: TimeRFC3339Nano, wantData: []byte("\x740000-01-01T00:00:00Z")},
				{opt: TimeRFC3339NanoUTC, wantData: []byte("\x740000-01-01T00:00:00Z")},
			},
		},
		{
			name: "9999-12-31T23:59:59.999999999Z",
			tm:   time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC),
			wants: []want{
				{opt: TimeRFC3339, wantData: []byte("\x749999-12-31T23:59:59Z")},
				{opt: TimeRFC3339Nano, wantData: []byte("\x78\x1e9999-12-31T23:59:59.999999999Z")},
				{opt: TimeRFC3339NanoUTC, wantData: []byte("\x78\x1e9999-12-31T23:59:59.999999999Z")},
			},
		},
		{
			name: "10000-01-01T00:00:00Z",
			tm:   time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC),
			wants: []want{
				{opt: TimeRFC3339, wantErrorMsg: yearRangeErrorMsg},
				{opt: TimeRFC3339Nano, wantErrorMsg: yearRangeErrorMsg},
				{opt: TimeRFC3339NanoUTC, wantErrorMsg: yearRangeErrorMsg},
			},
		},
		{
			name: "-0001-12-31T23:59:59.999999999Z",
			tm:   time.Date(-1, 12, 31, 23, 59, 59, 999999999, time.UTC),
			wants: []want{
				{opt: TimeRFC3339, wantErrorMsg: yearRangeErrorMsg},
				{opt: TimeRFC3339Nano, wantErrorMsg: yearRangeErrorMsg},
				{opt: TimeRFC3339NanoUTC, wantErrorMsg: yearRangeErrorMsg},
			},
		},
		{
			// Local year is 9999, but UTC year is 10000.
			name: "9999-12-31T23:00:00-02:00",
			tm:   time.Date(9999, 12, 31, 23, 0, 0, 0, time.FixedZone("", -2*60*60)),
			wants: []want{
				{opt: TimeRFC3339, wantData: []byte("\x78\x199999-12-31T23:00:00-02:00")},
				{opt: TimeRFC3339Nano, wantData: []byte("\x78\x199999-12-31T23:00:00-02:00")},
				{opt: TimeRFC3339NanoUTC, wantErrorMsg: yearRangeErrorMsg},
			},
		},
		{
			// Local year is 0, but UTC year is -1.
			name: "0000-01-01T01:00:00+02:00",
			tm:   time.Date(0, 1, 1, 1, 0, 0, 0, time.FixedZone("", 2*60*60)),
			wants: []want{
				{opt: TimeRFC3339, wantData: []byte("\x78\x190000-01-01T01:00:00+02:00")},
				{opt: TimeRFC3339Nano, wantData: []byte("\x78\x190000-01-01T01:00:00+02:00")},
				{opt: TimeRFC3339NanoUTC, wantErrorMsg: yearRangeErrorMsg},
			},
		},

		// UTC offsets
		{
			name: "2026-10-18T12:00:00+23:59",
			tm:   time.Date(2026, 10, 18, 12, 0, 0, 0, time.FixedZone("", 23*60*60+59*60)),
			wants: []want{
				{opt: TimeRFC3339, wantData: []byte("\x78\x192026-10-18T12:00:00+23:59")},
				{opt: TimeRFC3339Nano, wantData: []byte("\x78\x192026-10-18T12:00:00+23:59")},
				{opt: TimeRFC3339NanoUTC, wantData: []byte("\x742026-10-17T12:01:00Z")},
			},
		},
		{
			name: "2026-10-18T12:00:00-23:59",
			tm:   time.Date(2026, 10, 18, 12, 0, 0, 0, time.FixedZone("", -(23*60*60+59*60))),
			wants: []want{
				{opt: TimeRFC3339, wantData: []byte("\x78\x192026-10-18T12:00:00-23:59")},
				{opt: TimeRFC3339Nano, wantData: []byte("\x78\x192026-10-18T12:00:00-23:59")},
				{opt: TimeRFC3339NanoUTC, wantData: []byte("\x742026-10-19T11:59:00Z")},
			},
		},
		{
			name: "2026-10-18T12:00:00+24:00",
			tm:   time.Date(2026, 10, 18, 12, 0, 0, 0, time.FixedZone("", 24*60*60)),
			wants: []want{
				{opt: TimeRFC3339, wantErrorMsg: offsetRangeErrorMsg},
				{opt: TimeRFC3339Nano, wantErrorMsg: offsetRangeErrorMsg},
				{opt: TimeRFC3339NanoUTC, wantData: []byte("\x742026-10-17T12:00:00Z")},
			},
		},
		{
			name: "2026-10-18T12:00:00-24:00",
			tm:   time.Date(2026, 10, 18, 12, 0, 0, 0, time.FixedZone("", -24*60*60)),
			wants: []want{
				{opt: TimeRFC3339, wantErrorMsg: offsetRangeErrorMsg},
				{opt: TimeRFC3339Nano, wantErrorMsg: offsetRangeErrorMsg},
				{opt: TimeRFC3339NanoUTC, wantData: []byte("\x742026-10-19T12:00:00Z")},
			},
		},
		{
			name: "2026-10-18T12:00:00-00:30",
			tm:   time.Date(2026, 10, 18, 12, 0, 0, 0, time.FixedZone("", -30*60)),
			wants: []want{
				{opt: TimeRFC3339, wantData: []byte("\x78\x192026-10-18T12:00:00-00:30")},
				{opt: TimeRFC3339Nano, wantData: []byte("\x78\x192026-10-18T12:00:00-00:30")},
				{opt: TimeRFC3339NanoUTC, wantData: []byte("\x742026-10-18T12:30:00Z")},
			},
		},
		{
			name: "2026-10-18T12:00:00+29:59",
			tm:   time.Date(2026, 10, 18, 12, 0, 0, 0, time.FixedZone("", 29*60*60+59*60)),
			wants: []want{
				{opt: TimeRFC3339, wantErrorMsg: offsetRangeErrorMsg},
				{opt: TimeRFC3339Nano, wantErrorMsg: offsetRangeErrorMsg},
				{opt: TimeRFC3339NanoUTC, wantData: []byte("\x742026-10-17T06:01:00Z")},
			},
		},
		{
			name: "2026-10-18T12:00:00+30:00",
			tm:   time.Date(2026, 10, 18, 12, 0, 0, 0, time.FixedZone("", 30*60*60)),
			wants: []want{
				{opt: TimeRFC3339, wantErrorMsg: offsetRangeErrorMsg},
				{opt: TimeRFC3339Nano, wantErrorMsg: offsetRangeErrorMsg},
				{opt: TimeRFC3339NanoUTC, wantData: []byte("\x742026-10-17T06:00:00Z")},
			},
		},
		{
			name: "2026-10-18T12:00:00+100:00",
			tm:   time.Date(2026, 10, 18, 12, 0, 0, 0, time.FixedZone("", 100*60*60)),
			wants: []want{
				{opt: TimeRFC3339, wantErrorMsg: offsetRangeErrorMsg},
				{opt: TimeRFC3339Nano, wantErrorMsg: offsetRangeErrorMsg},
				{opt: TimeRFC3339NanoUTC, wantData: []byte("\x742026-10-14T08:00:00Z")},
			},
		},
	}

	for _, tc := range testCases {
		for _, w := range tc.wants {
			for _, tagMode := range []EncTagMode{EncTagNone, EncTagRequired} {
				name := tc.name + " with " + timeModeNames[w.opt] + " and " + encTagModeNames[tagMode] + " options"
				wantData := w.wantData
				if tagMode == EncTagRequired {
					wantData = append([]byte{0xc0}, w.wantData...)
				}
				t.Run(name, func(t *testing.T) {
					em, err := EncOptions{Time: w.opt, TimeTag: tagMode}.EncMode()
					if err != nil {
						t.Fatalf("EncMode() error = %v", err)
					}
					b, err := em.Marshal(tc.tm)
					checkMarshalCase[*UnsupportedValueError](t, "em.Marshal", b, err, wantData, w.wantErrorMsg != "", w.wantErrorMsg)
				})
			}
		}
	}
}
