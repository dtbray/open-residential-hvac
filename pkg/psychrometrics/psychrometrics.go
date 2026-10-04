// SPDX-License-Identifier: AGPL-3.0-only
// Saturation, wet-bulb, and air-property equations adapted from PsychroLib (MIT).
// See THIRD_PARTY_NOTICES.md for the upstream copyright and permission notice.
package psychrometrics

import (
	"fmt"
	"math"
)

const DryAirGasConstant = 287.042 // J/(kg K), PsychroLib SI
const DryAirSpecificHeat = 1006.0 // J/(kg K)
const VapourSpecificHeat = 1860.0 // J/(kg K)
const LatentHeat = 2501000.0      // J/kg, zero-C enthalpy convention

func SaturationPressure(t float64) (float64, error) {
	if math.IsNaN(t) || math.IsInf(t, 0) || t < -100 || t > 200 {
		return 0, fmt.Errorf("dry bulb must be finite and within -100..200 C")
	}
	k := t + 273.15
	var ln float64
	if t <= 0.01 {
		ln = -5.6745359e3/k + 6.3925247 - 9.677843e-3*k + 6.2215701e-7*k*k + 2.0747825e-9*k*k*k - 9.484024e-13*k*k*k*k + 4.1635019*math.Log(k)
	} else {
		ln = -5.8002206e3/k + 1.3914993 - 4.8640239e-2*k + 4.1764768e-5*k*k - 1.4452093e-8*k*k*k + 6.5459673*math.Log(k)
	}
	return math.Exp(ln), nil
}

func HumidityRatioRH(t, rh, p float64) (float64, error) {
	if !finite(rh) || rh < 0 || rh > 1 {
		return 0, fmt.Errorf("relative humidity must be finite and in [0,1]")
	}
	ps, err := SaturationPressure(t)
	if err != nil {
		return 0, err
	}
	return ratio(rh*ps, p)
}

func ratio(pw, p float64) (float64, error) {
	if !finite(p) || p <= 0 || pw >= p {
		return 0, fmt.Errorf("pressure must be finite, positive, and greater than vapour pressure")
	}
	return 0.621945 * pw / (p - pw), nil
}

func HumidityRatioWB(db, wb, p float64) (float64, error) {
	if !finite(wb) || wb > db {
		return 0, fmt.Errorf("wet bulb must be finite and no greater than dry bulb")
	}
	if _, err := SaturationPressure(db); err != nil {
		return 0, err
	}
	ps, err := SaturationPressure(wb)
	if err != nil {
		return 0, err
	}
	ws, err := ratio(ps, p)
	if err != nil {
		return 0, err
	}
	var w float64
	if wb >= 0 {
		w = ((2501-2.326*wb)*ws - 1.006*(db-wb)) / (2501 + 1.86*db - 4.186*wb)
	} else {
		w = ((2830-0.24*wb)*ws - 1.006*(db-wb)) / (2830 + 1.86*db - 2.1*wb)
	}
	if !finite(w) || w < 0 {
		return 0, fmt.Errorf("wet-bulb conditions imply invalid negative humidity ratio")
	}
	max, err := HumidityRatioRH(db, 1, p)
	if err != nil {
		return 0, err
	}
	if w > max+1e-10 {
		return 0, fmt.Errorf("wet-bulb conditions imply supersaturation")
	}
	return w, nil
}

// DryAirDensity returns kg dry air per m³ moist air.
func DryAirDensity(t, w, p float64) float64 {
	return p / (DryAirGasConstant * (t + 273.15) * (1 + 1.607858*w))
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
