package mise

import (
	"errors"
	"fmt"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/mise/nixpkgs"
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
		{
			name: "strict resolution with nixpkgs backend",
			tool: provider.ToolRequest{
				ToolName:           "nixpkgs:ruby",
				UnparsedVersion:    "3.3.0",
				ResolutionStrategy: provider.ResolutionStrategyStrict,
			},
			want:    "nixpkgs:ruby@3.3.0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			latestInstalledResolver := func(toolName provider.ToolID, version string) (string, error) {
				// Setup fake behavior based on test case.
				switch tt.tool.ToolName {
				case "go":
					// Fake successful resolution.
					return "1.21.5", nil
				case "java":
					// Fake no matching version found.
					return "", errNoMatchingVersion
				case "ruby":
					// Fake other error.
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

func TestCanBeInstalledWithNix(t *testing.T) {
	tests := []struct {
		name               string
		toolID             provider.ToolID
		version            string
		resolutionStrategy provider.ResolutionStrategy
		setupFake          func(m *fakeExecEnv)
		want               bool
	}{
		{
			name:               "concrete Ruby version that exists in index",
			toolID:             provider.ToolID("ruby"),
			version:            "3.3.9",
			resolutionStrategy: provider.ResolutionStrategyStrict,
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(fmt.Sprintf("plugin install %s %s", nixpkgs.PluginName, nixpkgs.PluginGitURL), "")
				m.setResponse(fmt.Sprintf("plugin update %s", nixpkgs.PluginName), "")
				m.setResponse("ls --installed --json --quiet ruby", "[]")
				m.setResponse("ls-remote --quiet nixpkgs:ruby@3.3.9", "3.3.9")
			},
			want: true,
		},
		{
			name:               "fuzzy Ruby version that matches an existing version in index",
			toolID:             provider.ToolID("ruby"),
			version:            "3.3",
			resolutionStrategy: provider.ResolutionStrategyLatestReleased,
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(fmt.Sprintf("plugin install %s %s", nixpkgs.PluginName, nixpkgs.PluginGitURL), "")
				m.setResponse(fmt.Sprintf("plugin update %s", nixpkgs.PluginName), "")
				m.setResponse("ls --installed --json --quiet ruby", "[]")
				m.setResponse("ls-remote --quiet nixpkgs:ruby@3.3", "3.3.8\n3.3.9")
			},
			want: true,
		},
		{
			name:               "concrete Ruby version that doesn't exist in index",
			toolID:             provider.ToolID("ruby"),
			version:            "0.0.1",
			resolutionStrategy: provider.ResolutionStrategyStrict,
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(fmt.Sprintf("plugin install %s %s", nixpkgs.PluginName, nixpkgs.PluginGitURL), "")
				m.setResponse(fmt.Sprintf("plugin update %s", nixpkgs.PluginName), "")
				m.setResponse("ls --installed --json --quiet ruby", "[]")
				m.setResponse("ls-remote --quiet nixpkgs:ruby@0.0.1", "")
			},
			want: false,
		},
		{
			name:               "nixpkgs plugin install error",
			toolID:             provider.ToolID("ruby"),
			version:            "3.3.9",
			resolutionStrategy: provider.ResolutionStrategyStrict,
			setupFake: func(m *fakeExecEnv) {
				m.setError(fmt.Sprintf("plugin install %s %s", nixpkgs.PluginName, nixpkgs.PluginGitURL), fmt.Errorf("fake error"))
			},
			want: false,
		},
		{
			name:               "nixpkgs index check error",
			toolID:             provider.ToolID("ruby"),
			version:            "3.3.9",
			resolutionStrategy: provider.ResolutionStrategyStrict,
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(fmt.Sprintf("plugin install %s %s", nixpkgs.PluginName, nixpkgs.PluginGitURL), "")
				m.setResponse(fmt.Sprintf("plugin update %s", nixpkgs.PluginName), "")
				m.setResponse("ls --installed --json --quiet ruby", "[]")
				m.setError("ls-remote --quiet nixpkgs:ruby@3.3.9", fmt.Errorf("fake error"))
			},
			want: false,
		},
	}

	t.Setenv("BITRISE_TOOLSETUP_FAST_INSTALL", "true")
	// Nix might not be available in the test environment, but we don't actually test Nix functionality here.
	t.Setenv("BITRISE_TEST_SKIP_NIX_CHECK", "true")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execEnv := newFakeExecEnv()
			tt.setupFake(execEnv)

			request := provider.ToolRequest{
				ToolName:           tt.toolID,
				UnparsedVersion:    tt.version,
				ResolutionStrategy: tt.resolutionStrategy,
			}

			got := canBeInstalledWithNix(request, execEnv)
			require.Equal(t, tt.want, got)

		})
	}
}

func TestInstallRequest(t *testing.T) {
	tests := []struct {
		name   string
		tool   provider.ToolRequest
		useNix bool
		want   provider.ToolRequest
	}{
		{
			name: "without nixpkgs",
			tool: provider.ToolRequest{
				ToolName:           "node",
				UnparsedVersion:    "18.20.0",
				ResolutionStrategy: provider.ResolutionStrategyStrict,
			},
			useNix: false,
			want: provider.ToolRequest{
				ToolName:           "node",
				UnparsedVersion:    "18.20.0",
				ResolutionStrategy: provider.ResolutionStrategyStrict,
			},
		},
		{
			name: "with nixpkgs",
			tool: provider.ToolRequest{
				ToolName:           "node",
				UnparsedVersion:    "18",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
			useNix: true,
			want: provider.ToolRequest{
				ToolName:           "nixpkgs:node",
				UnparsedVersion:    "18",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := installRequest(tt.tool, tt.useNix)
			require.Equal(t, tt.want, got)
		})
	}
}
