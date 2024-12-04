package integration

import (
	"testing"

	"github.com/bitrise-io/bitrise/plugins"
	"github.com/stretchr/testify/require"
)



const examplePluginGitURL = "https://github.com/bitrise-io/bitrise-plugins-example.git"

func TestDownloadPluginBin(t *testing.T) {
	t.Log("example plugin bin - specific version")
	{
		_, _, err := plugins.InstallPlugin(examplePluginGitURL, "0.9.0")
		require.NoError(t, err)
	}

	t.Log("example plugin bin - latest version")
	{
		_, _, err := plugins.InstallPlugin(examplePluginGitURL, "")
		require.NoError(t, err)
	}
}
