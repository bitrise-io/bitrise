package output

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/bitrise/v2/log"
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
func ConfigureOutputFormat(format string) error {
	switch format {
	case FormatRaw, FormatJSON, FormatYML:
		Format = format
	case "":
		Format = FormatRaw
	default:
		return fmt.Errorf("invalid output format: %s", format)
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
		log.Printf("%s", serBytes)
	case FormatYML:
		serBytes, err := yaml.Marshal(outModel)
		if err != nil {
			log.Errorf("[output.print] ERROR: %s", err)
			return
		}
		log.Printf("%s", serBytes)
	default:
		log.Errorf("[output.print] Invalid output format: %s", format)
	}
}
