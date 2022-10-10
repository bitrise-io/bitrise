package cli

import (
	"errors"
	"fmt"
	"os"

	log "github.com/bitrise-io/bitrise/advancedlog"
	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/urfave/cli"
)

var pluginInfoCommand = cli.Command{
	Name:  "info",
	Usage: "Installed bitrise plugin's info",
	Action: func(c *cli.Context) error {
		if err := pluginInfo(c); err != nil {
			log.Errorf("Plugin info failed, error: %s", err)
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
	ArgsUsage: "<plugin_name>",
}

func createPluginInfo(name string) (plugins.PluginInfoModel, error) {
	plugin, found, err := plugins.LoadPlugin(name)
	if err != nil {
		return plugins.PluginInfoModel{}, fmt.Errorf("failed to check if plugin installed, error: %s", err)
	} else if !found {
		return plugins.PluginInfoModel{}, fmt.Errorf("plugin is not installed")
	}

	route, found, err := plugins.ReadPluginRoute(plugin.Name)
	if err != nil {
		return plugins.PluginInfoModel{}, fmt.Errorf("failed to read plugin route, error: %s", err)
	} else if !found {
		return plugins.PluginInfoModel{}, errors.New("no route found for loaded plugin")
	}

	pluginVersionPtr, err := plugins.GetPluginVersion(plugin.Name)
	if err != nil {
		return plugins.PluginInfoModel{}, fmt.Errorf("failed to read plugin version, error: %s", err)
	}

	pluginDefinitionPth := plugins.GetPluginDefinitionPath(plugin.Name)

	pluginInfo := plugins.PluginInfoModel{
		Name:          plugin.Name,
		Version:       pluginVersionPtr.String(),
		Source:        route.Source,
		Plugin:        plugin,
		DefinitionPth: pluginDefinitionPth,
	}

	return pluginInfo, nil
}

func pluginInfo(c *cli.Context) error {
	// Input validation
	args := c.Args()
	if len(args) == 0 {
		showSubcommandHelp(c)
		return errors.New("plugin_name not defined")
	}

	name := args[0]
	if name == "" {
		showSubcommandHelp(c)
		return errors.New("plugin_name not defined")
	}

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
		logger = NewDefaultJSONLoger()
	}
	// ---

	// Info
	pluginInfo, err := createPluginInfo(name)
	if err != nil {
		return err
	}

	logger.Print(pluginInfo)
	// ---

	return nil
}
