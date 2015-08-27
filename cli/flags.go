package cli

import (
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/codegangsta/cli"
)

const (
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

	// InventoryKey ...
	InventoryKey      = "inventory"
	inventoryShortKey = "i"

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
		Usage: "Path where the workflow config file is located.",
	}
	flInventory = cli.StringFlag{
		Name:  InventoryKey + ", " + inventoryShortKey,
		Usage: "Path of the inventory file.",
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
