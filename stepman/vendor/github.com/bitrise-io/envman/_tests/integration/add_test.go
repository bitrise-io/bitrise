package integration

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func addCommand(key, value, envstore string) *command.Model {
	return command.New(binPath(), "-l", "debug", "-p", envstore, "add", "--key", key, "--value", value)
}

func addFileCommand(key, pth, envstore string) *command.Model {
	return command.New(binPath(), "-l", "debug", "-p", envstore, "add", "--key", key, "--valuefile", pth)
}

func addPipeCommand(key string, reader io.Reader, envstore string) *command.Model {
	return command.New(binPath(), "-l", "debug", "-p", envstore, "add", "--key", key).SetStdin(reader)
}

func TestAdd(t *testing.T) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("__envman__")
	require.NoError(t, err)

	envstore := filepath.Join(tmpDir, ".envstore")
	f, err := os.Create(envstore)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	t.Log("add flag value")
	{
		out, err := addCommand("KEY", "value", envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		cont, err := fileutil.ReadStringFromFile(envstore)
		require.NoError(t, err, out)
		require.Equal(t, "envs:\n- KEY: value\n", cont)
	}

	t.Log("add file flag value")
	{
		pth := filepath.Join(tmpDir, "file")
		require.NoError(t, fileutil.WriteStringToFile(pth, "some content"))

		out, err := addFileCommand("KEY", pth, envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		cont, err := fileutil.ReadStringFromFile(envstore)
		require.NoError(t, err, out)
		require.Equal(t, "envs:\n- KEY: some content\n", cont)
	}

	t.Log("add piped value")
	{
		out, err := addPipeCommand("KEY", strings.NewReader("some piped value"), envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		cont, err := fileutil.ReadStringFromFile(envstore)
		require.NoError(t, err, out)
		require.Equal(t, "envs:\n- KEY: some piped value\n", cont)
	}

	t.Log("add empty piped value")
	{
		out, err := addPipeCommand("KEY", strings.NewReader(""), envstore).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		cont, err := fileutil.ReadStringFromFile(envstore)
		require.NoError(t, err, out)
		require.Equal(t, "envs:\n- KEY: \"\"\n", cont)
	}
}
