// SPDX-License-Identifier: AGPL-3.0-only
package climate

import (
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/provenance"
	"git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"
)

type DesignConditions struct {
	Heating  Heating                        `json:"heating" yaml:"heating"`
	Cooling  Cooling                        `json:"cooling" yaml:"cooling"`
	Pressure *units.Pressure                `json:"pressure_pa" yaml:"pressure_pa"`
	Evidence map[string]provenance.Evidence `json:"evidence" yaml:"evidence"`
}

type Heating struct {
	OutdoorDB *units.Temperature `json:"outdoor_db_c" yaml:"outdoor_db_c"`
	IndoorDB  *units.Temperature `json:"indoor_db_c" yaml:"indoor_db_c"`
}

type Cooling struct {
	OutdoorDB *units.Temperature `json:"outdoor_db_c" yaml:"outdoor_db_c"`
	OutdoorWB *units.Temperature `json:"outdoor_wb_c,omitempty" yaml:"outdoor_wb_c,omitempty"`
	OutdoorRH *float64           `json:"outdoor_rh_fraction,omitempty" yaml:"outdoor_rh_fraction,omitempty"`
	IndoorDB  *units.Temperature `json:"indoor_db_c" yaml:"indoor_db_c"`
	IndoorRH  *float64           `json:"indoor_rh_fraction" yaml:"indoor_rh_fraction"`
}
