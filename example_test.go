// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

package cbor_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/fxamacker/cbor"
)

func ExampleMarshal() {
	type Animal struct {
		Age    int
		Name   string
		Owners []string
		Male   bool
	}
	animal := Animal{Age: 4, Name: "Candy", Owners: []string{"Mary", "Joe"}}
	b, err := cbor.Marshal(animal, cbor.EncOptions{})
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%0x\n", b)
	// Output:
	// a46341676504644e616d656543616e6479664f776e65727382644d617279634a6f65644d616c65f4
}

func ExampleMarshal_time() {
	tm, _ := time.Parse(time.RFC3339Nano, "2013-03-21T20:04:00Z")
	// Encode time as string in RFC3339 format.
	b, err := cbor.Marshal(tm, cbor.EncOptions{TimeRFC3339: true})
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%0x\n", b)
	// Encode time as numerical representation of seconds since January 1, 1970 UTC.
	b, err = cbor.Marshal(tm, cbor.EncOptions{TimeRFC3339: false})
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%0x\n", b)
	// Output:
	// 74323031332d30332d32315432303a30343a30305a
	// 1a514b67b0
}

// This example uses Marshal to encode struct and map in canonical form.
func ExampleMarshal_structAndMapCanonical() {
	type Animal struct {
		Age      int
		Name     string
		Contacts map[string]string
		Male     bool
	}
	animal := Animal{Age: 4, Name: "Candy", Contacts: map[string]string{"Mary": "111-111-1111", "Joe": "222-222-2222"}}
	b, err := cbor.Marshal(animal, cbor.EncOptions{Canonical: true})
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%0x\n", b)
	// Output:
	// a46341676504644d616c65f4644e616d656543616e647968436f6e7461637473a2634a6f656c3232322d3232322d32323232644d6172796c3131312d3131312d31313131
}

func ExampleUnmarshal() {
	type Animal struct {
		Age    int
		Name   string
		Owners []string
		Male   bool
	}
	cborHex := "a46341676504644e616d656543616e6479664f776e65727382644d617279634a6f65644d616c65f4"
	cborData, _ := hex.DecodeString(cborHex)
	var animal Animal
	err := cbor.Unmarshal(cborData, &animal)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", animal)
	// Output:
	// {Age:4 Name:Candy Owners:[Mary Joe] Male:false}
}

func ExampleUnmarshal_time() {
	cborRFC3339Time := hexDecode("74323031332d30332d32315432303a30343a30305a")
	cborUnixTime := hexDecode("1a514b67b0")
	tm := time.Time{}
	if err := cbor.Unmarshal(cborRFC3339Time, &tm); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v\n", tm.UTC().Format(time.RFC3339Nano))
	tm = time.Time{}
	if err := cbor.Unmarshal(cborUnixTime, &tm); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v\n", tm.UTC().Format(time.RFC3339Nano))
	// Output:
	// 2013-03-21T20:04:00Z
	// 2013-03-21T20:04:00Z
}

func ExampleEncoder() {
	type Animal struct {
		Age    int
		Name   string
		Owners []string
		Male   bool
	}
	animals := []Animal{
		{Age: 4, Name: "Candy", Owners: []string{"Mary", "Joe"}, Male: false},
		{Age: 6, Name: "Rudy", Owners: []string{"Cindy"}, Male: true},
		{Age: 2, Name: "Duke", Owners: []string{"Norton"}, Male: true},
	}
	var buf bytes.Buffer
	enc := cbor.NewEncoder(&buf, cbor.EncOptions{Canonical: true})
	for _, animal := range animals {
		err := enc.Encode(animal)
		if err != nil {
			fmt.Println("error:", err)
		}
	}
	fmt.Printf("%0x\n", buf.Bytes())
	// Output:
	// a46341676504644d616c65f4644e616d656543616e6479664f776e65727382644d617279634a6f65a46341676506644d616c65f5644e616d656452756479664f776e657273816543696e6479a46341676502644d616c65f5644e616d656444756b65664f776e65727381664e6f72746f6e
}

func ExampleDecoder() {
	type Animal struct {
		Age    int
		Name   string
		Owners []string
		Male   bool
	}
	cborHex := "a46341676504644d616c65f4644e616d656543616e6479664f776e65727382644d617279634a6f65a46341676506644d616c65f5644e616d656452756479664f776e657273816543696e6479a46341676502644d616c65f5644e616d656444756b65664f776e65727381664e6f72746f6e"
	cborData, _ := hex.DecodeString(cborHex)
	dec := cbor.NewDecoder(bytes.NewReader(cborData))
	for {
		var animal Animal
		if err := dec.Decode(&animal); err != nil {
			if err != io.EOF {
				fmt.Println("error:", err)
			}
			break
		}
		fmt.Printf("%+v\n", animal)
	}
	// Output:
	// {Age:4 Name:Candy Owners:[Mary Joe] Male:false}
	// {Age:6 Name:Rudy Owners:[Cindy] Male:true}
	// {Age:2 Name:Duke Owners:[Norton] Male:true}
}

