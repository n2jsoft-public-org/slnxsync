package config

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Profiles              map[string]Profile `yaml:"profiles"`
	Source                string             `yaml:"source"`
	duplicateProfileNames []string
	Version               int `yaml:"version"`
}

type Profile struct {
	Output  string   `yaml:"output"`
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

type ValidationError struct {
	Field   string
	Message string
}

type ValidationErrors struct {
	Errors []ValidationError
}

func (e ValidationErrors) Error() string {
	var builder strings.Builder
	builder.WriteString("configuration validation failed:\n")
	for _, validationError := range e.Errors {
		builder.WriteString("- ")
		if validationError.Field != "" {
			builder.WriteString(validationError.Field)
			builder.WriteString(": ")
		}
		builder.WriteString(validationError.Message)
		builder.WriteString("\n")
	}
	return strings.TrimRight(builder.String(), "\n")
}

func Load(configPath string) (*Config, error) {
	content, err := os.ReadFile(configPath) // #nosec G304 -- configPath from user input
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, fmt.Errorf("parse config yaml: %w", err)
	}

	var cfg Config
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config yaml: %w", err)
	}

	cfg.duplicateProfileNames = findDuplicateProfileNames(&root)
	return &cfg, nil
}

func (c *Config) Validate(baseDir string) []ValidationError {
	var validationErrs []ValidationError

	if c.Version == 0 {
		validationErrs = append(validationErrs, ValidationError{Field: "version", Message: "is required"})
	} else if c.Version != 1 {
		validationErrs = append(validationErrs, ValidationError{Field: "version", Message: "must be 1"})
	}

	if strings.TrimSpace(c.Source) == "" {
		validationErrs = append(validationErrs, ValidationError{Field: "source", Message: "is required"})
	} else {
		sourcePath := c.Source
		if !filepath.IsAbs(sourcePath) {
			sourcePath = filepath.Join(baseDir, sourcePath)
		}
		if _, err := os.Stat(sourcePath); err != nil {
			validationErrs = append(validationErrs, ValidationError{
				Field:   "source",
				Message: fmt.Sprintf("file does not exist: %s", c.Source),
			})
		}
	}

	if len(c.Profiles) == 0 {
		validationErrs = append(validationErrs, ValidationError{Field: "profiles", Message: "must contain at least one profile"})
	}

	for _, name := range c.duplicateProfileNames {
		validationErrs = append(validationErrs, ValidationError{
			Field:   fmt.Sprintf("profiles.%s", name),
			Message: "profile name is duplicated",
		})
	}

	profileNames := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		profileNames = append(profileNames, name)
	}
	sort.Strings(profileNames)

	outputOwners := make(map[string]string)
	for _, profileName := range profileNames {
		profile := c.Profiles[profileName]
		profileField := fmt.Sprintf("profiles.%s", profileName)

		if strings.TrimSpace(profileName) == "" {
			validationErrs = append(validationErrs, ValidationError{Field: "profiles", Message: "profile name cannot be empty"})
		}

		if strings.TrimSpace(profile.Output) == "" {
			validationErrs = append(validationErrs, ValidationError{Field: profileField + ".output", Message: "is required"})
		} else {
			normalizedOutput := normalizePath(baseDir, profile.Output)
			if owner, exists := outputOwners[normalizedOutput]; exists {
				validationErrs = append(validationErrs, ValidationError{
					Field:   profileField + ".output",
					Message: fmt.Sprintf("conflicts with profiles.%s.output", owner),
				})
			} else {
				outputOwners[normalizedOutput] = profileName
			}
		}

		validationErrs = append(validationErrs, validatePatterns(profileField+".include", profile.Include)...)
		validationErrs = append(validationErrs, validatePatterns(profileField+".exclude", profile.Exclude)...)
	}

	return validationErrs
}

func normalizePath(baseDir, value string) string {
	normalized := value
	if !filepath.IsAbs(normalized) {
		normalized = filepath.Join(baseDir, normalized)
	}
	normalized = filepath.Clean(normalized)
	return filepath.ToSlash(normalized)
}

func validatePatterns(fieldPrefix string, patterns []string) []ValidationError {
	validationErrs := make([]ValidationError, 0)
	for idx, patternValue := range patterns {
		field := fmt.Sprintf("%s[%d]", fieldPrefix, idx)
		if strings.TrimSpace(patternValue) == "" {
			validationErrs = append(validationErrs, ValidationError{Field: field, Message: "pattern cannot be empty"})
			continue
		}
		if _, err := path.Match(patternValue, "sample"); err != nil {
			validationErrs = append(validationErrs, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("invalid glob pattern %q", patternValue),
			})
		}
	}
	return validationErrs
}

func findDuplicateProfileNames(root *yaml.Node) []string {
	if root == nil || len(root.Content) == 0 {
		return nil
	}

	documentNode := root.Content[0]
	if documentNode.Kind != yaml.MappingNode {
		return nil
	}

	for idx := 0; idx+1 < len(documentNode.Content); idx += 2 {
		keyNode := documentNode.Content[idx]
		valueNode := documentNode.Content[idx+1]
		if keyNode.Value != "profiles" || valueNode.Kind != yaml.MappingNode {
			continue
		}

		counts := map[string]int{}
		for profileIdx := 0; profileIdx+1 < len(valueNode.Content); profileIdx += 2 {
			profileKey := valueNode.Content[profileIdx].Value
			counts[profileKey]++
		}

		duplicates := make([]string, 0)
		for name, count := range counts {
			if count > 1 {
				duplicates = append(duplicates, name)
			}
		}
		sort.Strings(duplicates)
		return duplicates
	}

	return nil
}
