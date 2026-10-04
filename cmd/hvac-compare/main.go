// SPDX-License-Identifier: AGPL-3.0-only
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/models/designload"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/project"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/validation"
	"os"
)

func main() { os.Exit(run()) }
func run() int {
	rel := flag.Float64("relative", 0.01, "relative tolerance")
	abs := flag.Float64("absolute", 0.5, "absolute tolerance in W")
	reportPath := flag.String("report", "validation-report.json", "output artifact")
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: hvac-compare [--relative 0.01] [--absolute 0.5] [--report report.json] PROJECT REFERENCE.json")
		return 2
	}
	p, err := project.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fixture, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	r, err := designload.New().Calculate(p.Building, p.Building.Design)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	raw, err := os.ReadFile(flag.Arg(1))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	var ref validation.Reference
	if err = json.Unmarshal(raw, &ref); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	report, err := validation.Compare(r, fixture, ref, *rel, *abs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	data, err := validation.EncodeReport(report)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err = os.WriteFile(*reportPath, append(data, '\n'), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(*reportPath)
	if !report.Pass {
		return 1
	}
	return 0
}
