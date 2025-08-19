//go:build linux_and_mac
// +build linux_and_mac

package steps

import (
	"testing"
	"os"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func TestSteplibStepExecutable(t *testing.T) {
	cmd := command.New(testhelpers.BinPath(), "run", "step-executable-test", "-c", "steplib_step_executable/bitrise.yml")
	envs := os.Environ()
	envs = append(envs, "BITRISE_EXPERIMENT_PRECOMPILED_STEPS=true")
	cmd.SetEnvs(envs...)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}