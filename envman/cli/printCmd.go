package cli

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/codegangsta/cli"
)

func printEnvs() error {
	environments, err := envman.LoadEnvMap()
	if err != nil {
		return err
	}

	for _, eModel := range environments {
		envString := "- " + eModel.Key + ": " + eModel.Value
		fmt.Println(envString)
		if eModel.IsExpand == false {
			expandString := "  " + "isExpand" + ": " + "false"
			fmt.Println(expandString)
		}
	}

	return nil
}

func printCmd(c *cli.Context) {
	log.Info("Work path:", envman.CurrentEnvStoreFilePath)

	if err := printEnvs(); err != nil {
		log.Fatal("Failed to print:", err)
	}
}
