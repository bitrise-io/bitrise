package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
)

// NewSetCommand returns the `config set` subcommand.
func NewSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "set KEY VALUE",
		Short:     "Set a config key and save the global config file",
		ValidArgs: internalconfig.Keys,
		Long: fmt.Sprintf(`Set a config key in the global config file.

Valid keys: %s`,
			strings.Join(internalconfig.Keys, ", "),
		),
		Example: `  bitrise config set api_base_url https://staging-api.bitrise.io/v0.1`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			key, value := args[0], args[1]
			cfg, err := internalconfig.Load()
			if err != nil {
				return err
			}
			if err := cfg.Set(key, value); err != nil {
				return err
			}
			if err := internalconfig.Save(cfg); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.ErrOrStderr(), "Saved %s\n", key)
			return err
		},
	}
}
