package cli

import (
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func before(c *cli.Context) error {
	levelString := c.String(LOG_LEVEL_KEY)
	if levelString == "" {
		log.SetLevel(log.DebugLevel)
	} else {
		if level, err := log.ParseLevel(levelString); err != nil {
			return err
		} else {
			log.SetLevel(level)
		}
	}
	return nil
}

func Run() {
	// Parse cl
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Bitrise Automations Workflow Runner"
	app.Version = "0.0.1"

	app.Author = ""
	app.Email = ""

	app.Before = before

	app.Flags = flags
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		log.Fatal("Finished with Error:", err)
	}
}
