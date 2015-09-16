package cli

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

// EnvInfoModel ...
type EnvInfoModel struct {
	Env         string `json:"env,omitempty" yaml:"env,omitempty"`
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// StepInfoModel ...
type StepInfoModel struct {
	ID          string         `json:"step_id,omitempty" yaml:"step_id,omitempty"`
	Version     string         `json:"step_version,omitempty" yaml:"step_version,omitempty"`
	Latest      string         `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	Source      string         `json:"source,omitempty" yaml:"source,omitempty"`
	StepLib     string         `json:"steplib,omitempty" yaml:"steplib,omitempty"`
	Inputs      []EnvInfoModel `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs     []EnvInfoModel `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

func printRawStepInfo(stepInfo StepInfoModel, isShort bool) error {
	fmt.Println(colorstring.Bluef("Step info in StepLib (%s):", stepInfo.StepLib))

	fmt.Printf("%s: %s\n", colorstring.Blue("ID"), stepInfo.ID)
	fmt.Printf("%s: %s\n", colorstring.Blue("version"), stepInfo.Version)
	fmt.Printf("%s: %s\n", colorstring.Blue("latest"), stepInfo.Latest)
	if !isShort {
		fmt.Printf("%s: %s\n", colorstring.Blue("source"), stepInfo.Source)
		fmt.Printf("%s:\n", colorstring.Blue("description"))
		fmt.Printf("%s\n", stepInfo.Description)
		fmt.Println()

		if len(stepInfo.Inputs) > 0 {
			fmt.Printf("%s:\n", colorstring.Blue("inputs"))
			for _, input := range stepInfo.Inputs {
				fmt.Printf("- %s\n", colorstring.Green(input.Env))
				if input.Description != "" {
					fmt.Printf("  %s:\n", colorstring.Green("description"))
					fmt.Printf("  %s\n", input.Description)
				}
			}
		}

		if len(stepInfo.Outputs) > 0 {
			fmt.Printf("%s:\n", colorstring.Blue("outputs"))
			for _, input := range stepInfo.Outputs {
				fmt.Printf("- %s\n", colorstring.Green(input.Env))
				if input.Description != "" {
					fmt.Printf("  %s:\n", colorstring.Green("description"))
					fmt.Printf("  %s\n", input.Description)
				}
			}
		}
	}
	fmt.Println()
	fmt.Println()
	return nil
}

func printJSONStepInfo(stepInfo StepInfoModel, isShort bool) error {
	bytes, err := json.Marshal(stepInfo)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))
	return nil
}

func getEnvInfos(envs []envmanModels.EnvironmentItemModel) ([]EnvInfoModel, error) {
	envInfos := []EnvInfoModel{}
	for _, env := range envs {
		key, _, err := env.GetKeyValuePair()
		if err != nil {
			return []EnvInfoModel{}, err
		}

		options, err := env.GetOptions()
		if err != nil {
			return []EnvInfoModel{}, err
		}

		envInfo := EnvInfoModel{
			Env:         key,
			Description: *options.Description,
		}
		envInfos = append(envInfos, envInfo)
	}
	return envInfos, nil
}

func stepInfo(c *cli.Context) {
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
		log.Fatal("[STEPMAN] - Missing step id")
	}

	format := c.String(FormatKey)
	if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		log.Fatalf("[STEPMAN] - Invalid format: %s", format)
	}

	version := c.String(VersionKey)
	isShort := c.Bool(ShortKey)

	for _, collectionURI := range collectionURIs {
		// Check if step exist in collection
		collection, err := stepman.ReadStepSpec(collectionURI)
		if err != nil {
			log.Fatalf("[STEPMAN] - Failed to read steps spec (spec.json), err: %s", err)
		}

		step, stepFound := collection.GetStep(id, version)
		if !stepFound {
			if version == "" {
				log.Fatalf("[STEPMAN] - Collection doesn't contain any version of step (id:%s)", id)
			} else {
				log.Fatalf("[STEPMAN] - Collection doesn't contain step (id:%s) (version:%s)", id, version)
			}
		}

		latest, err := collection.GetLatestStepVersion(id)
		if err != nil {
			log.Fatalf("[STEPMAN] - Failed to get latest version of step (id:%s)", id)
		}

		if version == "" {
			version = latest
		}

		inputs, err := getEnvInfos(step.Inputs)
		if err != nil {
			log.Fatalf("[STEPMAN] - Failed to get step (id:%s) input infos, err: %s", id, err)
		}

		outputs, err := getEnvInfos(step.Outputs)
		if err != nil {
			log.Fatalf("[STEPMAN] - Failed to get step (id:%s) output infos, err: %s", id, err)
		}

		stepInfo := StepInfoModel{
			ID:          id,
			Version:     version,
			Latest:      latest,
			Description: *step.Description,
			StepLib:     collectionURI,
			Source:      *step.SourceCodeURL,
			Inputs:      inputs,
			Outputs:     outputs,
		}

		switch format {
		case OutputFormatRaw:
			printRawStepInfo(stepInfo, isShort)
			break
		case OutputFormatJSON:
			printJSONStepInfo(stepInfo, isShort)
			break
		default:
			log.Fatalf("[STEPMAN] - Invalid format: %s", format)
		}
	}
}
