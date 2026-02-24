package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommandPrintsBuildInfo(t *testing.T) {
	t.Parallel()

	root := newRootCmd(BuildInfo{
		Version:   "1.2.3",
		Commit:    "abc1234",
		BuildDate: "2026-02-24T12:00:00Z",
	})

	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute version command: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "version: 1.2.3") {
		t.Fatalf("expected version in output, got: %q", got)
	}
	if !strings.Contains(got, "commit: abc1234") {
		t.Fatalf("expected commit in output, got: %q", got)
	}
	if !strings.Contains(got, "buildDate: 2026-02-24T12:00:00Z") {
		t.Fatalf("expected build date in output, got: %q", got)
	}
}

func TestRootVersionFlagPrintsBuildInfo(t *testing.T) {
	t.Parallel()

	root := newRootCmd(BuildInfo{
		Version:   "9.9.9",
		Commit:    "deadbee",
		BuildDate: "2026-02-24T13:00:00Z",
	})

	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"--version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute --version: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "version: 9.9.9") {
		t.Fatalf("expected version in output, got: %q", got)
	}
	if !strings.Contains(got, "commit: deadbee") {
		t.Fatalf("expected commit in output, got: %q", got)
	}
	if !strings.Contains(got, "buildDate: 2026-02-24T13:00:00Z") {
		t.Fatalf("expected build date in output, got: %q", got)
	}
}

func TestBuildInfoNormalization(t *testing.T) {
	t.Parallel()

	root := newRootCmd(BuildInfo{})
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute version command: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "version: dev") {
		t.Fatalf("expected default version, got: %q", got)
	}
	if !strings.Contains(got, "commit: unknown") {
		t.Fatalf("expected default commit, got: %q", got)
	}
	if !strings.Contains(got, "buildDate: unknown") {
		t.Fatalf("expected default buildDate, got: %q", got)
	}
}
