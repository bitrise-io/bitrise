package cli

import "github.com/codegangsta/cli"

var (
	commands = []cli.Command{
		{
			Name:      "init",
			ShortName: "i",
			Usage:     "Create an empty .envstore.yml into the current working directory, or to the path specified by the --path flag.",
			Action:    initCmd,
		},
		{
			Name:      "add",
			ShortName: "a",
			Usage:     "Add new, or update an exist environment variable.",
			Flags: []cli.Flag{
				flKey,
				flValue,
				flValueFile,
				flIsExpand,
			},
			Action: addCmd,
		},
		{
			Name:      "clear",
			ShortName: "c",
			Usage:     "Clear the envstore.",
			Action:    clearCmd,
		},
		{
			Name:      "print",
			ShortName: "p",
			Usage:     "Print out the environment variables in envstore.",
			Action:    printCmd,
		},
		{
			Name:            "run",
			ShortName:       "r",
			Usage:           "Run the specified command with the environment variables stored in the envstore.",
			SkipFlagParsing: true,
			Action:          runCmd,
		},
	}
)
