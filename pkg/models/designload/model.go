// SPDX-License-Identifier: AGPL-3.0-only
// Package designload implements the open simultaneous steady-state design-load
// methodology. It makes no claim of conformance with a national standard.
package designload

import (
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/models"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/provenance"
)

const ID = "designload"
const Version = "0.1"

// Model has no mutable state or hidden defaults. Configuration can be introduced
// when a real policy variation needs it; New needs no empty configuration object.
type Model struct{}

func New() *Model           { return &Model{} }
func (m *Model) ID() string { return ID }

func Methodology() provenance.Methodology {
	return provenance.Methodology{ID: ID, Name: "Open Design Load", Version: Version, Reference: "docs/methodology.md#designload-01"}
}

var _ models.Model = (*Model)(nil)
