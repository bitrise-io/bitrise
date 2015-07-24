package models

import (
	"errors"
	"fmt"
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
func CreateStepIDDataFromString(compositeVersionStr, defaultStepLibSource string) (StepIDData, error) {
	ID := ""
	version := ""
	source := ""
	path := ""

	components := strings.Split(compositeVersionStr, "::")
	if len(components) == 1 {
		idAndVersion := components[0]
		idAndVersionComponents := strings.Split(idAndVersion, "@")
		if len(idAndVersionComponents) == 1 {
			ID = idAndVersionComponents[0]
		} else if len(idAndVersionComponents) == 2 {
			ID = idAndVersionComponents[0]
			version = idAndVersionComponents[1]
		} else {
			return StepIDData{}, fmt.Errorf("Invalid idAndVersionComponents length (%d)", len(idAndVersionComponents))
		}

		if defaultStepLibSource == "" {
			return StepIDData{}, errors.New("No default StepLib source, in this case the composite ID should contain the source, separated with a '::' separator from the step ID (" + compositeVersionStr + ")")
		}
		source = defaultStepLibSource
	} else if len(components) == 2 {
		sourceStr := components[0]
		idAndVersion := components[1]
		idAndVersionComponents := strings.Split(idAndVersion, "@")
		if len(idAndVersionComponents) == 1 {
			ID = idAndVersionComponents[0]
		} else if len(idAndVersionComponents) == 2 {
			ID = idAndVersionComponents[0]
			version = idAndVersionComponents[1]
		} else {
			return StepIDData{}, fmt.Errorf("Invalid idAndVersionComponents length (%d)", len(idAndVersionComponents))
		}

		if sourceStr == "path" {
			pth, err := pathutil.AbsPath(ID)
			if err != nil {
				return StepIDData{}, err
			}
			path = pth
		} else {
			source = sourceStr
		}
	} else {
		return StepIDData{}, fmt.Errorf("Invalid components length (%d)", len(components))
	}

	return StepIDData{
		ID:            ID,
		Version:       version,
		SteplibSource: source,
		LocalPath:     path,
	}, nil

	/*
		createIDVersionAndSourceComponents := func(IDVersionAndSource string) (idAndVersion string, source *string, e error) {
			components := strings.Split(IDVersionAndSource, "::")
			if len(components) == 1 {
				idAndVersion = components[0]
			} else if len(components) == 2 {
				sourceStr := components[0]
				source = &sourceStr
				idAndVersion = components[1]
			} else {
				return "", new(string), fmt.Errorf("Invalid components length (%d)", len(components))
			}
			return idAndVersion, source, nil
		}

		createIDAndVersion := func(idAndVersionStr string) (ID, version string, e error) {
			idAndVersionComponents := strings.Split(idAndVersionStr, "@")
			if len(idAndVersionComponents) == 1 {
				ID = idAndVersionComponents[0]
			} else if len(idAndVersionComponents) == 2 {
				ID = idAndVersionComponents[0]
				version = idAndVersionComponents[1]
			} else {
				return "", "", fmt.Errorf("Invalid idAndVersionComponents length (%d)", len(idAndVersionComponents))
			}
			return ID, version, nil
		}

		createSource := func(sourceStr *string, id string) (source, path string, e error) {
			if sourceStr == nil {
				if defaultStepLibSource == "" {
					return "", "", errors.New("No default StepLib source, in this case the composite ID should contain the source, separated with a '::' separator from the step ID (" + compositeVersionStr + ")")
				}
				source = defaultStepLibSource
			} else {
				if *sourceStr == "path" {
					path = id
				} else {
					source = *sourceStr
				}
			}
			return source, path, nil
		}

		stepIDAndVersionStr, stepSourceStr, err := createIDVersionAndSourceComponents(compositeVersionStr)
		if err != nil {
			return StepIDData{}, err
		}
		fmt.Println("stepIDAndVersionStr:", stepIDAndVersionStr)
		fmt.Println("stepSourceStr:", stepSourceStr)

		stepID, stepVersion, err := createIDAndVersion(stepIDAndVersionStr)
		if err != nil {
			return StepIDData{}, err
		}
		fmt.Println("stepID:", stepID)
		fmt.Println("stepVersion:", stepVersion)

		steplibSource, localPath, err := createSource(stepSourceStr, stepID)
		if err != nil {
			return StepIDData{}, err
		}
		fmt.Println("steplibSource:", steplibSource)
		fmt.Println("localPath:", localPath)

	*/
}
