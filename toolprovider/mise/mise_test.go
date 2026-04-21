package mise

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/execenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewToolProvider_InvalidDirs(t *testing.T) {
	tests := []struct {
		name       string
		installDir string
		dataDir    string
	}{
		{name: "empty installDir", installDir: "", dataDir: t.TempDir()},
		{name: "empty dataDir", installDir: t.TempDir(), dataDir: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewToolProvider(tt.installDir, tt.dataDir, false, false, nil)
			require.Error(t, err)
		})
	}
}

func TestNewToolProvider_CreatesDirs(t *testing.T) {
	base := t.TempDir()
	installDir := filepath.Join(base, "install")
	dataDir := filepath.Join(base, "data")

	_, err := NewToolProvider(installDir, dataDir, false, false, nil)
	require.NoError(t, err)

	_, err = os.Stat(installDir)
	assert.NoError(t, err, "installDir should be created")
	_, err = os.Stat(dataDir)
	assert.NoError(t, err, "dataDir should be created")
}

func TestNewToolProvider_MiseEnvsAlwaysSet(t *testing.T) {
	dataDir := t.TempDir()

	p, err := NewToolProvider(t.TempDir(), dataDir, false, false, nil)
	require.NoError(t, err)

	envs := p.ExecEnv.(execenv.MiseExecEnv).ExtraEnvs
	assert.Equal(t, dataDir, envs["MISE_DATA_DIR"])
	assert.Equal(t, dataDir, envs["MISE_CONFIG_DIR"])
	assert.Equal(t, "1", envs["MISE_NODE_COREPACK"])
}

func TestNewToolProvider_ExtraEnvs(t *testing.T) {
	tests := []struct {
		name      string
		extraEnvs map[string]string
		assertFn  func(t *testing.T, envs map[string]string)
	}{
		{
			name:      "nil extraEnvs",
			extraEnvs: nil,
			assertFn: func(t *testing.T, envs map[string]string) {
				assert.NotContains(t, envs, "GITHUB_TOKEN")
			},
		},
		{
			name:      "extraEnvs are merged in",
			extraEnvs: map[string]string{"GITHUB_TOKEN": "ghp_test"},
			assertFn: func(t *testing.T, envs map[string]string) {
				assert.Equal(t, "ghp_test", envs["GITHUB_TOKEN"])
			},
		},
		{
			name:      "extraEnvs override mise-specific envs",
			extraEnvs: map[string]string{"MISE_NODE_COREPACK": "0"},
			assertFn: func(t *testing.T, envs map[string]string) {
				assert.Equal(t, "0", envs["MISE_NODE_COREPACK"])
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewToolProvider(t.TempDir(), t.TempDir(), false, false, tt.extraEnvs)
			require.NoError(t, err)
			tt.assertFn(t, p.ExecEnv.(execenv.MiseExecEnv).ExtraEnvs)
		})
	}
}
