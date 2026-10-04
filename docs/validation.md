# Validation status

v0.1 is not declared complete. There are no committed OpenStudio-HPXML design-load
results, so no agreement with that oracle is claimed. There is no evidence yet
that this simplified model meets any industry sizing accuracy requirement.

## Tests currently available

Unit conversions use absolute tolerance 1e-10 in the target numerical unit.
Conduction/solar/airflow tests use explicit independent hand results. Integration
component and aggregate checks use absolute tolerance 1e-8 W. Serialization checks
require exact deep equality through JSON and YAML. Repeated calculations require
identical serialized results. Invalid boundaries, unknown references, missing
properties, nonfinite values, duplicate IDs and excessive openings are rejected.

Psychrometric tests compare nine independently executed PsychroLib 2.5.0 cases:
five RH/saturation cases and four wet-bulb cases, including subfreezing conditions.
Tolerance is max(1e-12 absolute, 1e-10 relative). Those references independently
exercise a separate implementation, but use the same underlying correlations;
they are not experimental validation of the correlations themselves.

Seven further outdoor-air component and total metrics were independently generated
with PsychroLib dry-air density and moist-air specific-volume/enthalpy functions.
The corresponding comparison test uses max(1e-8 W absolute, 1e-10 relative).

The fixture matrix includes wall-conduction-only, simple-box, hot/cold boxes,
high-window-area, tight/leaky infiltration, high-latent climate, multi-room ranch,
two zones/stories, and an explicit attic boundary. `attic-duct-house` deliberately
does not model ducts: its name preserves the requested future validation category,
and its metadata states that duct loads are unsupported. Fixtures are small
synthetic physical abstractions, not validated full building design studies.

## Comparison harness

```sh
go run ./cmd/hvac-compare --relative 0.01 --absolute 0.5 \
  --report validation-report.json \
  testdata/buildings/wall-conduction-only.json \
  testdata/reference/wall-conduction-hand.json
```

The committed demonstration reference is a **hand calculation**, not an OpenStudio
result. The report provides reference/candidate/difference per selected total or
stable component ID, reference generator/version, fixture hash and tolerances.
Reference-zero metrics use absolute tolerance without a manufactured relative
percentage. Missing component mappings and fixture hash mismatches fail.

For OpenStudio-HPXML, save the exact matching HPXML, native project, weather/design
inputs, oracle version/commit and `results_design_load_details` source outputs
under `testdata/openstudio`. Map each supported reference component to a stable
engine ID in a version-1 reference JSON. Retain original output units and conversion
notes; fill `source_files`, `generator`, `generator_version`, `fixture_sha256` and
`metrics_w`. Do not create oracle results by copying this engine's output.

Comparison tolerances are explicit diagnostic choices, not engineering accuracy
claims. Investigate deviations before adopting fixture-specific thresholds.
Expected differences include solar timing/storage, opaque solar gains, airflow
allocation, peak coincidence, ground heat transfer, ducts, and sizing adjustments.
OpenStudio is a comparison target, not automatic normative truth.

## Desktop and HPXML gates

The UI service is tested for exact equality with the standalone library. Frontend
type checking and asset compilation are separate checks. Native GUI launch and
open/edit/save/inspect workflows still require platform acceptance testing.

The basic HPXML adapter imports a synthetic single-block walls/windows fixture,
with explicit supplemental design/airflow/internal/solar inputs. The fixture is
not asserted to be full XSD-valid HPXML. Representative real-world import, wider
envelope mapping, HPXML schema validation and room mapping remain outstanding.
