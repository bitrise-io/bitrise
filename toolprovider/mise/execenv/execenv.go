package execenv

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bitrise-io/bitrise/v3/log"
)

const (
	InstallTimeout = 5 * time.Minute
	DefaultTimeout = 1 * time.Minute

	textFileBusyAttempts  = 5
	textFileBusyRetryWait = 200 * time.Millisecond
)

// ExecEnv contains everything needed to run mise commands in a specific environment
// that is installed and pre-configured.
type ExecEnv interface {
	// InstallDir is the directory where mise is installed. This is not necessarily the same as the data directory.
	InstallDir() string
	RunMise(args ...string) (string, error)
	RunMiseWithTimeout(timeout time.Duration, args ...string) (string, error)
	RunMiseWithTimeoutAndEnvs(timeout time.Duration, extraEnvs map[string]string, args ...string) (string, error)
	RunMisePlugin(args ...string) (string, error)
}

type MiseExecEnv struct {
	installDir string
	ExtraEnvs  map[string]string
}

func NewMiseExecEnv(installDir string, extraEnvs map[string]string) MiseExecEnv {
	return MiseExecEnv{
		installDir: installDir,
		ExtraEnvs:  extraEnvs,
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
	return e.RunMiseWithTimeoutAndEnvs(timeout, nil, args...)
}

// RunMiseWithTimeoutAndEnvs runs mise commands with optional timeout and additional environment variables.
func (e MiseExecEnv) RunMiseWithTimeoutAndEnvs(timeout time.Duration, additionalEnvs map[string]string, args ...string) (string, error) {
	var ctx context.Context
	if timeout == 0 {
		ctx = context.Background()
	} else {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	executable := filepath.Join(e.installDir, "bin", "mise")
	env := os.Environ()
	for k, v := range e.ExtraEnvs {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range additionalEnvs {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	var output []byte
	var err error
	for attempt := 1; ; attempt++ {
		cmd := exec.CommandContext(ctx, executable, args...)
		cmd.Env = env
		output, err = cmd.CombinedOutput()
		// Bootstrap writes bin/mise, and a process another goroutine forks meanwhile
		// inherits that write fd until it execs, which fails running mise with ETXTBSY
		// (https://go.dev/issue/22315). The child execs soon, so the run succeeds on retry.
		if !errors.Is(err, syscall.ETXTBSY) || attempt == textFileBusyAttempts {
			break
		}
		log.Warnf("Running %s failed with text file busy, retrying (attempt %d of %d).", executable, attempt+1, textFileBusyAttempts)
		time.Sleep(textFileBusyRetryWait)
	}
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
