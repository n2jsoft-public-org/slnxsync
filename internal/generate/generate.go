package generate

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/n2jsoft-public-org/slnxsync/internal/config"
	"github.com/n2jsoft-public-org/slnxsync/internal/filter"
	"github.com/n2jsoft-public-org/slnxsync/internal/slnx"
)

type Request struct {
	ConfigPath string
	Profile    string
	OutDir     string
	DryRun     bool
	Strict     bool
}

type ProfileResult struct {
	Name             string
	OutputPath       string
	SelectedProjects int
	TotalProjects    int
	DryRun           bool
}

type ValidationError struct {
	Err error
}

func (e *ValidationError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *ValidationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func Run(req Request) ([]ProfileResult, error) {
	cfg, err := config.Load(req.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	configDir := filepath.Dir(req.ConfigPath)
	validationErrs := cfg.Validate(configDir)
	if len(validationErrs) > 0 {
		return nil, &ValidationError{Err: config.ValidationErrors{Errors: validationErrs}}
	}

	profileNames, err := selectedProfileNames(cfg, req.Profile)
	if err != nil {
		return nil, err
	}

	sourcePath := cfg.Source
	if !filepath.IsAbs(sourcePath) {
		sourcePath = filepath.Join(configDir, sourcePath)
	}
	sourceSolution, err := slnx.ParseFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("parse source .slnx: %w", err)
	}
	totalProjects := countProjects(sourceSolution)

	resolvedOutputs, err := resolveOutputPaths(configDir, req.OutDir, profileNames, cfg)
	if err != nil {
		return nil, err
	}

	results := make([]ProfileResult, 0, len(profileNames))
	for _, profileName := range profileNames {
		profile := cfg.Profiles[profileName]

		filtered, err := filter.Apply(sourceSolution, profile.Include, profile.Exclude, req.Strict)
		if err != nil {
			if _, ok := err.(*filter.UnmatchedIncludeError); ok {
				return nil, &ValidationError{Err: fmt.Errorf("profile %q: %w", profileName, err)}
			}
			return nil, fmt.Errorf("apply filter for profile %q: %w", profileName, err)
		}

		if !req.DryRun {
			if err := slnx.WriteFile(resolvedOutputs[profileName], filtered); err != nil {
				return nil, fmt.Errorf("write profile %q output: %w", profileName, err)
			}
		}

		results = append(results, ProfileResult{
			Name:             profileName,
			OutputPath:       resolvedOutputs[profileName],
			SelectedProjects: countProjects(filtered),
			TotalProjects:    totalProjects,
			DryRun:           req.DryRun,
		})
	}

	return results, nil
}

func selectedProfileNames(cfg *config.Config, profile string) ([]string, error) {
	if strings.TrimSpace(profile) != "" {
		if _, ok := cfg.Profiles[profile]; !ok {
			suggestions := cfg.SuggestProfileNames(profile, 3)
			if len(suggestions) == 0 {
				return nil, &ValidationError{Err: fmt.Errorf("profile %q not found", profile)}
			}
			return nil, &ValidationError{Err: fmt.Errorf("profile %q not found (did you mean: %s?)", profile, strings.Join(suggestions, ", "))}
		}
		return []string{profile}, nil
	}

	return cfg.ProfileNames(), nil
}

func resolveOutputPaths(configDir, outDirOverride string, profileNames []string, cfg *config.Config) (map[string]string, error) {
	resolved := make(map[string]string, len(profileNames))
	owners := make(map[string]string, len(profileNames))

	for _, profileName := range profileNames {
		output := cfg.Profiles[profileName].Output
		var outputPath string
		if strings.TrimSpace(outDirOverride) != "" {
			outputPath = filepath.Join(outDirOverride, filepath.Base(output))
		} else {
			outputPath = output
		}
		if !filepath.IsAbs(outputPath) {
			outputPath = filepath.Join(configDir, outputPath)
		}
		outputPath = filepath.Clean(outputPath)

		if owner, exists := owners[outputPath]; exists {
			return nil, &ValidationError{Err: fmt.Errorf("output path conflict between profiles %q and %q", owner, profileName)}
		}
		owners[outputPath] = profileName
		resolved[profileName] = outputPath
	}

	return resolved, nil
}

func countProjects(solution *slnx.Solution) int {
	total := 0
	for _, folder := range solution.Folders {
		total += len(folder.Projects)
	}
	return total
}
