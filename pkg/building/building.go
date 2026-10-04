// SPDX-License-Identifier: AGPL-3.0-only
// Package building models physical residential buildings independently of standards.
package building

import (
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/climate"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/provenance"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
)

type Building struct {
	ID           string                   `json:"id" yaml:"id"`
	Name         string                   `json:"name" yaml:"name"`
	Location     Location                 `json:"location" yaml:"location"`
	Zones        []Zone                   `json:"zones" yaml:"zones"`
	Assemblies   []Assembly               `json:"assemblies" yaml:"assemblies"`
	Infiltration InfiltrationSpec         `json:"infiltration" yaml:"infiltration"`
	Ventilation  VentilationSpec          `json:"ventilation" yaml:"ventilation"`
	Design       climate.DesignConditions `json:"design" yaml:"design"`
}

type Location struct {
	Description string   `json:"description" yaml:"description"`
	Latitude    *float64 `json:"latitude,omitempty" yaml:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty" yaml:"longitude,omitempty"`
}

type Zone struct {
	ID    string `json:"id" yaml:"id"`
	Name  string `json:"name" yaml:"name"`
	Rooms []Room `json:"rooms" yaml:"rooms"`
}

type Room struct {
	ID               string                         `json:"id" yaml:"id"`
	Name             string                         `json:"name" yaml:"name"`
	FloorArea        units.Area                     `json:"floor_area_m2" yaml:"floor_area_m2"`
	Volume           units.Volume                   `json:"volume_m3" yaml:"volume_m3"`
	Occupants        int                            `json:"occupants" yaml:"occupants"`
	OccupantSensible *units.HeatFlow                `json:"occupant_sensible_w_per_person" yaml:"occupant_sensible_w_per_person"`
	OccupantLatent   *units.HeatFlow                `json:"occupant_latent_w_per_person" yaml:"occupant_latent_w_per_person"`
	Lighting         *units.HeatFlow                `json:"lighting_w" yaml:"lighting_w"`
	InternalSensible *units.HeatFlow                `json:"internal_sensible_w" yaml:"internal_sensible_w"`
	InternalLatent   *units.HeatFlow                `json:"internal_latent_w" yaml:"internal_latent_w"`
	Walls            []Surface                      `json:"walls" yaml:"walls"`
	Windows          []Window                       `json:"windows" yaml:"windows"`
	Doors            []Door                         `json:"doors" yaml:"doors"`
	Floors           []Surface                      `json:"floors" yaml:"floors"`
	Ceilings         []Surface                      `json:"ceilings" yaml:"ceilings"`
	Evidence         map[string]provenance.Evidence `json:"evidence" yaml:"evidence"`
}

type Adjacency string

const (
	Outdoors     Adjacency = "outdoors"
	Ground       Adjacency = "ground"
	Conditioned  Adjacency = "conditioned"
	Attic        Adjacency = "unconditioned_attic"
	Garage       Adjacency = "garage"
	Crawlspace   Adjacency = "crawlspace"
	Basement     Adjacency = "basement"
	AdjacentUnit Adjacency = "adjacent_unit"
)

// Area is gross; window/door areas referenced to this surface are subtracted.
type Surface struct {
	ID                string                         `json:"id" yaml:"id"`
	Name              string                         `json:"name" yaml:"name"`
	Area              units.Area                     `json:"area_m2" yaml:"area_m2"`
	Orientation       string                         `json:"orientation" yaml:"orientation"`
	TiltDegrees       float64                        `json:"tilt_degrees" yaml:"tilt_degrees"`
	Adjacent          Adjacency                      `json:"adjacent" yaml:"adjacent"`
	Assembly          string                         `json:"assembly" yaml:"assembly"`
	HeatingAdjacentDB *units.Temperature             `json:"heating_adjacent_db_c,omitempty" yaml:"heating_adjacent_db_c,omitempty"`
	CoolingAdjacentDB *units.Temperature             `json:"cooling_adjacent_db_c,omitempty" yaml:"cooling_adjacent_db_c,omitempty"`
	Evidence          map[string]provenance.Evidence `json:"evidence" yaml:"evidence"`
}

type Window struct {
	Surface       `json:",inline" yaml:",inline"`
	ParentSurface string            `json:"parent_surface" yaml:"parent_surface"`
	SHGC          *float64          `json:"shgc" yaml:"shgc"`
	IncidentSolar *units.Irradiance `json:"incident_solar_w_m2" yaml:"incident_solar_w_m2"`
	ShadingFactor *float64          `json:"shading_factor" yaml:"shading_factor"`
}

type Door struct {
	Surface       `json:",inline" yaml:",inline"`
	ParentSurface string `json:"parent_surface" yaml:"parent_surface"`
}

type Assembly struct {
	ID       string               `json:"id" yaml:"id"`
	Name     string               `json:"name" yaml:"name"`
	UFactor  *units.Transmittance `json:"u_factor_w_m2_k,omitempty" yaml:"u_factor_w_m2_k,omitempty"`
	Layers   []Layer              `json:"layers,omitempty" yaml:"layers,omitempty"`
	Evidence provenance.Evidence  `json:"evidence" yaml:"evidence"`
}

// Layer R-values must include all films and parallel-path effects explicitly.
type Layer struct {
	Name       string              `json:"name" yaml:"name"`
	Resistance units.Resistance    `json:"r_m2_k_w" yaml:"r_m2_k_w"`
	Evidence   provenance.Evidence `json:"evidence" yaml:"evidence"`
}

type InfiltrationSpec struct {
	Mode             string                         `json:"mode" yaml:"mode"` // ach_natural, ach50, explicit
	ACH              *float64                       `json:"ach,omitempty" yaml:"ach,omitempty"`
	ACH50            *float64                       `json:"ach50,omitempty" yaml:"ach50,omitempty"`
	ConversionFactor *float64                       `json:"conversion_factor,omitempty" yaml:"conversion_factor,omitempty"`
	Airflow          *units.Airflow                 `json:"airflow_m3_s,omitempty" yaml:"airflow_m3_s,omitempty"`
	Evidence         map[string]provenance.Evidence `json:"evidence" yaml:"evidence"`
}

type VentilationSpec struct {
	Airflow   *units.Airflow                 `json:"airflow_m3_s" yaml:"airflow_m3_s"`
	RoomRates map[string]units.Airflow       `json:"room_rates_m3_s,omitempty" yaml:"room_rates_m3_s,omitempty"`
	Evidence  map[string]provenance.Evidence `json:"evidence" yaml:"evidence"`
}
