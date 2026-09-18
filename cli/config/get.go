package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
)

// NewGetCommand returns the `config get` subcommand.
func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "get KEY",
		Short:     "Print the value of a single config key",
		ValidArgs: internalconfig.Keys,
		Long: fmt.Sprintf(`Print the value of one config key from the global config file.

Valid keys: %s`,
			strings.Join(internalconfig.Keys, ", "),
		),
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			cfg, err := internalconfig.Load()
			if err != nil {
				return err
			}
			v, err := cfg.Get(args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), v)
			return err
		},
	}
}
