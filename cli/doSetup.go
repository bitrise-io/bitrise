package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise-cli/bitrise"
	"github.com/codegangsta/cli"
)

func doSetup(c *cli.Context) {
	log.Info("Setup")

	// Envman
	os.Setenv("ENVMAN_ENVSTORE_PATH", "/Users/godrei/develop/bitrise/bitrise-cli-test/help.yml")

	err := bitrise.RunEnvmanInit()
	if err != nil {
		log.Errorln("Failed to run init envman")
		return
	}

	envs := map[string]string{
		"BITRISE_STEP_FORMATTED_OUTPUT_FILE_PATH": "/Users/godrei/develop/bitrise/bitrise-cli-test/formout.md",
	}
	for key, value := range envs {
		err := bitrise.RunEnvmanAdd(key, value)
		if err != nil {
			log.Errorln("Failed to run envman add")
			return
		}
	}

	// Stepman
	err = bitrise.RunStepmanSetup()
	if err != nil {
		log.Errorln("Failed to run stepman setup")
		return
	}
}
