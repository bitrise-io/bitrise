package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
)

// NewPathCommand returns the `config path` subcommand.
func NewPathCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the absolute path of the global config file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			p, err := internalconfig.Path()
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), p)
			return err
		},
	}
}
