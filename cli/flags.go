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

	// VersionKey ...
	VersionKey      string = "version"
	versionKeyShort string = "v"

	// PathKey ...
	PathKey      string = "path"
	pathKeyShort string = "p"

	// InventoryKey ...
	InventoryKey      string = "inventory"
	inventoryShortKey string = "i"

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
	flPath = cli.StringFlag{
		Name:  PathKey + ", " + pathKeyShort,
		Value: "",
		Usage: "Path where the workflow config file is located.",
	}
	flInventory = cli.StringFlag{
		Name:  InventoryKey + ", " + inventoryShortKey,
		Value: "",
		Usage: "Path of the inventory file.",
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
