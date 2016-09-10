package cli

import (
	"fmt"
	"log"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/bitrise/version"
	"github.com/urfave/cli"
)

// VersionOutputModel ...
type VersionOutputModel struct {
	Version       string `json:"version"`
	FormatVersion string `json:"format_version"`
	BuildNumber   string `json:"build_number"`
	Commit        string `json:"commit"`
}

func printVersionCmd(c *cli.Context) error {
	fullVersion := c.Bool("full")

	if err := output.ConfigureOutputFormat(c); err != nil {
		log.Fatalf("Failed to configure output format, error: %s", err)
	}

	versionOutput := VersionOutputModel{
		Version: version.VERSION,
	}

	if fullVersion {
		versionOutput.FormatVersion = models.Version
		versionOutput.BuildNumber = version.BuildNumber
		versionOutput.Commit = version.Commit
	}

	if output.Format == output.FormatRaw {
		if fullVersion {
			fmt.Fprintf(c.App.Writer, "version: %v\nformat version: %v\nbuild number: %v\ncommit: %v\n", versionOutput.Version, versionOutput.FormatVersion, versionOutput.BuildNumber, versionOutput.Commit)
		} else {
			fmt.Fprintf(c.App.Writer, "%v\n", versionOutput.Version)
		}
	} else {
		output.Print(versionOutput, output.Format)
	}

	return nil
}