// ExampleEncoder_indefiniteLengthByteString encodes a stream of definite
// length byte string ("chunks") as an indefinite length byte string.
func ExampleEncoder_indefiniteLengthByteString() {
	var buf bytes.Buffer
	encoder := cbor.NewEncoder(&buf, cbor.EncOptions{})
	// Start indefinite length byte string encoding.
	if err := encoder.StartIndefiniteByteString(); err != nil {
		fmt.Println("error:", err)
	}
	// Encode definite length byte string.
	if err := encoder.Encode([]byte{1, 2}); err != nil {
		fmt.Println("error:", err)
	}
	// Encode definite length byte string.
	if err := encoder.Encode([3]byte{3, 4, 5}); err != nil {
		fmt.Println("error:", err)
	}
	// Close indefinite length byte string.
	if err := encoder.EndIndefinite(); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%0x\n", buf.Bytes())
	// Output:
	// 5f42010243030405ff
}

// ExampleEncoder_indefiniteLengthTextString encodes a stream of definite
// length text string ("chunks") as an indefinite length text string.
func ExampleEncoder_indefiniteLengthTextString() {
	var buf bytes.Buffer
	encoder := cbor.NewEncoder(&buf, cbor.EncOptions{})
	// Start indefinite length text string encoding.
	if err := encoder.StartIndefiniteTextString(); err != nil {
		fmt.Println("error:", err)
	}
	// Encode definite length text string.
	if err := encoder.Encode("strea"); err != nil {
		fmt.Println("error:", err)
	}
	// Encode definite length text string.
	if err := encoder.Encode("ming"); err != nil {
		fmt.Println("error:", err)
	}
	// Close indefinite length text string.
	if err := encoder.EndIndefinite(); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%0x\n", buf.Bytes())
	// Output:
	// 7f657374726561646d696e67ff
}

// ExampleEncoder_indefiniteLengthArray encodes a stream of elements as an
// indefinite length array.  Encoder supports nested indefinite length values.
func ExampleEncoder_indefiniteLengthArray() {
	var buf bytes.Buffer
	enc := cbor.NewEncoder(&buf, cbor.EncOptions{})
	// Start indefinite length array encoding.
	if err := enc.StartIndefiniteArray(); err != nil {
		fmt.Println("error:", err)
	}
	// Encode array element.
	if err := enc.Encode(1); err != nil {
		fmt.Println("error:", err)
	}
	// Encode array element.
	if err := enc.Encode([]int{2, 3}); err != nil {
		fmt.Println("error:", err)
	}
	// Start a nested indefinite length array as array element.
	if err := enc.StartIndefiniteArray(); err != nil {
		fmt.Println("error:", err)
	}
	// Encode nested array element.
	if err := enc.Encode(4); err != nil {
		fmt.Println("error:", err)
	}
	// Encode nested array element.
	if err := enc.Encode(5); err != nil {
		fmt.Println("error:", err)
	}
	// Close nested indefinite length array.
	if err := enc.EndIndefinite(); err != nil {
		fmt.Println("error:", err)
	}
	// Close outer indefinite length array.
	if err := enc.EndIndefinite(); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%0x\n", buf.Bytes())
	// Output:
	// 9f018202039f0405ffff
}

// ExampleEncoder_indefiniteLengthMap encodes a stream of elements as an
// indefinite length map.  Encoder supports nested indefinite length values.
func ExampleEncoder_indefiniteLengthMap() {
	var buf bytes.Buffer
	enc := cbor.NewEncoder(&buf, cbor.EncOptions{Canonical: true})
	// Start indefinite length map encoding.
	if err := enc.StartIndefiniteMap(); err != nil {
		fmt.Println("error:", err)
	}
	// Encode map key.
	if err := enc.Encode("a"); err != nil {
		fmt.Println("error:", err)
	}
	// Encode map value.
	if err := enc.Encode(1); err != nil {
		fmt.Println("error:", err)
	}
	// Encode map key.
	if err := enc.Encode("b"); err != nil {
		fmt.Println("error:", err)
	}
	// Start an indefinite length array as map value.
	if err := enc.StartIndefiniteArray(); err != nil {
		fmt.Println("error:", err)
	}
	// Encoded array element.
	if err := enc.Encode(2); err != nil {
		fmt.Println("error:", err)
	}
	// Encoded array element.
	if err := enc.Encode(3); err != nil {
		fmt.Println("error:", err)
	}
	// Close indefinite length array.
	if err := enc.EndIndefinite(); err != nil {
		fmt.Println("error:", err)
	}
	// Close indefinite length map.
	if err := enc.EndIndefinite(); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%0x\n", buf.Bytes())
	// Output:
	// bf61610161629f0203ffff
}
