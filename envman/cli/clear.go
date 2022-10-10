package cli

import (
	"errors"

	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pathutil"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func clear(c *cli.Context) error {
	log.Debugln("[ENVMAN] - Work path:", CurrentEnvStoreFilePath)

	if err := ClearEnvs(CurrentEnvStoreFilePath); err != nil {
		log.Fatal("[ENVMAN] - Failed to clear EnvStore:", err)
	}

	log.Info("[ENVMAN] - EnvStore cleared")

	return nil
}

// ClearEnvs ...
func ClearEnvs(envStorePth string) error {
	if isExists, err := pathutil.IsPathExists(envStorePth); err != nil {
		return err
	} else if !isExists {
		return errors.New("EnvStore not found in path:" + envStorePth)
	}

	return WriteEnvMapToFile(envStorePth, []models.EnvironmentItemModel{})
}
