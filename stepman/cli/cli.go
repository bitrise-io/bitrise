package cli

import (
	_ "errors"
	"fmt"
	"os"
	"path"

	"github.com/bitrise-io/stepman/paths"
	"github.com/codegangsta/cli"
)

func before(c *cli.Context) error {
	err := paths.CreateStepManDirIfNeeded()
	if err != nil {
		return err
	}

	paths.CollectionPath = c.String(COLLECTION_KEY)
	if paths.CollectionPath == "" {
		if os.Getenv(COLLECTION_PATH_ENV_KEY) != "" {
			paths.CollectionPath = os.Getenv(COLLECTION_PATH_ENV_KEY)
		}
		// TODO! remove default path
		if paths.CollectionPath == "" {
			paths.CollectionPath = paths.STEP_COLLECTION_GIT
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
