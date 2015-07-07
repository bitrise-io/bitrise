package cli

import (
	"fmt"

	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func setup(c *cli.Context) {
	fmt.Println("Setup")

	if err := stepman.SetupCurrentRouting(); err != nil {
		fmt.Println("Failed to setup routing:", err)
		return
	}

	pth := stepman.GetCurrentStepCollectionPath()
	if err := stepman.DoGitClone(stepman.STEP_COLLECTION_GIT, pth); err != nil {
		fmt.Println("Failed to initialize Stepman:", err)
		return
	}

	if err := stepman.WriteStepSpecToFile(); err != nil {
		fmt.Println("Failed to initialize Stepman:", err)
		return
	}

	fmt.Println("Stepman initialized")
}
