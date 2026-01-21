package mise

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

func TestExtractLastLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line",
			input:    "4.116.2",
			expected: "4.116.2",
		},
		{
			name:     "single line with whitespace",
			input:    "  4.116.2  \n",
			expected: "4.116.2",
		},
		{
			name: "multiple lines with plugin installation",
			input: `mise plugin:tuist    clone https://github.com/mise-plugins/mise-tuist.git
mise plugin:tuist  ✓ https://github.com/mise-plugins/mise-tuist.git#a24ea40
4.116.2`,
			expected: "4.116.2",
		},
		{
			name: "multiple lines with trailing whitespace",
			input: `mise plugin:tuist    clone https://github.com/mise-plugins/mise-tuist.git
mise plugin:tuist  ✓ https://github.com/mise-plugins/mise-tuist.git#a24ea40
4.116.2
`,
			expected: "4.116.2",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "  \n  \n  ",
			expected: "",
		},
		{
			name: "multiple lines all with content",
			input: `line1
line2
line3`,
			expected: "line3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLastLine(tt.input)
			if result != tt.expected {
				t.Errorf("extractLastLine() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Helper functions to construct mise command strings for mocking.

func miseLatestCmd(tool provider.ToolID, version string) string {
	return fmt.Sprintf("latest --quiet %s@%s", tool, version)
}

func miseLatestInstalledCmd(tool provider.ToolID, version string) string {
	var toolString = string(tool)
	if version != "" && version != "installed" {
		toolString = fmt.Sprintf("%s@%s", tool, version)
	}
	return fmt.Sprintf("latest --installed --quiet %s", toolString)
}

func miseLsInstalledCmd(tool provider.ToolID) string {
	return fmt.Sprintf("ls --installed --json --quiet %s", tool)
}

func miseLsRemoteCmd(tool provider.ToolID, version string) string {
	if version != "" && version != "latest" && version != "installed" {
		return fmt.Sprintf("ls-remote --quiet %s@%s", tool, version)
	}
	return fmt.Sprintf("ls-remote --quiet %s", tool)
}

var installedVersionsJSON = `[
  {
    "version": "3.3.8",
    "requested_version": "3.3",
    "install_path": "/Users/vagrant/.local/share/mise/installs/ruby/3.3.8",
    "source": {
      "type": ".tool-versions",
      "path": "/Users/vagrant/.tool-versions"
    },
    "installed": true,
    "active": true
  },
  {
    "version": "3.4.5",
    "install_path": "/Users/vagrant/.local/share/mise/installs/ruby/3.4.5",
    "installed": true,
    "active": false
  }
]`

func TestParseInstalledVersionsJSON(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		exists  bool
		wantErr bool
	}{
		{"empty", "[]", false, false},
		{"oneInstalled", `[{"installed":true}]`, true, false},
		{"multipleMixed", `[{"installed":false},{"installed":true}]`, true, false},
		{"multipleInstalled", installedVersionsJSON, true, false},
		{"noneInstalled", `[{"installed":false}]`, false, false},
		{"malformed", `{"installed":true}`, false, true},
	}
	for _, tc := range cases {
		got, err := parseInstalledVersionsJSON(tc.input, "")
		if tc.wantErr && err == nil {
			// nolint:staticcheck
			t.Fatalf("%s: expected error, got nil", tc.name)
		}
		if !tc.wantErr && err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		if got != tc.exists {
			t.Fatalf("%s: expected exists=%v got %v", tc.name, tc.exists, got)
		}
	}
}

func TestResolveToLatestReleased(t *testing.T) {
	tests := []struct {
		name            string
		toolName        provider.ToolID
		version         string
		setupFake       func(*fakeExecEnv)
		expectedVersion string
		wantErr         bool
	}{
		{
			name:     "resolve specific version",
			toolName: "ruby",
			version:  "3.3",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestCmd("ruby", "3.3"), "3.3.8")
			},
			expectedVersion: "3.3.8",
			wantErr:         false,
		},
		{
			name:     "resolve latest version with empty string",
			toolName: "node",
			version:  "",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestCmd("node", ""), "22.11.0")
			},
			expectedVersion: "22.11.0",
			wantErr:         false,
		},
		{
			name:     "resolve exact version",
			toolName: "go",
			version:  "1.23.0",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestCmd("go", "1.23.0"), "1.23.0")
			},
			expectedVersion: "1.23.0",
			wantErr:         false,
		},
		{
			name:     "no matching version - empty output",
			toolName: "python",
			version:  "9.9",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestCmd("python", "9.9"), "")
			},
			expectedVersion: "",
			wantErr:         true,
		},
		{
			name:     "command error",
			toolName: "java",
			version:  "17",
			setupFake: func(m *fakeExecEnv) {
				m.setError(miseLatestCmd("java", "17"), fmt.Errorf("network timeout"))
			},
			expectedVersion: "",
			wantErr:         true,
		},
		{
			name:     "output with whitespace is trimmed",
			toolName: "ruby",
			version:  "3.4",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestCmd("ruby", "3.4"), "  3.4.1  \n")
			},
			expectedVersion: "3.4.1",
			wantErr:         false,
		},
		{
			name:     "fuzzy version prefix",
			toolName: "node",
			version:  "20",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestCmd("node", "20"), "20.18.1")
			},
			expectedVersion: "20.18.1",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newFakeExecEnv()
			tt.setupFake(mock)

			version, err := resolveToLatestReleased(mock, tt.toolName, tt.version)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if version != tt.expectedVersion {
					t.Errorf("expected version=%q, got %q", tt.expectedVersion, version)
				}
			}
		})
	}
}

