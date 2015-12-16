package cli

import (
	"encoding/json"
	"fmt"
	"path"

	log "github.com/Sirupsen/logrus"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
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
		fmt.Printf(colorstring.Red("This step is deprecated!\n"))
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

func stepInfo(c *cli.Context) {
	// Input validation
	format := c.String(FormatKey)
	if format == "" {
		format = OutputFormatRaw
	} else if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		log.Fatalf("Invalid format: %s", format)
	}

	isShort := c.Bool(ShortKey)

	YMLPath := c.String(StepYMLKey)
	if YMLPath != "" {
		//
		// Local step info
		step, err := stepman.ParseStepYml(YMLPath, false)
		if err != nil {
			log.Fatalf("Failed to parse step.yml, err: %s", err)
		}

		inputs, err := getEnvInfos(step.Inputs)
		if err != nil {
			log.Fatalf("Failed to get step (path:%s) input infos, err: %s", YMLPath, err)
		}

		outputs, err := getEnvInfos(step.Outputs)
		if err != nil {
			log.Fatalf("Failed to get step (path:%s) output infos, err: %s", YMLPath, err)
		}

		stepInfo := models.StepInfoModel{
			StepLib:     YMLPath,
			Description: *step.Description,
			Source:      *step.SourceCodeURL,
			Inputs:      inputs,
			Outputs:     outputs,
		}

		dir := path.Dir(YMLPath)
		globalStepInfoPth := path.Join(dir, "step-info.yml")
		globalInfo, found, err := stepman.ParseGlobalStepInfoYML(globalStepInfoPth)
		if err != nil {
			log.Fatalf("Failed to get step (path:%s) output infos, err: %s", YMLPath, err)
		}

		if found {
			stepInfo.GlobalInfo = globalInfo
		}

		if err := printStepInfo(stepInfo, format, isShort, true); err != nil {
			log.Fatalf("Failed to print step info, err: %s", err)
		}
	} else {
		//
		// StepLib step info

		// Input validation
		collectionURIs := []string{}
		collectionURI := c.String(CollectionKey)
		if collectionURI == "" {
			collectionURIs = stepman.GetAllStepCollectionPath()
		} else {
			collectionURIs = []string{collectionURI}
		}

		id := c.String(IDKey)
		if id == "" {
			log.Fatal("Missing step id")
		}

		version := c.String(VersionKey)

		for _, collectionURI := range collectionURIs {
			// Check if step exist in collection
			collection, err := stepman.ReadStepSpec(collectionURI)
			if err != nil {
				log.Fatalf("Failed to read steps spec (spec.json), err: %s", err)
			}

			step, stepFound := collection.GetStep(id, version)
			if !stepFound {
				if version == "" {
					log.Fatalf("Collection doesn't contain any version of step (id:%s)", id)
				} else {
					log.Fatalf("Collection doesn't contain step (id:%s) (version:%s)", id, version)
				}
			}

			latest, err := collection.GetLatestStepVersion(id)
			if err != nil {
				log.Fatalf("Failed to get latest version of step (id:%s)", id)
			}

			if version == "" {
				version = latest
			}

			inputs, err := getEnvInfos(step.Inputs)
			if err != nil {
				log.Fatalf("Failed to get step (id:%s) input infos, err: %s", id, err)
			}

			outputs, err := getEnvInfos(step.Outputs)
			if err != nil {
				log.Fatalf("Failed to get step (id:%s) output infos, err: %s", id, err)
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
				log.Fatalf("No route found for collection: %s", collectionURI)
			}
			globalStepInfoPth := stepman.GetStepGlobalInfoPath(route, id)
			if globalStepInfoPth != "" {
				globalInfo, found, err := stepman.ParseGlobalStepInfoYML(globalStepInfoPth)
				if err != nil {
					log.Fatalf("Failed to get step (path:%s) output infos, err: %s", YMLPath, err)
				}

				if found {
					stepInfo.GlobalInfo = globalInfo
				}
			}

			if err := printStepInfo(stepInfo, format, isShort, false); err != nil {
				log.Fatalf("Failed to print step info, err: %s", err)
			}
		}
	}
}
