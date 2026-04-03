package docker

import (
	"fmt"

	"github.com/bitrise-io/go-utils/command"
)

type RunningContainer struct {
	ID   string
	Name string
}

func (rc *RunningContainer) Destroy() error {
	_, err := command.New("docker", "rm", "--force", "--volumes", rc.Name).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("remove docker container: %w", err)
	}
	return nil
}

func (rc *RunningContainer) ExecuteCommandArgs(envs []string) []string {
	args := []string{"exec"}

	for _, env := range envs {
		args = append(args, "-e", env)
	}

	args = append(args, rc.Name)

	return args
}
