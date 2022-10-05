package cli

import (
	"fmt"

	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/bitrise/tools"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/urfave/cli"
)

func stepList(c *cli.Context) error {
	warnings := []string{}

	// Expand cli.Context
	bitriseConfigBase64Data := c.String(ConfigBase64Key)

	bitriseConfigPath := c.String(ConfigKey)
	deprecatedBitriseConfigPath := c.String(PathKey)
	if bitriseConfigPath == "" && deprecatedBitriseConfigPath != "" {
		warnings = append(warnings, "'path' key is deprecated, use 'config' instead!")
		bitriseConfigPath = deprecatedBitriseConfigPath
	}

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
		break
	case output.FormatJSON:
		outStr, err := tools.StepmanJSONStepList(collectionURI)
		if err != nil {
			registerFatal(fmt.Sprintf("Failed to print step info, err: %s", err), warnings, format)
		}
		log.Print(outStr)
		break
	default:
		registerFatal(fmt.Sprintf("Invalid format: %s", format), warnings, output.FormatJSON)
	}

	return nil
}
