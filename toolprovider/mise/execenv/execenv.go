package execenv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"
)

const (
	InstallTimeout = 5 * time.Minute
	DefaultTimeout = 1 * time.Minute
)

// ExecEnv contains everything needed to run mise commands in a specific environment
// that is installed and pre-configured.
type ExecEnv struct {
	// InstallDir is the directory where mise is installed. This is not necessarily the same as the data directory.
	InstallDir string

	// Additional env vars that configure mise and are required for its operation.
	ExtraEnvs map[string]string
}

func (e *ExecEnv) RunMise(args ...string) (string, error) {
	return e.RunMiseWithTimeout(0, args...)
}

func (e *ExecEnv) RunMisePlugin(args ...string) (string, error) {
	cmdWithArgs := append([]string{"plugin"}, args...)

	// Use timeout for all plugin operations as they involve unknown code execution
	return e.RunMiseWithTimeout(DefaultTimeout, cmdWithArgs...)
}

// RunMiseWithTimeout runs mise commands that involve untrusted operations (plugin execution, remote network calls)
// with a timeout to prevent hanging
func (e *ExecEnv) RunMiseWithTimeout(timeout time.Duration, args ...string) (string, error) {
	var ctx context.Context
	if timeout == 0 {
		ctx = context.Background()
	} else {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	executable := path.Join(e.InstallDir, "bin", "mise")
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = os.Environ()
	for k, v := range e.ExtraEnvs {
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
