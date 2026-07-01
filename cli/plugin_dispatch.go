package cli

import (
	"slices"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/plugins"
	"github.com/spf13/cobra"
)

// detectPlugin decides plugin dispatch: it only happens when the first
// non-global-flag token is not a known command (or "help"), so e.g.
// `bitrise run a:b` stays a run invocation rather than being treated as a plugin.
func detectPlugin(root *cobra.Command, rawArgs []string) (string, []string, bool) {
	i := cmdutil.CommandTokenIndex(rawArgs, cmdutil.GlobalFlagNames)
	if i == len(rawArgs) {
		return "", nil, false
	}

	token := rawArgs[i]
	if token == "help" {
		return "", nil, false
	}
	for _, c := range root.Commands() {
		if c.Name() == token || slices.Contains(c.Aliases, token) {
			return "", nil, false
		}
	}

	// Pass the args from the command token onward (not globals-stripped) so that
	// flags following the plugin name — including ones that share a global flag's
	// name, e.g. the plugin's own --debug — are forwarded to the plugin verbatim.
	return plugins.ParseArgs(rawArgs[i:])
}

func runPlugin(root *cobra.Command, rawArgs []string, pluginName string, pluginArgs []string) {
	logger := log.NewLogger(log.GetGlobalLoggerOpts())

	cmdutil.ApplyGlobalFlagsFromArgs(root, rawArgs, cmdutil.GlobalFlagNames)
	if err := before(root, nil); err != nil {
		cmdutil.Failf("%s", err)
	}

	cmdutil.LogPluginCommandParameters(pluginName, pluginArgs)

	plugin, found, err := plugins.LoadPlugin(pluginName)
	if err != nil {
		cmdutil.Failf("failed to get plugin (%s), error: %s", pluginName, err)
	}
	if !found {
		cmdutil.Failf("plugin (%s) not installed", pluginName)
	}

	if err := bitrise.RunSetupIfNeeded(logger); err != nil {
		cmdutil.Failf("Setup failed, error: %s", err)
	}

	if err := plugins.RunPluginByCommand(plugin, pluginArgs); err != nil {
		cmdutil.Failf("failed to run plugin (%s), error: %s", pluginName, err)
	}
}
