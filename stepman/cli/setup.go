package cli

import (
	"fmt"

	"github.com/bitrise-io/go-pathutil"
	"github.com/bitrise-io/stepman/git"
	"github.com/bitrise-io/stepman/paths"
	"github.com/bitrise-io/stepman/step_util"
	"github.com/codegangsta/cli"
)

func setup(c *cli.Context) {
	fmt.Println("Setup")

	stepsSpecDir := pathutil.UserHomeDir() + paths.STEP_COLLECTION_DIR
	exists, err := pathutil.IsPathExists(stepsSpecDir)
	if err != nil {
		fmt.Println("Failed to update Stepman:", err)
		return
	}
	if exists == true {
		fmt.Println("Stepman already initialized")
		return
	}

	err = git.DoGitClone(paths.STEP_COLLECTION_GIT, stepsSpecDir)
	if err != nil {
		fmt.Println("Failed to initialize Stepman:", err)
		return
	}
	err = step_util.WriteStepSpecToFile()
	if err != nil {
		fmt.Println("Failed to initialize Stepman:", err)
		return
	}
	fmt.Println("Stepman initialized")
}
