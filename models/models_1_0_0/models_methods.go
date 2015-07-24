package models

import (
	"errors"
	"strings"

	"github.com/bitrise-io/go-pathutil/pathutil"
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

// GetStepIDStepDataPair ...
func GetStepIDStepDataPair(stepListItm StepListItemModel) (string, stepmanModels.StepModel, error) {
	if len(stepListItm) > 1 {
		return "", stepmanModels.StepModel{}, errors.New("StepListItem contains more than 1 key-value pair!")
	}
	for key, value := range stepListItm {
		return key, value, nil
	}
	return "", stepmanModels.StepModel{}, errors.New("StepListItem does not contain a key-value pair!")
}

// path::~/develop/steps-xcode-builder
// https://bitbucket.org/bitrise-team/bitrise-new-steps-spec::script@2.0.0
// script@2.0.0
// script

// CreateStepIDDataFromString ...
// CreateStepIDDataFromString ...
func CreateStepIDDataFromString(compositeVersionStr, defaultStepLibSource string) (StepIDData, error) {
	stepSrc := ""
	stepIDAndVersionOrURIStr := ""
	libsourceStepSplits := strings.Split(compositeVersionStr, "::")
	if len(libsourceStepSplits) == 2 {
		// long/verbose ID mode, ex: step-lib-src::step-id@1.0.0
		stepSrc = libsourceStepSplits[0]
		stepIDAndVersionOrURIStr = libsourceStepSplits[1]
	} else if len(libsourceStepSplits) == 1 {
		// missing steplib-src mode, ex: step-id@1.0.0
		//  in this case if we have a default StepLibSource we'll use that
		stepIDAndVersionOrURIStr = libsourceStepSplits[0]
	} else {
		return StepIDData{}, errors.New("No StepLib found, neither default provided (" + compositeVersionStr + ")")
	}

	stepIDOrURI := ""
	stepVersion := ""
	stepidVersionOrURISplits := strings.Split(stepIDAndVersionOrURIStr, "@")
	if len(stepidVersionOrURISplits) == 2 {
		stepIDOrURI = stepidVersionOrURISplits[0]
		stepVersion = stepidVersionOrURISplits[1]
	} else if len(stepidVersionOrURISplits) == 1 {
		stepIDOrURI = stepidVersionOrURISplits[0]
	} else {
		return StepIDData{}, errors.New("Step ID and version should be separated with a '@' separator (" + stepIDAndVersionOrURIStr + ")")
	}

	if stepIDOrURI == "" {
		return StepIDData{}, errors.New("No ID found at all (" + compositeVersionStr + ")")
	}

	if stepSrc == "" {
		if defaultStepLibSource == "" {
			return StepIDData{}, errors.New("No default StepLib source, in this case the composite ID should contain the source, separated with a '::' separator from the step ID (" + compositeVersionStr + ")")
		}
		stepSrc = defaultStepLibSource
	} else if stepSrc == "path" {
		uri, err := pathutil.AbsPath(stepIDOrURI)
		if err != nil {
			return StepIDData{}, err
		}
		stepIDOrURI = uri
	}

	return StepIDData{
		SteplibSource: stepSrc,
		ID:            stepIDOrURI,
		Version:       stepVersion,
	}, nil
}
