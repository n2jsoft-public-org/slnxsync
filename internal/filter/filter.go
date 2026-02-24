package filter

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/n2jsoft-public-org/slnxsync/internal/slnx"
)

type UnmatchedIncludeError struct {
	Patterns []string
}

func (e *UnmatchedIncludeError) Error() string {
	if e == nil || len(e.Patterns) == 0 {
		return "strict mode enabled and include patterns were not matched"
	}
	return fmt.Sprintf("strict mode enabled; unmatched include patterns: %s", strings.Join(e.Patterns, ", "))
}

func Apply(solution *slnx.Solution, includes, excludes []string, strict bool) (*slnx.Solution, error) {
	if solution == nil {
		return nil, fmt.Errorf("solution cannot be nil")
	}

	normalizedIncludes := normalizePatterns(includes)
	normalizedExcludes := normalizePatterns(excludes)
	matchedIncludes := make([]bool, len(normalizedIncludes))

	filtered := &slnx.Solution{Folders: make([]slnx.Folder, 0, len(solution.Folders))}
	for _, folder := range solution.Folders {
		folderPath := normalizeFolderPath(folder.Name)
		folderName := normalizeFolderName(folder.Name)
		markIncludeMatches(folderPath, folderName, normalizedIncludes, matchedIncludes)

		nextFolder := slnx.Folder{
			Name:     folder.Name,
			Files:    make([]slnx.File, 0, len(folder.Files)),
			Projects: make([]slnx.Project, 0, len(folder.Projects)),
		}

		for _, file := range folder.Files {
			filePath := strings.ReplaceAll(file.Path, "\\", "/")
			fileName := filepath.Base(filePath)
			if entrySelected(filePath, fileName, normalizedIncludes, normalizedExcludes, matchedIncludes) {
				nextFolder.Files = append(nextFolder.Files, file)
			}
		}

		for _, project := range folder.Projects {
			projectPath := strings.ReplaceAll(project.Path, "\\", "/")
			projectName := strings.TrimSuffix(filepath.Base(projectPath), filepath.Ext(projectPath))
			if entrySelected(projectPath, projectName, normalizedIncludes, normalizedExcludes, matchedIncludes) {
				nextFolder.Projects = append(nextFolder.Projects, project)
			}
		}

		if len(nextFolder.Files) == 0 && len(nextFolder.Projects) == 0 {
			continue
		}

		filtered.Folders = append(filtered.Folders, nextFolder)
	}

	if strict && len(normalizedIncludes) > 0 {
		unmatched := make([]string, 0)
		for idx, wasMatched := range matchedIncludes {
			if !wasMatched {
				unmatched = append(unmatched, includes[idx])
			}
		}
		if len(unmatched) > 0 {
			return nil, &UnmatchedIncludeError{Patterns: unmatched}
		}
	}

	return filtered, nil
}

func entrySelected(path, name string, includes, excludes []string, matchedIncludes []bool) bool {
	include := markIncludeMatches(path, name, includes, matchedIncludes)
	if !include {
		return false
	}

	for _, pattern := range excludes {
		if matches(pattern, path, name) {
			return false
		}
	}

	return true
}

func markIncludeMatches(path, name string, includes []string, matchedIncludes []bool) bool {
	include := len(includes) == 0
	for idx, pattern := range includes {
		if matches(pattern, path, name) {
			if idx < len(matchedIncludes) {
				matchedIncludes[idx] = true
			}
			include = true
		}
	}

	return include
}

func normalizePatterns(patterns []string) []string {
	normalized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		normalized = append(normalized, strings.ReplaceAll(pattern, "\\", "/"))
	}
	return normalized
}

func matches(pattern, projectPath, projectName string) bool {
	matchedPath, err := doublestar.PathMatch(pattern, projectPath)
	if err == nil && matchedPath {
		return true
	}

	matchedName, err := doublestar.PathMatch(pattern, projectName)
	if err != nil {
		return false
	}
	return matchedName
}

func normalizeFolderPath(folderName string) string {
	normalized := strings.ReplaceAll(folderName, "\\", "/")
	return strings.Trim(normalized, "/")
}

func normalizeFolderName(folderName string) string {
	normalized := normalizeFolderPath(folderName)
	if normalized == "" {
		return ""
	}

	parts := strings.Split(normalized, "/")
	return parts[len(parts)-1]
}
