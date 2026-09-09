package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestActivatorOptions covers the env-var reads that used to live in stepman:
// both feature flags default to on and only "false"/"0" opts out.
func TestActivatorOptions(t *testing.T) {
	// The two flags are set independently: with both keys always holding the
	// same value, reading the wrong one for either option would still satisfy
	// every assertion.
	tests := []struct {
		name                  string
		apiValue, binaryValue string // empty means the env var is unset
		wantAPI, wantBinary   bool
	}{
		{name: "Unset enables both", wantAPI: true, wantBinary: true},
		{name: "true enables both", apiValue: "true", binaryValue: "true", wantAPI: true, wantBinary: true},
		{name: "1 enables both", apiValue: "1", binaryValue: "1", wantAPI: true, wantBinary: true},
		{name: "Unrecognized value enables both", apiValue: "maybe", binaryValue: "maybe", wantAPI: true, wantBinary: true},
		{name: "false opts out of both", apiValue: "false", binaryValue: "false"},
		{name: "0 opts out of both", apiValue: "0", binaryValue: "0"},
		{name: "API off alone leaves prebuilt executables on", apiValue: "false", wantBinary: true},
		{name: "Prebuilt executables off alone leaves the API on", binaryValue: "false", wantAPI: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range map[string]string{
				"BITRISE_STEPLIB_USE_API":    tt.apiValue,
				"BITRISE_STEPLIB_USE_BINARY": tt.binaryValue,
			} {
				// t.Setenv registers the restore even when the value is then
				// removed, so an unset case cannot leak into the rest of the suite.
				t.Setenv(key, value)
				if value == "" {
					require.NoError(t, os.Unsetenv(key))
				}
			}

			opts := activatorOptions(false)
			require.Equal(t, tt.wantAPI, !opts.DisableSteplibAPI)
			require.Equal(t, tt.wantBinary, !opts.DisablePrecompiled)
		})
	}
}

func TestActivatorOptions_StorageURLOverride(t *testing.T) {
	t.Setenv("BITRISE_STEPLIB_STORAGE_URLS", "https://a.example.com,https://b.example.com")
	require.Equal(t, []string{"https://a.example.com", "https://b.example.com"},
		activatorOptions(false).PrecompiledStorageURLs)

	require.NoError(t, os.Unsetenv("BITRISE_STEPLIB_STORAGE_URLS"))
	require.Nil(t, activatorOptions(false).PrecompiledStorageURLs,
		"no override leaves the defaults to be filled in by stepman")
}

// Offline mode is a run mode rather than an env var read here: the caller
// resolves it and it reaches stepman as an option instead of being threaded
// through every activation.
func TestActivatorOptions_OfflineMode(t *testing.T) {
	require.False(t, activatorOptions(false).IsOfflineMode)
	require.True(t, activatorOptions(true).IsOfflineMode)
}
