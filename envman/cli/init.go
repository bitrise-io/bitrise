package cli

import (
	"github.com/bitrise-io/go-utils/command"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func initEnvStore(c *cli.Context) error {
	log.Debugln("[ENVMAN] - Work path:", CurrentEnvStoreFilePath)

	clear := c.Bool(ClearKey)
	if clear {
		if err := command.RemoveFile(CurrentEnvStoreFilePath); err != nil {
			log.Fatal("[ENVMAN] - Failed to clear path:", err)
		}
	}

	if err := InitAtPath(CurrentEnvStoreFilePath); err != nil {
		log.Fatal("[ENVMAN] - Failed to init at path:", err)
	}

	log.Debugln("[ENVMAN] - Initialized")

	return nil
}
