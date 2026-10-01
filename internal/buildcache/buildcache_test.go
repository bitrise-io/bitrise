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
		assert.False(t, enabled(nil))
	})
	t.Run("process env", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "true")
		assert.True(t, enabled(nil))
	})
	t.Run("build env", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "")
		assert.True(t, enabled(buildEnvs(EnvActivateAll, "true")))
	})
}

func TestCacheEnvs(t *testing.T) {
	envs := buildEnvs(
		"BITRISE_BUILD_CACHE_AUTH_TOKEN", "auth",
		servicesTokenKey, "jwt",
		"BITRISE_BUILD_API_TOKEN", "unrelated",
		"GRADLE_ENCRYPTION_KEY", "unrelated",
	)

	assert.Equal(t, []string{"BITRISE_BUILD_CACHE_AUTH_TOKEN=auth", servicesTokenKey + "=jwt"}, cacheEnvs(envs))
}

func TestActivateAll_RunsTheCLIOnceWithCacheEnvsOnly(t *testing.T) {
	setPin(t)
	t.Setenv("HOME", t.TempDir())
	out := filepath.Join(t.TempDir(), "calls")
	installFakeCLI(t, `echo "$@ $BITRISEIO_BITRISE_SERVICES_ACCESS_TOKEN ${BITRISE_BUILD_API_TOKEN:-absent}" >> `+out)

	err := activateAll(context.Background(), testLogger(), buildEnvs(servicesTokenKey, "jwt", "BITRISE_BUILD_API_TOKEN", "secret"))

	require.NoError(t, err)
	calls, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, "activate all --auto jwt absent", strings.TrimSpace(string(calls)))

	target, err := os.Readlink(filepath.Join(configs.GetBitriseToolsDirPath(), binaryName))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(configs.GetBitriseHomeDirPath(), "build-cache", "3.14.1", binaryName), target)
}

func TestActivateAll_ReportsAFailingCLI(t *testing.T) {
	setPin(t)
	t.Setenv("HOME", t.TempDir())
	installFakeCLI(t, "exit 3")

	err := activateAll(context.Background(), testLogger(), nil)

	require.ErrorContains(t, err, "activate all: exit status 3")
}

func TestActivateAll_RequiresPin(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, pin := range [][2]string{{"", ""}, {"3.14.1", ""}, {"../3.14.1", strings.Repeat("a", 64)}, {"3.14.1", "not-a-sha"}} {
		t.Setenv(EnvCLIVersion, pin[0])
		t.Setenv(EnvCLISHA256, pin[1])

		err := activateAll(context.Background(), testLogger(), nil)

		require.ErrorContains(t, err, "must be set by the VM setup")
	}
}

func TestActivateIfEnabled_SkipsNestedRuns(t *testing.T) {
	setPin(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv(EnvActivateAll, "true")
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
