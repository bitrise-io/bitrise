package cli

import (
	"fmt"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/tools"
	envmanModels "github.com/bitrise-io/envman/models"
)

type prepareStepInputParams struct {
	environment                 []envmanModels.EnvironmentItemModel
	inputs                      []envmanModels.EnvironmentItemModel
	buildRunResults             models.BuildRunResultsModel
	inputEnvstorePath           string
	isCIMode, isPullRequestMode bool
}

func prepareStepEnvironment(params prepareStepInputParams) ([]envmanModels.EnvironmentItemModel, error) {
	// Collect step inputs
	if err := tools.EnvmanInitAtPath(params.inputEnvstorePath); err != nil {
		return []envmanModels.EnvironmentItemModel{},
			fmt.Errorf("failed to init envman for the Step, error: %s", err)
	}

	if err := tools.ExportEnvironmentsList(params.inputEnvstorePath, params.environment); err != nil {
		return []envmanModels.EnvironmentItemModel{},
			fmt.Errorf("failed to export environment list for the Step, error: %s", err)
	}

	evaluatedInputs := []envmanModels.EnvironmentItemModel{}
	for _, input := range params.inputs {
		key, value, err := input.GetKeyValuePair()
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("failed to get input key: %s", err)
		}

		options, err := input.GetOptions()
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("failed to get options: %s", err)
		}

		if options.IsTemplate != nil && *options.IsTemplate {
			outStr, err := tools.EnvmanJSONPrint(params.inputEnvstorePath)
			if err != nil {
				return []envmanModels.EnvironmentItemModel{},
					fmt.Errorf("EnvmanJSONPrint failed, err: %s", err)
			}

			envList, err := envmanModels.NewEnvJSONList(outStr)
			if err != nil {
				return []envmanModels.EnvironmentItemModel{},
					fmt.Errorf("NewEnvJSONList failed, err: %s", err)
			}

			evaluatedValue, err := bitrise.EvaluateTemplateToString(value, params.isCIMode, params.isPullRequestMode, params.buildRunResults, envList)
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("failed to evaluate template: %s", err)
			}

			input[key] = evaluatedValue
		}

		evaluatedInputs = append(evaluatedInputs, input)
	}

	return append(params.environment, evaluatedInputs...), nil
}
