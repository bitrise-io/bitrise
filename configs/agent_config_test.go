package configs

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadAgentConfig(t *testing.T) {
	t.Setenv("BITRISE_APP_SLUG", "ef7a9665e8b6408b")
	t.Setenv("BITRISE_BUILD_SLUG", "80b66786-d011-430f-9c68-00e9416a7325")
	tempDir := t.TempDir()
	t.Setenv("HOOKS_DIR", tempDir)
	err := ioutil.WriteFile(filepath.Join(tempDir, "cleanup.sh"), []byte("echo cleanup.sh"), 0644)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		configFile     string
		expectedConfig AgentConfig
		expectedErr    bool
	}{
		{
			name:       "Full config file",
			configFile: "testdata/full-agent-config.yml",
			expectedConfig: AgentConfig{
				BitriseDirs{
					BitriseDataHomeDir: "/opt/bitrise",
					SourceDir:          "/opt/bitrise/workspace/ef7a9665e8b6408b",
					DeployDir:          "/opt/bitrise/ef7a9665e8b6408b/80b66786-d011-430f-9c68-00e9416a7325/artifacts",
					TestDeployDir:      "/opt/bitrise/ef7a9665e8b6408b/80b66786-d011-430f-9c68-00e9416a7325/test_results",
				},
				AgentHooks{
					CleanupOnBuildStart: []string{"$BITRISE_DEPLOY_DIR"},
					CleanupOnBuildEnd:   []string{"$BITRISE_TEST_DEPLOY_DIR"},
					DoOnBuildStart:      filepath.Join(tempDir, "cleanup.sh"),
					DoOnBuildEnd:        filepath.Join(tempDir, "cleanup.sh"),
				},
			},
			expectedErr: false,
		},
		{
			name:       "Minimal config file",
			configFile: "testdata/minimal-agent-config.yml",
			expectedConfig: AgentConfig{
				BitriseDirs{
					BitriseDataHomeDir: "/opt/bitrise",
					SourceDir:          "/opt/bitrise/workspace",
					DeployDir:          "/opt/bitrise/ef7a9665e8b6408b/80b66786-d011-430f-9c68-00e9416a7325/artifacts",
					TestDeployDir:      "/opt/bitrise/ef7a9665e8b6408b/80b66786-d011-430f-9c68-00e9416a7325/test_results",
				},
				AgentHooks{},
			},
			expectedErr: false,
		},
		{
			name:           "Non-existent config file",
			configFile:     "nonexistent",
			expectedConfig: AgentConfig{},
			expectedErr:    true,
		},
		{
			name:           "Config file with invalid YAML",
			configFile:     "testdata",
			expectedConfig: AgentConfig{},
			expectedErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := ReadAgentConfig(tc.configFile)
			if (err != nil) != tc.expectedErr {
				t.Errorf("Unexpected error: %v", err)
			}
			require.Equal(t, tc.expectedConfig, config)
		})
	}
}
