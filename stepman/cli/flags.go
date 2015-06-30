package cli

import "github.com/codegangsta/cli"

const (
	COLLECTION_PATH_ENV_KEY string = "STEPMAN_COLLECTION"
	COLLECTION_KEY          string = "collection"
	C_KEY                   string = "c"

	ID_KEY string = "id"
	I_KEY  string = "i"

	VERSION_KEY string = "tag"
	V_KEY       string = "t"

	PATH_KEY string = "path"
	P_KEY    string = "p"
)

var (
	// App flags
	flCollection = cli.StringFlag{
		Name:   COLLECTION_KEY + ", " + C_KEY,
		Value:  "",
		EnvVar: COLLECTION_PATH_ENV_KEY,
		Usage:  "Collection of step.",
	}
	flags = []cli.Flag{
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
