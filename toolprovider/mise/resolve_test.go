package mise

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// Helper functions to construct mise command strings for mocking

func miseLatestCmd(tool provider.ToolID, version string) string {
	return fmt.Sprintf("latest %s@%s", tool, version)
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
		got, err := parseInstalledVersionsJSON(tc.input)
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

func TestVersionExists(t *testing.T) {
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
			name:     "installed version exists",
			toolName: "ruby",
			version:  "installed",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLsInstalledCmd("ruby"), installedVersionsJSON)
			},
			expectedExists: true,
			wantErr:        false,
		},
		{
			name:     "installed version does not exist - empty array",
			toolName: "ruby",
			version:  "installed",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLsInstalledCmd("ruby"), "[]")
				// ls-remote called for fallback
				m.setResponse(miseLsRemoteCmd("ruby", ""), "3.3.0\n3.3.1")
			},
			expectedExists: true, // No installed versions, but remote has versions
			wantErr:        false,
		},
		{
			name:     "installed version with none actually installed - falls through to ls-remote",
			toolName: "ruby",
			version:  "installed",
			setupFake: func(m *fakeExecEnv) {
				// Entries exist but none are installed
				m.setResponse(miseLsInstalledCmd("ruby"), `[{"installed":false}]`)
				// Falls through to ls-remote
				m.setResponse(miseLsRemoteCmd("ruby", ""), "3.3.0\n3.3.1")
			},
			expectedExists: true, // Falls through to ls-remote which has results
			wantErr:        false,
		},
		{
			name:     "installed version does not exist - empty response",
			toolName: "node",
			version:  "installed",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLsInstalledCmd("node"), "")
				m.setResponse(miseLsRemoteCmd("node", ""), "")
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
		{
			name:     "ls installed error",
			toolName: "ruby",
			version:  "installed",
			setupFake: func(m *fakeExecEnv) {
				m.setError(miseLsInstalledCmd("ruby"), fmt.Errorf("command failed"))
			},
			expectedExists: false,
			wantErr:        true,
		},
		{
			name:     "installed with malformed JSON",
			toolName: "ruby",
			version:  "installed",
			setupFake: func(m *fakeExecEnv) {
				m.setResponse(miseLsInstalledCmd("ruby"), `{"invalid": "json"}`)
			},
			expectedExists: false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeExecEnv()
			tt.setupFake(fake)

			exists, err := versionExists(fake, tt.toolName, tt.version)

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
