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
				var v interface{}
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var configPath string
			var tmpDir string

			tmpDir = t.TempDir()
			configPath = filepath.Join(tmpDir, tc.fileName)
			err := os.WriteFile(configPath, []byte(tc.fileContent), 0644)
			require.NoError(t, err)

			args := []string{"tools", "setup", "--config", configPath, "--format", tc.outputFormat}
			if tc.workflow != "" {
				args = append(args, "--workflow", tc.workflow)
			}

			cmd := command.New(testhelpers.BinPath(), args...)
			if tmpDir != "" {
				cmd.SetDir(tmpDir)
			}
			out, err := cmd.RunAndReturnTrimmedCombinedOutput()

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
