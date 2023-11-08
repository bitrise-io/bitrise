package configs

import (
	"reflect"
	"testing"
)

func TestReadAgentConfig(t *testing.T) {
	t.Setenv("BITRISE_APP_SLUG", "ef7a9665e8b6408b")
	t.Setenv("BITRISE_BUILD_SLUG", "80b66786-d011-430f-9c68-00e9416a7325")
	t.Setenv("HOME", "/Users/bitrise")
	testCases := []struct {
		name          string
		configFile     string
		expectedConfig AgentConfig
		expectedErr   bool
	}{
		{
			name:        "Valid config file",
			configFile:   "testdata/valid-agent-config.yml",
			expectedConfig: AgentConfig{
				BitriseDirs {
					SourceDir:     "/opt/bitrise/workspace/ef7a9665e8b6408b",
					DeployDir:     "/opt/bitrise/ef7a9665e8b6408b/80b66786-d011-430f-9c68-00e9416a7325/artifacts",
					TestDeployDir: "/opt/bitrise/ef7a9665e8b6408b/80b66786-d011-430f-9c68-00e9416a7325/test_results",
				},
				AgentHooks {
					CleanupOnWorkflowStart: []string { "$BITRISE_DEPLOY_DIR" },
					CleanupOnWorkflowEnd: []string { "$BITRISE_TEST_DEPLOY_DIR" },
					DoOnWorkflowStart: "/Users/bitrise/hooks/pre-build.sh",
					DoOnWorkflowEnd: "/Users/bitrise/hooks/post-build.sh",
				},
			},
			expectedErr: false,
		},
		{
			name:        "Non-existent config file",
			configFile:   "nonexistent",
			expectedConfig: AgentConfig{},
			expectedErr:   true,
		},
		{
			name:        "Config file with invalid YAML",
			configFile:   "testdata",
			expectedConfig: AgentConfig{},
			expectedErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := readAgentConfig(tc.configFile)
			if (err != nil) != tc.expectedErr {
				t.Errorf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(config, tc.expectedConfig) {
				t.Errorf("Expected config: %v, but got: %v", tc.expectedConfig, config)
			}
		})
	}
}
