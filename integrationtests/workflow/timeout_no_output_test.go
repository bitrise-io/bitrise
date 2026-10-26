//go:build linux_and_mac
// +build linux_and_mac

package workflow

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v3/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

const configPath = "timeout_no_output.yml"

func Test_GivenHangDetectionOn_WhenThereIsOutput_ThenDoesNotAbort(t *testing.T) {
	cmd := command.New(testhelpers.BinPath(), "run", "output_consistent", "--config", configPath)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()

	require.NoError(t, err, "Bitrise CLI failed, output: %s", out)
}

func Test_GivenHangDetectionOn_WhenOutputSlowsDown_ThenAborts(t *testing.T) {
	cmd := command.New(testhelpers.BinPath(), "run", "output_slows_down", "--config", configPath)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()

	require.EqualError(t, err, "exit status 92", "Bitrise CLI did not abort hanged build, output: %s", out)
}

func Test_GivenHangDetectionSetByAppEnv_WhenOutputSlowsDown_ThenAborts(t *testing.T) {
	absConfigPath, err := filepath.Abs(configPath)
	require.NoError(t, err)

	cmd := command.New(testhelpers.BinPath(), "run", "output_slows_down_app_env", "--config", absConfigPath)
	cmd.SetDir(t.TempDir())
	cmd.AppendEnvs(
		"XDG_CONFIG_HOME="+t.TempDir(),
		"BITRISE_TOKEN=",
		"BITRISE_ANALYTICS_DISABLED=true",
		"BITRISE_NO_OUTPUT_TIMEOUT=",
	)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()

	require.EqualError(t, err, "exit status 92", "Bitrise CLI did not abort hanged build, output: %s", out)
}
