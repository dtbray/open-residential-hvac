# International load-calculation architecture addendum

This addendum preserves the original MVP scope and makes its calculation
architecture explicitly standards-neutral. The engine describes physical buildings;
national standards and guidance can supply future calculation methodologies. The
project is an open residential thermal-load engine with inspectable calculations
and provenance, with no national-standard conformance claim.

## Layers and implemented boundary

Building description → physical properties → thermal primitives → calculation
model → result tree.

The building/zone/room/surface model stores physical area, volume, assemblies,
glazing properties, orientation, boundaries, occupancy, infiltration, ventilation
and project design conditions. It has no construction-table identifiers, national
climate percentiles, heat-transfer multipliers or standard-specific documents.

Assemblies resolve direct transmittance or explicit series resistance. Shared
psychrometrics resolve physical air properties. `physics` exposes signed conduction,
glazing solar gain, volume-to-mass air exchange, sensible/latent air transfer and
internal gain. These functions have typed SI inputs and do not decide defaults,
clipping, airflow allocation, design climates, safety margins or aggregation.

`models/designload` is the only supported methodology. It chooses applicable
primitives, supported inputs/boundaries, seasonal conditions, positive-load policy,
airflow allocation and room/zone/block aggregation. Model-specific validation lives
with the model rather than in the shared results package. `loads` holds reusable
result/error contracts and trace-tree helpers, and imports no model.

```go
model := designload.New()
result, err := model.Calculate(building, conditions)
```

Explicit physical conditions are authoritative, even if the building contains a
different stored project design. Calculation does not mutate either input. There
is no empty configuration object or configurable standards profile without an
actual use case. Future real model policy variations can justify configuration.

The small `models.Model` interface contains `ID()` and
`Calculate(building.Building, climate.DesignConditions)`. There is no plugin
registry, discovery, loading protocol or provider lifecycle. Conceptual future
IDs include `en12831`, `iso52016`, `csaf280`, and `acca`; no packages, selectable
options or conformance labels for these methods have been implemented. Future UK,
Australian and other regional guidance can fit this boundary when defensible
methods, source rights and model-specific inputs are available.

## Physical units, climate and boundaries

Canonical quantities use SI. Absolute temperature, temperature difference, area,
volume, heat flow, airflow, resistance, transmittance, irradiance, pressure, air
properties and azimuth have distinct Go types. Imperial callers can use explicit
conversions at the library/adapter boundary; neither the model interface nor its
physical primitives takes customary-unit raw parameters. Native version-1 project
files remain explicitly SI, so no ambiguous mixed-unit JSON/YAML dialect is added.

Text output defaults to watts and supports `--units ip` for Btu/h. JSON remains
canonical SI. Metric/imperial library inputs have a numerical equivalence test.
The HPXML adapter converts its explicitly declared customary input units before
calling a model. File-format extensions and weather resolvers belong at boundaries,
not in methodology-specific versions of the physical types.

Climate stores physical design temperatures, wet bulb or RH, pressure and input
provenance. It does not store standard-specific weather lookup identifiers or
constrain selection to a national design percentile.

Boundaries include outdoors, ground, conditioned space, generic unconditioned
space, adjacent building, attic, garage, crawlspace, basement and adjacent unit.
`designload` requires explicit seasonal temperatures for unconditioned/adjacent
boundaries; it does not infer them from surface type. Ground remains unsupported.
Optional azimuth is clockwise from true north in degrees; descriptive cardinal
orientation is retained for existing projects. Neither infers solar exposure in
the initial model; plane irradiance is an explicit physical input.

## Airtightness, ventilation and heat recovery

Infiltration and ventilation remain independent. Measured or assumed airtightness
can be preserved separately from selected infiltration airflow, with air-change
rate, test pressure and evidence. Test pressure is a physical value, not fixed to
one country's conventions. `designload` warns that stored test results are not
converted automatically; it uses the explicitly chosen airflow mode and its
traceable inputs. An absent measurement is different from explicit zero.

Mechanical ventilation can retain building/room airflow and optional sensible/
latent heat-recovery efficiencies with evidence. Serialization preserves these
properties even before a model implements them. `designload` returns an explicit
unsupported error when heat-recovery properties are supplied; it does not silently
apply a no-recovery approximation to a specified device.

## Methodology and results

Every complete calculation result and comparison report includes:

```json
{
  "methodology": {
    "id": "designload",
    "name": "Open Design Load",
    "version": "0.1",
    "reference": "docs/methodology.md#designload-01"
  }
}
```

This identifies model policy independently of input provenance: weather data,
assembly measurements, assumptions and imported properties retain their own
sources. Stable result IDs and explicit aggregation remain shared contracts.
An isolated CLI JSON node response wraps `node` with `methodology` so the
calculation context survives selection. UI results display the actual methodology
and use generic heating-design, cooling-design, sensible/latent and envelope labels.

`hvac load house.yaml --model designload` selects the supported model explicitly.
Unsupported national-model IDs are errors. Defaults still select `designload`;
the CLI does not need a registry for one real implementation.

## Dynamic models and validation

Hourly simulation is outside v0.1. Thermal mass, capacitance, time-series weather,
solar, internal gains, control setpoints and inter-zone coupling are not added
speculatively. Future dynamic methods can extend physical inputs and introduce a
richer temporal input/result API without making the building a steady-state
document. The current small design-condition interface is not a universal hourly
simulation contract.

`testdata/designload` scopes the existing load fixtures/references to this model;
`testdata/physics` contains primitive psychrometric references and `testdata/openstudio`
remains reserved for real US-fixture oracle outputs. Future models get their own
validation suites, source methodology, assumptions and numerical tolerances.
Different methods need not produce identical results. Comparison should identify
methodological differences rather than normalize them away.

The architectural refactor preserves the existing model's numerical loads and
node identities. This is a regression requirement for unchanged physics, not a
requirement that future national or dynamic models agree with `designload`.
Real OpenStudio-HPXML comparisons remain an outstanding MVP acceptance gate.

## Development API migration

Replace the preview `loads.Calculate(b)` API with
`designload.New().Calculate(b, b.Design)`, or pass independently selected conditions.
`loads.Validate(b)` becomes `designload.New().Validate(b, conditions)` because
input support is model policy. Shared result/error types remain in `loads`.

Result `model` now identifies `designload`; the versioned `methodology` field is
the authoritative method descriptor. CLI text defaults change from Btu/h to W;
use `--units ip` to retain customary output. JSON isolated-node output now has a
methodology envelope. Native project version and existing project field values
are preserved; new physical metadata fields are optional. Fixture paths have
moved under `testdata/designload`.

