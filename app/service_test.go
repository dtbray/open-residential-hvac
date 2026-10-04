// SPDX-License-Identifier: AGPL-3.0-only
package app

import (
	"bytes"
	"encoding/json"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/loads"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/project"
	"os"
	"testing"
)

func TestDesktopServiceMatchesLibrary(t *testing.T) {
	b, e := os.ReadFile("../testdata/buildings/multi-room-ranch.json")
	if e != nil {
		t.Fatal(e)
	}
	p, e := project.Decode(b, "json")
	if e != nil {
		t.Fatal(e)
	}
	want, e := loads.Calculate(p.Building)
	if e != nil {
		t.Fatal(e)
	}
	got, e := (&Service{}).Calculate(string(b))
	if e != nil {
		t.Fatal(e)
	}
	a, _ := json.Marshal(want)
	c, _ := json.Marshal(got)
	if !bytes.Equal(a, c) {
		t.Fatal("desktop results differ from library")
	}
}
