package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func unsetCommand(key, envstore string) *command.Model {
	return command.New(binPath(), "-p", envstore, "unset", "--key", key)
}

func runCommand(cmd, envstore string) *command.Model {
	return command.New(binPath(), "-p", envstore, "run", cmd)
}

func TestUnset(t *testing.T) {
	t.Log("only unset on an empty envstore")
	{
		// create a fully empty envstore
		tmpDir, err := pathutil.NormalizedOSTempDirPath("__envman__")
		require.NoError(t, err)

		envstore := filepath.Join(tmpDir, ".envstore")
		f, err := os.Create(envstore)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		randomEnvKEY := "DONOTEXPORT"

		// unset DONOTEXPORT env
		out, err := unsetCommand(randomEnvKEY, envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		// run env command through envman and see the exported env's list
		out, err = runCommand("env", envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		// check if the env is surely not exported
		if strings.Contains(out, randomEnvKEY) {
			t.Errorf("env is exported however it should be unset, complete list of exported envs:\n%s\n", out)
		}
	}

	t.Log("add env then unset it in an empty envstore")
	{
		// create a fully empty envstore
		tmpDir, err := pathutil.NormalizedOSTempDirPath("__envman__")
		require.NoError(t, err)

		envstore := filepath.Join(tmpDir, ".envstore")
		f, err := os.Create(envstore)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		randomEnvKEY := "DONOTEXPORT"

		// add DONOTEXPORT env
		out, err := addCommand(randomEnvKEY, "sample value", envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		// unset DONOTEXPORT env
		out, err = unsetCommand(randomEnvKEY, envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		// run env command through envman and see the exported env's list
		out, err = runCommand("env", envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		// check if the env is surely not exported
		if strings.Contains(out, randomEnvKEY) {
			t.Errorf("env is exported however it should be unset, complete list of exported envs:\n%s\n", out)
		}
	}

	t.Log("set env externally then only unset on an empty envstore")
	{
		// create a fully empty envstore
		tmpDir, err := pathutil.NormalizedOSTempDirPath("__envman__")
		require.NoError(t, err)

		envstore := filepath.Join(tmpDir, ".envstore")
		f, err := os.Create(envstore)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		randomEnvKEY := "DONOTEXPORT"

		require.NoError(t, os.Setenv(randomEnvKEY, "value"))

		// unset DONOTEXPORT env
		out, err := unsetCommand(randomEnvKEY, envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		// run env command through envman and see the exported env's list
		out, err = runCommand("env", envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		// check if the env is surely not exported
		if strings.Contains(out, randomEnvKEY) {
			t.Errorf("env is exported however it should be unset, complete list of exported envs:\n%s\n", out)
		}
	}

	t.Log("set env externally then add env then unset it in an empty envstore")
	{
		// create a fully empty envstore
		tmpDir, err := pathutil.NormalizedOSTempDirPath("__envman__")
		require.NoError(t, err)

		envstore := filepath.Join(tmpDir, ".envstore")
		f, err := os.Create(envstore)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		controlEnvKey := "EXPORT_THIS"
		randomEnvKEY := "DONOTEXPORT"

		require.NoError(t, os.Setenv(controlEnvKey, "value"))
		require.NoError(t, os.Setenv(randomEnvKEY, "value"))

		// add DONOTEXPORT env
		out, err := addCommand(randomEnvKEY, "sample value", envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		// unset DONOTEXPORT env
		out, err = unsetCommand(randomEnvKEY, envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		// run env command through envman and see the exported env's list
		out, err = runCommand("env", envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		// check if the env is surely not exported
		if !strings.Contains(out, controlEnvKey) {
			t.Errorf("env %s is not exported, complete list of exported envs:\n%s\n", controlEnvKey, out)
		}

		// check if the env is surely not exported
		if strings.Contains(out, randomEnvKEY) {
			t.Errorf("env is exported however it should be unset, complete list of exported envs:\n%s\n", out)
		}
	}
}