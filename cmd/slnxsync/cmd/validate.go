package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/n2jsoft-public-org/slnxsync/internal/config"
	"github.com/n2jsoft-public-org/slnxsync/internal/logging"
	"github.com/spf13/cobra"
)

func newValidateCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the filter configuration",
		Example: "slnxsync validate -c filters.yml\n" +
			"slnxsync validate -c filters.yml --strict",
		RunE: func(_ *cobra.Command, _ []string) error {
			logging.Verbosef("validate called with config=%q strict=%t", opts.ConfigPath, opts.Strict)

			if opts.ConfigPath == "" {
				return &ExitError{Code: 1, Err: errors.New("--config is required")}
			}

			cfg, err := config.Load(opts.ConfigPath)
			if err != nil {
				return &ExitError{Code: 1, Err: fmt.Errorf("load config: %w", err)}
			}

			validationErrs := cfg.Validate(filepath.Dir(opts.ConfigPath))
			if len(validationErrs) > 0 {
				return &ExitError{Code: 2, Err: config.ValidationErrors{Errors: validationErrs}}
			}

			fmt.Printf("configuration is valid (%d profile(s))\n", len(cfg.Profiles))
			return nil
		},
	}
}
