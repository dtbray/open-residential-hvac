// SPDX-License-Identifier: AGPL-3.0-only
// Package hpxml is an adapter boundary, never an engineering calculation model.
package hpxml

import (
	"encoding/xml"
	"fmt"

	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/provenance"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
)

// Options supplies design/airflow/internal inputs that this deliberately narrow
// adapter cannot infer. Template must contain exactly one native room. Imported
// geometry replaces the template geometry. Solar must be supplied per window ID.
type Options struct {
	Template building.Building
	Solar    map[string]SolarInput
}
type SolarInput struct {
	Irradiance    units.Irradiance
	ShadingFactor float64
	Evidence      provenance.Evidence
}
type ImportResult struct {
	Building building.Building
	Warnings []string
}

type identifier struct {
	ID string `xml:"id,attr"`
}
type reference struct {
	ID string `xml:"idref,attr"`
}
type xmlWall struct {
	ID       identifier `xml:"SystemIdentifier"`
	Area     *float64   `xml:"Area"`
	Interior string     `xml:"InteriorAdjacentTo"`
	Exterior string     `xml:"ExteriorAdjacentTo"`
	R        *float64   `xml:"Insulation>AssemblyEffectiveRValue"`
}
type xmlWindow struct {
	ID   identifier `xml:"SystemIdentifier"`
	Wall reference  `xml:"AttachedToWall"`
	Area *float64   `xml:"Area"`
	U    *float64   `xml:"UFactor"`
	SHGC *float64   `xml:"SHGC"`
}
type xmlBuilding struct {
	ID              identifier  `xml:"BuildingID"`
	Area            *float64    `xml:"BuildingDetails>BuildingSummary>BuildingConstruction>ConditionedFloorArea"`
	Volume          *float64    `xml:"BuildingDetails>BuildingSummary>BuildingConstruction>ConditionedBuildingVolume"`
	Walls           []xmlWall   `xml:"BuildingDetails>Enclosure>Walls>Wall"`
	Windows         []xmlWindow `xml:"BuildingDetails>Enclosure>Windows>Window"`
	Roofs           []struct{}  `xml:"BuildingDetails>Enclosure>Roofs>Roof"`
	Floors          []struct{}  `xml:"BuildingDetails>Enclosure>Floors>Floor"`
	Slabs           []struct{}  `xml:"BuildingDetails>Enclosure>Slabs>Slab"`
	Doors           []struct{}  `xml:"BuildingDetails>Enclosure>Doors>Door"`
	RimJoists       []struct{}  `xml:"BuildingDetails>Enclosure>RimJoists>RimJoist"`
	FoundationWalls []struct{}  `xml:"BuildingDetails>Enclosure>FoundationWalls>FoundationWall"`
}

