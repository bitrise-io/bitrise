package mise

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	homeDir            = "/home/user"
	expectedInstallDir = "/home/user/.cache/bitrise/toolprovider/mise/"
	customInstallDir   = "/custom/cache/bitrise/toolprovider/mise/"
	expectedDataDir    = "/home/user/.local/share/mise"
	customDataDir      = "/custom/data/mise"
	customMiseDataDir  = "/override/mise"
	customDataHome     = "/custom/data"
	customCacheHome    = "/custom/cache"
)

func TestDirs(t *testing.T) {
	sanitizedVersion := GetMiseVersion()

	tests := []struct {
		name               string
		xdgDataHome        string
		xdgCacheHome       string
		miseDataDir        string
		homeDir            string
		expectedInstallDir string
		expectedDataDir    string
	}{
		{
			name:               "default paths",
			xdgDataHome:        "",
			xdgCacheHome:       "",
			miseDataDir:        "",
			homeDir:            homeDir,
			expectedInstallDir: expectedInstallDir + sanitizedVersion,
			expectedDataDir:    expectedDataDir,
		},
		{
			name:               "XDG_DATA_HOME and XDG_CACHE_HOME set",
			xdgDataHome:        customDataHome,
			xdgCacheHome:       customCacheHome,
			miseDataDir:        "",
			homeDir:            homeDir,
			expectedInstallDir: customInstallDir + sanitizedVersion,
			expectedDataDir:    customDataDir,
		},
		{
			name:               "MISE_DATA_DIR overrides data dir",
			xdgDataHome:        customDataHome,
			xdgCacheHome:       customCacheHome,
			miseDataDir:        customMiseDataDir,
			homeDir:            homeDir,
			expectedInstallDir: customInstallDir + sanitizedVersion,
			expectedDataDir:    customMiseDataDir,
		},
		{
			name:               "MISE_DATA_DIR overrides default data dir",
			xdgDataHome:        "",
			xdgCacheHome:       "",
			miseDataDir:        customMiseDataDir,
			homeDir:            homeDir,
			expectedInstallDir: expectedInstallDir + sanitizedVersion,
			expectedDataDir:    customMiseDataDir,
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

			installDir, dataDir := Dirs(GetMiseVersion())

			require.Equal(t, tt.expectedInstallDir, installDir)
			require.Equal(t, tt.expectedDataDir, dataDir)
		})
	}
}
