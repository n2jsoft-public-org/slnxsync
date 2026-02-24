package slnx

import (
	"strings"
	"testing"
)

func TestParseNormalizesAndSorts(t *testing.T) {
	t.Parallel()

	input := `<Solution>
  <Folder Name="/b/">
		<Project Path="src\z\Z.csproj" />
    <Project Path="src/a/A.csproj" />
		<File Path="src\z\z.props" />
    <File Path="src/a/a.props" />
  </Folder>
  <Folder Name="/a/">
		<Project Path="src\b\B.csproj" />
  </Folder>
</Solution>`

	solution, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if got, want := solution.Folders[0].Name, "/a/"; got != want {
		t.Fatalf("folder order mismatch: got %q want %q", got, want)
	}
	if strings.Contains(solution.Folders[1].Projects[0].Path, "\\") {
		t.Fatalf("expected normalized project path, got: %q", solution.Folders[1].Projects[0].Path)
	}
	if got, want := solution.Folders[1].Projects[0].Path, "src/a/A.csproj"; got != want {
		t.Fatalf("project order mismatch: got %q want %q", got, want)
	}
}

func TestRoundTripParseWriteParse(t *testing.T) {
	t.Parallel()

	input := `<Solution>
  <Folder Name="/src/">
		<Project Path="src\b\B.csproj" Id="abc-123">
			<BuildDependency Project="src\a\A.csproj" />
    </Project>
    <Project Path="src/a/A.csproj" />
  </Folder>
</Solution>`

	first, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("parse first: %v", err)
	}

	marshaled, err := Marshal(first)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	second, err := Parse(marshaled)
	if err != nil {
		t.Fatalf("parse second: %v", err)
	}

	if len(first.Folders) != len(second.Folders) {
		t.Fatalf("folders length mismatch: %d vs %d", len(first.Folders), len(second.Folders))
	}
	if first.Folders[0].Projects[0].Path != second.Folders[0].Projects[0].Path {
		t.Fatalf("project path mismatch after roundtrip: %q vs %q", first.Folders[0].Projects[0].Path, second.Folders[0].Projects[0].Path)
	}
	if strings.TrimSpace(first.Folders[0].Projects[1].InnerXML) != strings.TrimSpace(second.Folders[0].Projects[1].InnerXML) {
		t.Fatalf("project inner xml mismatch after roundtrip")
	}
}
