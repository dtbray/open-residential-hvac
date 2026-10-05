// SPDX-License-Identifier: AGPL-3.0-only
package psychrometrics

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

func close(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > math.Max(1e-12, math.Abs(want)*1e-10) {
		t.Fatalf("got %.12g, want %.12g", got, want)
	}
}
func TestPsychroLibReference(t *testing.T) {
	data, err := os.ReadFile("../../testdata/physics/psychrolib-2.5.0.json")
	if err != nil {
		t.Fatal(err)
	}
	var ref struct {
		RH []struct {
			DB  float64 `json:"db_c"`
			RH  float64 `json:"rh"`
			P   float64 `json:"p_pa"`
			W   float64 `json:"humidity_ratio"`
			Sat float64 `json:"saturation_pa"`
		} `json:"rh_cases"`
		WB []struct {
			DB float64 `json:"db_c"`
			WB float64 `json:"wb_c"`
			P  float64 `json:"p_pa"`
			W  float64 `json:"humidity_ratio"`
		} `json:"wb_cases"`
	}
	if err := json.Unmarshal(data, &ref); err != nil {
		t.Fatal(err)
	}
	if len(ref.RH) != 5 || len(ref.WB) != 4 {
		t.Fatal("incomplete reference")
	}
	for _, c := range ref.RH {
		w, e := HumidityRatioRH(c.DB, c.RH, c.P)
		if e != nil {
			t.Fatal(e)
		}
		close(t, w, c.W)
		sat, e := SaturationPressure(c.DB)
		if e != nil {
			t.Fatal(e)
		}
		close(t, sat, c.Sat)
	}
	for _, c := range ref.WB {
		w, e := HumidityRatioWB(c.DB, c.WB, c.P)
		if e != nil {
			t.Fatal(e)
		}
		close(t, w, c.W)
	}
}
func TestInvalidConditions(t *testing.T) {
	for _, c := range [][3]float64{{25, -0.1, 101325}, {25, 1.1, 101325}, {25, 0.5, 0}, {25, 0.5, math.NaN()}, {300, 0.5, 101325}, {25, math.NaN(), 101325}} {
		if _, err := HumidityRatioRH(c[0], c[1], c[2]); err == nil {
			t.Fatal("invalid RH accepted")
		}
	}
	if _, err := HumidityRatioWB(20, 21, 101325); err == nil {
		t.Fatal("wet bulb above dry bulb accepted")
	}
}
