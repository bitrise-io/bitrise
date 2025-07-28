package execenv

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"al.essio.dev/pkg/shellescape"
)

// ExecEnv contains everything needed to run asdf commands in a specific environment
// that is installed and pre-configured.
type ExecEnv struct {
	// Env vars that confiure asdf and are required for its operation.
	EnvVars map[string]string

	// When set to true, env vars inherited from the parent process are cleared for maximum isolation.
	ClearInheritedEnvs bool

	// ShellInit is a shell command that initializes asdf in the shell session.
	// This is required because classic asdf is written in bash and we can't assume that
	// its init command is sourced in .bashrc or similar (and we don't want to modify
	// anything system-wide).
	ShellInit string
}

func (e *ExecEnv) RunAsdf(args ...string) (string, error) {
	cmdWithArgs := append([]string{"asdf"}, args...)
	return e.RunCommand(nil, cmdWithArgs...)
}

func (e *ExecEnv) RunAsdfPlugin(args ...string) (string, error) {
	cmdWithArgs := append([]string{"asdf", "plugin"}, args...)
	return e.RunCommand(nil, cmdWithArgs...)
}

func (e *ExecEnv) RunCommand(extraEnvs map[string]string, args ...string) (string, error) {
	innerShellCmd := []string{}
	if e.ShellInit != "" {
		innerShellCmd = append(innerShellCmd, e.ShellInit+" &&")
	}
	innerShellCmd = append(innerShellCmd, shellescape.QuoteCommand(args))

	// We need to spawn a sub-shell because classic asdf is implemented in bash and
	// relies on shell features.
	bashArgs := []string{"-c", strings.Join(innerShellCmd, " ")}
	bashCmd := exec.Command("bash", bashArgs...)
	if !e.ClearInheritedEnvs {
		bashCmd.Env = os.Environ()
	}
	for k, v := range e.EnvVars {
		bashCmd.Env = append(bashCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range extraEnvs {
		bashCmd.Env = append(bashCmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	output, err := bashCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %v: %w\n\nOutput:\n%s", "bash", bashArgs, err, output)
	}

	return string(output), nil
}
