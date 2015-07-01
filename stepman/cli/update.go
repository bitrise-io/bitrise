package cli

import (
	"fmt"

	"github.com/bitrise-io/go-pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func update(c *cli.Context) {
	fmt.Println("Update")

	stepCollectionPath := stepman.GetCurrentStepCollectionPath()
	exists, err := pathutil.IsPathExists(stepCollectionPath)
	if err != nil {
		fmt.Println("Failed to update Stepman:", err)
		return
	}
	if exists == false {
		fmt.Println("Stepman is not initialized")
		return
	}

	err = stepman.DoGitPull(stepCollectionPath)
	if err != nil {
		fmt.Println("Failed tp update Stepman:", err)
		return
	}

	err = stepman.WriteStepSpecToFile()
	if err != nil {
		fmt.Println("Failed to initialize Stepman:", err)
		return
	}

	fmt.Println("Stepman updated")
}
