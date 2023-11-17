package integration

import (
	"os"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/stretchr/testify/require"
)

func Test_AgentConfigBitriseDirs(t *testing.T) {
	cleanupFn := setupAgentConfig(t)
	defer cleanupFn()

	cmd := command.New(binPath(), "run", "test_bitrise_dirs", "--config", "agent_config_test_bitrise.yml")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}

func Test_AgentConfigBuildHooks(t *testing.T) {
	cleanupFn := setupAgentConfig(t)
	defer cleanupFn()

	cmd := command.New(binPath(), "run", "test_build_hooks", "--config", "agent_config_test_bitrise.yml")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}

func setupAgentConfig(t *testing.T) func() {
	cfg, err := os.ReadFile("agent-config.yml")
	require.NoError(t, err)

	absPath, err := pathutil.AbsPath("$HOME/.bitrise/agent-config.yml")
	require.NoError(t, err)

	err = os.WriteFile(absPath, cfg, 0644)
	require.NoError(t, err)
	cleanupFn := func() { os.Remove(absPath) }

	t.Setenv("BITRISE_APP_SLUG", "ef7a9665e8b6408b")
	t.Setenv("BITRISE_BUILD_SLUG", "80b66786-d011-430f-9c68-00e9416a7325")

	return cleanupFn
}
