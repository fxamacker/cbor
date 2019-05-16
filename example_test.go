// Copyright (c) 2019 Faye Amacker. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

package cbor_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

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
