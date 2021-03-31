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
	workflow := "test"
	moduleModes := []string{"on", "auto"}

	for _, mode := range moduleModes {
		t.Logf("Test go modules and GOPATH modes, GO111MODULE=%s", mode)

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err, "failed to get HOME dir")

		err = os.RemoveAll(filepath.Join(homeDir, ".bitrise", "toolkits", "go", "cache"))
		require.NoError(t, err, "faield to clean Step binary cache")

		cmd := command.New(binPath(), "--debug", "run", workflow, "--config", configPth).
			AppendEnvs("GO111MODULE=" + mode)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, "Bitrise CLI failed, output: %s", out)
		t.Logf("Bitrise CLI ouptut: %s", out)
	}
}
