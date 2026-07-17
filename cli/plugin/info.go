package plugin

import (
	"errors"
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/output"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/spf13/cobra"
)

func newInfoCommand() *cobra.Command {
	pluginInfoCommand := &cobra.Command{
		Use:   "info <plugin_name>",
		Short: "Installed bitrise plugin's info",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := pluginInfo(cmd, args); err != nil {
				log.Errorf("Plugin info failed, error: %s", err)
				os.Exit(1)
			}
			return nil
		},
	}

	pluginInfoCommand.Flags().String(output.FormatKey, "", "Output format. Accepted: raw, json.")

	return pluginInfoCommand
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

func loggerForFormat(cmd *cobra.Command) (cmdutil.Logger, error) {
	format, _ := cmd.Flags().GetString(output.FormatKey)
	if format == "" {
		format = output.FormatRaw
	}
	if format != output.FormatRaw && format != output.FormatJSON {
		cmdutil.ShowSubcommandHelp(cmd)
		return nil, fmt.Errorf("invalid format: %s", format)
	}
	if format == output.FormatJSON {
		return cmdutil.NewDefaultJSONLogger(), nil
	}
	return cmdutil.NewDefaultRawLogger(), nil
}

func pluginInfo(cmd *cobra.Command, args []string) error {
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

	logger, err := loggerForFormat(cmd)
	if err != nil {
		return err
	}

	pluginInfo, err := createPluginInfo(name)
	if err != nil {
		return err
	}

	logger.Print(pluginInfo)
	// ---

	return nil
}
