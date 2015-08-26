package cli

import "github.com/codegangsta/cli"

const (
	// PathEnvKey ...
	PathEnvKey = "ENVMAN_ENVSTORE_PATH"
	// PathKey ...
	PathKey      = "path"
	pathKeyShort = "p"

	// LogLevelEnvKey ...
	LogLevelEnvKey = "LOGLEVEL"
	// LogLevelKey ...
	LogLevelKey      = "loglevel"
	logLevelKeyShort = "l"

	// KeyKey ...
	KeyKey       = "key"
	keyKeyShortT = "k"

	// ValueKey ...
	ValueKey      = "value"
	valueKeyShort = "v"

	// ValueFileKey ...
	ValueFileKey      = "valuefile"
	valueFileKeyShort = "f"

	// NoExpandKey ...
	NoExpandKey      = "no-expand"
	noExpandKeyShort = "n"

	// AppendKey ...
	AppendKey      = "append"
	appendKeyShort = "a"

	// ToolEnvKey ...
	ToolEnvKey = "ENVMAN_TOOLMODE"
	// ToolKey ...
	ToolKey      = "tool"
	toolKeyShort = "t"

	// ClearKey ....
	ClearKey      = "clear"
	clearKeyShort = "c"

	// HelpKey ...
	HelpKey      = "help"
	helpKeyShort = "h"

	// VersionKey ...
	VersionKey      = "version"
	versionKeyShort = "v"
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
		EnvVar: PathEnvKey,
		Value:  "",
		Usage:  "Path of the envstore.",
	}
	flTool = cli.BoolFlag{
		Name:   ToolKey + ", " + toolKeyShort,
		EnvVar: ToolEnvKey,
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
		Usage: "Key of the environment variable. Empty string (\"\") is NOT accepted.",
	}
	flValue = cli.StringFlag{
		Name:  ValueKey + ", " + valueKeyShort,
		Usage: "Value of the environment variable. Empty string is accepted.",
	}
	flValueFile = cli.StringFlag{
		Name:  ValueFileKey + ", " + valueFileKeyShort,
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

func initHelpAndVersionFlags() {
	cli.HelpFlag = cli.BoolFlag{
		Name:  HelpKey + ", " + helpKeyShort,
		Usage: "Show help.",
	}

	cli.VersionFlag = cli.BoolFlag{
		Name:  VersionKey + ", " + versionKeyShort,
		Usage: "Print the version.",
	}
}
