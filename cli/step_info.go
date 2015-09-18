package cli

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/codegangsta/cli"
)

func stepInfo(c *cli.Context) {
	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		bitriseConfig, err := CreateBitriseConfigFromCLIParams(c)
		if err != nil {
			log.Fatalf("No collection defined and faild to read bitrise cofing, err: %s", err)
		}

		if bitriseConfig.DefaultStepLibSource == "" {
			log.Fatal("No collection defined and no default collection found in bitrise cofing")
		}

		collectionURI = bitriseConfig.DefaultStepLibSource
	}

	id := c.String(IDKey)
	if id == "" {
		log.Fatal("Missing step id")
	}

	format := c.String(OuputFormatKey)
	if format == "" {
		format = OutputFormatRaw
	} else if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		log.Fatalf("Invalid format: %s", format)
	}

	version := c.String(VersionKey)

	switch format {
	case OutputFormatRaw:
		if err := bitrise.StepmanPrintRawStepInfo(collectionURI, id, version); err != nil {
			log.Fatalf("Failed to print step info, err: %s", err)
		}
		break
	case OutputFormatJSON:
		stepInfo, err := bitrise.StepmanStepInfo(collectionURI, id, version)
		if err != nil {
			log.Fatalf("Failed to print step info, err: %s", err)
		}
		bytes, err := json.Marshal(stepInfo)
		if err != nil {
			if err != nil {
				log.Fatalf("Failed to print step info, err: %s", err)
			}
		}
		fmt.Println(string(bytes))
		break
	default:
		log.Fatalf("Invalid format: %s", format)
	}
}
