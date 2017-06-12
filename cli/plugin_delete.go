package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/go-utils/log"
	"github.com/urfave/cli"
)

var pluginDeleteCommand = cli.Command{
	Name:  "delete",
	Usage: "Delete bitrise plugin.",
	Action: func(c *cli.Context) error {
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

	versionPtr, err := plugins.GetPluginVersion(name)
	if err != nil {
		return fmt.Errorf("failed to read plugin version, error: %s", err)
	}

	if versionPtr != nil {
		log.Infof("=> Deleting plugin (%s) with version (%s) ...", name, versionPtr.String())
	} else {
		log.Infof("=> Deleting local plugin (%s) ...", name)
	}
	if err := plugins.DeletePlugin(name); err != nil {
		return fmt.Errorf("failed to delete plugin, error: %s", err)
	}

	fmt.Println()
	log.Donef("Plugin deleted")
	// ---

	return nil
}
