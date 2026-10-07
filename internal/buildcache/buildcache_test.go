package buildcache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

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
		assert.True(t, enabled(eval(t, buildEnvs(EnvActivateAll, "true")), EnvActivateAll))
	})
	t.Run("the last build env wins", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "")
		assert.False(t, enabled(eval(t, buildEnvs(EnvActivateAll, "true", EnvActivateAll, "false")), EnvActivateAll))
		assert.True(t, enabled(eval(t, buildEnvs(EnvActivateAll, "false", EnvActivateAll, "true")), EnvActivateAll))
	})
	t.Run("a build env beats the process env", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "true")
		assert.False(t, enabled(eval(t, buildEnvs(EnvActivateAll, "false")), EnvActivateAll))
	})
	t.Run("the two opt-ins are independent", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "")
		assert.False(t, enabled(eval(t, buildEnvs(EnvActivateGradleMirrors, "true")), EnvActivateAll))
	})
}

func TestEnabled_EvaluatesTheDeclarationsLikeAStep(t *testing.T) {
	flag := func(value string, opts *envmanModels.EnvironmentItemOptionsModel) envmanModels.EnvironmentItemModel {
		item := envmanModels.EnvironmentItemModel{EnvActivateAll: value}
		if opts != nil {
			item[envmanModels.OptionsKey] = *opts
		}

		return item
	}
	yes, no := true, false

	t.Run("a reference to an earlier declaration", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "")
		items := []envmanModels.EnvironmentItemModel{{"CACHE_ENABLED": "true"}, flag("$CACHE_ENABLED", nil)}

		assert.True(t, enabled(eval(t, items), EnvActivateAll))
	})
	t.Run("a reference to the process env", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "")
		t.Setenv("CACHE_ENABLED_FROM_PROCESS", "true")

		assert.True(t, enabled(eval(t, []envmanModels.EnvironmentItemModel{flag("$CACHE_ENABLED_FROM_PROCESS", nil)}), EnvActivateAll))
	})
	t.Run("a reference that resolves to false turns it off", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "true")
		items := []envmanModels.EnvironmentItemModel{{"CACHE_ENABLED": "false"}, flag("$CACHE_ENABLED", nil)}

		assert.False(t, enabled(eval(t, items), EnvActivateAll))
	})
	t.Run("is_expand false keeps the value literal", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "")
		items := []envmanModels.EnvironmentItemModel{{"CACHE_ENABLED": "true"}, flag("$CACHE_ENABLED", &envmanModels.EnvironmentItemOptionsModel{IsExpand: &no})}

		assert.False(t, enabled(eval(t, items), EnvActivateAll))
	})
	t.Run("skip_if_empty leaves the earlier declaration in force", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "")
		items := []envmanModels.EnvironmentItemModel{flag("true", nil), flag("", &envmanModels.EnvironmentItemOptionsModel{SkipIfEmpty: &yes})}

		assert.True(t, enabled(eval(t, items), EnvActivateAll))
	})
	t.Run("unset turns it off even when the process env has it on", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "true")
		items := []envmanModels.EnvironmentItemModel{flag("true", &envmanModels.EnvironmentItemOptionsModel{Unset: &yes})}

		assert.False(t, enabled(eval(t, items), EnvActivateAll))
	})
	t.Run("an unset beats an earlier declaration", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "")
		items := []envmanModels.EnvironmentItemModel{flag("true", nil), flag("true", &envmanModels.EnvironmentItemOptionsModel{Unset: &yes})}

		assert.False(t, enabled(eval(t, items), EnvActivateAll))
	})
	t.Run("a later declaration after an unset is in force again", func(t *testing.T) {
		t.Setenv(EnvActivateAll, "")
		items := []envmanModels.EnvironmentItemModel{flag("true", &envmanModels.EnvironmentItemOptionsModel{Unset: &yes}), flag("true", nil)}

		assert.True(t, enabled(eval(t, items), EnvActivateAll))
	})
}

func TestEnabled_AnUnsetDeclarationIsGoneForLaterReferences(t *testing.T) {
	t.Setenv(EnvActivateAll, "")
	t.Setenv(EnvActivateGradleMirrors, "")
	yes := true
	items := []envmanModels.EnvironmentItemModel{
		{EnvActivateAll: "true", envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{Unset: &yes}},
		{EnvActivateGradleMirrors: "$" + EnvActivateAll},
	}

	values := eval(t, items)

	assert.False(t, enabled(values, EnvActivateAll))
	assert.False(t, enabled(values, EnvActivateGradleMirrors), "a step would see the first variable as unset when expanding the second")
}

