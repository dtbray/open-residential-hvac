# MVP specification and acceptance contract

This document records the implementation contract supplied by the project owner.
It is an organized summary of that specification, with repository/license choices
added on 2026-10-04. The project is a standalone AGPL-3.0-only repository;
Forgejo is authoritative and GitHub is a public one-way distribution mirror.

## Purpose and first principles

Build a small, transparent open-source residential HVAC engineering engine focused
on room and whole-building heating/cooling design loads. Every important result
must explain its inputs, equation, units, source, assumptions and aggregation.
Traceability belongs to the Go engine, not a desktop-only inspector.

All engineering/domain logic lives in Go. The same library powers CLI, Wails,
tests, and future HTTP callers. Use the standard library where practical, small
focused dependencies, deterministic results and no runtime EnergyPlus/OpenStudio
or network requirement. Ordinary single-family calculations should take well
under one second; correctness and explanation precede optimization.

Prefer physical domain concepts over standard-specific documents; measured values
over assumptions; derived values over unexplained constants; explicit missing
data over hidden defaults; and documented validation over plausible-looking loads.

## Domain and storage

Building: ID/name/location, zones/rooms, reusable assemblies, infiltration,
ventilation and explicit design conditions. Multiple zones must be structurally
supported even if one conditioned zone is common. Rooms need ID/name, floor area,
volume, occupants, walls/windows/doors/floors/ceilings and internal gains.

Surfaces express gross area, orientation, tilt, adjacent condition, assembly and
openings. Boundary conditions include outdoors, ground, conditioned, attic,
garage, crawlspace, basement and adjacent unit; surface type must not determine
the boundary by itself. Reusable assemblies support direct U or derived explicit
layer R values, with component provenance.

Use typed quantities and one canonical unit system; support boundary conversions
for °F/°C, ft²/m², Btu/h/W and CFM/L/s. Human-readable versioned JSON/YAML projects
must round-trip without data loss. SQLite is optional and not needed to save a home.
Design temperatures and indoor/outdoor cooling humidity must be explicit and
carry provenance; manual climate entry is sufficient initially.

## Calculations and result contract

Heating: opaque walls, ceilings/roofs, floors, windows/doors, infiltration sensible
and ventilation sensible. Unconditioned-space effects require defensible boundary
inputs. Ground/slab loss must either use a documented public method or be explicitly
unsupported. Unsupported categories must never silently receive invented values.

Cooling: opaque/opening conduction, glazing solar, infiltration/ventilation
sensible and latent, occupants, lighting and other internal sensible/latent gains.
A simplified solar method is permitted when documented, inspectable and tested.
Psychrometrics belong in a dedicated package with public implementations such as
PsychroLib used for numerical validation.

Infiltration must expose explicit modes (natural ACH, explicit airflow, and a
documented ACH50 conversion if offered), returning airflow derivation/provenance.
Ventilation supports explicit continuous airflow and stays separate from
infiltration. Full ventilation standards logic is outside the initial scope.

Result nodes need stable IDs, names, values, methods/equations, inputs, sources,
assumptions, warnings and children. Aggregation must be explicit. Cooling sensible
and latent must be distinguishable at room and block levels. Every material input
must be classifiable as measured, user-entered, inferred/derived, assumed/default,
imported, dataset or public-reference-based. Missing information and influential
assumptions must remain visible through dependent results; no statistical
uncertainty engine is required.

## Interfaces and interoperability

CLI minimum: validate, load, room selection, structured JSON output and explain
(optionally selecting a stable node ID). Errors/warnings should be structured;
invalid inputs must not silently become zero.

Wails desktop minimum: project/design metadata, room/zone list, structured envelope
and assembly editor, load totals/per-room sensible/latent results, tree inspector,
warnings/assumptions, and project open/save. TypeScript with Svelte or React is
permitted. Frontend work is input gathering, display and formatting only.
A graphical plan is optional; a complete structured editor is sufficient.

HPXML must remain an adapter to the physical native domain. Prefer basic supported
subset import; acceptable initial fallback is a mapping layer and fixtures.
Full export and wider import can follow without coupling calculations to XML types.

## Validation and source rights

Each calculation package needs meaningful numerical, dimensional, edge-case and
regression checks with explicit tolerances. Fixtures should isolate effects before
combining them: simple boxes, hot/cold climate, window-heavy, tight/leaky,
high-latent, multi-room ranch, two-story, attic/duct category.

OpenStudio-HPXML is a key independent validation target. Retain real reference
outputs, versioned matching fixtures, component-level diffs and documented
significant deviations. Agreement is not a substitute for method justification.
Do not fabricate oracle results or assume the oracle is normative truth.

Document all adopted formulas, coefficients, datasets, sources and licenses.
Do not copy proprietary ACCA prose/tables/figures or numeric standard tables whose
redistribution rights are unclear. A BSD/MIT repository license alone may not
establish independent rights in embedded third-party material.
No Manual J/S/D compliance, certification or marketing claims are permitted.

## Non-goals and later releases

v0.1 excludes equipment matching, manufacturer catalogs, ducts/airflow balancing,
register selection, annual dynamic simulation, CAD, photo geometry, accounts,
SaaS, collaboration and ACCA certification. Do not clone proprietary software.

v0.2: equipment/performance data, interpolation, heat-pump derating, sensible/latent
capacity checks, modulation and candidate ranking. v0.3: room airflow targets,
duct graphs, fittings/losses, pressure budgets, return paths and balancing.
v0.4: alternate load models, validation/standards profiles and investigation of
ACCA approval/licensing. Introduce interfaces for actual variation; do not build
a speculative plugin system before a second implementation exists.

## Milestones and completion gate

1. Foundation: Go module, CI, formatting/testing, schema version and unit primitives.
2. Domain: fully represent and losslessly round-trip a ranch-style home.
3. Heating: hand-checked component/room/block results and result tree.
4. Cooling: sensible/latent components with complete input provenance.
5. Validation: actual OpenStudio-HPXML references, comparison/report tooling and
   component-level investigation.
6. CLI: complete independent calculation workflow.
7. Desktop: create/edit/save projects and inspect the same Go results.
8. HPXML: representative supported fixture imports without manual restructuring.

v0.1 is complete only when a human-readable multi-room home calculates heating
and separate sensible/latent cooling; important results and assumptions are
traceable; Go library, CLI and Wails share the engine; representative real
OpenStudio comparisons and deviations are documented; no runtime simulation
dependency or proprietary unauthorized material is present; no ACCA compliance
claim is made; and deterministic CI runs pass.

