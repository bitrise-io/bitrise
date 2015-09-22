package cli

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/codegangsta/cli"
)

func stepList(c *cli.Context) {
	// Input validation
	format := c.String(OuputFormatKey)
	if format == "" {
		format = OutputFormatRaw
	} else if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), OutputFormatJSON)
	}

	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		bitriseConfig, err := CreateBitriseConfigFromCLIParams(c)
		if err != nil {
			registerFatal(fmt.Sprintf("No collection defined and faild to read bitrise cofing, err: %s", err), format)
		}

		if bitriseConfig.DefaultStepLibSource == "" {
			registerFatal("No collection defined and no default collection found in bitrise cofing", format)
		}

		collectionURI = bitriseConfig.DefaultStepLibSource
	}

	switch format {
	case OutputFormatRaw:
		if err := bitrise.StepmanPrintRawStepList(collectionURI); err != nil {
			registerFatal(fmt.Sprintf("Failed to print step info, err: %s", err), format)
		}
		break
	case OutputFormatJSON:
		stepInfo, err := bitrise.StepmanStepList(collectionURI)
		if err != nil {
			registerFatal(fmt.Sprintf("Failed to print step info, err: %s", err), format)
		}
		bytes, err := json.Marshal(stepInfo)
		if err != nil {
			if err != nil {
				registerFatal(fmt.Sprintf("Failed to print step info, err: %s", err), format)
			}
		}
		fmt.Println(string(bytes))
		break
	default:
		log.Fatalf("Invalid format: %s", format)
		registerFatal(fmt.Sprintf("Invalid format: %s", format), OutputFormatJSON)
	}
}
