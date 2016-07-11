package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func getEnvInfos(envs []envmanModels.EnvironmentItemModel) ([]models.EnvInfoModel, error) {
	envInfos := []models.EnvInfoModel{}
	for _, env := range envs {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return []models.EnvInfoModel{}, err
		}

		options, err := env.GetOptions()
		if err != nil {
			return []models.EnvInfoModel{}, err
		}

		envInfo := models.EnvInfoModel{
			Key:          key,
			Description:  *options.Description,
			ValueOptions: options.ValueOptions,
			DefaultValue: value,
			IsExpand:     *options.IsExpand,
		}
		envInfos = append(envInfos, envInfo)
	}
	return envInfos, nil
}

func printRawEnvInfo(env models.EnvInfoModel) {
	if env.DefaultValue != "" {
		fmt.Printf("- %s: %s\n", colorstring.Green(env.Key), env.DefaultValue)
	} else {
		fmt.Printf("- %s\n", colorstring.Green(env.Key))
	}

	fmt.Printf("  %s: %v\n", colorstring.Green("is expand"), env.IsExpand)

	if len(env.ValueOptions) > 0 {
		fmt.Printf("  %s:\n", colorstring.Green("value options"))
		for _, option := range env.ValueOptions {
			fmt.Printf("  - %s\n", option)
		}
	}

	if env.Description != "" {
		fmt.Printf("  %s:\n", colorstring.Green("description"))
		fmt.Printf("  %s\n", env.Description)
	}
}

func printRawStepInfo(stepInfo models.StepInfoModel, isShort, isLocal bool) {
	if isLocal {
		fmt.Println(colorstring.Bluef("Local step info, yml path (%s):", stepInfo.StepLib))
	} else {
		fmt.Println(colorstring.Bluef("Step info in StepLib (%s):", stepInfo.StepLib))
	}

	if stepInfo.GlobalInfo.RemovalDate != "" {
		fmt.Println("")
		fmt.Println(colorstring.Red("This step is deprecated!"))
		fmt.Printf("%s %s\n", colorstring.Red("removal date:"), stepInfo.GlobalInfo.RemovalDate)

		if stepInfo.GlobalInfo.DeprecateNotes != "" {
			fmt.Printf("%s\n%s\n", colorstring.Red("deprecate notes:"), stepInfo.GlobalInfo.DeprecateNotes)
		}
	}

	if stepInfo.ID != "" {
		fmt.Printf("%s: %s\n", colorstring.Blue("ID"), stepInfo.ID)
	}
	if stepInfo.Version != "" {
		fmt.Printf("%s: %s\n", colorstring.Blue("version"), stepInfo.Version)
	}
	if stepInfo.Latest != "" {
		fmt.Printf("%s: %s\n", colorstring.Blue("latest"), stepInfo.Latest)
	}

	if !isShort {
		fmt.Printf("%s: %s\n", colorstring.Blue("source"), stepInfo.Source)
		fmt.Printf("%s:\n", colorstring.Blue("description"))
		fmt.Printf("%s\n", stepInfo.Description)
		fmt.Println()

		if len(stepInfo.Inputs) > 0 {
			fmt.Printf("%s:\n", colorstring.Blue("inputs"))
			for _, input := range stepInfo.Inputs {
				printRawEnvInfo(input)
			}
		}

		if len(stepInfo.Outputs) > 0 {
			if len(stepInfo.Inputs) > 0 {
				fmt.Println()
			}
			fmt.Printf("%s:\n", colorstring.Blue("outputs"))
			for _, output := range stepInfo.Outputs {
				printRawEnvInfo(output)
			}
		}
	}

	fmt.Println()
	fmt.Println()
}

func printJSONStepInfo(stepInfo models.StepInfoModel, isShort bool) error {
	bytes, err := json.Marshal(stepInfo)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))
	return nil
}

func printStepInfo(stepInfo models.StepInfoModel, format string, isShort, isLocal bool) error {
	switch format {
	case OutputFormatRaw:
		printRawStepInfo(stepInfo, isShort, isLocal)
		break
	case OutputFormatJSON:
		if err := printJSONStepInfo(stepInfo, isShort); err != nil {
			return err
		}
		break
	default:
		return fmt.Errorf("Invalid format: %s", format)
	}
	return nil
}

