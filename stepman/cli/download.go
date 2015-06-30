package cli

import (
	"fmt"

	"github.com/codegangsta/cli"
)

func download(c *cli.Context) {
	fmt.Println("Download")

	/*
		stepCollection, err := readStepSpec()
		if err != nil {
			return err
		}

		exist, step := stepCollection.GetStep(id, version)
		if exist {
			return step.Download()
		} else {
			return errors.New(fmt.Sprintf("Step: %s - (%s) dos not exist", id, version))
		}
	*/
}
