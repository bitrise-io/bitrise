package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_GoModMigration(t *testing.T) {
	configPth := "gomodmigrate.yml"
	workflows := []string{"test_gopath_disabled", "test_gopath_enabled"}

	for _, wf := range workflows {
		t.Logf("Test go modules and GOPATH modes: %s", wf)

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err, "failed to get HOME dir")

		err = os.RemoveAll(filepath.Join(homeDir, ".bitrise", "toolkits", "go", "cache"))
		require.NoError(t, err, "faield to clean Step binary cache")

		cmd := command.New(binPath(), "--debug", "run", wf, "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, "Bitrise CLI failed, output: %s", out)
	}
}
