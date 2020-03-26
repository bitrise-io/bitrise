package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_OutputAlias(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

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
