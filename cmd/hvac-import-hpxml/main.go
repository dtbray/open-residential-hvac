// SPDX-License-Identifier: AGPL-3.0-only
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/hpxml"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/project"
	"os"
)

func main() {
	optionsPath := flag.String("options", "", "explicit template, solar and adjacent conditions JSON")
	output := flag.String("output", "project.json", "native project destination")
	flag.Parse()
	if flag.NArg() != 1 || *optionsPath == "" {
		fmt.Fprintln(os.Stderr, "usage: hvac-import-hpxml --options options.json --output project.json INPUT.xml")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *optionsPath, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(input, optionsPath, output string) error {
	raw, err := os.ReadFile(optionsPath)
	if err != nil {
		return err
	}
	var o hpxml.Options
	if err = json.Unmarshal(raw, &o); err != nil {
		return err
	}
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	imported, err := hpxml.ImportBasic(data, o)
	if err != nil {
		return err
	}
	data, err = project.Encode(project.Project{Version: 1, Building: imported.Building}, "json")
	if err != nil {
		return err
	}
	if err = os.WriteFile(output, data, 0644); err != nil {
		return err
	}
	for _, warning := range imported.Warnings {
		fmt.Fprintln(os.Stderr, warning)
	}
	return nil
}
