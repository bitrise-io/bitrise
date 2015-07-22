package cli

import "github.com/codegangsta/cli"

var (
	commands = []cli.Command{
		{
			Name:      "setup",
			ShortName: "s",
			Usage:     "Setup initialize the specified collection, it's required befor using a collection.",
			Action:    setup,
			Flags: []cli.Flag{
				flCollection,
			},
		},
		{
			Name:      "update",
			ShortName: "u",
			Usage:     "Updates the collection, if no --collection flag provided, all collections will updated.",
			Action:    update,
			Flags: []cli.Flag{
				flCollection,
			},
		},
		{
			Name:      "download",
			ShortName: "d",
			Usage:     "Downloads the step wit provided --id, and --version, from specified --collection into local step cache. If no --version defined, the latest version of step will be cached.",
			Action:    download,
			Flags: []cli.Flag{
				flCollection,
				flID,
				flVersion,
			},
		},
		{
			Name:      "activate",
			ShortName: "a",
			Usage:     "Copies the step with specified --id, and --version, into provided path. If --version flag is unset, the latest version of step will be used. If --copyyml flag is set, step.yml will copied into given path.",
			Action:    activate,
			Flags: []cli.Flag{
				flCollection,
				flID,
				flVersion,
				flPath,
				flCopyYML,
			},
		},
	}
)
