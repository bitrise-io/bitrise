package cli

import (
	"fmt"
	"runtime"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/output"
	"github.com/bitrise-io/bitrise/v2/version"
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
	logCommandParameters(c)

	fullVersion := c.Bool("full")

	if err := output.ConfigureOutputFormat(c); err != nil {
		failf("Failed to configure output format, error: %s", err)
	}

	versionOutput := VersionOutputModel{
		Version: version.VERSION,
	}

	if fullVersion {
		versionOutput.FormatVersion = models.FormatVersion
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
			log.Print(versionStr)
		} else {
			log.Print(versionOutput.Version)
		}
	} else {
		output.Print(versionOutput, output.Format)
	}

	return nil
}
