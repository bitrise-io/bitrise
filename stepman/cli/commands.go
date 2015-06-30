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
			/*
				Flags: []cli.Flag{
					flKey,
					flValue,
					flValueFile,
					flIsExpand,
				},
			*/
			Action: update,
		},
		{
			Name:      "download",
			ShortName: "d",
			Usage:     "",
			Action:    download,
		},
		{
			Name:      "activate",
			ShortName: "a",
			Usage:     "",
			Action:    activate,
		},
	}
)
