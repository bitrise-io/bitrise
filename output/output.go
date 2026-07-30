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
func ConfigureOutputFormat(outFmt string) error {
	switch outFmt {
	case FormatRaw, FormatJSON, FormatYML:
		// valid
		Format = outFmt
	case "":
		// default
		Format = FormatRaw
	default:
		// invalid
		return fmt.Errorf("invalid output format: %s", outFmt)
	}
	return nil
}

// Print marshals outModel per format and writes it via the logger. Returns an
// error instead of logging it, so a marshaling failure surfaces as a non-zero
// exit rather than a silently successful command.
func Print(outModel interface{}, format string) error {
	switch format {
	case FormatJSON:
		serBytes, err := json.Marshal(outModel)
		if err != nil {
			return fmt.Errorf("marshal json output: %w", err)
		}
		log.Printf("%s", serBytes)
	case FormatYML:
		serBytes, err := yaml.Marshal(outModel)
		if err != nil {
			return fmt.Errorf("marshal yml output: %w", err)
		}
		log.Printf("%s", serBytes)
	default:
		return fmt.Errorf("invalid output format: %s", format)
	}
	return nil
}
