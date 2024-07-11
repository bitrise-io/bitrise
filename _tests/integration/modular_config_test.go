package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_ModularConfig_Run(t *testing.T) {
	configPth := "modular_config_main.yml"

	cmd := command.New(binPath(), "validate", "--config", configPth)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Equal(t, "Config is valid: \u001B[32;1mtrue\u001B[0m", out)

	cmd = command.New(binPath(), "workflows", "--id-only", "--config", configPth)
	out, err = cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Equal(t, "print_hello_bitrise print_hello_world", out)

	cmd = command.New(binPath(), "run", "print_hello_bitrise", "--config", configPth)
	out, err = cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Contains(t, out, "Hello Bitrise!")

	cmd = command.New(binPath(), "run", "print_hello_world", "--config", configPth)
	out, err = cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
	require.Contains(t, out, "Hello World!")
}
