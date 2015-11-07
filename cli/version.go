package cli

import (
	"fmt"
	"log"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/output"
	"github.com/codegangsta/cli"
)

// VersionOutputModel ...
type VersionOutputModel struct {
	Version string `json:"version"`
}

func printVersionCmd(c *cli.Context) {
	if err := configs.ConfigureOutputFormat(c); err != nil {
		log.Fatalf("Error: %s", err)
	}

	if configs.OutputFormat == configs.OutputFormatRaw {
		fmt.Fprintf(c.App.Writer, "%v\n", c.App.Version)
	} else {
		output.Print(VersionOutputModel{c.App.Version}, configs.OutputFormat)
	}
}
