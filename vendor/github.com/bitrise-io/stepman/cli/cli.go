package cli

import (
	"os"
	"path"

	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/bitrise-io/stepman/version"
	"github.com/urfave/cli"
)

func before(c *cli.Context) error {
	initHelpAndVersionFlags()
	initAppHelpTemplate()

	if c.String(LogLevelKey) == "debug" {
		log.SetEnableDebugLog(true)
	}

	// Setup
	err := stepman.CreateStepManDirIfNeeded()
	if err != nil {
		return err
	}

	return nil
}

func printVersion(c *cli.Context) {
	log.Println(c.App.Version)
}

// Run ...
func Run() {
	cli.VersionPrinter = printVersion

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Step manager"
	app.Version = version.Version

	app.Author = ""
	app.Email = ""

	app.Before = before

	app.Flags = flags
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		fail(err)
	}
}
