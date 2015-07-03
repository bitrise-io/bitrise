package bitrise

import (
	_ "fmt"
	_ "io"
	"os"
	"os/exec"
	"strings"
)

// Stepman
func RunStepmanSetup() error {
	return RunCommand("stepman", "setup")
}

func RunStepmanActivate(stepId, stepVersion, dir string) error {
	args := []string{"activate", "-i", stepId, "-v", stepVersion, "-p", dir}
	return RunCommand("stepman", args...)
}

// Envman
func RunEnvmanInit() error {
	return RunCommand("envman", "init")
}

func RunPipedEnvmanAdd(key, value string) error {
	args := []string{"add", "-k", key}
	envman := exec.Command("envman", args...)
	envman.Stdin = strings.NewReader(value)
	envman.Stdout = os.Stdout
	envman.Stderr = os.Stderr
	return envman.Run()
}

func RunEnvmanAdd(key, value string) error {
	args := []string{"add", "-k", key, "-v", value}
	return RunCommand("envman", args...)
}

func RunEnvmanRun(cmd []string) error {
	args := []string{"run"}
	args = append(args, cmd...)

	return RunCommand("envman", args...)
}

// Common
func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
