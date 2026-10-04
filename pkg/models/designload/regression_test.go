// SPDX-License-Identifier: AGPL-3.0-only
package designload

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/loads"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/project"
	"os"
	"testing"
)

// A captured same-method baseline protects numerical and stable-ID behaviour
// during architectural migration. It is not an independent validation oracle.
func TestPreRefactorRegression(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/designload/reference/pre-refactor-regression.json")
	if err != nil {
		t.Fatal(err)
	}
	var baseline struct {
		Kind     string `json:"kind"`
		Fixtures map[string]struct {
			Hash     string             `json:"fixture_sha256"`
			Heating  float64            `json:"heating_load_w"`
			Sensible float64            `json:"cooling_sensible_w"`
			Latent   float64            `json:"cooling_latent_w"`
			Cooling  float64            `json:"cooling_load_w"`
			Nodes    map[string]float64 `json:"nodes_w"`
		} `json:"fixtures"`
	}
	if err = json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	if baseline.Kind != "regression" || len(baseline.Fixtures) != 11 {
		t.Fatal("incomplete or mislabeled baseline")
	}
	for name, expected := range baseline.Fixtures {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile("../../../testdata/designload/buildings/" + name)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%x", sha256.Sum256(raw)) != expected.Hash {
				t.Fatal("fixture no longer matches recorded regression input")
			}
			p, err := project.Decode(raw, "json")
			if err != nil {
				t.Fatal(err)
			}
			r, err := New().Calculate(p.Building, p.Building.Design)
			if err != nil {
				t.Fatal(err)
			}
			equal(t, float64(r.HeatingLoad), expected.Heating)
			equal(t, float64(r.CoolingSensible), expected.Sensible)
			equal(t, float64(r.CoolingLatent), expected.Latent)
			equal(t, float64(r.CoolingLoad), expected.Cooling)
			actual := map[string]float64{}
			var walk func(loads.ResultNode)
			walk = func(n loads.ResultNode) {
				actual[n.ID] = float64(n.Value)
				for _, c := range n.Children {
					walk(c)
				}
			}
			walk(r.Heating)
			walk(r.Cooling)
			if len(actual) != len(expected.Nodes) {
				t.Fatal("result node count changed")
			}
			for id, value := range expected.Nodes {
				v, ok := actual[id]
				if !ok {
					t.Fatal("stable result ID missing", id)
				}
				equal(t, v, value)
			}
		})
	}
}
