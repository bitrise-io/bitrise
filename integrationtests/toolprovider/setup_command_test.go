//go:build linux_and_mac
// +build linux_and_mac

package toolprovider

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type setupCommandCase struct {
	name           string
	fileName       string
	fileContent    string
	outputFormat   string
	workflow       string // Workflow flag (optional, for bitrise.yml)
	validateOutput func(t *testing.T, output string)
	useAsdf        bool
	wantErr        bool
	errContains    string // substring expected in output when wantErr is true
}

func TestToolsSetupCommand(t *testing.T) {
	cases := []setupCommandCase{
		{
			name:         "setup from .tool-versions file",
			fileContent:  "golang 1.21.0",
			fileName:     ".tool-versions",
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "golang")
				assert.Contains(t, output, "1.21.0")
			},
		},
		{
			name: "setup from .tool-versions with multiple tools",
			fileContent: `nodejs 20.0.0
python 3.11.0`,
			fileName:     ".tool-versions",
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "nodejs")
				assert.Contains(t, output, "20.0.0")
				assert.Contains(t, output, "python")
				assert.Contains(t, output, "3.11.0")
			},
		},
		{
			name: "setup from bitrise.yml with global tools",
			fileContent: `format_version: "17"
tools:
  nodejs: 20.0.0
workflows:
  test:
    steps:
      - script:
          inputs:
            - content: echo "test"`,
			fileName:     "bitrise.yml",
			workflow:     "test",
			outputFormat: "plaintext",
		},
		{
			name: "setup from bitrise.yml with workflow-specific tools",
			fileContent: `format_version: "17"
tools:
  golang: 1.21.0
workflows:
  test:
    tools:
      nodejs: 20.0.0
      python: 3.11.0
    steps:
      - script:
          inputs:
            - content: echo "test"`,
			fileName:     "bitrise.yml",
			workflow:     "test",
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "nodejs")
				assert.Contains(t, output, "20.0.0")
				assert.Contains(t, output, "python")
				assert.Contains(t, output, "3.11.0")
				// Global tool should also be included
				assert.Contains(t, output, "golang")
				assert.Contains(t, output, "1.21.0")

			},
		},
		{
			name:         "JSON output",
			fileContent:  "golang 1.21.0",
			fileName:     ".tool-versions",
			outputFormat: "json",
			validateOutput: func(t *testing.T, output string) {
				var v any
				err := json.Unmarshal([]byte(output), &v)
				assert.NoError(t, err, "Output should be valid JSON")
			},
		},
		{
			name:         "output format bash",
			fileContent:  "golang 1.21.0",
			fileName:     ".tool-versions",
			outputFormat: "bash",
			validateOutput: func(t *testing.T, output string) {
				cmd := exec.Command("bash", "-c", fmt.Sprintf("eval %s", output))
				out, err := cmd.CombinedOutput()
				assert.NoError(t, err, "Should be able to eval bash output without error: %s", string(out))
			},
		},
		{
			name:         "use asdf provider",
			useAsdf:      true,
			fileContent:  "golang 1.21.0",
			fileName:     ".tool-versions",
			outputFormat: "bash",
			validateOutput: func(t *testing.T, output string) {
				cmd := exec.Command("bash", "-c", fmt.Sprintf("eval %s", output))
				out, err := cmd.CombinedOutput()
				assert.NoError(t, err, "Should be able to eval bash output without error: %s", string(out))
			},
		},
		// .fvmrc tests
		{
			name:         "setup from .fvmrc with exact version",
			fileContent:  `{"flutter": "3.32.1"}`,
			fileName:     ".fvmrc",
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "flutter")
				assert.Contains(t, output, "3.32.1")
			},
		},
		{
			name:         "setup from .fvmrc with version and channel",
			fileContent:  `{"flutter": "3.32.1@stable"}`,
			fileName:     ".fvmrc",
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "flutter")
				assert.Contains(t, output, "3.32.1")
			},
		},
		{
			name:         "setup from .fvmrc with channel only fails",
			fileContent:  `{"flutter": "stable"}`,
			fileName:     ".fvmrc",
			outputFormat: "plaintext",
			wantErr:      true,
			errContains:  "channel-only value",
		},
		{
			name:         "setup from .fvmrc with flavors",
			fileContent:  `{"flutter": "3.32.1", "flavors": {"staging": "3.29.0"}}`,
			fileName:     ".fvmrc",
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "flutter")
				assert.Contains(t, output, "3.32.1")
				assert.Contains(t, output, "3.29.0")
			},
		},
		{
			name:         "setup from .fvmrc with same flavor version",
			fileContent:  `{"flutter": "3.32.1", "flavors": {"staging": "3.32.1"}}`,
			fileName:     ".fvmrc",
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "flutter")
				assert.Contains(t, output, "3.32.1")
			},
		},
		{
			name:         "setup from .fvmrc with flavor channel only fails",
			fileContent:  `{"flutter": "3.32.1", "flavors": {"staging": "beta"}}`,
			fileName:     ".fvmrc",
			outputFormat: "plaintext",
			wantErr:      true,
			errContains:  "channel-only value",
		},
		// fvm_config.json tests
		{
			name:         "setup from fvm_config.json with exact version",
			fileContent:  `{"flutterSdkVersion": "3.32.1"}`,
			fileName:     "fvm_config.json",
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "flutter")
				assert.Contains(t, output, "3.32.1")
			},
		},
		{
			name:         "setup from .fvm/fvm_config.json subdirectory",
			fileContent:  `{"flutterSdkVersion": "3.32.1"}`,
			fileName:     filepath.Join(".fvm", "fvm_config.json"),
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "flutter")
				assert.Contains(t, output, "3.32.1")
			},
		},
		{
			name:         "setup from fvm_config.json with version and channel",
			fileContent:  `{"flutterSdkVersion": "3.32.1@stable"}`,
			fileName:     "fvm_config.json",
			outputFormat: "plaintext",
			validateOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "flutter")
				assert.Contains(t, output, "3.32.1")
			},
		},
		{
			name:         "setup from fvm_config.json with channel only fails",
			fileContent:  `{"flutterSdkVersion": "stable"}`,
			fileName:     "fvm_config.json",
			outputFormat: "plaintext",
			wantErr:      true,
			errContains:  "channel-only value",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, tc.fileName)

			err := os.MkdirAll(filepath.Dir(configPath), 0755)
			require.NoError(t, err)
			err = os.WriteFile(configPath, []byte(tc.fileContent), 0644)
			require.NoError(t, err)

			args := []string{"tools", "setup", "--config", configPath, "--format", tc.outputFormat}
			if tc.workflow != "" {
				args = append(args, "--workflow", tc.workflow)
			}

			if tc.useAsdf {
				args = append(args, "--provider", "asdf")
			}

			cmd := command.New(testhelpers.BinPath(), args...)
			cmd.SetDir(tmpDir)
			out, err := cmd.RunAndReturnTrimmedCombinedOutput()

			if tc.wantErr {
				require.Error(t, err, "expected command to fail")
				if tc.errContains != "" {
					assert.Contains(t, out, tc.errContains)
				}
				return
			}

			if err != nil {
				t.Logf("Setup output: %s", out)
				t.Logf("Setup error (may be expected): %v", err)
			}
			if tc.validateOutput != nil {
				tc.validateOutput(t, out)
			}
		})
	}
}

