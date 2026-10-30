package asdf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/asdf/execenv"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivateEnv(t *testing.T) {
	execEnvDataDir := t.TempDir()
	osEnvDataDir := t.TempDir()
	homeDir := t.TempDir()
	dataDirWithoutShims := t.TempDir()

	require.NoError(t, os.Mkdir(filepath.Join(execEnvDataDir, "shims"), 0700))
	require.NoError(t, os.Mkdir(filepath.Join(osEnvDataDir, "shims"), 0700))
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".asdf", "shims"), 0700))

	tests := []struct {
		name           string
		execEnvDataDir string
		osEnvDataDir   string
		expectedPaths  []string
	}{
		{
			name:           "data dir of the exec env wins over the process env",
			execEnvDataDir: execEnvDataDir,
			osEnvDataDir:   osEnvDataDir,
			expectedPaths:  []string{filepath.Join(execEnvDataDir, "shims")},
		},
		{
			name:          "data dir of the process env is the fallback",
			osEnvDataDir:  osEnvDataDir,
			expectedPaths: []string{filepath.Join(osEnvDataDir, "shims")},
		},
		{
			name:          "home dir is the last resort",
			expectedPaths: []string{filepath.Join(homeDir, ".asdf", "shims")},
		},
		{
			name:           "no path is contributed when the shims dir doesn't exist",
			execEnvDataDir: dataDirWithoutShims,
			expectedPaths:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", homeDir)
			t.Setenv("ASDF_DATA_DIR", tt.osEnvDataDir)

			toolProvider := AsdfToolProvider{
				ExecEnv: execenv.ExecEnv{
					EnvVars: map[string]string{"ASDF_DATA_DIR": tt.execEnvDataDir},
				},
			}

			activation, err := toolProvider.ActivateEnv(provider.ToolInstallResult{
				ToolName:        "nodejs",
				ConcreteVersion: "22.23.2",
			})
			require.NoError(t, err)

			assert.Equal(t, map[string]string{"ASDF_NODEJS_VERSION": "22.23.2"}, activation.ContributedEnvVars)
			assert.Equal(t, tt.expectedPaths, activation.ContributedPaths)
		})
	}
}
