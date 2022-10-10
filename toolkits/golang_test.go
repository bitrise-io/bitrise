package toolkits

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_stepBinaryFilename(t *testing.T) {
	{
		sIDData := models.StepIDData{SteplibSource: "path", IDorURI: "./", Version: ""}
		require.Equal(t, "path-._", stepBinaryFilename(sIDData))
	}

	{
		sIDData := models.StepIDData{SteplibSource: "git", IDorURI: "https://github.com/bitrise-steplib/steps-go-toolkit-hello-world.git", Version: "master"}
		require.Equal(t, "git-https___github.com_bitrise-steplib_steps-go-toolkit-hello-world.git-master", stepBinaryFilename(sIDData))
	}

	{
		sIDData := models.StepIDData{SteplibSource: "git", IDorURI: "https://github.com/bitrise-steplib/steps-go-toolkit-hello-world.git", Version: ""}
		require.Equal(t, "git-https___github.com_bitrise-steplib_steps-go-toolkit-hello-world.git", stepBinaryFilename(sIDData))
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

	t.Log("Beta OK")
	{
		verStr, err := parseGoVersionFromGoVersionOutput("go version go1.17beta1 darwin/amd64")
		require.NoError(t, err)
		require.Equal(t, "1.17", verStr)
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

func (m *mockRunner) run(cmd *command.Model) error {
	m.cmds = append(m.cmds, cmd.PrintableCommandArgs())

	return nil
}

func (m *mockRunner) runForOutput(cmd *command.Model) (string, error) {
	m.cmds = append(m.cmds, cmd.PrintableCommandArgs())
	if val, ok := m.outputs[cmd.PrintableCommandArgs()]; ok {
		return val, nil
	}

	return "", nil
}

func Test_goBuildStep(t *testing.T) {
	type args struct {
		packageName   string
		outputBinPath string
	}
	tests := []struct {
		name        string
		isGoModStep bool
		args        args
		mockOutputs map[string]string
		wantCmds    []string
		wantGoMod   bool
	}{
		{
			name:        "Go module step -> Run in Go module mode",
			isGoModStep: true,
			args: args{
				packageName:   "github.com/bitrise-steplib/my-step",
				outputBinPath: "/output",
			},
			wantCmds: []string{
				`go "build" "-o" "/output"`,
			},
		},
		{
			name: "GOPATH step, GO111MODULES=on -> should migrate",
			args: args{
				packageName:   "github.com/bitrise-steplib/my-step",
				outputBinPath: "/output",
			},
			mockOutputs: map[string]string{
				`go "env" "-json" "GO111MODULE"`: `{"GO111MODULE": "on"}`,
			},
			wantCmds: []string{
				`go "env" "-json" "GO111MODULE"`,
				`go "build" "-mod=vendor" "-o" "/output"`,
			},
			wantGoMod: true,
		},
		{
			name: "GOPATH step, GO111MODULES='' -> should migrate",
			args: args{
				packageName:   "github.com/bitrise-steplib/my-step",
				outputBinPath: "/output",
			},
			mockOutputs: map[string]string{
				`go "env" "-json" "GO111MODULE"`: `{"GO111MODULE": ""}`,
			},
			wantCmds: []string{
				`go "env" "-json" "GO111MODULE"`,
				`go "build" "-mod=vendor" "-o" "/output"`,
			},
			wantGoMod: true,
		},
		{
			name: "GOPATH step, GO111MODULES=auto -> Run in GOPATH mode",
			args: args{
				packageName:   "github.com/bitrise-steplib/my-step",
				outputBinPath: "/output",
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
			stepDir, err := ioutil.TempDir("", "")
			require.NoError(t, err, "failed to create temp dir")

			goModPath := filepath.Join(stepDir, "go.mod")
			if tt.isGoModStep {
				err := ioutil.WriteFile(goModPath, []byte{}, 0600)
				require.NoError(t, err, "failed to create file")
			}

			mockRunner := mockRunner{outputs: tt.mockOutputs}
			goConfig := GoConfigurationModel{
				GoBinaryPath: "go",
				GOROOT:       "/goroot",
			}

			err = goBuildStep(&mockRunner, goConfig, tt.args.packageName, stepDir, tt.args.outputBinPath)

			require.NoError(t, err, "goBuildStep()")
			require.Equal(t, tt.wantCmds, mockRunner.cmds, "goBuildStep() run commands do not match")

			if tt.wantGoMod {
				_, err := os.Stat(goModPath)
				require.NoError(t, err, "go.mod was not created")
			}
		})
	}
}

func Benchmark_goBuildStep(b *testing.B) {
	isInstallRequired, _, goConfig, err := selectGoConfiguration()
	require.NoError(b, err, "Failed to select an appropriate Go installation for compiling the Step")

	if isInstallRequired {
		b.Fatalf("Failed to select an appropriate Go installation for compiling the Step")
	}

	stepDir, err := ioutil.TempDir("", "")
	require.NoError(b, err)

	defer func() {
		err := os.RemoveAll(stepDir)
		require.NoError(b, err)
	}()

	outputDir, err := ioutil.TempDir("", "")
	require.NoError(b, err)

	defer func() {
		err := os.RemoveAll(outputDir)
		require.NoError(b, err)
	}()

	stepmanCmd := command.New("stepman", "activate",
		"--collection", "https://github.com/bitrise-io/bitrise-steplib",
		"--id", "xcode-test",
		"--version", "2.5.0", "--update",
		"--path", stepDir)
	err = stepmanCmd.Run()
	require.NoError(b, err)

	packageName := "github.com/bitrise-steplib/steps-xcode-test"

	for _, mode := range []string{"on", "auto"} {
		b.Run(fmt.Sprintf("Benchmarking GO111MODULE=%s", mode), func(b *testing.B) {
			err := os.Setenv("GO111MODULE", mode)
			require.NoError(b, err)

			for i := 0; i < b.N; i++ {
				stepPerTestDir, err := ioutil.TempDir("", "")
				require.NoError(b, err)

				defer func() {
					err := os.RemoveAll(stepPerTestDir)
					require.NoError(b, err)
				}()

				err = command.CopyDir(stepDir, stepPerTestDir, true)
				require.NoError(b, err)

				err = goBuildStep(&defaultRunner{},
					goConfig,
					packageName,
					stepPerTestDir,
					filepath.Join(outputDir, fmt.Sprintf("%s_%d", mode, i)))
				require.NoError(b, err)
			}
		})
	}
}
