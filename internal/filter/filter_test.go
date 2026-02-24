package filter

import (
	"errors"
	"testing"

	"github.com/n2jsoft-public-org/slnxsync/internal/slnx"
)

func TestApplyIncludeDefaultsToAll(t *testing.T) {
	t.Parallel()

	solution := fixtureSolution()
	filtered, err := Apply(solution, nil, nil, false)
	if err != nil {
		t.Fatalf("apply filter: %v", err)
	}

	if got := countProjects(filtered); got != 4 {
		t.Fatalf("expected 4 projects, got %d", got)
	}
	if got := countFiles(filtered); got != 3 {
		t.Fatalf("expected 3 files, got %d", got)
	}
	if got := len(filtered.Folders); got != 3 {
		t.Fatalf("expected 3 folders, got %d", got)
	}
}

func TestApplyIncludeExcludePrecedence(t *testing.T) {
	t.Parallel()

	solution := fixtureSolution()
	filtered, err := Apply(solution, []string{"src/**"}, []string{"**/*.Tests"}, false)
	if err != nil {
		t.Fatalf("apply filter: %v", err)
	}

	if got := countProjects(filtered); got != 2 {
		t.Fatalf("expected 2 projects after exclude, got %d", got)
	}
	if got := countFiles(filtered); got != 1 {
		t.Fatalf("expected 1 file after exclude, got %d", got)
	}
	if got := len(filtered.Folders); got != 1 {
		t.Fatalf("expected 1 folder after pruning, got %d", got)
	}
}

func TestApplyMatchesPathAndProjectName(t *testing.T) {
	t.Parallel()

	solution := fixtureSolution()
	filtered, err := Apply(solution, []string{"Card*.Core", "**/Expenses.csproj"}, nil, false)
	if err != nil {
		t.Fatalf("apply filter: %v", err)
	}

	if got := countProjects(filtered); got != 2 {
		t.Fatalf("expected 2 projects selected by path/name, got %d", got)
	}
}

func TestApplyStrictUnmatchedInclude(t *testing.T) {
	t.Parallel()

	solution := fixtureSolution()
	_, err := Apply(solution, []string{"src/**", "missing/**"}, nil, true)
	if err == nil {
		t.Fatal("expected strict unmatched include error")
	}

	var unmatchedErr *UnmatchedIncludeError
	if !errors.As(err, &unmatchedErr) {
		t.Fatalf("expected UnmatchedIncludeError, got %T", err)
	}
	if len(unmatchedErr.Patterns) != 1 || unmatchedErr.Patterns[0] != "missing/**" {
		t.Fatalf("unexpected unmatched patterns: %+v", unmatchedErr.Patterns)
	}
}

func fixtureSolution() *slnx.Solution {
	return &slnx.Solution{
		Folders: []slnx.Folder{
			{
				Name: "/src/",
				Files: []slnx.File{
					{Path: "src/Directory.Build.props"},
				},
				Projects: []slnx.Project{
					{Path: "src/cards/Cards.Core/Cards.Core.csproj"},
					{Path: "src/expenses/Expenses/Expenses.csproj"},
					{Path: "src/expenses/Expenses.Tests/Expenses.Tests.csproj"},
				},
			},
			{
				Name: "/test/",
				Files: []slnx.File{
					{Path: "test/Directory.Build.props"},
				},
				Projects: []slnx.Project{
					{Path: "test/cards/Cards.Core.Tests/Cards.Core.Tests.csproj"},
				},
			},
			{
				Name: "/Build/",
				Files: []slnx.File{
					{Path: "Directory.Build.props"},
				},
			},
		},
	}
}

func countProjects(solution *slnx.Solution) int {
	total := 0
	for _, folder := range solution.Folders {
		total += len(folder.Projects)
	}
	return total
}

func countFiles(solution *slnx.Solution) int {
	total := 0
	for _, folder := range solution.Folders {
		total += len(folder.Files)
	}
	return total
}
