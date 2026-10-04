// SPDX-License-Identifier: AGPL-3.0-only
package designload

import (
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/loads"
	"math"

	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/climate"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/envelope"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/infiltration"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/physics"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/provenance"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/psychrometrics"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/ventilation"
)

// Calculate implements a simultaneous, steady-state open-physics load model.
// Positive loads are summed; negative component credits are clamped explicitly.
func (m *Model) Calculate(b building.Building, conditions climate.DesignConditions) (*loads.Result, error) {
	b.Design = conditions
	if errors := m.Validate(b, conditions); len(errors) > 0 {
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
	result := &loads.Result{Model: m.ID(), Methodology: Methodology()}
	var heatingZones, coolingZones []loads.ResultNode
	for _, z := range b.Zones {
		var heatingRooms, coolingRooms []loads.ResultNode
		for _, r := range z.Rooms {
			prefix := "room/" + r.ID
			var heatEnvelope, coolEnvelope, solarNodes []loads.ResultNode
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
				base := []loads.InputValue{loads.Input(s.ID+".gross_area", float64(gross), "m2", s.Evidence["area_m2"]), loads.Input(s.ID+".opening_area", float64(gross-area), "m2", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: "sum of opening areas on " + s.ID}}), loads.Input(s.ID+".net_area", float64(area), "m2", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: "gross area minus openings"}}), loads.Input(a.ID+".u_factor", float64(u), "W/(m2 K)", a.Evidence)}
				if len(a.Layers) > 0 {
					base[3].Source = provenance.Source{Kind: provenance.Derived, Reference: "1 / sum(layer resistance): " + a.ID}
					for _, l := range a.Layers {
						base = append(base, loads.Input(a.ID+".layer."+l.Name, float64(l.Resistance), "m2 K/W", l.Evidence))
					}
				}
				for _, w := range r.Windows {
					if w.ParentSurface == s.ID {
						base = append(base, loads.Input(w.ID+".opening_area", float64(w.Area), "m2", w.Evidence["area_m2"]))
					}
				}
				for _, door := range r.Doors {
					if door.ParentSurface == s.ID {
						base = append(base, loads.Input(door.ID+".opening_area", float64(door.Area), "m2", door.Evidence["area_m2"]))
					}
				}
				hi := append(append([]loads.InputValue{}, base...), loads.Input("adjacent.heating_db", float64(hOut), "C", boundaryH), loads.Input("indoor.heating_db", float64(*d.Heating.IndoorDB), "C", d.Evidence["heating.indoor_db_c"]))
				ci := append(append([]loads.InputValue{}, base...), loads.Input("adjacent.cooling_db", float64(cOut), "C", boundaryC), loads.Input("indoor.cooling_db", float64(*d.Cooling.IndoorDB), "C", d.Evidence["cooling.indoor_db_c"]))
				h := loads.Leaf(prefix+"/heating/envelope/"+s.ID, s.Name, "steady-state conduction", "Q = max(0, U × A_net × (T_indoor - T_adjacent))", physics.Conduction(u, area, units.TemperatureDifference(math.Max(0, float64(*d.Heating.IndoorDB-hOut)))), hi...)
				c := loads.Leaf(prefix+"/cooling/sensible/envelope/"+s.ID, s.Name, "steady-state conduction", "Q = max(0, U × A_net × (T_adjacent - T_indoor))", physics.Conduction(u, area, units.TemperatureDifference(math.Max(0, float64(cOut-*d.Cooling.IndoorDB)))), ci...)
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
				gain := physics.SolarGain(w.Area, *w.IncidentSolar, *w.SHGC, *w.ShadingFactor)
				n := loads.Leaf(prefix+"/cooling/sensible/solar/"+w.ID, w.Name+" solar", "instantaneous glazing gain", "Q = A × I_plane × SHGC × shading_factor", gain, loads.Input(w.ID+".area", float64(w.Area), "m2", w.Evidence["area_m2"]), loads.Input(w.ID+".irradiance", float64(*w.IncidentSolar), "W/m2", w.Evidence["incident_solar_w_m2"]), loads.Input(w.ID+".shgc", *w.SHGC, "fraction", w.Evidence["shgc"]), loads.Input(w.ID+".shading_factor", *w.ShadingFactor, "fraction", w.Evidence["shading_factor"]))
				if gain > 0 {
					n.Warnings = []string{"Instantaneous solar heat gain is treated as immediate sensible load; thermal storage and time lag are unsupported."}
				}
				solarNodes = append(solarNodes, n)
			}
			for _, door := range r.Doors {
				conduction(inherit(door.Surface, door.ParentSurface), door.Area, door.Area)
			}
			roomInf := units.Airflow(float64(inf) * float64(r.Volume) / float64(volume))
			flowInputs := []loads.InputValue{loads.Input("building.infiltration_airflow", float64(inf), "m3/s", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: infEquation}}), loads.Input("building.volume", float64(volume), "m3", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: "sum(room volumes)"}}), loads.Input(r.ID+".volume", float64(r.Volume), "m3", r.Evidence["volume_m3"])}
			for _, other := range rooms {
				if other.ID != r.ID {
					flowInputs = append(flowInputs, loads.Input(other.ID+".volume", float64(other.Volume), "m3", other.Evidence["volume_m3"]))
				}
			}
			if b.Infiltration.ACH != nil {
				flowInputs = append(flowInputs, loads.Input("infiltration.ach", *b.Infiltration.ACH, "1/h", b.Infiltration.Evidence["ach"]))
			}
			if b.Infiltration.ACH50 != nil {
				flowInputs = append(flowInputs, loads.Input("infiltration.ach50", *b.Infiltration.ACH50, "1/h", b.Infiltration.Evidence["ach50"]), loads.Input("infiltration.conversion_factor", *b.Infiltration.ConversionFactor, "ratio", b.Infiltration.Evidence["conversion_factor"]))
			}
			if b.Infiltration.Airflow != nil {
				flowInputs = append(flowInputs, loads.Input("infiltration.explicit_airflow", float64(*b.Infiltration.Airflow), "m3/s", b.Infiltration.Evidence["airflow_m3_s"]))
			}
			vInputs := []loads.InputValue{loads.Input("building.ventilation_airflow", float64(*b.Ventilation.Airflow), "m3/s", b.Ventilation.Evidence["airflow_m3_s"])}
			vMethod := "explicit room ventilation airflow"
			if len(b.Ventilation.RoomRates) == 0 {
				vMethod = "building ventilation × room volume / building volume"
				vInputs = append(vInputs, loads.Input("building.volume", float64(volume), "m3", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: "sum(room volumes)"}}), loads.Input(r.ID+".volume", float64(r.Volume), "m3", r.Evidence["volume_m3"]))
				for _, other := range rooms {
					if other.ID != r.ID {
						vInputs = append(vInputs, loads.Input(other.ID+".volume", float64(other.Volume), "m3", other.Evidence["volume_m3"]))
					}
				}
			} else {
				vInputs = append(vInputs, loads.Input(r.ID+".ventilation_airflow", float64(vent[r.ID]), "m3/s", b.Ventilation.Evidence["room_rates_m3_s."+r.ID]))
			}
			hair, cair, lair := airNodes(prefix, "infiltration", roomInf, infEquation+"; room airflow = building airflow × room volume / building volume", flowInputs, b, wi, wo)
			hvent, cvent, lvent := airNodes(prefix, "ventilation", vent[r.ID], vMethod, vInputs, b, wi, wo)
			if b.Infiltration.Mode == "ach50" {
				for _, n := range []*loads.ResultNode{&hair, &cair, &lair} {
					n.Warnings = append(n.Warnings, "ACH50 divisor is supplied by the user; design wind/stack behaviour is not independently modeled.")
				}
			}
			internal := func(kind, name, key string, q units.HeatFlow) loads.ResultNode {
				return loads.Leaf(prefix+"/cooling/"+kind+"/internal/"+key, name, "explicit design internal gain", "Q = explicit heat gain", q, loads.Input(r.ID+"."+key, float64(q), "W", r.Evidence[key]))
			}
			occupant := func(kind string, q units.HeatFlow) loads.ResultNode {
				return loads.Leaf(prefix+"/cooling/"+kind+"/internal/occupants", "Occupants", "user-specified per-person design gain", "Q = occupants × per_person_gain", physics.InternalGain(*r.Occupants, q), loads.Input(r.ID+".occupants", float64(*r.Occupants), "people", r.Evidence["occupants"]), loads.Input(r.ID+".occupant_"+kind, float64(q), "W/person", r.Evidence["occupant_"+kind+"_w_per_person"]))
			}
			h := loads.Aggregate(prefix+"/heating", r.Name+" heating", loads.Aggregate(prefix+"/heating/envelope", "Envelope", heatEnvelope...), hair, hvent)
			cs := loads.Aggregate(prefix+"/cooling/sensible", r.Name+" cooling sensible", loads.Aggregate(prefix+"/cooling/sensible/envelope", "Envelope", coolEnvelope...), loads.Aggregate(prefix+"/cooling/sensible/solar", "Solar", solarNodes...), cair, cvent, loads.Aggregate(prefix+"/cooling/sensible/internal", "Internal gains", occupant("sensible", *r.OccupantSensible), internal("sensible", "Lighting", "lighting_w", *r.Lighting), internal("sensible", "Other sensible gains", "internal_sensible_w", *r.InternalSensible)))
			cl := loads.Aggregate(prefix+"/cooling/latent", r.Name+" cooling latent", lair, lvent, loads.Aggregate(prefix+"/cooling/latent/internal", "Internal gains", occupant("latent", *r.OccupantLatent), internal("latent", "Other latent gains", "internal_latent_w", *r.InternalLatent)))
			c := loads.Aggregate(prefix+"/cooling", r.Name+" cooling", cs, cl)
			heatingRooms = append(heatingRooms, h)
			coolingRooms = append(coolingRooms, c)
			result.Rooms = append(result.Rooms, loads.RoomResult{ID: r.ID, ZoneID: z.ID, Name: r.Name, HeatingLoad: h.Value, CoolingSensible: cs.Value, CoolingLatent: cl.Value})
			result.CoolingSensible += cs.Value
			result.CoolingLatent += cl.Value
		}
		heatingZones = append(heatingZones, loads.Aggregate("zone/"+z.ID+"/heating", z.Name, heatingRooms...))
		coolingZones = append(coolingZones, loads.Aggregate("zone/"+z.ID+"/cooling", z.Name, coolingRooms...))
	}
	result.Heating = loads.Aggregate("building/"+b.ID+"/heating", "Building heating", heatingZones...)
	result.Cooling = loads.Aggregate("building/"+b.ID+"/cooling", "Building cooling", coolingZones...)
	result.HeatingLoad = result.Heating.Value
	result.CoolingLoad = result.Cooling.Value
	if b.Infiltration.Airtightness != nil {
		warning := "Stored airtightness measurement is retained but not converted automatically; infiltration uses the explicitly selected airflow mode."
		result.Heating.Warnings = append(result.Heating.Warnings, warning)
		result.Cooling.Warnings = append(result.Cooling.Warnings, warning)
	}
	if !units.Finite(float64(result.HeatingLoad)) || !units.Finite(float64(result.CoolingLoad)) {
		return nil, loads.ValidationErrors{{Code: "numerical_overflow", Path: "result", Message: "load overflow; review input magnitudes"}}
	}
	return result, nil
}