func TestToolsSetupCommand_MultipleConfigs(t *testing.T) {
	tmpDir := t.TempDir()

	bitriseYml1 := filepath.Join(tmpDir, "bitrise1.yml")
	bitriseYml2 := filepath.Join(tmpDir, "bitrise2.yml")

	content := `format_version: "17"
workflows:
  test:
    steps:
      - script:
          inputs:
            - content: echo "test"`

	err := os.WriteFile(bitriseYml1, []byte(content), 0644)
	require.NoError(t, err)
	err = os.WriteFile(bitriseYml2, []byte(content), 0644)
	require.NoError(t, err)

	cmd := command.New(testhelpers.BinPath(), "tools", "setup",
		"--config", bitriseYml1,
		"--config", bitriseYml2)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()

	require.Error(t, err)
	assert.Contains(t, out, "multiple bitrise config files")
}

func TestToolsSetupCommand_EnvmanInitialization(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".tool-versions")
	err := os.WriteFile(configPath, []byte("go 1.21.13\n"), 0644)
	require.NoError(t, err)

	cmd := command.New(testhelpers.BinPath(), "tools", "setup",
		"--config", configPath)
	cmd.SetDir(tmpDir)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, "tools setup should succeed: %s", out)

	// Output should NOT contain the envman error
	assert.NotContains(t, out, "Failed to expose tool envs with envman",
		"Should not fail to expose envs with envman (envman should be properly initialized)")
	assert.NotContains(t, out, "No file found at path",
		"Should not fail with envman file not found error")
}

func TestToolsSetupCommand_GlobalToolsWithoutWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "bitrise.yml")

	// bitrise.yml with global tools but no workflow specified
	content := `format_version: "17"
tools:
  nodejs: 20.0.0
  python: 3.11.0
workflows:
  test:
    steps:
      - script:
          inputs:
            - content: echo "test"`

	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	// Run setup WITHOUT --workflow flag
	cmd := command.New(testhelpers.BinPath(), "tools", "setup",
		"--config", configPath,
		"--format", "plaintext")
	cmd.SetDir(tmpDir)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()

	require.NoError(t, err, "tools setup should succeed without workflow flag: %s", out)

	// Global tools should be installed
	assert.Contains(t, out, "nodejs")
	assert.Contains(t, out, "20.0.0")
	assert.Contains(t, out, "python")
	assert.Contains(t, out, "3.11.0")
}
