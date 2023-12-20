package toolversions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/stretchr/testify/assert"
)

const validASDFOutput = `alias           ______          No version is set. Run "asdf <global|shell|local> alias <version>"
flutter         3.16.1-stable   /Users/bitrise/.tool-versions
golang          1.18            /Users/bitrise/Projects/steps/.tool-versions
java            17              Not installed. Run "asdf install java 17"
nodejs          19.7.0          Not installed. Run "asdf install nodejs 19.7.0"
ruby            3.1.3           /Users/bitrise/.tool-versions
python		    ______          No version is set. Run "asdf <global|shell|local> python <version>"`

func TestIsAvailable(t *testing.T) {
	tests := []struct {
		name        string
		systemPath  string
		cmdOutput   string
		cmdExitCode int
		expected    bool
	}{
		{
			name:        "asdf is available",
			systemPath:  "/bin:/usr/bin:/root/.asdf/bin/asdf",
			cmdOutput:   validASDFOutput,
			cmdExitCode: 0,
			expected:    true,
		},
		{
			name:        "asdf is not available",
			systemPath:  "/bin:/usr/bin",
			cmdOutput:   "",
			cmdExitCode: 1,
			expected:    false,
		},
		{
			name:        "asdf is not working",
			systemPath:  "/bin:/usr/bin:/root/.asdf/bin/asdf",
			cmdOutput:   "",
			cmdExitCode: 1,
			expected:    false,
		},
		
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewLogger()
			logger.EnableDebugLog(true)
			r := NewASDFVersionReporter(
				fakeCommandLocator{path: tt.systemPath},
				fakeCommandFactory{stdout: tt.cmdOutput, exitCode: tt.cmdExitCode},
				logger,
				"/Users/bitrise",
			)

			result := r.IsAvailable()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCurrentToolVersions(t *testing.T) {
	tests := []struct {
		name      string
		cmdOutput string
		cmdError  error
		expected  map[string]ToolVersion
		expectErr bool
	}{
		{
			name:      "valid output",
			cmdOutput: validASDFOutput,
			expected: map[string]ToolVersion{
				"flutter": {
					Version:        "3.16.1-stable",
					IsInstalled:    true,
					DeclaredByFile: ".tool-versions",
					IsGlobal:       true,
				},
				"golang": {
					Version:        "1.18",
					IsInstalled:    true,
					DeclaredByFile: ".tool-versions",
					IsGlobal:       false,
				},
				"java": {
					Version:        "17",
					IsInstalled:    false,
					DeclaredByFile: "",
					IsGlobal:       false,
				},
				"nodejs": {
					Version:        "19.7.0",
					IsInstalled:    false,
					DeclaredByFile: "",
					IsGlobal:       false,
				},
				"ruby": {
					Version:        "3.1.3",
					IsInstalled:    true,
					DeclaredByFile: ".tool-versions",
					IsGlobal:       true,
				},
			},
			expectErr: false,
		},
		{
			name:      "empty output",
			cmdOutput: "",
			expected:  map[string]ToolVersion{},
			expectErr: false,
		},
		{
			name:      "invalid output",
			cmdOutput: "error",
			expected:  map[string]ToolVersion{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewLogger()
			logger.EnableDebugLog(true)
			r := NewASDFVersionReporter(
				fakeCommandLocator{path: "/root/.asdf/bin/asdf"},
				fakeCommandFactory{stdout: tt.cmdOutput},
				logger,
				"/Users/bitrise",
			)

			result, err := r.CurrentToolVersions()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

type fakeCommandLocator struct {
	path string
}

func (f fakeCommandLocator) LookPath(name string) (string, error) {
	return f.path, nil
}

type fakeCommandFactory struct {
	stdout   string
	exitCode int
}

func (f fakeCommandFactory) Create(name string, args []string, opts *command.Opts) command.Command {
	return fakeCommand{
		command:  fmt.Sprintf("%s %s", name, strings.Join(args, " ")),
		stdout:   f.stdout,
		exitCode: f.exitCode,
	}
}

type fakeCommand struct {
	command  string
	stdout   string
	stderr   string
	exitCode int
}

func (c fakeCommand) PrintableCommandArgs() string {
	return c.command
}

func (c fakeCommand) Run() error {
	if c.exitCode != 0 {
		return fmt.Errorf("exit code %d", c.exitCode)
	}
	return nil
}

func (c fakeCommand) RunAndReturnExitCode() (int, error) {
	if c.exitCode != 0 {
		return c.exitCode, fmt.Errorf("exit code %d", c.exitCode)
	}
	return c.exitCode, nil
}

func (c fakeCommand) RunAndReturnTrimmedOutput() (string, error) {
	if c.exitCode != 0 {
		return "", fmt.Errorf("exit code %d", c.exitCode)
	}
	return c.stdout, nil
}

func (c fakeCommand) RunAndReturnTrimmedCombinedOutput() (string, error) {
	if c.exitCode != 0 {
		return "", fmt.Errorf("exit code %d", c.exitCode)
	}
	return fmt.Sprintf("%s%s", c.stdout, c.stderr), nil
}

func (c fakeCommand) Start() error {
	return nil
}

func (c fakeCommand) Wait() error {
	return nil
}
