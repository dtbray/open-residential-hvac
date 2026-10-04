# Open Residential HVAC Engine

Forgejo is authoritative: https://git.thomas-bray.com/thomas/open-residential-hvac
GitHub is a public distribution mirror. Push changes to Forgejo only. Open issues
and pull requests on Forgejo. Do not edit the GitHub mirror independently.

License: AGPL-3.0-only. Preserve compatible third-party notices separately.

Read docs/mvp-specification.md and docs/international-architecture.md before
implementation. Keep physical domain/properties,
shared physics, model-specific policy and result contracts separate. `designload`
is the only implemented model; do not add placeholder national-standard packages
or a plugin framework. Include methodology identity/version in structured results.
All engineering calculations belong in Go. CLI and desktop must call the same engine. No ACCA compliance claims,
proprietary tables, hidden engineering defaults, or invented validation outputs.

Use canonical SI quantities internally and convert at serialization boundaries.
Missing engineering inputs must produce errors; explicit zero is different from
missing. Preserve provenance and assumptions through aggregation. Numerical tests
must use explicit expected values and tolerances.

Before publishing: gofmt, go vet ./..., go test -race ./..., and go build ./cmd/hvac.
Document unsupported categories and validation gaps honestly. Do not claim that
OpenStudio comparisons ran unless real versioned oracle output is present.
