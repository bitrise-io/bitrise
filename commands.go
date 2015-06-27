package main

import "github.com/codegangsta/cli"

var (
	commands = []cli.Command{
		{
			Name:      "setup",
			ShortName: "s",
			Usage:     "Setup the current host. Install every required tool to run Workflows.",
			Action:    setupCmd,
		},
	}
)
