// SPDX-License-Identifier: AGPL-3.0-only
package physics

import (
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
	"math"
	"testing"
)

func TestThermalPrimitivesPreserveSigns(t *testing.T) {
	cases := []struct {
		name     string
		actual   units.HeatFlow
		expected float64
	}{
		{"conduction", Conduction(0.4, 20, 35), 280},
		{"signed conduction", Conduction(0.4, 20, -11), -88},
		{"solar", SolarGain(15, 400, 0.5, 0.8), 2400},
		{"sensible air", SensibleAirLoad(AirExchange(0.05, 1.2), 1006, 10), 603.6},
		{"signed sensible air", SensibleAirLoad(AirExchange(0.05, 1.2), 1006, -10), -603.6},
		{"latent air", LatentAirLoad(0.06, 2501000, 0.004), 600.24},
		{"signed latent air", LatentAirLoad(0.06, 2501000, -0.004), -600.24},
		{"internal", InternalGain(3, 70), 210},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if math.Abs(float64(c.actual)-c.expected) > 1e-9 {
				t.Fatalf("got %.12g W; want %.12g W", c.actual, c.expected)
			}
		})
	}
}
