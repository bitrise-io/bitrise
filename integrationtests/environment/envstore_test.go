//go:build linux_and_mac
// +build linux_and_mac

package environment

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_EnvstoreTest(t *testing.T) {
	configPth := "envstore_test_bitrise.yml"

	t.Log("exit_code_test_fail")
	{
		cmd := command.New(testhelpers.BinPath(), "run", "envstore_test", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}