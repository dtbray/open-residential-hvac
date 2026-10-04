// SPDX-License-Identifier: AGPL-3.0-only
package loads

import (
	"math"

	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/envelope"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/infiltration"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/provenance"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/psychrometrics"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/solar"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/ventilation"
)

// Calculate implements a simultaneous, steady-state open-physics load model.
// Positive loads are summed; negative component credits are clamped explicitly.
func Calculate(b building.Building) (*Result, error) {
	if errors := Validate(b); len(errors) > 0 {
		return nil, errors
	}
	d := b.Design
	p := float64(*d.Pressure)
	wi, _ := psychrometrics.HumidityRatioRH(float64(*d.Cooling.IndoorDB), *d.Cooling.IndoorRH, p)
	var wo float64
	if d.Cooling.OutdoorRH != nil {
		wo, _ = psychrometrics.HumidityRatioRH(float64(*d.Cooling.OutdoorDB), *d.Cooling.OutdoorRH, p)
	} else {
		wo, _ = psychrometrics.HumidityRatioWB(float64(*d.Cooling.OutdoorDB), float64(*d.Cooling.OutdoorWB), p)
	}
	assemblies := map[string]building.Assembly{}
	for _, a := range b.Assemblies {
		assemblies[a.ID] = a
	}
	var rooms []building.Room
	var volume units.Volume
	for _, z := range b.Zones {
		for _, r := range z.Rooms {
			rooms = append(rooms, r)
			volume += r.Volume
		}
	}
	inf, infEquation, _ := infiltration.Calculate(b.Infiltration, volume)
	vent, _ := ventilation.Allocate(b.Ventilation, rooms)
	result := &Result{Model: "openphysics-steady-state-v1"}
	var heatingZones, coolingZones []ResultNode
	for _, z := range b.Zones {
		var heatingRooms, coolingRooms []ResultNode
		for _, r := range z.Rooms {
			prefix := "room/" + r.ID
			var heatEnvelope, coolEnvelope, solarNodes []ResultNode
			all := append(append(append([]building.Surface{}, r.Walls...), r.Floors...), r.Ceilings...)
			parents := map[string]building.Surface{}
			areas := map[string]units.Area{}
			for _, s := range all {
				parents[s.ID] = s
				areas[s.ID] = s.Area
			}
			for _, w := range r.Windows {
				areas[w.ParentSurface] -= w.Area
			}
			for _, door := range r.Doors {
				areas[door.ParentSurface] -= door.Area
			}
			conduction := func(s building.Surface, area units.Area, gross units.Area) {
				a := assemblies[s.Assembly]
				u, _ := envelope.Transmittance(a)
				hOut, cOut := *d.Heating.OutdoorDB, *d.Cooling.OutdoorDB
				boundaryH, boundaryC := d.Evidence["heating.outdoor_db_c"], d.Evidence["cooling.outdoor_db_c"]
				if s.Adjacent == building.Conditioned {
					hOut = *d.Heating.IndoorDB
					cOut = *d.Cooling.IndoorDB
					boundaryH = d.Evidence["heating.indoor_db_c"]
					boundaryC = d.Evidence["cooling.indoor_db_c"]
				} else if s.Adjacent != building.Outdoors {
					hOut = *s.HeatingAdjacentDB
					cOut = *s.CoolingAdjacentDB
					boundaryH = s.Evidence["heating_adjacent_db_c"]
					boundaryC = s.Evidence["cooling_adjacent_db_c"]
				}
				base := []InputValue{input(s.ID+".gross_area", float64(gross), "m2", s.Evidence["area_m2"]), input(s.ID+".opening_area", float64(gross-area), "m2", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: "sum of opening areas on " + s.ID}}), input(s.ID+".net_area", float64(area), "m2", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: "gross area minus openings"}}), input(a.ID+".u_factor", float64(u), "W/(m2 K)", a.Evidence)}
				if len(a.Layers) > 0 {
					base[3].Source = provenance.Source{Kind: provenance.Derived, Reference: "1 / sum(layer resistance): " + a.ID}
					for _, l := range a.Layers {
						base = append(base, input(a.ID+".layer."+l.Name, float64(l.Resistance), "m2 K/W", l.Evidence))
					}
				}
				for _, w := range r.Windows {
					if w.ParentSurface == s.ID {
						base = append(base, input(w.ID+".opening_area", float64(w.Area), "m2", w.Evidence["area_m2"]))
					}
				}
				for _, door := range r.Doors {
					if door.ParentSurface == s.ID {
						base = append(base, input(door.ID+".opening_area", float64(door.Area), "m2", door.Evidence["area_m2"]))
					}
				}
				hi := append(append([]InputValue{}, base...), input("adjacent.heating_db", float64(hOut), "C", boundaryH), input("indoor.heating_db", float64(*d.Heating.IndoorDB), "C", d.Evidence["heating.indoor_db_c"]))
				ci := append(append([]InputValue{}, base...), input("adjacent.cooling_db", float64(cOut), "C", boundaryC), input("indoor.cooling_db", float64(*d.Cooling.IndoorDB), "C", d.Evidence["cooling.indoor_db_c"]))
				h := leaf(prefix+"/heating/envelope/"+s.ID, s.Name, "steady-state conduction", "Q = max(0, U × A_net × (T_indoor - T_adjacent))", envelope.Conduction(u, area, math.Max(0, float64(*d.Heating.IndoorDB-hOut))), hi...)
				c := leaf(prefix+"/cooling/sensible/envelope/"+s.ID, s.Name, "steady-state conduction", "Q = max(0, U × A_net × (T_adjacent - T_indoor))", envelope.Conduction(u, area, math.Max(0, float64(cOut-*d.Cooling.IndoorDB))), ci...)
				heatEnvelope = append(heatEnvelope, h)
				coolEnvelope = append(coolEnvelope, c)
			}
			for _, s := range all {
				conduction(s, areas[s.ID], s.Area)
			}
			inherit := func(s building.Surface, parent string) building.Surface {
				if s.Adjacent == "" {
					ps := parents[parent]
					s.Adjacent = ps.Adjacent
					s.HeatingAdjacentDB = ps.HeatingAdjacentDB
					s.CoolingAdjacentDB = ps.CoolingAdjacentDB
					if s.Evidence == nil {
						s.Evidence = map[string]provenance.Evidence{}
					} else {
						copyMap := map[string]provenance.Evidence{}
						for k, v := range s.Evidence {
							copyMap[k] = v
						}
						s.Evidence = copyMap
					}
					s.Evidence["heating_adjacent_db_c"] = ps.Evidence["heating_adjacent_db_c"]
					s.Evidence["cooling_adjacent_db_c"] = ps.Evidence["cooling_adjacent_db_c"]
				}
				return s
			}
			for _, w := range r.Windows {
				conduction(inherit(w.Surface, w.ParentSurface), w.Area, w.Area)
				gain := solar.Glazing(w.Area, *w.IncidentSolar, *w.SHGC, *w.ShadingFactor)
				n := leaf(prefix+"/cooling/sensible/solar/"+w.ID, w.Name+" solar", "instantaneous glazing gain", "Q = A × I_plane × SHGC × shading_factor", gain, input(w.ID+".area", float64(w.Area), "m2", w.Evidence["area_m2"]), input(w.ID+".irradiance", float64(*w.IncidentSolar), "W/m2", w.Evidence["incident_solar_w_m2"]), input(w.ID+".shgc", *w.SHGC, "fraction", w.Evidence["shgc"]), input(w.ID+".shading_factor", *w.ShadingFactor, "fraction", w.Evidence["shading_factor"]))
				if gain > 0 {
					n.Warnings = []string{"Instantaneous solar heat gain is treated as immediate sensible load; thermal storage and time lag are unsupported."}
				}
				solarNodes = append(solarNodes, n)
			}
			for _, door := range r.Doors {
				conduction(inherit(door.Surface, door.ParentSurface), door.Area, door.Area)
			}
			roomInf := units.Airflow(float64(inf) * float64(r.Volume) / float64(volume))
			flowInputs := []InputValue{input("building.infiltration_airflow", float64(inf), "m3/s", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: infEquation}}), input("building.volume", float64(volume), "m3", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: "sum(room volumes)"}}), input(r.ID+".volume", float64(r.Volume), "m3", r.Evidence["volume_m3"])}
			for _, other := range rooms {
				if other.ID != r.ID {
					flowInputs = append(flowInputs, input(other.ID+".volume", float64(other.Volume), "m3", other.Evidence["volume_m3"]))
				}
			}
			if b.Infiltration.ACH != nil {
				flowInputs = append(flowInputs, input("infiltration.ach", *b.Infiltration.ACH, "1/h", b.Infiltration.Evidence["ach"]))
			}
			if b.Infiltration.ACH50 != nil {
				flowInputs = append(flowInputs, input("infiltration.ach50", *b.Infiltration.ACH50, "1/h", b.Infiltration.Evidence["ach50"]), input("infiltration.conversion_factor", *b.Infiltration.ConversionFactor, "ratio", b.Infiltration.Evidence["conversion_factor"]))
			}
			if b.Infiltration.Airflow != nil {
				flowInputs = append(flowInputs, input("infiltration.explicit_airflow", float64(*b.Infiltration.Airflow), "m3/s", b.Infiltration.Evidence["airflow_m3_s"]))
			}
			vInputs := []InputValue{input("building.ventilation_airflow", float64(*b.Ventilation.Airflow), "m3/s", b.Ventilation.Evidence["airflow_m3_s"])}
			vMethod := "explicit room ventilation airflow"
			if len(b.Ventilation.RoomRates) == 0 {
				vMethod = "building ventilation × room volume / building volume"
				vInputs = append(vInputs, input("building.volume", float64(volume), "m3", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: "sum(room volumes)"}}), input(r.ID+".volume", float64(r.Volume), "m3", r.Evidence["volume_m3"]))
				for _, other := range rooms {
					if other.ID != r.ID {
						vInputs = append(vInputs, input(other.ID+".volume", float64(other.Volume), "m3", other.Evidence["volume_m3"]))
					}
				}
			} else {
				vInputs = append(vInputs, input(r.ID+".ventilation_airflow", float64(vent[r.ID]), "m3/s", b.Ventilation.Evidence["room_rates_m3_s."+r.ID]))
			}
			hair, cair, lair := airNodes(prefix, "infiltration", roomInf, infEquation+"; room airflow = building airflow × room volume / building volume", flowInputs, b, wi, wo)
			hvent, cvent, lvent := airNodes(prefix, "ventilation", vent[r.ID], vMethod, vInputs, b, wi, wo)
			if b.Infiltration.Mode == "ach50" {
				for _, n := range []*ResultNode{&hair, &cair, &lair} {
					n.Warnings = append(n.Warnings, "ACH50 divisor is supplied by the user; design wind/stack behaviour is not independently modeled.")
				}
			}
			internal := func(kind, name, key string, q units.HeatFlow) ResultNode {
				return leaf(prefix+"/cooling/"+kind+"/internal/"+key, name, "explicit design internal gain", "Q = explicit heat gain", q, input(r.ID+"."+key, float64(q), "W", r.Evidence[key]))
			}
			occupant := func(kind string, q units.HeatFlow) ResultNode {
				return leaf(prefix+"/cooling/"+kind+"/internal/occupants", "Occupants", "user-specified per-person design gain", "Q = occupants × per_person_gain", units.HeatFlow(float64(r.Occupants)*float64(q)), input(r.ID+".occupants", float64(r.Occupants), "people", r.Evidence["occupants"]), input(r.ID+".occupant_"+kind, float64(q), "W/person", r.Evidence["occupant_"+kind+"_w_per_person"]))
			}
			h := aggregate(prefix+"/heating", r.Name+" heating", aggregate(prefix+"/heating/envelope", "Envelope", heatEnvelope...), hair, hvent)
			cs := aggregate(prefix+"/cooling/sensible", r.Name+" cooling sensible", aggregate(prefix+"/cooling/sensible/envelope", "Envelope", coolEnvelope...), aggregate(prefix+"/cooling/sensible/solar", "Solar", solarNodes...), cair, cvent, aggregate(prefix+"/cooling/sensible/internal", "Internal gains", occupant("sensible", *r.OccupantSensible), internal("sensible", "Lighting", "lighting_w", *r.Lighting), internal("sensible", "Other sensible gains", "internal_sensible_w", *r.InternalSensible)))
			cl := aggregate(prefix+"/cooling/latent", r.Name+" cooling latent", lair, lvent, aggregate(prefix+"/cooling/latent/internal", "Internal gains", occupant("latent", *r.OccupantLatent), internal("latent", "Other latent gains", "internal_latent_w", *r.InternalLatent)))
			c := aggregate(prefix+"/cooling", r.Name+" cooling", cs, cl)
			heatingRooms = append(heatingRooms, h)
			coolingRooms = append(coolingRooms, c)
			result.Rooms = append(result.Rooms, RoomResult{ID: r.ID, ZoneID: z.ID, Name: r.Name, HeatingLoad: h.Value, CoolingSensible: cs.Value, CoolingLatent: cl.Value})
			result.CoolingSensible += cs.Value
			result.CoolingLatent += cl.Value
		}
		heatingZones = append(heatingZones, aggregate("zone/"+z.ID+"/heating", z.Name, heatingRooms...))
		coolingZones = append(coolingZones, aggregate("zone/"+z.ID+"/cooling", z.Name, coolingRooms...))
	}
	result.Heating = aggregate("building/"+b.ID+"/heating", "Building heating", heatingZones...)
	result.Cooling = aggregate("building/"+b.ID+"/cooling", "Building cooling", coolingZones...)
	result.HeatingLoad = result.Heating.Value
	result.CoolingLoad = result.Cooling.Value
	if !units.Finite(float64(result.HeatingLoad)) || !units.Finite(float64(result.CoolingLoad)) {
		return nil, ValidationErrors{{Code: "numerical_overflow", Path: "result", Message: "load overflow; review input magnitudes"}}
	}
	return result, nil
}

