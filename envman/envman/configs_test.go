package envman

import (
	"os"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestCheckIfConfigsSaved(t *testing.T) {
	configsPth := getEnvmanConfigsFilePath()
	exist, err := pathutil.IsPathExists(configsPth)
	require.Equal(t, nil, err)
	if exist {
		require.Equal(t, nil, os.RemoveAll(configsPth))
	}

	exist = CheckIfConfigsSaved()
	require.Equal(t, false, exist)

	require.Equal(t, nil, SaveDefaultConfigs())
	exist = CheckIfConfigsSaved()
	require.Equal(t, true, exist)
}

func TestGetConfigs(t *testing.T) {
	configsPth := getEnvmanConfigsFilePath()
	exist, err := pathutil.IsPathExists(configsPth)
	require.Equal(t, nil, err)
	if exist {
		require.Equal(t, nil, os.RemoveAll(configsPth))
	}

	_, err = GetConfigs()
	require.NotEqual(t, nil, err)

	require.Equal(t, nil, SaveDefaultConfigs())
	configs, err := GetConfigs()
	require.Equal(t, nil, err)
	require.Equal(t, defaultEnvBytesLimitInKB, configs.EnvBytesLimitInKB)
	require.Equal(t, defaultEnvListBytesLimitInKB, configs.EnvListBytesLimitInKB)
}

func TestSaveDefaultConfigs(t *testing.T) {
	require.Equal(t, nil, SaveDefaultConfigs())
}
