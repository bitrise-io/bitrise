package envman

import (
	"errors"
	"os"
	"os/exec"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/models"
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

// RunCmd ...
func RunCmd(commandToRun CommandModel) (int, error) {
	cmdEnvs, err := commandEnvs(commandToRun.Environments)
	if err != nil {
		return 1, err
	}

	cmd := exec.Command(commandToRun.Command, commandToRun.Argumentums...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = cmdEnvs

	log.Debugln("Command to execute:", cmd)

	cmdExitCode := 0
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus, ok := exitError.Sys().(syscall.WaitStatus)
			if !ok {
				return 1, errors.New("Failed to cast exit status")
			}
			cmdExitCode = waitStatus.ExitStatus()
		}
		return cmdExitCode, err
	}

	return 0, nil
}
