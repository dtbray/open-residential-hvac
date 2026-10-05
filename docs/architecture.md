# Architecture

The physical building model is independent of ACCA documents, HPXML XML types,
UI frameworks, and project storage. Typed SI quantities cross calculation package
boundaries. JSON/YAML field names declare their units, and `units` offers explicit
IP and SI conversions for callers.

| Boundary | Responsibility |
| --- | --- |
| `building`, `climate` | Zones, rooms, surfaces/openings, assemblies, airflows and design conditions |
| `units` | Distinct SI quantity, temperature-difference, air-property and azimuth types; explicit SI/IP conversions |
| `project` | Version 1 strict human-readable serialization |
| `provenance` | Source classifications and assumption metadata |
| `envelope` | Series resistance derivation and conductive heat transfer |
| `physics` | Signed conduction, solar, air-exchange, sensible/latent air and internal-gain primitives |
| `psychrometrics` | Saturation pressure, humidity ratios and moist-air properties |
| `infiltration`, `ventilation` | Explicit airflow modes and room allocation |
| `solar` | Incident glazing-plane solar gain |
| `loads` | Shared result/error types and inspectable tree construction; no model imports |
| `models` | Small `ID`/`Calculate` interface; no plugin registry |
| `models/designload` | Model validation, primitive selection, clipping/airflow policy, orchestration and aggregation choices |
| `hpxml` | Narrow mapping adapter with explicit supplemental inputs |
| `validation` | Independent reference comparison and deterministic reports |
| `app` | UI-neutral Go application service |
| `cmd/hvac`, `cmd/hvac-compare` | CLI boundaries |
| `cmd/desktop`, `app/frontend` | Wails lifecycle/files and Svelte editing/inspection |

`designload.New().Calculate(building, conditions)` validates first, resolves assemblies and airflow,
calculates individual room components, then aggregates rooms into zones and zones
into the whole building. A project can store design conditions on its building,
but explicit method arguments are authoritative for calculation and validation.
They do not mutate the stored project. Core climate values contain physical
temperatures, humidity and pressure, with no national-standard lookup identifiers.
There is one supported model and a small structural interface, with no discovery,
registry or plugin lifecycle. `en12831`, `iso52016`, `csaf280` and `acca` remain
conceptual until actual implementations justify their packages and inputs.

The calculation layers are building description → resolved physical properties →
shared thermal primitives → calculation-model policy → result tree. The primitive
layer preserves signed heat flow. `designload` decides which terms are included,
clips negative credits, selects seasonal boundaries and indoor air properties,
allocates airflows, and builds room/zone/block sums. Its steady-state assumptions
do not live in the building types or the shared result contract.

Heating and cooling have separate roots. Each cooling room has sensible and latent
children. IDs derive from immutable domain IDs, not labels or numeric values.
Leaf values use watts; inputs carry individual units and source/assumption metadata.
Parent values are sums of children and propagate warnings and assumptions.
The result's `methodology` contains ID, name, version and a methodology reference,
separately from each input's source. Repeated assumptions can appear in detailed trees because each dependency retains
its metadata; the CLI summary de-duplicates them by ID.

No calculation mutates the project. No calculation makes network calls, starts an
external process, or depends on Wails. Desktop-only imports are build-tagged so
ordinary library/CLI tests require no native GUI development libraries.

JSON/YAML v1 deliberately uses canonical SI fields rather than permitting ambiguous
bare values or mixed-unit keys. Missing required design quantities and gains are
pointer-backed to distinguish absence from explicit zero. Areas/volumes must be
positive; nonfinite inputs, unresolved references and duplicate IDs are rejected.

Future equipment, airflow targets, duct graphs and standards profiles should use
this physical model and stable results. They are outside v0.1. HPXML export can
share the adapter boundary without introducing XML types into calculations.

Generic `unconditioned_space` and `adjacent_building` boundaries require explicit
seasonal temperatures in `designload`. Airtightness measurements retain their own
test pressure and provenance independently of selected infiltration airflow.
Heat-recovery properties survive serialization but currently cause an explicit
unsupported-model error. Optional azimuth is degrees clockwise from true north;
the legacy orientation string is descriptive. Neither is used to infer solar
irradiance in this model.

Dynamic methods may need thermal mass, time series, setpoints and inter-zone
coupling. Those fields and hourly simulation are not introduced by this change.
A later dynamic provider may need a richer input/result API, alongside the small
current design-condition model interface. No steady-state policy is mandated by
the fundamental physical building representation.
