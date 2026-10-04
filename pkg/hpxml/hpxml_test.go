// SPDX-License-Identifier: AGPL-3.0-only
package hpxml

import (
	"bytes"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/models/designload"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/project"
	"math"
	"os"
	"testing"
)

func TestBasicAdapter(t *testing.T) {
	data, e := os.ReadFile("../../testdata/hpxml/basic-wall-window.xml")
	if e != nil {
		t.Fatal(e)
	}
	p, e := project.Open("../../testdata/designload/buildings/simple-box.json")
	if e != nil {
		t.Fatal(e)
	}
	options := Options{Template: p.Building, Solar: map[string]SolarInput{"west-glass": {Irradiance: 0, ShadingFactor: 1}}}
	imported, e := ImportBasic(data, options)
	if e != nil {
		t.Fatal(e)
	}
	r, e := designload.New().Calculate(imported.Building, imported.Building.Design)
	if e != nil {
		t.Fatal(e)
	}
	want := (150.0/14 + 30*0.3) * (35 * 1.8)
	if math.Abs(r.HeatingLoad.Btuh()-want) > 1e-7 {
		t.Fatalf("got %g Btu/h, want %g", r.HeatingLoad.Btuh(), want)
	}
	if len(imported.Warnings) == 0 {
		t.Fatal("adapter limitations missing")
	}
	if _, e := ImportBasic(data, Options{Template: p.Building}); e == nil {
		t.Fatal("missing solar input accepted")
	}
	bad := bytes.Replace(data, []byte("</Enclosure>"), []byte("<Slabs><Slab/></Slabs></Enclosure>"), 1)
	if _, e := ImportBasic(bad, options); e == nil {
		t.Fatal("unsupported envelope silently discarded")
	}
}
