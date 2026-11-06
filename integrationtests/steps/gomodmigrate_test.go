//go:build linux_and_mac
// +build linux_and_mac

package steps

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/bitrise/v3/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_GoModMigration(t *testing.T) {
	configPth := "gomodmigrate.yml"
	workflow := "test"
	moduleModes := []string{"on", "auto"}

	for _, mode := range moduleModes {
		t.Run(fmt.Sprintf("Go module modes, GO111MODULE='%s'", mode), func(t *testing.T) {
			// An empty Step binary cache, so the Go step is built from scratch.
			homeDir := testhelpers.IsolatedHome(t)

			cmd := command.New(testhelpers.BinPath(), "--debug", "run", workflow, "--config", configPth).
				AppendEnvs("HOME="+homeDir, "GO111MODULE="+mode)
			out, err := cmd.RunAndReturnTrimmedCombinedOutput()
			require.NoError(t, err, "Bitrise CLI failed, output: %s", out)
		})
	}
}
