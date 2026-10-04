# Architecture

The physical building model is independent of ACCA documents, HPXML XML types,
UI frameworks, and project storage. Typed SI quantities cross calculation package
boundaries. JSON/YAML field names declare their units, and `units` offers explicit
IP and SI conversions for callers.

| Boundary | Responsibility |
| --- | --- |
| `building`, `climate` | Zones, rooms, surfaces/openings, assemblies, airflows and design conditions |
| `units` | Distinct quantity types and explicit conversion constants/functions |
| `project` | Version 1 strict human-readable serialization |
| `provenance` | Source classifications and assumption metadata |
| `envelope` | Series resistance derivation and conductive heat transfer |
| `psychrometrics` | Saturation pressure, humidity ratios and moist-air properties |
| `infiltration`, `ventilation` | Explicit airflow modes and room allocation |
| `solar` | Incident glazing-plane solar gain |
| `loads` | Domain validation, orchestration and hierarchical aggregation |
| `hpxml` | Narrow mapping adapter with explicit supplemental inputs |
| `validation` | Independent reference comparison and deterministic reports |
| `app` | UI-neutral Go application service |
| `cmd/hvac`, `cmd/hvac-compare` | CLI boundaries |
| `cmd/desktop`, `app/frontend` | Wails lifecycle/files and Svelte editing/inspection |

`loads.Calculate(building)` validates first, resolves assemblies and airflow,
calculates individual room components, then aggregates rooms into zones and zones
into the whole building. Building design conditions are part of the domain object.
There is one load model now; no speculative plugin/provider registry is introduced.
A second real implementation can motivate a small provider interface later.

Heating and cooling have separate roots. Each cooling room has sensible and latent
children. IDs derive from immutable domain IDs, not labels or numeric values.
Leaf values use watts; inputs carry individual units and source/assumption metadata.
Parent values are sums of children and propagate warnings and assumptions.
Repeated assumptions can appear in detailed trees because each dependency retains
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

