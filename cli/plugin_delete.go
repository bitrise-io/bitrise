package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/codegangsta/cli"
)

func pluginDelete(c *cli.Context) {
	// Input validation
	printableName := c.Args()[0]
	if printableName == "" {
		log.Fatalf("Missing plugin name")
	}

	pluginName, pluginType, err := plugins.ParsePrintableName(printableName)
	if err != nil {
		log.Fatalf("Failed to parse plugin name, err: %s", err)
	}

	// Delete
	log.Infof("=> Deleting plugin (%s) ...", printableName)
	if err := plugins.DeletePlugin(pluginName, pluginType); err != nil {
		log.Fatalf("Failed to delete plugin, err: %s", err)
	}
	log.Infof("Done")
}
