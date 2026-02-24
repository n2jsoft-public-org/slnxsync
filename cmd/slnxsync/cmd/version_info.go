package cmd

import (
	"fmt"
	"io"
)

type BuildInfo struct {
	Version   string
	Commit    string
	BuildDate string
}

func (bi BuildInfo) normalized() BuildInfo {
	if bi.Version == "" {
		bi.Version = "dev"
	}
	if bi.Commit == "" {
		bi.Commit = "unknown"
	}
	if bi.BuildDate == "" {
		bi.BuildDate = "unknown"
	}
	return bi
}

func printBuildInfo(w io.Writer, bi BuildInfo) error {
	normalized := bi.normalized()
	if _, err := fmt.Fprintf(w, "version: %s\n", normalized.Version); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "commit: %s\n", normalized.Commit); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "buildDate: %s\n", normalized.BuildDate); err != nil {
		return err
	}
	return nil
}
