package cli

import "github.com/codegangsta/cli"

const (
	// PathKey ...
	PathKey      string = "path"
	pathKeyShort string = "p"

	// LogLevelEnvKey ...
	LogLevelEnvKey string = "LOGLEVEL"
	// LogLevelKey ...
	LogLevelKey      string = "loglevel"
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

	// NoExpandKey ...
	NoExpandKey      string = "no-expand"
	noExpandKeyShort string = "n"

	// AppendKey ...
	AppendKey      string = "append"
	appendKeyShort string = "a"

	// ToolKey ...
	ToolKey      string = "tool"
	toolKeyShort string = "t"

	// ClearKey ....
	ClearKey      string = "clear"
	clearKeyShort string = "c"

	// HelpKey ...
	HelpKey      string = "help"
	helpKeyShort string = "h"

	// VersionKey ...
	VersionKey      string = "version"
	versionKeyShort string = "v"
)

var (
	// App flags
	flLogLevel = cli.StringFlag{
		Name:   LogLevelKey + ", " + logLevelKeyShort,
		Value:  "info",
		Usage:  "Log level (options: debug, info, warn, error, fatal, panic).",
		EnvVar: LogLevelEnvKey,
	}
	flPath = cli.StringFlag{
		Name:   PathKey + ", " + pathKeyShort,
		EnvVar: "ENVMAN_ENVSTORE_PATH",
		Value:  "",
		Usage:  "Path of the envstore.",
	}
	flTool = cli.BoolFlag{
		Name:   ToolKey + ", " + toolKeyShort,
		EnvVar: "ENVMAN_TOOLMODE",
		Usage:  "If true, envman will NOT ask for user inputs.",
	}
	flags = []cli.Flag{
		flLogLevel,
		flPath,
		flTool,
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
	flNoExpand = cli.BoolFlag{
		Name:  NoExpandKey + ", " + noExpandKeyShort,
		Usage: "If flag is set, envman will NOT replaces ${var} or $var in the string according to the values of the current environment variables.",
	}
	flAppend = cli.BoolFlag{
		Name:  AppendKey + ", " + appendKeyShort,
		Usage: "If flag is set, new env will append to envstore, otherwise if env exist with specified key, will replaced.",
	}
	flClear = cli.BoolFlag{
		Name:  ClearKey + ", " + clearKeyShort,
		Usage: "If flag is set, 'envman init' removes envstore if exist.",
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
