package cli

import (
	"fmt"

	"github.com/bitrise-io/go-utils/command"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func initEnvStore(c *cli.Context) error {
	log.Debugln("[ENVMAN] - Work path:", CurrentEnvStoreFilePath)
	clearEnvstore := c.Bool(ClearKey)
	err := InitEnvStore(CurrentEnvStoreFilePath, clearEnvstore)
	log.Debugln("[ENVMAN] - Initialized")
	return err
}

// InitEnvStore ...
func InitEnvStore(envStorePth string, clearEnvstore bool) error {
	if clearEnvstore {
		if err := command.RemoveFile(envStorePth); err != nil {
			return fmt.Errorf("failed to clear path: %s", err)
		}
	}

	if err := initAtPath(envStorePth); err != nil {
		return fmt.Errorf("failed to init at path: %s", err)
	}

	return nil
}
