package cmd

import (
	"errors"
	"fmt"

	"github.com/n2jsoft-public-org/slnxsync/internal/generate"
	"github.com/n2jsoft-public-org/slnxsync/internal/logging"
	"github.com/spf13/cobra"
)

func newGenerateCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate one or more filtered .slnx outputs",
		Example: "slnxsync generate -c filters.yml\n" +
			"slnxsync generate -c filters.yml --profile api-ci\n" +
			"slnxsync generate -c filters.yml --profile api-ci --dry-run\n" +
			"slnxsync generate -c filters.yml --out-dir ./out",
		RunE: func(_ *cobra.Command, _ []string) error {
			logging.Verbosef("generate called with config=%q profile=%q out-dir=%q dry-run=%t strict=%t", opts.ConfigPath, opts.Profile, opts.OutDir, opts.DryRun, opts.Strict)

			if opts.ConfigPath == "" {
				return &ExitError{Code: 1, Err: errors.New("--config is required")}
			}

			results, err := generate.Run(generate.Request{
				ConfigPath: opts.ConfigPath,
				Profile:    opts.Profile,
				OutDir:     opts.OutDir,
				DryRun:     opts.DryRun,
				Strict:     opts.Strict,
			})
			if err != nil {
				var validationErr *generate.ValidationError
				if errors.As(err, &validationErr) {
					return &ExitError{Code: 2, Err: validationErr}
				}
				return &ExitError{Code: 1, Err: err}
			}

			for _, result := range results {
				if result.DryRun {
					fmt.Printf("[dry-run] profile=%s selected=%d/%d output=%s\n", result.Name, result.SelectedProjects, result.TotalProjects, result.OutputPath)
					continue
				}
				fmt.Printf("profile=%s selected=%d/%d output=%s\n", result.Name, result.SelectedProjects, result.TotalProjects, result.OutputPath)
			}

			return nil
		},
	}
}
