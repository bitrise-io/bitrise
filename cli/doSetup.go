package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const (
	// EnvstorePathEnvKey ...
	EnvstorePathEnvKey string = "ENVMAN_ENVSTORE_PATH"
	// EnvstorePath ...
	EnvstorePath string = "/Users/godrei/develop/bitrise/bitrise-cli-test/envstore.yml"
	// FormattedOutputPathEnvKey ...
	FormattedOutputPathEnvKey string = "BITRISE_STEP_FORMATTED_OUTPUT_FILE_PATH"
	// FormattedOutputPath ...
	FormattedOutputPath string = "/Users/godrei/develop/bitrise/bitrise-cli-test/formout.md"
)

func doSetup(c *cli.Context) {
	log.Info("[BITRISE_CLI] - Setup -- Coming soon!")
	os.Exit(1)
}
