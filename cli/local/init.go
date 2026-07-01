package local

import (
	"fmt"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/spf13/cobra"
)

// NewInitCommand ...
func NewInitCommand() *cobra.Command {
	initCommand := &cobra.Command{
		Use:     "init",
		Aliases: []string{"i"},
		Short:   "Init bitrise config.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			logger := log.NewLogger(log.GetGlobalLoggerOpts())
			if err := initConfig(cmd); err != nil {

				// If the plugin is not installed yet run the bitrise setup first and try it again
				perr, ok := err.(plugins.NotInstalledError)
				if ok {
					logger.Warn(perr)
					logger.Print("Running setup to install the default plugins")
					logger.Print()

					if err := bitrise.RunSetup(logger, version.VERSION, bitrise.SetupModeDefault, false, false); err != nil {
						return fmt.Errorf("setup failed, error: %s", err)
					}

				if err := initConfig(cmd); err != nil {
					failf("%s", err)
				}
			} else {
				failf("%s", err)
			}
		}
		return nil
	},
}

	initCommand.Flags().Bool("minimal", false, "creates empty bitrise config and secrets")

	return initCommand
}

func initConfig(cmd *cobra.Command) error {
	minimal, _ := cmd.Flags().GetBool("minimal")

	pluginName := "init"
	plugin, found, err := plugins.LoadPlugin(pluginName)
	if err != nil {
		return fmt.Errorf("failed to get plugin (%s), error: %s", pluginName, err)
	}

	if !found {
		return plugins.NewNotInstalledError("init")
	}

	pluginArgs := []string{}
	if minimal {
		pluginArgs = []string{"--minimal"}
	}
	if err := plugins.RunPluginByCommand(plugin, pluginArgs); err != nil {
		return fmt.Errorf("failed to run plugin (%s), error: %s", pluginName, err)
	}

	return nil
}
