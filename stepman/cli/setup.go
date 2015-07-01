package cli

import (
	"fmt"

	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func setup(c *cli.Context) {
	fmt.Println("Setup")

	err := stepman.SetupCurrentRouting()
	if err != nil {
		fmt.Println("Failed to setup routing:", err)
		return
	}

	pth := stepman.GetCurrentStepCollectionPath()

	err = stepman.DoGitClone(stepman.STEP_COLLECTION_GIT, pth)
	if err != nil {
		fmt.Println("Failed to initialize Stepman:", err)
		return
	}

	err = stepman.WriteStepSpecToFile()
	if err != nil {
		fmt.Println("Failed to initialize Stepman:", err)
		return
	}

	fmt.Println("Stepman initialized")
}
