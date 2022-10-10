package cli

import (
	"encoding/json"
	"fmt"
	"runtime"

	flog "github.com/bitrise-io/go-utils/log"
	ver "github.com/bitrise-io/stepman/version"
	"github.com/urfave/cli"
)

// VersionOutputModel ...
type VersionOutputModel struct {
	Version     string `json:"version,omitempty"`
	OS          string `json:"os,omitempty"`
	GO          string `json:"go,omitempty"`
	BuildNumber string `json:"build_number,omitempty"`
	Commit      string `json:"commit,omitempty"`

	FullVersion bool `json:"-"`
}

// String ...
func (version VersionOutputModel) String() string {
	str := ""
	if version.FullVersion {
		str += fmt.Sprintf("version: %s\n", version.Version)
		str += fmt.Sprintf("os: %s\n", version.OS)
		str += fmt.Sprintf("go: %s\n", version.GO)
		str += fmt.Sprintf("build_number: %s\n", version.BuildNumber)
		str += fmt.Sprintf("commit: %s", version.Commit)
	} else {
		str = version.Version
	}

	return str
}

// JSON ...
func (version VersionOutputModel) JSON() string {
	if version.FullVersion {
		bytes, err := json.Marshal(version)
		if err != nil {
			return fmt.Sprintf(`"Failed to marshal version (%#v), err: %s"`, version, err)
		}
		return string(bytes) + "\n"
	}

	return fmt.Sprintf(`"%s"`+"\n", version.Version)
}

func printVersionCmd(c *cli.Context) error {
	fullVersion := c.Bool("full")
	format := c.String("format")
	if format == "" {
		format = "raw"
	}

	var log flog.Logger
	if format == "raw" {
		log = flog.NewDefaultRawLogger()
	} else if format == "json" {
		log = flog.NewDefaultJSONLoger()
	} else {
		failf("Invalid format: %s", format)
	}

	versionOutput := VersionOutputModel{}
	versionOutput.Version = ver.Version
	versionOutput.OS = fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)
	versionOutput.GO = runtime.Version()
	versionOutput.BuildNumber = ver.BuildNumber
	versionOutput.Commit = ver.Commit
	versionOutput.FullVersion = fullVersion

	log.Print(versionOutput)

	return nil
}
