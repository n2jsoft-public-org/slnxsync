package preview

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSingleProfile(t *testing.T) {
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
		t.Fatalf("run preview: %v", err)
	}
	if len(results) != 1 || results[0].ProfileName != "api" {
		t.Fatalf("expected only api profile result, got %+v", results)
	}
	if len(results[0].SelectedProjects) != 1 {
		t.Fatalf("expected 1 selected project, got %d", len(results[0].SelectedProjects))
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

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if !strings.Contains(err.Error(), "unmatched include patterns") {
		t.Fatalf("unexpected strict error message: %v", err)
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
