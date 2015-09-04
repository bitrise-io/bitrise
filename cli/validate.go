package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func validate(c *cli.Context) {
	// Config validation
	_, err := CreateBitriseConfigFromCLIParams(c)
	if err != nil {
		log.Fatalf("Failed to create bitrise cofing, err: %s", err)
	}

	log.Info("Valid bitrise config")
}
