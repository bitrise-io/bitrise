package plugin

import (
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/spf13/cobra"
)

func newInstallCommand() *cobra.Command {
	pluginInstallCommand := &cobra.Command{
		Use:   "install <plugin_source_remote_or_local_url>",
		Short: "Install bitrise plugin.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := pluginInstall(cmd, args); err != nil {
				log.Errorf("Plugin install failed, error: %s", err)
				os.Exit(1)
			}
			return nil
		},
	}

	pluginInstallCommand.Flags().String("version", "", "Plugin version tag.")
	pluginInstallCommand.Flags().String("source", "", "Deprecated!!! Specify as arg instead - Plugin source url (can be local path or remote url).")

	return pluginInstallCommand
}

func pluginInstall(cmd *cobra.Command, args []string) error {
	// Input validation
	pluginSource := ""
	if len(args) > 0 {
		pluginSource = args[0]
	} else {
		pluginSource, _ = cmd.Flags().GetString("source")
	}

	pluginVersionTag, _ := cmd.Flags().GetString("version")

	if pluginSource == "" {
		cmdutil.ShowSubcommandHelp(cmd)
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
