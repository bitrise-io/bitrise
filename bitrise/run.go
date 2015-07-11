package bitrise

import (
	"os"
	"os/exec"
	"strings"
)

// ------------------
// --- Stepman

// RunStepmanSetup ...
func RunStepmanSetup(collection string) error {
	args := []string{"-d", "true", "-c", collection, "setup"}
	return RunCommand("stepman", args...)
}

// RunStepmanActivate ...
func RunStepmanActivate(collection, stepID, stepVersion, dir string) error {
	args := []string{"-d", "true", "-c", collection, "activate", "-i", stepID, "-v", stepVersion, "-p", dir}
	return RunCommand("stepman", args...)
}

// ------------------
// --- Envman

// RunEnvmanInit ...
func RunEnvmanInit() error {
	return RunCommand("envman", "init")
}

// RunEnvmanAdd ...
func RunEnvmanAdd(key, value string) error {
	args := []string{"add", "-k", key}
	envman := exec.Command("envman", args...)
	envman.Stdin = strings.NewReader(value)
	envman.Stdout = os.Stdout
	envman.Stderr = os.Stderr
	return envman.Run()
}

// RunEnvmanRun ...
func RunEnvmanRun(cmd []string) error {
	args := []string{"run"}
	args = append(args, cmd...)
	return RunCommand("envman", args...)
}

// RunEnvmanRunInDir ...
func RunEnvmanRunInDir(dir string, cmd []string) error {
	args := []string{"run"}
	args = append(args, cmd...)
	return RunCommandInDir(dir, "envman", args...)
}

// ------------------
// --- Common

// RunCommand ...
func RunCommand(name string, args ...string) error {
	return RunCommandInDir("", name, args...)
}

// RunCommandInDir ...
func RunCommandInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if dir != "" {
		cmd.Dir = dir
	}
	return cmd.Run()
}
