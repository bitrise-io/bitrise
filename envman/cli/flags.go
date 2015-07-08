package cli

import "github.com/codegangsta/cli"

const (
	// PathKey ...
	PathKey      string = "path"
	pathKeyShort string = "p"

	// LogLevelKey ...
	LogLevelKey      string = "log-level"
	logLevelKeyShort string = "l"

	// KeyKey ...
	KeyKey       string = "key"
	keyKeyShortT string = "k"

	// ValueKey ...
	ValueKey      string = "value"
	valueKeyShort string = "v"

	// ValueFileKey ...
	ValueFileKey      string = "valuefile"
	valueFileKeyShort string = "f"

	// ExpandKey ...
	ExpandKey      string = "expand"
	expandKeyShort string = "e"

	// HelpKey ...
	HelpKey      string = "help"
	helpKeyShort string = "h"

	// VersionKey ...
	VersionKey      string = "version"
	versionKeyShort string = "v"
)

var (
	// App flags
	flPath = cli.StringFlag{
		Name:   PathKey + ", " + pathKeyShort,
		EnvVar: "ENVMAN_ENVSTORE_PATH",
		Value:  "",
		Usage:  "Path of the envstore.",
	}
	flLogLevel = cli.StringFlag{
		Name:  LogLevelKey + ", " + logLevelKeyShort,
		Value: "info",
		Usage: "Log level (options: debug, info, warn, error, fatal, panic).",
	}
	flags = []cli.Flag{
		flPath,
		flLogLevel,
	}

	// Command flags
	flKey = cli.StringFlag{
		Name:  KeyKey + ", " + keyKeyShortT,
		Value: "",
		Usage: "Key of the environment variable. Empty string (\"\") is NOT accepted.",
	}
	flValue = cli.StringFlag{
		Name:  ValueKey + ", " + valueKeyShort,
		Value: "",
		Usage: "Value of the environment variable. Empty string is accepted.",
	}
	flValueFile = cli.StringFlag{
		Name:  ValueFileKey + ", " + valueFileKeyShort,
		Value: "",
		Usage: "Path of a file which contains the environment variable's value to be stored.",
	}
	flIsExpand = cli.StringFlag{
		Name:  ExpandKey + ", " + expandKeyShort,
		Value: "true",
		Usage: "If true, replaces ${var} or $var in the string according to the values of the current environment variables.",
	}
)

func init() {
	// Override default help and version flags
	cli.HelpFlag = cli.BoolFlag{
		Name:  HelpKey + ", " + helpKeyShort,
		Usage: "Show help.",
	}

	cli.VersionFlag = cli.BoolFlag{
		Name:  VersionKey + ", " + versionKeyShort,
		Usage: "Print the version.",
	}
}
