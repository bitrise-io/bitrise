package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/pathutil"
	ver "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)

const examplePluginGitURL = "https://github.com/bitrise-io/bitrise-plugins-example.git"
const analyticsPluginBinURL = "https://github.com/bitrise-io/bitrise-plugins-analytics/releases/download/0.9.1/analytics-Darwin-x86_64"

func TestIsLocalURL(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

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

func TestValidateVersion(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	t.Log("required min - pass")
	{
		requiredMin, err := ver.NewVersion("1.0.0")
		require.NoError(t, err)

		current, err := ver.NewVersion("1.0.1")
		require.NoError(t, err)

		err = validateVersion(*current, *requiredMin, nil)
		require.NoError(t, err)
	}

	t.Log("required min - fail")
	{
		requiredMin, err := ver.NewVersion("1.0.2")
		require.NoError(t, err)

		current, err := ver.NewVersion("1.0.1")
		require.NoError(t, err)

		err = validateVersion(*current, *requiredMin, nil)
		require.Error(t, err)
	}

	t.Log("required min + required max - pass")
	{
		requiredMin, err := ver.NewVersion("1.0.0")
		require.NoError(t, err)

		requiredMax, err := ver.NewVersion("1.0.2")
		require.NoError(t, err)

		current, err := ver.NewVersion("1.0.1")
		require.NoError(t, err)

		err = validateVersion(*current, *requiredMin, requiredMax)
		require.NoError(t, err)
	}

	t.Log("required min + required max - pass")
	{
		requiredMin, err := ver.NewVersion("1.0.0")
		require.NoError(t, err)

		requiredMax, err := ver.NewVersion("1.0.0")
		require.NoError(t, err)

		current, err := ver.NewVersion("1.0.0")
		require.NoError(t, err)

		err = validateVersion(*current, *requiredMin, requiredMax)
		require.NoError(t, err)
	}

	t.Log("required min + required max - min fail")
	{
		requiredMin, err := ver.NewVersion("1.0.1")
		require.NoError(t, err)

		requiredMax, err := ver.NewVersion("1.0.2")
		require.NoError(t, err)

		current, err := ver.NewVersion("1.0.0")
		require.NoError(t, err)

		err = validateVersion(*current, *requiredMin, requiredMax)
		require.Error(t, err)
	}

	t.Log("required min + required max - max fail")
	{
		requiredMin, err := ver.NewVersion("1.0.0")
		require.NoError(t, err)

		requiredMax, err := ver.NewVersion("1.0.1")
		require.NoError(t, err)

		current, err := ver.NewVersion("1.0.2")
		require.NoError(t, err)

		err = validateVersion(*current, *requiredMin, requiredMax)
		require.Error(t, err)
	}
}

func TestValidateRequirements(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	bitriseVersion, err := ver.NewVersion("1.0.0")
	require.NoError(t, err)

	envmanVersion, err := ver.NewVersion("1.0.0")
	require.NoError(t, err)

	stepmanVersion, err := ver.NewVersion("1.0.0")
	require.NoError(t, err)

	currentVersionMap := map[string]ver.Version{
		"bitrise": *bitriseVersion,
		"envman":  *envmanVersion,
		"stepman": *stepmanVersion,
	}

	t.Log("valid requirements")
	{
		requirements := []Requirement{
			Requirement{
				Tool:       "bitrise",
				MinVersion: "1.0.0",
				MaxVersion: "1.0.0",
			},
			Requirement{
				Tool:       "envman",
				MinVersion: "0.9.0",
				MaxVersion: "1.1.0",
			},
			Requirement{
				Tool:       "stepman",
				MinVersion: "1.0.0",
				MaxVersion: "1.0.0",
			},
		}

		err := validateRequirements(requirements, currentVersionMap)
		require.NoError(t, err)
	}

	t.Log("invalid requirements")
	{
		requirements := []Requirement{
			Requirement{
				Tool:       "bitrise",
				MinVersion: "1.0.0",
				MaxVersion: "1.0.0",
			},
			Requirement{
				Tool:       "envman",
				MinVersion: "1.1.0",
				MaxVersion: "1.1.0",
			},
			Requirement{
				Tool:       "stepman",
				MinVersion: "1.0.0",
				MaxVersion: "1.0.0",
			},
		}

		err := validateRequirements(requirements, currentVersionMap)
		require.Error(t, err)
	}
}

func TestDownloadPluginBin(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	t.Log("example plugin bin - ")
	{
		pluginBinURL := analyticsPluginBinURL
		destinationDir, err := pathutil.NormalizedOSTempDirPath("TestDownloadPluginBin")
		require.NoError(t, err)

		exist, err := pathutil.IsPathExists(destinationDir)
		require.NoError(t, err)
		if exist {
			err := os.RemoveAll(destinationDir)
			require.NoError(t, err)
		}

		require.NoError(t, os.MkdirAll(destinationDir, 0777))

		destinationPth := filepath.Join(destinationDir, "example")

		require.NoError(t, downloadPluginBin(pluginBinURL, destinationPth))

		exist, err = pathutil.IsPathExists(destinationPth)
		require.NoError(t, err)
		require.Equal(t, true, exist)
	}
}

func Test_isSourceURIChanged(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	for _, tt := range []struct {
		installed string
		new       string
		want      bool
	}{
		{installed: "https://github.com/bitrise-core/bitrise-plugins-analytics.git", new: "https://github.com/bitrise-core/bitrise-plugins-analytics.git", want: false},
		{installed: "https://github.com/bitrise-core/bitrise-plugins-analytics.git", new: "https://github.com/bitrise-io/bitrise-plugins-analytics.git", want: false},
		{installed: "https://github.com/bitrise-core/bitrise-plugins-analytics.git", new: "https://github.com/myorg/bitrise-plugins-analytics.git", want: true},
		{installed: "https://github.com/bitrise-core/bitrise-plugins-analytics.git", new: "https://github.com/bitrise-team/bitrise-plugins-analytics.git", want: true},
		{installed: "https://github.com/bitrise-custom-org/bitrise-plugins-analytics.git", new: "https://github.com/bitrise-core/bitrise-plugins-analytics.git", want: true},
		{installed: "https://github.com/bitrise-custom-org/bitrise-plugins-analytics.git", new: "https://github.com/bitrise-io/bitrise-plugins-analytics.git", want: true},
	} {
		t.Run("", func(t *testing.T) {
			start := time.Now().UnixNano()
			defer func(s int64) {
				debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
			}(start)

			if got := isSourceURIChanged(tt.installed, tt.new); got != tt.want {
				t.Log(tt.installed, tt.new)
				t.Errorf("isSourceURIChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}
