package cli

import (
	"os"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/log"
	"github.com/urfave/cli"
)

var setupCommand = cli.Command{
	Name:  "setup",
	Usage: "Setup the current host. Install every required tool to run Workflows.",
	Action: func(c *cli.Context) error {
		if err := setup(c); err != nil {
			log.Errorf("Setup failed, error: %s", err)
			os.Exit(1)
		}
		return nil
	},
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "clean",
			Usage: "Removes bitrise's workdir before setup.",
		},
		cli.BoolFlag{
			Name:  "minimal",
			Usage: "Only installs the required tools for running in CI mode.",
		},
	},
}

func setup(c *cli.Context) error {
	clean := c.Bool("clean")
	minimal := c.Bool("minimal")

	setupMode := bitrise.SetupModeDefault
	if minimal {
		setupMode = bitrise.SetupModeMinimal
	}

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	if err := bitrise.RunSetup(logger, c.App.Version, setupMode, clean); err != nil {
		return err
	}

	logger.Print()
	logger.Infof("To start using bitrise:")
	logger.Printf("* cd into your project's directory (if you're not there already)")
	logger.Printf("* call: bitrise init")
	logger.Printf("* follow the guide")
	logger.Print()
	logger.Donef("That's all :)")

	return nil
}
