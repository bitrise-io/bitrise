package cli

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/plugins"
)

// printInstalledPlugins appends the installed plugin list to the root help.
// cobra's native help does not know about bitrise plugins (invoked via the colon
// syntax), so this preserves their discoverability from `bitrise --help`.
func printInstalledPlugins(w io.Writer) {
	pluginList, err := plugins.InstalledPluginList()
	if err != nil {
		cmdutil.Failf("Failed to list plugins, error: %s", err)
	}
	if len(pluginList) == 0 {
		return
	}
	if len(pluginList) == 0 {
		return
	}

	plugins.SortByName(pluginList)
	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	fmt.Fprintln(tw, "\nPlugins:")
	for _, plugin := range pluginList {
		fmt.Fprintf(tw, "  :%s\t%s\n", plugin.Name, strings.Split(plugin.Description, "\n")[0])
	}
	_ = tw.Flush()
}
