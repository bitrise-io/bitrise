package cli

import (
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
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
