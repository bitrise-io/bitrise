package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/urfave/cli"
)

func stepList(c *cli.Context) error {
	warnings := []string{}

	// Expand cli.Context
	bitriseConfigBase64Data := c.String(ConfigBase64Key)

	bitriseConfigPath := c.String(ConfigKey)

	format := c.String(OuputFormatKey)

	collectionURI := c.String(CollectionKey)
	//

	// Input validation
	if format == "" {
		format = output.FormatRaw
	} else if !(format == output.FormatRaw || format == output.FormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), warnings, output.FormatJSON)
	}

	if collectionURI == "" {
		bitriseConfig, warns, err := CreateBitriseConfigFromCLIParams(bitriseConfigBase64Data, bitriseConfigPath)
		warnings = append(warnings, warns...)
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
			log.Print("Step list:")
			log.Printf("%s", out)
		}
		if err != nil {
			registerFatal(fmt.Sprintf("Failed to print step info, err: %s", err), warnings, format)
		}
	case output.FormatJSON:
		outStr, err := tools.StepmanJSONStepList(collectionURI)
		if err != nil {
			registerFatal(fmt.Sprintf("Failed to print step info, err: %s", err), warnings, format)
		}
		log.Print(outStr)
	default:
		registerFatal(fmt.Sprintf("Invalid format: %s", format), warnings, output.FormatJSON)
	}

	return nil
}

func registerFatal(errorMsg string, warnings []string, format string) {
	message := ValidationItemModel{
		IsValid:  (len(errorMsg) > 0),
		Error:    errorMsg,
		Warnings: warnings,
	}

	if format == output.FormatRaw {
		for _, warning := range message.Warnings {
			log.Warnf("warning: %s", warning)
		}
		failf(message.Error)
	} else {
		bytes, err := json.Marshal(message)
		if err != nil {
			failf("Failed to parse error model, error: %s", err)
		}

		log.Print(string(bytes))
		os.Exit(1)
	}
}
