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
	flDebug = cli.StringFlag{
		Name:   DebugKey + ", " + debugKeyShort,
		Value:  "false",
		EnvVar: DebugEnvKey,
		Usage:  "Debug mode.",
	}
	flCollection = cli.StringFlag{
		Name:   CollectionKey + ", " + collectionKeyShort,
		Value:  "",
		EnvVar: CollectionPathEnvKey,
		Usage:  "Collection of step.",
	}
	flags = []cli.Flag{
		flDebug,
		flCollection,
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
