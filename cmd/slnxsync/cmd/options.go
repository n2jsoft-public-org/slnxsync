package cmd

type Options struct {
	ConfigPath string
	Profile    string
	OutDir     string
	DryRun     bool
	Strict     bool
	Verbose    bool
	Version    bool
}
