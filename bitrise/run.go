package bitrise

import (
	"os"
	"os/exec"
	"strings"
)

// ------------------
// --- Stepman

// RunStepmanSetup ...
func RunStepmanSetup() error {
	return RunCommand("stepman", "setup")
}

// RunStepmanActivate ...
func RunStepmanActivate(stepID, stepVersion, dir string) error {
	args := []string{"activate", "-i", stepID, "-v", stepVersion, "-p", dir}
	return RunCommand("stepman", args...)
}

// ------------------
// --- Envman

// RunEnvmanInit ...
func RunEnvmanInit() error {
	return RunCommand("envman", "init")
}

// RunPipedEnvmanAdd ...
func RunPipedEnvmanAdd(key, value string) error {
	args := []string{"add", "-k", key}
	envman := exec.Command("envman", args...)
	envman.Stdin = strings.NewReader(value)
	envman.Stdout = os.Stdout
	envman.Stderr = os.Stderr
	return envman.Run()
}

// RunEnvmanAdd ...
func RunEnvmanAdd(key, value string) error {
	args := []string{"add", "-k", key, "-v", value}
	return RunCommand("envman", args...)
}

// RunEnvmanRun ...
func RunEnvmanRun(cmd []string) error {
	args := []string{"run"}
	args = append(args, cmd...)
	return RunCommand("envman", args...)
}

// ------------------
// --- Common

// RunCommand ...
func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
