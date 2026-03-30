package workarounds

import (
	"errors"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestAdjustFlutterStableVersion(t *testing.T) {
	tests := []struct {
		name                    string
		toolName                provider.ToolID
		version                 string
		silent                  bool
		versionChecks           map[string]bool
		versionCheckError       error
		expectedAdjustedVersion string
		expectedError           string
	}{
		{
			name:                    "Not flutter tool - returns empty",
			toolName:                "node",
			version:                 "20.0.0-stable",
			silent:                  true,
			versionChecks:           map[string]bool{},
			expectedAdjustedVersion: "",
		},
		{
			name:                    "Version without -stable suffix - returns empty",
			toolName:                "flutter",
			version:                 "3.32.1",
			silent:                  true,
			versionChecks:           map[string]bool{},
			expectedAdjustedVersion: "",
		},
		{
			name:     "Version with -stable exists remotely - returns empty",
			toolName: "flutter",
			version:  "3.32.1-stable",
			silent:   true,
			versionChecks: map[string]bool{
				"3.32.1-stable": true,
			},
			expectedAdjustedVersion: "",
		},
		{
			name:     "Version with -stable doesn't exist, fallback exists - returns fallback",
			toolName: "flutter",
			version:  "3.32.1-stable",
			silent:   true,
			versionChecks: map[string]bool{
				"3.32.1-stable": false,
				"3.32.1":        true,
			},
			expectedAdjustedVersion: "3.32.1",
		},
		{
			name:     "Version with -stable doesn't exist, fallback also doesn't exist - returns empty",
			toolName: "flutter",
			version:  "3.32.1-stable",
			silent:   true,
			versionChecks: map[string]bool{
				"3.32.1-stable": false,
				"3.32.1":        false,
			},
			expectedAdjustedVersion: "",
		},
		{
			name:     "Silent mode false - should still work",
			toolName: "flutter",
			version:  "3.32.1-stable",
			silent:   false,
			versionChecks: map[string]bool{
				"3.32.1-stable": false,
				"3.32.1":        true,
			},
			expectedAdjustedVersion: "3.32.1",
		},
		{
			name:              "Error checking original version - returns error",
			toolName:          "flutter",
			version:           "3.32.1-stable",
			silent:            true,
			versionCheckError: errors.New("network error"),
			expectedError:     "check if flutter 3.32.1-stable exists remotely: network error",
		},
		{
			name:     "Error checking fallback version - returns error",
			toolName: "flutter",
			version:  "3.32.1-stable",
			silent:   true,
			versionChecks: map[string]bool{
				"3.32.1-stable": false,
			},
			versionCheckError: errors.New("network error"),
			expectedError:     "check if flutter 3.32.1 exists remotely: network error",
		},
		{
			name:     "Version starting with -stable suffix",
			toolName: "flutter",
			version:  "-stable",
			silent:   true,
			versionChecks: map[string]bool{
				"-stable": false,
				"":        false,
			},
			expectedAdjustedVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			mockVersionChecker := func(toolName provider.ToolID, version string) (bool, error) {
				callCount++

				if tt.versionCheckError != nil {
					if tt.name == "Error checking original version - returns error" && callCount == 1 {
						return false, tt.versionCheckError
					}
					if tt.name == "Error checking fallback version - returns error" && callCount == 2 {
						return false, tt.versionCheckError
					}
				}

				exists, ok := tt.versionChecks[version]
				if !ok {
					t.Errorf("unexpected version check call for version: %s", version)
					return false, nil
				}
				return exists, nil
			}

			adjustedVersion, err := AdjustFlutterStableVersion(
				mockVersionChecker,
				tt.toolName,
				tt.version,
				tt.silent,
			)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedAdjustedVersion, adjustedVersion)
			}
		})
	}
}
