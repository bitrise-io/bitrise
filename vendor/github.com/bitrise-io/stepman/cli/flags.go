package cli

import "github.com/urfave/cli"

const (
	// DebugEnvKey ...
	DebugEnvKey = "STEPMAN_DEBUG"
	// LogLevelEnvKey ...
	LogLevelEnvKey = "LOGLEVEL"
	// CollectionPathEnvKey ...
	CollectionPathEnvKey = "STEPMAN_COLLECTION"

	// HelpKey ...
	HelpKey      = "help"
	helpKeyShort = "h"

	// VersionKey ...
	VersionKey      = "version"
	versionKeyShort = "v"

	// CollectionKey ...
	CollectionKey      = "collection"
	collectionKeyShort = "c"
	// LocalCollectionKey ...
	LocalCollectionKey = "local"
	// CopySpecJSONKey ...
	CopySpecJSONKey = "copy-spec-json"

	// DebugKey ...
	DebugKey      = "debug"
	debugKeyShort = "d"

	// LogLevelKey ...
	LogLevelKey      = "loglevel"
	logLevelKeyShort = "l"

	// IDKey ...
	IDKey      = "id"
	idKeyShort = "i"

	// PathKey ...
	PathKey      = "path"
	pathKeyShort = "p"

	// CopyYMLKey ...
	CopyYMLKey      = "copyyml"
	copyYMLKeyShort = "y"

	// UpdateKey ...
	UpdateKey      = "update"
	updateKeyShort = "u"

	// TagKey ...
	TagKey      = "tag"
	tagKeyShort = "t"

	// GitKey ...
	GitKey      = "git"
	gitKeyShort = "g"

	// StepIDKey ...
	StepIDKey      = "stepid"
	stepIDKeyShort = "s"

	// ShortKey ...
	ShortKey = "short"

	// ToolMode ...
	ToolMode = "toolmode"

	// FormatKey ...
	FormatKey      = "format"
	formatKeyShort = "f"
	// OutputFormatRaw ...
	OutputFormatRaw = "raw"
	// OutputFormatJSON ...
	OutputFormatJSON = "json"

	// StepYMLKey ...
	StepYMLKey = "step-yml"

	StepYMLOverrideKey = "stepyml-override"
)

var (
	// App flags
	flLogLevel = cli.StringFlag{
		Name:   LogLevelKey + ", " + logLevelKeyShort,
		Value:  "info",
		Usage:  "Log level (options: debug, info, warn, error, fatal, panic).",
		EnvVar: LogLevelEnvKey,
	}
	flags = []cli.Flag{
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
		Usage: "[Deprecated!!!][Use 'file://' in steplib uri instead] Allow the --collection to be a local path.",
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
	flUpdate = cli.BoolFlag{
		Name:  UpdateKey + ", " + updateKeyShort,
		Usage: "If flag is set, and collection doesn't contains the specified step, the collection will updated.",
	}
	flTag = cli.StringFlag{
		Name:  TagKey + ", " + tagKeyShort,
		Usage: "Git (version) tag.",
	}
	flGit = cli.StringFlag{
		Name:  GitKey + ", " + gitKeyShort,
		Usage: "Git clone url of the step repository.",
	}
	flStepID = cli.StringFlag{
		Name:  StepIDKey + ", " + stepIDKeyShort,
		Usage: "ID of the step.",
	}
	flFormat = cli.StringFlag{
		Name:  FormatKey + ", " + formatKeyShort,
		Usage: "Output format (options: raw, json).",
	}
	flToolMode = cli.BoolFlag{
		Name:  ToolMode,
		Usage: "Stepman called as tool.",
	}
	flStepYMLOverride = cli.StringFlag{
		Name:  StepYMLOverrideKey,
		Usage: "Path to a step.yml file that will override the one from the git checkout.",
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
