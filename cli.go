package main

import (
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

// Run the Envman CLI.
func run() {
	log.SetLevel(log.DebugLevel)

	// Parse cl
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Bitrise Automations Workflow Runner"
	app.Version = VERSION

	app.Author = ""
	app.Email = ""

	app.Before = func(c *cli.Context) error {
		level, err := log.ParseLevel(c.String("loglevel"))
		if err != nil {
			log.Fatal(err.Error())
		}
		log.SetLevel(level)

		return nil
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "loglevel" + ", " + "l",
			Value: "info",
			Usage: "Log level (options: debug, info, warn, error, fatal, panic).",
		},
		cli.BoolFlag{
			Name:  "tool" + ", " + "ci" + ", " + "quiet" + ", " + "q",
			Usage: "If true it indicates that we're used by another tool so don't require any user input!",
			EnvVar: "CI",
		},
	}
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		log.Fatal("Finished with Error:", err)
	}
}
