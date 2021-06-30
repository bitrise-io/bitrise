package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_EnvSizeTest(t *testing.T) {
	configPth := "env_size_test_bitrise.yml"

	t.Log("exit_code_test_fail")
	{
		cmd := command.New(binPath(), "run", "env-size-test", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}
