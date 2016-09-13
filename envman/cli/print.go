package cli

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/envman/models"
	"github.com/urfave/cli"
)

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

func convertToEnsJSONModel(envs []models.EnvironmentItemModel, expand bool) (models.EnvsJSONListModel, error) {
	JSONModels := models.EnvsJSONListModel{}
	for _, env := range envs {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return models.EnvsJSONListModel{}, err
		}

		opts, err := env.GetOptions()
		if err != nil {
			return models.EnvsJSONListModel{}, err
		}

		if expand && (opts.IsExpand != nil && *opts.IsExpand) {
			value = expandEnvsInString(value)
		}

		JSONModels[key] = value

		if err := os.Setenv(key, value); err != nil {
			return models.EnvsJSONListModel{}, err
		}
	}
	return JSONModels, nil
}

func print(c *cli.Context) error {
	// Input validation
	format := c.String(FormatKey)
	if format == "" {
		format = OutputFormatRaw
	} else if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		log.Fatalf("Invalid format: %s", format)
	}

	expand := c.Bool(ExpandKey)

	// Read envs
	environments, err := envman.ReadEnvs(envman.CurrentEnvStoreFilePath)
	if err != nil {
		log.Fatalf("Failed to read envs, error: %s", err)
	}

	envsJSONList, err := convertToEnsJSONModel(environments, expand)
	if err != nil {
		log.Fatalf("Failed to convert envs, error: %s", err)
	}

	// Print envs
	switch format {
	case OutputFormatRaw:
		printRawEnvs(envsJSONList)
		break
	case OutputFormatJSON:
		if err := printJSONEnvs(envsJSONList); err != nil {
			log.Fatalf("Failed to print env list, err: %s", err)
		}
		break
	default:
		log.Fatalf("[STEPMAN] - Invalid format: %s", format)
	}

	return nil
}
