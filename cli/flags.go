package cli

import "github.com/codegangsta/cli"

const (
	// CIEnvKey ...
	CIEnvKey string = "CI"
	// CIKey ...
	CIKey string = "ci"
	cKey  string = "c"

	// LogLevelEnvKey ...
	LogLevelEnvKey string = "LOGLEVEL"
	// LogLevelKey ...
	LogLevelKey      string = "loglevel"
	logLevelKeyShort string = "l"

	// IDKey ...
	IDKey      string = "id"
	idKeyShort string = "i"

	// VersionKey ...
	VersionKey      string = "version"
	versionKeyShort string = "v"

	// PathKey ...
	PathKey      string = "path"
	pathKeyShort string = "p"

	// HelpKey ...
	HelpKey      string = "help"
	helpKeyShort string = "h"
)

var (
	// App flags
	flLogLevel = cli.StringFlag{
		Name:   LogLevelKey + ", " + logLevelKeyShort,
		Value:  "info",
		Usage:  "Log level (options: debug, info, warn, error, fatal, panic).",
		EnvVar: LogLevelEnvKey,
	}
	flTool = cli.BoolFlag{
		Name:   CIKey + ", " + cKey,
		Usage:  "If true it indicates that we're used by another tool so don't require any user input!",
		EnvVar: CIEnvKey,
	}
	flags = []cli.Flag{
		flLogLevel,
		flTool,
	}
	// Command flags
	flID = cli.StringFlag{
		Name:  IDKey + ", " + idKeyShort,
		Value: "",
		Usage: "Step id.",
	}
	flVersion = cli.StringFlag{
		Name:  VersionKey + ", " + versionKeyShort,
		Value: "",
		Usage: "Step version.",
	}
	flPath = cli.StringFlag{
		Name:  PathKey + ", " + pathKeyShort,
		Value: "",
		Usage: "Path where the step will copied.",
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
