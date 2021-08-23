package cli

import (
	"os"

	"github.com/bitrise-io/envman/env"
	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/command"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// CommandModel ...
type CommandModel struct {
	Command      string
	Argumentums  []string
	Environments []models.EnvironmentItemModel
}

func expandEnvsInString(inp string) string {
	return os.ExpandEnv(inp)
}

func commandEnvs(newEnvs []models.EnvironmentItemModel) ([]string, error) {
	result, err := env.GetDeclarationsSideEffects(newEnvs, &env.DefaultEnvironmentSource{})
	if err != nil {
		return nil, err
	}

	for _, command := range result.CommandHistory {
		if err := env.ExecuteCommand(command); err != nil {
			return nil, err
		}
	}

	return os.Environ(), nil
}

func runCommandModel(cmdModel CommandModel) (int, error) {
	cmdEnvs, err := commandEnvs(cmdModel.Environments)
	if err != nil {
		return 1, err
	}

	return command.RunCommandWithEnvsAndReturnExitCode(cmdEnvs, cmdModel.Command, cmdModel.Argumentums...)
}

func run(c *cli.Context) error {
	log.Debug("[ENVMAN] - Work path:", envman.CurrentEnvStoreFilePath)

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

		cmdToExecute := CommandModel{
			Command:      doCommand,
			Environments: doCmdEnvs,
			Argumentums:  doArgs,
		}

		log.Debug("[ENVMAN] - Executing command:", cmdToExecute)

		if exit, err := runCommandModel(cmdToExecute); err != nil {
			log.Debug("[ENVMAN] - Failed to execute command:", err)
			if exit == 0 {
				log.Error("[ENVMAN] - Failed to execute command:", err)
				exit = 1
			}
			os.Exit(exit)
		}

		log.Debug("[ENVMAN] - Command executed")
	} else {
		log.Fatal("[ENVMAN] - No command specified")
	}

	return nil
}
