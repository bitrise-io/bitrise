package cli

import _ "github.com/codegangsta/cli"

const (
	PATH_KEY string = "path"
	P_KEY    string = "p"

	LOG_LEVEL_KEY string = "log-level"
	L_KEY         string = "l"

	KEY_KEY string = "key"
	K_KEY   string = "k"

	VALUE_KEY string = "value"
	V_KEY     string = "v"

	VALUE_FILE_KEY string = "valuefile"
	VF_KEY         string = "f"

	EXPAND_KEY string = "expand"
	E_KEY      string = "e"
)

var (
/*
	// App flags
	flPath = cli.StringFlag{
		Name:   PATH_KEY + ", " + P_KEY,
		EnvVar: "ENVMAN_ENVSTORE_PATH",
		Value:  "",
		Usage:  "Envman's working path (SOME_DIR/envstore.yml).",
	}
	flLogLevel = cli.StringFlag{
		Name:  LOG_LEVEL_KEY + ", " + L_KEY,
		Value: "info",
		Usage: "Log level (options: debug, info, warn, error, fatal, panic).",
	}
	flags = []cli.Flag{
		flPath,
		flLogLevel,
	}
*/

/*
	// Command flags
	flKey = cli.StringFlag{
		Name:  KEY_KEY + ", " + K_KEY,
		Value: "",
		Usage: "Key of the environment variable.",
	}
	flValue = cli.StringFlag{
		Name:  VALUE_KEY + ", " + V_KEY,
		Value: "",
		Usage: "Value of the environment variable.",
	}
	flValueFile = cli.StringFlag{
		Name:  VALUE_FILE_KEY + ", " + VF_KEY,
		Value: "",
		Usage: "Path of the environment variable value.",
	}
	flIsExpand = cli.StringFlag{
		Name:  EXPAND_KEY + ", " + E_KEY,
		Value: "true",
		Usage: "If true, replaces ${var} or $var in the string according to the values of the current environment variables.",
	}
*/
)
