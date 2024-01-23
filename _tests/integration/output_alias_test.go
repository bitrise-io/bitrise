//go:build linux_and_mac
// +build linux_and_mac

package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_OutputAlias(t *testing.T) {
	configPth := "output_alias_test_bitrise.yml"

	for _, aTestWFID := range []string{
		"test-single-step-single-alias", "test-double-step-single-alias",
		"test-double-step-double-alias", "test-wf-env",
	} {
		t.Log(aTestWFID)
		{
			cmd := command.New(binPath(), "run", aTestWFID, "--config", configPth)
			out, err := cmd.RunAndReturnTrimmedCombinedOutput()
			require.NoError(t, err, out)
		}
	}
}
