//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestMiseInstallNixpkgsRuby(t *testing.T) {
	tests := []struct {
		name               string
		tool               string
		version            string
		resolutionStrategy provider.ResolutionStrategy
		want               string
		wantErr            bool
	}{
		{
			name:               "Install specific version",
			tool:               "ruby",
			version:            "3.3.9",
			resolutionStrategy: provider.ResolutionStrategyStrict,
			want:               "3.3.9",
		},
		{
			name:               "Install fuzzy version and released strategy",
			tool:               "ruby",
			version:            "3.1", // EOL version, won't receive new patch versions suddenly
			resolutionStrategy: provider.ResolutionStrategyLatestReleased,
			want:               "3.1.7",
		},
		{
			name:               "Install fuzzy version and installed strategy",
			tool:               "ruby",
			version:            "3.1", // EOL version, won't receive new patch versions
			resolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			want:               "3.1.7",
		},
		{
			name:               "Nonexistent version in nixpkgs index",
			tool:               "ruby",
			version:            "0.1.999",
			resolutionStrategy: provider.ResolutionStrategyStrict,
			wantErr:            true,
		},
		{
			name:               "Install some other tool with forced nixpkgs backend",
			tool:               "node",
			version:            "22.22.1",
			resolutionStrategy: provider.ResolutionStrategyStrict,
			wantErr:            true,
		},
	}

	t.Setenv("BITRISE_TOOLSETUP_FAST_INSTALL", "true")
	t.Setenv("BITRISE_TOOLSETUP_FAST_INSTALL_FORCE", "true")

	for _, tt := range tests {
		miseInstallDir := t.TempDir()
		miseDataDir := t.TempDir()
		miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir)
		require.NoError(t, err)

		err = miseProvider.Bootstrap()
		require.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			request := provider.ToolRequest{
				ToolName:           provider.ToolID(tt.tool),
				UnparsedVersion:    tt.version,
				ResolutionStrategy: tt.resolutionStrategy,
			}
			result, installErr := miseProvider.InstallTool(request)

			if tt.wantErr {
				require.Error(t, installErr)
				return
			}
			require.NoError(t, installErr)
			if tt.tool == "ruby" {
				// We purposely return ruby with the nixpkgs: prefix for environment activation later
				require.Equal(t, provider.ToolID("nixpkgs:ruby"), result.ToolName)
			} else {
				require.Equal(t, provider.ToolID(tt.tool), result.ToolName)
			}
			require.Equal(t, tt.want, result.ConcreteVersion)
			require.False(t, result.IsAlreadyInstalled)
		})
	}
}
