package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bitrise-io/envman/models"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func print(c *cli.Context) error {
	// Input validation
	format := c.String(FormatKey)
	if format == "" {
		format = OutputFormatRaw
	} else if !(format == OutputFormatRaw || format == OutputFormatJSON) {
		log.Fatalf("Invalid format: %s", format)
	}

	expand := c.Bool(ExpandKey)
	sensitiveOnly := c.Bool(SensitiveOnlyKey)

	// Read envs
	envSet, err := ReadEnvsJSONList(CurrentEnvStoreFilePath, expand, sensitiveOnly)
	if err != nil {
		log.Fatal(err)
	}

	// Print envs
	switch format {
	case OutputFormatRaw:
		printRawEnvs(envSet)
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
func ReadEnvsJSONList(envStorePth string, expand, sensitiveOnly bool) (models.EnvsJSONListModel, error) {
	// Read envs
	environments, err := ReadEnvs(envStorePth)
	if err != nil {
		return nil, fmt.Errorf("failed to read envs: %s", err)
	}

	envsJSONList, err := convertToEnvsJSONModel(environments, expand, sensitiveOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to convert envs: %s", err)
	}
	return envsJSONList, nil
}

func convertToEnvsJSONModel(envs []models.EnvironmentItemModel, expand, sensitiveOnly bool) (models.EnvsJSONListModel, error) {
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

		if sensitiveOnly {
			if opts.IsSensitive == nil || !*opts.IsSensitive {
				continue
			}
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

func expandEnvsInString(inp string) string {
	return os.ExpandEnv(inp)
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
