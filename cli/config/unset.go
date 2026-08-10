package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
)

// NewUnsetCommand returns the `config unset` subcommand.
func NewUnsetCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "unset KEY",
		Short:     "Remove a config key and save the global config file",
		ValidArgs: internalconfig.Keys,
		Long: fmt.Sprintf(`Remove a config key from the global config file.

Valid keys: %s`, strings.Join(internalconfig.Keys, ", ")),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			key := args[0]
			cfg, err := internalconfig.Load()
			if err != nil {
				return err
			}
			if err := cfg.Unset(key); err != nil {
				return err
			}
			if err := internalconfig.Save(cfg); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.ErrOrStderr(), "Cleared %s\n", key)
			return err
		},
	}
}
