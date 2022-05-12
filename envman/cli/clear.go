package cli

import (
	"errors"

	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pathutil"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func clearEnvs() error {
	if isExists, err := pathutil.IsPathExists(CurrentEnvStoreFilePath); err != nil {
		return err
	} else if !isExists {
		errMsg := "EnvStore not found in path:" + CurrentEnvStoreFilePath
		return errors.New(errMsg)
	}

	return WriteEnvMapToFile(CurrentEnvStoreFilePath, []models.EnvironmentItemModel{})
}

func clear(c *cli.Context) error {
	log.Debugln("[ENVMAN] - Work path:", CurrentEnvStoreFilePath)

	if err := clearEnvs(); err != nil {
		log.Fatal("[ENVMAN] - Failed to clear EnvStore:", err)
	}

	log.Info("[ENVMAN] - EnvStore cleared")

	return nil
}
