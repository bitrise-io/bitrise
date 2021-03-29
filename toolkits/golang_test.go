package toolkits

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_stepBinaryFilename(t *testing.T) {
	{
		sIDData := models.StepIDData{SteplibSource: "path", IDorURI: "./", Version: ""}
		require.Equal(t, "path-._-", stepBinaryFilename(sIDData))
	}

	{
		sIDData := models.StepIDData{SteplibSource: "git", IDorURI: "https://github.com/bitrise-steplib/steps-go-toolkit-hello-world.git", Version: "master"}
		require.Equal(t, "git-https___github.com_bitrise-steplib_steps-go-toolkit-hello-world.git-master", stepBinaryFilename(sIDData))
	}

	{
		sIDData := models.StepIDData{SteplibSource: "_", IDorURI: "https://github.com/bitrise-steplib/steps-go-toolkit-hello-world.git", Version: "master"}
		require.Equal(t, "_-https___github.com_bitrise-steplib_steps-go-toolkit-hello-world.git-master", stepBinaryFilename(sIDData))
	}

	{
		sIDData := models.StepIDData{SteplibSource: "https://github.com/bitrise-io/bitrise-steplib.git", IDorURI: "script", Version: "1.2.3"}
		require.Equal(t, "https___github.com_bitrise-io_bitrise-steplib.git-script-1.2.3", stepBinaryFilename(sIDData))
	}
}

func Test_parseGoVersionFromGoVersionOutput(t *testing.T) {
	t.Log("Example OK")
	{
		verStr, err := parseGoVersionFromGoVersionOutput("go version go1.7 darwin/amd64")
		require.NoError(t, err)
		require.Equal(t, "1.7", verStr)
	}

	t.Log("Example OK 2")
	{
		verStr, err := parseGoVersionFromGoVersionOutput(`go version go1.7 darwin/amd64

`)
		require.NoError(t, err)
		require.Equal(t, "1.7", verStr)
	}

	t.Log("Example OK 3")
	{
		verStr, err := parseGoVersionFromGoVersionOutput("go version go1.7.1 darwin/amd64")
		require.NoError(t, err)
		require.Equal(t, "1.7.1", verStr)
	}

	t.Log("Empty")
	{
		verStr, err := parseGoVersionFromGoVersionOutput("")
		require.EqualError(t, err, "Failed to parse Go version, error: version call output was empty")
		require.Equal(t, "", verStr)
	}

	t.Log("Empty 2")
	{
		verStr, err := parseGoVersionFromGoVersionOutput(`

`)
		require.EqualError(t, err, "Failed to parse Go version, error: version call output was empty")
		require.Equal(t, "", verStr)
	}

	t.Log("Invalid")
	{
		verStr, err := parseGoVersionFromGoVersionOutput("go version REMOVED darwin/amd64")
		require.EqualError(t, err, "Failed to parse Go version, error: failed to find version in input: go version REMOVED darwin/amd64")
		require.Equal(t, "", verStr)
	}
}

type mockRunner struct {
	outputs map[string]string
	cmds    []string
}

func (m *mockRunner) run(cmd *command.Model) (string, error) {
	m.cmds = append(m.cmds, cmd.PrintableCommandArgs())
	if val, ok := m.outputs[cmd.PrintableCommandArgs()]; ok {
		return val, nil
	}

	return "", nil
}

func Test_goBuildStep(t *testing.T) {
	goModStep, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	if err := ioutil.WriteFile(filepath.Join(goModStep, "go.mod"), []byte{}, 0600); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	goPathStep, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	type args struct {
		packageName    string
		stepAbsDirPath string
		outputBinPath  string
	}
	tests := []struct {
		name        string
		args        args
		mockOutputs map[string]string
		wantErr     bool
		wantCmds    []string
	}{
		{
			name: "Go module step -> Run in Go module mode",
			args: args{
				packageName:    "github.com/bitrise-steplib/my-step",
				stepAbsDirPath: goModStep,
				outputBinPath:  "/output",
			},
			wantCmds: []string{
				`go "build" "-o" "/output"`,
			},
		},
		{
			name: "GOPATH step, GO111MODULES=on -> should migrate",
			args: args{
				packageName:    "github.com/bitrise-steplib/my-step",
				stepAbsDirPath: goPathStep,
				outputBinPath:  "/output",
			},
			mockOutputs: map[string]string{
				`go "env" "-json" "GO111MODULE"`: `{"GO111MODULE": "on"}`,
			},
			wantCmds: []string{
				`go "env" "-json" "GO111MODULE"`,
				`go "build" "-mod=vendor" "-o" "/output"`,
			},
		},
		{
			name: "GOPATH step, GO111MODULES='' -> should migrate",
			args: args{
				packageName:    "github.com/bitrise-steplib/my-step",
				stepAbsDirPath: goPathStep,
				outputBinPath:  "/output",
			},
			mockOutputs: map[string]string{
				`go "env" "-json" "GO111MODULE"`: `{"GO111MODULE": ""}`,
			},
			wantCmds: []string{
				`go "env" "-json" "GO111MODULE"`,
				`go "build" "-mod=vendor" "-o" "/output"`,
			},
		},
		{
			name: "GOPATH step, GO111MODULES=auto -> Run in GOPATH mode",
			args: args{
				packageName:    "github.com/bitrise-steplib/my-step",
				stepAbsDirPath: goPathStep,
				outputBinPath:  "/output",
			},
			mockOutputs: map[string]string{
				`go "env" "-json" "GO111MODULE"`: `{"GO111MODULE": "auto"}`,
			},
			wantCmds: []string{
				`go "env" "-json" "GO111MODULE"`,
				`go "build" "-o" "/output" "github.com/bitrise-steplib/my-step"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := mockRunner{outputs: tt.mockOutputs}
			goConfig := GoConfigurationModel{
				GoBinaryPath: "go",
				GOROOT:       "/goroot",
			}

			if err := goBuildStep(&mockRunner, goConfig, tt.args.packageName, tt.args.stepAbsDirPath, tt.args.outputBinPath); (err != nil) != tt.wantErr {
				t.Errorf("goBuildStep() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(tt.wantCmds) != len(mockRunner.cmds) {
				t.Fatalf("goBuildStep() wantCmds = %v gotCmds = %v", tt.wantCmds, mockRunner.cmds)
			}
			for i, cmd := range tt.wantCmds {
				if cmd != mockRunner.cmds[i] {
					t.Errorf("goBuildStep() wantCmds = %v gotCmds = %v", tt.wantCmds, mockRunner.cmds)
				}
			}
		})
	}
}
