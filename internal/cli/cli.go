// SPDX-License-Identifier: AGPL-3.0-only
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"

	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/loads"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/project"
)

func Run(args []string, out, errOut io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(errOut, "usage: hvac validate|load|explain PROJECT.{json,yaml} [--room ID_OR_NAME] [--format text|json] [--node ID] [--verbose]")
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
	if err := f.Parse(args[2:]); err != nil {
		return 2
	}
	if f.NArg() > 0 || (*format != "text" && *format != "json") {
		fmt.Fprintln(errOut, "invalid arguments or output format")
		return 2
	}
	p, err := project.Open(path)
	if err != nil {
		return failure(err, *format, errOut)
	}
	if command == "validate" {
		if e := loads.Validate(p.Building); len(e) > 0 {
			return failure(e, *format, errOut)
		}
		if *format == "json" {
			writeJSON(out, map[string]any{"valid": true, "version": p.Version})
		} else {
			fmt.Fprintln(out, "Valid project:", p.Building.Name)
		}
		return 0
	}
	r, err := loads.Calculate(p.Building)
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
			writeJSON(out, n)
		} else {
			tree(out, *n, "", true)
		}
		return 0
	}
	if *format == "json" {
		writeJSON(out, r)
		return 0
	}
	fmt.Fprintf(out, "Heating Load: %.0f Btu/h\nCooling Sensible: %.0f Btu/h\nCooling Latent: %.0f Btu/h\nCooling Load: %.0f Btu/h\n", r.HeatingLoad.Btuh(), r.CoolingSensible.Btuh(), r.CoolingLatent.Btuh(), r.CoolingLoad.Btuh())
	for _, v := range r.Rooms {
		fmt.Fprintf(out, "  %s: heating %.0f; cooling sensible %.0f; latent %.0f Btu/h\n", v.Name, v.HeatingLoad.Btuh(), v.CoolingSensible.Btuh(), v.CoolingLatent.Btuh())
	}
	if command == "explain" || *verbose {
		tree(out, r.Heating, "", true)
		tree(out, r.Cooling, "", true)
	} else {
		tree(out, r.Cooling, "", false)
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
func tree(w io.Writer, n loads.ResultNode, indent string, details bool) {
	fmt.Fprintf(w, "%s%s: %.1f Btu/h [%s]\n", indent, n.Name, n.Value.Btuh(), n.ID)
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
		tree(w, c, indent+strings.Repeat(" ", 2), details)
	}
}
