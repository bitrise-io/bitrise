package plugins

import (
	"testing"

	ver "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)


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

func TestValidateVersion(t *testing.T) {
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