func stepInfo(c *cli.Context) error {
	// Input validation
	format := c.String(FormatKey)
	collectionURI := c.String(CollectionKey)
	YMLPath := c.String(StepYMLKey)
	isShort := c.Bool(ShortKey)
	id := c.String(IDKey)
	version := c.String(VersionKey)

	if format == "" {
		format = OutputFormatRaw
	} else if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		return fmt.Errorf("Invalid output format: %s", format)
	}

	if YMLPath == "" && collectionURI == "" {
		return fmt.Errorf("Missing required input: no StepLib, nor step.yml path defined as step info source")
	}

	if YMLPath != "" {
		//
		// Local step info
		step, err := stepman.ParseStepYml(YMLPath, false)
		if err != nil {
			return fmt.Errorf("Failed to parse step.yml (path:%s), err: %s", YMLPath, err)
		}

		inputs, err := getEnvInfos(step.Inputs)
		if err != nil {
			return fmt.Errorf("Failed to get step (path:%s) input infos, err: %s", YMLPath, err)
		}

		outputs, err := getEnvInfos(step.Outputs)
		if err != nil {
			return fmt.Errorf("Failed to get step (path:%s) output infos, err: %s", YMLPath, err)
		}

		stepInfo := models.StepInfoModel{
			StepLib:     YMLPath,
			Description: *step.Description,
			Source:      *step.SourceCodeURL,
			Inputs:      inputs,
			Outputs:     outputs,
		}

		if err := printStepInfo(stepInfo, format, isShort, true); err != nil {
			return fmt.Errorf("Failed to print step info, err: %s", err)
		}
	} else {
		//
		// StepLib step info

		// Input validation
		if id == "" {
			return errors.New("Missing required input: step id")
		}

		// Check if setup was done for collection
		if exist, err := stepman.RootExistForCollection(collectionURI); err != nil {
			return fmt.Errorf("Failed to check if setup was done for steplib (%s), error: %s", collectionURI, err)
		} else if !exist {
			if err := setupSteplib(collectionURI, format != OutputFormatRaw); err != nil {
				return errors.New("Failed to setup steplib")
			}
		}

		// Check if step exist in collection
		collection, err := stepman.ReadStepSpec(collectionURI)
		if err != nil {
			return fmt.Errorf("Failed to read steps spec (spec.json), err: %s", err)
		}

		step, stepFound := collection.GetStep(id, version)
		if !stepFound {
			if version == "" {
				return fmt.Errorf("Collection doesn't contain any version of step (id:%s)", id)
			}
			return fmt.Errorf("Collection doesn't contain step (id:%s) (version:%s)", id, version)
		}

		latest, err := collection.GetLatestStepVersion(id)
		if err != nil {
			return fmt.Errorf("Failed to get latest version of step (id:%s)", id)
		}

		if version == "" {
			version = latest
		}

		inputs, err := getEnvInfos(step.Inputs)
		if err != nil {
			return fmt.Errorf("Failed to get step (id:%s) input infos, err: %s", id, err)
		}

		outputs, err := getEnvInfos(step.Outputs)
		if err != nil {
			return fmt.Errorf("Failed to get step (id:%s) output infos, err: %s", id, err)
		}

		stepInfo := models.StepInfoModel{
			ID:          id,
			Version:     version,
			Latest:      latest,
			Description: *step.Description,
			StepLib:     collectionURI,
			Source:      *step.SourceCodeURL,
			Inputs:      inputs,
			Outputs:     outputs,
		}

		route, found := stepman.ReadRoute(collectionURI)
		if !found {
			return fmt.Errorf("No route found for collection: %s", collectionURI)
		}
		globalStepInfoPth := stepman.GetStepGlobalInfoPath(route, id)
		if globalStepInfoPth != "" {
			globalInfo, found, err := stepman.ParseGlobalStepInfoYML(globalStepInfoPth)
			if err != nil {
				return fmt.Errorf("Failed to get step (path:%s) output infos, err: %s", YMLPath, err)
			}

			if found {
				stepInfo.GlobalInfo = globalInfo
			}
		}

		if err := printStepInfo(stepInfo, format, isShort, false); err != nil {
			return fmt.Errorf("Failed to print step info, err: %s", err)
		}
	}

	return nil
}
