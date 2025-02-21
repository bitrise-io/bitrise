package cli

import (
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/output"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/urfave/cli"
)

var pluginListCommand = cli.Command{
	Name:  "list",
	Usage: "List installed bitrise plugins.",
	Action: func(c *cli.Context) error {
		logCommandParameters(c)

		if err := pluginList(c); err != nil {
			log.Errorf("Plugin list failed, error: %s", err)
			os.Exit(1)
		}
		return nil
	},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  output.FormatKey,
			Usage: "Output format. Accepted: raw, json.",
		},
	},
	ArgsUsage: "",
}

func pluginList(c *cli.Context) error {
	// Input validation
	format := c.String(output.FormatKey)
	if format == "" {
		format = output.FormatRaw
	}
	if format != output.FormatRaw && format != output.FormatJSON {
		showSubcommandHelp(c)
		return fmt.Errorf("invalid format: %s", format)
	}

	var logger Logger
	logger = NewDefaultRawLogger()
	if format == output.FormatJSON {
		logger = NewDefaultJSONLogger()
	}
	// ---

	// List
	installedPlugins, err := plugins.InstalledPluginList()
	if err != nil {
		return fmt.Errorf("failed to list plugins, error: %s", err)
	}

	if len(installedPlugins) == 0 {
		log.Warnf("No installed plugin found")
		return nil
	}

	plugins.SortByName(installedPlugins)

	pluginInfos := plugins.PluginInfos{}

	for _, plugin := range installedPlugins {
		pluginInfo, err := createPluginInfo(plugin.Name)
		if err != nil {
			return err
		}
		pluginInfos = append(pluginInfos, pluginInfo)
	}

	logger.Print(pluginInfos)
	// ---

	return nil
}
