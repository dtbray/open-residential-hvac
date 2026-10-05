// SPDX-License-Identifier: AGPL-3.0-only
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/models/designload"
	"io"
	"strings"

	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/loads"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/project"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
)

func Run(args []string, out, errOut io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(errOut, "usage: hvac validate|load|explain PROJECT.{json,yaml} [--model designload] [--units si|ip] [--room ID_OR_NAME] [--format text|json] [--node ID] [--verbose]")
		return 2
	}
	command, path := args[0], args[1]
	if command != "validate" && command != "load" && command != "explain" {
		fmt.Fprintln(errOut, "unknown command:", command)
		return 2
	}
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	f.SetOutput(errOut)
	format := f.String("format", "text", "text or json")
	room := f.String("room", "", "room ID or unique name")
	node := f.String("node", "", "stable result node ID")
	verbose := f.Bool("verbose", false, "display all calculations")
	modelID := f.String("model", designload.ID, "calculation model (designload)")
	unitSystem := f.String("units", "si", "text heat-flow units: si (W) or ip (Btu/h); JSON always uses SI")
	if err := f.Parse(args[2:]); err != nil {
		return 2
	}
	if f.NArg() > 0 || (*format != "text" && *format != "json") {
		fmt.Fprintln(errOut, "invalid arguments or output format")
		return 2
	}
	if *modelID != designload.ID {
		return failure(fmt.Errorf("unsupported calculation model %q; supported: designload", *modelID), *format, errOut)
	}
	if *unitSystem != "si" && *unitSystem != "ip" {
		fmt.Fprintln(errOut, "units must be si or ip")
		return 2
	}
	p, err := project.Open(path)
	if err != nil {
		return failure(err, *format, errOut)
	}
	if command == "validate" {
		if e := designload.New().Validate(p.Building, p.Building.Design); len(e) > 0 {
			return failure(e, *format, errOut)
		}
		if *format == "json" {
			writeJSON(out, map[string]any{"valid": true, "version": p.Version, "methodology": designload.Methodology()})
		} else {
			fmt.Fprintln(out, "Valid project:", p.Building.Name)
		}
		return 0
	}
	r, err := designload.New().Calculate(p.Building, p.Building.Design)
	if err != nil {
		return failure(err, *format, errOut)
	}
	if *room != "" {
		var matches []loads.RoomResult
		for _, v := range r.Rooms {
			if v.ID == *room || v.Name == *room {
				matches = append(matches, v)
			}
		}
		if len(matches) != 1 {
			return failure(fmt.Errorf("room %q must identify exactly one room; found %d", *room, len(matches)), *format, errOut)
		}
		v := matches[0]
		h := loads.Find(r.Heating, "room/"+v.ID+"/heating")
		c := loads.Find(r.Cooling, "room/"+v.ID+"/cooling")
		r.Heating = *h
		r.Cooling = *c
		r.Rooms = matches
		r.HeatingLoad = v.HeatingLoad
		r.CoolingSensible = v.CoolingSensible
		r.CoolingLatent = v.CoolingLatent
		r.CoolingLoad = c.Value
	}
	if *node != "" {
		n := loads.Find(r.Heating, *node)
		if n == nil {
			n = loads.Find(r.Cooling, *node)
		}
		if n == nil {
			return failure(fmt.Errorf("unknown result node %q", *node), *format, errOut)
		}
		if *format == "json" {
			writeJSON(out, map[string]any{"methodology": r.Methodology, "node": n})
		} else {
			fmt.Fprintf(out, "Calculation method: %s %s (%s)\n", r.Methodology.Name, r.Methodology.Version, r.Methodology.ID)
			tree(out, *n, "", true, *unitSystem)
		}
		return 0
	}
	if *format == "json" {
		writeJSON(out, r)
		return 0
	}
	value := func(q units.HeatFlow) float64 { return heatValue(q, *unitSystem) }
	unit := heatUnit(*unitSystem)
	fmt.Fprintf(out, "Calculation method: %s %s (%s)\n", r.Methodology.Name, r.Methodology.Version, r.Methodology.ID)
	fmt.Fprintf(out, "Heating design load: %.1f %s\nSensible cooling load: %.1f %s\nLatent cooling load: %.1f %s\nCooling design load: %.1f %s\n", value(r.HeatingLoad), unit, value(r.CoolingSensible), unit, value(r.CoolingLatent), unit, value(r.CoolingLoad), unit)
	for _, v := range r.Rooms {
		fmt.Fprintf(out, "  %s: heating %.1f; cooling sensible %.1f; latent %.1f %s\n", v.Name, value(v.HeatingLoad), value(v.CoolingSensible), value(v.CoolingLatent), unit)
	}
	if command == "explain" || *verbose {
		tree(out, r.Heating, "", true, *unitSystem)
		tree(out, r.Cooling, "", true, *unitSystem)
	} else {
		tree(out, r.Cooling, "", false, *unitSystem)
	}
	seen := map[string]bool{}
	for _, root := range []loads.ResultNode{r.Heating, r.Cooling} {
		for _, w := range root.Warnings {
			if !seen[w] {
				fmt.Fprintln(out, "Warning:", w)
				seen[w] = true
			}
		}
		for _, a := range root.Assumptions {
			if !seen[a.ID] {
				fmt.Fprintln(out, "Assumption:", a.Description)
				seen[a.ID] = true
			}
		}
	}
	return 0
}

func failure(err error, format string, w io.Writer) int {
	if format == "json" {
		if e, ok := err.(loads.ValidationErrors); ok {
			writeJSON(w, map[string]any{"valid": false, "errors": e})
		} else {
			writeJSON(w, map[string]any{"valid": false, "error": err.Error()})
		}
	} else {
		fmt.Fprintln(w, err)
	}
	return 1
}

func writeJSON(w io.Writer, v any) { e := json.NewEncoder(w); e.SetIndent("", "  "); _ = e.Encode(v) }
func heatValue(q units.HeatFlow, system string) float64 {
	if system == "ip" {
		return q.Btuh()
	}
	return float64(q)
}
func heatUnit(system string) string {
	if system == "ip" {
		return "Btu/h"
	}
	return "W"
}
func tree(w io.Writer, n loads.ResultNode, indent string, details bool, system string) {
	fmt.Fprintf(w, "%s%s: %.1f %s [%s]\n", indent, n.Name, heatValue(n.Value, system), heatUnit(system), n.ID)
	if details && len(n.Children) == 0 {
		fmt.Fprintln(w, indent+"  Method: "+n.Method)
		fmt.Fprintln(w, indent+"  Equation: "+n.Equation)
		for _, i := range n.Inputs {
			fmt.Fprintf(w, "%s  %s: %.8g %s; %s (%s)\n", indent, i.Name, i.Value, i.Unit, i.Source.Kind, i.Source.Reference)
			for _, a := range i.Assumptions {
				fmt.Fprintln(w, indent+"    Assumption: "+a.Description)
			}
		}
	}
	for _, c := range n.Children {
		tree(w, c, indent+strings.Repeat(" ", 2), details, system)
	}
}
