package main

import (
	"errors"
	"fmt"
	_ "os"

	"github.com/bitrise-io/go-pathutil"
)

const (
	STEP_COLLECTION_GIT string = "https://github.com/steplib/steplib.git"
	STEP_COLLECTION_DIR string = "/.stepMan/step_collection"
	STEPS_DIR           string = "/.stepMan/step_collection/steps"

	STEP_SPEC_DIR string = "/.stepMan/step_spec/step_spec.json"

	STEP_CACHE_DIR string = "/.stepMan/step_cache/"
)

// Interface commands
func doUpdateCommand(stepsSpecDir string) error {
	return doGitUpdate(STEP_COLLECTION_GIT, stepsSpecDir)
}

func doGenerateStepsSpec() error {
	err := writeStepSpecToFile()
	if err != nil {
		return err
	}
	return nil
}

func doActivateStep(id, version, pth string) error {
	stepCollection, err := readStepSpec()
	if err != nil {
		return err
	}

	exist, step := stepCollection.GetStep(id, version)
	if exist {
		git := step.Source["git"]
		pth := pathutil.UserHomeDir() + STEP_CACHE_DIR + id + "/" + version + "/"

		return doGitUpdate(git, pth)
	} else {
		return errors.New(fmt.Sprintf("Step: %s - (%s) dos not exist", id, version))
	}
}

func main() {
	fmt.Println("stepman")

	fmt.Println("start updating")

	stepsSpecDir := pathutil.UserHomeDir() + STEP_COLLECTION_DIR
	err := doUpdateCommand(stepsSpecDir)
	if err != nil {
		fmt.Println("Failed to update:", err)
	} else {
		fmt.Println("Update success!")
	}

	err = doGenerateStepsSpec()
	if err != nil {
		fmt.Println("Failed to write spec:", err)
	}

	stepId := "activate-ssh-key"
	stepVersion := "1.0.0"
	pth := STEP_CACHE_DIR
	err = doActivateStep(stepId, stepVersion, pth)
	if err != nil {
		fmt.Println(err)
	}
}
