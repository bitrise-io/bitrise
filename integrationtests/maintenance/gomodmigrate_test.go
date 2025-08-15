//go:build linux_and_mac
// +build linux_and_mac

package maintenance

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_GoModMigration(t *testing.T) {
	configPth := "gomodmigrate.yml"
	workflow := "test"
	moduleModes := []string{"on", "auto"}

	for _, mode := range moduleModes {
		t.Run(fmt.Sprintf("Go module modes, GO111MODULE='%s'", mode), func(t *testing.T) {
			homeDir, err := os.UserHomeDir()
			require.NoError(t, err, "failed to get HOME dir")

			err = os.RemoveAll(filepath.Join(homeDir, ".bitrise", "toolkits", "go", "cache"))
			require.NoError(t, err, "faield to clean Step binary cache")

			cmd := command.New(testhelpers.BinPath(), "--debug", "run", workflow, "--config", configPth).
				AppendEnvs("GO111MODULE=" + mode)
			out, err := cmd.RunAndReturnTrimmedCombinedOutput()
			require.NoError(t, err, "Bitrise CLI failed, output: %s", out)
		})
	}
}