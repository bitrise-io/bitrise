package cli

import (
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/output"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/spf13/cobra"
)

var pluginListCommand = &cobra.Command{
	Use:   "list",
	Short: "List installed bitrise plugins.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		logCommandParameters(cmd)

		if err := pluginList(cmd); err != nil {
			log.Errorf("Plugin list failed, error: %s", err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	pluginListCommand.Flags().String(output.FormatKey, "", "Output format. Accepted: raw, json.")
}

func pluginList(cmd *cobra.Command) error {
	logger, err := loggerForFormat(cmd)
	if err != nil {
		return err
	}

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
