package preview

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/n2jsoft-public-org/slnxsync/internal/config"
	"github.com/n2jsoft-public-org/slnxsync/internal/filter"
	"github.com/n2jsoft-public-org/slnxsync/internal/slnx"
)

type Request struct {
	ConfigPath string
	Profile    string
	Strict     bool
}

type Result struct {
	ProfileName      string
	SelectedProjects []string
	TotalProjects    int
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

func Run(req Request) ([]Result, error) {
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

	results := make([]Result, 0, len(profileNames))
	for _, profileName := range profileNames {
		profile := cfg.Profiles[profileName]

		filtered, err := filter.Apply(sourceSolution, profile.Include, profile.Exclude, req.Strict)
		if err != nil {
			if _, ok := err.(*filter.UnmatchedIncludeError); ok {
				return nil, &ValidationError{Err: fmt.Errorf("profile %q: %w", profileName, err)}
			}
			return nil, fmt.Errorf("apply filter for profile %q: %w", profileName, err)
		}

		projects := collectProjectPaths(filtered)
		results = append(results, Result{
			ProfileName:      profileName,
			SelectedProjects: projects,
			TotalProjects:    totalProjects,
		})
	}

	return results, nil
}

func selectedProfileNames(cfg *config.Config, requested string) ([]string, error) {
	if strings.TrimSpace(requested) == "" {
		return cfg.ProfileNames(), nil
	}

	if _, ok := cfg.Profiles[requested]; ok {
		return []string{requested}, nil
	}

	suggestions := cfg.SuggestProfileNames(requested, 3)
	if len(suggestions) == 0 {
		return nil, &ValidationError{Err: fmt.Errorf("profile %q not found", requested)}
	}

	return nil, &ValidationError{Err: fmt.Errorf("profile %q not found (did you mean: %s?)", requested, strings.Join(suggestions, ", "))}
}

func collectProjectPaths(solution *slnx.Solution) []string {
	projects := make([]string, 0)
	for _, folder := range solution.Folders {
		for _, project := range folder.Projects {
			projects = append(projects, strings.ReplaceAll(project.Path, "\\", "/"))
		}
	}
	sort.Strings(projects)
	return projects
}

func countProjects(solution *slnx.Solution) int {
	total := 0
	for _, folder := range solution.Folders {
		total += len(folder.Projects)
	}
	return total
}
