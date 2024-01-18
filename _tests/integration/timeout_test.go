//go:build linux_and_mac
// +build linux_and_mac

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func Test_TimeoutTest(t *testing.T) {
	configPth := "timeout_test_bitrise.yml"

	t.Log("Step with timeout")
	{
		tmpDir, err := pathutil.NormalizedOSTempDirPath("__timeout_test__")
		require.NoError(t, err)
		defer func() {
			require.NoError(t, os.RemoveAll(tmpDir))
		}()

		testFilePth1 := filepath.Join(tmpDir, "file1")
		testFilePth2 := filepath.Join(tmpDir, "file2")

		envs := []string{
			fmt.Sprintf("TIMEOUT_TEST_FILE_PTH_1=%s", testFilePth1),
			fmt.Sprintf("TIMEOUT_TEST_FILE_PTH_2=%s", testFilePth2),
		}
		cmd := command.New(binPath(), "run", "timeout", "--config", configPth)
		cmd.AppendEnvs(envs...)

		out, err := cmd.RunAndReturnTrimmedCombinedOutput()

		require.EqualError(t, err, "exit status 91", out)

		t.Log("Should exist")
		{
			exist, err := pathutil.IsPathExists(testFilePth1)
			require.NoError(t, err)
			require.Equal(t, true, exist)
		}

		t.Log("Should NOT exist")
		{
			exist, err := pathutil.IsPathExists(testFilePth2)
			require.NoError(t, err)
			require.Equal(t, false, exist)
		}
	}

	t.Log("Multiple steps with timeout")
	{
		tmpDir, err := pathutil.NormalizedOSTempDirPath("__multiple_timeout_test__")
		require.NoError(t, err)
		defer func() {
			require.NoError(t, os.RemoveAll(tmpDir))
		}()

		testFilePth1 := filepath.Join(tmpDir, "file1")
		testFilePth2 := filepath.Join(tmpDir, "file2")

		envs := []string{
			fmt.Sprintf("TIMEOUT_TEST_FILE_PTH_1=%s", testFilePth1),
			fmt.Sprintf("TIMEOUT_TEST_FILE_PTH_2=%s", testFilePth2),
		}
		cmd := command.New(binPath(), "run", "multiple_timeout", "--config", configPth)
		cmd.AppendEnvs(envs...)

		out, err := cmd.RunAndReturnTrimmedCombinedOutput()

		require.NoError(t, err, out)

		t.Log("Should exist")
		{
			exist, err := pathutil.IsPathExists(testFilePth1)
			require.NoError(t, err)
			require.Equal(t, true, exist)
		}

		t.Log("Should existt")
		{
			exist, err := pathutil.IsPathExists(testFilePth2)
			require.NoError(t, err)
			require.Equal(t, true, exist)
		}
	}
}