func airNodes(prefix, kind string, flow units.Airflow, flowEquation string, flowInputs []InputValue, b building.Building, wi, wo float64) (ResultNode, ResultNode, ResultNode) {
	d := b.Design
	p := float64(*d.Pressure)
	rhoH := psychrometrics.DryAirDensity(float64(*d.Heating.IndoorDB), 0, p)
	rhoC := psychrometrics.DryAirDensity(float64(*d.Cooling.IndoorDB), wi, p)
	ref := provenance.Evidence{Source: provenance.Source{Kind: provenance.PublicReference, Name: "PsychroLib SI air properties", Reference: "https://github.com/psychrometrics/psychrolib", License: "MIT"}}
	derived := provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: flowEquation}}
	base := append([]InputValue{}, flowInputs...)
	base = append(base, input("room.airflow", float64(flow), "m3/s", derived), input("pressure", p, "Pa", d.Evidence["pressure_pa"]), input("dry_air_gas_constant", psychrometrics.DryAirGasConstant, "J/(kg K)", ref), input("dry_air_specific_heat", psychrometrics.DryAirSpecificHeat, "J/(kg K)", ref))
	hi := append(append([]InputValue{}, base...), input("indoor.heating_db", float64(*d.Heating.IndoorDB), "C", d.Evidence["heating.indoor_db_c"]), input("outdoor.heating_db", float64(*d.Heating.OutdoorDB), "C", d.Evidence["heating.outdoor_db_c"]))
	h := leaf(prefix+"/heating/"+kind, kind, "dry-air sensible heat transfer; "+flowEquation, "rho = p/(R_air × T_indoor_K); Q = max(0, rho × airflow × cp_air × ΔT)", units.HeatFlow(rhoH*float64(flow)*psychrometrics.DryAirSpecificHeat*math.Max(0, float64(*d.Heating.IndoorDB-*d.Heating.OutdoorDB))), hi...)
	h.Assumptions = append(h.Assumptions, provenance.Assumption{ID: "dry-air-heating-density", Description: "Heating sensible airflow uses dry-air density at the indoor heating temperature."})
	ci := append(append([]InputValue{}, base...), input("indoor.cooling_db", float64(*d.Cooling.IndoorDB), "C", d.Evidence["cooling.indoor_db_c"]), input("outdoor.cooling_db", float64(*d.Cooling.OutdoorDB), "C", d.Evidence["cooling.outdoor_db_c"]), input("indoor.rh", *d.Cooling.IndoorRH, "fraction", d.Evidence["cooling.indoor_rh_fraction"]), input("indoor.humidity_ratio", wi, "kg/kg dry air", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: "humidity ratio from indoor dry bulb, RH, pressure"}}), input("outdoor.humidity_ratio", wo, "kg/kg dry air", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: "humidity ratio from outdoor dry bulb, humidity condition, pressure"}}), input("vapour_specific_heat", psychrometrics.VapourSpecificHeat, "J/(kg K)", ref), input("latent_heat", psychrometrics.LatentHeat, "J/kg", ref))
	if d.Cooling.OutdoorRH != nil {
		ci = append(ci, input("outdoor.rh", *d.Cooling.OutdoorRH, "fraction", d.Evidence["cooling.outdoor_rh_fraction"]))
	} else {
		ci = append(ci, input("outdoor.wet_bulb", float64(*d.Cooling.OutdoorWB), "C", d.Evidence["cooling.outdoor_wb_c"]))
	}
	s := leaf(prefix+"/cooling/sensible/"+kind, kind, "moist-air sensible heat transfer; "+flowEquation, "rho_dry = p/[R_air × T_K × (1 + 1.607858 × W_indoor)]; Q = max(0, rho_dry × airflow × (1006 + 1860 × W_indoor) × ΔT)", units.HeatFlow(rhoC*float64(flow)*(psychrometrics.DryAirSpecificHeat+psychrometrics.VapourSpecificHeat*wi)*math.Max(0, float64(*d.Cooling.OutdoorDB-*d.Cooling.IndoorDB))), ci...)
	l := leaf(prefix+"/cooling/latent/"+kind, kind, "humidity-ratio latent heat transfer; "+flowEquation, "Q = max(0, rho_dry × airflow × 2501000 × (W_outdoor - W_indoor))", units.HeatFlow(rhoC*float64(flow)*psychrometrics.LatentHeat*math.Max(0, wo-wi)), ci...)
	if kind == "ventilation" && len(b.Ventilation.RoomRates) == 0 {
		a := provenance.Assumption{ID: "ventilation-volume-allocation", Description: "Building ventilation allocated to rooms in proportion to volume; explicit room rates can replace this allocation."}
		h.Assumptions = append(h.Assumptions, a)
		s.Assumptions = append(s.Assumptions, a)
		l.Assumptions = append(l.Assumptions, a)
	}
	a := provenance.Assumption{ID: "airflow-reference-volume", Description: "Airflows interpreted as volume flow at indoor design conditions; no heat/moisture recovery modeled."}
	h.Assumptions = append(h.Assumptions, a)
	s.Assumptions = append(s.Assumptions, a)
	l.Assumptions = append(l.Assumptions, a)
	return h, s, l
}
