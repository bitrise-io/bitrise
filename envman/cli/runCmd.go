package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/codegangsta/cli"
)

func runCmd(c *cli.Context) {
	log.Info("Work path:", envman.CurrentEnvStoreFilePath)

	if len(c.Args()) > 0 {
		doCmdEnvs, err := envman.LoadEnvMap()
		if err != nil {
			log.Fatal("Failed to load EnvStore:", err)
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

		log.Info("Executing command:", cmdToExecute)

		if err := envman.ExecuteCmd(cmdToExecute); err != nil {
			log.Fatal("Failed to execute command:", err)
		}

		log.Info("Command executed")
	} else {
		log.Fatal("No command specified")
	}
}
