//go:build linux_and_mac
// +build linux_and_mac

package mise

import (
	"context"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestMiseInstallNixpkgsRuby(t *testing.T) {
	tests := []struct {
		name               string
		requestedVersion   string
		resolutionStrategy provider.ResolutionStrategy
		expectedVersion    string
	}{
		{"Install specific version", "3.3.9", provider.ResolutionStrategyStrict, "3.3.9"},
	}

	for _, tt := range tests {
		miseInstallDir := t.TempDir()
		miseDataDir := t.TempDir()
		miseProvider, err := mise.NewToolProvider(miseInstallDir, miseDataDir)
		require.NoError(t, err)

		err = miseProvider.Bootstrap()
		require.NoError(t, err)

		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			done := make(chan bool)
			var result provider.ToolInstallResult
			var installErr error

			go func() {
				request := provider.ToolRequest{
					ToolName:           "ruby",
					UnparsedVersion:    tt.requestedVersion,
					ResolutionStrategy: tt.resolutionStrategy,
				}
				result, installErr = miseProvider.InstallTool(request)
				done <- true
			}()

			select {
			case <-done:
				require.NoError(t, installErr)
				require.Equal(t, provider.ToolID("ruby"), result.ToolName)
				require.Equal(t, tt.expectedVersion, result.ConcreteVersion)
				require.False(t, result.IsAlreadyInstalled)
			case <-ctx.Done():
				t.Fatal("Test exceeded 1 minute timeout, installation was too slow for nixpkgs ruby")
			}
		})
	}
}
