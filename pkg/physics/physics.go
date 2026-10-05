// SPDX-License-Identifier: AGPL-3.0-only
// Package physics contains reusable thermal primitives with canonical SI types.
// Functions preserve signs and impose no margins, clipping, airflow allocation,
// climate percentiles, standards defaults, or model aggregation policies.
package physics

import "git.thomas-bray.com/thomas/open-residential-hvac/pkg/units"

func Conduction(u units.Transmittance, area units.Area, delta units.TemperatureDifference) units.HeatFlow {
	return units.HeatFlow(float64(u) * float64(area) * float64(delta))
}

func SolarGain(area units.Area, irradiance units.Irradiance, shgc, shading float64) units.HeatFlow {
	return units.HeatFlow(float64(area) * float64(irradiance) * shgc * shading)
}

// AirExchange converts indoor volumetric flow to dry-air mass flow.
func AirExchange(flow units.Airflow, density units.DryAirDensity) units.MassFlow {
	return units.MassFlow(float64(density) * float64(flow))
}

func SensibleAirLoad(massFlow units.MassFlow, cp units.SpecificHeat, delta units.TemperatureDifference) units.HeatFlow {
	return units.HeatFlow(float64(massFlow) * float64(cp) * float64(delta))
}

func LatentAirLoad(massFlow units.MassFlow, latent units.SpecificEnergy, delta units.HumidityRatio) units.HeatFlow {
	return units.HeatFlow(float64(massFlow) * float64(latent) * float64(delta))
}

func InternalGain(count int, perPerson units.HeatFlow) units.HeatFlow {
	return units.HeatFlow(float64(count) * float64(perPerson))
}
