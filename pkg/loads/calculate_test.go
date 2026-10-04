// SPDX-License-Identifier: AGPL-3.0-only
package loads

import (
	"bytes"
	"encoding/json"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/project"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/provenance"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
	"math"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T, name string) building.Building {
	t.Helper()
	p, e := project.Open("../../testdata/buildings/" + name + ".json")
	if e != nil {
		t.Fatal(e)
	}
	return p.Building
}
func equal(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-8 {
		t.Fatalf("got %.12g W, want %.12g W (absolute tolerance 1e-8 W)", got, want)
	}
}
func TestHandConduction(t *testing.T) {
	r, e := Calculate(fixture(t, "wall-conduction-only"))
	if e != nil {
		t.Fatal(e)
	}
	equal(t, float64(r.HeatingLoad), 0.4*20*35)
	equal(t, float64(r.CoolingSensible), 0.4*20*11)
	equal(t, float64(r.CoolingLatent), 0)
}
func TestOpeningsAndSolar(t *testing.T) {
	r, e := Calculate(fixture(t, "high-window-area"))
	if e != nil {
		t.Fatal(e)
	}
	equal(t, float64(r.HeatingLoad), (0.4*5+2*15)*35)
	equal(t, float64(r.CoolingSensible), (0.4*5+2*15)*11+15*400*0.5*0.8)
	n := Find(r.Cooling, "room/living/cooling/sensible/envelope/living-west")
	if n == nil {
		t.Fatal("wall node missing")
	}
	equal(t, float64(n.Value), 0.4*5*11)
}
func TestAllFixturesAndAggregation(t *testing.T) {
	files, _ := filepath.Glob("../../testdata/buildings/*.json")
	if len(files) < 10 {
		t.Fatal("fixture matrix missing")
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			p, e := project.Open(f)
			if e != nil {
				t.Fatal(e)
			}
			r, e := Calculate(p.Building)
			if e != nil {
				t.Fatal(e)
			}
			var h, s, l units.HeatFlow
			for _, v := range r.Rooms {
				h += v.HeatingLoad
				s += v.CoolingSensible
				l += v.CoolingLatent
			}
			equal(t, float64(r.HeatingLoad), float64(h))
			equal(t, float64(r.CoolingSensible), float64(s))
			equal(t, float64(r.CoolingLatent), float64(l))
			equal(t, float64(r.CoolingLoad), float64(s+l))
			ids := map[string]bool{}
			var check func(ResultNode)
			check = func(n ResultNode) {
				if ids[n.ID] {
					t.Fatal("duplicate result ID", n.ID)
				}
				ids[n.ID] = true
				if len(n.Children) > 0 {
					var sum units.HeatFlow
					for _, c := range n.Children {
						sum += c.Value
						check(c)
					}
					equal(t, float64(n.Value), float64(sum))
				} else {
					if n.Equation == "" || n.Method == "" {
						t.Fatal("unexplained result", n.ID)
					}
					for _, i := range n.Inputs {
						if !i.Source.Valid() {
							t.Fatal("unlabeled input", i.Name)
						}
					}
				}
			}
			check(r.Heating)
			check(r.Cooling)
			a, _ := json.Marshal(r)
			r2, e := Calculate(p.Building)
			if e != nil {
				t.Fatal(e)
			}
			b, _ := json.Marshal(r2)
			if !bytes.Equal(a, b) {
				t.Fatal("nondeterministic result")
			}
		})
	}
}
func TestRejectsInvalidInputs(t *testing.T) {
	for _, mutate := range []func(*building.Building){func(b *building.Building) { b.Design.Cooling.OutdoorDB = nil }, func(b *building.Building) { b.Zones[0].Rooms[0].Walls[0].Adjacent = building.Ground }, func(b *building.Building) { b.Zones[0].Rooms[0].Windows[0].ParentSurface = "missing" }, func(b *building.Building) { b.Zones[0].Rooms[0].Windows[0].Area = 100 }, func(b *building.Building) { b.Assemblies[0].UFactor = nil }, func(b *building.Building) { b.Design.Cooling.IndoorRH = nil }, func(b *building.Building) { b.Zones[0].Rooms[0].Lighting = nil }, func(b *building.Building) { b.Zones[0].Rooms[0].Volume = units.Volume(math.NaN()) }, func(b *building.Building) { b.Zones[0].Rooms[0].Walls[0].ID = b.ID }} {
		b := fixture(t, "high-window-area")
		mutate(&b)
		if _, e := Calculate(b); e == nil {
			t.Fatal("invalid engineering input accepted")
		}
	}
}
func TestAssumptionsSurvive(t *testing.T) {
	b := fixture(t, "simple-box")
	b.Assemblies[0].Evidence.Assumptions = []provenance.Assumption{{ID: "insulation", Description: "Assumed assembly resistance"}}
	r, e := Calculate(b)
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, a := range r.Cooling.Assumptions {
		if a.ID == "insulation" {
			found = true
		}
	}
	if !found {
		t.Fatal("assembly assumption lost in aggregation")
	}
}
func TestNoInputMutation(t *testing.T) {
	b := fixture(t, "multi-room-ranch")
	b.Zones[0].Rooms[0].Windows[0].Adjacent = ""
	before, _ := json.Marshal(b)
	if _, e := Calculate(b); e != nil {
		t.Fatal(e)
	}
	after, _ := json.Marshal(b)
	if !bytes.Equal(before, after) {
		t.Fatal("calculation mutated input")
	}
}

func BenchmarkRanch(b *testing.B) {
	p, e := project.Open("../../testdata/buildings/multi-room-ranch.json")
	if e != nil {
		b.Fatal(e)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := Calculate(p.Building); e != nil {
			b.Fatal(e)
		}
	}
}
