package main

import (
	"os"

	"github.com/n2jsoft-public-org/slnxsync/cmd/slnxsync/cmd"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func main() {
	os.Exit(cmd.Execute(cmd.BuildInfo{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}))
}
