package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/codegangsta/cli"
)

func initEnvStore(c *cli.Context) {
	log.Debugln("[ENVMAN] - Work path:", envman.CurrentEnvStoreFilePath)

	clear := c.Bool(ClearKey)
	if clear {
		if err := cmdex.RemoveFile(envman.CurrentEnvStoreFilePath); err != nil {
			log.Fatal("[ENVMAN] - Failed to clear path:", err)
		}
	}

	if err := envman.InitAtPath(envman.CurrentEnvStoreFilePath); err != nil {
		log.Fatal("[ENVMAN] - Failed to init at path:", err)
	}

	log.Debugln("[ENVMAN] - Initialized")
}
