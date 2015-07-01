package cli

import "github.com/codegangsta/cli"

const (
	COLLECTION_PATH_ENV_KEY string = "STEPMAN_COLLECTION"
	COLLECTION_KEY          string = "collection"
	C_KEY                   string = "c"

	DEBUG_ENV_KEY string = "STEPMAN_DEBUG"
	DEBUG_KEY     string = "debug"
	D_KEY         string = "d"

	ID_KEY string = "id"
	I_KEY  string = "i"

	VERSION_KEY string = "version"
	V_KEY       string = "v"

	PATH_KEY string = "path"
	P_KEY    string = "p"
)

var (
	// App flags
	flDebug = cli.StringFlag{
		Name:   DEBUG_KEY + ", " + D_KEY,
		Value:  "false",
		EnvVar: DEBUG_ENV_KEY,
		Usage:  "Debug mode.",
	}
	flCollection = cli.StringFlag{
		Name:   COLLECTION_KEY + ", " + C_KEY,
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
		Name:  ID_KEY + ", " + I_KEY,
		Value: "",
		Usage: "Step id.",
	}
	flVersion = cli.StringFlag{
		Name:  VERSION_KEY + ", " + V_KEY,
		Value: "",
		Usage: "Step version.",
	}
	flPath = cli.StringFlag{
		Name:  PATH_KEY + ", " + P_KEY,
		Value: "",
		Usage: "Path where the step will copied.",
	}
)
