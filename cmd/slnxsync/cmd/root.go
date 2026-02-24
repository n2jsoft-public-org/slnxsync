package cmd

import (
	"fmt"
	"os"

	"github.com/n2jsoft-public-org/slnxsync/internal/logging"
	"github.com/spf13/cobra"
)

func Execute(buildInfo BuildInfo) int {
	if err := newRootCmd(buildInfo).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if exitErr, ok := err.(*ExitError); ok {
			return exitErr.Code
		}
		return 1
	}
	return 0
}

func newRootCmd(buildInfo BuildInfo) *cobra.Command {
	opts := &Options{}

	rootCmd := &cobra.Command{
		Use:           "slnxsync",
		Short:         "Generate CI-focused solution variants from a .slnx source",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.Version {
				return printBuildInfo(cmd.OutOrStdout(), buildInfo)
			}
			return cmd.Help()
		},
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			logging.SetVerbose(opts.Verbose)
		},
	}

	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&opts.ConfigPath, "config", "c", "", "Path to filter configuration file")
	flags.StringVarP(&opts.Profile, "profile", "p", "", "Run one profile only")
	flags.StringVarP(&opts.OutDir, "out-dir", "o", "", "Override output directory")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "Print actions without writing files")
	flags.BoolVar(&opts.Strict, "strict", false, "Fail on unmatched include patterns")
	flags.BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose logs")
	rootCmd.Flags().BoolVar(&opts.Version, "version", false, "Print version information")

	rootCmd.AddCommand(newGenerateCmd(opts))
	rootCmd.AddCommand(newValidateCmd(opts))
	rootCmd.AddCommand(newPreviewCmd(opts))
	rootCmd.AddCommand(newVersionCmd(buildInfo))

	return rootCmd
}