func TestResolveToLatestInstalled(t *testing.T) {
	tests := []struct {
		name            string
		toolName        provider.ToolID
		version         string
		setupFake       func(*fakeExecEnv)
		expectedVersion string
		wantErr         bool
	}{
		{
			name:     "resolve latest installed for fuzzy version",
			toolName: "ruby",
			version:  "3.3",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("ruby", "3.3"), "3.3.8")
			},
			expectedVersion: "3.3.8",
			wantErr:         false,
		},
		{
			name:     "resolve latest installed with empty version",
			toolName: "node",
			version:  "",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("node", ""), "22.10.0")
			},
			expectedVersion: "22.10.0",
			wantErr:         false,
		},
		{
			name:     "exact version already installed",
			toolName: "go",
			version:  "1.23.0",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("go", "1.23.0"), "1.23.0")
			},
			expectedVersion: "1.23.0",
			wantErr:         false,
		},
		{
			name:     "no matching installed version - empty output",
			toolName: "python",
			version:  "3.12",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("python", "3.12"), "")
			},
			expectedVersion: "",
			wantErr:         true,
		},
		{
			name:     "command error - not installed",
			toolName: "java",
			version:  "21",
			setupFake: func(m *fakeExecEnv) {
				m.setError(miseLatestInstalledCmd("java", "21"), fmt.Errorf("no versions installed"))
			},
			expectedVersion: "",
			wantErr:         true,
		},
		{
			name:     "output with whitespace is trimmed",
			toolName: "ruby",
			version:  "3.4",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("ruby", "3.4"), "\n  3.4.1  \n")
			},
			expectedVersion: "3.4.1",
			wantErr:         false,
		},
		{
			name:     "multiple versions installed - returns latest",
			toolName: "node",
			version:  "20",
			setupFake: func(m *fakeExecEnv) {
				// mise latest returns only the highest matching version
				m.setResponse(miseLatestInstalledCmd("node", "20"), "20.18.0")
			},
			expectedVersion: "20.18.0",
			wantErr:         false,
		},
		{
			name:     "latest installed without version constraint",
			toolName: "python",
			version:  "",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("python", ""), "3.13.0")
			},
			expectedVersion: "3.13.0",
			wantErr:         false,
		},
		{
			name:     "tool name with backend prefix",
			toolName: "nixpkgs:ruby",
			version:  "",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("nixpkgs:ruby", ""), "3.3.7")
				m.setResponse(miseLatestInstalledCmd("ruby", ""), "3.13.0")
			},
			expectedVersion: "3.3.7",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newFakeExecEnv()
			tt.setupFake(mock)

			version, err := resolveToLatestInstalled(mock, tt.toolName, tt.version)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if version != tt.expectedVersion {
					t.Errorf("expected version=%q, got %q", tt.expectedVersion, version)
				}
			}
		})
	}
}

