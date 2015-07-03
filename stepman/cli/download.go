package cli

import (
	"fmt"

	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

func download(c *cli.Context) {
	fmt.Println("Download")

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

	stepCollection, err := stepman.ReadStepSpec()
	if err != nil {
		return
	}

	exist, step := stepCollection.GetStep(id, version)
	if exist == false {
		fmt.Printf("Step: %s - (%s) dos not exist", id, version)
		return
	}

	err = stepman.DownloadStep(step)
	if err != nil {
		fmt.Println("Failed to download step")
	}
}
