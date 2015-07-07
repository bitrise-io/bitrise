package cli

import "github.com/codegangsta/cli"

const (
	CI_ENV_KEY     string = "CI"
	TOOL_KEY       string = "tool"
	CI_KEY         string = "ci"
	QUIT_KEY       string = "quite"
	QUIT_KEY_SHORT string = "q"

	LOG_LEVEL_KEY       string = "loglevel"
	LOG_LEVEL_KEY_SHORT string = "l"

	ID_KEY       string = "id"
	ID_KEY_SHORT string = "i"

	VERSION_KEY       string = "version"
	VERSION_KEY_SHORT string = "v"

	PATH_KEY       string = "path"
	PATH_KEY_SHORT string = "p"

	HELP_KEY       string = "help"
	HELP_KEY_SHORT string = "h"
)

var (
	// App flags
	flLogLevel = cli.StringFlag{
		Name:  LOG_LEVEL_KEY + ", " + LOG_LEVEL_KEY_SHORT,
		Value: "info",
		Usage: "Log level (options: debug, info, warn, error, fatal, panic).",
	}
	flTool = cli.BoolFlag{
		Name:   TOOL_KEY + ", " + CI_KEY + ", " + QUIT_KEY + ", " + QUIT_KEY_SHORT,
		Usage:  "If true it indicates that we're used by another tool so don't require any user input!",
		EnvVar: CI_ENV_KEY,
	}
	flags = []cli.Flag{
		flLogLevel,
		flTool,
	}
	// Command flags
	flId = cli.StringFlag{
		Name:  ID_KEY + ", " + ID_KEY_SHORT,
		Value: "",
		Usage: "Step id.",
	}
	flVersion = cli.StringFlag{
		Name:  VERSION_KEY + ", " + VERSION_KEY_SHORT,
		Value: "",
		Usage: "Step version.",
	}
	flPath = cli.StringFlag{
		Name:  PATH_KEY + ", " + PATH_KEY_SHORT,
		Value: "",
		Usage: "Path where the step will copied.",
	}
)

func init() {
	// Override default help and version flags
	cli.HelpFlag = cli.BoolFlag{
		Name:  HELP_KEY + ", " + HELP_KEY_SHORT,
		Usage: "Show help.",
	}

	cli.VersionFlag = cli.BoolFlag{
		Name:  VERSION_KEY + ", " + VERSION_KEY_SHORT,
		Usage: "Print the version.",
	}
}
