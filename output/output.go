package output

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	// FormatKey ...
	FormatKey = "format"
	// FormatRaw ...
	FormatRaw = "raw"
	// FormatJSON ...
	FormatJSON = "json"
	// FormatYML ...
	FormatYML = "yml"
)

// Format ...
var Format = FormatRaw

// ConfigureOutputFormat ...
func ConfigureOutputFormat(c *cli.Context) error {
	outFmt := c.String(FormatKey)
	switch outFmt {
	case FormatRaw, FormatJSON, FormatYML:
		// valid
		Format = outFmt
	case "":
		// default
		Format = FormatRaw
	default:
		// invalid
		return fmt.Errorf("Invalid Output Format: %s", outFmt)
	}
	return nil
}

// Print ...
func Print(outModel interface{}, format string) {
	switch format {
	case FormatJSON:
		serBytes, err := json.Marshal(outModel)
		if err != nil {
			log.Errorf("[.print] ERROR: %s", err)
			return
		}
		fmt.Printf("%s\n", serBytes)
	case FormatYML:
		serBytes, err := yaml.Marshal(outModel)
		if err != nil {
			log.Errorf("[output.print] ERROR: %s", err)
			return
		}
		fmt.Printf("%s\n", serBytes)
	default:
		log.Errorf("[output.print] Invalid output format: %s", format)
	}
}
