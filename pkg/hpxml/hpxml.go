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
	Adjacent map[string]AdjacentConditions
}

// AdjacentConditions explicitly supplies seasonal temperatures for shared/unconditioned boundaries.
type AdjacentConditions struct {
	Heating  units.Temperature
	Cooling  units.Temperature
	Evidence provenance.Evidence
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
	Azimuth  *float64   `xml:"Azimuth"`
}
type xmlWindow struct {
	ID   identifier `xml:"SystemIdentifier"`
	Wall reference  `xml:"AttachedToWall"`
	Area *float64   `xml:"Area"`
	U    *float64   `xml:"UFactor"`
	SHGC *float64   `xml:"SHGC"`
}
type xmlFloor struct {
	xmlWall
	Kind string `xml:"FloorOrCeiling"`
}
type xmlDoor struct {
	ID   identifier `xml:"SystemIdentifier"`
	Wall reference  `xml:"AttachedToWall"`
	Area *float64   `xml:"Area"`
	R    *float64   `xml:"RValue"`
}
type xmlBuilding struct {
	ID              identifier  `xml:"BuildingID"`
	Area            *float64    `xml:"BuildingDetails>BuildingSummary>BuildingConstruction>ConditionedFloorArea"`
	Volume          *float64    `xml:"BuildingDetails>BuildingSummary>BuildingConstruction>ConditionedBuildingVolume"`
	Walls           []xmlWall   `xml:"BuildingDetails>Enclosure>Walls>Wall"`
	Windows         []xmlWindow `xml:"BuildingDetails>Enclosure>Windows>Window"`
	Roofs           []struct{}  `xml:"BuildingDetails>Enclosure>Roofs>Roof"`
	Floors          []xmlFloor  `xml:"BuildingDetails>Enclosure>Floors>Floor"`
	Slabs           []struct{}  `xml:"BuildingDetails>Enclosure>Slabs>Slab"`
	Doors           []xmlDoor   `xml:"BuildingDetails>Enclosure>Doors>Door"`
	RimJoists       []struct{}  `xml:"BuildingDetails>Enclosure>RimJoists>RimJoist"`
	FoundationWalls []struct{}  `xml:"BuildingDetails>Enclosure>FoundationWalls>FoundationWall"`
}

