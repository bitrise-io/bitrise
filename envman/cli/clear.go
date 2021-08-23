package cli

import (
	"errors"

	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pathutil"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func clearEnvs() error {
	if isExists, err := pathutil.IsPathExists(envman.CurrentEnvStoreFilePath); err != nil {
		return err
	} else if !isExists {
		errMsg := "EnvStore not found in path:" + envman.CurrentEnvStoreFilePath
		return errors.New(errMsg)
	}

	return envman.WriteEnvMapToFile(envman.CurrentEnvStoreFilePath, []models.EnvironmentItemModel{})
}

func clear(c *cli.Context) error {
	log.Debugln("[ENVMAN] - Work path:", envman.CurrentEnvStoreFilePath)

	if err := clearEnvs(); err != nil {
		log.Fatal("[ENVMAN] - Failed to clear EnvStore:", err)
	}

	log.Info("[ENVMAN] - EnvStore cleared")

	return nil
}
