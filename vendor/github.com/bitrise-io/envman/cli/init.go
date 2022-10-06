package cli

import (
	"fmt"

	"github.com/bitrise-io/go-utils/command"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func initEnvStore(c *cli.Context) error {
	log.Debugln("[ENVMAN] - Work path:", CurrentEnvStoreFilePath)
	clear := c.Bool(ClearKey)
	err := Init(CurrentEnvStoreFilePath, clear)
	log.Debugln("[ENVMAN] - Initialized")
	return err
}

// Init ...
func Init(envStorePth string, clear bool) error {
	if clear {
		if err := command.RemoveFile(envStorePth); err != nil {
			return fmt.Errorf("failed to clear path: %s", err)
		}
	}

	if err := InitAtPath(envStorePth); err != nil {
		return fmt.Errorf("failed to init at path: %s", err)
	}

	return nil
}
