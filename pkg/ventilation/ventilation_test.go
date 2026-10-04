// SPDX-License-Identifier: AGPL-3.0-only
package ventilation

import (
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
	"math"
	"testing"
)

func TestAllocation(t *testing.T) {
	total := units.Airflow(0.03)
	rooms := []building.Room{{ID: "a", Volume: 100}, {ID: "b", Volume: 50}}
	s := building.VentilationSpec{Airflow: &total}
	r, e := Allocate(s, rooms)
	if e != nil {
		t.Fatal(e)
	}
	if math.Abs(float64(r["a"])-0.02) > 1e-12 || math.Abs(float64(r["b"])-0.01) > 1e-12 {
		t.Fatal("incorrect volume allocation")
	}
	s.RoomRates = map[string]units.Airflow{"a": 0.01, "b": 0.02}
	r, e = Allocate(s, rooms)
	if e != nil || r["a"] != 0.01 {
		t.Fatal("explicit allocation failed", e)
	}
	s.RoomRates["b"] = 0.03
	if _, e := Allocate(s, rooms); e == nil {
		t.Fatal("nonconserving room airflow accepted")
	}
}
