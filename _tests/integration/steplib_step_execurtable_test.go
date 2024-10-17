//go:build linux_and_mac
// +build linux_and_mac

package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func TestSteplibStepExecutable(t *testing.T) {
	cmd := command.New(binPath(), "run", "step-executable-test")
	cmd.SetDir("steplib_step_executable")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}
