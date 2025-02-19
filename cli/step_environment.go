package cli

import (
	"fmt"

	"github.com/bitrise-io/bitrise/v2/bitrise"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/envman/env"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/parseutil"
)

type prepareStepInputParams struct {
	environment                 []envmanModels.EnvironmentItemModel
	inputs                      []envmanModels.EnvironmentItemModel
	buildRunResults             models.BuildRunResultsModel
	isCIMode, isPullRequestMode bool
}

func prepareStepEnvironment(params prepareStepInputParams, envSource env.EnvironmentSource) (stepEnvironment []envmanModels.EnvironmentItemModel, resultEnvironment map[string]string, nonStringInputs map[string]interface{}, err error) {
	for _, envVar := range params.environment {
		if err := envVar.Normalize(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, map[string]string{}, map[string]interface{}{},
				fmt.Errorf("failed to normalize declared environment variable: %s", err)
		}

		if err := envVar.FillMissingDefaults(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, map[string]string{}, map[string]interface{}{},
				fmt.Errorf("failed to fill missing declared environment variable defaults: %s", err)
		}
	}

	// Expand templates
	var evaluatedInputs []envmanModels.EnvironmentItemModel
	nonStringInputs = map[string]interface{}{}
	for _, input := range params.inputs {
		if err := input.FillMissingDefaults(); err != nil {
			return []envmanModels.EnvironmentItemModel{}, map[string]string{}, map[string]interface{}{},
				fmt.Errorf("failed to fill input missing default properties: %s", err)
		}

		key, value, err := input.GetKeyValuePairWithType()
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, map[string]string{}, map[string]interface{}{},
				fmt.Errorf("failed to get input key: %s", err)
		}

		options, err := input.GetOptions()
		if err != nil {
			return []envmanModels.EnvironmentItemModel{}, map[string]string{}, map[string]interface{}{},
				fmt.Errorf("failed to get options: %s", err)
		}

		if options.IsTemplate != nil && *options.IsTemplate {
			envs, err := env.GetDeclarationsSideEffects(params.environment, &env.DefaultEnvironmentSource{})
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, map[string]string{}, map[string]interface{}{},
					fmt.Errorf("GetDeclarationsSideEffects() failed, %s", err)
			}

			evaluatedValue, err := bitrise.EvaluateTemplateToString(parseutil.CastToString(value), params.isCIMode, params.isPullRequestMode, params.buildRunResults, envs.ResultEnvironment)
			if err != nil {
				return []envmanModels.EnvironmentItemModel{}, map[string]string{}, map[string]interface{}{},
					fmt.Errorf("failed to evaluate template: %s", err)
			}

			value = evaluatedValue
			input[key] = evaluatedValue
		}
		if _, ok := value.(string); !ok {
			nonStringInputs[key] = value
		}
		evaluatedInputs = append(evaluatedInputs, input)
	}

	stepEnvironment = append(params.environment, evaluatedInputs...)

	declarationSideEffects, err := env.GetDeclarationsSideEffects(stepEnvironment, envSource)

	if err != nil {
		return []envmanModels.EnvironmentItemModel{}, map[string]string{}, map[string]interface{}{},
			fmt.Errorf("failed to get environment variable declaration results: %s", err)
	}

	return stepEnvironment, declarationSideEffects.ResultEnvironment, nonStringInputs, nil
}

func getSensitiveEnvs(envList []envmanModels.EnvironmentItemModel, expandedStepEnvironment map[string]string) ([]envmanModels.EnvironmentItemModel, error) {
	var sensitiveKeys []string
	for _, env := range envList {
		opts, err := env.GetOptions()
		if err != nil {
			return nil, err
		}
		if opts.IsSensitive == nil || !*opts.IsSensitive {
			continue
		}
		key, _, err := env.GetKeyValuePair()
		if err != nil {
			return nil, err
		}

		sensitiveKeys = append(sensitiveKeys, key)
	}

	var sensitiveValues []envmanModels.EnvironmentItemModel
	for _, key := range sensitiveKeys {
		if value, ok := expandedStepEnvironment[key]; ok {
			sensitiveValues = append(sensitiveValues, envmanModels.EnvironmentItemModel{key: value})
		}
	}

	return sensitiveValues, nil
}
