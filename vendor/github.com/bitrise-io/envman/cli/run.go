package cli

import (
	"fmt"
	"os"

	"github.com/bitrise-io/envman/env"
	"github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/command"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// CommandModel ...
type CommandModel struct {
	Command      string
	Arguments    []string
	Environments []models.EnvironmentItemModel
}

func run(c *cli.Context) error {
	if len(c.Args()) == 0 {
		log.Fatal("[ENVMAN] - No command specified")
	}

	cmd, err := createCommand(CurrentEnvStoreFilePath, c.Args())
	if err != nil {
		log.Errorf("command failed: %s", err)
	}
	cmd.SetStdin(os.Stdin)
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)
	exitCode, err := cmd.RunAndReturnExitCode()
	if err != nil {
		log.Errorf("command failed: %s", err)
	}
	if err != nil && exitCode == 0 {
		exitCode = 1
	}
	os.Exit(exitCode)
	return nil
}

func createCommand(envStorePth string, args []string) (*command.Model, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	cmdEnvs, err := ReadAndEvaluateEnvs(envStorePth, &env.DefaultEnvironmentSource{})
	if err != nil {
		return nil, fmt.Errorf("failed to load EnvStore: %s", err)
	}

	cmdName := args[0]
	var cmdArgs []string
	if len(args) > 1 {
		cmdArgs = args[1:]
	}

	cmd := command.New(cmdName, cmdArgs...)
	cmd.SetEnvs(cmdEnvs...)
	return cmd, nil
}
