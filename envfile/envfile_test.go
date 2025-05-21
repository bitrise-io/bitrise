package envfile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMergeEnvfileWithRuntimeEnvs(t *testing.T) {
	tests := []struct {
		name           string
		envfileContent map[string]string
		runtimeEnvs    envmanModels.EnvsJSONListModel
		expected       map[string]string
		wantErr        bool
	}{
		{
			name: "cleared_runtime_envs",
			envfileContent: map[string]string {
				"BITRISE_GIT_CHANGED_FILES":   "README.md",
				"BITRISE_GIT_COMMIT_MESSAGES": "Merge x into y",
				"KEY3":                        "original_value3",
				"ENV_ONLY_DEFINED_HERE":      "original_value4",
			},
			runtimeEnvs: envmanModels.EnvsJSONListModel{
				"BITRISE_GIT_CHANGED_FILES":   "README.md",      // This should stay as is
				"BITRISE_GIT_COMMIT_MESSAGES": "",               // This should be restored from original
				"KEY3":                        "runtime_value3", // This should stay as is
				"KEY5":                        "runtime_value5", // This is only in runtime, should stay
			},
			expected: map[string]string{
				"BITRISE_GIT_CHANGED_FILES":   "README.md",
				"BITRISE_GIT_COMMIT_MESSAGES": "Merge x into y", // Restored from original
				"KEY3":                        "runtime_value3",
				"KEY5":                        "runtime_value5",
			},
			wantErr: false,
		},
		{
			name:           "empty_envfile",
			envfileContent: map[string]string{},
			runtimeEnvs: envmanModels.EnvsJSONListModel{
				"BITRISE_GIT_CHANGED_FILES":   "README.md",      // This should stay as is
				"BITRISE_GIT_COMMIT_MESSAGES": "Fix things",     // This should stay as is
				"KEY3":                        "runtime_value3", // This should stay as is
				"KEY5":                        "runtime_value5", // This is only in runtime, should stay
			},
			expected: map[string]string{
				"BITRISE_GIT_CHANGED_FILES":   "README.md",
				"BITRISE_GIT_COMMIT_MESSAGES": "Fix things",
				"KEY3":                        "runtime_value3",
				"KEY5":                        "runtime_value5",
			},
			wantErr: false,
		},
		{
			name: "runtime_env_overrides_envfile",
			envfileContent: map[string]string{
				"CI":              "true",
				"BITRISE_SRC_DIR": "/bitrise/src",
			},
			runtimeEnvs: envmanModels.EnvsJSONListModel{
				"BITRISE_SRC_DIR": "/bitrise/src", // This should stay as is
				"CI":              "false",        // This should override the envfile value
			},
			expected: map[string]string{
				"BITRISE_SRC_DIR": "/bitrise/src",
				"CI":              "false",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testEnvPath := filepath.Join(t.TempDir(), ".env")

			out, err := yaml.Marshal(&EnvFile{Envs: tt.envfileContent})
			require.NoError(t, err)
			err = os.WriteFile(testEnvPath, out, 0644)
			require.NoError(t, err)

			got, err := MergeEnvfileWithRuntimeEnvs(tt.runtimeEnvs, testEnvPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeEnvfileWithRuntimeEnvs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("MergeEnvfileWithRuntimeEnvs() = %v, want %v", got, tt.expected)
			}
		})
	}
}