func TestVersionExistsRemote(t *testing.T) {
	tests := []struct {
		name           string
		toolName       provider.ToolID
		version        string
		setupFake      func(*fakeExecEnv)
		expectedExists bool
		wantErr        bool
	}{
		{
			name:     "version exists in ls-remote",
			toolName: "ruby",
			version:  "3.3.0",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLsRemoteCmd("ruby", "3.3.0"), "3.3.0\n3.3.1\n3.3.2")
			},
			expectedExists: true,
			wantErr:        false,
		},
		{
			name:     "version does not exist in ls-remote",
			toolName: "ruby",
			version:  "9.9.9",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLsRemoteCmd("ruby", "9.9.9"), "")
			},
			expectedExists: false,
			wantErr:        false,
		},
		{
			name:     "latest version",
			toolName: "go",
			version:  "latest",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLsRemoteCmd("go", "latest"), "1.21.0\n1.22.0\n1.23.0")
			},
			expectedExists: true,
			wantErr:        false,
		},
		{
			name:     "empty version defaults to tool name search",
			toolName: "python",
			version:  "",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLsRemoteCmd("python", ""), "3.11.0\n3.12.0")
			},
			expectedExists: true,
			wantErr:        false,
		},
		{
			name:     "ls-remote error",
			toolName: "java",
			version:  "17",
			setupFake: func(m *fakeExecEnv) {
				m.setError(miseLsRemoteCmd("java", "17"), fmt.Errorf("network error"))
			},
			expectedExists: false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeExecEnv()
			tt.setupFake(fake)

			exists, err := versionExistsRemote(fake, tt.toolName, tt.version)

			if tt.wantErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if exists != tt.expectedExists {
				t.Errorf("expected exists=%v, got %v", tt.expectedExists, exists)
			}
		})
	}
}

func TestNormalizeRequest(t *testing.T) {
	tests := []struct {
		name            string
		tool            provider.ToolRequest
		setupFake       func(*fakeExecEnv)
		expectedRequest provider.ToolRequest
		wantErr         bool
	}{
		{
			name: "Strict strategy - no normalization",
			tool: provider.ToolRequest{
				ToolName:           "node",
				UnparsedVersion:    "20.18.1",
				ResolutionStrategy: provider.ResolutionStrategyStrict,
			},
			setupFake: func(m *fakeExecEnv) {},
			expectedRequest: provider.ToolRequest{
				ToolName:           "node",
				UnparsedVersion:    "20.18.1",
				ResolutionStrategy: provider.ResolutionStrategyStrict,
			},
			wantErr: false,
		},
		{
			name: "LatestReleased strategy - no normalization",
			tool: provider.ToolRequest{
				ToolName:           "python",
				UnparsedVersion:    "3.11",
				ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
			},
			setupFake: func(m *fakeExecEnv) {},
			expectedRequest: provider.ToolRequest{
				ToolName:           "python",
				UnparsedVersion:    "3.11",
				ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
			},
			wantErr: false,
		},
		{
			name: "LatestInstalled strategy - version is installed",
			tool: provider.ToolRequest{
				ToolName:           "ruby",
				UnparsedVersion:    "3.3",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("ruby", "3.3"), "3.3.8")
			},
			expectedRequest: provider.ToolRequest{
				ToolName:           "ruby",
				UnparsedVersion:    "3.3",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
			wantErr: false,
		},
		{
			name: "LatestInstalled strategy - no version installed, normalizes to LatestReleased",
			tool: provider.ToolRequest{
				ToolName:           "go",
				UnparsedVersion:    "1.21",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("go", "1.21"), "")
			},
			expectedRequest: provider.ToolRequest{
				ToolName:           "go",
				UnparsedVersion:    "1.21",
				ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
			},
			wantErr: false,
		},
		{
			name: "LatestInstalled strategy - command error",
			tool: provider.ToolRequest{
				ToolName:           "java",
				UnparsedVersion:    "17",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
			setupFake: func(m *fakeExecEnv) {
				m.setError(miseLatestInstalledCmd("java", "17"), fmt.Errorf("command failed"))
			},
			expectedRequest: provider.ToolRequest{
				ToolName:           "java",
				UnparsedVersion:    "17",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
			wantErr: true,
		},
		{
			name: "Installed keyword - version is installed",
			tool: provider.ToolRequest{
				ToolName:           "node",
				UnparsedVersion:    "installed",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("node", ""), "20.18.1")
			},
			expectedRequest: provider.ToolRequest{
				ToolName:           "node",
				UnparsedVersion:    "",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
			wantErr: false,
		},
		{
			name: "Installed keyword - no version installed, normalizes to LatestReleased",
			tool: provider.ToolRequest{
				ToolName:           "python",
				UnparsedVersion:    "installed",
				ResolutionStrategy: provider.ResolutionStrategyLatestInstalled,
			},
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("python", ""), "")
			},
			expectedRequest: provider.ToolRequest{
				ToolName:           "python",
				UnparsedVersion:    "",
				ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
			},
			wantErr: false,
		},
		{
			name: "Latest keyword",
			tool: provider.ToolRequest{
				ToolName:           "node",
				UnparsedVersion:    "latest",
				ResolutionStrategy: provider.ResolutionStrategyStrict,
			},
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("node", ""), "20.18.1")
			},
			expectedRequest: provider.ToolRequest{
				ToolName:           "node",
				UnparsedVersion:    "",
				ResolutionStrategy: provider.ResolutionStrategyLatestReleased,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeExecEnv()
			tt.setupFake(fake)

			request, err := normalizeRequest(fake, tt.tool, false)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %+v", err)
				}
				if request != tt.expectedRequest {
					t.Errorf("expected request=%+v, got %+v", tt.expectedRequest, request)
				}
			}
		})
	}
}

