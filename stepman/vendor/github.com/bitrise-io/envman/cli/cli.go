package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/envman/version"
	"github.com/urfave/cli"
)

const (
	defaultEnvStoreName string = ".envstore.yml"
)

var (
	stdinValue string
)

func isPipedData() bool {
	if stat, err := os.Stdin.Stat(); err != nil {
		return false
	} else if (stat.Mode() & os.ModeCharDevice) == 0 {
		return true
	}
	return false
}

func envStorePathInCurrentDir() (string, error) {
	return filepath.Abs(path.Join("./", defaultEnvStoreName))
}

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
	if logLevel, err := log.ParseLevel(c.String(LogLevelKey)); err != nil {
		log.Fatal("[BITRISE_CLI] - Failed to parse log level:", err)
	} else {
		log.SetLevel(logLevel)
	}

	// Befor parsing cli, and running command
	// we need to decide wich path will be used by envman
	envman.CurrentEnvStoreFilePath = c.String(PathKey)
	if envman.CurrentEnvStoreFilePath == "" {
		if path, err := envStorePathInCurrentDir(); err != nil {
			log.Fatal("[ENVMAN] - Failed to set envman work path in current dir:", err)
		} else {
			envman.CurrentEnvStoreFilePath = path
		}
	}

	envman.ToolMode = c.Bool(ToolKey)
	if envman.ToolMode {
		log.Info("[ENVMAN] - Tool mode on")
	}

	if _, err := envman.GetConfigs(); err != nil {
		log.Fatal("[ENVMAN] - Failed to init configs:", err)
	}

	return nil
}

func printVersion(c *cli.Context) {
	fmt.Fprintf(c.App.Writer, "%v\n", c.App.Version)
}

// Run the Envman CLI.
func Run() {
	// Read piped data
	if isPipedData() {
		if bytes, err := ioutil.ReadAll(os.Stdin); err != nil {
			log.Error("[ENVMAN] - Failed to read stdin:", err)
		} else if len(bytes) > 0 {
			stdinValue = string(bytes)
		}
	}

	// Parse cl
	cli.VersionPrinter = printVersion

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Environment variable manager"
	app.Version = version.VERSION

	app.Author = ""
	app.Email = ""

	app.Before = before

	app.Flags = flags
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		log.Fatal("[ENVMAN] - Finished:", err)
	}
}
