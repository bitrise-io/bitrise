package cli

import (
	_ "errors"
	"fmt"
	"os"
	"path"

	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func before(c *cli.Context) error {
	err := stepman.CreateStepManDirIfNeeded()
	if err != nil {
		return err
	}

	stepman.CollectionPath = c.String(COLLECTION_KEY)
	if stepman.CollectionPath == "" {
		if os.Getenv(COLLECTION_PATH_ENV_KEY) != "" {
			stepman.CollectionPath = os.Getenv(COLLECTION_PATH_ENV_KEY)
		}
		// TODO! remove default path
		if stepman.CollectionPath == "" {
			stepman.CollectionPath = stepman.STEP_COLLECTION_GIT
		}
		//return errors.New("No collection path specified")
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
