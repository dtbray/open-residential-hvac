# Continuous integration and release gates

Forgejo is authoritative. GitHub mirrors code and tags, not workflow results or
release artifacts. Workflows live in `.forgejo/workflows`; they do not require a
GitHub credential. Action revisions, oracle archives and scanner versions are pinned.

## Pull requests and pushes

`CI / verify` checks formatting, vet, race tests, numerical/serialization regressions,
CLI process workflows, frontend types and assets, stored reference comparisons,
and a **runnable production** Linux desktop. Native acceptance uses a virtual X
display and AT-SPI to create/edit/calculate, expand the inspector, save through the
native dialog, compare rendered totals to the CLI, reopen and recalculate. It uses
the real Wails bridge and Go engine, with no frontend HVAC formulas or mocked bridge.
Windows production cross-compilation remains an additional compile check.

Reports and application/accessibility logs are uploaded as workflow artifacts.
Stored OpenStudio comparisons are diagnostic: CI checks input/reference integrity,
component mapping, documented differences and candidate regressions. It does not
turn large cross-method differences into passing accuracy claims. Review changes
to reference snapshots as engineering changes, with the generation evidence.

Linux native prerequisites are installed by `tools/install-linux-ci.sh` when
missing. The installer locks concurrent installation and supports root or passwordless
sudo. The existing `security` label is a persistent instance-wide Linux host runner.
Public untrusted PRs must be approved before executing on a self-hosted runner;
native acceptance needs no repository secrets or privileged service access.
WebKit sandbox disabling is limited to nested headless CI tests. Packaged apps do
not set that environment variable.

## Scheduled and manual checks

`Scheduled validation` runs weekly on Monday at 05:17 UTC and supports manual
dispatch. One job checksum-verifies pinned OpenStudio 3.11.0 and HPXML 1.12.0 archives,
validates input/weather hashes, executes schema/schematron-validated design-only
runs, and compares the regenerated raw oracle reports exactly with stored reports.
It never overwrites committed evidence. A following job (also run if the oracle fails) runs pinned `govulncheck` against
the desktop build tags and `npm audit --audit-level=high`. Findings or infrastructure
errors fail the job; both retain reports. Scanner/database access is confined to
CI, not runtime calculations. Reachable Go findings are blocking; the scanner can
also report unreachable module/package advisories for review.

Local equivalents:

```sh
make check build
node tools/cli-smoke.mjs
node tools/check-references.mjs
node tools/regenerate-oracle.mjs
go install golang.org/x/vuln/cmd/govulncheck@v1.8.0
govulncheck -tags desktop,production,webkit2_41 ./...
cd app/frontend && npm ci && npm audit --audit-level=high
```

Oracle regeneration is currently Linux x86-64 only. Go 1.26 is the CI toolchain;
the module minimum is 1.25 after the fixed parser dependency update.

## Native release packaging

`Native release packages` runs for `v*` tags or manual dispatch. Manual `target`
accepts `linux`, `windows`, `macos`, or `all`. `tools/package-release.mjs` builds the
three CLIs and the production desktop on the host platform, executes CLI acceptance,
and packages binaries, an example project, source commit/toolchain metadata,
AGPL license, third-party notices, actual Go dependency license texts and the
frontend runtime license. Archives have SHA256 sidecars. Linux also exercises the
packaged desktop using the native acceptance script.

**Available now:** Linux x86-64. Windows and macOS require native Forgejo runners;
none were online/configured when this workflow was added. Set repository variables
`HVAC_WINDOWS_RUNNER` and `HVAC_MACOS_RUNNER` to their actual labels after provisioning.
Windows runners need Git/tar, Node, Go and WebView2; macOS runners need Node, Go,
Git/tar and Xcode command-line tools. The workflows select the runner's native CPU
architecture. Windows/macOS jobs use native builds and CLI tests but still require
native GUI acceptance before claiming desktop support. No cross-build is labeled
as a completed native platform test.

Unconfigured conditional platform jobs use the Linux label only to evaluate their
skip condition on the current Forgejo runner; they never build a Windows/macOS
package on Linux. This avoids indefinitely queued jobs on nonexistent labels.

An all-platform/tag release **fails preflight** if either runner variable is
missing. Missing platforms are never silently advertised as released. A manual
Linux-only build remains available. Packages are unsigned; macOS signing/notarization
and Windows signing require future release credentials/policy. Workflow artifacts
are build evidence, not an automatically published stable release. Publishing a
version and copying release attachments to GitHub are separate steps; Git push
mirroring cannot perform that distribution.

Do not add a release tag merely to test the workflow. Dispatch a development build
and inspect packages first. A dependency upgrade, oracle refresh, or platform
failure requires a reviewed change rather than an automatic baseline rewrite.
