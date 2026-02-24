# slnxsync

[![CI](https://github.com/n2jsoft-public-org/slnxsync/actions/workflows/ci.yml/badge.svg)](https://github.com/n2jsoft-public-org/slnxsync/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/n2jsoft-public-org/slnxsync)](https://goreportcard.com/report/github.com/n2jsoft-public-org/slnxsync)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A Go CLI tool that reads a source `.slnx` solution file and a filter configuration, then generates targeted `.slnx` files for CI-oriented workflows.

## Features

- **Filter solution files** by project path and name using wildcard patterns
- **Generate multiple profiles** from a single configuration file
- **Deterministic output** with stable project ordering
- **CI-friendly** with dry-run mode and exit code conventions
- **Helpful error messages** with profile name suggestions

## Installation

```bash
go install github.com/n2jsoft-public-org/slnxsync/cmd/slnxsync@latest
```

Or build from source:

```bash
git clone https://github.com/n2jsoft-public-org/slnxsync.git
cd slnxsync
go build -o slnxsync ./cmd/slnxsync
```

## Quick Start

Create a filter configuration file (`filters.yml`):

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
    output: ./out/MyProduct.Api.CI.slnx

  tests-only:
    include:
      - "**/*.Tests"
    output: ./out/MyProduct.Tests.slnx
```

Generate filtered solution files:

```bash
slnxsync generate -c filters.yml
```

Preview what would be generated:

```bash
slnxsync preview -c filters.yml --profile api-ci
```

Validate configuration:

```bash
slnxsync validate -c filters.yml
```

Check CLI version metadata:

```bash
slnxsync version
slnxsync --version
```

## Usage

### Generate Command

Generate filtered solution files from a configuration:

```bash
# Generate all profiles
slnxsync generate -c filters.yml

# Generate specific profile only
slnxsync generate -c filters.yml --profile api-ci

# Dry run (no files written)
slnxsync generate -c filters.yml --dry-run

# Override output directory
slnxsync generate -c filters.yml --out-dir ./build

# Strict mode (fail on unmatched include patterns)
slnxsync generate -c filters.yml --strict

# Verbose output
slnxsync generate -c filters.yml -v
```

### Preview Command

Preview which projects would be selected for a profile:

```bash
# Preview specific profile
slnxsync preview -c filters.yml --profile api-ci

# Strict mode validation
slnxsync preview -c filters.yml --profile api-ci --strict
```

### Validate Command

Validate configuration file syntax and rules:

```bash
slnxsync validate -c filters.yml
```

### Version Command

Print CLI version/build metadata:

```bash
slnxsync version
slnxsync --version
```

Output format:

```text
version: 1.2.3
commit: abc1234
buildDate: 2026-02-24T12:00:00Z
```

Build-time metadata can be injected with linker flags:

```bash
go build -ldflags "-X main.Version=1.2.3 -X main.Commit=$(git rev-parse --short HEAD) -X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o slnxsync .
```

Docker image builds support the same values via build args:

```bash
docker build -f cmd/slnxsync/Dockerfile \
  --build-arg VERSION=1.2.3 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t slnxsync:1.2.3 .
```

## Configuration

### Filter File Format

```yaml
version: 1                  # Config schema version (required)
source: ./path/to/source.slnx  # Source solution file (required)
profiles:                   # Profile definitions (required)
  profile-name:             # Profile name (used with --profile flag)
    include:                # Include patterns (optional, default: all entries)
      - "pattern1"
      - "pattern2"
    exclude:                # Exclude patterns (optional)
      - "pattern3"
    output: ./out/filtered.slnx  # Output path (required)
```

### Pattern Matching

Patterns support glob wildcards:
- `*` - matches any sequence except path separator
- `**` - matches any sequence including path separators
- `?` - matches single character
- `[abc]` - matches character class

Patterns are matched against:
- **Project path** - relative path from solution root
- **Project name** - project file name
- **File path** - file path from solution root
- **File name** - file name
- **Folder path** - folder name/path (without leading/trailing `/`)
- **Folder name** - terminal folder segment

Examples:
- `src/Api/**` - all projects under src/Api
- `**/*.Tests` - all test projects
- `src/*/Core` - Core projects in any src subdirectory
- `MyProject.*` - projects matching name pattern

### Evaluation Order

1. **Include phase**: Select entries matching any include pattern (default: all if omitted)
2. **Exclude phase**: Remove entries matching any exclude pattern
3. **Prune phase**: Remove folders left empty after filtering

## Exit Codes

- `0` - Success
- `1` - Runtime error (file not found, parse error, etc.)
- `2` - Validation error (invalid config, bad pattern, etc.)

## Examples

### CI Pipeline Integration

**Azure Pipelines:**

```yaml
- task: GoTool@0
  inputs:
    version: '1.23'

- script: |
    go install github.com/n2jsoft-public-org/slnxsync/cmd/slnxsync@latest
    slnxsync generate -c filters.yml --profile api-ci
  displayName: 'Generate API solution'

- task: VSBuild@1
  inputs:
    solution: 'out/MyProduct.Api.CI.slnx'
```

**GitHub Actions:**

```yaml
- name: Setup Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.23'

- name: Generate filtered solution
  run: |
    go install github.com/n2jsoft-public-org/slnxsync/cmd/slnxsync@latest
    slnxsync generate -c filters.yml --profile api-ci

- name: Build solution
  run: dotnet build out/MyProduct.Api.CI.slnx
```

### Multiple Profile Generation

Generate different solution variants for different pipeline stages:

```yaml
version: 1
source: ./MyProduct.slnx
profiles:
  fast-ci:
    include:
      - "src/Core/**"
      - "src/Api/**"
    exclude:
      - "**/*.IntegrationTests"
    output: ./ci/fast.slnx

  full-ci:
    include:
      - "src/**"
    output: ./ci/full.slnx

  nightly:
    include:
      - "src/**"
      - "benchmarks/**"
      - "tools/**"
    output: ./ci/nightly.slnx
```

## Development

### Prerequisites

- Go 1.22 or later
- golangci-lint (for linting)

### Build

```bash
go build -v ./...
```

### Test

```bash
go test -v ./...
```

### Lint

```bash
golangci-lint run ./...
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run `golangci-lint run ./...` and `go test ./...`
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [doublestar](https://github.com/bmatcuk/doublestar) for glob matching
