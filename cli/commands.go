package cli

import "github.com/codegangsta/cli"

var (
	commands = []cli.Command{
		{
			Name:    "setup",
			Aliases: []string{"s"},
			Usage:   "Setup the current host. Install every required tool to run Workflows.",
			Action:  doSetup,
			Flags: []cli.Flag{
				flMinimalSetup,
			},
		},
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Generates a Workflow/app config file in the current directory, which then can be run immediately.",
			Action:  doInit,
		},
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "Runs a specified Workflow.",
			Action:  doRun,
			Flags: []cli.Flag{
				flPath,
				flInventory,
			},
		},
	}
)
