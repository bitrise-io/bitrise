package mise

import (
	"errors"
	"fmt"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/require"
)

func TestMiseVersionString(t *testing.T) {
	tests := []struct {
		name    string
		tool    provider.ToolRequest
		want    string
		wantErr bool
	}{
		{
			name: "strict resolution strategy, exact version",
			tool: provider.ToolRequest{
				ToolName:           "node",
				UnparsedVersion:    "18.20.0",
				ResolutionStrategy: provider.ResolutionStrategyStrict,
			},
			want:    "node@18.20.0",
			wantErr: false,
		},
		{
			name: "latest released resolution strategy, partial version",
			tool: provider.ToolRequest{
				ToolName:           "python",
				UnparsedVersion:    "3.11",
				ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
			},
			want:    "python@prefix:3.11",
			wantErr: false,
		},
		{
			name: "latest installed resolution strategy, partial version, version found",
			tool: provider.ToolRequest{
				ToolName:           "go",
				UnparsedVersion:    "1.21",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
			want:    "go@1.21.5",
			wantErr: false,
		},
		{
			name: "latest installed resolution strategy, no version found, fallback to latest released",
			tool: provider.ToolRequest{
				ToolName:           "java",
				UnparsedVersion:    "17",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
			want:    "java@prefix:17",
			wantErr: false,
		},
		{
			name: "latest installed resolution strategy - error resolving",
			tool: provider.ToolRequest{
				ToolName:           "ruby",
				UnparsedVersion:    "3.0",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unknown resolution strategy",
			tool: provider.ToolRequest{
				ToolName:           "node",
				UnparsedVersion:    "18.0.0",
				ResolutionStrategy: provider.ResolutionStrategy(999),
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			latestInstalledResolver := func(toolName provider.ToolID, version string) (string, error) {
				// Setup fake behavior based on test case
				switch tt.tool.ToolName {
				case "go":
					// Fake successful resolution
					return "1.21.5", nil
				case "java":
					// Fake no matching version found
					return "", errNoMatchingVersion
				case "ruby":
					// Fake other error
					return "", errors.New("some other error")
				}
				return "", fmt.Errorf("no fake behavior defined for tool %s", toolName)
			}

			got, err := miseVersionString(tt.tool, latestInstalledResolver)

			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIsAlreadyInstalled(t *testing.T) {
	tests := []struct {
		name                 string
		tool                 provider.ToolRequest
		latestInstalledError error
		want                 bool
		wantErr              bool
	}{
		{
			name: "tool is already installed",
			tool: provider.ToolRequest{
				ToolName:        "node",
				UnparsedVersion: "18.20.0",
			},
			latestInstalledError: nil,
			want:                 true,
			wantErr:              false,
		},
		{
			name: "tool is not installed - no matching version",
			tool: provider.ToolRequest{
				ToolName:        "python",
				UnparsedVersion: "3.11",
			},
			latestInstalledError: errNoMatchingVersion,
			want:                 false,
			wantErr:              false,
		},
		{
			name: "error resolving installed versions",
			tool: provider.ToolRequest{
				ToolName:        "ruby",
				UnparsedVersion: "3.0",
			},
			latestInstalledError: errors.New("failed to list installed versions"),
			want:                 false,
			wantErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			latestInstalledResolver := func(toolName provider.ToolID, version string) (string, error) {
				if tt.latestInstalledError != nil {
					return "", tt.latestInstalledError
				}
				return "fake.version", nil
			}

			got, err := isAlreadyInstalled(tt.tool, latestInstalledResolver)

			if tt.wantErr {
				require.Error(t, err)
				require.False(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
