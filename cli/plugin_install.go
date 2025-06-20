package cli

import (
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/urfave/cli"
)

var pluginInstallCommand = cli.Command{
	Name:  "install",
	Usage: "Install bitrise plugin.",
	Action: func(c *cli.Context) error {
		logCommandParameters(c)

		if err := pluginInstall(c); err != nil {
			log.Errorf("Plugin install failed, error: %s", err)
			os.Exit(1)
		}
		return nil
	},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "version",
			Usage: "Plugin version tag.",
		},
		cli.StringFlag{
			Name:  "source",
			Usage: "Deprecated!!! Specify as arg instead - Plugin source url (can be local path or remote url).",
		},
	},
	ArgsUsage: "<plugin_source_remote_or_local_url>",
}

func pluginInstall(c *cli.Context) error {
	// Input validation
	pluginSource := ""
	if args := c.Args(); len(args) > 0 {
		pluginSource = args[0]
	} else {
		pluginSource = c.String("source")
	}

	pluginVersionTag := c.String("version")

	if pluginSource == "" {
		showSubcommandHelp(c)
		return fmt.Errorf("plugin source not defined")
	}
	// ---

	// Install
	log.Infof("Installing plugin")

	plugin, version, err := plugins.InstallPlugin(pluginSource, pluginVersionTag)
	if err != nil {
		return err
	}

	if len(plugin.Description) > 0 {
		log.Print()
		log.Infof("Description:")
		log.Print(plugin.Description)
	}

	log.Print()
	if version == "" {
		log.Donef("Local plugin (%s) installed ", plugin.Name)
	} else {
		log.Donef("Plugin (%s) with version (%s) installed ", plugin.Name, version)
	}
	// ---

	return nil
}
