# AGENTS.md

## Mission
Build a Go CLI that reads a source `.slnx` file and a filter config, then generates one or more smaller `.slnx` files for CI-oriented workflows.

The tool must:
- Accept include/exclude rules with wildcard support.
- Produce deterministic output `.slnx` files.
- Make it easy to generate targeted solutions (per area, layer, or pipeline stage).

---

## Product Scope

### Inputs
- A source solution file: `*.slnx`
- A filter configuration file (YAML initially)

### Outputs
- One or many generated `*.slnx` files, each corresponding to a named profile in the filter config.

### Core Use Case
Given a large monolithic solution, teams define profiles (for example `api-ci`, `domain-ci`, `smoke`) and generate slimmed `.slnx` variants to speed up CI jobs.

---

## Non-Goals (MVP)
- Editing project files (`*.csproj`) or dependency graph rewriting.
- Intelligent project graph solving beyond explicit include/exclude matching.
- IDE integration.

---

## Proposed Filter File (v1)

Use YAML for readability and CI friendliness.

```yaml
version: 1
source: ./MyProduct.slnx
profiles:
  api-ci:
    include:
      - "src/Api/**"
      - "src/Common/**"
    exclude:
      - "**/*.Tests"
      - "src/Legacy/**"
    output: ./out/MyProduct.Api.CI.slnx

  tests-only:
    include:
      - "**/*.Tests"
    output: ./out/MyProduct.Tests.slnx
```

### Rule Semantics
- `include`: whitelist candidates (if omitted, default is all entries from source `.slnx`).
- `exclude`: remove matched entries after include phase.
- Matching targets project path/name, file path/name, and folder path/name.
- Wildcards follow glob semantics (`*`, `**`, `?`).
- Evaluation order: `include` then `exclude`, then prune empty folders.

---

## CLI Design (Cobra)

### Command Shape
- `slnxsync generate -c filters.yml`
- `slnxsync validate -c filters.yml`
- `slnxsync preview -c filters.yml --profile api-ci`

### Flags (initial)
- `-c, --config`: filter config path (required)
- `-p, --profile`: run one profile only (optional)
- `-o, --out-dir`: override output directory (optional)
- `--dry-run`: no files written; print actions
- `--strict`: fail on unmatched include patterns
- `-v, --verbose`: verbose logs

### Exit Codes
- `0` success
- `1` runtime/config/parsing error
- `2` validation failure (invalid profile/pattern/output conflict)

---

## Architecture

### Package Layout (proposed)
- `cmd/slnxsync/` → cobra commands and flags
- `internal/config/` → config schema + parsing + validation
- `internal/slnx/` → `.slnx` parser/writer
- `internal/filter/` → wildcard matching and include/exclude engine
- `internal/generate/` → orchestration for profile generation
- `internal/logging/` → output formatting / verbosity

### Data Flow
1. Load and validate config.
2. Parse source `.slnx` into model (projects + metadata).
3. For each selected profile:
   - Resolve include set.
   - Apply excludes.
   - Build output model.
   - Write output `.slnx`.
4. Print summary (selected/total projects, output path).

### Determinism Rules
- Stable ordering of projects in generated files.
- Normalized path separators.
- Repeatable output for same inputs.

---

## Roadmap

### Phase 0 — Bootstrap
- Initialize Cobra app skeleton with `root` command.
- Add commands: `generate`, `validate`, `preview`.
- Add base logging and shared flag handling.

### Phase 1 — Config Model + Validation
- Define YAML schema (`version`, `source`, `profiles`, `include`, `exclude`, `output`).
- Implement validation:
  - required fields
  - unique profile names
  - output path conflicts
  - pattern syntax checks
- Add `validate` command output with actionable errors.

### Phase 2 — `.slnx` Parsing/Writing
- Implement parser to read source `.slnx` project entries.
- Implement writer preserving required format conventions.
- Add round-trip tests (`parse -> write -> parse`).

### Phase 3 — Filter Engine
- Implement glob matcher for project path + name.
- Apply include/exclude pipeline with deterministic behavior.
- Add strict-mode handling for unmatched include rules.

### Phase 4 — Generation Command
- Implement `generate` orchestration for all or selected profiles.
- Add dry-run and summary reporting.
- Ensure output directory creation and safe overwrite behavior.

### Phase 5 — Preview & DX
- Implement `preview` command with table/text output of selected projects.
- Improve error messages (suggest nearest profile names).
- Add examples in help text and README.

### Phase 6 — Quality Gates
- Unit tests for config, filter engine, and parser.
- Golden file tests for generated `.slnx` output.
- CI pipeline:
  - `go test ./...`
  - `go vet ./...`
  - optional lint (`golangci-lint`).

### Phase 7 — v1 Release
- Versioning strategy (SemVer tags).
- Changelog and release notes.
- Publish usage examples for CI systems (GitHub Actions, Azure Pipelines).

---

## Testing Strategy
- Unit tests:
  - Pattern matching behavior (`*`, `**`, `?`)
  - Include/exclude precedence
  - Config validation errors
- Integration tests:
  - Generate multiple profiles from sample `.slnx`
  - Dry-run output snapshots
- Golden tests:
  - Compare generated `.slnx` against expected files

---

## Risks & Mitigations
- `.slnx` format complexity or undocumented edge cases
  - Mitigation: start with supported subset + fixtures from real solutions.
- Ambiguous wildcard behavior
  - Mitigation: document exact matching rules and add explicit tests.
- CI instability due to path differences across OS
  - Mitigation: normalize separators and use path-cleaning consistently.

---

## Definition of Done (v1)
- `generate`, `validate`, and `preview` commands implemented.
- Include/exclude wildcard rules behave as documented.
- Multiple profile outputs generated from a single config.
- Deterministic `.slnx` output validated by tests.
- CI checks green and release tagged.

---

## Suggested Next Task
Implement Phase 0 + Phase 1 together:
1. Scaffold Cobra commands.
2. Add config schema and validation.
3. Ship `validate` command first to lock contract before parser work.
