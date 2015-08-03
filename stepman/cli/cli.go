package cli

import (
	"fmt"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func parseDebug(c *cli.Context) bool {
	if c.IsSet(DebugKey) {
		return c.Bool(DebugKey)
	}
	return false
}

func before(c *cli.Context) error {
	// Log level
	if logLevel, err := log.ParseLevel(c.String(LogLevelKey)); err != nil {
		log.Fatal("[BITRISE_CLI] - Failed to parse log level:", err)
	} else {
		log.SetLevel(logLevel)
	}

	// Setup
	err := stepman.CreateStepManDirIfNeeded()
	if err != nil {
		return err
	}

	// Debug mode
	stepman.DebugMode = parseDebug(c)
	return nil
}

func printVersion(c *cli.Context) {
	fmt.Fprintf(c.App.Writer, "%v\n", c.App.Version)
}

// Run ...
func Run() {
	cli.VersionPrinter = printVersion

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Step manager"
	app.Version = "0.9.7"

	app.Author = ""
	app.Email = ""

	app.Before = before

	app.Flags = flags
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		log.Error("[STEPMAN] - Stepman finished:", err)
	}
}
