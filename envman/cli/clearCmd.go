package cli

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/go-pathutil"
	"github.com/codegangsta/cli"
)

func clearEnvs() error {
	if isExists, err := pathutil.IsPathExists(envman.CurrentEnvStoreFilePath); err != nil {
		return err
	} else if !isExists {
		errMsg := "EnvStore not found in path:" + envman.CurrentEnvStoreFilePath
		return errors.New(errMsg)
	}

	if err := envman.WriteEnvMapToFile(envman.CurrentEnvStoreFilePath, []envman.EnvModel{}); err != nil {
		return err
	}

	return nil
}

func clearCmd(c *cli.Context) {
	log.Info("Work path:", envman.CurrentEnvStoreFilePath)

	if err := clearEnvs(); err != nil {
		log.Fatal("Failed to clear EnvStore:", err)
	}

	log.Info("EnvStore cleared")
}
