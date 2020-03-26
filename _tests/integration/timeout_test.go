package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func Test_TimeoutTest(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

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

		require.EqualError(t, err, "exit status 1", out)

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

		require.NoError(t, err, "exit status 1", out)

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