func TestCacheEnvs_ForwardsTheEvaluatedValues(t *testing.T) {
	no := false
	items := []envmanModels.EnvironmentItemModel{
		{"CACHE_TOKEN": "secret-token"},
		{"BITRISE_BUILD_CACHE_AUTH_TOKEN": "$CACHE_TOKEN"},
		{"BITRISE_BUILD_CACHE_USERNAME": "$NOT_EXPANDED", envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{IsExpand: &no}},
		{"UNRELATED": "x"},
	}

	assert.Equal(t, []string{
		"BITRISE_BUILD_CACHE_AUTH_TOKEN=secret-token",
		"BITRISE_BUILD_CACHE_USERNAME=$NOT_EXPANDED",
	}, cacheEnvs(eval(t, items)))
}

func TestActivateIfEnabled_AReferencedOptInAndTokenReachTheCLI(t *testing.T) {
	isolate(t)
	t.Setenv(analytics.StepExecutionIDEnvKey, "")
	out := filepath.Join(t.TempDir(), "calls")
	installFakeCLI(t, `echo "$@ token=$BITRISE_BUILD_CACHE_AUTH_TOKEN" >> `+out)
	items := []envmanModels.EnvironmentItemModel{
		{"WANT_CACHE": "true"},
		{"MY_TOKEN": "tok-1"},
		{EnvActivateAll: "$WANT_CACHE"},
		{"BITRISE_BUILD_CACHE_AUTH_TOKEN": "$MY_TOKEN"},
	}

	require.True(t, ActivateIfEnabled(testLogger(), items))

	calls, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, "activate all --auto token=tok-1", strings.TrimSpace(string(calls)))
}

func TestActivateIfEnabled_RunsTheVersionedBinaryNotTheSharedLink(t *testing.T) {
	isolate(t)
	t.Setenv(EnvActivateAll, "true")
	out := filepath.Join(t.TempDir(), "path")
	installFakeCLI(t, `echo "$0" >> `+out)

	require.True(t, ActivateIfEnabled(testLogger(), nil))

	got, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(configs.GetBitriseHomeDirPath(), "build-cache", testVersion, binaryName), strings.TrimSpace(string(got)))
}

func TestCacheEnvs(t *testing.T) {
	envs := buildEnvs(
		"BITRISE_BUILD_CACHE_AUTH_TOKEN", "auth",
		servicesTokenKey, "jwt",
		buildHubVMTokenKey, "vm",
		buildHubVMTokenURLKey, "https://hub",
		"BITRISE_BUILD_API_TOKEN", "unrelated",
		"GRADLE_ENCRYPTION_KEY", "unrelated",
		gradleUserHomeKey, "/work/gradle",
	)

	assert.Equal(t, []string{
		servicesTokenKey + "=jwt",
		buildHubVMTokenKey + "=vm",
		buildHubVMTokenURLKey + "=https://hub",
		"BITRISE_BUILD_CACHE_AUTH_TOKEN=auth",
		gradleUserHomeKey + "=/work/gradle",
	}, cacheEnvs(eval(t, envs)))
}

func TestActivateIfEnabled_InstallsOnceAndRunsBothCommands(t *testing.T) {
	isolate(t)
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
	isolate(t)
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
	isolate(t)
	t.Setenv(EnvActivateAll, "")
	t.Setenv(EnvActivateGradleMirrors, "true")
	t.Setenv(envDisableHostsOverride, "true")
	out := filepath.Join(t.TempDir(), "calls")
	installFakeCLI(t, `echo called >> `+out)

	assert.False(t, ActivateIfEnabled(testLogger(), nil))
	assert.NoFileExists(t, out)
}

func TestActivateIfEnabled_AFailingCommandDoesNotStopTheOther(t *testing.T) {
	isolate(t)
	t.Setenv(EnvActivateAll, "true")
	t.Setenv(EnvActivateGradleMirrors, "true")
	out := filepath.Join(t.TempDir(), "calls")
	installFakeCLI(t, `echo "$2" >> `+out+`; exit 3`)

	require.True(t, ActivateIfEnabled(testLogger(), nil))

	calls, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, []string{"gradle-mirrors", "all"}, strings.Split(strings.TrimSpace(string(calls)), "\n"))
}

func TestActivateIfEnabled_HostsOverrideDisabledStillRunsAll(t *testing.T) {
	isolate(t)
	t.Setenv(EnvActivateAll, "true")
	t.Setenv(EnvActivateGradleMirrors, "true")
	t.Setenv(envDisableHostsOverride, "true")
	out := filepath.Join(t.TempDir(), "calls")
	installFakeCLI(t, `echo "$2" >> `+out)

	require.True(t, ActivateIfEnabled(testLogger(), nil))

	calls, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, "all", strings.TrimSpace(string(calls)))
}

