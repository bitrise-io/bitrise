package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/stretchr/testify/require"
)

func Test_EnvstoreTest(t *testing.T) {
	configPth := "envstore_test_bitrise.yml"

	t.Log("exit_code_test_fail")
	{
		cmd := cmdex.NewCommand(binPath(), "run", "envstore_test", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}
