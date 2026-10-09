package plugins

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

const examplePluginGitURL = "https://github.com/bitrise-io/bitrise-plugins-example.git"
const testPluginBinURL = "https://github.com/bitrise-io/bitrise-plugins-step/releases/download/0.10.4/bitrise-plugins-step-Darwin-arm64"

func TestIsLocalURL(t *testing.T) {
	t.Log("local url - absolute")
	{
		require.Equal(t, true, isLocalURL("/usr/bin"))
	}

	t.Log("local url - relative")
	{
		require.Equal(t, true, isLocalURL("../usr/bin"))
	}

	t.Log("local url - with prefix: file://")
	{
		require.Equal(t, true, isLocalURL("file:///usr/bin"))
	}

	t.Log("local url - relative with prefix: file://")
	{
		require.Equal(t, true, isLocalURL("file://./../usr/bin"))
	}

	t.Log("remote url")
	{
		require.Equal(t, false, isLocalURL("https://bitrise.io"))
	}

	t.Log("remote url- git ssh url")
	{
		require.Equal(t, false, isLocalURL("git@github.com:bitrise-io/bitrise.git"))
	}
}

func TestDownloadPluginBin(t *testing.T) {
	t.Log("example plugin bin - ")
	{
		destinationDir, err := pathutil.NormalizedOSTempDirPath("TestDownloadPluginBin")
		require.NoError(t, err)

		exist, err := pathutil.IsPathExists(destinationDir)
		require.NoError(t, err)
		if exist {
			err := os.RemoveAll(destinationDir)
			require.NoError(t, err)
		}

		require.NoError(t, os.MkdirAll(destinationDir, 0755))

		destinationPth := filepath.Join(destinationDir, "example")

		require.NoError(t, downloadPluginBin(testPluginBinURL, destinationPth))

		exist, err = pathutil.IsPathExists(destinationPth)
		require.NoError(t, err)
		require.Equal(t, true, exist)
	}
}

func Test_isSourceURIChanged(t *testing.T) {
	for _, tt := range []struct {
		installed string
		new       string
		want      bool
	}{
		{installed: "https://github.com/bitrise-core/bitrise-plugins-step.git", new: "https://github.com/bitrise-core/bitrise-plugins-step.git", want: false},
		{installed: "https://github.com/bitrise-core/bitrise-plugins-step.git", new: "https://github.com/bitrise-io/bitrise-plugins-step.git", want: false}, // resolves to same real URL
		{installed: "https://github.com/bitrise-core/bitrise-plugins-step.git", new: "https://github.com/myorg/bitrise-plugins-step.git", want: true},
		{installed: "https://github.com/bitrise-core/bitrise-plugins-step.git", new: "https://github.com/bitrise-team/bitrise-plugins-step.git", want: true},
		{installed: "https://github.com/bitrise-custom-org/bitrise-plugins-step.git", new: "https://github.com/bitrise-core/bitrise-plugins-step.git", want: true},
		{installed: "https://github.com/bitrise-custom-org/bitrise-plugins-step.git", new: "https://github.com/bitrise-io/bitrise-plugins-step.git", want: true},
	} {
		t.Run("", func(t *testing.T) {
			if got := isSourceURIChanged(tt.installed, tt.new); got != tt.want {
				t.Log(tt.installed, tt.new)
				t.Errorf("isSourceURIChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}
