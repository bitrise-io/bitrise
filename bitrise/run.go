package bitrise

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
)

// ------------------
// --- Stepman

// StepmanSetup ...
func StepmanSetup(collection string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "setup", "--collection", collection}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanActivate ...
func StepmanActivate(collection, stepID, stepVersion, dir, ymlPth string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "activate", "--collection", collection,
		"--id", stepID, "--version", stepVersion, "--path", dir, "--copyyml", ymlPth}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanUpdate ...
func StepmanUpdate(collection string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "update", "--collection", collection}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanPrintRawStepLibStepInfo ...
func StepmanPrintRawStepLibStepInfo(collection, stepID, stepVersion string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--collection", collection,
		"--id", stepID, "--version", stepVersion, "--format", "raw"}
	out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
	if err != nil {
		return err
	}

	fmt.Println(out)
	return nil
}

// StepmanPrintRawLocalStepInfo ...
func StepmanPrintRawLocalStepInfo(pth string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--step-yml", pth, "--format", "raw"}
	out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
	if err != nil {
		return err
	}

	fmt.Println(out)
	return nil
}

// StepmanJSONStepLibStepInfo ...
func StepmanJSONStepLibStepInfo(collection, stepID, stepVersion string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--collection", collection,
		"--id", stepID, "--version", stepVersion, "--format", "json"}

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	if err := cmdex.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), "stepman", args...); err != nil {
		return outBuffer.String(), fmt.Errorf("Error: %s, details: %s", err, errBuffer.String())
	}

	return outBuffer.String(), nil
}

// StepmanJSONLocalStepInfo ...
func StepmanJSONLocalStepInfo(pth string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--step-yml", pth, "--format", "json"}

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	if err := cmdex.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), "stepman", args...); err != nil {
		return outBuffer.String(), fmt.Errorf("Error: %s, details: %s", err, errBuffer.String())
	}

	return outBuffer.String(), nil
}

// StepmanPrintRawStepList ...
func StepmanPrintRawStepList(collection string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-list", "--collection", collection, "--format", "raw"}
	out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
	if err != nil {
		return err
	}

	fmt.Println(out)
	return nil
}

// StepmanJSONStepList ...
func StepmanJSONStepList(collection string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-list", "--collection", collection, "--format", "json"}

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	if err := cmdex.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), "stepman", args...); err != nil {
		return outBuffer.String(), fmt.Errorf("Error: %s, details: %s", err, errBuffer.String())
	}

	return outBuffer.String(), nil
}

// ------------------
// --- Envman

// EnvmanInit ...
func EnvmanInit() error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "init"}
	return cmdex.RunCommand("envman", args...)
}

// EnvmanInitAtPath ...
func EnvmanInitAtPath(envstorePth string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "init", "--clear"}
	return cmdex.RunCommand("envman", args...)
}

// EnvmanAdd ...
func EnvmanAdd(envstorePth, key, value string, expand bool) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "add", "--key", key, "--append"}
	if !expand {
		args = []string{"--loglevel", logLevel, "--path", envstorePth, "add", "--key", key, "--no-expand", "--append"}
	}

	envman := exec.Command("envman", args...)
	envman.Stdin = strings.NewReader(value)
	envman.Stdout = os.Stdout
	envman.Stderr = os.Stderr
	return envman.Run()
}

// EnvmanRun ...
func EnvmanRun(envstorePth, workDirPth string, cmd []string) (int, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "run"}
	args = append(args, cmd...)

	return cmdex.RunCommandInDirAndReturnExitCode(workDirPth, "envman", args...)
}

// EnvmanJSONPrint ...
func EnvmanJSONPrint(envstorePth string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "print", "--format", "json", "--expand"}

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	if err := cmdex.RunCommandWithWriters(io.Writer(&outBuffer), io.Writer(&errBuffer), "envman", args...); err != nil {
		return outBuffer.String(), fmt.Errorf("Error: %s, details: %s", err, errBuffer.String())
	}

	return outBuffer.String(), nil
}
