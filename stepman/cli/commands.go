package cli

import "github.com/codegangsta/cli"

var (
	commands = []cli.Command{
		{
			Name:      "setup",
			ShortName: "s",
			Usage:     "",
			Action:    setup,
			Flags: []cli.Flag{
				flCollection,
			},
		},
		{
			Name:      "update",
			ShortName: "u",
			Usage:     "",
			Action:    update,
			Flags: []cli.Flag{
				flCollection,
			},
		},
		{
			Name:      "download",
			ShortName: "d",
			Usage:     "",
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
			Usage:     "",
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
