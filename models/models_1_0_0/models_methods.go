package models

import (
	"errors"
	"fmt"
	"strings"

	envmanModels "github.com/bitrise-io/envman/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// Normalize ...
func (config *BitriseDataModel) Normalize() error {
	for _, workflow := range config.Workflows {
		if err := workflow.Normalize(); err != nil {
			return err
		}
	}
	for _, env := range config.App.Environments {
		if err := env.Normalize(); err != nil {
			return err
		}
	}
	return nil
}

func containsWorkflowName(title string, workflowStack []string) bool {
	for _, t := range workflowStack {
		if t == title {
			return true
		}
	}
	return false
}

func removeWorkflowName(title string, workflowStack []string) []string {
	newStack := []string{}
	for _, t := range workflowStack {
		if t != title {
			newStack = append(newStack, t)
		}
	}
	return newStack
}

func checkWorkflowReferenceCycle(workflowID string, workflow WorkflowModel, bitriseConfig BitriseDataModel, workflowStack []string) error {
	if containsWorkflowName(workflowID, workflowStack) {
		stackStr := ""
		for _, aWorkflowID := range workflowStack {
			stackStr += aWorkflowID + " -> "
		}
		stackStr += workflowID
		return fmt.Errorf("Workflow reference cycle found: %s", stackStr)
	}
	workflowStack = append(workflowStack, workflowID)

	for _, beforeWorkflowName := range workflow.BeforeRun {
		beforeWorkflow, exist := bitriseConfig.Workflows[beforeWorkflowName]
		if !exist {
			return errors.New("Workflow not exist wit name " + beforeWorkflowName)
		}

		err := checkWorkflowReferenceCycle(beforeWorkflowName, beforeWorkflow, bitriseConfig, workflowStack)
		if err != nil {
			return err
		}
	}

	for _, afterWorkflowName := range workflow.AfterRun {
		afterWorkflow, exist := bitriseConfig.Workflows[afterWorkflowName]
		if !exist {
			return errors.New("Workflow not exist wit name " + afterWorkflowName)
		}

		err := checkWorkflowReferenceCycle(afterWorkflowName, afterWorkflow, bitriseConfig, workflowStack)
		if err != nil {
			return err
		}
	}

	workflowStack = removeWorkflowName(workflowID, workflowStack)

	return nil
}

// Validate ...
func (config *BitriseDataModel) Validate() error {
	for ID, workflow := range config.Workflows {
		if err := workflow.Validate(ID); err != nil {
			return err
		}
		if err := checkWorkflowReferenceCycle(ID, workflow, *config, []string{}); err != nil {
			return err
		}
	}
	return nil
}

// FillMissingDefaults ...
func (config *BitriseDataModel) FillMissingDefaults() error {
	for title, workflow := range config.Workflows {
		if err := workflow.FillMissingDefaults(title); err != nil {
			return err
		}
	}
	for _, env := range config.App.Environments {
		if err := env.FillMissingDefaults(); err != nil {
			return err
		}
	}
	return nil
}

// Normalize ...
func (workflow *WorkflowModel) Normalize() error {
	for _, env := range workflow.Environments {
		if err := env.Normalize(); err != nil {
			return err
		}
	}
	return nil
}

// FillMissingDefaults ...
func (workflow *WorkflowModel) FillMissingDefaults(title string) error {
	for _, env := range workflow.Environments {
		if err := env.FillMissingDefaults(); err != nil {
			return err
		}
	}
	if workflow.Title == "" {
		workflow.Title = title
	}
	return nil
}

// Validate ...
func (workflow *WorkflowModel) Validate(title string) error {
	// Validate envs
	for _, env := range workflow.Environments {
		if err := env.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// MergeEnvironmentWith ...
func MergeEnvironmentWith(env *envmanModels.EnvironmentItemModel, otherEnv envmanModels.EnvironmentItemModel) error {
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
		if options.Title == nil {
			options.Title = new(string)
		}
		*options.Title = *otherOptions.Title
	}
	if otherOptions.Description != nil {
		if options.Description == nil {
			options.Description = new(string)
		}
		*options.Description = *otherOptions.Description
	}
	if len(otherOptions.ValueOptions) > 0 {
		options.ValueOptions = otherOptions.ValueOptions
	}
	if otherOptions.IsRequired != nil {
		if options.IsRequired == nil {
			options.IsRequired = new(bool)
		}
		*options.IsRequired = *otherOptions.IsRequired
	}
	if otherOptions.IsExpand != nil {
		if options.IsExpand == nil {
			options.IsExpand = new(bool)
		}
		*options.IsExpand = *otherOptions.IsExpand
	}
	if otherOptions.IsDontChangeValue != nil {
		if options.IsDontChangeValue == nil {
			options.IsDontChangeValue = new(bool)
		}
		*options.IsDontChangeValue = *otherOptions.IsDontChangeValue
	}
	(*env)[envmanModels.OptionsKey] = options
	return nil
}

// MergeStepWith ...
func MergeStepWith(step, otherStep stepmanModels.StepModel) (stepmanModels.StepModel, error) {
	if err := step.FillMissingDefaults(); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if otherStep.Title != nil {
		if step.Title == nil {
			step.Title = new(string)
		}
		*step.Title = *otherStep.Title
	}
	if otherStep.Description != nil {
		if step.Description == nil {
			step.Description = new(string)
		}
		*step.Description = *otherStep.Description
	}
	if otherStep.Summary != nil {
		if step.Summary == nil {
			step.Summary = new(string)
		}
		*step.Summary = *otherStep.Summary
	}
	if otherStep.Website != nil {
		if step.Website == nil {
			step.Website = new(string)
		}
		*step.Website = *otherStep.Website
	}
	if otherStep.SourceCodeURL != nil {
		if step.SourceCodeURL == nil {
			step.SourceCodeURL = new(string)
		}
		*step.SourceCodeURL = *otherStep.SourceCodeURL
	}
	if otherStep.SupportURL != nil {
		if step.SupportURL == nil {
			step.SupportURL = new(string)
		}
		*step.SupportURL = *otherStep.SupportURL
	}
	if otherStep.Source.Git != "" {
		step.Source.Git = otherStep.Source.Git
	}
	if otherStep.Source.Commit != "" {
		step.Source.Commit = otherStep.Source.Commit
	}
	if len(otherStep.Dependencies) > 0 {
		step.Dependencies = otherStep.Dependencies
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
		if step.IsRequiresAdminUser == nil {
			step.IsRequiresAdminUser = new(bool)
		}
		*step.IsRequiresAdminUser = *otherStep.IsRequiresAdminUser
	}
	if otherStep.IsAlwaysRun != nil {
		if step.IsAlwaysRun == nil {
			step.IsAlwaysRun = new(bool)
		}
		*step.IsAlwaysRun = *otherStep.IsAlwaysRun
	}
	if otherStep.IsSkippable != nil {
		if step.IsSkippable == nil {
			step.IsSkippable = new(bool)
		}
		*step.IsSkippable = *otherStep.IsSkippable
	}
	if otherStep.RunIf != nil {
		if step.RunIf == nil {
			step.RunIf = new(string)
		}
		*step.RunIf = *otherStep.RunIf
	}

	for _, input := range step.Inputs {
		key, _, err := input.GetKeyValuePair()
		if err != nil {
			return stepmanModels.StepModel{}, err
		}
		otherInput, found := getInputByKey(otherStep, key)
		if found {
			err := MergeEnvironmentWith(&input, otherInput)
			if err != nil {
				return stepmanModels.StepModel{}, err
			}
		}
	}

	for _, output := range step.Outputs {
		key, _, err := output.GetKeyValuePair()
		if err != nil {
			return stepmanModels.StepModel{}, err
		}
		otherOutput, found := getOutputByKey(otherStep, key)
		if found {
			err := MergeEnvironmentWith(&output, otherOutput)
			if err != nil {
				return stepmanModels.StepModel{}, err
			}
		}
	}

	return step, nil
}

func getInputByKey(step stepmanModels.StepModel, key string) (envmanModels.EnvironmentItemModel, bool) {
	for _, input := range step.Inputs {
		k, _, err := input.GetKeyValuePair()
		if err != nil {
			return envmanModels.EnvironmentItemModel{}, false
		}

		if k == key {
			return input, true
		}
	}
	return envmanModels.EnvironmentItemModel{}, false
}

func getOutputByKey(step stepmanModels.StepModel, key string) (envmanModels.EnvironmentItemModel, bool) {
	for _, output := range step.Outputs {
		k, _, err := output.GetKeyValuePair()
		if err != nil {
			return envmanModels.EnvironmentItemModel{}, false
		}

		if k == key {
			return output, true
		}
	}
	return envmanModels.EnvironmentItemModel{}, false
}

// GetStepIDStepDataPair ...
func GetStepIDStepDataPair(stepListItem StepListItemModel) (string, stepmanModels.StepModel, error) {
	if len(stepListItem) > 1 {
		return "", stepmanModels.StepModel{}, errors.New("StepListItem contains more than 1 key-value pair!")
	}
	for key, value := range stepListItem {
		return key, value, nil
	}
	return "", stepmanModels.StepModel{}, errors.New("StepListItem does not contain a key-value pair!")
}

// CreateStepIDDataFromString ...
// compositeVersionStr examples:
//  * local path:
//    * path::~/path/to/step/dir
//  * direct git url and branch or tag:
//    * git::https://github.com/bitrise-io/steps-timestamp.git@master
//  * full ID with steplib, stepid and version:
//    * https://github.com/bitrise-io/bitrise-steplib::script@2.0.0
//  * only stepid and version (requires a default steplib source to be provided):
//    * script@2.0.0
//  * only stepid, latest version will be used (requires a default steplib source to be provided):
//    * script
func CreateStepIDDataFromString(compositeVersionStr, defaultStepLibSource string) (StepIDData, error) {
	// first, determine the steplib-source/type
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

	if stepSrc == "" {
		if defaultStepLibSource == "" {
			return StepIDData{}, errors.New("No default StepLib source, in this case the composite ID should contain the source, separated with a '::' separator from the step ID (" + compositeVersionStr + ")")
		}
		stepSrc = defaultStepLibSource
	}

	// now determine the ID-or-URI and the version (if provided)
	stepIDOrURI := ""
	stepVersion := ""
	stepidVersionOrURISplits := strings.Split(stepIDAndVersionOrURIStr, "@")
	if len(stepidVersionOrURISplits) >= 2 {
		splitsCnt := len(stepidVersionOrURISplits)
		allButLastSplits := stepidVersionOrURISplits[:splitsCnt-1]
		// the ID or URI is all components except the last @version component
		//  which will be the version itself
		// for example in case it's a git direct URI like:
		//  git@github.com:bitrise-io/steps-timestamp.git@develop
		// which contains 2 at (@) signs only the last should be the version,
		//  the first one is part of the URI
		stepIDOrURI = strings.Join(allButLastSplits, "@")
		// version is simply the last component
		stepVersion = stepidVersionOrURISplits[splitsCnt-1]
	} else if len(stepidVersionOrURISplits) == 1 {
		stepIDOrURI = stepidVersionOrURISplits[0]
	} else {
		return StepIDData{}, errors.New("Step ID and version should be separated with a '@' separator (" + stepIDAndVersionOrURIStr + ")")
	}

	if stepIDOrURI == "" {
		return StepIDData{}, errors.New("No ID found at all (" + compositeVersionStr + ")")
	}

	return StepIDData{
		SteplibSource: stepSrc,
		IDorURI:       stepIDOrURI,
		Version:       stepVersion,
	}, nil
}

// IsBuildFailed ...
func (buildRes BuildRunResultsModel) IsBuildFailed() bool {
	return len(buildRes.FailedSteps) > 0
}

// Append ...
func (buildRes *BuildRunResultsModel) Append(res BuildRunResultsModel) {
	for _, success := range res.SuccessSteps {
		buildRes.SuccessSteps = append(buildRes.SuccessSteps, success)
	}
	for _, failed := range res.FailedSteps {
		buildRes.FailedSteps = append(buildRes.FailedSteps, failed)
	}
	for _, notImportant := range res.FailedNotImportantSteps {
		buildRes.FailedNotImportantSteps = append(buildRes.FailedNotImportantSteps, notImportant)
	}
	for _, skipped := range res.SkippedSteps {
		buildRes.SkippedSteps = append(buildRes.SkippedSteps, skipped)
	}
}
