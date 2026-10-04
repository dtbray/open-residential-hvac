// SPDX-License-Identifier: AGPL-3.0-only
package envelope

import (
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
	"math"
	"testing"
)

func TestConductionAndLayers(t *testing.T) {
	a := building.Assembly{ID: "test", Layers: []building.Layer{{Resistance: 0.5}, {Resistance: 2}}}
	u, err := Transmittance(a)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(float64(Conduction(u, 20, 35))-280) > 1e-10 {
		t.Fatal("series layer or conduction mismatch")
	}
	v := units.Transmittance(0.4)
	a.UFactor = &v
	if _, err := Transmittance(a); err == nil {
		t.Fatal("ambiguous thermal spec accepted")
	}
}
func TestInvalidAssembly(t *testing.T) {
	for _, a := range []building.Assembly{{}, {Layers: []building.Layer{{Resistance: 0}}}, {Layers: []building.Layer{{Resistance: -1}}}} {
		if _, err := Transmittance(a); err == nil {
			t.Fatal("invalid resistance accepted")
		}
	}
}
