package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/codegangsta/cli"
)

func runCmd(c *cli.Context) {
	log.Info("[ENVMAN] - Work path:", envman.CurrentEnvStoreFilePath)

	if len(c.Args()) > 0 {
		doCmdEnvs, err := envman.LoadEnvMap()
		if err != nil {
			log.Fatal("[ENVMAN] - Failed to load EnvStore:", err)
		}

		doCommand := c.Args()[0]

		doArgs := []string{}
		if len(c.Args()) > 1 {
			doArgs = c.Args()[1:]
		}

		cmdToExecute := envman.CommandModel{
			Command:      doCommand,
			Environments: doCmdEnvs,
			Argumentums:  doArgs,
		}

		log.Info("[ENVMAN] - Executing command:", cmdToExecute)

		if err := envman.ExecuteCmd(cmdToExecute); err != nil {
			log.Fatal("[ENVMAN] - Failed to execute command:", err)
		}

		log.Info("[ENVMAN] - Command executed")
	} else {
		log.Fatal("[ENVMAN] - No command specified")
	}
}
