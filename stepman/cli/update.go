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
	if exists, err := pathutil.IsPathExists(stepCollectionPath); err != nil {
		fmt.Println("Failed to update Stepman:", err)
		return
	} else if exists == false {
		fmt.Println("Stepman is not initialized")
		return
	}

	if err := stepman.DoGitPull(stepCollectionPath); err != nil {
		fmt.Println("Failed tp update Stepman:", err)
		return
	}

	if err := stepman.WriteStepSpecToFile(); err != nil {
		fmt.Println("Failed to initialize Stepman:", err)
		return
	}

	fmt.Println("Stepman updated")
}
