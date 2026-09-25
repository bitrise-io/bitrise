//go:build linux_and_mac
// +build linux_and_mac

package asdf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"al.essio.dev/pkg/shellescape"
	"github.com/bitrise-io/bitrise/v2/toolprovider"
	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

// Other version managers (such as mise) provide shims for the same tools and ignore the
// ASDF_TOOL_VERSION env vars, so an activated asdf tool is only correct if the asdf shims
// come first in the $PATH of the step.
func TestActivatedToolWinsOverOtherShims(t *testing.T) {
	testEnv, err := createTestEnv(t, asdfInstallation{
		flavor:  flavorAsdfClassic,
		version: "0.14.0",
		plugins: []string{"nodejs"},
	})
	require.NoError(t, err)

	asdfProvider := asdf.AsdfToolProvider{
		ExecEnv: testEnv.toExecEnv(),
	}
	result, err := asdfProvider.InstallTool(provider.ToolRequest{
		ToolName:           "nodejs",
		UnparsedVersion:    "24.20.0",
		ResolutionStrategy: provider.ResolutionStrategyStrict,
	})
	require.NoError(t, err)
	require.Equal(t, "24.20.0", result.ConcreteVersion)

	// A different global default is what stacks set up, and what the activation has to override.
	_, err = asdfProvider.InstallTool(provider.ToolRequest{
		ToolName:           "nodejs",
		UnparsedVersion:    "18.16.0",
		ResolutionStrategy: provider.ResolutionStrategyStrict,
	})
	require.NoError(t, err)
	_, err = testEnv.runAsdf("global", "nodejs", "18.16.0")
	require.NoError(t, err)

	otherShimsDir := t.TempDir()
	otherNodeShim := filepath.Join(otherShimsDir, "node")
	require.NoError(t, os.WriteFile(otherNodeShim, []byte("#!/bin/bash\necho v99.0.0\n"), 0700))

	activation, err := asdfProvider.ActivateEnv(result)
	require.NoError(t, err)
	require.NotEmpty(t, activation.ContributedPaths, "activation should contribute the asdf shims dir")

	out, err := runInActivatedEnv(t, testEnv, activation, otherShimsDir, "node", "--version")
	require.NoError(t, err)
	require.Equal(t, "v24.20.0", strings.TrimSpace(out))
}

// runInActivatedEnv runs a command the way a step would see it: the competing shims dir comes first in
// the inherited $PATH, and the activation is applied by the same code that bitrise runs for a step.
func runInActivatedEnv(
	t *testing.T,
	testEnv testEnv,
	activation provider.EnvironmentActivation,
	competingPathEntry string,
	args ...string,
) (string, error) {
	basePath := testEnv.envVars["PATH"]
	if basePath == "" {
		basePath = os.Getenv("PATH")
	}
	competingFirstPath := competingPathEntry + ":" + basePath

	// ConvertToEnvMap prepends the contributed paths to the $PATH of the current process.
	t.Setenv("PATH", competingFirstPath)

	envs := map[string]string{}
	for k, v := range testEnv.envVars {
		envs[k] = v
	}
	envs["PATH"] = competingFirstPath
	for k, v := range toolprovider.ConvertToEnvMap([]provider.EnvironmentActivation{activation}) {
		envs[k] = v
	}

	cmd := exec.Command("bash", "-c", shellescape.QuoteCommand(args))
	cmd.Env = os.Environ()
	for k, v := range envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %w\n\nOutput:\n%s", args, err, out)
	}
	return string(out), nil
}
