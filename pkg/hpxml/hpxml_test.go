// SPDX-License-Identifier: AGPL-3.0-only
package hpxml

import (
	"bytes"
	"encoding/json"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/models/designload"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/project"
	"math"
	"os"
	"testing"
)

func TestRepresentativeApartmentImport(t *testing.T) {
	for _, name := range []string{"apartment", "apartment-hot"} {
		t.Run(name, func(t *testing.T) {
			base := "../../testdata/openstudio/" + name + "/"
			xml, err := os.ReadFile(base + "resolved.xml")
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(base + "import-options.json")
			if err != nil {
				t.Fatal(err)
			}
			var options Options
			if err = json.Unmarshal(data, &options); err != nil {
				t.Fatal(err)
			}
			imported, err := ImportBasic(xml, options)
			if err != nil {
				t.Fatal(err)
			}
			room := imported.Building.Zones[0].Rooms[0]
			if len(room.Walls) != 2 || len(room.Windows) != 3 || len(room.Doors) != 1 || len(room.Floors) != 1 || len(room.Ceilings) != 1 {
				t.Fatal("representative envelope incomplete")
			}
			// Gross external wall minus all openings is the raw oracle's 542.4 ft2.
			net := room.Walls[0].Area.SquareFeet() - room.Doors[0].Area.SquareFeet()
			for _, window := range room.Windows {
				net -= window.Area.SquareFeet()
			}
			if math.Abs(net-542.4) > 1e-8 {
				t.Fatalf("net wall area: %g", net)
			}
			got, err := designload.New().Calculate(imported.Building, imported.Building.Design)
			if err != nil {
				t.Fatal(err)
			}
			expected, err := project.Open(base + "project.json")
			if err != nil {
				t.Fatal(err)
			}
			want, err := designload.New().Calculate(expected.Building, expected.Building.Design)
			if err != nil {
				t.Fatal(err)
			}
			if got.HeatingLoad != want.HeatingLoad || got.CoolingLoad != want.CoolingLoad {
				t.Fatal("adapter and committed native project diverged")
			}
			delete(options.Adjacent, "Wall2")
			if _, err = ImportBasic(xml, options); err == nil {
				t.Fatal("adjacent-unit temperatures silently inferred")
			}
		})
	}
}

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
