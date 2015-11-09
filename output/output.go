package output

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/configs"
)

// Print ...
func Print(outModel interface{}, format string) {
	switch format {
	case configs.OutputFormatJSON:
		serBytes, err := json.Marshal(outModel)
		if err != nil {
			log.Errorf("[output.print] ERROR: %s", err)
			return
		}
		fmt.Printf("%s\n", serBytes)
	case configs.OutputFormatYML:
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
