//go:build linux_and_mac
// +build linux_and_mac

package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v3/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CLIShims_EnvmanDispatch(t *testing.T) {
	bin, err := exec.LookPath(testhelpers.BinPath())
	require.NoError(t, err)
	bin, err = filepath.Abs(bin)
	require.NoError(t, err)
	shim := filepath.Join(t.TempDir(), "envman")
	require.NoError(t, os.Symlink(bin, shim))
	dir := t.TempDir()

	cliVersion, err := command.New(bin, "version").SetDir(dir).AppendEnvs("BITRISE_ANALYTICS_DISABLED=true").RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, cliVersion)
	for _, flag := range []string{"-version", "--version", "version"} {
		out, err := command.New(shim, flag).SetDir(dir).RunAndReturnTrimmedOutput()
		require.NoError(t, err, out)
		assert.Equal(t, cliVersion, out, flag)
	}

	envstore := "ENVMAN_ENVSTORE_PATH=" + filepath.Join(dir, ".envstore.yml")
	out, err := command.New(shim, "init", "--clear").SetDir(dir).AppendEnvs(envstore).RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	// Single-dash long flags are part of the envman interface that bitrise itself rejects.
	out, err = command.New(shim, "add", "-key", "FOO", "-value", "bar").SetDir(dir).AppendEnvs(envstore).RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	out, err = command.New(shim, "print", "--format", "json").SetDir(dir).AppendEnvs(envstore).RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	assert.Contains(t, out, `"FOO":"bar"`)
}

func Test_CLIShims_SetupAndRun(t *testing.T) {
	home := t.TempDir()
	// Runs its content input, like the script step, without a StepLib setup in the empty HOME.
	stepDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(stepDir, "step.yml"), []byte("inputs:\n- content:\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(stepDir, "step.sh"), []byte("#!/usr/bin/env bash\nset -euo pipefail\nbash -ec \"$content\"\n"), 0o755))
	envs := []string{
		"BITRISE_CLI_SHIMS_STEP_DIR=" + stepDir,
		"HOME=" + home,
		"XDG_CONFIG_HOME=" + t.TempDir(),
		"BITRISE_ANALYTICS_DISABLED=true",
		"BITRISE_TOKEN=",
	}
	shim := filepath.Join(home, ".bitrise", "shims", "envman")
	configPth, err := filepath.Abs("cli_shims_test_bitrise.yml")
	require.NoError(t, err)
	workDir := t.TempDir()
	runBitrise := func(args ...string) (string, error) {
		return command.New(testhelpers.BinPath(), args...).SetDir(workDir).AppendEnvs(envs...).RunAndReturnTrimmedCombinedOutput()
	}

	out, err := runBitrise("setup", "--minimal", "--no-update", "--experimental-gen-cli-shims")
	require.NoError(t, err, out)
	_, err = os.Readlink(shim)
	require.NoError(t, err, out)

	out, err = runBitrise("setup", "--minimal", "--no-update")
	require.NoError(t, err, out)
	_, err = os.Readlink(shim)
	require.NoError(t, err, "setup without the flag keeps the shim")

	out, err = runBitrise("run", "envman-through-shim", "--config", configPth)
	require.NoError(t, err, out)

	// The shim exists and is first on PATH now, so setup's version check has to skip it.
	out, err = runBitrise("setup", "--minimal", "--no-update", "--experimental-gen-cli-shims=false")
	require.NoError(t, err, out)
	assert.NotContains(t, out, shim, "setup checked the shim instead of the installed envman")
	_, err = os.Lstat(shim)
	assert.ErrorIs(t, err, os.ErrNotExist, out)
}
