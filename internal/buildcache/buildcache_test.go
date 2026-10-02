package buildcache

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/v3/analytics"
	"github.com/bitrise-io/bitrise/v3/configs"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnabled(t *testing.T) {
	t.Run("off by default", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "")
		assert.False(t, enabled(nil, EnvActivateAll))
	})
	t.Run("process env", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "true")
		assert.True(t, enabled(nil, EnvActivateAll))
	})
	t.Run("build env", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "")
		assert.True(t, enabled(buildEnvs(EnvActivateAll, "true"), EnvActivateAll))
	})
	t.Run("the two opt-ins are independent", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "")
		assert.False(t, enabled(buildEnvs(EnvActivateGradleMirrors, "true"), EnvActivateAll))
	})
}

func TestCacheEnvs(t *testing.T) {
	envs := buildEnvs(
		"BITRISE_BUILD_CACHE_AUTH_TOKEN", "auth",
		servicesTokenKey, "jwt",
		buildHubVMTokenKey, "vm",
		buildHubVMTokenURLKey, "https://hub",
		"BITRISE_BUILD_API_TOKEN", "unrelated",
		"GRADLE_ENCRYPTION_KEY", "unrelated",
	)

	assert.Equal(t, []string{
		"BITRISE_BUILD_CACHE_AUTH_TOKEN=auth",
		servicesTokenKey + "=jwt",
		buildHubVMTokenKey + "=vm",
		buildHubVMTokenURLKey + "=https://hub",
	}, cacheEnvs(envs))
}

func TestActivateIfEnabled_InstallsOnceAndRunsBothCommands(t *testing.T) {
	setPin(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv(analytics.StepExecutionIDEnvKey, "")
	t.Setenv(envDisableHostsOverride, "")
	t.Setenv(EnvActivateAll, "true")
	t.Setenv(EnvActivateGradleMirrors, "true")
	t.Setenv("BITRISE_BUILD_API_TOKEN", "")
	out := filepath.Join(t.TempDir(), "calls")
	installFakeCLI(t, `echo "$@ proxy=$BITRISE_MAVENCENTRAL_PROXY_ENABLED jwt=$BITRISEIO_BITRISE_SERVICES_ACCESS_TOKEN api=${BITRISE_BUILD_API_TOKEN:-absent}" >> `+out)

	ran := ActivateIfEnabled(testLogger(), buildEnvs(servicesTokenKey, "jwt", "BITRISE_BUILD_API_TOKEN", "secret"))

	require.True(t, ran)
	calls, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"activate gradle-mirrors -d proxy=true jwt= api=absent",
		"activate all --auto proxy= jwt=jwt api=absent",
	}, strings.Split(strings.TrimSpace(string(calls)), "\n"))

	target, err := os.Readlink(filepath.Join(configs.GetBitriseToolsDirPath(), binaryName))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(configs.GetBitriseHomeDirPath(), "build-cache", "3.14.1", binaryName), target)
}

func TestActivateIfEnabled_MirrorsOnly(t *testing.T) {
	setPin(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv(analytics.StepExecutionIDEnvKey, "")
	t.Setenv(envDisableHostsOverride, "")
	t.Setenv(EnvActivateAll, "")
	t.Setenv(EnvActivateGradleMirrors, "true")
	out := filepath.Join(t.TempDir(), "calls")
	installFakeCLI(t, `echo "$@" >> `+out)

	require.True(t, ActivateIfEnabled(testLogger(), nil))

	calls, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, "activate gradle-mirrors -d", strings.TrimSpace(string(calls)))
}

func TestActivateIfEnabled_HostsOverrideDisabledSkipsMirrors(t *testing.T) {
	setPin(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv(analytics.StepExecutionIDEnvKey, "")
	t.Setenv(EnvActivateAll, "")
	t.Setenv(EnvActivateGradleMirrors, "true")
	t.Setenv(envDisableHostsOverride, "true")
	out := filepath.Join(t.TempDir(), "calls")
	installFakeCLI(t, `echo called >> `+out)

	assert.False(t, ActivateIfEnabled(testLogger(), nil))
	assert.NoFileExists(t, out)
}

func TestActivateIfEnabled_AFailingCommandDoesNotStopTheOther(t *testing.T) {
	setPin(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv(analytics.StepExecutionIDEnvKey, "")
	t.Setenv(envDisableHostsOverride, "")
	t.Setenv(EnvActivateAll, "true")
	t.Setenv(EnvActivateGradleMirrors, "true")
	out := filepath.Join(t.TempDir(), "calls")
	installFakeCLI(t, `echo "$2" >> `+out+`; exit 3`)

	require.True(t, ActivateIfEnabled(testLogger(), nil))

	calls, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, []string{"gradle-mirrors", "all"}, strings.Split(strings.TrimSpace(string(calls)), "\n"))
}

func TestInstallCLI_RequiresPin(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, pin := range [][2]string{{"", ""}, {"3.14.1", ""}, {"../3.14.1", strings.Repeat("a", 64)}, {"3.14.1", "not-a-sha"}} {
		t.Setenv(EnvCLIVersion, pin[0])
		t.Setenv(EnvCLISHA256, pin[1])

		_, err := installCLI(context.Background(), testLogger())

		require.ErrorContains(t, err, "must be set by the VM setup")
	}
}

func TestActivateIfEnabled_SkipsNestedRuns(t *testing.T) {
	setPin(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv(EnvActivateAll, "true")
	t.Setenv(EnvActivateGradleMirrors, "true")
	t.Setenv(analytics.StepExecutionIDEnvKey, "outer-step")
	out := filepath.Join(t.TempDir(), "calls")
	installFakeCLI(t, `echo called >> `+out)

	ActivateIfEnabled(testLogger(), nil)

	assert.NoFileExists(t, out)
}

func installFakeCLI(t *testing.T, script string) {
	t.Helper()
	dir := filepath.Join(configs.GetBitriseHomeDirPath(), "build-cache", os.Getenv(EnvCLIVersion))
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, binaryName), []byte("#!/bin/sh\n"+script+"\n"), 0o755))
}

func setPin(t *testing.T) {
	t.Helper()
	t.Setenv(EnvCLIVersion, "3.14.1")
	t.Setenv(EnvCLISHA256, strings.Repeat("a", 64))
}

func buildEnvs(kv ...string) []envmanModels.EnvironmentItemModel {
	var envs []envmanModels.EnvironmentItemModel
	for i := 0; i < len(kv); i += 2 {
		envs = append(envs, envmanModels.EnvironmentItemModel{kv[i]: kv[i+1]})
	}
	return envs
}
