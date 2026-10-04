// SPDX-License-Identifier: AGPL-3.0-only
// Package provenance describes the origin and confidence of engineering inputs.
package provenance

type Kind string

// Methodology identifies calculation policy independently of input provenance.
type Methodology struct {
	ID        string `json:"id" yaml:"id"`
	Name      string `json:"name" yaml:"name"`
	Version   string `json:"version" yaml:"version"`
	Reference string `json:"reference" yaml:"reference"`
}

const (
	UserInput       Kind = "user_input"
	Derived         Kind = "derived"
	Default         Kind = "default"
	Dataset         Kind = "dataset"
	PublicReference Kind = "public_reference"
	Measurement     Kind = "measurement"
	Imported        Kind = "imported"
)

type Source struct {
	Kind      Kind   `json:"kind" yaml:"kind"`
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Reference string `json:"reference,omitempty" yaml:"reference,omitempty"`
	License   string `json:"license,omitempty" yaml:"license,omitempty"`
}

func (s Source) Valid() bool {
	switch s.Kind {
	case UserInput, Derived, Default, Dataset, PublicReference, Measurement, Imported:
		return true
	}
	return false
}

type Assumption struct {
	ID          string `json:"id" yaml:"id"`
	Description string `json:"description" yaml:"description"`
}

type Evidence struct {
	Source      Source       `json:"source" yaml:"source"`
	Assumptions []Assumption `json:"assumptions,omitempty" yaml:"assumptions,omitempty"`
}
