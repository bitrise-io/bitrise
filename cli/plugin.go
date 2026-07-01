package cli

import (
	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/cli/legacy"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/spf13/cobra"
)

var pluginCommand = &cobra.Command{
	Use:   "plugin",
	Short: "Plugin handling.",
	RunE:  requireKnownSubcommand,
}

func init() {
	pluginCommand.AddCommand(
		pluginInstallCommand,
		pluginUpdateCommand,
		pluginDeleteCommand,
		pluginInfoCommand,
		pluginListCommand,
	)
}

func showSubcommandHelp(cmd *cobra.Command) {
	if err := cmd.Help(); err != nil {
		log.Warnf("Failed to show help, error: %s", err)
	}
}

// detectPlugin decides plugin dispatch: it only happens when the first
// non-global-flag token is not a known command, so e.g. `bitrise run a:b` stays
// a run invocation rather than being treated as a plugin.
func detectPlugin(root *cobra.Command, rawArgs []string) (string, []string, bool) {
	i := legacy.CommandTokenIndex(rawArgs, globalFlagNames)
	if i == len(rawArgs) {
		return "", nil, false
	}
	if legacy.IsKnownCommand(root, rawArgs[i]) {
		return "", nil, false
	}
	// Pass the args from the command token onward (not globals-stripped) so that
	// flags following the plugin name — including ones that share a global flag's
	// name, e.g. the plugin's own --debug — are forwarded to the plugin verbatim.
	return plugins.ParseArgs(rawArgs[i:])
}

func runPlugin(root *cobra.Command, rawArgs []string, pluginName string, pluginArgs []string) {
	logger := log.NewLogger(log.GetGlobalLoggerOpts())

	legacy.ApplyGlobalFlagsFromArgs(root, rawArgs, globalFlagNames)
	if err := before(root, nil); err != nil {
		failf("%s", err)
	}

	logPluginCommandParameters(pluginName, pluginArgs)

	plugin, found, err := plugins.LoadPlugin(pluginName)
	if err != nil {
		failf("failed to get plugin (%s), error: %s", pluginName, err)
	}
	if !found {
		failf("plugin (%s) not installed", pluginName)
	}

	if err := bitrise.RunSetupIfNeeded(logger); err != nil {
		failf("Setup failed, error: %s", err)
	}

	if err := plugins.RunPluginByCommand(plugin, pluginArgs); err != nil {
		failf("failed to run plugin (%s), error: %s", pluginName, err)
	}
}
