package cmd

import "github.com/spf13/cobra"

func newVersionCmd(buildInfo BuildInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return printBuildInfo(cmd.OutOrStdout(), buildInfo)
		},
	}
}
