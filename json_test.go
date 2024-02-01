package cbor_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestJSONInteroperability(t *testing.T) {
	diagMode, err := cbor.DiagOptions{
		ByteStringText: true,
	}.DiagMode()
	if err != nil {
		t.Fatal(err)
	}
	encMode, err := cbor.EncOptions{
		ByteSlice: cbor.ByteSliceToByteStringWithExpectedConversionToBase64,
		String:    cbor.StringToByteString,
		FieldName: cbor.FieldNameToByteString,
		ByteArray: cbor.ByteArrayToArray,
	}.EncMode()
	if err != nil {
		t.Fatal(err)
	}
	decMode, err := cbor.DecOptions{
		DefaultMapType:        reflect.TypeOf(map[string]interface{}(nil)),
		DefaultByteStringType: reflect.TypeOf(""),
		ByteStringToString:    cbor.ByteStringToStringAllowed,
		FieldNameByteString:   cbor.FieldNameByteStringAllowed,
		TextConversions: func(dt reflect.Type) cbor.TextConversionMode {
			switch {
			case dt.Kind() == reflect.String:
				return cbor.TextConversionEncodeToText
			case dt.Kind() == reflect.Slice && dt.Elem().Kind() == reflect.Uint8:
				return cbor.TextConversionDecodeFromText
			default:
				return cbor.TextConversionNone
			}
		},
		DefaultTextEncoding: cbor.TextEncodingBase64,
	}.DecMode()
	if err != nil {
		t.Fatal(err)
	}

	type S struct {
		Bytes []byte  `json:"bytes"`
		Arr   [5]byte `json:"arr"`
	}

	original := S{
		Bytes: []byte("hello world"),
		Arr:   [5]byte{'h', 'e', 'l', 'l', 'o'},
	}

	t.Logf("original: %#v", original)

	j1, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("original to json: %s", string(j1))

	c1, err := encMode.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	diag1, err := diagMode.Diagnose(c1)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("original to cbor: %s", diag1)

	var jintf interface{}
	err = json.Unmarshal(j1, &jintf)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("json to interface{}: %#v", jintf)

	var cintf interface{}
	err = decMode.Unmarshal(c1, &cintf)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("cbor to interface{}: %#v", cintf)

	j2, err := json.Marshal(jintf)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("interface{} to json: %s", string(j2))

	c2, err := encMode.Marshal(cintf)
	if err != nil {
		t.Fatal(err)
	}
	diag2, err := diagMode.Diagnose(c2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("interface{} to cbor: %s", diag2)

	if !reflect.DeepEqual(jintf, cintf) {
		// expected: encoding/json decodes the array elements to float64
		t.Logf("native-to-interface{} via cbor differed from native-to-interface{} via json")
	}

	var jfinal S
	err = json.Unmarshal(j2, &jfinal)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("json to native: %#v", jfinal)
	if !reflect.DeepEqual(original, jfinal) {
		t.Error("diff in json roundtrip")
	}

	var cfinal S
	err = decMode.Unmarshal(c2, &cfinal)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("cbor to native: %#v", cfinal)
	if !reflect.DeepEqual(original, cfinal) {
		t.Error("diff in cbor roundtrip")
	}
}
