package cli

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/codegangsta/cli"
)

func runCmd(c *cli.Context) {
	log.Debugln("[ENVMAN] - Work path:", envman.CurrentEnvStoreFilePath)

	if len(c.Args()) > 0 {
		doCmdEnvs, err := envman.ReadEnvs(envman.CurrentEnvStoreFilePath)
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

		log.Debugln("[ENVMAN] - Executing command:", cmdToExecute)

		if exit, err := envman.RunCmd(cmdToExecute); err != nil {
			log.Error("[ENVMAN] - Failed to execute command:", err)
			os.Exit(exit)
		}

		log.Debugln("[ENVMAN] - Command executed")
	} else {
		log.Fatal("[ENVMAN] - No command specified")
	}
}
