// SPDX-License-Identifier: AGPL-3.0-only
package solar

import "git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"

// Glazing requires incident irradiance on the glazing plane at one simultaneous
// design snapshot. No weather lookup, time lag, or orientation inference is used.
func Glazing(area units.Area, irradiance units.Irradiance, shgc, shading float64) units.HeatFlow {
	return units.HeatFlow(float64(area) * float64(irradiance) * shgc * shading)
}
