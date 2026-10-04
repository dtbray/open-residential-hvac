// SPDX-License-Identifier: AGPL-3.0-only
package units

import (
	"math"
	"testing"
)

func TestConversions(t *testing.T) {
	checks := []struct {
		name         string
		actual, want float64
	}{
		{"freezing", float64(Fahrenheit(32)), 0},
		{"boiling", Celsius(100).Fahrenheit(), 212},
		{"area", float64(SquareFeet(1)), 0.09290304},
		{"volume", float64(CubicFeet(1)), 0.028316846592},
		{"heat", float64(Btuh(1)), 0.2930710701722222},
		{"flow", CFM(1).LitresPerSecond(), 0.4719474432},
		{"R reciprocal U", float64(RValueIP(10)) * float64(UFactorIP(0.1)), 1},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if math.Abs(c.actual-c.want) > 1e-10 {
				t.Fatalf("got %.12g, want %.12g", c.actual, c.want)
			}
		})
	}
}
