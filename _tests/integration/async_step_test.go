package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_AsyncStep(t *testing.T) {
	configPth := "async_step_test_bitrise.yml"

	for _, aTestWFID := range []string{
		"asynctest",
	} {
		t.Log(aTestWFID)
		{
			cmd := command.New(binPath(), "run", aTestWFID, "--config", configPth)
			out, err := cmd.RunAndReturnTrimmedCombinedOutput()
			require.NoError(t, err, out)
		}
	}
}
