package main

import (
	"fmt"

	"github.com/bitrise-io/go-pathutil"
)

const (
	STEP_COLLECTION_GIT string = "https://github.com/steplib/steplib.git"
	STEP_COLLECTION_DIR string = "/.stepMan/step_collection"
	STEPS_DIR           string = "/.stepMan/step_collection/steps"
	STEP_SPEC_DIR       string = "/.stepMan/step_spec/step_spec.json"
)

// Interface commands
func doUpdateCommand(stepsSpecDir string) error {
	isStorePathExists, err := pathutil.IsPathExists(stepsSpecDir)
	if err != nil {
		return err
	}
	if isStorePathExists == false {
		fmt.Println("StepsSpec path does not exist, do clone")
		return cloneStepsSpecs(stepsSpecDir)
	}

	fmt.Println("StepsSpec path exist, do pull")
	return pullStepsSpec(stepsSpecDir)
}

func doGenerateStepsSpec() error {
	err := writeStepSpecToFile()
	if err != nil {
		return err
	}
	return nil
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
}
