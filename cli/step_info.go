package cli

import (
	"fmt"

	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/bitrise/tools"
	log "github.com/bitrise-io/go-utils/v2/advancedlog"
	"github.com/urfave/cli"
)

func printStepLibStep(collectionURI, id, version, format string) error {
	switch format {
	case output.FormatRaw:
		out, err := tools.StepmanRawStepLibStepInfo(collectionURI, id, version)
		if out != "" {
			log.Println("Step info:")
			log.Printf("%s", out)
		}
		return err
	case output.FormatJSON:
		outStr, err := tools.StepmanJSONStepLibStepInfo(collectionURI, id, version)
		if err != nil {
			return fmt.Errorf("StepmanJSONStepLibStepInfo failed, err: %s", err)
		}
		log.Println(outStr)
		break
	default:
		return fmt.Errorf("Invalid format: %s", format)
	}
	return nil
}

func printLocalStepInfo(pth, format string) error {
	switch format {
	case output.FormatRaw:
		out, err := tools.StepmanRawLocalStepInfo(pth)
		if out != "" {
			log.Println("Step info:")
			log.Printf("%s", out)
		}
		return err
	case output.FormatJSON:
		outStr, err := tools.StepmanJSONLocalStepInfo(pth)
		if err != nil {
			return fmt.Errorf("StepmanJSONLocalStepInfo failed, err: %s", err)
		}
		log.Println(outStr)
		break
	default:
		return fmt.Errorf("Invalid format: %s", format)
	}
	return nil
}

func stepInfo(c *cli.Context) error {
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

	YMLPath := c.String(StepYMLKey)
	collectionURI := c.String(CollectionKey)

	id := ""
	if len(c.Args()) < 1 {
		registerFatal("No step specified!", warnings, format)
	} else {
		id = c.Args()[0]
	}

	version := c.String(VersionKey)
	//

	if format == "" {
		format = output.FormatRaw
	} else if !(format == output.FormatRaw || format == output.FormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), warnings, output.FormatJSON)
	}

	if YMLPath != "" {
		//
		// Local step info
		if err := printLocalStepInfo(YMLPath, format); err != nil {
			registerFatal(fmt.Sprintf("Failed to print step info (yml path: %s), err: %s", YMLPath, err), warnings, format)
		}
	} else {

		//
		// Steplib step info
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

		if err := printStepLibStep(collectionURI, id, version, format); err != nil {
			registerFatal(fmt.Sprintf("Failed to print step info, err: %s", err), warnings, format)
		}
	}

	return nil
}
