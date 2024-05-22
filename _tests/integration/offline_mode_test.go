package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/stretchr/testify/require"
)

const (
	offlineModeConfigPath = "offline_mode.yml"
	defaultStepLibURI     = "https://github.com/bitrise-io/bitrise-steplib.git"
)

func preloadSteps(t *testing.T) {
	// Clean steplib
	route, found := stepman.ReadRoute(defaultStepLibURI)
	require.True(t, found, "Failed to read route for Steplib: %s", defaultStepLibURI)

	err := stepman.CleanupRoute(route)
	require.NoError(t, err, "Failed to cleanup route: %s", route)

	// Preload steps
	cmd := command.New(binPath(), "steps", "preload", "--majors=1", "--minors=1", "--minors-since=0", "--patches-since=0")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, "Preload failed, output: %s", out)
}

func Test_GivenOfflineMode_WhenStepNotCached_ThenFails(t *testing.T) {
	preloadSteps(t)

	cmd := command.New(binPath(), "run", "not_cached", "--config", offlineModeConfigPath)
	cmd.AppendEnvs("BITRISE_OFFLINE_MODE=true")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()

	require.Error(t, err, "Bitrise CLI failed, output: %s", out)
	require.Contains(t, out, "Other versions available in the local cache:")
}

func Test_GivenOnlineMode_WhenStepNotCached_ThenSucceeds(t *testing.T) {
	preloadSteps(t)

	cmd := command.New(binPath(), "run", "not_cached", "--config", offlineModeConfigPath)
	cmd.AppendEnvs("BITRISE_OFFLINE_MODE=false")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()

	require.NoError(t, err, "Bitrise CLI failed, output: %s", out)
}

func Test_GivenOfflineMode_WhenStepCached_ThenSuceeds(t *testing.T) {
	preloadSteps(t)

	cmd := command.New(binPath(), "run", "cached", "--config", offlineModeConfigPath)
	cmd.AppendEnvs("BITRISE_OFFLINE_MODE=true")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()

	require.NoError(t, err, "Bitrise CLI failed, output: %s", out)
	t.Log(out)
}
