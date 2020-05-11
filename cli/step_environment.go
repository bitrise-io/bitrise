package cli

import (
	"fmt"

	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/envman/env"
	envmanModels "github.com/bitrise-io/envman/models"
)

type prepareStepInputParams struct {
	environment                 []envmanModels.EnvironmentItemModel
	inputs                      []envmanModels.EnvironmentItemModel
	buildRunResults             models.BuildRunResultsModel
	isCIMode, isPullRequestMode bool
}

func prepareStepEnvironment(params prepareStepInputParams) ([]envmanModels.EnvironmentItemModel, error) {
	// Expand templates
	evaluatedInputs := []envmanModels.EnvironmentItemModel{}
	for _, input := range params.inputs {
		key, value, err := input.GetKeyValuePair()
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("prepareStepEnvironment() failed to get input key: %s", err)
		}

		options, err := input.GetOptions()
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("prepareStepEnvironment() failed to get options: %s", err)
		}

		if options.IsTemplate != nil && *options.IsTemplate {
			envs, err := env.GetDeclarationsSideEffects(params.environment, &env.DefaultEnvironmentSource{})
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("GetDeclarationsSideEffects() failed, %s", err)
			}

			evaluatedValue, err := bitrise.EvaluateTemplateToString(value, params.isCIMode, params.isPullRequestMode, params.buildRunResults, envs.ResultEnvironment)
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, fmt.Errorf("prepareStepEnvironment() failed to evaluate template: %s", err)
			}

			input[key] = evaluatedValue
		}

		evaluatedInputs = append(evaluatedInputs, input)
	}

	stepEnvironment := append(params.environment, evaluatedInputs...)

	for _, stepEnv := range stepEnvironment {
		if err := stepEnv.FillMissingDefaults(); err != nil {
			return []envmanModels.EnvironmentItemModel{},
				fmt.Errorf("prepareStepEnvironment() failed to fill missing defaults: %s", err)
		}
	}

	return stepEnvironment, nil
}
