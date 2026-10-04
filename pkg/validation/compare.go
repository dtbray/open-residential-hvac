// SPDX-License-Identifier: AGPL-3.0-only
// Package validation compares explicit independently sourced reference metrics.
package validation

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/loads"
	"math"
	"sort"
)

type Reference struct {
	Version          int                `json:"version"`
	Generator        string             `json:"generator"`
	GeneratorVersion string             `json:"generator_version"`
	FixtureSHA256    string             `json:"fixture_sha256"`
	SourceFiles      []string           `json:"source_files"`
	Notes            string             `json:"notes"`
	Metrics          map[string]float64 `json:"metrics_w"`
}
type Difference struct {
	Metric             string   `json:"metric"`
	Reference          float64  `json:"reference_w"`
	Candidate          float64  `json:"candidate_w"`
	Difference         float64  `json:"difference_w"`
	RelativeDifference *float64 `json:"relative_difference,omitempty"`
	Pass               bool     `json:"pass"`
}
type Report struct {
	Generator         string       `json:"reference_generator"`
	GeneratorVersion  string       `json:"reference_generator_version"`
	FixtureSHA256     string       `json:"fixture_sha256"`
	RelativeTolerance float64      `json:"relative_tolerance"`
	AbsoluteTolerance float64      `json:"absolute_tolerance_w"`
	Differences       []Difference `json:"differences"`
	Pass              bool         `json:"pass"`
}

func Hash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }
func Compare(result *loads.Result, fixture []byte, ref Reference, relative, absolute float64) (*Report, error) {
	if ref.Version != 1 || ref.Generator == "" || ref.GeneratorVersion == "" || len(ref.SourceFiles) == 0 || len(ref.Metrics) == 0 {
		return nil, fmt.Errorf("reference requires version 1, generator/version, source files and nonempty metrics")
	}
	if ref.FixtureSHA256 != Hash(fixture) {
		return nil, fmt.Errorf("reference fixture SHA256 mismatch")
	}
	if !finite(relative) || !finite(absolute) || relative < 0 || absolute < 0 {
		return nil, fmt.Errorf("tolerances must be finite and nonnegative")
	}
	values := map[string]float64{"heating_load_w": float64(result.HeatingLoad), "cooling_sensible_w": float64(result.CoolingSensible), "cooling_latent_w": float64(result.CoolingLatent), "cooling_load_w": float64(result.CoolingLoad)}
	var walk func(loads.ResultNode)
	walk = func(n loads.ResultNode) {
		values[n.ID] = float64(n.Value)
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(result.Heating)
	walk(result.Cooling)
	r := &Report{Generator: ref.Generator, GeneratorVersion: ref.GeneratorVersion, FixtureSHA256: ref.FixtureSHA256, RelativeTolerance: relative, AbsoluteTolerance: absolute, Pass: true}
	keys := make([]string, 0, len(ref.Metrics))
	for k := range ref.Metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		want := ref.Metrics[key]
		got, ok := values[key]
		if !ok {
			return nil, fmt.Errorf("unknown reference metric %s", key)
		}
		if !finite(want) {
			return nil, fmt.Errorf("nonfinite reference metric %s", key)
		}
		diff := got - want
		pass := math.Abs(diff) <= math.Max(absolute, math.Abs(want)*relative)
		d := Difference{Metric: key, Reference: want, Candidate: got, Difference: diff, Pass: pass}
		if want != 0 {
			v := diff / math.Abs(want)
			d.RelativeDifference = &v
		}
		r.Differences = append(r.Differences, d)
		r.Pass = r.Pass && pass
	}
	return r, nil
}
func finite(v float64) bool                  { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func EncodeReport(r *Report) ([]byte, error) { return json.MarshalIndent(r, "", "  ") }
