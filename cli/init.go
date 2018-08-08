package cli

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/plugins"
	"github.com/bitrise-io/bitrise/version"
	"github.com/bitrise-io/go-utils/log"
	"github.com/urfave/cli"
)

var initCmd = cli.Command{
	Name:    "init",
	Aliases: []string{"i"},
	Usage:   "Init bitrise config.",
	Action: func(c *cli.Context) error {
		if err := initConfig(c); err != nil {

			// If the plugin is not installed yet run the bitrise setup first and try it again
			perr, ok := err.(PluginError)
			if ok {
				log.Warnf(perr.Error())
				log.Printf("Runing setup to install the default plugins")
				fmt.Println()

				if err := bitrise.RunSetup(version.VERSION, false, false); err != nil {
					return fmt.Errorf("Setup failed, error: %s", err)
				}

				if err := initConfig(c); err != nil {
					logrus.Fatal(err)
				}
			} else {
				logrus.Fatal(err)
			}
		}
		return nil
	},
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "minimal",
			Usage: "creates empty bitrise config and secrets",
		},
	},
}

func initConfig(c *cli.Context) error {
	minimal := c.Bool("minimal")

	pluginName := "init"
	plugin, found, err := plugins.LoadPlugin(pluginName)
	if err != nil {
		return fmt.Errorf("Failed to get plugin (%s), error: %s", pluginName, err)
	}

	if !found {
		return NewPluginError("Default plugin (init) NOT found.")
	}

	pluginArgs := []string{}
	if minimal {
		pluginArgs = []string{"--minimal"}
	}
	if err := plugins.RunPluginByCommand(plugin, pluginArgs); err != nil {
		return fmt.Errorf("Failed to run plugin (%s), error: %s", pluginName, err)
	}

	return nil
}
