package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/stretchr/testify/require"
)

func cleanupStepmanHome() (func() error, error) {
	stepmanHomeDir := stepman.GetStepmanDirPath()
	tmpDir, err := pathutil.NormalizedOSTempDirPath("_tmp_stepman_home_")
	if err != nil {
		return nil, err
	}

	if err := command.CopyDir(stepmanHomeDir, tmpDir, true); err != nil {
		return nil, err
	}

	if err := os.RemoveAll(stepmanHomeDir); err != nil {
		return nil, err
	}

	restoreFunc := func() error {
		if err := os.RemoveAll(stepmanHomeDir); err != nil {
			return err
		}
		return command.CopyDir(tmpDir, stepmanHomeDir, true)
	}

	return restoreFunc, nil
}

func cloneDeafultStepLib() (string, error) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("_tmp_step_lib_")
	if err != nil {
		return "", err
	}

	repo, err := git.New(tmpDir)
	if err != nil {
		return "", err
	}

	cloneCmd := repo.Clone(defaultLibraryURI)
	if out, err := cloneCmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return "", fmt.Errorf("%s failed: %s", cloneCmd.PrintableCommandArgs(), out)
		}
		return "", fmt.Errorf("%s failed: %s", cloneCmd.PrintableCommandArgs(), err)
	}

	return tmpDir, nil
}

func TestActivateLocalStepLibStep(t *testing.T) {
	// backup and clean stepman home
	restoreFunc, err := cleanupStepmanHome()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, restoreFunc())
	}()

	// setup a local collection
	stepLibDir, err := cloneDeafultStepLib()
	require.NoError(t, err)

	out, err := command.New(binPath(), "setup", "-c", "file://"+stepLibDir).RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)

	// activate step from local collection
	tmpDir, err := pathutil.NormalizedOSTempDirPath("_tmp_step_lib_")
	require.NoError(t, err)

	out, err = command.New(
		binPath(), "activate",
		"-c", "file://"+stepLibDir,
		"--id", "script",
		"--path", tmpDir,
	).RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}
