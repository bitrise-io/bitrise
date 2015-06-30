package cli

import "github.com/codegangsta/cli"

const (
	ID_KEY string = "id"
	I_KEY  string = "i"

	VERSION_KEY string = "version"
	V_KEY       string = "v"

	COLLECTION_KEY string = "collection"
	C_KEY          string = "c"
)

var (
	// Command flags
	flId = cli.StringFlag{
		Name:  ID_KEY + ", " + I_KEY,
		Value: "",
		Usage: "Step id.",
	}
	flVersions = cli.StringFlag{
		Name:  VERSION_KEY + ", " + V_KEY,
		Value: "",
		Usage: "Step version.",
	}
	flCollection = cli.StringFlag{
		Name:  COLLECTION_KEY + ", " + C_KEY,
		Value: "",
		Usage: "Collection of step.",
	}
)
