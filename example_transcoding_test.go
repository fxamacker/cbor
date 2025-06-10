// Copyright (c) Faye Amacker. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root for license information.

package cbor_test

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fxamacker/cbor/v2"
)

type TranscoderFunc func(io.Writer, io.Reader) error

func (f TranscoderFunc) Transcode(w io.Writer, r io.Reader) error {
	return f(w, r)
}

func ExampleTranscoder_fromJSON() {
	enc, _ := cbor.EncOptions{
		JSONMarshalerTranscoder: TranscoderFunc(func(w io.Writer, r io.Reader) error {
			d := json.NewDecoder(r)

			for {
				token, err := d.Token()
				if err == io.EOF {
					return nil
				}
				if err != nil {
					return err
				}
				switch token {
				case json.Delim('['):
					if _, err := w.Write([]byte{0x9f}); err != nil {
						return err
					}
				case json.Delim('{'):
					if _, err := w.Write([]byte{0xbf}); err != nil {
						return err
					}
				case json.Delim(']'), json.Delim('}'):
					if _, err := w.Write([]byte{0xff}); err != nil {
						return err
					}
				default:
					b, err := cbor.Marshal(token)
					if err != nil {
						return err
					}
					if _, err := w.Write(b); err != nil {
						return err
					}
				}
			}
		}),
	}.EncMode()

	got, _ := enc.Marshal(json.RawMessage(`{"a": [true, "z", {"y": 3.14}], "b": {"c": null}}`))
	diag, _ := cbor.Diagnose(got)
	fmt.Println(diag)
	// Output: {_ "a": [_ true, "z", {_ "y": 3.14}], "b": {_ "c": null}}
}
