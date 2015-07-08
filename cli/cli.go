package cli

import (
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func before(c *cli.Context) error {
	levelString := c.String(LogLevelKey)
	if levelString == "" {
		log.SetLevel(log.DebugLevel)
	} else {
		level, err := log.ParseLevel(levelString)
		if err != nil {
			return err
		}
		log.SetLevel(level)
	}
	return nil
}

// Run ...
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
