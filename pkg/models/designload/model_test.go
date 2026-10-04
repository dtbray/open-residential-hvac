// SPDX-License-Identifier: AGPL-3.0-only
package designload

import (
	"bytes"
	"encoding/json"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/loads"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/provenance"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
	"math"
	"strings"
	"testing"
)

func TestExplicitConditionsOverrideStoredDesign(t *testing.T) {
	b := fixture(t, "simple-box")
	conditions := b.Design
	out := units.Celsius(-30)
	conditions.Heating.OutdoorDB = &out
	conditions.Evidence = map[string]provenance.Evidence{"heating.outdoor_db_c": {Source: provenance.Source{Kind: provenance.Dataset, Reference: "independent physical design conditions"}}}
	// An invalid stored cooling condition proves the model validates the explicit
	// conditions instead of accidentally reading the project's stored defaults.
	b.Design.Cooling.IndoorRH = nil
	before, _ := json.Marshal(b)
	r, err := New().Calculate(b, conditions)
	if err != nil {
		t.Fatal(err)
	}
	equal(t, float64(r.HeatingLoad), 400)
	if r.Methodology != Methodology() || r.Model != ID {
		t.Fatal("methodology not identified", r.Methodology)
	}
	n := loads.Find(r.Heating, "room/living/heating/envelope/living-west")
	if n == nil {
		t.Fatal("missing wall result")
	}
	found := false
	for _, input := range n.Inputs {
		if input.Name == "adjacent.heating_db" && input.Source.Reference == "independent physical design conditions" {
			found = true
		}
	}
	if !found {
		t.Fatal("design-condition provenance lost")
	}
	after, _ := json.Marshal(b)
	if !bytes.Equal(before, after) {
		t.Fatal("explicit conditions mutated original building")
	}
}

func TestMetricAndImperialInputEquivalence(t *testing.T) {
	b := fixture(t, "simple-box")
	b.Infiltration.Mode = "explicit"
	flow := units.LitresPerSecond(25)
	b.Infiltration.Airflow = &flow
	want, err := calculate(b)
	if err != nil {
		t.Fatal(err)
	}
	for zi := range b.Zones {
		for ri := range b.Zones[zi].Rooms {
			r := &b.Zones[zi].Rooms[ri]
			r.FloorArea = units.SquareFeet(r.FloorArea.SquareFeet())
			r.Volume = units.CubicFeet(r.Volume.CubicFeet())
			for si := range r.Walls {
				r.Walls[si].Area = units.SquareFeet(r.Walls[si].Area.SquareFeet())
			}
		}
	}
	for _, value := range []*units.Temperature{b.Design.Heating.OutdoorDB, b.Design.Heating.IndoorDB, b.Design.Cooling.OutdoorDB, b.Design.Cooling.IndoorDB} {
		*value = units.Fahrenheit(value.Fahrenheit())
	}
	flow = units.CFM(flow.CFM())
	b.Infiltration.Airflow = &flow
	got, err := calculate(b)
	if err != nil {
		t.Fatal(err)
	}
	equal(t, float64(got.HeatingLoad), float64(want.HeatingLoad))
	equal(t, float64(got.CoolingLoad), float64(want.CoolingLoad))
}

func TestGenericBoundaries(t *testing.T) {
	for _, boundary := range []building.Adjacency{building.UnconditionedSpace, building.AdjacentBuilding} {
		b := fixture(t, "simple-box")
		s := &b.Zones[0].Rooms[0].Walls[0]
		s.Adjacent = boundary
		h, c := units.Celsius(0), units.Celsius(30)
		s.HeatingAdjacentDB = &h
		s.CoolingAdjacentDB = &c
		r, err := calculate(b)
		if err != nil {
			t.Fatal(err)
		}
		equal(t, float64(r.HeatingLoad), 160)
		equal(t, float64(r.CoolingSensible), 48)
		s.CoolingAdjacentDB = nil
		if _, err = calculate(b); err == nil {
			t.Fatal("missing generic boundary temperature accepted")
		}
	}
}

func TestPreservedPropertiesAreNotSilentlyApplied(t *testing.T) {
	b := fixture(t, "simple-box")
	ach := units.AirChangeRate(5)
	b.Infiltration.Airtightness = &building.AirtightnessMeasurement{AirChangesPerHour: &ach, TestPressure: 50, Evidence: provenance.Evidence{Source: provenance.Source{Kind: provenance.Measurement, Reference: "blower-door test"}}}
	r, err := calculate(b)
	if err != nil {
		t.Fatal(err)
	}
	equal(t, float64(r.HeatingLoad), 280)
	if !strings.Contains(strings.Join(r.Heating.Warnings, " "), "Stored airtightness") {
		t.Fatal("unused airtightness conversion was not disclosed")
	}
	efficiency := 0.8
	b.Ventilation.HeatRecovery = &building.HeatRecovery{SensibleEfficiency: &efficiency}
	if _, err = calculate(b); err == nil || !strings.Contains(err.Error(), "heat-recovery") {
		t.Fatal("unsupported heat recovery silently ignored", err)
	}
}

func TestInvalidPreservedProperties(t *testing.T) {
	for _, value := range []units.Azimuth{-1, 360, units.Azimuth(math.NaN())} {
		b := fixture(t, "simple-box")
		b.Zones[0].Rooms[0].Walls[0].Azimuth = &value
		if _, err := calculate(b); err == nil {
			t.Fatal("invalid azimuth accepted", value)
		}
	}
	b := fixture(t, "simple-box")
	b.Infiltration.Airtightness = &building.AirtightnessMeasurement{TestPressure: 50}
	if _, err := calculate(b); err == nil {
		t.Fatal("missing airtightness rate silently assumed zero")
	}
}
