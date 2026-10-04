// SPDX-License-Identifier: AGPL-3.0-only
package validation

import (
	"encoding/json"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/models/designload"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/project"
	"os"
	"testing"
)

func TestReferenceComparison(t *testing.T) {
	data, e := os.ReadFile("../../testdata/designload/buildings/wall-conduction-only.json")
	if e != nil {
		t.Fatal(e)
	}
	p, e := project.Decode(data, "json")
	if e != nil {
		t.Fatal(e)
	}
	r, e := designload.New().Calculate(p.Building, p.Building.Design)
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile("../../testdata/designload/reference/wall-conduction-hand.json")
	if e != nil {
		t.Fatal(e)
	}
	var ref Reference
	if e = json.Unmarshal(b, &ref); e != nil {
		t.Fatal(e)
	}
	report, e := Compare(r, data, ref, 0, 1e-8)
	if e == nil && report.Methodology != r.Methodology {
		t.Fatal("comparison report lost calculation methodology")
	}
	if e != nil || !report.Pass || len(report.Differences) != 6 {
		t.Fatal("hand comparison failed", e)
	}
	ref.Metrics["heating_load_w"] = 300
	report, e = Compare(r, data, ref, 0.01, 0.5)
	if e != nil || report.Pass {
		t.Fatal("deviation concealed", e)
	}
	if _, e = Compare(r, append(data, ' '), ref, 0.01, 0.5); e == nil {
		t.Fatal("changed fixture accepted")
	}
	ref.Metrics["unknown-component"] = 1
	if _, e = Compare(r, data, ref, 0.01, 0.5); e == nil {
		t.Fatal("unknown component silently ignored")
	}
}

func TestOutdoorAirReference(t *testing.T) {
	data, e := os.ReadFile("../../testdata/designload/buildings/leaky-house.json")
	if e != nil {
		t.Fatal(e)
	}
	p, e := project.Decode(data, "json")
	if e != nil {
		t.Fatal(e)
	}
	r, e := designload.New().Calculate(p.Building, p.Building.Design)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("../../testdata/designload/reference/leaky-house-psychrolib.json")
	if e != nil {
		t.Fatal(e)
	}
	var ref Reference
	if e = json.Unmarshal(raw, &ref); e != nil {
		t.Fatal(e)
	}
	report, e := Compare(r, data, ref, 1e-10, 1e-8)
	if e != nil || !report.Pass || len(report.Differences) != 7 {
		t.Fatalf("independent outdoor-air comparison failed: %+v, %v", report, e)
	}
}