func airNodes(prefix, kind string, flow units.Airflow, flowEquation string, flowInputs []loads.InputValue, b building.Building, wi, wo float64) (loads.ResultNode, loads.ResultNode, loads.ResultNode) {
	d := b.Design
	p := float64(*d.Pressure)
	rhoH := psychrometrics.DryAirDensity(float64(*d.Heating.IndoorDB), 0, p)
	rhoC := psychrometrics.DryAirDensity(float64(*d.Cooling.IndoorDB), wi, p)
	ref := provenance.Evidence{Source: provenance.Source{Kind: provenance.PublicReference, Name: "PsychroLib SI air properties", Reference: "https://github.com/psychrometrics/psychrolib", License: "MIT"}}
	derived := provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: flowEquation}}
	base := append([]loads.InputValue{}, flowInputs...)
	base = append(base, loads.Input("room.airflow", float64(flow), "m3/s", derived), loads.Input("pressure", p, "Pa", d.Evidence["pressure_pa"]), loads.Input("dry_air_gas_constant", psychrometrics.DryAirGasConstant, "J/(kg K)", ref), loads.Input("dry_air_specific_heat", psychrometrics.DryAirSpecificHeat, "J/(kg K)", ref))
	hi := append(append([]loads.InputValue{}, base...), loads.Input("indoor.heating_db", float64(*d.Heating.IndoorDB), "C", d.Evidence["heating.indoor_db_c"]), loads.Input("outdoor.heating_db", float64(*d.Heating.OutdoorDB), "C", d.Evidence["heating.outdoor_db_c"]))
	h := loads.Leaf(prefix+"/heating/"+kind, kind, "dry-air sensible heat transfer; "+flowEquation, "rho = p/(R_air × T_indoor_K); Q = max(0, rho × airflow × cp_air × ΔT)", physics.SensibleAirLoad(physics.AirExchange(flow, units.DryAirDensity(rhoH)), units.SpecificHeat(psychrometrics.DryAirSpecificHeat), units.TemperatureDifference(math.Max(0, float64(*d.Heating.IndoorDB-*d.Heating.OutdoorDB)))), hi...)
	h.Assumptions = append(h.Assumptions, provenance.Assumption{ID: "dry-air-heating-density", Description: "Heating sensible airflow uses dry-air density at the indoor heating temperature."})
	ci := append(append([]loads.InputValue{}, base...), loads.Input("indoor.cooling_db", float64(*d.Cooling.IndoorDB), "C", d.Evidence["cooling.indoor_db_c"]), loads.Input("outdoor.cooling_db", float64(*d.Cooling.OutdoorDB), "C", d.Evidence["cooling.outdoor_db_c"]), loads.Input("indoor.rh", *d.Cooling.IndoorRH, "fraction", d.Evidence["cooling.indoor_rh_fraction"]), loads.Input("indoor.humidity_ratio", wi, "kg/kg dry air", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: "humidity ratio from indoor dry bulb, RH, pressure"}}), loads.Input("outdoor.humidity_ratio", wo, "kg/kg dry air", provenance.Evidence{Source: provenance.Source{Kind: provenance.Derived, Reference: "humidity ratio from outdoor dry bulb, humidity condition, pressure"}}), loads.Input("vapour_specific_heat", psychrometrics.VapourSpecificHeat, "J/(kg K)", ref), loads.Input("latent_heat", psychrometrics.LatentHeat, "J/kg", ref))
	if d.Cooling.OutdoorRH != nil {
		ci = append(ci, loads.Input("outdoor.rh", *d.Cooling.OutdoorRH, "fraction", d.Evidence["cooling.outdoor_rh_fraction"]))
	} else {
		ci = append(ci, loads.Input("outdoor.wet_bulb", float64(*d.Cooling.OutdoorWB), "C", d.Evidence["cooling.outdoor_wb_c"]))
	}
	s := loads.Leaf(prefix+"/cooling/sensible/"+kind, kind, "moist-air sensible heat transfer; "+flowEquation, "rho_dry = p/[R_air × T_K × (1 + 1.607858 × W_indoor)]; Q = max(0, rho_dry × airflow × (1006 + 1860 × W_indoor) × ΔT)", physics.SensibleAirLoad(physics.AirExchange(flow, units.DryAirDensity(rhoC)), psychrometrics.MoistAirSpecificHeat(units.HumidityRatio(wi)), units.TemperatureDifference(math.Max(0, float64(*d.Cooling.OutdoorDB-*d.Cooling.IndoorDB)))), ci...)
	l := loads.Leaf(prefix+"/cooling/latent/"+kind, kind, "humidity-ratio latent heat transfer; "+flowEquation, "Q = max(0, rho_dry × airflow × 2501000 × (W_outdoor - W_indoor))", physics.LatentAirLoad(physics.AirExchange(flow, units.DryAirDensity(rhoC)), units.SpecificEnergy(psychrometrics.LatentHeat), units.HumidityRatio(math.Max(0, wo-wi))), ci...)
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
