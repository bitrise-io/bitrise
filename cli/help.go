package cli

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise/v2/plugins"
)

func getPluginsList() string {
	pluginListString := "PLUGINS:\n"

	pluginList, err := plugins.InstalledPluginList()
	if err != nil {
		failf("Failed to list plugins, error: %s", err)
	}

	if len(pluginList) > 0 {
		plugins.SortByName(pluginList)
		for _, plugin := range pluginList {
			pluginListString += fmt.Sprintf("  :%s\t%s\n", plugin.Name, strings.Split(plugin.Description, "\n")[0])
		}
	} else {
		pluginListString += "  No plugins installed\n"
	}

	return pluginListString
}