func TestActivateIfEnabled_InstallsThroughTheHookAndRunsBothCommands(t *testing.T) {
	isolate(t)
	out := filepath.Join(t.TempDir(), "calls")
	tarball, checksum := releaseTarball(t, "#!/bin/sh\necho \"$2\" >> "+out+"\n")
	var downloads atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		downloads.Add(1)
		_, _ = w.Write(tarball)
	}))
	defer srv.Close()
	t.Setenv(EnvCLISHA256, checksum)
	t.Setenv(EnvHostCacheURL, srv.URL)
	t.Setenv(EnvActivateAll, "true")
	t.Setenv(EnvActivateGradleMirrors, "true")

	require.True(t, ActivateIfEnabled(testLogger(), nil))

	calls, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, []string{"gradle-mirrors", "all"}, strings.Split(strings.TrimSpace(string(calls)), "\n"))
	assert.EqualValues(t, 1, downloads.Load())
}

func TestActivateIfEnabled_InstallFailureIsAttemptedButRunsNothing(t *testing.T) {
	isolate(t)
	t.Setenv(EnvCLIVersion, "")
	t.Setenv(EnvActivateAll, "true")

	assert.True(t, ActivateIfEnabled(testLogger(), nil))
}

func TestActivateIfEnabled_AHangingCommandDoesNotStarveTheNext(t *testing.T) {
	isolate(t)
	t.Setenv(EnvActivateAll, "true")
	t.Setenv(EnvActivateGradleMirrors, "true")
	setTimeouts(t, 3*time.Second, 200*time.Millisecond)
	out := filepath.Join(t.TempDir(), "calls")
	installFakeCLI(t, `if [ "$2" = gradle-mirrors ]; then sleep 30 & wait; fi; echo "$2" >> `+out)

	start := time.Now()
	require.True(t, ActivateIfEnabled(testLogger(), nil))

	assert.Less(t, time.Since(start), 8*time.Second)
	calls, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, "all", strings.TrimSpace(string(calls)))
}

func TestRun_ReturnsWhenADescendantKeepsTheOutputOpen(t *testing.T) {
	setTimeouts(t, 10*time.Second, 200*time.Millisecond)
	bin := filepath.Join(t.TempDir(), "daemonizing")
	require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\nsleep 30 &\n"), 0o755))

	start := time.Now()
	require.NoError(t, run(context.Background(), testLogger(), bin, os.Environ()))

	assert.Less(t, time.Since(start), 5*time.Second)
}

func setTimeouts(t *testing.T, command, waitDelay time.Duration) {
	t.Helper()
	oldCommand, oldWait := commandTimeout, commandWaitDelay
	commandTimeout, commandWaitDelay = command, waitDelay
	t.Cleanup(func() { commandTimeout, commandWaitDelay = oldCommand, oldWait })
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
	isolate(t)
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

// isolate clears every env var the package reads, so a CI or VM export cannot leak into a test.
func isolate(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv(analytics.StepExecutionIDEnvKey, "")
	for _, key := range []string{
		EnvActivateAll, EnvActivateGradleMirrors, EnvHostCacheURL, envDisableHostsOverride, envMavenCentralProxy,
		servicesTokenKey, buildHubVMTokenKey, buildHubVMTokenURLKey,
	} {
		t.Setenv(key, "")
	}
	t.Setenv(EnvCLIVersion, testVersion)
	t.Setenv(EnvCLISHA256, strings.Repeat("a", 64))
}

func eval(t *testing.T, items []envmanModels.EnvironmentItemModel) map[string]string {
	t.Helper()

	values, err := evaluate(items)
	require.NoError(t, err)

	return values
}

func buildEnvs(kv ...string) []envmanModels.EnvironmentItemModel {
	var envs []envmanModels.EnvironmentItemModel
	for i := 0; i < len(kv); i += 2 {
		envs = append(envs, envmanModels.EnvironmentItemModel{kv[i]: kv[i+1]})
	}
	return envs
}

func TestActivateIfEnabled_BothCommandsSeeABuildDeclaredGradleUserHome(t *testing.T) {
	isolate(t)
	t.Setenv(EnvActivateAll, "true")
	t.Setenv(EnvActivateGradleMirrors, "true")
	t.Setenv(gradleUserHomeKey, "")
	out := filepath.Join(t.TempDir(), "calls")
	installFakeCLI(t, `echo "$@ home=${GRADLE_USER_HOME:-unset}" >> `+out)

	require.True(t, ActivateIfEnabled(testLogger(), buildEnvs(gradleUserHomeKey, "/work/gradle")))

	calls, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, "activate gradle-mirrors -d home=/work/gradle\nactivate all --auto home=/work/gradle", strings.TrimSpace(string(calls)))
}
