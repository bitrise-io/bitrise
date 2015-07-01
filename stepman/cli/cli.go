package cli

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func before(c *cli.Context) error {
	err := stepman.CreateStepManDirIfNeeded()
	if err != nil {
		return err
	}

	// StepSpec collection path
	stepman.CollectionPath = c.String(COLLECTION_KEY)
	if stepman.CollectionPath == "" {
		stepman.CollectionPath = os.Getenv(COLLECTION_PATH_ENV_KEY)
	}
	// TODO! remove default path
	if stepman.CollectionPath == "" {
		stepman.CollectionPath = stepman.STEP_COLLECTION_GIT
	}

	// Debug mode
	debugString := c.String(DEBUG_KEY)
	if debugString == "" {
		debugString = os.Getenv(DEBUG_ENV_KEY)
	}
	if debugString == "" {
		debugString = "false"
	}
	stepman.DebugMode, err = strconv.ParseBool(debugString)
	if err != nil {
		fmt.Println("Failed to parse debug mode flag:", err)
		stepman.DebugMode = false
	}

	return nil
}

func Run() {
	// Parse cl
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Step manager"
	app.Version = "0.0.1"

	app.Author = ""
	app.Email = ""

	app.Before = before

	app.Flags = flags
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		fmt.Println("Stepman finished:", err)
	}
}
