package cli

import "github.com/codegangsta/cli"

const (
	// HelpKey ...
	HelpKey      string = "help"
	helpKeyShort string = "h"

	// VersionKey ...
	VersionKey      string = "version"
	versionKeyShort string = "v"

	// CollectionPathEnvKey ...
	CollectionPathEnvKey string = "STEPMAN_COLLECTION"
	// CollectionKey ...
	CollectionKey      string = "collection"
	collectionKeyShort string = "c"
	// LocalCollectionKey ...
	LocalCollectionKey string = "local"
	// CopySpecJSONKey ...
	CopySpecJSONKey string = "copy-spec-json"

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

	// PathKey ...
	PathKey      string = "path"
	pathKeyShort string = "p"

	// CopyYMLKey ...
	CopyYMLKey      string = "copyyml"
	copyYMLKeyShort string = "y"

	// UpdateKey ...
	UpdateKey      string = "update"
	updateKeyShort string = "u"
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
		Usage:  "Debug mode.",
		EnvVar: DebugEnvKey,
	}
	flags = []cli.Flag{
		flDebug,
		flLogLevel,
	}
	// Command flags
	flCollection = cli.StringFlag{
		Name:   CollectionKey + ", " + collectionKeyShort,
		Usage:  "Collection of step.",
		EnvVar: CollectionPathEnvKey,
	}
	flLocalCollection = cli.BoolFlag{
		Name:  LocalCollectionKey,
		Usage: "Allow the --collection to be a local path.",
	}
	flCopySpecJSON = cli.StringFlag{
		Name:  CopySpecJSONKey,
		Usage: "If setup succeeds copy the generates spec.json to this path.",
	}
	flID = cli.StringFlag{
		Name:  IDKey + ", " + idKeyShort,
		Usage: "Step id.",
	}
	flVersion = cli.StringFlag{
		Name:  VersionKey + ", " + versionKeyShort,
		Usage: "Step version.",
	}
	flPath = cli.StringFlag{
		Name:  PathKey + ", " + pathKeyShort,
		Usage: "Path where the step will copied.",
	}
	flCopyYML = cli.StringFlag{
		Name:  CopyYMLKey + ", " + copyYMLKeyShort,
		Usage: "Path where the selected/activated step's step.yml will be copied.",
	}
	flUpdate = cli.BoolFlag{
		Name:  UpdateKey + ", " + updateKeyShort,
		Usage: "If flag is set, and collection doesn't contains the specified step, the collection will updated.",
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
