package cli

import "github.com/codegangsta/cli"

const (
	PATH_KEY       string = "path"
	PATH_KEY_SHORT string = "p"

	LOG_LEVEL_KEY       string = "log-level"
	LOG_LEVEL_KEY_SHORT string = "l"

	KEY_KEY       string = "key"
	KEY_KEY_SHORT string = "k"

	VALUE_KEY       string = "value"
	VALUE_KEY_SHORT string = "v"

	VALUE_FILE_KEY       string = "valuefile"
	VALUE_FILE_KEY_SHORT string = "f"

	EXPAND_KEY       string = "expand"
	EXPAND_KEY_SHORT string = "e"

	HELP_KEY       string = "help"
	HELP_KEY_SHORT string = "h"

	VERSION_KEY       string = "version"
	VERSION_KEY_SHORT string = "v"
)

var (
	// App flags
	flPath = cli.StringFlag{
		Name:   PATH_KEY + ", " + PATH_KEY_SHORT,
		EnvVar: "ENVMAN_ENVSTORE_PATH",
		Value:  "",
		Usage:  "Path of the envstore.",
	}
	flLogLevel = cli.StringFlag{
		Name:  LOG_LEVEL_KEY + ", " + LOG_LEVEL_KEY_SHORT,
		Value: "info",
		Usage: "Log level (options: debug, info, warn, error, fatal, panic).",
	}
	flags = []cli.Flag{
		flPath,
		flLogLevel,
	}

	// Command flags
	flKey = cli.StringFlag{
		Name:  KEY_KEY + ", " + KEY_KEY_SHORT,
		Value: "",
		Usage: "Key of the environment variable. Empty string (\"\") is NOT accepted.",
	}
	flValue = cli.StringFlag{
		Name:  VALUE_KEY + ", " + VALUE_KEY_SHORT,
		Value: "",
		Usage: "Value of the environment variable. Empty string is accepted.",
	}
	flValueFile = cli.StringFlag{
		Name:  VALUE_FILE_KEY + ", " + VALUE_FILE_KEY_SHORT,
		Value: "",
		Usage: "Path of a file which contains the environment variable's value to be stored.",
	}
	flIsExpand = cli.StringFlag{
		Name:  EXPAND_KEY + ", " + EXPAND_KEY_SHORT,
		Value: "true",
		Usage: "If true, replaces ${var} or $var in the string according to the values of the current environment variables.",
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
