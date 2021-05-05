package cli

import (
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/bitrise-io/stepman/version"
	"github.com/urfave/cli"
)

func initLogFormatter() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})
}

func before(c *cli.Context) error {
	initLogFormatter()
	initHelpAndVersionFlags()
	initAppHelpTemplate()

	// Log level
	logLevel, err := log.ParseLevel(c.String(LogLevelKey))
	if err != nil {
		return fmt.Errorf("Failed to parse log level, error: %s", err)
	}
	log.SetLevel(logLevel)

	// Setup
	err = stepman.CreateStepManDirIfNeeded()
	if err != nil {
		return err
	}

	return nil
}

func printVersion(c *cli.Context) {
	fmt.Println(c.App.Version)
}

// Run ...
func Run() {
	cli.VersionPrinter = printVersion

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Step manager"
	app.Version = version.VERSION

	app.Author = ""
	app.Email = ""

	app.Before = before

	app.Flags = flags
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
