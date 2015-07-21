package models

import (
	"errors"

	stepmanModels "github.com/bitrise-io/stepman/models"
)

// MergeEnvironmentWith ...
func MergeEnvironmentWith(env *stepmanModels.EnvironmentItemModel, otherEnv stepmanModels.EnvironmentItemModel) error {
	// merge key-value
	key, _, err := env.GetKeyValuePair()
	if err != nil {
		return err
	}

	otherKey, otherValue, err := otherEnv.GetKeyValuePair()
	if err != nil {
		return err
	}

	if otherKey != key {
		return errors.New("Env keys are diferent")
	}

	(*env)[key] = otherValue

	//merge options
	options, err := env.GetOptions()
	if err != nil {
		return err
	}

	otherOptions, err := otherEnv.GetOptions()
	if err != nil {
		return err
	}

	if otherOptions.Title != nil {
		*options.Title = *otherOptions.Title
	}
	if otherOptions.Description != nil {
		*options.Description = *otherOptions.Description
	}
	if len(otherOptions.ValueOptions) > 0 {
		options.ValueOptions = otherOptions.ValueOptions
	}
	if otherOptions.IsRequired != nil {
		*options.IsRequired = *otherOptions.IsRequired
	}
	if otherOptions.IsExpand != nil {
		*options.IsExpand = *otherOptions.IsExpand
	}
	if otherOptions.IsDontChangeValue != nil {
		*options.IsDontChangeValue = *otherOptions.IsDontChangeValue
	}
	return nil
}

// MergeStepWith ...
func MergeStepWith(step, otherStep stepmanModels.StepModel) error {
	if otherStep.Title != nil {
		*step.Title = *otherStep.Title
	}
	if otherStep.Description != nil {
		*step.Description = *otherStep.Description
	}
	if otherStep.Summary != nil {
		*step.Summary = *otherStep.Summary
	}
	if otherStep.Website != nil {
		*step.Website = *otherStep.Website
	}
	if otherStep.SourceCodeURL != nil {
		*step.SourceCodeURL = *otherStep.SourceCodeURL
	}
	if otherStep.SupportURL != nil {
		*step.SupportURL = *otherStep.SupportURL
	}
	if otherStep.Source.Git != nil {
		*step.Source.Git = *otherStep.Source.Git
	}
	if len(otherStep.HostOsTags) > 0 {
		step.HostOsTags = otherStep.HostOsTags
	}
	if len(otherStep.ProjectTypeTags) > 0 {
		step.ProjectTypeTags = otherStep.ProjectTypeTags
	}
	if len(otherStep.TypeTags) > 0 {
		step.TypeTags = otherStep.TypeTags
	}
	if otherStep.IsRequiresAdminUser != nil {
		*step.IsRequiresAdminUser = *otherStep.IsRequiresAdminUser
	}
	if otherStep.IsAlwaysRun != nil {
		*step.IsAlwaysRun = *otherStep.IsAlwaysRun
	}
	if otherStep.IsNotImportant != nil {
		*step.IsNotImportant = *otherStep.IsNotImportant
	}

	for _, input := range step.Inputs {
		key, _, err := input.GetKeyValuePair()
		if err != nil {
			return err
		}
		otherInput, found := getInputByKey(otherStep, key)
		if found {
			err := MergeEnvironmentWith(&input, otherInput)
			if err != nil {
				return err
			}
		}
	}

	for _, output := range step.Outputs {
		key, _, err := output.GetKeyValuePair()
		if err != nil {
			return err
		}
		otherOutput, found := getOutputByKey(otherStep, key)
		if found {
			err := MergeEnvironmentWith(&output, otherOutput)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getInputByKey(step stepmanModels.StepModel, key string) (stepmanModels.EnvironmentItemModel, bool) {
	for _, input := range step.Inputs {
		k, _, err := input.GetKeyValuePair()
		if err != nil {
			return stepmanModels.EnvironmentItemModel{}, false
		}

		if k == key {
			return input, true
		}
	}
	return stepmanModels.EnvironmentItemModel{}, false
}

func getOutputByKey(step stepmanModels.StepModel, key string) (stepmanModels.EnvironmentItemModel, bool) {
	for _, output := range step.Outputs {
		k, _, err := output.GetKeyValuePair()
		if err != nil {
			return stepmanModels.EnvironmentItemModel{}, false
		}

		if k == key {
			return output, true
		}
	}
	return stepmanModels.EnvironmentItemModel{}, false
}
