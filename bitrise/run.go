package bitrise

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/cmdex"
)

// StepInfoModel ...
type StepInfoModel struct {
	StepID        string `json:"step_id,omitempty" yaml:"step_id,omitempty"`
	StepVersion   string `json:"step_version,omitempty" yaml:"step_version,omitempty"`
	LatestVersion string `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`
}

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

// StepmanStepInfo ...
func StepmanStepInfo(collection, stepID, stepVersion string) (StepInfoModel, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--collection", collection,
		"--id", stepID, "--version", stepVersion}
	out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
	if err != nil {
		return StepInfoModel{}, fmt.Errorf("Failed to run stepman step-info, err: %s", err)
	}

	stepInfo := StepInfoModel{}
	if err := json.Unmarshal([]byte(out), &stepInfo); err != nil {
		return StepInfoModel{}, err
	}

	return stepInfo, nil
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
func EnvmanInitAtPath(pth string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", pth, "init", "--clear"}
	return cmdex.RunCommand("envman", args...)
}

// EnvmanAdd ...
func EnvmanAdd(pth, key, value string, expand bool) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", pth, "add", "--key", key, "--append"}
	if !expand {
		args = []string{"--loglevel", logLevel, "--path", pth, "add", "--key", key, "--no-expand", "--append"}
	}

	envman := exec.Command("envman", args...)
	envman.Stdin = strings.NewReader(value)
	envman.Stdout = os.Stdout
	envman.Stderr = os.Stderr
	return envman.Run()
}

// EnvmanRun ...
func EnvmanRun(pth, dir string, cmd []string, logLevel string) (int, error) {
	if logLevel == "" {
		logLevel = log.GetLevel().String()
	}
	args := []string{"--loglevel", logLevel, "--path", pth, "run"}
	args = append(args, cmd...)

	return cmdex.RunCommandInDirAndReturnExitCode(dir, "envman", args...)
}

// EnvmanEnvstoreTest ...
func EnvmanEnvstoreTest(pth string) error {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", pth, "print"}
	cmd := exec.Command("envman", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
