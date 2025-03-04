package cli

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/bitrise-io/envman/v2/envman"
	"github.com/bitrise-io/envman/v2/version"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	defaultEnvStoreName = ".envstore.yml"

	helpTemplate = `
NAME: {{.Name}} - {{.Usage}}

USAGE: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]

VERSION: {{.Version}}{{if or .Author .Email}}

AUTHOR:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}
{{if .Flags}}
GLOBAL OPTIONS:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
COMMANDS:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
COMMAND HELP: {{.Name}} COMMAND --help/-h

`
)

func before(c *cli.Context) error {
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: "15:04:05"})

	// Log level
	logLevel, err := log.ParseLevel(c.String(LogLevelKey))
	if err != nil {
		log.Fatal("[BITRISE_CLI] - Failed to parse log level:", err)
	}
	log.SetLevel(logLevel)

	// Before parsing cli, and running command
	// we need to decide which path will be used by envman
	CurrentEnvStoreFilePath = c.String(PathKey)
	if CurrentEnvStoreFilePath == "" {
		pth, err := filepath.Abs(path.Join("./", defaultEnvStoreName))
		if err != nil {
			log.Fatal("[ENVMAN] - Failed to set envman work path in current dir:", err)
		}
		CurrentEnvStoreFilePath = pth
	}

	ToolMode = c.Bool(ToolKey)
	if ToolMode {
		log.Info("[ENVMAN] - Tool mode on")
	}

	if _, err := envman.GetConfigs(); err != nil {
		log.Fatal("[ENVMAN] - Failed to init configs:", err)
	}

	return nil
}

// Run the Envman CLI.
func Run() {
	cli.HelpFlag = cli.BoolFlag{Name: HelpKey + ", " + helpKeyShort, Usage: "Show help."}
	cli.AppHelpTemplate = helpTemplate

	cli.VersionFlag = cli.BoolFlag{Name: VersionKey + ", " + versionKeyShort, Usage: "Print the version."}
	cli.VersionPrinter = func(c *cli.Context) { fmt.Println(c.App.Version) }

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Environment variable manager"
	app.Version = version.Version

	app.Author = ""
	app.Email = ""

	app.Before = before

	app.Flags = flags
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		log.Fatal("[ENVMAN] - Finished:", err)
	}
}
