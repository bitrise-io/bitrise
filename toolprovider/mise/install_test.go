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
		name            string
		toolName        provider.ToolID
		concreteVersion string
		want            string
	}{
		{
			name:            "concrete node version",
			toolName:        "node",
			concreteVersion: "18.20.0",
			want:            "node@18.20.0",
		},
		{
			name:            "concrete python version",
			toolName:        "python",
			concreteVersion: "3.11.5",
			want:            "python@3.11.5",
		},
		{
			name:            "nixpkgs backend",
			toolName:        "nixpkgs:ruby",
			concreteVersion: "3.3.0",
			want:            "nixpkgs:ruby@3.3.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := miseVersionString(tt.toolName, tt.concreteVersion)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestIsAlreadyInstalled(t *testing.T) {
	tests := []struct {
		name            string
		toolName        provider.ToolID
		concreteVersion string
		setupFake       func(*fakeExecEnv)
		want            bool
		wantErr         bool
	}{
		{
			name:            "concrete version is already installed",
			toolName:        "node",
			concreteVersion: "18.20.0",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLsInstalledCmd("node"), `[{"version": "18.20.0", "installed": true}]`)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:            "concrete version is not installed",
			toolName:        "python",
			concreteVersion: "3.11.5",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLsInstalledCmd("python"), `[{"version": "3.11.4", "installed": true}]`)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:            "no versions installed",
			toolName:        "ruby",
			concreteVersion: "3.3.0",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLsInstalledCmd("ruby"), "[]")
			},
			want:    false,
			wantErr: false,
		},
		{
			name:            "error listing installed versions",
			toolName:        "go",
			concreteVersion: "1.21.5",
			setupFake: func(m *fakeExecEnv) {
				m.setError(miseLsInstalledCmd("go"), errors.New("failed to list"))
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeExecEnv()
			tt.setupFake(fake)

			provider := &MiseToolProvider{ExecEnv: fake}
			got, err := provider.isAlreadyInstalled(tt.toolName, tt.concreteVersion)

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
				m.setResponse("settings experimental=true", "")
				m.setResponse(fmt.Sprintf("plugin install %s %s", nixpkgs.PluginName, nixpkgs.PluginGitURL), "")
				m.setResponse(fmt.Sprintf("plugin update %s", nixpkgs.PluginName), "")
				m.setResponse("ls --installed --json --quiet ruby", "[]")
				m.setResponse("ls-remote --quiet --json nixpkgs:ruby@3.3.9", `[{"version":"3.3.9"}]`)
			},
			want: true,
		},
		{
			name:               "fuzzy Ruby version that matches an existing version in index",
			toolID:             provider.ToolID("ruby"),
			version:            "3.3",
			resolutionStrategy: provider.ResolutionStrategyLatestReleased,
			setupFake: func(m *fakeExecEnv) {
				m.setResponse("settings experimental=true", "")
				m.setResponse(fmt.Sprintf("plugin install %s %s", nixpkgs.PluginName, nixpkgs.PluginGitURL), "")
				m.setResponse(fmt.Sprintf("plugin update %s", nixpkgs.PluginName), "")
				m.setResponse("ls --installed --json --quiet ruby", "[]")
				m.setResponse("ls-remote --quiet --json nixpkgs:ruby@3.3", `[{"version":"3.3.8"},{"version":"3.3.9"}]`)
			},
			want: true,
		},
		{
			name:               "concrete Ruby version that doesn't exist in index",
			toolID:             provider.ToolID("ruby"),
			version:            "0.0.1",
			resolutionStrategy: provider.ResolutionStrategyStrict,
			setupFake: func(m *fakeExecEnv) {
				m.setResponse("settings experimental=true", "")
				m.setResponse(fmt.Sprintf("plugin install %s %s", nixpkgs.PluginName, nixpkgs.PluginGitURL), "")
				m.setResponse(fmt.Sprintf("plugin update %s", nixpkgs.PluginName), "")
				m.setResponse("ls --installed --json --quiet ruby", "[]")
				m.setResponse("ls-remote --quiet --json nixpkgs:ruby@0.0.1", "[]")
			},
			want: false,
		},
		{
			name:               "nixpkgs plugin install error",
			toolID:             provider.ToolID("ruby"),
			version:            "3.3.9",
			resolutionStrategy: provider.ResolutionStrategyStrict,
			setupFake: func(m *fakeExecEnv) {
				m.setResponse("settings experimental=true", "")
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
				m.setResponse("settings experimental=true", "")
				m.setResponse(fmt.Sprintf("plugin install %s %s", nixpkgs.PluginName, nixpkgs.PluginGitURL), "")
				m.setResponse(fmt.Sprintf("plugin update %s", nixpkgs.PluginName), "")
				m.setResponse("ls --installed --json --quiet ruby", "[]")
				m.setError("ls-remote --quiet --json nixpkgs:ruby@3.3.9", fmt.Errorf("fake error"))
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execEnv := newFakeExecEnv()
			tt.setupFake(execEnv)

			request := provider.ToolRequest{
				ToolName:           tt.toolID,
				UnparsedVersion:    tt.version,
				ResolutionStrategy: tt.resolutionStrategy,
			}

			nixChecker := func(tool provider.ToolRequest, silent bool) bool {
				return true
			}

			got := canBeInstalledWithNix(request, execEnv, true, nixChecker, false)
			require.Equal(t, tt.want, got)

		})
	}
}

