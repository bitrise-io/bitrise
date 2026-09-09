package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestActivatorOptions covers the env-var reads that used to live in stepman:
// both feature flags default to on and only "false"/"0" opts out.
func TestActivatorOptions(t *testing.T) {
	tests := []struct {
		name                string
		envValue            string // empty means the env var is unset
		wantAPI, wantBinary bool
	}{
		{name: "Unset enables both", wantAPI: true, wantBinary: true},
		{name: "true enables both", envValue: "true", wantAPI: true, wantBinary: true},
		{name: "1 enables both", envValue: "1", wantAPI: true, wantBinary: true},
		{name: "Unrecognized value enables both", envValue: "maybe", wantAPI: true, wantBinary: true},
		{name: "false opts out of both", envValue: "false"},
		{name: "0 opts out of both", envValue: "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range []string{"BITRISE_STEPLIB_USE_API", "BITRISE_STEPLIB_USE_BINARY"} {
				// t.Setenv registers the restore even when the value is then
				// removed, so an unset case cannot leak into the rest of the suite.
				t.Setenv(key, tt.envValue)
				if tt.envValue == "" {
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
