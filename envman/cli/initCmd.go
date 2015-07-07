package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/codegangsta/cli"
)

func initCmd(c *cli.Context) {
	log.Info("Work path:", envman.CurrentEnvStoreFilePath)

	if err := envman.InitAtPath(envman.CurrentEnvStoreFilePath); err != nil {
		log.Fatal(err)
	}

	log.Info("Envman initialized")
}
