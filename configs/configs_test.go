package configs

import (
	"os"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestNewConfigFromBytes(t *testing.T) {
	t.Log("Check config with string true")
	{
		configStr := `is_analytics_disabled: "true"`
		config, err := NewConfigFromBytes([]byte(configStr))

		require.Error(t, err)
		require.Equal(t, false, config.IsAnalyticsDisabled)
	}

	t.Log("Check config with bolean true")
	{
		configStr := `is_analytics_disabled: true`
		config, err := NewConfigFromBytes([]byte(configStr))

		require.Equal(t, nil, err)
		require.Equal(t, true, config.IsAnalyticsDisabled)
	}

	t.Log("Check config with bolean false")
	{
		configStr := `is_analytics_disabled: false`
		config, err := NewConfigFromBytes([]byte(configStr))

		require.Equal(t, nil, err)
		require.Equal(t, false, config.IsAnalyticsDisabled)
	}

	t.Log("Check config with string false")
	{
		configStr := `is_analytics_disabled: "false"`
		config, err := NewConfigFromBytes([]byte(configStr))

		require.Error(t, err)
		require.Equal(t, false, config.IsAnalyticsDisabled)
	}

	t.Log("Check config with integer")
	{
		configStr := `is_analytics_disabled: 0`
		config, err := NewConfigFromBytes([]byte(configStr))
		require.Error(t, err)
		require.Equal(t, false, config.IsAnalyticsDisabled)
	}
}

func TestSetupForVersionChecks(t *testing.T) {
	fakeHomePth, err := pathutil.NormalizedOSTempDirPath("_FAKE_HOME")
	require.Equal(t, nil, err)
	originalHome := os.Getenv("HOME")

	defer func() {
		require.Equal(t, nil, os.Setenv("HOME", originalHome))
		require.Equal(t, nil, os.RemoveAll(fakeHomePth))
	}()

	require.Equal(t, nil, os.Setenv("HOME", fakeHomePth))

	require.Equal(t, false, CheckIsSetupWasDoneForVersion("0.9.7"))

	require.Equal(t, nil, SaveSetupSuccessForVersion("0.9.7"))

	require.Equal(t, true, CheckIsSetupWasDoneForVersion("0.9.7"))

	require.Equal(t, false, CheckIsSetupWasDoneForVersion("0.9.8"))
}