func TestInstallToolVersion(t *testing.T) {
	tests := []struct {
		name            string
		toolName        provider.ToolID
		concreteVersion string
		mockErr         error
		wantCause       string
		wantRecommend   string
		wantRawOutput   string
	}{
		{
			name:            "success",
			toolName:        "node",
			concreteVersion: "18.20.0",
			mockErr:         nil,
		},
		{
			name:            "generic failure has no GitHub recommendation",
			toolName:        "node",
			concreteVersion: "18.20.0",
			mockErr:         errors.New("exit status 1\nsome unrelated mise error"),
			wantCause:       "mise install node@18.20.0: exit status 1\nsome unrelated mise error",
			wantRecommend:   "",
		},
		{
			name:            "GitHub rate limit error",
			toolName:        "tuist",
			concreteVersion: "4.70.0",
			mockErr:         errors.New("exit status 1\nmise ERROR Failed to install aqua:tuist/tuist@4.70.0: HTTP status client error (403 rate limit exceeded) for url (https://api.github.com/repos/tuist/tuist/releases/tags/4.70.0)"),
			wantCause:       "GitHub API rate limit exceeded while installing tuist",
			wantRecommend:   "GITHUB_TOKEN, GH_TOKEN, GITHUB_API_TOKEN, MISE_GITHUB_TOKEN, MISE_GITHUB_ENTERPRISE_TOKEN",
			wantRawOutput:   "exit status 1\nmise ERROR Failed to install aqua:tuist/tuist@4.70.0: HTTP status client error (403 rate limit exceeded) for url (https://api.github.com/repos/tuist/tuist/releases/tags/4.70.0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeExecEnv()
			versionString := miseVersionString(tt.toolName, tt.concreteVersion)
			cmdKey := fmt.Sprintf("install --yes %s", versionString)
			if tt.mockErr != nil {
				fake.setError(cmdKey, tt.mockErr)
			} else {
				fake.setResponse(cmdKey, "")
			}

			p := &MiseToolProvider{ExecEnv: fake}
			err := p.installToolVersion(tt.toolName, tt.concreteVersion)

			if tt.mockErr == nil {
				require.NoError(t, err)
				return
			}

			var toolErr provider.ToolInstallError
			require.ErrorAs(t, err, &toolErr)
			require.Equal(t, tt.wantCause, toolErr.Cause)
			if tt.wantRecommend == "" {
				require.Empty(t, toolErr.Recommendation)
			} else {
				require.Contains(t, toolErr.Recommendation, tt.wantRecommend)
			}
			if tt.wantRawOutput != "" {
				require.Equal(t, tt.wantRawOutput, toolErr.RawOutput)
			}
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
