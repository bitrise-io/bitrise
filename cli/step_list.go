package cli

import (
	"fmt"
	"log"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/codegangsta/cli"
)

func stepList(c *cli.Context) {
	// Input validation
	format := c.String(OuputFormatKey)
	if format == "" {
		format = configs.OutputFormatRaw
	} else if !(format == configs.OutputFormatRaw || format == configs.OutputFormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), configs.OutputFormatJSON)
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
	case configs.OutputFormatRaw:
		out, err := bitrise.StepmanRawStepList(collectionURI)
		if out != "" {
			fmt.Println("Step list:")
			fmt.Printf("%s", out)
		}
		if err != nil {
			registerFatal(fmt.Sprintf("Failed to print step info, err: %s", err), format)
		}
		break
	case configs.OutputFormatJSON:
		outStr, err := bitrise.StepmanJSONStepList(collectionURI)
		if err != nil {
			registerFatal(fmt.Sprintf("Failed to print step info, err: %s", err), format)
		}
		fmt.Println(outStr)
		break
	default:
		log.Fatalf("Invalid format: %s", format)
		registerFatal(fmt.Sprintf("Invalid format: %s", format), configs.OutputFormatJSON)
	}
}
