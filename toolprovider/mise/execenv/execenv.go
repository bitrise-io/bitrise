package execenv

import (
	"fmt"
	"os"
	"os/exec"
	"path"
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
	executable := path.Join(e.InstallDir, "bin", "mise")
	cmd := exec.Command(executable, args...)
	cmd.Env = os.Environ()
	for k, v := range e.ExtraEnvs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s\n%s", err, output)
	}

	return string(output), nil
}
