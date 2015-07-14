package cli

import (
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func parseLogLevelString(c *cli.Context) string {
	if c.IsSet(LogLevelKey) {
		return c.String(LogLevelKey)
	}
	return log.DebugLevel.String()
}

func parseLogLevel(c *cli.Context) (log.Level, error) {
	return log.ParseLevel(c.String(LogLevelKey))
}

func before(c *cli.Context) error {
	// Log level
	if err := os.Setenv(LogLevelEnvKey, parseLogLevelString(c)); err != nil {
		log.Fatal("Faild to set log level env:", err)
	}

	if logLevel, err := log.ParseLevel(parseLogLevelString(c)); err != nil {
		log.Fatal("[BITRISE_CLI] - Failed to parse log level:", err)
	} else {
		log.SetLevel(logLevel)
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
