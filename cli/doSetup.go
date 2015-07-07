package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const (
	ENVSTORE_PATH_ENV_KEY         string = "ENVMAN_ENVSTORE_PATH"
	ENVSTORE_PATH                 string = "/Users/godrei/develop/bitrise/bitrise-cli-test/envstore.yml"
	FORMATTED_OUTPUT_PATH_ENV_KEY string = "BITRISE_STEP_FORMATTED_OUTPUT_FILE_PATH"
	FORMATTED_OUTPUT_PATH         string = "/Users/godrei/develop/bitrise/bitrise-cli-test/formout.md"
)

func doSetup(c *cli.Context) {
	log.Info("[BITRISE_CLI] - Setup -- Coming soon!")
}
