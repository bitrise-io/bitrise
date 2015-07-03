package cli

import "github.com/codegangsta/cli"

var (
	commands = []cli.Command{
		{
			Name:      "setup",
			ShortName: "s",
			Usage:     "Setup the current host. Install every required tool to run Workflows.",
			Action:    doSetup,
		},
		{
			Name:      "init",
			ShortName: "i",
			Usage:     "Generates a Workflow/app config file in the current directory, which then can be run immediately.",
			Action:    doInit,
		},
		{
			Name:      "run",
			ShortName: "r",
			Usage:     "Runs a specified Workflow",
			Action:    doRun,
			Flags: []cli.Flag{
				flPath,
			},
		},
	}
)
