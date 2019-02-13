package integration

import (
	"os"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	t.Log("remote library")
	{
		out, err := command.New(binPath(), "delete", "-c", defaultLibraryURI).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "setup", "-c", defaultLibraryURI).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "update", "-c", defaultLibraryURI).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("local library")
	{
		tmpDir, err := pathutil.NormalizedOSTempDirPath("__library__")
		require.NoError(t, err)
		defer func() {
			require.NoError(t, os.RemoveAll(tmpDir))
		}()
		repo, err := git.New(tmpDir)
		require.NoError(t, err)
		require.NoError(t, repo.Clone(defaultLibraryURI).Run())

		out, err := command.New(binPath(), "delete", "-c", "file://"+tmpDir).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "setup", "-c", "file://"+tmpDir).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "update", "-c", tmpDir).RunAndReturnTrimmedCombinedOutput()
		require.Error(t, err, out)

		out, err = command.New(binPath(), "update", "-c", "file://"+tmpDir).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)

		out, err = command.New(binPath(), "delete", "-c", "file://"+tmpDir).RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}
