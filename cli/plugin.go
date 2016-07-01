package cli

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/codegangsta/cli"
)

func pluginInstall(c *cli.Context) error {
	// Input validation
	pluginSource := c.String("source")
	if pluginSource == "" {
		return errors.New("Missing required input: source")
	}

	pluginBinary := c.String("bin-source")
	pluginVersionTag := c.String("version")

	// Install
	if pluginVersionTag == "" {
		log.Infof("=> Installing plugin from (%s) with latest version...", pluginSource)
	} else {
		log.Infof("=> Installing plugin (%s) with version (%s)...", pluginSource, pluginVersionTag)
	}

	plugin, version, err := plugins.InstallPlugin(pluginSource, pluginBinary, pluginVersionTag)
	if err != nil {
		return fmt.Errorf("Failed to install plugin from (%s), error: %s", pluginSource, err)
	}

	fmt.Println()
	log.Infoln(colorstring.Greenf("Plugin (%s) with version (%s) installed ", plugin.Name, version))

	if len(plugin.Description) > 0 {
		fmt.Println()
		fmt.Println(plugin.Description)
		fmt.Println()
	}

	return nil
}

func pluginDelete(c *cli.Context) error {
	// Input validation
	if len(c.Args()) == 0 {
		return errors.New("Missing plugin name")
	}

	name := c.Args()[0]
	if name == "" {
		return errors.New("Missing plugin name")
	}
	if _, found, err := plugins.LoadPlugin(name); err != nil {
		return fmt.Errorf("Failed to check if plugin (%s) installed, error: %s", name, err)
	} else if !found {
		return fmt.Errorf("Plugin (%s) is not installed", name)
	}

	versionPtr, err := plugins.GetPluginVersion(name)
	if err != nil {
		return fmt.Errorf("Failed to read plugin (%s) version, error: %s", name, err)
	}

	// Delete
	version := "local"
	if versionPtr != nil {
		version = versionPtr.String()
	}
	log.Infof("=> Deleting plugin (%s) with version (%s) ...", name, version)
	if err := plugins.DeletePlugin(name); err != nil {
		return fmt.Errorf("Failed to delete plugin (%s) with version (%s), error: %s", name, version, err)
	}

	fmt.Println()
	log.Infof(colorstring.Greenf("Plugin (%s) with version (%s) deleted", name, version))

	return nil
}

func pluginList(c *cli.Context) error {
	pluginList, err := plugins.InstalledPluginList()
	if err != nil {
		return fmt.Errorf("Failed to list plugins, error: %s", err)
	}

	if len(pluginList) == 0 {
		log.Info("No installed plugin found")
	}

	plugins.SortByName(pluginList)

	fmt.Println()
	log.Infof("Installed plugins:")
	for _, plugin := range pluginList {
		fmt.Println()
		fmt.Println(plugin.String())
		fmt.Println()
	}

	return nil
}
