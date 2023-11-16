package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
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
	if hooks.DoOnBuildStart == "" {
		return nil
	}

	log.Print()
	log.Infof("Run build start hook")
	log.Print(hooks.DoOnBuildStart)
	log.Print()

	cmd := exec.Command(hooks.DoOnBuildStart)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runBuildEndHooks(hooks configs.AgentHooks) error {
	if hooks.DoOnBuildEnd == "" {
		return nil
	}

	log.Print()
	log.Infof("Run build end hook")
	log.Print(hooks.DoOnBuildEnd)
	log.Print()

	cmd := exec.Command(hooks.DoOnBuildEnd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
