package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/codegangsta/cli"
)

func pluginInstall(c *cli.Context) {
	// Input validation
	pluginSource := c.String("source")
	if pluginSource == "" {
		log.Fatalln("Missing required input: source")
	}

	pluginName := c.String("name")
	if pluginName == "" {
		log.Fatalf("Missing required input: name")
	}

	pluginType := c.String("type")
	if pluginType == "" {
		pluginType = plugins.TypeGeneric
	}

	// Install
	log.Infof("=> Installing plugin (%s)...", plugins.PrintableName(pluginName, pluginType))
	printableName, err := plugins.InstallPlugin(c.App.Version, pluginSource, pluginName, pluginType)
	if err != nil {
		log.Fatalf("Failed to install plugin, err: %s", err)
	}
	fmt.Println()
	log.Infoln(colorstring.Greenf("Plugin (%s) installed", printableName))
}
