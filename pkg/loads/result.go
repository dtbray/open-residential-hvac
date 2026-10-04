// SPDX-License-Identifier: AGPL-3.0-only
package loads

import (
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/provenance"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
)

type InputValue struct {
	Name        string                  `json:"name"`
	Value       float64                 `json:"value"`
	Unit        string                  `json:"unit"`
	Source      provenance.Source       `json:"source"`
	Assumptions []provenance.Assumption `json:"assumptions,omitempty"`
}

type ResultNode struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Value       units.HeatFlow          `json:"value_w"`
	Method      string                  `json:"method"`
	Equation    string                  `json:"equation"`
	Inputs      []InputValue            `json:"inputs,omitempty"`
	Sources     []provenance.Source     `json:"sources,omitempty"`
	Assumptions []provenance.Assumption `json:"assumptions,omitempty"`
	Warnings    []string                `json:"warnings,omitempty"`
	Children    []ResultNode            `json:"children,omitempty"`
}

type RoomResult struct {
	ID              string         `json:"id"`
	ZoneID          string         `json:"zone_id"`
	Name            string         `json:"name"`
	HeatingLoad     units.HeatFlow `json:"heating_load_w"`
	CoolingSensible units.HeatFlow `json:"cooling_sensible_w"`
	CoolingLatent   units.HeatFlow `json:"cooling_latent_w"`
}

type Result struct {
	Model           string         `json:"model"`
	HeatingLoad     units.HeatFlow `json:"heating_load_w"`
	CoolingSensible units.HeatFlow `json:"cooling_sensible_w"`
	CoolingLatent   units.HeatFlow `json:"cooling_latent_w"`
	CoolingLoad     units.HeatFlow `json:"cooling_load_w"`
	Heating         ResultNode     `json:"heating"`
	Cooling         ResultNode     `json:"cooling"`
	Rooms           []RoomResult   `json:"rooms"`
}

func aggregate(id, name string, children ...ResultNode) ResultNode {
	n := ResultNode{ID: id, Name: name, Method: "explicit aggregation", Equation: "Q = sum(children)", Children: children}
	for _, c := range children {
		n.Value += c.Value
		n.Assumptions = append(n.Assumptions, c.Assumptions...)
		n.Warnings = append(n.Warnings, c.Warnings...)
	}
	return n
}

func input(name string, v float64, unit string, e provenance.Evidence) InputValue {
	if !e.Source.Valid() {
		e.Source = provenance.Source{Kind: provenance.UserInput, Reference: "project input: " + name}
	}
	return InputValue{Name: name, Value: v, Unit: unit, Source: e.Source, Assumptions: e.Assumptions}
}

func leaf(id, name, method, equation string, v units.HeatFlow, in ...InputValue) ResultNode {
	n := ResultNode{ID: id, Name: name, Value: v, Method: method, Equation: equation, Inputs: in}
	for _, i := range in {
		n.Sources = append(n.Sources, i.Source)
		n.Assumptions = append(n.Assumptions, i.Assumptions...)
		if i.Source.Kind == provenance.Default {
			n.Warnings = append(n.Warnings, "Default input influences result: "+i.Name)
			if len(i.Assumptions) == 0 {
				n.Assumptions = append(n.Assumptions, provenance.Assumption{ID: "default/" + i.Name, Description: "Default value used for " + i.Name})
			}
		}
	}
	return n
}

func Find(n ResultNode, id string) *ResultNode {
	if n.ID == id {
		return &n
	}
	for _, c := range n.Children {
		if found := Find(c, id); found != nil {
			return found
		}
	}
	return nil
}
