# Open Residential HVAC Engine

A standards-neutral residential thermal-load engine, implemented as a Go library
with a CLI and Wails/Svelte desktop editor. Its initial `designload` model provides
transparent room and whole-building heating/cooling design loads. Each result has stable
IDs, equations, inputs, units, sources, assumptions, and explicit aggregation.

**Development preview. The complete v0.1 acceptance gate has not been met.**
OpenStudio-HPXML design-load comparisons and representative real HPXML import
remain outstanding. The engine uses a deliberately simplified steady-state model;
it does not claim Manual J, S, D, or ACCA compliance.

## Source and license

[Forgejo is authoritative](https://git.thomas-bray.com/thomas/open-residential-hvac).
[GitHub is a public distribution mirror](https://github.com/dtbray/open-residential-hvac).
Submit changes, issues, and reviews on Forgejo. GitHub receives one-way pushes;
edits there can be overwritten. Mirroring copies Git branches and tags, not issues,
pull requests, release attachments, or packages.

Original code is **AGPL-3.0-only**; see [LICENSE](LICENSE). Third-party components
retain their own notices in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md). No
proprietary ACCA tables are included.

## CLI

Go 1.24 or newer:

```sh
go test ./...
go build -o bin/hvac ./cmd/hvac
bin/hvac validate testdata/designload/buildings/ranch.yaml
bin/hvac load testdata/designload/buildings/ranch.yaml
bin/hvac load testdata/designload/buildings/ranch.yaml --model designload --units ip
bin/hvac load testdata/designload/buildings/ranch.yaml --room "Bedroom"
bin/hvac load testdata/designload/buildings/ranch.yaml --format json > result.json
bin/hvac explain testdata/designload/buildings/ranch.yaml --node room/bedroom/heating/envelope/bedroom-west
```

Text heat-flow output defaults to watts; use `--units ip` for Btu/h. JSON always
uses canonical SI and includes methodology metadata, room results and complete
trees. `--model designload` explicitly selects the only supported model.
`--verbose` prints equations and input provenance. An isolated JSON `--node`
response contains `methodology` and `node` fields together.
Room selection calculates the whole building first, retaining its airflow
allocation, then selects that room's results. Validation errors exit nonzero;
`--format json` produces structured errors on stderr.

## Project file

Projects use strict versioned JSON or YAML, with explicit SI field names. Unknown
fields and unsupported schema versions are errors. Required engineering inputs
must be present; zero is a permitted explicit input. A complete, synthetic
two-room example is [ranch.yaml](testdata/designload/buildings/ranch.yaml).

```yaml
version: 1
building:
  id: house
  name: Example Ranch
  # Additional required fields are shown in the complete example.
  design:
    heating:
      outdoor_db_c: -15
      indoor_db_c: 20
    cooling:
      outdoor_db_c: 35
      outdoor_rh_fraction: 0.5
      indoor_db_c: 24
      indoor_rh_fraction: 0.5
    pressure_pa: 101325
    evidence: {}
```

This excerpt is illustrative, not a complete valid project. The library provides
explicit °F/°C, ft²/m², Btu/h/W and CFM/L/s conversions; project schema v1 and
the desktop editor use SI. Reference fixtures are synthetic inputs, not suggested
insulation values, infiltration rates, or climate design conditions.

## Go library

```go
p, err := project.Open("house.yaml")
if err != nil { return err }
model := designload.New()
result, err := model.Calculate(p.Building, p.Building.Design)
```

Explicit conditions can instead come from another physical climate source; the
model uses those conditions without mutating the building's stored project design.
Calculations need no network or external processes. Shared physics and result types
are independent of model policy. See [architecture](docs/architecture.md) and the
[international architecture addendum](docs/international-architecture.md).

## Desktop

Use Node.js 22+, Go, and the Wails v2 platform prerequisites. On Debian/Ubuntu,
native builds require a C compiler, pkg-config, GTK3 and WebKitGTK 4.1 development
packages. On Windows, WebView2 is required. On macOS, install Xcode command-line tools.

```sh
cd app/frontend
npm ci
npm run check
npm run build
cd ../..
go build -tags 'desktop,webkit2_41' -o bin/open-residential-hvac ./cmd/desktop
bin/open-residential-hvac
```

For Windows/macOS omit `webkit2_41`. The frontend must be built before the Go
desktop build, since its assets are embedded. `make desktop` performs the Linux
build sequence. Wails development configuration is in `cmd/desktop/wails.json`;
invoke Wails with the `desktop` build tag.

The editor opens/saves JSON and YAML, edits project/design fields, zones, rooms,
surfaces, windows, doors, assemblies and airflows, and expands the Go result trees.
Optional absent fields and alternate infiltration/assembly modes can be edited in
the complete JSON screen. Native project save accepts incomplete engineering
inputs so projects can be saved mid-edit; Calculate enforces domain validation.

## Methods and limitations

Implemented: surface and opening conduction, natural-ACH/explicit/ACH50-divisor
infiltration, explicit ventilation, sensible/latent air loads, glazing gains,
occupants, lighting and explicit internal sensible/latent gains. Explicit boundary
temperatures support attics, garages, crawlspaces, basements and adjacent units.

Unsupported: ground/slab heat transfer, opaque solar absorption, thermal storage,
design wind/stack infiltration models, heat recovery, ducts, equipment selection,
automatic climate resolution, and standards compliance. Ground boundaries fail
validation. Solar irradiance must be entered on the glazing plane for one common
design snapshot. The simplified solar calculation uses immediate heat gain as
sensible load and does not reproduce time-dependent cooling sizing methods.

Read [methodology](docs/methodology.md), [provenance](docs/provenance.md),
[validation status](docs/validation.md), and [the MVP specification](docs/mvp-specification.md)
before interpreting these development results.
