package cli

import "github.com/codegangsta/cli"

var (
	commands = []cli.Command{
		{
			Name:      "setup",
			ShortName: "s",
			Usage:     "",
			Action:    setup,
		},
		{
			Name:      "update",
			ShortName: "u",
			Usage:     "",
			Action:    update,
		},
		{
			Name:      "download",
			ShortName: "d",
			Usage:     "",
			Action:    download,
			Flags: []cli.Flag{
				flCollection,
				flId,
				flVersions,
			},
		},
		{
			Name:      "activate",
			ShortName: "a",
			Usage:     "",
			Action:    activate,
		},
	}
)
