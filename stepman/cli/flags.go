package cli

import "github.com/codegangsta/cli"

const (
	COLLECTION_PATH_ENV_KEY string = "STEPMAN_COLLECTION"
	COLLECTION_KEY          string = "collection"
	COLLECTION_KEY_SHORT    string = "c"

	DEBUG_ENV_KEY   string = "STEPMAN_DEBUG"
	DEBUG_KEY       string = "debug"
	DEBUG_KEY_SHORT string = "d"

	ID_KEY       string = "id"
	ID_KEY_SHORT string = "i"

	VERSION_KEY       string = "version"
	VERSION_KEY_SHORT string = "v"

	PATH_KEY       string = "path"
	PATH_KEY_SHORT string = "p"

	HELP_KEY       string = "help"
	HELP_KEY_SHORT string = "h"
)

var (
	// App flags
	flDebug = cli.StringFlag{
		Name:   DEBUG_KEY + ", " + DEBUG_KEY_SHORT,
		Value:  "false",
		EnvVar: DEBUG_ENV_KEY,
		Usage:  "Debug mode.",
	}
	flCollection = cli.StringFlag{
		Name:   COLLECTION_KEY + ", " + COLLECTION_KEY_SHORT,
		Value:  "",
		EnvVar: COLLECTION_PATH_ENV_KEY,
		Usage:  "Collection of step.",
	}
	flags = []cli.Flag{
		flDebug,
		flCollection,
	}
	// Command flags
	flId = cli.StringFlag{
		Name:  ID_KEY + ", " + ID_KEY_SHORT,
		Value: "",
		Usage: "Step id.",
	}
	flVersion = cli.StringFlag{
		Name:  VERSION_KEY + ", " + VERSION_KEY_SHORT,
		Value: "",
		Usage: "Step version.",
	}
	flPath = cli.StringFlag{
		Name:  PATH_KEY + ", " + PATH_KEY_SHORT,
		Value: "",
		Usage: "Path where the step will copied.",
	}
)

func init() {
	// Override default help and version flags
	cli.HelpFlag = cli.BoolFlag{
		Name:  HELP_KEY + ", " + HELP_KEY_SHORT,
		Usage: "Show help.",
	}

	cli.VersionFlag = cli.BoolFlag{
		Name:  VERSION_KEY + ", " + VERSION_KEY_SHORT,
		Usage: "Print the version.",
	}
}
