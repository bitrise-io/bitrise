package cli

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/codegangsta/cli"
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

func parseTool(c *cli.Context) bool {
	if c.IsSet(ToolKey) {
		return c.Bool(ToolKey)
	}
	return false
}

func before(c *cli.Context) error {
	if level, err := log.ParseLevel(c.String("log-level")); err != nil {
		log.Fatal(err.Error())
	} else {
		log.SetLevel(level)
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

	envman.ToolMode = parseTool(c)
	if envman.ToolMode {
		log.Info("[ENVMAN] - Tool mode on")
	}

	return nil
}

// Run the Envman CLI.
func Run() {
	log.SetLevel(log.DebugLevel)

	// Read piped data
	if isPipedData() {
		if bytes, err := ioutil.ReadAll(os.Stdin); err != nil {
			log.Error("[ENVMAN] - Failed to read stdin:", err)
		} else if len(bytes) > 0 {
			stdinValue = string(bytes)
		}
	}

	// Parse cl
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Environment variable manager"
	app.Version = "0.0.7"

	app.Author = ""
	app.Email = ""

	app.Before = before

	app.Flags = flags
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		log.Fatal("[ENVMAN] - Finished:", err)
	}
}
