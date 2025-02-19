package cli

import (
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/urfave/cli"
)

var pluginCommand = cli.Command{
	Name:  "plugin",
	Usage: "Plugin handling.",
	Subcommands: []cli.Command{
		pluginInstallCommand,
		pluginUpdateCommand,
		pluginDeleteCommand,
		pluginInfoCommand,
		pluginListCommand,
	},
}

func showSubcommandHelp(c *cli.Context) {
	if err := cli.ShowSubcommandHelp(c); err != nil {
		log.Warnf("Failed to show help, error: %s", err)
	}
}