func TestResolveToConcreteVersion(t *testing.T) {
	tests := []struct {
		name            string
		toolName        provider.ToolID
		version         string
		strategy        provider.ResolutionStrategy
		setupFake       func(*fakeExecEnv)
		expectedVersion string
		wantErr         bool
	}{
		{
			name:     "Strict strategy resolves fuzzy version",
			toolName: "node",
			version:  "20.18.1",
			strategy: provider.ResolutionStrategyStrict,
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestCmd("node", "20.18.1"), "20.18.1")
			},
			expectedVersion: "20.18.1",
			wantErr:         false,
		},
		{
			name:     "LatestReleased strategy resolves fuzzy version",
			toolName: "python",
			version:  "3.11",
			strategy: provider.ResolutionStrategyLatestReleased,
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestCmd("python", "3.11"), "3.11.8")
			},
			expectedVersion: "3.11.8",
			wantErr:         false,
		},
		{
			name:     "LatestInstalled strategy resolves to installed version",
			toolName: "ruby",
			version:  "3.3",
			strategy: provider.ResolutionStrategyLatestInstalled,
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestInstalledCmd("ruby", "3.3"), "3.3.8")
			},
			expectedVersion: "3.3.8",
			wantErr:         false,
		},
		{
			name:     "LatestReleased strategy with empty version",
			toolName: "go",
			version:  "",
			strategy: provider.ResolutionStrategyLatestReleased,
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestCmd("go", ""), "1.23.5")
			},
			expectedVersion: "1.23.5",
			wantErr:         false,
		},
		{
			name:     "Resolution error",
			toolName: "java",
			version:  "999",
			strategy: provider.ResolutionStrategyLatestReleased,
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLatestCmd("java", "999"), "")
			},
			expectedVersion: "",
			wantErr:         true,
		},
		{
			name:            "Unknown strategy",
			toolName:        "node",
			version:         "20",
			strategy:        provider.ResolutionStrategy(999),
			setupFake:       func(m *fakeExecEnv) {},
			expectedVersion: "",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeExecEnv()
			tt.setupFake(fake)

			version, err := resolveToConcreteVersion(fake, tt.toolName, tt.version, tt.strategy)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if version != tt.expectedVersion {
					t.Errorf("expected version=%q, got %q", tt.expectedVersion, version)
				}
			}
		})
	}
}
