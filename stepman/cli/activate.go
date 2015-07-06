package cli

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func activate(c *cli.Context) {
	fmt.Println("Activate")

	// Input validation
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

	// Get step
	stepCollection, err := stepman.ReadStepSpec()
	if err != nil {
		return
	}

	exist, step := stepCollection.GetStep(id, version)
	if exist == false {
		fmt.Printf("Step: %s - (%s) dos not exist", id, version)
		return
	}

	pth := stepman.GetStepPath(step)
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		fmt.Printf("Failed to check path:", err)
		return
	} else if exist == false {
		fmt.Println("Step dos not exist, download it")
		if err := stepman.DownloadStep(step); err != nil {
			fmt.Println("Failed to download step:", err)
		}
	}

	// Copy to specified path
	srcFolder := pth
	destFolder := path

	if exist, err = pathutil.IsPathExists(destFolder); err != nil {
		fmt.Printf("Failed to check path:", err)
		return
	} else if exist == false {
		if err := os.MkdirAll(destFolder, 0777); err != nil {
			fmt.Printf("Failed to create path:", err)
			return
		}
	}

	if err = stepman.RunCommand("cp", []string{"-rf", srcFolder, destFolder}...); err != nil {
		fmt.Printf("Failed to copy step:", err)
	}
}
