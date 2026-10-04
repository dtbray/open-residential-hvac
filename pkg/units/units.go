// SPDX-License-Identifier: AGPL-3.0-only
// Package units supplies distinct canonical SI types and explicit conversions.
package units

import "math"

// Temperatures are degrees Celsius. All other types are SI.
type Temperature float64
type Area float64
type Volume float64
type HeatFlow float64
type Airflow float64       // m³/s
type Resistance float64    // m² K/W
type Transmittance float64 // W/(m² K)
type Irradiance float64    // W/m²
type Pressure float64      // Pa

const (
	SquareFootToM2   = 0.09290304
	CubicFootToM3    = 0.028316846592
	BtuPerHourToW    = 0.2930710701722222 // international-table Btu
	CFMToM3PerSecond = CubicFootToM3 / 60
	IPUToSI          = BtuPerHourToW / SquareFootToM2 * 1.8
)

func Fahrenheit(v float64) Temperature     { return Temperature((v - 32) / 1.8) }
func (t Temperature) Fahrenheit() float64  { return float64(t)*1.8 + 32 }
func Celsius(v float64) Temperature        { return Temperature(v) }
func SquareFeet(v float64) Area            { return Area(v * SquareFootToM2) }
func (a Area) SquareFeet() float64         { return float64(a) / SquareFootToM2 }
func SquareMetres(v float64) Area          { return Area(v) }
func CubicFeet(v float64) Volume           { return Volume(v * CubicFootToM3) }
func (v Volume) CubicFeet() float64        { return float64(v) / CubicFootToM3 }
func Btuh(v float64) HeatFlow              { return HeatFlow(v * BtuPerHourToW) }
func (h HeatFlow) Btuh() float64           { return float64(h) / BtuPerHourToW }
func Watts(v float64) HeatFlow             { return HeatFlow(v) }
func CFM(v float64) Airflow                { return Airflow(v * CFMToM3PerSecond) }
func (a Airflow) CFM() float64             { return float64(a) / CFMToM3PerSecond }
func LitresPerSecond(v float64) Airflow    { return Airflow(v / 1000) }
func (a Airflow) LitresPerSecond() float64 { return float64(a) * 1000 }
func UFactorIP(v float64) Transmittance    { return Transmittance(v * IPUToSI) }
func RValueIP(v float64) Resistance        { return Resistance(v / IPUToSI) }
func Finite(v float64) bool                { return !math.IsNaN(v) && !math.IsInf(v, 0) }
