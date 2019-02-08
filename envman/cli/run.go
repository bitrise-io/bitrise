package cli

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/command"
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

func commandEnvs(envs []models.EnvironmentItemModel) ([]string, error) {
	for _, env := range envs {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return []string{}, err
		}

		opts, err := env.GetOptions()
		if err != nil {
			return []string{}, err
		}

		if opts.Unset != nil && *opts.Unset {
			if err := os.Unsetenv(key); err != nil {
				return []string{}, fmt.Errorf("unset env (%s): %s", key, err)
			}
			continue
		}

		if *opts.SkipIfEmpty && value == "" {
			continue
		}

		var valueStr string
		if *opts.IsExpand {
			valueStr = expandEnvsInString(value)
		} else {
			valueStr = value
		}

		if err := os.Setenv(key, valueStr); err != nil {
			return []string{}, err
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
