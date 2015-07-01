package cli

import "github.com/codegangsta/cli"

const (
	CI_ENV_KEY string = "CI"
	TOOL_KEY   string = "tool"
	CI_KEY     string = "ci"
	QUIT_KEY   string = "quite"
	Q_KEY      string = "q"

	LOG_LEVEL_KEY string = "loglevel"
	L_KEY         string = "l"

	ID_KEY string = "id"
	I_KEY  string = "i"

	VERSION_KEY string = "version"
	V_KEY       string = "v"

	PATH_KEY string = "path"
	P_KEY    string = "p"
)

var (
	// App flags
	flLogLevel = cli.StringFlag{
		Name:  LOG_LEVEL_KEY + ", " + L_KEY,
		Value: "info",
		Usage: "Log level (options: debug, info, warn, error, fatal, panic).",
	}
	flTool = cli.BoolFlag{
		Name:   TOOL_KEY + ", " + CI_KEY + ", " + QUIT_KEY + ", " + Q_KEY,
		Usage:  "If true it indicates that we're used by another tool so don't require any user input!",
		EnvVar: CI_ENV_KEY,
	}
	flags = []cli.Flag{
		flLogLevel,
		flTool,
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
