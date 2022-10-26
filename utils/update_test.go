package utils

import (
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsUpdateAvailable(t *testing.T) {
	t.Log("simple compare versions - ture")
	{
		stepInfo1 := stepmanModels.StepInfoModel{
			Version:       "1.0.0",
			LatestVersion: "1.1.0",
		}

		updateAvailable, err := IsUpdateAvailable(stepInfo1.Version, stepInfo1.LatestVersion)
		require.NoError(t, err)
		require.Equal(t, true, updateAvailable)
	}

	t.Log("simple compare versions - false")
	{
		stepInfo1 := stepmanModels.StepInfoModel{
			Version:       "1.0.0",
			LatestVersion: "1.0.0",
		}

		updateAvailable, err := IsUpdateAvailable(stepInfo1.Version, stepInfo1.LatestVersion)
		require.NoError(t, err)
		require.Equal(t, false, updateAvailable)
	}

	t.Log("issue - no latest - false")
	{
		stepInfo1 := stepmanModels.StepInfoModel{
			Version:       "1.0.0",
			LatestVersion: "",
		}

		updateAvailable, err := IsUpdateAvailable(stepInfo1.Version, stepInfo1.LatestVersion)
		require.NoError(t, err)
		require.Equal(t, false, updateAvailable)
	}

	t.Log("issue - no current - false")
	{
		stepInfo1 := stepmanModels.StepInfoModel{
			Version:       "",
			LatestVersion: "1.0.0",
		}

		updateAvailable, err := IsUpdateAvailable(stepInfo1.Version, stepInfo1.LatestVersion)
		require.True(t, err != nil)
		require.Equal(t, false, updateAvailable)
	}
}
