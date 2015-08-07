package cli

import "github.com/codegangsta/cli"

var (
	commands = []cli.Command{
		{
			Name:      "setup",
			ShortName: "s",
			Usage:     "Initialize the specified collection, it's required before using a collection.",
			Action:    setup,
			Flags: []cli.Flag{
				flCollection,
				flLocalCollection,
				flCopySpecJSON,
			},
		},
		{
			Name:      "update",
			ShortName: "u",
			Usage:     "Update the collection, if no --collection flag provided, all collections will updated.",
			Action:    update,
			Flags: []cli.Flag{
				flCollection,
			},
		},
		{
			Name:      "download",
			ShortName: "d",
			Usage:     "Download the step with provided --id and --version, from specified --collection, into local step downloads cache. If no --version defined, the latest version of the step (latest found in the collection) will be downloaded into the cache.",
			Action:    download,
			Flags: []cli.Flag{
				flCollection,
				flID,
				flVersion,
				flUpdate,
			},
		},
		{
			Name:      "activate",
			ShortName: "a",
			Usage:     "Copy the step with specified --id, and --version, into provided path. If --version flag is not set, the latest version of the step will be used. If --copyyml flag is set, step.yml will be copied to the given path.",
			Action:    activate,
			Flags: []cli.Flag{
				flCollection,
				flID,
				flVersion,
				flPath,
				flCopyYML,
				flUpdate,
			},
		},
		{
			Name:    "share",
			Aliases: []string{"s"},
			Usage:   "Coming soon.",
			Action:  share,
			Subcommands: []cli.Command{
				{
					Name:    "start",
					Aliases: []string{"s"},
					Usage:   "Coming soon.",
					Action:  start,
					Flags: []cli.Flag{
						flCollection,
					},
				},
				{
					Name:    "create",
					Aliases: []string{"c"},
					Usage:   "Coming soon.",
					Action:  create,
					Flags: []cli.Flag{
						flTag,
						flURL,
					},
				},
				{
					Name:    "finish",
					Aliases: []string{"f"},
					Usage:   "Coming soon.",
					Action:  finish,
					Flags: []cli.Flag{
						flCollection,
					},
				},
			},
		},
	}
)
