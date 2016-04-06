package cli

import (
	"fmt"
	"log"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/bitrise/version"
	"github.com/codegangsta/cli"
)

// VersionOutputModel ...
type VersionOutputModel struct {
	Version     string `json:"version"`
	BuildNumber string `json:"build_number,omitempty"`
}

func printVersionCmd(c *cli.Context) {
	fullVersion := c.Bool("full")

	if err := configs.ConfigureOutputFormat(c); err != nil {
		log.Fatalf("Error: %s", err)
	}

	versionOutput := VersionOutputModel{
		Version: version.VERSION,
	}

	if fullVersion {
		versionOutput.BuildNumber = version.BuildNumber
	}

	if configs.OutputFormat == configs.OutputFormatRaw {
		if fullVersion {
			fmt.Fprintf(c.App.Writer, "%v (%v)\n", versionOutput.Version, versionOutput.BuildNumber)
		} else {
			fmt.Fprintf(c.App.Writer, "%v\n", versionOutput.Version)
		}
	} else {
		output.Print(versionOutput, configs.OutputFormat)
	}
}
