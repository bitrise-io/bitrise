package cli

import "github.com/codegangsta/cli"

var (
	commands = []cli.Command{
		{
			Name:    "setup",
			Aliases: []string{"s"},
			Usage:   "Setup the current host. Install every required tool to run Workflows.",
			Action:  setup,
			Flags: []cli.Flag{
				flMinimalSetup,
			},
		},
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Generates a Workflow/app config file in the current directory, which then can be run immediately.",
			Action:  initConfig,
		},
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "Runs a specified Workflow.",
			Action:  run,
			Flags: []cli.Flag{
				flPath,
				flInventory,
			},
		},
		{
			Name:    "trigger",
			Aliases: []string{"t"},
			Usage:   "Triggers a specified Workflow.",
			Action:  trigger,
			Flags: []cli.Flag{
				flPath,
				flInventory,
			},
		},
		{
			Name:   "export",
			Usage:  "Export the bitrise configuration.",
			Action: export,
			Flags: []cli.Flag{
				flPath,
				flFormat,
				flOutputPath,
				flPretty,
			},
		},
		{
			Name:   "normalize",
			Usage:  "Normalize the bitrise configuration.",
			Action: normalize,
			Flags: []cli.Flag{
				flPath,
			},
		},
	}
)
