package bitrise

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Stepman
func RunStepmanSetup() error {
	return RunCommand("stepman", "setup")
}

func RunStepmanActivate(stepId, stepVersion, dir string) error {
	args := strings.Split(fmt.Sprintf("activate -i %s -v %s -p %s", stepId, stepVersion, dir), " ")
	return RunCommand("stepman", args...)
}

// Envman
func RunEnvmanInit() error {
	return RunCommand("envman", "init")
}

func RunPipedEnvmanAdd(key, value string) error {
	echo := exec.Command("echo", value)

	args := strings.Split(fmt.Sprintf("add -k %s", key), " ")
	envman := exec.Command("envman", args...)

	reader, writer := io.Pipe()

	// push echo command output to writer
	echo.Stdout = writer

	// read from echo command output
	envman.Stdin = reader

	// prepare a buffer to capture the output
	// after envman command finished executing
	envman.Stdout = os.Stdout

	echo.Start()
	envman.Start()
	echo.Wait()
	writer.Close()
	envman.Wait()

	return nil
}

func RunEnvmanAdd(key, value string) error {
	//argsString := fmt.Sprintf("add -k %s -v %s", key, value)
	args := []string{"add", "-k", key, "-v", value}
	//args := strings.Split(fmt.Sprintf("add -k %s -v %s", key, value), " ")
	return RunCommand("envman", args...)
}

func RunEnvmanRun(cmd string) error {
	args := strings.Split(fmt.Sprintf("run %s", cmd), " ")
	return RunCommand("envman", args...)
}

// Common
func RunBashCommand(cmd string) error {
	c := exec.Command("bash", cmd)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func RunBashCommandInDir(dir, cmd string) error {
	c := exec.Command("bash", cmd)
	c.Dir = dir
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func commandEnvs(env map[string]string) []string {
	cmdEnvs := []string{}

	for key, value := range env {
		cmdEnvs = append(cmdEnvs, key+"="+value)
	}

	return append(os.Environ(), cmdEnvs...)
}

func RunCommandWithEnv(name string, env map[string]string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = commandEnvs(env)
	return cmd.Run()
}

func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunCommandInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
