//go:build linux_and_mac
// +build linux_and_mac

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v3/analytics"
	"github.com/bitrise-io/bitrise/v3/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_AgentConfigBitriseDirs(t *testing.T) {
	homeEnv := setupAgentConfig(t)

	cmd := command.New(testhelpers.BinPath(), "run", "test_bitrise_dirs", "--config", "agent_config_test_bitrise.yml")
	cmd.AppendEnvs(homeEnv)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}

func Test_AgentConfigBuildHooksSuccess(t *testing.T) {
	homeEnv := setupAgentConfig(t)

	cmd := command.New(testhelpers.BinPath(), "run", "test_build_hooks_success", "--config", "agent_config_test_bitrise.yml")
	cmd.AppendEnvs(homeEnv)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)

	hookDataDir := filepath.Join(testhelpers.BinPath(), "..", "hooks")

	_, err = os.Stat(filepath.Join(hookDataDir, "build_start"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(hookDataDir, "build_end"))
	require.NoError(t, err)
}

func Test_AgentConfigBuildHooksFailure(t *testing.T) {
	homeEnv := setupAgentConfig(t)

	cmd := command.New(testhelpers.BinPath(), "run", "test_build_hooks_failure", "--config", "agent_config_test_bitrise.yml")
	cmd.AppendEnvs(homeEnv)
	_, _ = cmd.RunAndReturnTrimmedCombinedOutput()

	hookDataDir := filepath.Join(testhelpers.BinPath(), "..", "hooks")

	_, err := os.Stat(filepath.Join(hookDataDir, "build_start"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(hookDataDir, "build_end"))
	require.NoError(t, err)
}

// setupAgentConfig writes the agent config into its own HOME and returns the HOME env that
// runs bitrise against it. The agent config turns every bitrise run that reads it into an
// agent run, so it can't go in the HOME the other tests share.
func setupAgentConfig(t *testing.T) string {
	cfg, err := os.ReadFile("agent-config.yml")
	require.NoError(t, err)

	homeDir := testhelpers.IsolatedHome(t)
	bitriseHomeDir := filepath.Join(homeDir, ".bitrise")
	err = os.MkdirAll(bitriseHomeDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(bitriseHomeDir, "agent-config.yml"), cfg, 0644)
	require.NoError(t, err)

	t.Setenv("BITRISE_APP_SLUG", "ef7a9665e8b6408b")
	t.Setenv("BITRISE_BUILD_SLUG", "80b66786-d011-430f-9c68-00e9416a7325")

	// Clear the execution ID env var as we are already running inside a Bitrise CLI execution.
	// This would confuse the logic that prevents running agent hooks in nested executions.
	t.Setenv(analytics.StepExecutionIDEnvKey, "")

	return "HOME=" + homeDir
}
