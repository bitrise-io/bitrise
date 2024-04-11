package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/log/logwriter"
	"github.com/bitrise-io/colorstring"
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
		{
			dir:    dirs.HTMLReportDir,
			envKey: configs.BitriseHtmlReportDirEnvKey,
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
	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	logWriter := logwriter.NewLogWriter(logger)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
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
	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	logWriter := logwriter.NewLogWriter(logger)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	return cmd.Run()
}

func cleanupDirs(dirs []string) error {
	if len(dirs) == 0 {
		return nil
	}

	log.Print()
	log.Infof("Run directory cleanups")
	for _, dir := range dirs {
		expandedPath := os.ExpandEnv(dir)
		if expandedPath == "" {
			continue
		}
		expandedPath, err := pathutil.ExpandTilde(expandedPath)
		if err != nil {
			return fmt.Errorf("cleaning up %s: %w", dir, err)
		}
		if expandedPath == "" {
			continue
		}
		absPath, err := pathutil.AbsPath(expandedPath)
		if err != nil {
			return fmt.Errorf("cleaning up %s: %w", dir, err)
		}
		if err := os.RemoveAll(absPath); err != nil {
			return fmt.Errorf("cleaning up %s: %w", dir, err)
		}
		log.Donef("- Cleaned %s", colorstring.Cyan(expandedPath))
	}

	return nil
}