// ImportBasic supports a single conditioned block with exterior walls and
// attached windows. Other major envelope categories are errors, never skipped.
// Full HPXML schema validation and ordinary-house import remain future work.
func ImportBasic(data []byte, o Options) (*ImportResult, error) {
	var doc struct {
		XMLName   xml.Name      `xml:"HPXML"`
		Buildings []xmlBuilding `xml:"Building"`
	}
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("HPXML: %w", err)
	}
	if len(doc.Buildings) != 1 || len(o.Template.Zones) != 1 || len(o.Template.Zones[0].Rooms) != 1 {
		return nil, fmt.Errorf("basic HPXML import requires one building and one template room")
	}
	b := doc.Buildings[0]
	if len(b.Roofs)+len(b.Floors)+len(b.Slabs)+len(b.Doors)+len(b.RimJoists)+len(b.FoundationWalls) > 0 {
		return nil, fmt.Errorf("basic HPXML adapter does not support roofs, floors, slabs, doors, rim joists, or foundation walls")
	}
	if b.Area == nil || b.Volume == nil || *b.Area <= 0 || *b.Volume <= 0 {
		return nil, fmt.Errorf("HPXML requires positive conditioned floor area and volume")
	}
	native := o.Template
	native.Zones = append([]building.Zone{}, o.Template.Zones...)
	native.Zones[0].Rooms = append([]building.Room{}, o.Template.Zones[0].Rooms...)
	room := native.Zones[0].Rooms[0]
	room.FloorArea = units.SquareFeet(*b.Area)
	room.Volume = units.CubicFeet(*b.Volume)
	room.Walls = nil
	room.Windows = nil
	room.Doors = nil
	room.Floors = nil
	room.Ceilings = nil
	native.Assemblies = nil
	source := provenance.Evidence{Source: provenance.Source{Kind: provenance.Imported, Name: "HPXML supplied document", Reference: "HPXML areas in ft2, volume in ft3, R in h ft2 F/Btu, U in Btu/(h ft2 F)"}}
	room.Evidence = map[string]provenance.Evidence{}
	for k, v := range o.Template.Zones[0].Rooms[0].Evidence {
		room.Evidence[k] = v
	}
	room.Evidence["floor_area_m2"] = source
	room.Evidence["volume_m3"] = source
	for _, w := range b.Walls {
		if w.ID.ID == "" || w.Area == nil || w.R == nil || *w.Area <= 0 || *w.R <= 0 {
			return nil, fmt.Errorf("HPXML wall requires ID, positive Area and AssemblyEffectiveRValue")
		}
		if w.Interior != "conditioned space" || w.Exterior != "outside" {
			return nil, fmt.Errorf("HPXML wall %s: only conditioned space / outside supported", w.ID.ID)
		}
		assembly := "hpxml-assembly-" + w.ID.ID
		u := units.UFactorIP(1 / *w.R)
		native.Assemblies = append(native.Assemblies, building.Assembly{ID: assembly, Name: assembly, UFactor: &u, Evidence: source})
		room.Walls = append(room.Walls, building.Surface{ID: w.ID.ID, Name: w.ID.ID, Area: units.SquareFeet(*w.Area), Adjacent: building.Outdoors, Assembly: assembly, TiltDegrees: 90, Evidence: map[string]provenance.Evidence{"area_m2": source}})
	}
	for _, w := range b.Windows {
		if w.ID.ID == "" || w.Area == nil || w.U == nil || w.SHGC == nil {
			return nil, fmt.Errorf("HPXML window requires ID, Area, UFactor and SHGC")
		}
		s, ok := o.Solar[w.ID.ID]
		if !ok {
			return nil, fmt.Errorf("HPXML window %s requires explicit design-plane solar inputs", w.ID.ID)
		}
		assembly := "hpxml-assembly-" + w.ID.ID
		u := units.UFactorIP(*w.U)
		native.Assemblies = append(native.Assemblies, building.Assembly{ID: assembly, Name: assembly, UFactor: &u, Evidence: source})
		room.Windows = append(room.Windows, building.Window{Surface: building.Surface{ID: w.ID.ID, Name: w.ID.ID, Area: units.SquareFeet(*w.Area), Adjacent: building.Outdoors, Assembly: assembly, TiltDegrees: 90, Evidence: map[string]provenance.Evidence{"area_m2": source, "shgc": source, "incident_solar_w_m2": s.Evidence, "shading_factor": s.Evidence}}, ParentSurface: w.Wall.ID, SHGC: w.SHGC, IncidentSolar: &s.Irradiance, ShadingFactor: &s.ShadingFactor})
	}
	if len(room.Walls) == 0 {
		return nil, fmt.Errorf("HPXML requires at least one supported wall")
	}
	native.Zones[0].Rooms[0] = room
	return &ImportResult{Building: native, Warnings: []string{"Basic adapter imports only exterior walls, windows, conditioned floor area and volume; design conditions, airflow, internal gains and solar come from explicit options.", "HPXML room geometry, shading, framing paths, ducts, systems and schedules are not imported; full HPXML schema validation has not been performed."}}, nil
}
