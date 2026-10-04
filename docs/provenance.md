# Provenance and redistribution policy

Original code: AGPL-3.0-only. No ACCA prose, tables, figures, coefficients or diagrams
have been copied. No ACCA compliance or certification is claimed.

| Material | Source and use | License/status |
| --- | --- | --- |
| Psychrometric correlations and SI air properties | [PsychroLib 2.5.0](https://github.com/psychrometrics/psychrolib/tree/fca789d70d61aa4a4b5b6594d3e7299a7f4c86db), adapted equations and independently executed JS validation cases | MIT; full notice in THIRD_PARTY_NOTICES.md |
| Conduction and series resistance | Algebraic definitions of U/R and steady-state heat transfer; original implementation | No external prose/tables copied |
| Glazing gain | [US DOE definition of SHGC](https://www.energy.gov/energysaver/energy-performance-ratings-windows-doors-and-skylights); original product-of-explicit-inputs implementation | Definition used; no optical tables copied |
| SI/IP conversions | Exact international foot and IT Btu unit definitions, documented in units.go | Numerical unit definitions; no dataset |
| OpenStudio-HPXML | [Design-load detail output documentation](https://openstudio-hpxml.readthedocs.io/en/latest/workflow_outputs.html); intended independent comparison | No implementation or numeric tables imported; actual comparisons pending |
| HPXML | Explicit field mapping for a narrow synthetic fixture | No HPXML schema or third-party fixture redistributed |
| Example house/fixture values | Original synthetic test inputs | AGPL-3.0-only project material; not real measurements |
| Go YAML, Wails, frontend dependencies | Pinned via go.mod/go.sum and package-lock.json | Preserve each dependency's upstream terms |

The PsychroLib JS oracle file was retrieved from the pinned commit; its SHA256 is
`c1622038dfd28fbbf5ad1a197fc30eac43f6bf94f4f85e946913a0ad239f7b59`.
The output fixture records that hash and the source commit. Reference code is a
development oracle, not a runtime dependency.

For future external sources, record version, URL/reference, license, retrieval
method, input fixture hashes and modifications. Review rights to underlying
coefficient tables separately from a repository's general software license.
Permissively licensed surrounding code is not proof that embedded third-party
standard tables can be redistributed. Missing rights mean do not import them.

Property `evidence` maps use the exact schema field name as the key; design keys
include `heating.` or `cooling.` prefixes. Example:

```yaml
evidence:
  area_m2:
    source:
      kind: measurement
      reference: Site survey 2026-10-04
    assumptions:
      - id: approximate-wall-outline
        description: Measured as a rectangular wall; small recesses omitted
```

