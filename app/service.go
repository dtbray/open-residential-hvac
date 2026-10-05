// SPDX-License-Identifier: AGPL-3.0-only
// Package app is the UI-neutral application boundary; engineering stays in pkg.
package app

import (
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/loads"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/models/designload"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/project"
)

type Service struct{}

func (s *Service) Calculate(data string) (*loads.Result, error) {
	p, err := project.Decode([]byte(data), "json")
	if err != nil {
		return nil, err
	}
	return designload.New().Calculate(p.Building, p.Building.Design)
}
func (s *Service) Validate(data string) (loads.ValidationErrors, error) {
	p, err := project.Decode([]byte(data), "json")
	if err != nil {
		return nil, err
	}
	return designload.New().Validate(p.Building, p.Building.Design), nil
}
