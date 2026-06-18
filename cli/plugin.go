package cli

import (
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/spf13/cobra"
)

var pluginCommand = &cobra.Command{
	Use:   "plugin",
	Short: "Plugin handling.",
	RunE:  requireKnownSubcommand,
}

func init() {
	pluginCommand.AddCommand(
		pluginInstallCommand,
		pluginUpdateCommand,
		pluginDeleteCommand,
		pluginInfoCommand,
		pluginListCommand,
	)
}

func showSubcommandHelp(cmd *cobra.Command) {
	if err := cmd.Help(); err != nil {
		log.Warnf("Failed to show help, error: %s", err)
	}
}