// ImportBasic supports a single conditioned block with exterior walls and
// attached windows, doors, and supported floors/ceilings. Roof/ground categories
// remain errors, never skipped. Schema validation belongs to the validation harness.
// Standalone HPXML schema validation and complete single-family import remain future work.
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
	if len(b.Roofs)+len(b.Slabs)+len(b.RimJoists)+len(b.FoundationWalls) > 0 {
		return nil, fmt.Errorf("basic HPXML adapter does not support roofs, slabs, rim joists, or foundation walls")
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
		if w.Interior != "conditioned space" {
			return nil, fmt.Errorf("HPXML wall %s: interior must be conditioned space", w.ID.ID)
		}
		assembly := "hpxml-assembly-" + w.ID.ID
		u := units.UFactorIP(1 / *w.R)
		native.Assemblies = append(native.Assemblies, building.Assembly{ID: assembly, Name: assembly, UFactor: &u, Evidence: source})
		surface := building.Surface{ID: w.ID.ID, Name: w.ID.ID, Area: units.SquareFeet(*w.Area), Assembly: assembly, TiltDegrees: 90, Evidence: map[string]provenance.Evidence{"area_m2": source}}
		if err := setBoundary(&surface, w.Exterior, o); err != nil {
			return nil, err
		}
		if w.Azimuth != nil {
			a := units.Azimuth(*w.Azimuth)
			surface.Azimuth = &a
		}
		room.Walls = append(room.Walls, surface)
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

	for _, f := range b.Floors {
		if f.ID.ID == "" || f.Area == nil || f.R == nil || *f.Area <= 0 || *f.R <= 0 || f.Interior != "conditioned space" {
			return nil, fmt.Errorf("HPXML floor/ceiling requires ID, positive area/R and conditioned interior")
		}
		if f.Kind != "floor" && f.Kind != "ceiling" {
			return nil, fmt.Errorf("HPXML floor %s requires FloorOrCeiling", f.ID.ID)
		}
		assembly := "hpxml-assembly-" + f.ID.ID
		u := units.UFactorIP(1 / *f.R)
		native.Assemblies = append(native.Assemblies, building.Assembly{ID: assembly, Name: assembly, UFactor: &u, Evidence: source})
		surface := building.Surface{ID: f.ID.ID, Name: f.ID.ID, Area: units.SquareFeet(*f.Area), Assembly: assembly, Evidence: map[string]provenance.Evidence{"area_m2": source}}
		if err := setBoundary(&surface, f.Exterior, o); err != nil {
			return nil, err
		}
		if f.Kind == "floor" {
			room.Floors = append(room.Floors, surface)
		} else {
			room.Ceilings = append(room.Ceilings, surface)
		}
	}
	for _, door := range b.Doors {
		if door.ID.ID == "" || door.Area == nil || door.R == nil || *door.Area <= 0 || *door.R <= 0 {
			return nil, fmt.Errorf("HPXML door requires ID and positive area/R")
		}
		var parent *building.Surface
		for i := range room.Walls {
			if room.Walls[i].ID == door.Wall.ID {
				parent = &room.Walls[i]
				break
			}
		}
		if parent == nil {
			return nil, fmt.Errorf("HPXML door %s references unknown wall", door.ID.ID)
		}
		assembly := "hpxml-assembly-" + door.ID.ID
		u := units.UFactorIP(1 / *door.R)
		native.Assemblies = append(native.Assemblies, building.Assembly{ID: assembly, Name: assembly, UFactor: &u, Evidence: source})
		surface := *parent
		surface.ID = door.ID.ID
		surface.Name = door.ID.ID
		surface.Area = units.SquareFeet(*door.Area)
		surface.Assembly = assembly
		room.Doors = append(room.Doors, building.Door{Surface: surface, ParentSurface: parent.ID})
	}
	if len(room.Walls) == 0 {
		return nil, fmt.Errorf("HPXML requires at least one supported wall")
	}
	native.Zones[0].Rooms[0] = room
	return &ImportResult{Building: native, Warnings: []string{"Basic adapter imports supported walls, windows, doors, floors/ceilings, conditioned floor area and volume; design conditions, airflow, internal gains and solar come from explicit options.", "HPXML room geometry, shading, framing paths, ducts, systems and schedules are not imported; full HPXML schema validation has not been performed."}}, nil
}

func setBoundary(surface *building.Surface, exterior string, o Options) error {
	if exterior == "outside" {
		surface.Adjacent = building.Outdoors
		return nil
	}
	boundaries := map[string]building.Adjacency{"other housing unit": building.AdjacentUnit, "attic - unvented": building.Attic, "attic - vented": building.Attic, "garage": building.Garage, "crawlspace - vented": building.Crawlspace, "crawlspace - unvented": building.Crawlspace}
	boundary, ok := boundaries[exterior]
	if !ok {
		return fmt.Errorf("HPXML surface %s: unsupported exterior %q", surface.ID, exterior)
	}
	adjacent, ok := o.Adjacent[surface.ID]
	if !ok {
		return fmt.Errorf("HPXML surface %s requires explicit adjacent temperatures", surface.ID)
	}
	surface.Adjacent = boundary
	surface.HeatingAdjacentDB = &adjacent.Heating
	surface.CoolingAdjacentDB = &adjacent.Cooling
	surface.Evidence["heating_adjacent_db_c"] = adjacent.Evidence
	surface.Evidence["cooling_adjacent_db_c"] = adjacent.Evidence
	return nil
}
