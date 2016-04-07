package cli

import (
	"fmt"
	"log"

	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/codegangsta/cli"
)

func stepList(c *cli.Context) {
	warnings := []string{}

	// Input validation
	format := c.String(OuputFormatKey)
	if format == "" {
		format = output.FormatRaw
	} else if !(format == output.FormatRaw || format == output.FormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), []string{}, output.FormatJSON)
	}

	collectionURI := c.String(CollectionKey)
	if collectionURI == "" {
		bitriseConfig, warns, err := CreateBitriseConfigFromCLIParams(c)
		warnings = warns
		if err != nil {
			registerFatal(fmt.Sprintf("No collection defined and failed to read bitrise config, err: %s", err), warnings, format)
		}

		if bitriseConfig.DefaultStepLibSource == "" {
			registerFatal("No collection defined and no default collection found in bitrise config", warnings, format)
		}

		collectionURI = bitriseConfig.DefaultStepLibSource
	}

	switch format {
	case output.FormatRaw:
		out, err := tools.StepmanRawStepList(collectionURI)
		if out != "" {
			fmt.Println("Step list:")
			fmt.Printf("%s", out)
		}
		if err != nil {
			registerFatal(fmt.Sprintf("Failed to print step info, err: %s", err), warnings, format)
		}
		break
	case output.FormatJSON:
		outStr, err := tools.StepmanJSONStepList(collectionURI)
		if err != nil {
			registerFatal(fmt.Sprintf("Failed to print step info, err: %s", err), warnings, format)
		}
		fmt.Println(outStr)
		break
	default:
		log.Fatalf("Invalid format: %s", format)
		registerFatal(fmt.Sprintf("Invalid format: %s", format), warnings, output.FormatJSON)
	}
}
