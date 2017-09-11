package integration

import (
	"testing"

	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	t.Log("remote library")
	{
		out, err := command.New(binPath(), "delete", "-c", defaultLibraryURI).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "setup", "-c", defaultLibraryURI).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "delete", "-c", defaultLibraryURI).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("local library")
	{
		tmpDir, err := pathutil.NormalizedOSTempDirPath("__library__")
		require.NoError(t, err)
		defer func() {
			require.NoError(t, os.RemoveAll(tmpDir))
		}()
		require.NoError(t, git.Clone(defaultLibraryURI, tmpDir))

		out, err := command.New(binPath(), "delete", "-c", tmpDir).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "setup", "--local", "-c", tmpDir).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "delete", "-c", "file://"+tmpDir).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "setup", "--local", "-c", "file://"+tmpDir).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "delete", "-c", "file://"+tmpDir).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "setup", "-c", "file://"+tmpDir).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "delete", "-c", "file://"+tmpDir).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}
