# Validation status

v0.1 is not declared complete. Two real, schema/schematron-validated apartment
design-load comparisons are committed under `testdata/openstudio`. They show
substantial cooling differences. There is no evidence yet that this simplified
model meets any industry sizing accuracy requirement.

Fixtures and model-specific references now live under `testdata/designload`.
Psychrometric primitive references live under `testdata/physics`. Each comparison
report identifies the candidate methodology independently of the reference
generator. Future methodology suites should use their own fixtures, assumptions
and tolerances; differing models are not required to produce identical loads.

The international architecture refactor preserves all existing numerical results,
stable node IDs, equations, inputs and source/assumption metadata across 11 fixtures.
`pre-refactor-regression.json` stores 411 node values captured from commit
`582d69f419c98a60fb814693bce91cd1a2148584`. This is a refactor regression baseline,
not an independent validation oracle or an OpenStudio result.

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
  testdata/designload/buildings/wall-conduction-only.json \
  testdata/designload/reference/wall-conduction-hand.json
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
type checking and asset compilation are separate checks. Native Linux launch/edit/calculate/inspect/save/reopen workflows are exercised through
AT-SPI using the real Wails bridge and compared to CLI totals. Windows and macOS
still require native GUI acceptance. See [CI gates](ci.md).

The HPXML adapter imports the pinned schema-valid resolved apartment fixtures,
including walls/windows/doors, floors/ceilings and explicit shared-unit boundary
temperatures. Supplemental climate/airflow/internal/solar inputs remain explicit.
The adapter itself does not validate the full HPXML schema; the oracle workflow
validates the fixtures. Roofs, ground/foundation categories, full-house systems and
room mapping remain unsupported. The older synthetic fixture remains a unit test.

## Investigated apartment comparisons

| Fixture | Heating engine / oracle W | Difference | Cooling engine / oracle W | Difference |
|---|---:|---:|---:|---:|
| Denver apartment | 1704.14 / 1724.43 | -1.18% | 3461.37 / 2346.91 | +47.49% |
| Phoenix apartment | 766.35 / 766.97 | -0.08% | 4020.93 / 2793.26 | +43.95% |

The full `comparison.json` files retain sensible/latent and individual mapped
components; `metric-mapping.json` identifies each raw oracle field and unit conversion.
Both fixtures derive from the same residential apartment geometry and therefore do
not establish broad building coverage. The 1% / 0.5 W diagnostic threshold is not
an accepted sizing accuracy requirement. CI intentionally permits documented
cross-method failures while rejecting changed mappings, references or candidate
regressions. The `pass` flags remain false where differences exceed that threshold.

Heating differences predominantly come from air-property/infiltration treatment:
the model uses physical dry-air density and heat capacity at explicit pressure,
while oracle design airflow and sensible-air treatment are different. Surface
conduction differences include output rounding and method-specific U adjustments.
The native input uses the reported heating infiltration CFM for both seasons;
OpenStudio reports a lower cooling infiltration flow. This is a known policy
mismatch, not an inferred natural-ACH conversion.

Cooling solar uses simultaneous explicit benchmark plane irradiances of 100/400/500
W/m2 for north/south/west with shading factor 1. The oracle uses its own solar,
shading, peak/timing and AED treatment; its raw window cooling includes both solar
and conduction, so those values are not mapped to conduction-only result nodes.
The engine does not implement AED excursions or cooling storage. Outdoor RH and
pressure are explicit benchmark choices (Denver 15%/83000 Pa; Phoenix 25%/97000 Pa),
not claimed equivalents of the oracle's humidity-difference inputs. Negative oracle
latent infiltration remains in raw references, whereas `designload` clips negative
cooling contributions to zero. Internal sensible/latent gains are normalized to
reported oracle gains to isolate envelope/air/solar differences; this is not an
independent occupant-gain calibration. No coefficients are fitted to totals.

Required next coverage includes isolated infiltration/solar comparisons with matched
physical conditions, a complete single-family fixture, ground/attic behavior where
supported, and more room/zone configurations. Tighten engineering acceptance only
after investigating component-level differences. Reproduction instructions, archive
hashes, source rights and limitations are in `testdata/openstudio/README.md`.
