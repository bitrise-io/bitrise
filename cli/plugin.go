package cli

import (
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Plugin handling.",
}

func init() {
	pluginCmd.AddCommand(pluginInstallCmd, pluginUpdateCmd, pluginDeleteCmd, pluginInfoCmd, pluginListCmd)
}

