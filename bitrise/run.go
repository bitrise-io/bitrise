package bitrise

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/cmdex"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

const (
	// VerifiedStepLibURI ...
	VerifiedStepLibURI = "https://github.com/bitrise-io/bitrise-steplib"
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

// StepmanPrintRawStepInfo ...
func StepmanPrintRawStepInfo(collection, stepID, stepVersion string) error {
	if collection == "" {
		collection = VerifiedStepLibURI
	}
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--collection", collection,
		"--id", stepID, "--version", stepVersion}
	out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
	if err != nil {
		return fmt.Errorf("Failed to run stepman step-info, err: %s", err)
	}

	fmt.Println(out)
	return nil
}

// StepmanStepInfo ...
func StepmanStepInfo(collection, stepID, stepVersion string) (stepmanModels.StepInfoModel, error) {
	if collection == "" {
		collection = VerifiedStepLibURI
	}
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-info", "--collection", collection,
		"--id", stepID, "--version", stepVersion, "--format", "json"}
	out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
	if err != nil {
		return stepmanModels.StepInfoModel{}, fmt.Errorf("Failed to run stepman step-info, err: %s", err)
	}

	stepInfo := stepmanModels.StepInfoModel{}
	if err := json.Unmarshal([]byte(out), &stepInfo); err != nil {
		return stepmanModels.StepInfoModel{}, err
	}

	return stepInfo, nil
}

// StepmanPrintRawStepList ...
func StepmanPrintRawStepList(collection string) error {
	if collection == "" {
		collection = VerifiedStepLibURI
	}
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-list", "--collection", collection, "--format", "raw"}
	out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
	if err != nil {
		return fmt.Errorf("Failed to run stepman step-list, err: %s", err)
	}

	fmt.Println(out)
	return nil
}

// StepmanStepList ...
func StepmanStepList(collection string) (stepmanModels.StepListModel, error) {
	if collection == "" {
		collection = VerifiedStepLibURI
	}
	logLevel := log.GetLevel().String()
	args := []string{"--debug", "--loglevel", logLevel, "step-list", "--collection", collection, "--format", "json"}
	out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("stepman", args...)
	if err != nil {
		return stepmanModels.StepListModel{}, fmt.Errorf("Failed to run stepman step-list, err: %s", err)
	}

	stepList := stepmanModels.StepListModel{}
	if err := json.Unmarshal([]byte(out), &stepList); err != nil {
		return stepmanModels.StepListModel{}, err
	}

	return stepList, nil
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
func EnvmanRun(envstorePth, workDirPth string, cmd []string, logLevel string) (int, error) {
	if logLevel == "" {
		logLevel = log.GetLevel().String()
	}
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "run"}
	args = append(args, cmd...)

	return cmdex.RunCommandInDirAndReturnExitCode(workDirPth, "envman", args...)
}

// EnvmanJSONPrint ...
func EnvmanJSONPrint(envstorePth string) (envmanModels.EnvsJSONListModel, error) {
	logLevel := log.GetLevel().String()
	args := []string{"--loglevel", logLevel, "--path", envstorePth, "print", "--format", "json", "--expand"}

	out, err := cmdex.RunCommandAndReturnCombinedStdoutAndStderr("envman", args...)
	if err != nil {
		return envmanModels.EnvsJSONListModel{}, fmt.Errorf("Failed to run envman print, err: %s", err)
	}

	envList := envmanModels.EnvsJSONListModel{}
	if err := json.Unmarshal([]byte(out), &envList); err != nil {
		return envmanModels.EnvsJSONListModel{}, err
	}

	return envList, nil
}
