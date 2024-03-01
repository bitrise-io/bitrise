//go:build linux_and_mac
// +build linux_and_mac

package integration

import (
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_SensitiveInputs(t *testing.T) {
	configPth := "sensitive_inputs_test_bitrise.yml"

	cmd := command.New(binPath(), "run", "test-sensitive-env-and-output", "--config", configPth)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()

	require.NoError(t, err, out)

	require.Equal(t, 1, strings.Count(out, "realvalue"))
	require.Equal(t, 1, strings.Count(out, "mysupersecret"))
	require.Equal(t, 1, strings.Count(out, "myotherverysecret"))

	require.Equal(t, 3, strings.Count(out, "[REDACTED]"))
}
