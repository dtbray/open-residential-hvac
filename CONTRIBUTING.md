# Contributing

Open issues and pull requests on the authoritative
[Forgejo repository](https://git.thomas-bray.com/thomas/open-residential-hvac).
The GitHub repository is a public Git mirror. New work should use a feature branch
and a ready-for-review pull request on Forgejo.

Contributions are provided under AGPL-3.0-only. Preserve external licenses and
document the source of every adopted formula, coefficient set or dataset. Verify
underlying material rights separately from the license of surrounding software.

Run `make check`, `make build`, and the frontend check/build when changing UI code.
Engineering changes need numerical expected values, explicit tolerances, invalid
input cases and provenance checks. Document method limitations and differences
from reference implementations. A copied engine output is a regression baseline,
not an independent validation oracle.

No frontend HVAC calculations; no silent engineering defaults; no proprietary
ACCA content; no standards-compliance marketing without the required approval.

