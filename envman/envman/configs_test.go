package envman

import (
	"os"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func TestGetConfigs(t *testing.T) {
	// fake home, to save the configs into
	fakeHomePth, err := pathutil.NormalizedOSTempDirPath("_FAKE_HOME")
	t.Logf("fakeHomePth: %s", fakeHomePth)
	require.NoError(t, err)
	originalHome := os.Getenv("HOME")
	defer func() {
		require.NoError(t, os.Setenv("HOME", originalHome))
		require.NoError(t, os.RemoveAll(fakeHomePth))
	}()
	require.Equal(t, nil, os.Setenv("HOME", fakeHomePth))

	configPth := getEnvmanConfigsFilePath()
	t.Logf("configPth: %s", configPth)

	// --- TESTING

	baseConf, err := GetConfigs()
	t.Logf("baseConf: %#v", baseConf)
	require.NoError(t, err)
	require.Equal(t, defaultEnvBytesLimitInKB, baseConf.EnvBytesLimitInKB)
	require.Equal(t, defaultEnvListBytesLimitInKB, baseConf.EnvListBytesLimitInKB)

	// modify it
	baseConf.EnvBytesLimitInKB = 123
	baseConf.EnvListBytesLimitInKB = 321

	// save to file
	require.NoError(t, saveConfigs(baseConf))

	// read it back
	configs, err := GetConfigs()
	t.Logf("configs: %#v", configs)
	require.NoError(t, err)
	require.Equal(t, configs, baseConf)
	require.Equal(t, 123, configs.EnvBytesLimitInKB)
	require.Equal(t, 321, configs.EnvListBytesLimitInKB)

	// delete the tmp config file
	require.NoError(t, os.Remove(configPth))
}

func TestGetConfigsEnvVarOverride(t *testing.T) {
	// fake home, to control the configs file
	fakeHomePth, err := pathutil.NormalizedOSTempDirPath("_FAKE_HOME")
	require.NoError(t, err)
	originalHome := os.Getenv("HOME")
	defer func() {
		require.NoError(t, os.Setenv("HOME", originalHome))
		require.NoError(t, os.RemoveAll(fakeHomePth))
	}()
	require.NoError(t, os.Setenv("HOME", fakeHomePth))

	unsetEnvs := func() {
		require.NoError(t, os.Unsetenv(envBytesLimitInKBEnvKey))
		require.NoError(t, os.Unsetenv(envListBytesLimitInKBEnvKey))
	}
	defer unsetEnvs()

	t.Run("env vars override the defaults when no config file exists", func(t *testing.T) {
		unsetEnvs()
		require.NoError(t, os.Setenv(envBytesLimitInKBEnvKey, "111"))
		require.NoError(t, os.Setenv(envListBytesLimitInKBEnvKey, "222"))

		configs, err := GetConfigs()
		require.NoError(t, err)
		require.Equal(t, 111, configs.EnvBytesLimitInKB)
		require.Equal(t, 222, configs.EnvListBytesLimitInKB)
	})

	t.Run("env vars take precedence over the config file", func(t *testing.T) {
		unsetEnvs()
		require.NoError(t, saveConfigs(ConfigsModel{EnvBytesLimitInKB: 123, EnvListBytesLimitInKB: 321}))
		defer func() { require.NoError(t, os.Remove(getEnvmanConfigsFilePath())) }()

		require.NoError(t, os.Setenv(envBytesLimitInKBEnvKey, "111"))
		require.NoError(t, os.Setenv(envListBytesLimitInKBEnvKey, "222"))

		configs, err := GetConfigs()
		require.NoError(t, err)
		require.Equal(t, 111, configs.EnvBytesLimitInKB)
		require.Equal(t, 222, configs.EnvListBytesLimitInKB)
	})

	t.Run("only the set env var overrides, the other falls back to the config file", func(t *testing.T) {
		unsetEnvs()
		require.NoError(t, saveConfigs(ConfigsModel{EnvBytesLimitInKB: 123, EnvListBytesLimitInKB: 321}))
		defer func() { require.NoError(t, os.Remove(getEnvmanConfigsFilePath())) }()

		require.NoError(t, os.Setenv(envBytesLimitInKBEnvKey, "111"))

		configs, err := GetConfigs()
		require.NoError(t, err)
		require.Equal(t, 111, configs.EnvBytesLimitInKB)
		require.Equal(t, 321, configs.EnvListBytesLimitInKB)
	})

	t.Run("invalid env var value returns an error", func(t *testing.T) {
		unsetEnvs()
		require.NoError(t, os.Setenv(envBytesLimitInKBEnvKey, "not-a-number"))

		_, err := GetConfigs()
		require.Error(t, err)
	})
}
