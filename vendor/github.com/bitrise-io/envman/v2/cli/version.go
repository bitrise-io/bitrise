package cli

import (
	"fmt"
	"log"

	"github.com/bitrise-io/envman/v2/output"
	"github.com/bitrise-io/envman/v2/version"
	"github.com/urfave/cli"
)

// VersionOutputModel ...
type VersionOutputModel struct {
	Version     string `json:"version"`
	BuildNumber string `json:"build_number"`
	Commit      string `json:"commit"`
}

func printVersionCmd(c *cli.Context) error {
	fullVersion := c.Bool("full")

	if err := output.ConfigureOutputFormat(c); err != nil {
		log.Fatalf("Error: %s", err)
	}

	versionOutput := VersionOutputModel{
		Version: version.Version,
	}

	if fullVersion {
		versionOutput.BuildNumber = version.BuildNumber
		versionOutput.Commit = version.Commit
	}

	if output.Format == output.FormatRaw {
		if fullVersion {
			fmt.Printf("version: %v\nbuild_number: %v\ncommit: %v\n", versionOutput.Version, versionOutput.BuildNumber, versionOutput.Commit)
		} else {
			fmt.Println(versionOutput.Version)
		}
	} else {
		output.Print(versionOutput, output.Format)
	}

	return nil
}
