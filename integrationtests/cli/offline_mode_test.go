//go:build linux_and_mac
// +build linux_and_mac

package cli

import (
	"testing"

	"github.com/bitrise-io/bitrise/v3/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

const offlineModeConfigPath = "offline_mode.yml"

// preloadSteps fills an empty steplib cache with preloaded steps only, and returns the
// HOME env that runs bitrise against it. The cache is in ~/.stepman, so a test that starts
// it from scratch needs its own HOME rather than the one the other tests share.
func preloadSteps(t *testing.T) string {
	homeEnv := "HOME=" + testhelpers.IsolatedHome(t)

	cmd := command.New(testhelpers.BinPath(), "steps", "preload", "--majors=1", "--minors=1", "--minors-since=0", "--patches-since=0")
	cmd.AppendEnvs(homeEnv)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, "Preload failed, output: %s", out)

	return homeEnv
}

func Test_GivenOfflineMode_WhenStepNotCached_ThenFails(t *testing.T) {
	homeEnv := preloadSteps(t)

	cmd := command.New(testhelpers.BinPath(), "run", "not_cached", "--config", offlineModeConfigPath)
	cmd.AppendEnvs(homeEnv, "BITRISE_OFFLINE_MODE=true", "BITRISE_STEPLIB_USE_API=false")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()

	require.Error(t, err, "Bitrise CLI failed, output: %s", out)
	require.Contains(t, out, "Other versions available in the local cache:")
}

func Test_GivenOnlineMode_WhenStepNotCached_ThenSucceeds(t *testing.T) {
	homeEnv := preloadSteps(t)

	cmd := command.New(testhelpers.BinPath(), "run", "not_cached", "--config", offlineModeConfigPath)
	cmd.AppendEnvs(homeEnv, "BITRISE_OFFLINE_MODE=false")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()

	require.NoError(t, err, "Bitrise CLI failed, output: %s", out)
}

func Test_GivenOfflineMode_WhenStepCached_ThenSuceeds(t *testing.T) {
	homeEnv := preloadSteps(t)

	cmd := command.New(testhelpers.BinPath(), "run", "cached", "--config", offlineModeConfigPath)
	cmd.AppendEnvs(homeEnv, "BITRISE_OFFLINE_MODE=true", "BITRISE_STEPLIB_USE_API=false")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()

	require.NoError(t, err, "Bitrise CLI failed, output: %s", out)
	t.Log(out)
}
