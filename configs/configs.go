package configs

import (
	"fmt"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/version"
	"github.com/codegangsta/cli"
	ver "github.com/hashicorp/go-version"
)

// ---------------------------
// --- Project level vars / configs

var (
	// IsCIMode ...
	IsCIMode = false
	// IsDebugMode ...
	IsDebugMode = false
	// IsPullRequestMode ...
	IsPullRequestMode = false
)

// OutputFormat ...
var OutputFormat = OutputFormatRaw

// ---------------------------
// --- Consts

const (
	// OuputFormatKey ...
	OuputFormatKey = "format"
	// OutputFormatRaw ...
	OutputFormatRaw = "raw"
	// OutputFormatJSON ...
	OutputFormatJSON = "json"
	// OutputFormatYML ...
	OutputFormatYML = "yml"
)

// ConfigureOutputFormat ...
func ConfigureOutputFormat(c *cli.Context) error {
	outFmt := c.String(OuputFormatKey)
	switch outFmt {
	case OutputFormatRaw, OutputFormatJSON, OutputFormatYML:
		// valid
		OutputFormat = outFmt
	case "":
		// default
		OutputFormat = OutputFormatRaw
	default:
		// invalid
		return fmt.Errorf("Invalid Output Format: %s", outFmt)
	}
	return nil
}

// BitriseVersion ...
func BitriseVersion() (ver.Version, error) {
	bitriseVersionPtr, err := ver.NewVersion(version.VERSION)
	if err != nil {
		return ver.Version{}, err
	}
	if bitriseVersionPtr == nil {
		return ver.Version{}, fmt.Errorf("Failed to parse version (%s)", version.VERSION)
	}

	return *bitriseVersionPtr, nil
}

// VersionMap ...
func VersionMap() (map[string]ver.Version, error) {
	envmanVersion, err := bitrise.EnvmanVersion()
	if err != nil {
		return map[string]ver.Version{}, err
	}

	stepmanVersion, err := bitrise.StepmanVersion()
	if err != nil {
		return map[string]ver.Version{}, err
	}

	bitriseVersion, err := BitriseVersion()
	if err != nil {
		return map[string]ver.Version{}, err
	}

	return map[string]ver.Version{
		"bitrise": bitriseVersion,
		"envman":  envmanVersion,
		"stepman": stepmanVersion,
	}, nil
}
