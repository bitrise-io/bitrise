package cli

import (
	"fmt"

	"github.com/bitrise-io/stepman/git"
	"github.com/bitrise-io/stepman/paths"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func setup(c *cli.Context) {
	fmt.Println("Setup")

	err := paths.SetupCurrentRouting()
	if err != nil {
		fmt.Println("Failed to setup routing:", err)
		return
	}

	pth := paths.GetCurrentStepCollectionPath()

	err = git.DoGitClone(paths.STEP_COLLECTION_GIT, pth)
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
