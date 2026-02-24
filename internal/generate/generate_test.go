package generate

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/n2jsoft-public-org/slnxsync/internal/slnx"
)

func TestRunDryRunAllProfiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := writeFixtureFiles(t, tempDir, `version: 1
source: ./sample.slnx
profiles:
  api:
    include:
      - src/api/**
    output: ./out/api.slnx
  tests:
    include:
      - test/**
    output: ./out/tests.slnx
`)

	results, err := Run(Request{ConfigPath: configPath, DryRun: true})
	if err != nil {
		t.Fatalf("run generate: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if _, err := os.Stat(filepath.Join(tempDir, "out", "api.slnx")); !os.IsNotExist(err) {
		t.Fatalf("expected no output files in dry-run mode")
	}
}

func TestRunSingleProfileWritesFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := writeFixtureFiles(t, tempDir, `version: 1
source: ./sample.slnx
profiles:
  api:
    include:
      - src/api/**
    output: ./out/api.slnx
  tests:
    include:
      - test/**
    output: ./out/tests.slnx
`)

	results, err := Run(Request{ConfigPath: configPath, Profile: "api"})
	if err != nil {
		t.Fatalf("run generate: %v", err)
	}
	if len(results) != 1 || results[0].Name != "api" {
		t.Fatalf("expected only api profile result, got %+v", results)
	}

	outputPath := filepath.Join(tempDir, "out", "api.slnx")
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected output file %q: %v", outputPath, err)
	}
}

func TestRunStrictUnmatchedIncludeReturnsValidation(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := writeFixtureFiles(t, tempDir, `version: 1
source: ./sample.slnx
profiles:
  bad:
    include:
      - missing/**
    output: ./out/bad.slnx
`)

	_, err := Run(Request{ConfigPath: configPath, Strict: true})
	if err == nil {
		t.Fatal("expected strict validation error")
	}

	var typedValidationErr *ValidationError
	if !errors.As(err, &typedValidationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if !strings.Contains(err.Error(), "profile \"bad\"") || !strings.Contains(err.Error(), "unmatched include patterns") {
		t.Fatalf("unexpected strict error message: %v", err)
	}
}

func TestRunOutDirOverrideConflict(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := writeFixtureFiles(t, tempDir, `version: 1
source: ./sample.slnx
profiles:
  one:
    include:
      - src/**
    output: ./out/shared.slnx
  two:
    include:
      - test/**
    output: ./another/shared.slnx
`)

	_, err := Run(Request{ConfigPath: configPath, OutDir: filepath.Join(tempDir, "override")})
	if err == nil {
		t.Fatal("expected output path conflict validation error")
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if !strings.Contains(err.Error(), "output path conflict") {
		t.Fatalf("expected output path conflict message, got: %v", err)
	}
}

func TestRunUnknownProfileSuggestsClosest(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := writeFixtureFiles(t, tempDir, `version: 1
source: ./sample.slnx
profiles:
  api-ci:
    include:
      - src/api/**
    output: ./out/api.slnx
  smoke:
    include:
      - src/**
    output: ./out/smoke.slnx
`)

	_, err := Run(Request{ConfigPath: configPath, Profile: "api-c"})
	if err == nil {
		t.Fatal("expected profile-not-found validation error")
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if !strings.Contains(err.Error(), "did you mean") || !strings.Contains(err.Error(), "api-ci") {
		t.Fatalf("expected suggestion in error, got: %v", err)
	}
}

func TestRunFiltersFilesAndFoldersByInclude(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	slnxContent := `<Solution>
  <Folder Name="/Build/">
    <File Path="Directory.Build.props" />
  </Folder>
  <Folder Name="/src/">
    <File Path="src/Directory.Build.props" />
    <Project Path="src/api/Api.csproj" />
  </Folder>
  <Folder Name="/Mobile/">
    <Project Path="Mobile/App/App.csproj" />
  </Folder>
</Solution>
`
	if err := os.WriteFile(filepath.Join(tempDir, "sample.slnx"), []byte(slnxContent), 0o644); err != nil { // #nosec G306 -- test fixture
		t.Fatalf("write sample slnx: %v", err)
	}

	configPath := filepath.Join(tempDir, "filters.yml")
	configContent := `version: 1
source: ./sample.slnx
profiles:
  qodana:
    include:
      - src/**
    output: ./out/qodana.slnx
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil { // #nosec G306 -- test fixture
		t.Fatalf("write config: %v", err)
	}

	if _, err := Run(Request{ConfigPath: configPath, Profile: "qodana"}); err != nil {
		t.Fatalf("run generate: %v", err)
	}

	generated, err := slnx.ParseFile(filepath.Join(tempDir, "out", "qodana.slnx"))
	if err != nil {
		t.Fatalf("parse generated slnx: %v", err)
	}

	if len(generated.Folders) != 1 {
		t.Fatalf("expected 1 folder, got %d", len(generated.Folders))
	}
	if generated.Folders[0].Name != "/src/" {
		t.Fatalf("expected only /src/ folder, got %q", generated.Folders[0].Name)
	}
	if len(generated.Folders[0].Files) != 1 || generated.Folders[0].Files[0].Path != "src/Directory.Build.props" {
		t.Fatalf("expected only src file in output, got %+v", generated.Folders[0].Files)
	}
	if len(generated.Folders[0].Projects) != 1 || generated.Folders[0].Projects[0].Path != "src/api/Api.csproj" {
		t.Fatalf("expected only src project in output, got %+v", generated.Folders[0].Projects)
	}
}

func writeFixtureFiles(t *testing.T, tempDir, configContent string) string {
	t.Helper()

	slnxContent := `<Solution>
  <Folder Name="/src/">
    <Project Path="src/api/Api.csproj" />
    <Project Path="src/core/Core.csproj" />
  </Folder>
  <Folder Name="/test/">
    <Project Path="test/api/Api.Tests.csproj" />
  </Folder>
</Solution>
`
	if err := os.WriteFile(filepath.Join(tempDir, "sample.slnx"), []byte(slnxContent), 0o644); err != nil { // #nosec G306 -- test fixture
		t.Fatalf("write sample slnx: %v", err)
	}

	configPath := filepath.Join(tempDir, "filters.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil { // #nosec G306 -- test fixture
		t.Fatalf("write config: %v", err)
	}

	return configPath
}
