//go:build linux_and_mac
// +build linux_and_mac

package workflow

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_AsyncStep(t *testing.T) {
	configPth := "async_step_test_bitrise.yml"

	aTestWFID := "asynctest"
	{
		t.Log(aTestWFID)
		{
			cmd := command.New(testhelpers.BinPath(), "run", aTestWFID, "--config", configPth)
			out, err := cmd.RunAndReturnTrimmedCombinedOutput()
			require.NoError(t, err, out)
		}
	}
}
