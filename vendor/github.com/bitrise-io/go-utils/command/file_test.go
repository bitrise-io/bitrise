package command

import (
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestCopyFileErrorIfDirectory(t *testing.T) {
	t.Log("It fails if source is a directory")
	{
		tmpFolder, err := pathutil.NormalizedOSTempDirPath("_tmp")
		require.NoError(t, err)
		require.EqualError(t, CopyFile(tmpFolder, "./nothing/whatever"), "Source is a directory: "+tmpFolder)
	}
}
