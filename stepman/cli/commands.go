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
				flId,
				flVersion,
			},
		},
		{
			Name:      "activate",
			ShortName: "a",
			Usage:     "",
			Action:    activate,
			Flags: []cli.Flag{
				flId,
				flVersion,
				flPath,
			},
		},
	}
)
