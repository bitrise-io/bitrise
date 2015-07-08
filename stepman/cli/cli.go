package cli

import (
	"os"
	"path"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func before(c *cli.Context) error {
	err := stepman.CreateStepManDirIfNeeded()
	if err != nil {
		return err
	}

	// StepSpec collection path
	stepman.CollectionURI = c.String(CollectionKey)
	if stepman.CollectionURI == "" {
		stepman.CollectionURI = os.Getenv(CollectionPathEnvKey)
	}
	if stepman.CollectionURI == "" {
		log.Fatalln("[STEPMAN] - No step collection specified")
	}

	// Debug mode
	debugString := c.String(DebugKey)
	if debugString == "" {
		debugString = os.Getenv(DebugEnvKey)
	}
	if debugString == "" {
		debugString = "false"
	}

	if stepman.DebugMode, err = strconv.ParseBool(debugString); err != nil {
		log.Error("[STEPMAN] - Failed to parse debug mode flag:", err)
		stepman.DebugMode = false
	}

	return nil
}

// Run ...
func Run() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Step manager"
	app.Version = "0.0.2"

	app.Author = ""
	app.Email = ""

	app.Before = before

	app.Flags = flags
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		log.Error("[STEPMAN] - Stepman finished:", err)
	}
}
