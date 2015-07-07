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
	stepman.CollectionUri = c.String(COLLECTION_KEY)
	if stepman.CollectionUri == "" {
		stepman.CollectionUri = os.Getenv(COLLECTION_PATH_ENV_KEY)
	}
	// TODO! remove default path
	if stepman.CollectionUri == "" {
		stepman.CollectionUri = stepman.OPEN_STEPLIB_GIT
	}

	// Debug mode
	debugString := c.String(DEBUG_KEY)
	if debugString == "" {
		debugString = os.Getenv(DEBUG_ENV_KEY)
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

func Run() {
	// Parse cl
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
