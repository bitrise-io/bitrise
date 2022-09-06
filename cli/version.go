package cli

import (
	"fmt"
	"log"

	"runtime"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/bitrise/version"
	"github.com/urfave/cli"
)

// VersionOutputModel ...
type VersionOutputModel struct {
	Version       string `json:"version"`
	FormatVersion string `json:"format_version"`
	OS            string `json:"os"`
	GO            string `json:"go"`
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
		versionOutput.OS = fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)
		versionOutput.GO = runtime.Version()
	}

	if output.Format == output.FormatRaw {
		if fullVersion {
			versionStr := fmt.Sprintf(`version: %s
format version: %s
os: %s
go: %s
build number: %s
commit: %s
`, versionOutput.Version, versionOutput.FormatVersion, versionOutput.OS, versionOutput.GO, versionOutput.BuildNumber, versionOutput.Commit)
			log.Println(versionStr)
		} else {
			versionStr := fmt.Sprintf("%s", versionOutput.Version)
			log.Println(versionStr)
		}
	} else {
		output.Print(versionOutput, output.Format)
	}

	return nil
}
