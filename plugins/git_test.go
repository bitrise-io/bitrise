package plugins

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestFilterVersionTags(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	t.Log("single version tag")
	{
		versionTags := filterVersionTags([]string{"1.0.0"})
		require.Equal(t, 1, len(versionTags))
		require.Equal(t, "1.0.0", versionTags[0].String())
	}

	t.Log("version tag list")
	{
		versionTags := filterVersionTags([]string{"1.0.0", "1.1.0", "1.1.1"})
		require.Equal(t, 3, len(versionTags))
		require.Equal(t, "1.0.0", versionTags[0].String())
		require.Equal(t, "1.1.0", versionTags[1].String())
		require.Equal(t, "1.1.1", versionTags[2].String())
	}

	t.Log("non version tag")
	{
		versionTags := filterVersionTags([]string{"release"})
		require.Equal(t, 0, len(versionTags))
	}

	t.Log("version tag + non version tag")
	{
		versionTags := filterVersionTags([]string{"1.0.0", "release"})
		require.Equal(t, 1, len(versionTags))
		require.Equal(t, "1.0.0", versionTags[0].String())
	}
}

func TestClonePluginSrc(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	t.Log("example plugin - latest version")
	{
		pluginSource := examplePluginGitURL
		versionTag := ""
		destinationDir, err := pathutil.NormalizedOSTempDirPath("TestClonePluginSrc")
		require.NoError(t, err)

		exist, err := pathutil.IsPathExists(destinationDir)
		require.NoError(t, err)
		if exist {
			err := os.RemoveAll(destinationDir)
			require.NoError(t, err)
		}

		version, err := GitCloneAndCheckoutVersionOrLatestVersion(destinationDir, pluginSource, versionTag)
		require.NoError(t, err)
		require.NotNil(t, version)

		exist, err = pathutil.IsPathExists(destinationDir)
		require.NoError(t, err)
		require.Equal(t, true, exist)
	}

	t.Log("example plugin - 0.9.0 version")
	{
		pluginSource := examplePluginGitURL
		versionTag := "0.9.0"
		destinationDir, err := pathutil.NormalizedOSTempDirPath("TestClonePluginSrc")
		require.NoError(t, err)

		exist, err := pathutil.IsPathExists(destinationDir)
		require.NoError(t, err)
		if exist {
			err := os.RemoveAll(destinationDir)
			require.NoError(t, err)
		}

		version, err := GitCloneAndCheckoutVersionOrLatestVersion(destinationDir, pluginSource, versionTag)
		require.NoError(t, err)
		require.NotNil(t, version)
		require.Equal(t, "0.9.0", version)

		exist, err = pathutil.IsPathExists(destinationDir)
		require.NoError(t, err)
		require.Equal(t, true, exist)
	}
}
