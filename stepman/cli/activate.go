package cli

import (
	"fmt"

	"github.com/bitrise-io/go-pathutil"
	"github.com/bitrise-io/stepman/git"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func activate(c *cli.Context) {
	fmt.Println("Activate -- Coming soon")

	id := c.String(ID_KEY)
	if id == "" {
		fmt.Println("Missing step id")
		return
	}

	version := c.String(VERSION_KEY)
	if version == "" {
		fmt.Println("Missing step version")
		return
	}

	path := c.String(PATH_KEY)
	if path == "" {
		fmt.Println("Missing copy path")
		return
	}

	stepCollection, err := stepman.ReadStepSpec()
	if err != nil {
		return
	}

	exist, step := stepCollection.GetStep(id, version)
	if exist == false {
		fmt.Printf("Step: %s - (%s) dos not exist", id, version)
		return
	}

	pth := step.Path()
	exist, err = pathutil.IsPathExists(pth)
	if err != nil {
		fmt.Printf("Failed to check path:", err)
		return
	}
	if exist == false {
		fmt.Println("Step dos not exist, download it")
		step.Download()
	}

	srcFolder := pth
	destFolder := path
	git.RunCommand("cp", []string{"-rf", srcFolder, destFolder}...)
}
