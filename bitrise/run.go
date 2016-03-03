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
	version "github.com/hashicorp/go-version"
)

// ------------------
// --- Stepman

// StepmanVersion ...
func StepmanVersion() (version.Version, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "--version"}

	versionOut, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
	if err != nil {
		return version.Version{}, err
	}

	versionPtr, err := version.NewVersion(versionOut)
	if err != nil {
		return version.Version{}, err
	}
	if versionPtr == nil {
		return version.Version{}, fmt.Errorf("Failed to parse version (%s)", versionOut)
	}

	return *versionPtr, nil
}

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

// StepmanRawStepLibStepInfo ...
func StepmanRawStepLibStepInfo(collection, stepID, stepVersion string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--collection", collection,
		"--id", stepID, "--version", stepVersion, "--format", "raw"}
	return cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
}

// StepmanRawLocalStepInfo ...
func StepmanRawLocalStepInfo(pth string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--step-yml", pth, "--format", "raw"}
	return cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
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

// StepmanRawStepList ...
func StepmanRawStepList(collection string) (string, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-list", "--collection", collection, "--format", "raw"}
	return cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
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

//
// Share

// StepmanShare ...
func StepmanShare() error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "--toolmode"}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanShareAudit ...
func StepmanShareAudit() error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "audit", "--toolmode"}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanShareCreate ...
func StepmanShareCreate(tag, git, stepID string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "create", "--tag", tag, "--git", git, "--stepid", stepID, "--toolmode"}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanShareFinish ...
func StepmanShareFinish() error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "finish", "--toolmode"}
	return cmdex.RunCommand("stepman", args...)
}

// StepmanShareStart ...
func StepmanShareStart(collection string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "share", "start", "--collection", collection, "--toolmode"}
	return cmdex.RunCommand("stepman", args...)
}

// ------------------
// --- Envman

// EnvmanVersion ...
func EnvmanVersion() (version.Version, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--version"}
	versionOut, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("envman", args...)
	if err != nil {
		return version.Version{}, err
	}

	versionPtr, err := version.NewVersion(versionOut)
	if err != nil {
		return version.Version{}, err
	}
	if versionPtr == nil {
		return version.Version{}, fmt.Errorf("Failed to parse version (%s)", versionOut)
	}

	return *versionPtr, nil
}

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
