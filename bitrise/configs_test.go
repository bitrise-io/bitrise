package bitrise

import (
	"os"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestNewConfigFromBytes(t *testing.T) {
	t.Log("Check config with string true")
	{
		configStr := `opt_out_analytics: "true"`

		config, err := NewConfigFromBytes([]byte(configStr))
		require.Equal(t, nil, err)

		require.Equal(t, true, config.OptOutAnalytics)
	}

	t.Log("Check config with bolean true")
	{
		configStr := `opt_out_analytics: true`

		config, err := NewConfigFromBytes([]byte(configStr))
		require.Equal(t, nil, err)

		require.Equal(t, true, config.OptOutAnalytics)
	}

	t.Log("Check config with bolean false")
	{
		configStr := `opt_out_analytics: false`

		config, err := NewConfigFromBytes([]byte(configStr))
		require.Equal(t, nil, err)

		require.Equal(t, false, config.OptOutAnalytics)
	}

	t.Log("Check config with string false")
	{
		configStr := `opt_out_analytics: "false"`

		config, err := NewConfigFromBytes([]byte(configStr))
		require.Equal(t, nil, err)

		require.Equal(t, false, config.OptOutAnalytics)
	}

	t.Log("Check config with integer")
	{
		configStr := `opt_out_analytics: 0`

		_, err := NewConfigFromBytes([]byte(configStr))
		require.NotEqual(t, nil, err)
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
