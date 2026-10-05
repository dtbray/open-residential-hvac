# Real design-load comparison evidence

Two residential apartment fixtures use pinned OpenStudio-HPXML 1.12.0 with
OpenStudio 3.11.0+241b8abb4d. `apartment` uses the upstream Denver apartment
sample unchanged. `apartment-hot` changes only its state/weather file to Phoenix.
Both ran with schema and schematron validation enabled and `--skip-simulation`.
These are design-load runs, not annual energy simulations or experimental validation.

Each fixture retains input HPXML, resolved HPXML, raw design-load output,
generation log/manifest and weather hash, explicit adapter options, native project,
reference metrics/mapping, candidate comparison and diagnostic regression policy.
`tools/oracle-lock.json` pins downloadable archives and their SHA256 hashes;
`node tools/regenerate-oracle.mjs` reproduces raw results into `dist/oracle`, without
replacing reviewed evidence. The archived workflow contains both weather files.
Input fixture redistribution retains `LICENSE.upstream.md`. Generated reference
outputs are kept for comparison; no proprietary coefficient/lookup tables or
OpenStudio calculation implementation are copied into the engine.

Mappings convert reported Btu/h into watts and identify exact report/entry/field.
Heating window metrics map to conduction. Cooling window metrics include solar
timing and cannot map to a conduction-only node, so they are represented in total
comparisons, not mislabeled as window-conduction equivalence. AED adjustments,
zero duct/piping/blower categories, and whole-building totals remain in the raw
report even where the native model has no corresponding component.

The native adapter imports the resolved file without manual geometry restructuring:

```sh
go run ./cmd/hvac-import-hpxml --options testdata/openstudio/apartment/import-options.json \
  --output project.json testdata/openstudio/apartment/resolved.xml
hvac load project.json
```

Inputs deliberately make methodology differences explicit: heating infiltration
CFM is used for both seasons, cooling RH/pressure are benchmark choices, adjacent
units are held at explicit indoor setpoints, internal gains are normalized to the
oracle's reported gains, and simultaneous plane irradiances are explicit choices.
Those choices are not national design defaults or back-calculated solar factors.
See `docs/validation.md` for quantitative differences and outstanding coverage.
