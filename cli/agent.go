package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/go-utils/pathutil"
)

func registerAgentOverrides(dirs configs.BitriseDirs) error {
	params := []struct {
		dir    string
		envKey string
	}{
		{
			dir:    dirs.BitriseDataHomeDir,
			envKey: configs.BitriseDataHomeDirEnvKey,
		},
		{
			dir:    dirs.SourceDir,
			envKey: configs.BitriseSourceDirEnvKey,
		},
		{
			dir:    dirs.DeployDir,
			envKey: configs.BitriseDeployDirEnvKey,
		},
		{
			dir:    dirs.TestDeployDir,
			envKey: configs.BitriseTestDeployDirEnvKey,
		},
	}
	for _, param := range params {
		err := pathutil.EnsureDirExist(param.dir)
		if err != nil {
			return fmt.Errorf("can't create %s: %w", param.envKey, err)
		}
		err = os.Setenv(param.envKey, param.dir)
		if err != nil {
			return fmt.Errorf("set %s: %w", param.envKey, err)
		}
	}

	return nil
}

func runBuildStartHooks(hooks configs.AgentHooks) error {
	if hooks.DoOnWorkflowStart == "" {
		return nil
	}

	log.Print()
	log.Info("Run build start hook")

	cmd := exec.Command(hooks.DoOnWorkflowStart)



	return nil
}

func runBuildEndHooks(hooks configs.AgentHooks) error {
	return nil
}
