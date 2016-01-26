package configs

import (
	"fmt"

	"github.com/codegangsta/cli"
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
	// OutputFormat ...
	OutputFormat = OutputFormatRaw
	// OptOutUsageData ...
	OptOutUsageData = false
)

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
