package models

import (
	"errors"
	"fmt"
	"strings"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// ----------------------------
// --- Normalize

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
			return errors.New("Workflow does not exist with name " + beforeWorkflowName)
		}

		err := checkWorkflowReferenceCycle(beforeWorkflowName, beforeWorkflow, bitriseConfig, workflowStack)
		if err != nil {
			return err
		}
	}

	for _, afterWorkflowName := range workflow.AfterRun {
		afterWorkflow, exist := bitriseConfig.Workflows[afterWorkflowName]
		if !exist {
			return errors.New("Workflow does not exist with name " + afterWorkflowName)
		}

		err := checkWorkflowReferenceCycle(afterWorkflowName, afterWorkflow, bitriseConfig, workflowStack)
		if err != nil {
			return err
		}
	}

	workflowStack = removeWorkflowName(workflowID, workflowStack)

	return nil
}

// Normalize ...
func (workflow *WorkflowModel) Normalize() error {
	for _, env := range workflow.Environments {
		if err := env.Normalize(); err != nil {
			return err
		}
	}
	for _, aWfStepItem := range workflow.Steps {
		_, stepData, err := GetStepIDStepDataPair(aWfStepItem)
		if err != nil {
			return err
		}
		if err := stepData.Normalize(); err != nil {
			return err
		}
	}
	return nil
}

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

// ----------------------------
// --- Validate

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

// ----------------------------
// --- FillMissingDefaults

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

// ----------------------------
// --- RemoveRedundantFields

func removeEnvironmentRedundantFields(env *envmanModels.EnvironmentItemModel) error {
	options, err := env.GetOptions()
	if err != nil {
		return err
	}

	hasOptions := false

	if options.Title != nil {
		if *options.Title == "" {
			options.Title = nil
		} else {
			hasOptions = true
		}
	}
	if options.Description != nil {
		if *options.Description == "" {
			options.Description = nil
		} else {
			hasOptions = true
		}
	}
	if options.Summary != nil {
		if *options.Summary == "" {
			options.Summary = nil
		} else {
			hasOptions = true
		}
	}
	if options.IsRequired != nil {
		if *options.IsRequired == envmanModels.DefaultIsRequired {
			options.IsRequired = nil
		} else {
			hasOptions = true
		}
	}
	if options.IsExpand != nil {
		if *options.IsExpand == envmanModels.DefaultIsExpand {
			options.IsExpand = nil
		} else {
			hasOptions = true
		}
	}
	if options.IsDontChangeValue != nil {
		if *options.IsDontChangeValue == envmanModels.DefaultIsDontChangeValue {
			options.IsDontChangeValue = nil
		} else {
			hasOptions = true
		}
	}

	if hasOptions {
		(*env)[envmanModels.OptionsKey] = options
	} else {
		delete(*env, envmanModels.OptionsKey)
	}

	return nil
}

func removeStepRedundantFields(step *stepmanModels.StepModel) error {
	if step.Title != nil && *step.Title == "" {
		step.Title = nil
	}
	if step.Description != nil && *step.Description == "" {
		step.Description = nil
	}
	if step.Summary != nil && *step.Summary == "" {
		step.Summary = nil
	}
	if step.Website != nil && *step.Website == "" {
		step.Website = nil
	}
	if step.SourceCodeURL != nil && *step.SourceCodeURL == "" {
		step.SourceCodeURL = nil
	}
	if step.SupportURL != nil && *step.SupportURL == "" {
		step.SupportURL = nil
	}
	if step.IsRequiresAdminUser != nil && *step.IsRequiresAdminUser == stepmanModels.DefaultIsRequiresAdminUser {
		step.IsRequiresAdminUser = nil
	}
	if step.IsAlwaysRun != nil && *step.IsAlwaysRun == stepmanModels.DefaultIsAlwaysRun {
		step.IsAlwaysRun = nil
	}
	if step.IsSkippable != nil && *step.IsSkippable == stepmanModels.DefaultIsSkippable {
		step.IsSkippable = nil
	}
	if step.RunIf != nil && *step.RunIf == "" {
		step.RunIf = nil
	}
	for _, env := range step.Inputs {
		if err := removeEnvironmentRedundantFields(&env); err != nil {
			return err
		}
	}
	for _, env := range step.Outputs {
		if err := removeEnvironmentRedundantFields(&env); err != nil {
			return err
		}
	}
	return nil
}

func (workflow *WorkflowModel) removeRedundantFields() error {
	for _, env := range workflow.Environments {
		if err := removeEnvironmentRedundantFields(&env); err != nil {
			return err
		}
	}

	for _, stepListItem := range workflow.Steps {
		_, step, err := GetStepIDStepDataPair(stepListItem)
		if err != nil {
			return err
		}
		if err := removeStepRedundantFields(&step); err != nil {
			return err
		}
	}
	return nil
}

