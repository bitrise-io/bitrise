package stepman

import (
	"os"
	"os/exec"
)

func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunCommandInDir(workingDir, commandName string, args ...string) error {
	cmd := exec.Command(commandName, args...)
	cmd.Dir = workingDir
	return cmd.Run()
}
