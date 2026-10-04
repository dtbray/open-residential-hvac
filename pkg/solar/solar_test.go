// SPDX-License-Identifier: AGPL-3.0-only
package solar

import (
	"math"
	"testing"
)

func TestGlazing(t *testing.T) {
	if got := Glazing(15, 400, 0.5, 0.8); math.Abs(float64(got)-2400) > 1e-12 {
		t.Fatalf("got %g, want 2400 W", got)
	}
	if Glazing(15, 400, 0.5, 0) != 0 {
		t.Fatal("fully shaded gain must be zero")
	}
}
