package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAndValidateValidConfig(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "MyProduct.slnx")
	if err := os.WriteFile(sourcePath, []byte("<Solution/>"), 0o644); err != nil { // #nosec G306 -- test fixture
		t.Fatalf("write source file: %v", err)
	}

	configPath := filepath.Join(tempDir, "filters.yml")
	content := "version: 1\n" +
		"source: ./MyProduct.slnx\n" +
		"profiles:\n" +
		"  api-ci:\n" +
		"    include:\n" +
		"      - src/Api/**\n" +
		"    exclude:\n" +
		"      - '**/*.Tests'\n" +
		"    output: ./out/Api.slnx\n"
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil { // #nosec G306 -- test fixture
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	validationErrs := cfg.Validate(filepath.Dir(configPath))
	if len(validationErrs) != 0 {
		t.Fatalf("expected no validation errors, got: %+v", validationErrs)
	}
}

func TestValidateMissingRequiredFields(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	validationErrs := cfg.Validate(t.TempDir())
	if len(validationErrs) < 3 {
		t.Fatalf("expected required-field errors, got: %+v", validationErrs)
	}

	summary := ValidationErrors{Errors: validationErrs}.Error()
	if !strings.Contains(summary, "version") || !strings.Contains(summary, "source") || !strings.Contains(summary, "profiles") {
		t.Fatalf("expected version/source/profiles in error summary, got: %s", summary)
	}
}

func TestValidateOutputConflictAndPattern(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "Product.slnx")
	if err := os.WriteFile(sourcePath, []byte("<Solution/>"), 0o644); err != nil { // #nosec G306 -- test fixture
		t.Fatalf("write source file: %v", err)
	}

	cfg := &Config{
		Version: 1,
		Source:  "./Product.slnx",
		Profiles: map[string]Profile{
			"a": {
				Include: []string{"["},
				Output:  "./out/shared.slnx",
			},
			"b": {
				Output: "./out/shared.slnx",
			},
		},
	}

	validationErrs := cfg.Validate(tempDir)
	if len(validationErrs) == 0 {
		t.Fatal("expected validation errors")
	}

	summary := ValidationErrors{Errors: validationErrs}.Error()
	if !strings.Contains(summary, "conflicts") {
		t.Fatalf("expected output conflict error, got: %s", summary)
	}
	if !strings.Contains(summary, "invalid glob pattern") {
		t.Fatalf("expected invalid pattern error, got: %s", summary)
	}
}

func TestLoadDuplicateProfileNames(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "Product.slnx")
	if err := os.WriteFile(sourcePath, []byte("<Solution/>"), 0o644); err != nil { // #nosec G306 -- test fixture
		t.Fatalf("write source file: %v", err)
	}

	configPath := filepath.Join(tempDir, "filters.yml")
	content := "version: 1\n" +
		"source: ./Product.slnx\n" +
		"profiles:\n" +
		"  api-ci:\n" +
		"    output: ./out/a.slnx\n" +
		"  api-ci:\n" +
		"    output: ./out/b.slnx\n"
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil { // #nosec G306 -- test fixture
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected duplicate profile load error")
	}
	if !strings.Contains(err.Error(), "already defined") {
		t.Fatalf("expected duplicate mapping key error, got: %v", err)
	}
}