func (app *AppModel) removeRedundantFields() error {
	for _, env := range app.Environments {
		if err := removeEnvironmentRedundantFields(&env); err != nil {
			return err
		}
	}
	return nil
}

// RemoveRedundantFields ...
func (config *BitriseDataModel) RemoveRedundantFields() error {
	if err := config.App.removeRedundantFields(); err != nil {
		return err
	}
	for _, workflow := range config.Workflows {
		if err := workflow.removeRedundantFields(); err != nil {
			return err
		}
	}
	return nil
}

// ----------------------------
// --- Merge

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
		options.Title = pointers.NewStringPtr(*otherOptions.Title)
	}
	if otherOptions.Description != nil {
		options.Description = pointers.NewStringPtr(*otherOptions.Description)
	}
	if otherOptions.Summary != nil {
		options.Summary = pointers.NewStringPtr(*otherOptions.Summary)
	}
	if len(otherOptions.ValueOptions) > 0 {
		options.ValueOptions = otherOptions.ValueOptions
	}
	if otherOptions.IsRequired != nil {
		options.IsRequired = pointers.NewBoolPtr(*otherOptions.IsRequired)
	}
	if otherOptions.IsExpand != nil {
		options.IsExpand = pointers.NewBoolPtr(*otherOptions.IsExpand)
	}
	if otherOptions.IsDontChangeValue != nil {
		options.IsDontChangeValue = pointers.NewBoolPtr(*otherOptions.IsDontChangeValue)
	}
	(*env)[envmanModels.OptionsKey] = options
	return nil
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

// MergeStepWith ...
func MergeStepWith(step, otherStep stepmanModels.StepModel) (stepmanModels.StepModel, error) {
	if err := step.FillMissingDefaults(); err != nil {
		return stepmanModels.StepModel{}, err
	}

	if otherStep.Title != nil {
		step.Title = pointers.NewStringPtr(*otherStep.Title)
	}
	if otherStep.Description != nil {
		step.Description = pointers.NewStringPtr(*otherStep.Description)
	}
	if otherStep.Summary != nil {
		step.Summary = pointers.NewStringPtr(*otherStep.Summary)
	}
	if otherStep.Website != nil {
		step.Website = pointers.NewStringPtr(*otherStep.Website)
	}
	if otherStep.SourceCodeURL != nil {
		step.SourceCodeURL = pointers.NewStringPtr(*otherStep.SourceCodeURL)
	}
	if otherStep.SupportURL != nil {
		step.SupportURL = pointers.NewStringPtr(*otherStep.SupportURL)
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
		step.IsRequiresAdminUser = pointers.NewBoolPtr(*otherStep.IsRequiresAdminUser)
	}
	if otherStep.IsAlwaysRun != nil {
		step.IsAlwaysRun = pointers.NewBoolPtr(*otherStep.IsAlwaysRun)
	}
	if otherStep.IsSkippable != nil {
		step.IsSkippable = pointers.NewBoolPtr(*otherStep.IsSkippable)
	}
	if otherStep.RunIf != nil {
		step.RunIf = pointers.NewStringPtr(*otherStep.RunIf)
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

// ----------------------------
// --- StepIDData

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
//    * https://github.com/bitrise-io/bitrise-steplib.git::script@2.0.0
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

// ----------------------------
// --- BuildRunResults

// IsBuildFailed ...
func (buildRes BuildRunResultsModel) IsBuildFailed() bool {
	return len(buildRes.FailedSteps) > 0
}

// HasFailedSkippableSteps ...
func (buildRes BuildRunResultsModel) HasFailedSkippableSteps() bool {
	return len(buildRes.FailedSkippableSteps) > 0
}

// ResultsCount ...
func (buildRes BuildRunResultsModel) ResultsCount() int {
	return len(buildRes.SuccessSteps) + len(buildRes.FailedSteps) + len(buildRes.FailedSkippableSteps) + len(buildRes.SkippedSteps)
}

func (buildRes BuildRunResultsModel) unorderedResults() []StepRunResultsModel {
	results := append([]StepRunResultsModel{}, buildRes.SuccessSteps...)
	results = append(results, buildRes.FailedSteps...)
	results = append(results, buildRes.FailedSkippableSteps...)
	return append(results, buildRes.SkippedSteps...)
}

//OrderedResults ...
func (buildRes BuildRunResultsModel) OrderedResults() []StepRunResultsModel {
	results := make([]StepRunResultsModel, buildRes.ResultsCount())
	unorderedResults := buildRes.unorderedResults()
	for _, result := range unorderedResults {
		results[result.Idx] = result
	}
	return results
}
