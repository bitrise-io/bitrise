package cli

import (
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/codegangsta/cli"
)

const (
	// OutputFormatRaw ...
	OutputFormatRaw = "raw"
	// OutputFormatJSON ...
	OutputFormatJSON = "json"
	// CollectionPathEnvKey ...
	CollectionPathEnvKey = "STEPMAN_COLLECTION"
	// CIKey ...
	CIKey = "ci"
	// DebugModeKey ...
	DebugModeKey = "debug"

	// LogLevelKey ...
	LogLevelKey      = "loglevel"
	logLevelKeyShort = "l"

	// VersionKey ...
	VersionKey      = "version"
	versionKeyShort = "v"

	// PathKey ...
	PathKey      = "path"
	pathKeyShort = "p"

	// CollectionKey ...
	CollectionKey      = "collection"
	collectionKeyShort = "c"

	// InventoryKey ...
	InventoryKey      = "inventory"
	inventoryShortKey = "i"

	// InventoryBase64Key ...
	InventoryBase64Key = "inventory-base64"

	// ConfigKey ...
	ConfigKey      = "config"
	configShortKey = "c"

	// ConfigBase64Key ...
	ConfigBase64Key = "config-base64"

	// HelpKey ...
	HelpKey      = "help"
	helpKeyShort = "h"

	// MinimalModeKey ...
	MinimalModeKey = "minimal"

	// OuputFormatKey ...
	OuputFormatKey = "format"
	// OuputPathKey ...
	OuputPathKey = "outpath"
	// PrettyFormatKey ...
	PrettyFormatKey = "pretty"

	// IDKey ...
	IDKey      = "id"
	idKeyShort = "i"
	// ShortKey ...
	ShortKey = "short"
)

var (
	// App flags
	flLogLevel = cli.StringFlag{
		Name:   LogLevelKey + ", " + logLevelKeyShort,
		Usage:  "Log level (options: debug, info, warn, error, fatal, panic).",
		EnvVar: bitrise.LogLevelEnvKey,
	}
	flDebugMode = cli.BoolFlag{
		Name:   DebugModeKey,
		Usage:  "If true it enabled DEBUG mode. If no separate Log Level is specified this will also set the loglevel to debug.",
		EnvVar: bitrise.DebugModeEnvKey,
	}
	flTool = cli.BoolFlag{
		Name:   CIKey,
		Usage:  "If true it indicates that we're used by another tool so don't require any user input!",
		EnvVar: bitrise.CIModeEnvKey,
	}
	flags = []cli.Flag{
		flLogLevel,
		flDebugMode,
		flTool,
	}
	// Command flags
	flPath = cli.StringFlag{
		Name:  PathKey + ", " + pathKeyShort,
		Usage: "[Deprecated!!! Use 'config'] Path where the workflow config file is located.",
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
	// Setup
	flMinimalSetup = cli.BoolFlag{
		Name:  MinimalModeKey,
		Usage: "Minimal setup mode: skips more thorough checking, like brew doctor.",
	}
	// Export
	flFormat = cli.StringFlag{
		Name:  OuputFormatKey,
		Usage: "Output format. Accepted: json, yml",
	}
	flOutputPath = cli.StringFlag{
		Name:  OuputPathKey,
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
