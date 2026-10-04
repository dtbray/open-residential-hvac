// SPDX-License-Identifier: AGPL-3.0-only
package envelope

import (
	"fmt"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
)

func Transmittance(a building.Assembly) (units.Transmittance, error) {
	if a.UFactor != nil && len(a.Layers) > 0 {
		return 0, fmt.Errorf("assembly %s specifies both U-factor and layers", a.ID)
	}
	if a.UFactor != nil {
		if !units.Finite(float64(*a.UFactor)) || *a.UFactor <= 0 {
			return 0, fmt.Errorf("assembly %s U-factor must be positive and finite", a.ID)
		}
		return *a.UFactor, nil
	}
	var r units.Resistance
	for _, l := range a.Layers {
		if !units.Finite(float64(l.Resistance)) || l.Resistance <= 0 {
			return 0, fmt.Errorf("assembly %s layer %s resistance must be positive and finite", a.ID, l.Name)
		}
		r += l.Resistance
	}
	if r <= 0 || !units.Finite(float64(r)) {
		return 0, fmt.Errorf("assembly %s has neither U-factor nor sufficient layers", a.ID)
	}
	return units.Transmittance(1 / float64(r)), nil
}

// Conduction is signed; aggregation policy belongs to the load model.
func Conduction(u units.Transmittance, a units.Area, deltaC float64) units.HeatFlow {
	return units.HeatFlow(float64(u) * float64(a) * deltaC)
}
