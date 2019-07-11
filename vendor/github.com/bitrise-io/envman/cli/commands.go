package cli

import "github.com/urfave/cli"

var (
	commands = []cli.Command{
		{
			Name:   "version",
			Usage:  "Prints the version",
			Action: printVersionCmd,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "format",
					Usage: "Output format. Accepted: json, yml",
				},
				cli.BoolFlag{
					Name:  "full",
					Usage: "Prints the build number and commit as well.",
				},
			},
		},
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Create an empty .envstore.yml into the current working directory, or to the path specified by the --path flag.",
			Action:  initEnvStore,
			Flags: []cli.Flag{
				flClear,
			},
		},
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "Add new, or update an exist environment variable.",
			Action:  add,
			Flags: []cli.Flag{
				flKey,
				flValue,
				flValueFile,
				flNoExpand,
				flAppend,
				cli.BoolFlag{
					Name:  SkipIfEmptyKey,
					Usage: "If enabled the added environment variable will be skipped during envman run, if the value is empty. If not set then the empty value will be used.",
				},
			},
		},
		{
			Name:    "clear",
			Aliases: []string{"c"},
			Usage:   "Clear the envstore.",
			Action:  clear,
		},
		{
			Name:    "print",
			Aliases: []string{"p"},
			Usage:   "Print out the environment variables in envstore.",
			Action:  print,
			Flags: []cli.Flag{
				flFormat,
				flExpand,
			},
		},
		{
			Name:            "run",
			Aliases:         []string{"r"},
			Usage:           "Run the specified command with the environment variables stored in the envstore.",
			SkipFlagParsing: true,
			Action:          run,
		},
		{
			Name:    "unset",
			Aliases: []string{"rm"},
			Usage:   "Enlist an environment variable to be unset (for example to clear OS inherited vars for the process).",
			Action:  unset,
			Flags: []cli.Flag{
				flKey,
			},
		},
	}
)
