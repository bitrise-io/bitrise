package local

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bitrise-io/bitrise/v2/configs"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/log/logwriter"
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

func runBuildHook(hookCmd, label string) error {
	if hookCmd == "" {
		return nil
	}

	log.Print()
	log.Infof("Run %s hook", label)
	log.Print(hookCmd)
	log.Print()

	cmd := exec.Command(hookCmd)
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
		log.Donef("- Cleaned %s", colorstring.Cyan("%s", expandedPath))
	}

	return nil
}
