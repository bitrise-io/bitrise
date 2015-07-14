package cli

import (
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

func parseLogLevel(c *cli.Context) (log.Level, error) {
	if c.IsSet(LogLevelKey) {
		return log.ParseLevel(c.String(LogLevelKey))
	}
	return log.DebugLevel, nil
}

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
	stepman.DebugMode = parseDebug(c)

	// Log level
	if logLevel, err := parseLogLevel(c); err != nil {
		log.Fatal("[BITRISE_CLI] - Failed to parse log level:", err)
	} else {
		log.SetLevel(logLevel)
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
