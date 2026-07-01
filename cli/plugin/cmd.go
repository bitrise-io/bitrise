package plugin

import (
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmd returns the `bitrise plugin` parent command.
func NewCmd() *cobra.Command {
	pluginCommand := &cobra.Command{
		Use:   "plugin",
		Short: "Plugin handling.",
		RunE:  cmdutil.RequireKnownSubcommand,
	}

	pluginCommand.AddCommand(
		newInstallCommand(),
		newUpdateCommand(),
		newDeleteCommand(),
		newInfoCommand(),
		newListCommand(),
	)

	return pluginCommand
}
