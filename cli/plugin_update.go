package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/go-utils/log"
	"github.com/urfave/cli"
)

var pluginUpdateCommand = cli.Command{
	Name:  "update",
	Usage: "Update bitrise plugin.",
	Action: func(c *cli.Context) error {
		if err := pluginUpdate(c); err != nil {
			log.Errorf("Plugin update failed, error: %s", err)
			os.Exit(1)
		}
		return nil
	},
	ArgsUsage: "<plugin_name>",
}

func pluginUpdate(c *cli.Context) error {
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
	// ---

	// Update
	plugin, found, err := plugins.LoadPlugin(name)
	if err != nil {
		return fmt.Errorf("failed to check if plugin installed, error: %s", err)
	} else if !found {
		return fmt.Errorf("plugin is not installed")
	}

	if newVersion, err := plugins.CheckForNewVersion(plugin); err != nil {
		return fmt.Errorf("failed to check for plugin new version, error: %s", err)
	} else if newVersion != "" {
		log.Infof("Installing new version (%s)", newVersion)

		route, found, err := plugins.ReadPluginRoute(plugin.Name)
		if err != nil {
			return fmt.Errorf("failed to read plugin route, error: %s", err)
		}
		if !found {
			return errors.New("no route found for already loaded plugin")
		}

		plugin, version, err := plugins.InstallPlugin(route.Source, newVersion)
		if err != nil {
			return fmt.Errorf("failed to install plugin from (%s), error: %s", route.Source, err)
		}

		fmt.Println()
		log.Donef("Plugin (%s) with version (%s) installed ", plugin.Name, version)

		if len(plugin.Description) > 0 {
			fmt.Println()
			log.Infof("Description:")
			fmt.Println(plugin.Description)
			fmt.Println()
		}
	} else {
		log.Donef("No new version available")
	}
	// ---

	return nil
}
