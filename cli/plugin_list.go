package cli

import (
	"fmt"
	"sort"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/codegangsta/cli"
)

func pluginList(c *cli.Context) {
	pluginMap, err := plugins.ListPlugins()
	if err != nil {
		log.Fatalf("Failed to list plugins, err: %s", err)
	}

	pluginNames := []string{}
	for _, plugins := range pluginMap {
		for _, plugin := range plugins {
			pluginNames = append(pluginNames, plugin.PrintableName())
		}
	}
	sort.Strings(pluginNames)

	if len(pluginNames) > 0 {
		fmt.Println("")
		for _, name := range pluginNames {
			fmt.Printf(" ⚡️ %s\n", colorstring.Green(name))
		}
		fmt.Println("")
	} else {
		fmt.Println("")
		fmt.Println("No installed plugin found")
		fmt.Println("")
	}
}
