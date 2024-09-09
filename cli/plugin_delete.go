package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/urfave/cli"
)

var pluginDeleteCommand = cli.Command{
	Name:  "delete",
	Usage: "Delete bitrise plugin.",
	Action: func(c *cli.Context) error {
		logSubcommandParameters(c)

		if err := pluginDelete(c); err != nil {
			log.Errorf("Plugin delete failed, error: %s", err)
			os.Exit(1)
		}
		return nil
	},
	ArgsUsage: "<plugin_name>",
}

func pluginDelete(c *cli.Context) error {
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

	// Delete
	if _, found, err := plugins.LoadPlugin(name); err != nil {
		return fmt.Errorf("failed to check if plugin installed, error: %s", err)
	} else if !found {
		log.Warnf("Plugin not installed")
		return nil
	}

	log.Infof("Deleting plugin")
	if err := plugins.DeletePlugin(name); err != nil {
		return fmt.Errorf("failed to delete plugin, error: %s", err)
	}

	log.Donef("Plugin deleted")
	// ---

	return nil
}
