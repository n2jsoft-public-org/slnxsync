package cmd

import (
	"errors"
	"fmt"

	"github.com/n2jsoft-public-org/slnxsync/internal/logging"
	"github.com/n2jsoft-public-org/slnxsync/internal/preview"
	"github.com/spf13/cobra"
)

func newPreviewCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "preview",
		Short: "Preview selected projects for a profile",
		Example: "slnxsync preview -c filters.yml\n" +
			"slnxsync preview -c filters.yml --profile api-ci\n" +
			"slnxsync preview -c filters.yml --profile api-ci --strict",
		RunE: func(_ *cobra.Command, _ []string) error {
			logging.Verbosef("preview called with config=%q profile=%q strict=%t", opts.ConfigPath, opts.Profile, opts.Strict)

			if opts.ConfigPath == "" {
				return &ExitError{Code: 1, Err: errors.New("--config is required")}
			}

			results, err := preview.Run(preview.Request{
				ConfigPath: opts.ConfigPath,
				Profile:    opts.Profile,
				Strict:     opts.Strict,
			})
			if err != nil {
				var validationErr *preview.ValidationError
				if errors.As(err, &validationErr) {
					return &ExitError{Code: 2, Err: validationErr}
				}
				return &ExitError{Code: 1, Err: err}
			}

			for _, result := range results {
				fmt.Printf("profile=%s selected=%d/%d\n", result.ProfileName, len(result.SelectedProjects), result.TotalProjects)
				for _, projectPath := range result.SelectedProjects {
					fmt.Printf("  - %s\n", projectPath)
				}
			}

			return nil
		},
	}
}
