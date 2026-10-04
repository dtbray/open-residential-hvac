// SPDX-License-Identifier: AGPL-3.0-only
package infiltration

import (
	"fmt"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
)

func Calculate(s building.InfiltrationSpec, volume units.Volume) (units.Airflow, string, error) {
	var a units.Airflow
	var equation string
	switch s.Mode {
	case "explicit":
		if s.Airflow == nil || s.ACH != nil || s.ACH50 != nil || s.ConversionFactor != nil {
			return 0, "", fmt.Errorf("explicit infiltration requires airflow only")
		}
		a = *s.Airflow
		equation = "airflow = user-specified airflow"
	case "ach_natural":
		if s.ACH == nil || s.Airflow != nil || s.ACH50 != nil || s.ConversionFactor != nil {
			return 0, "", fmt.Errorf("ach_natural requires ACH only")
		}
		if !units.Finite(*s.ACH) || *s.ACH < 0 {
			return 0, "", fmt.Errorf("natural ACH must be nonnegative and finite")
		}
		a = units.Airflow(*s.ACH * float64(volume) / 3600)
		equation = "airflow = ACH × volume / 3600"
	case "ach50":
		if s.ACH50 == nil || s.ConversionFactor == nil || s.ACH != nil || s.Airflow != nil {
			return 0, "", fmt.Errorf("ach50 requires ACH50 and an explicit conversion_factor only")
		}
		if !units.Finite(*s.ACH50) || *s.ACH50 < 0 || !units.Finite(*s.ConversionFactor) || *s.ConversionFactor <= 0 {
			return 0, "", fmt.Errorf("ACH50 must be nonnegative; conversion factor must be positive; both finite")
		}
		a = units.Airflow(*s.ACH50 / *s.ConversionFactor * float64(volume) / 3600)
		equation = "ACHnatural = ACH50 / conversion_factor; airflow = ACHnatural × volume / 3600"
	default:
		return 0, "", fmt.Errorf("unsupported infiltration mode %q", s.Mode)
	}
	if !units.Finite(float64(a)) || a < 0 {
		return 0, "", fmt.Errorf("infiltration airflow must be nonnegative and finite")
	}
	return a, equation, nil
}
