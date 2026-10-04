// SPDX-License-Identifier: AGPL-3.0-only
package loads

import (
	"math"
	"sort"
	"strings"

	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/envelope"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/infiltration"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/provenance"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/psychrometrics"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/ventilation"
)

type DomainError struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

func (e DomainError) Error() string { return e.Path + ": " + e.Message }

type ValidationErrors []DomainError

func (e ValidationErrors) Error() string {
	var lines []string
	for _, v := range e {
		lines = append(lines, v.Error())
	}
	return strings.Join(lines, "\n")
}

func Validate(b building.Building) ValidationErrors {
	var errors ValidationErrors
	add := func(path, code, msg string) { errors = append(errors, DomainError{code, path, msg}) }
	ids := map[string]bool{}
	id := func(path, v string) {
		if v == "" || strings.Contains(v, "/") {
			add(path, "invalid_id", "ID must be nonempty and contain no slash")
		} else if ids[v] {
			add(path, "duplicate_id", "duplicate ID "+v)
		}
		ids[v] = true
	}
	evidence := func(path string, e map[string]provenance.Evidence) {
		for k, v := range e {
			if v.Source.Kind != "" && !v.Source.Valid() {
				add(path+"."+k, "invalid_source", "unknown source kind")
			}
			for _, a := range v.Assumptions {
				if a.ID == "" || a.Description == "" {
					add(path+"."+k, "invalid_assumption", "assumptions require ID and description")
				}
			}
		}
	}
	positive := func(path string, v float64) {
		if !units.Finite(v) || v <= 0 {
			add(path, "invalid_quantity", "must be positive and finite")
		}
	}
	nonnegative := func(path string, v float64) {
		if !units.Finite(v) || v < 0 {
			add(path, "invalid_quantity", "must be nonnegative and finite")
		}
	}
	heat := func(path string, v *units.HeatFlow) {
		if v == nil {
			add(path, "missing_input", "explicit heat gain required (zero is permitted)")
		} else {
			nonnegative(path, float64(*v))
		}
	}
	fraction := func(path string, v *float64) {
		if v == nil || !units.Finite(*v) || *v < 0 || *v > 1 {
			add(path, "invalid_fraction", "requires a finite fraction in [0,1]")
		}
	}
	temp := func(path string, v *units.Temperature) {
		if v == nil {
			add(path, "missing_input", "explicit temperature required")
		} else if _, err := psychrometrics.SaturationPressure(float64(*v)); err != nil {
			add(path, "invalid_temperature", err.Error())
		}
	}
	id("building.id", b.ID)
	d := b.Design
	temp("design.heating.outdoor_db_c", d.Heating.OutdoorDB)
	temp("design.heating.indoor_db_c", d.Heating.IndoorDB)
	temp("design.cooling.outdoor_db_c", d.Cooling.OutdoorDB)
	temp("design.cooling.indoor_db_c", d.Cooling.IndoorDB)
	fraction("design.cooling.indoor_rh_fraction", d.Cooling.IndoorRH)
	if d.Pressure == nil {
		add("design.pressure_pa", "missing_input", "explicit atmospheric pressure required")
	} else {
		positive("design.pressure_pa", float64(*d.Pressure))
	}
	if (d.Cooling.OutdoorWB == nil) == (d.Cooling.OutdoorRH == nil) {
		add("design.cooling", "missing_humidity", "specify exactly one outdoor wet bulb or RH")
	}
	if d.Cooling.OutdoorRH != nil {
		fraction("design.cooling.outdoor_rh_fraction", d.Cooling.OutdoorRH)
	}
	if d.Cooling.OutdoorWB != nil {
		temp("design.cooling.outdoor_wb_c", d.Cooling.OutdoorWB)
	}
	evidence("design.evidence", d.Evidence)
	if d.Cooling.OutdoorDB != nil && d.Cooling.IndoorDB != nil && d.Pressure != nil && d.Cooling.IndoorRH != nil {
		if _, err := psychrometrics.HumidityRatioRH(float64(*d.Cooling.IndoorDB), *d.Cooling.IndoorRH, float64(*d.Pressure)); err != nil {
			add("design.cooling.indoor", "invalid_psychrometrics", err.Error())
		}
		if d.Cooling.OutdoorRH != nil {
			if _, err := psychrometrics.HumidityRatioRH(float64(*d.Cooling.OutdoorDB), *d.Cooling.OutdoorRH, float64(*d.Pressure)); err != nil {
				add("design.cooling.outdoor", "invalid_psychrometrics", err.Error())
			}
		}
		if d.Cooling.OutdoorWB != nil {
			if _, err := psychrometrics.HumidityRatioWB(float64(*d.Cooling.OutdoorDB), float64(*d.Cooling.OutdoorWB), float64(*d.Pressure)); err != nil {
				add("design.cooling.outdoor", "invalid_psychrometrics", err.Error())
			}
		}
	}
	assemblies := map[string]bool{}
	for _, a := range b.Assemblies {
		id("assembly.id", a.ID)
		assemblies[a.ID] = true
		if _, err := envelope.Transmittance(a); err != nil {
			add("assembly."+a.ID, "invalid_assembly", err.Error())
		}
		evidence("assembly."+a.ID, map[string]provenance.Evidence{"thermal": a.Evidence})
		for _, l := range a.Layers {
			evidence("assembly."+a.ID+".layer."+l.Name, map[string]provenance.Evidence{"resistance": l.Evidence})
		}
	}
	var rooms []building.Room
	var volume units.Volume
	for _, z := range b.Zones {
		id("zone.id", z.ID)
		if len(z.Rooms) == 0 {
			add("zone."+z.ID, "missing_rooms", "zone requires at least one room")
		}
		for _, r := range z.Rooms {
			p := "room." + r.ID
			id(p, r.ID)
			rooms = append(rooms, r)
			volume += r.Volume
			positive(p+".floor_area_m2", float64(r.FloorArea))
			positive(p+".volume_m3", float64(r.Volume))
			if r.Occupants < 0 {
				add(p+".occupants", "invalid_quantity", "occupants must be nonnegative")
			}
			heat(p+".occupant_sensible", r.OccupantSensible)
			heat(p+".occupant_latent", r.OccupantLatent)
			heat(p+".lighting", r.Lighting)
			heat(p+".internal_sensible", r.InternalSensible)
			heat(p+".internal_latent", r.InternalLatent)
			evidence(p+".evidence", r.Evidence)
			parents := map[string]building.Surface{}
			openingArea := map[string]units.Area{}
			all := append(append(append([]building.Surface{}, r.Walls...), r.Floors...), r.Ceilings...)
			for _, s := range all {
				parents[s.ID] = s
			}
			check := func(s building.Surface) {
				sp := p + ".surface." + s.ID
				id(sp, s.ID)
				positive(sp+".area_m2", float64(s.Area))
				if !assemblies[s.Assembly] {
					add(sp+".assembly", "unknown_assembly", "unknown assembly "+s.Assembly)
				}
				if !units.Finite(s.TiltDegrees) || s.TiltDegrees < 0 || s.TiltDegrees > 180 {
					add(sp+".tilt_degrees", "invalid_orientation", "tilt must be within 0..180 degrees")
				}
				switch s.Adjacent {
				case building.Outdoors, building.Conditioned:
					if s.HeatingAdjacentDB != nil || s.CoolingAdjacentDB != nil {
						add(sp, "conflicting_boundary", "outdoor and conditioned surfaces cannot override boundary temperatures")
					}
				case building.Ground:
					add(sp, "unsupported_ground", "ground/slab heat transfer is unsupported; no silent steady-state approximation")
				case building.Attic, building.Garage, building.Crawlspace, building.Basement, building.AdjacentUnit:
					temp(sp+".heating_adjacent_db_c", s.HeatingAdjacentDB)
					temp(sp+".cooling_adjacent_db_c", s.CoolingAdjacentDB)
				default:
					add(sp, "unsupported_adjacency", "unknown adjacency "+string(s.Adjacent))
				}
				evidence(sp+".evidence", s.Evidence)
			}
			for _, s := range all {
				check(s)
			}
			opening := func(s building.Surface, parent string) {
				ps, ok := parents[parent]
				if !ok {
					add(p+".opening."+s.ID, "unknown_parent", "unknown parent surface "+parent)
				} else {
					if s.Adjacent == "" {
						s.Adjacent = ps.Adjacent
						s.HeatingAdjacentDB = ps.HeatingAdjacentDB
						s.CoolingAdjacentDB = ps.CoolingAdjacentDB
					}
					if s.Adjacent != ps.Adjacent {
						add(p+".opening."+s.ID, "conflicting_boundary", "opening adjacency must match parent")
					}
					openingArea[parent] += s.Area
				}
				check(s)
			}
			for _, w := range r.Windows {
				opening(w.Surface, w.ParentSurface)
				fraction(p+".window."+w.ID+".shgc", w.SHGC)
				fraction(p+".window."+w.ID+".shading_factor", w.ShadingFactor)
				if w.IncidentSolar == nil {
					add(p+".window."+w.ID, "missing_solar", "explicit incident plane irradiance required, including explicit zero")
				} else {
					nonnegative(p+".window."+w.ID+".solar", float64(*w.IncidentSolar))
				}
				ps, ok := parents[w.ParentSurface]
				if ok && ps.Adjacent != building.Outdoors && w.IncidentSolar != nil && *w.IncidentSolar != 0 {
					add(p+".window."+w.ID, "unsupported_solar", "solar gains supported only for outdoor glazing")
				}
			}
			for _, door := range r.Doors {
				opening(door.Surface, door.ParentSurface)
			}
			for parent, a := range openingArea {
				if a > parents[parent].Area {
					add(p+".surface."+parent, "invalid_openings", "opening area exceeds gross surface area")
				}
			}
		}
	}
	if len(rooms) == 0 {
		add("building.zones", "missing_rooms", "building requires at least one room")
	}
	if _, _, err := infiltration.Calculate(b.Infiltration, volume); err != nil {
		add("infiltration", "invalid_infiltration", err.Error())
	}
	evidence("infiltration.evidence", b.Infiltration.Evidence)
	if _, err := ventilation.Allocate(b.Ventilation, rooms); err != nil {
		add("ventilation", "invalid_ventilation", err.Error())
	}
	evidence("ventilation.evidence", b.Ventilation.Evidence)
	// Overflow in accumulated volume is also an invalid engineering input.
	if math.IsInf(float64(volume), 0) {
		add("building.volume", "invalid_quantity", "aggregate volume overflow")
	}
	sort.SliceStable(errors, func(i, j int) bool {
		return errors[i].Path+errors[i].Code+errors[i].Message < errors[j].Path+errors[j].Code+errors[j].Message
	})
	return errors
}
