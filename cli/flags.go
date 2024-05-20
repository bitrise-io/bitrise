package cli

import (
	"github.com/bitrise-io/bitrise/configs"
	"github.com/urfave/cli"
)

const (
	CollectionPathEnvKey = "STEPMAN_COLLECTION"
	CIKey = "ci"
	PRKey = "pr"
	DebugModeKey = "debug"

	VersionKey      = "version"
	versionKeyShort = "v"

	CollectionKey      = "collection"
	collectionKeyShort = "c"

	inventoryShortKey = "i"

	InventoryBase64Key = "inventory-base64"

	configShortKey = "c"

	ConfigBase64Key = "config-base64"

	HelpKey      = "help"
	helpKeyShort = "h"

	MinimalModeKey = "minimal"
	FullModeKey = "full"

	outputFormatKeyShort = "f"
	OutputPathKey = "outpath"
	PrettyFormatKey = "pretty"

	IDKey      = "id"
	idKeyShort = "i"
	ShortKey = "short"

	StepYMLKey = "step-yml"

	//
	// Stepman share

	TagKey = "tag"
	GitKey = "git"
	StepIDKey = "stepid"
)

var (
	// App flags
	flDebugMode = cli.BoolFlag{
		Name:   DebugModeKey,
		Usage:  "If true it enables DEBUG mode.",
		EnvVar: configs.DebugModeEnvKey,
	}
	flTool = cli.BoolFlag{
		Name:   CIKey,
		Usage:  "If true it indicates that we're used by another tool so don't require any user input!",
		EnvVar: configs.CIModeEnvKey,
	}
	flPRMode = cli.BoolFlag{
		Name:  PRKey,
		Usage: "If true bitrise runs in pull request mode.",
	}
	flags = []cli.Flag{
		flDebugMode,
		flTool,
		flPRMode,
	}
	// Command flags
	flOutputFormat = cli.StringFlag{
		Name:  OutputFormatKey + ", " + outputFormatKeyShort,
		Usage: "Output format. Accepted: raw (default), json, yml",
	}
	flCollection = cli.StringFlag{
		Name:   CollectionKey + ", " + collectionKeyShort,
		Usage:  "Collection of step.",
		EnvVar: CollectionPathEnvKey,
	}
	flConfig = cli.StringFlag{
		Name:  ConfigKey + ", " + configShortKey,
		Usage: "Path where the workflow config file is located.",
	}
	flConfigBase64 = cli.StringFlag{
		Name:  ConfigBase64Key,
		Usage: "base64 decoded config data.",
	}
	flInventory = cli.StringFlag{
		Name:  InventoryKey + ", " + inventoryShortKey,
		Usage: "Path of the inventory file.",
	}
	flInventoryBase64 = cli.StringFlag{
		Name:  InventoryBase64Key,
		Usage: "base64 decoded inventory data.",
	}

	// Export
	flFormat = cli.StringFlag{
		Name:  OutputFormatKey,
		Usage: "Output format. Accepted: json, yml.",
	}
	flOutputPath = cli.StringFlag{
		Name:  OutputPathKey,
		Usage: "Output path, where the exported file will be saved.",
	}
	flPretty = cli.BoolFlag{
		Name:  PrettyFormatKey,
		Usage: "Pretty printed export?",
	}
	flID = cli.StringFlag{
		Name:  IDKey + ", " + idKeyShort,
		Usage: "Step id.",
	}
	flVersion = cli.StringFlag{
		Name:  VersionKey + ", " + versionKeyShort,
		Usage: "Step version.",
	}
	flShort = cli.BoolFlag{
		Name:  ShortKey,
		Usage: "Show short version of infos.",
	}
	flStepYML = cli.StringFlag{
		Name:  StepYMLKey,
		Usage: "Path of step.yml",
	}

	// Stepman share
	flTag = cli.StringFlag{
		Name:  TagKey,
		Usage: "Git (version) tag.",
	}
	flGit = cli.StringFlag{
		Name:  GitKey,
		Usage: "Git clone url of the step repository.",
	}
	flStepID = cli.StringFlag{
		Name:  StepIDKey,
		Usage: "ID of the step.",
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
