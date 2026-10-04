// SPDX-License-Identifier: AGPL-3.0-only
package project

import (
	"bytes"
	"reflect"
	"testing"
)

func TestRanchRoundTrip(t *testing.T) {
	p, err := Open("../../testdata/buildings/multi-room-ranch.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"json", "yaml"} {
		data, err := Encode(*p, format)
		if err != nil {
			t.Fatal(err)
		}
		q, err := Decode(data, format)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(p, q) {
			t.Fatalf("%s roundtrip lost data", format)
		}
	}
}
func TestStrictDecoding(t *testing.T) {
	if _, err := Decode([]byte(`{"version":1,"version":1}`), "json"); err == nil {
		t.Fatal("duplicate JSON keys accepted")
	}
	p, _ := Open("../../testdata/buildings/simple-box.json")
	b, _ := Encode(*p, "json")
	for _, data := range [][]byte{bytes.Replace(b, []byte(`"version": 1`), []byte(`"version": 2`), 1), bytes.Replace(b, []byte(`"version": 1`), []byte(`"version": 1, "typo": true`), 1), append(b, []byte(` {}`)...)} {
		if _, err := Decode(data, "json"); err == nil {
			t.Fatal("bad JSON accepted")
		}
	}
	yaml, _ := Encode(*p, "yaml")
	if _, err := Decode(append(yaml, []byte("\n---\n{}\n")...), "yaml"); err == nil {
		t.Fatal("multiple YAML documents accepted")
	}
	if _, err := Decode(append(yaml, []byte("\ntypo: true\n")...), "yaml"); err == nil {
		t.Fatal("unknown YAML field accepted")
	}
}
