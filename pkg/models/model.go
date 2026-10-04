// SPDX-License-Identifier: AGPL-3.0-only
// Package models defines a small model boundary without a plugin registry.
package models

import (
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/building"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/climate"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/loads"
)

// Model consumes physical inputs and produces shared inspectable results.
// Model-specific validation and policy belong in the implementation.
type Model interface {
	ID() string
	Calculate(building.Building, climate.DesignConditions) (*loads.Result, error)
}
