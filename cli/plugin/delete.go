package plugin

import (
	"errors"
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <plugin_name>",
		Short: "Delete bitrise plugin.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := pluginDelete(cmd, args); err != nil {
				log.Errorf("Plugin delete failed, error: %s", err)
				os.Exit(1)
			}
			return nil
		},
	}
}

func pluginDelete(cmd *cobra.Command, args []string) error {
	// Input validation
	if len(args) == 0 {
		cmdutil.ShowSubcommandHelp(cmd)
		return errors.New("plugin_name not defined")
	}

	name := args[0]
	if name == "" {
		cmdutil.ShowSubcommandHelp(cmd)
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
