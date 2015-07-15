package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/codegangsta/cli"
)

func parseClear(c *cli.Context) bool {
	if c.IsSet(ClearKey) {
		return c.Bool(ClearKey)
	}
	return false
}

func initCmd(c *cli.Context) {
	log.Debugln("[ENVMAN] - Work path:", envman.CurrentEnvStoreFilePath)

	clear := parseClear(c)
	if clear {
		if err := envman.ClearPathIfExist(envman.CurrentEnvStoreFilePath); err != nil {
			log.Fatal("[ENVMAN] - Failed to clear path:", err)
		}
	}

	if err := envman.InitAtPath(envman.CurrentEnvStoreFilePath); err != nil {
		log.Fatal("[ENVMAN] - Failed to init at path:", err)
	}

	log.Debugln("[ENVMAN] - Initialized")
}
