// SPDX-License-Identifier: AGPL-3.0-only
package infiltration

import (
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"math"
	"testing"
)

func TestModels(t *testing.T) {
	ach, ach50, factor := 0.5, 5.0, 10.0
	for _, s := range []building.InfiltrationSpec{{Mode: "ach_natural", ACH: &ach}, {Mode: "ach50", ACH50: &ach50, ConversionFactor: &factor}} {
		flow, _, err := Calculate(s, 360)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(float64(flow)-0.05) > 1e-12 {
			t.Fatalf("got %g, expected 0.05 m3/s", flow)
		}
	}
	if _, _, err := Calculate(building.InfiltrationSpec{Mode: "ach50", ACH50: &ach50}, 360); err == nil {
		t.Fatal("missing conversion accepted")
	}
}
