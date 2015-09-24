package cli

import (
	"encoding/json"
	"fmt"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/codegangsta/cli"
)

func printStepLibStep(collectionURI, id, version, format string) error {
	switch format {
	case OutputFormatRaw:
		if err := bitrise.StepmanPrintRawStepLibStepInfo(collectionURI, id, version); err != nil {
			return err
		}
		break
	case OutputFormatJSON:
		stepInfo, err := bitrise.StepmanStepLibStepInfo(collectionURI, id, version)
		if err != nil {
			return err
		}
		bytes, err := json.Marshal(stepInfo)
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		break
	default:
		return fmt.Errorf("Invalid format: %s", format)
	}
	return nil
}

func printLocalStepInfo(pth, format string) error {
	switch format {
	case OutputFormatRaw:
		if err := bitrise.StepmanPrintRawLocalStepInfo(pth); err != nil {
			return err
		}
		break
	case OutputFormatJSON:
		stepInfo, err := bitrise.StepmanLocalStepInfo(pth)
		if err != nil {
			return err
		}
		bytes, err := json.Marshal(stepInfo)
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		break
	default:
		return fmt.Errorf("Invalid format: %s", format)
	}
	return nil
}

func stepInfo(c *cli.Context) {
	format := c.String(OuputFormatKey)
	if format == "" {
		format = OutputFormatRaw
	} else if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), OutputFormatJSON)
	}

	YMLPath := c.String(StepYMLKey)
	if YMLPath != "" {
		//
		// Local step info
		if err := printLocalStepInfo(YMLPath, format); err != nil {
			registerFatal(fmt.Sprintf("Faild to print step info, err: %s", err), format)
		}
	} else {
		//
		// Steplib step info
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

		id := ""
		if len(c.Args()) < 1 {
			registerFatal("No step specified!", format)
		} else {
			id = c.Args()[0]
		}

		version := c.String(VersionKey)

		if err := printStepLibStep(collectionURI, id, version, format); err != nil {
			registerFatal(fmt.Sprintf("Faild to print step info, err: %s", err), format)
		}
	}
}
