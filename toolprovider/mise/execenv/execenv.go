package execenv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	InstallTimeout = 5 * time.Minute
	DefaultTimeout = 1 * time.Minute
)

// ExecEnv contains everything needed to run mise commands in a specific environment
// that is installed and pre-configured.
type ExecEnv interface {
	// InstallDir is the directory where mise is installed. This is not necessarily the same as the data directory.
	InstallDir() string
	RunMise(args ...string) (string, error)
	RunMiseWithTimeout(timeout time.Duration, args ...string) (string, error)
	RunMisePlugin(args ...string) (string, error)
}

type MiseExecEnv struct {
	installDir string

	extraEnvs map[string]string
}

// extraEnvs: additional env vars that configure mise and are required for its operation.
func NewMiseExecEnv(installDir string, extraEnvs map[string]string) MiseExecEnv {
	return MiseExecEnv{
		installDir: installDir,
		extraEnvs:  extraEnvs,
	}
}

func (e MiseExecEnv) RunMise(args ...string) (string, error) {
	return e.RunMiseWithTimeout(0, args...)
}

func (e MiseExecEnv) RunMisePlugin(args ...string) (string, error) {
	cmdWithArgs := append([]string{"plugin"}, args...)

	// Use timeout for all plugin operations as they involve unknown code execution.
	return e.RunMiseWithTimeout(DefaultTimeout, cmdWithArgs...)
}

// RunMiseWithTimeout runs mise commands that involve untrusted operations (plugin execution, remote network calls)
// with a timeout to prevent hanging.
func (e MiseExecEnv) RunMiseWithTimeout(timeout time.Duration, args ...string) (string, error) {
	var ctx context.Context
	if timeout == 0 {
		ctx = context.Background()
	} else {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	executable := filepath.Join(e.installDir, "bin", "mise")
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = os.Environ()
	for k, v := range e.extraEnvs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("mise command timed out: %v", args)
		}
		return "", fmt.Errorf("%s\n%s", err, output)
	}

	return string(output), nil
}

func (e MiseExecEnv) InstallDir() string {
	return e.installDir
}
