// SPDX-License-Identifier: AGPL-3.0-only
package ventilation

import (
	"fmt"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
	"math"
)

// Allocate requires explicit room rates or distributes by room volume.
func Allocate(s building.VentilationSpec, rooms []building.Room) (map[string]units.Airflow, error) {
	if s.Airflow == nil || !units.Finite(float64(*s.Airflow)) || *s.Airflow < 0 {
		return nil, fmt.Errorf("ventilation requires explicit nonnegative finite airflow")
	}
	out := map[string]units.Airflow{}
	var volume units.Volume
	valid := map[string]bool{}
	for _, r := range rooms {
		volume += r.Volume
		valid[r.ID] = true
	}
	if volume <= 0 {
		return nil, fmt.Errorf("building volume must be positive")
	}
	if len(s.RoomRates) > 0 {
		var total units.Airflow
		for id, a := range s.RoomRates {
			if !valid[id] || !units.Finite(float64(a)) || a < 0 {
				return nil, fmt.Errorf("invalid ventilation room rate %s", id)
			}
			out[id] = a
			total += a
		}
		if math.Abs(float64(total-*s.Airflow)) > 1e-9 {
			return nil, fmt.Errorf("ventilation room rates must sum to building airflow (tolerance 1e-9 m3/s)")
		}
		for _, r := range rooms {
			if _, ok := out[r.ID]; !ok {
				return nil, fmt.Errorf("ventilation room rate missing for %s", r.ID)
			}
		}
		return out, nil
	}
	for _, r := range rooms {
		out[r.ID] = units.Airflow(float64(*s.Airflow) * float64(r.Volume) / float64(volume))
	}
	return out, nil
}
