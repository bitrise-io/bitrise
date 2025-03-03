package cli

import (
	"encoding/json"
	"fmt"

	"github.com/bitrise-io/envman/env"
	"github.com/bitrise-io/envman/models"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func printCmd(c *cli.Context) error {
	// Input validation
	format := c.String(FormatKey)
	if format == "" {
		format = OutputFormatRaw
	} else if !(format == OutputFormatRaw || format == OutputFormatJSON || format == OutputFormatEnvList) {
		log.Fatalf("Invalid format: %s", format)
	}

	expand := c.Bool(ExpandKey)
	sensitiveOnly := c.Bool(SensitiveOnlyKey)

	// Read envs
	envSet, err := ReadEnvsJSONList(CurrentEnvStoreFilePath, expand, sensitiveOnly, &env.DefaultEnvironmentSource{})
	if err != nil {
		log.Fatal(err)
	}

	// Print envs
	switch format {
	case OutputFormatRaw:
		printRawEnvs(envSet)
	case OutputFormatEnvList:
		printEnvsList(envSet)
	case OutputFormatJSON:
		if err := printJSONEnvs(envSet); err != nil {
			log.Fatalf("Failed to print env list, err: %s", err)
		}
	default:
		log.Fatalf("[STEPMAN] - Invalid format: %s", format)
	}

	return nil
}

// ReadEnvsJSONList ...
func ReadEnvsJSONList(envStorePth string, expand, sensitiveOnly bool, envSource env.EnvironmentSource) (models.EnvsJSONListModel, error) {
	envs, err := ReadEnvs(envStorePth)
	if err != nil {
		return nil, fmt.Errorf("failed to read envs: %s", err)
	}

	return ConvertToEnvsJSONModel(envs, expand, sensitiveOnly, envSource)
}

func ConvertToEnvsJSONModel(envs []models.EnvironmentItemModel, expand, sensitiveOnly bool, envSource env.EnvironmentSource) (models.EnvsJSONListModel, error) {
	if sensitiveOnly {
		var err error
		envs, err = sensitiveEnvs(envs)
		if err != nil {
			return nil, fmt.Errorf("failed to filter sensitive envs: %s", err)
		}
	}

	var resultEnvs map[string]string
	if expand {
		result, err := env.GetDeclarationsSideEffects(envs, envSource)
		if err != nil {
			return nil, fmt.Errorf("failed to expand envs: %s", err)
		}
		resultEnvs = result.EvaluatedNewEnvs
	} else {
		resultEnvs = map[string]string{}
		for _, env := range envs {
			key, value, err := env.GetKeyValuePair()
			if err != nil {
				return nil, err
			}

			resultEnvs[key] = value
		}
	}

	return resultEnvs, nil
}

func sensitiveEnvs(envs []models.EnvironmentItemModel) ([]models.EnvironmentItemModel, error) {
	var filtered []models.EnvironmentItemModel
	for _, env := range envs {
		opts, err := env.GetOptions()
		if err != nil {
			return nil, err
		}

		if opts.IsSensitive != nil && *opts.IsSensitive {
			filtered = append(filtered, env)
		}
	}
	return filtered, nil
}

func printJSONEnvs(envList models.EnvsJSONListModel) error {
	bytes, err := json.Marshal(envList)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))
	return nil
}

func printRawEnvs(envList models.EnvsJSONListModel) {
	fmt.Println()
	for key, value := range envList {
		fmt.Printf("%s: %s\n", key, value)
	}
	fmt.Println()
}

func printEnvsList(envList models.EnvsJSONListModel) {
	fmt.Println()
	for key, value := range envList {
		fmt.Printf("export %s=%q\n", key, value)
	}
	fmt.Println()
}
