//go:build linux_and_mac
// +build linux_and_mac

package steps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

// TestSteplibStepExecutableV1 exercises the V1 steplib activation route (the
// only route the bitrise CLI wires today) against the real production steplib
// with the precompiled-steps experiment enabled. create-zip publishes a prebuilt
// Go binary, so activation must download and run that executable.
//
// The workflow has a single step, so the only way create-zip's activation falls
// back to source is logged as "fallback to step source activation"; asserting
// that line is absent proves the precompiled binary was actually used.
func TestSteplibStepExecutableV1(t *testing.T) {
	sourceDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "hello.txt"), []byte("precompiled v1 route"), 0o600))
	destZip := filepath.Join(t.TempDir(), "out.zip")

	cmd := command.New(testhelpers.BinPath(), "run", "steplib-executable-v1-test", "-c", "steplib_step_executable_v1/bitrise.yml")
	envs := os.Environ()
	envs = append(envs,
		"BITRISE_EXPERIMENT_PRECOMPILED_STEPS=true",
		"ZIP_SOURCE_PATH="+sourceDir,
		"ZIP_DEST_PATH="+destZip,
	)
	cmd.SetEnvs(envs...)

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.NotContains(t, out, "fallback to step source activation", "create-zip should activate via the precompiled binary, not source")
	require.FileExists(t, destZip)
}
