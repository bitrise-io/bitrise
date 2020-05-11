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

func prepareStepEnvironment(params prepareStepInputParams, envSource env.EnvironmentSource) ([]envmanModels.EnvironmentItemModel, map[string]string, error) {
	initialEnvs := params.environment

	for _, envVar := range initialEnvs {
		if err := envVar.FillMissingDefaults(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, map[string]string{},
				fmt.Errorf("failed to fill missing defaults: %s", err)
		}
	}

	// Expand templates
	evaluatedInputs := []envmanModels.EnvironmentItemModel{}
	for _, input := range params.inputs {
		if err := input.FillMissingDefaults(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, map[string]string{},
				fmt.Errorf("failed to fill input missing default properties: %s", err)
		}

		key, value, err := input.GetKeyValuePair()
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, map[string]string{},
				fmt.Errorf("failed to get input key: %s", err)
		}

		options, err := input.GetOptions()
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, map[string]string{},
				fmt.Errorf("failed to get options: %s", err)
		}

		if options.IsTemplate != nil && *options.IsTemplate {
			envs, err := env.GetDeclarationsSideEffects(initialEnvs, &env.DefaultEnvironmentSource{})
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, map[string]string{},
					fmt.Errorf("GetDeclarationsSideEffects() failed, %s", err)
			}

			evaluatedValue, err := bitrise.EvaluateTemplateToString(value, params.isCIMode, params.isPullRequestMode, params.buildRunResults, envs.ResultEnvironment)
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, map[string]string{},
					fmt.Errorf("failed to evaluate template: %s", err)
			}

			input[key] = evaluatedValue
		}

		evaluatedInputs = append(evaluatedInputs, input)
	}

	stepEnvironment := append(params.environment, evaluatedInputs...)

	declarationSideEffects, err := env.GetDeclarationsSideEffects(stepEnvironment, envSource)
	if err != nil {
		return []envmanModels.EnvironmentItemModel{}, map[string]string{},
			fmt.Errorf("failed to get environment variable declaration results: %s", err)
	}

	return stepEnvironment, declarationSideEffects.ResultEnvironment, nil
}
