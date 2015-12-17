package cli

import (
	"fmt"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/configs"
	"github.com/codegangsta/cli"
)

func printStepLibStep(collectionURI, id, version, format string) error {
	switch format {
	case configs.OutputFormatRaw:
		out, err := bitrise.StepmanRawStepLibStepInfo(collectionURI, id, version)
		if out != "" {
			fmt.Println("Step info:")
			fmt.Printf("%s", out)
		}
		return err
	case configs.OutputFormatJSON:
		outStr, err := bitrise.StepmanJSONStepLibStepInfo(collectionURI, id, version)
		if err != nil {
			return fmt.Errorf("StepmanJSONStepLibStepInfo failed, err: %s", err)
		}
		fmt.Println(outStr)
		break
	default:
		return fmt.Errorf("Invalid format: %s", format)
	}
	return nil
}

func printLocalStepInfo(pth, format string) error {
	switch format {
	case configs.OutputFormatRaw:
		out, err := bitrise.StepmanRawLocalStepInfo(pth)
		if out != "" {
			fmt.Println("Step info:")
			fmt.Printf("%s", out)
		}
		return err
	case configs.OutputFormatJSON:
		outStr, err := bitrise.StepmanJSONLocalStepInfo(pth)
		if err != nil {
			return fmt.Errorf("StepmanJSONLocalStepInfo failed, err: %s", err)
		}
		fmt.Println(outStr)
		break
	default:
		return fmt.Errorf("Invalid format: %s", format)
	}
	return nil
}

func stepInfo(c *cli.Context) {
	format := c.String(OuputFormatKey)
	if format == "" {
		format = configs.OutputFormatRaw
	} else if !(format == configs.OutputFormatRaw || format == configs.OutputFormatJSON) {
		registerFatal(fmt.Sprintf("Invalid format: %s", format), configs.OutputFormatJSON)
	}

	YMLPath := c.String(StepYMLKey)
	if YMLPath != "" {
		//
		// Local step info
		if err := printLocalStepInfo(YMLPath, format); err != nil {
			registerFatal(fmt.Sprintf("Failed to print step info (yml path: %s), err: %s", YMLPath, err), format)
		}
	} else {
		//
		// Steplib step info
		collectionURI := c.String(CollectionKey)
		if collectionURI == "" {
			bitriseConfig, err := CreateBitriseConfigFromCLIParams(c)
			if err != nil {
				registerFatal(fmt.Sprintf("No collection defined and failed to read bitrise config, err: %s", err), format)
			}

			if bitriseConfig.DefaultStepLibSource == "" {
				registerFatal("No collection defined and no default collection found in bitrise config", format)
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
			registerFatal(fmt.Sprintf("Failed to print step info, err: %s", err), format)
		}
	}
}
