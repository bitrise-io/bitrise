package cli

import "github.com/codegangsta/cli"

const (
	// CollectionPathEnvKey ...
	CollectionPathEnvKey string = "STEPMAN_COLLECTION"
	// CollectionKey ...
	CollectionKey      string = "collection"
	collectionKeyShort string = "c"

	// DebugEnvKey ...
	DebugEnvKey string = "STEPMAN_DEBUG"
	// DebugKey ...
	DebugKey      string = "debug"
	debugKeyShort string = "d"

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

	// CopyYMLKey ...
	CopyYMLKey      string = "copyyml"
	copyYMLKeyShort string = "y"

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
	flDebug = cli.BoolFlag{
		Name:   DebugKey + ", " + debugKeyShort,
		EnvVar: DebugEnvKey,
		Usage:  "Debug mode.",
	}
	flags = []cli.Flag{
		flDebug,
		flLogLevel,
	}
	// Command flags
	flCollection = cli.StringFlag{
		Name:   CollectionKey + ", " + collectionKeyShort,
		Value:  "",
		EnvVar: CollectionPathEnvKey,
		Usage:  "Collection of step.",
	}
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
	flCopyYML = cli.StringFlag{
		Name:  CopyYMLKey + ", " + copyYMLKeyShort,
		Value: "",
		Usage: "Path where the stp.yml will copied.",
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
