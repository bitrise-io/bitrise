package mise

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirs(t *testing.T) {
	tests := []struct {
		name                string
		xdgDataHome         string
		xdgCacheHome        string
		miseDataDir         string
		homeDir             string
		expectedInstallDir  string
		expectedDataDir     string
	}{
		{
			name:                "default paths",
			xdgDataHome:         "",
			xdgCacheHome:        "",
			miseDataDir:         "",
			homeDir:             "/home/user",
			expectedInstallDir:  "/home/user/.cache/bitrise/toolprovider/mise",
			expectedDataDir:     "/home/user/.local/share/mise",
		},
		{
			name:                "XDG_DATA_HOME and XDG_CACHE_HOME set",
			xdgDataHome:         "/custom/data",
			xdgCacheHome:        "/custom/cache",
			miseDataDir:         "",
			homeDir:             "/home/user",
			expectedInstallDir:  "/custom/cache/bitrise/toolprovider/mise",
			expectedDataDir:     "/custom/data/mise",
		},
		{
			name:                "MISE_DATA_DIR overrides data dir",
			xdgDataHome:         "/custom/data",
			xdgCacheHome:        "/custom/cache",
			miseDataDir:         "/override/mise",
			homeDir:             "/home/user",
			expectedInstallDir:  "/custom/cache/bitrise/toolprovider/mise",
			expectedDataDir:     "/override/mise",
		},
		{
			name:                "MISE_DATA_DIR overrides default data dir",
			xdgDataHome:         "",
			xdgCacheHome:        "",
			miseDataDir:         "/override/mise",
			homeDir:             "/home/user",
			expectedInstallDir:  "/home/user/.cache/bitrise/toolprovider/mise",
			expectedDataDir:     "/override/mise",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", tt.homeDir)
			if tt.xdgDataHome != "" {
				t.Setenv("XDG_DATA_HOME", tt.xdgDataHome)
			}
			if tt.xdgCacheHome != "" {
				t.Setenv("XDG_CACHE_HOME", tt.xdgCacheHome)
			}
			if tt.miseDataDir != "" {
				t.Setenv("MISE_DATA_DIR", tt.miseDataDir)
			}

			installDir, dataDir := Dirs()

			require.Equal(t, tt.expectedInstallDir, installDir)
			require.Equal(t, tt.expectedDataDir, dataDir)
		})
	}
}
