package cli

import (
	"fmt"
	"runtime"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/output"
	"github.com/bitrise-io/bitrise/v2/version"
	"github.com/spf13/cobra"
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

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version",
	RunE:  printVersionCmd,
}

var versionOpts struct {
	full   bool
	format string
}

func init() {
	versionCmd.Flags().BoolVar(&versionOpts.full, "full", false, "Prints the build number as well.")
	versionCmd.Flags().StringVarP(&versionOpts.format, OuputFormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
}

func printVersionCmd(cmd *cobra.Command, args []string) error {
	logCommandParameters(cmd)

	fullVersion := versionOpts.full

	if err := output.ConfigureOutputFormat(versionOpts.format); err != nil {
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
