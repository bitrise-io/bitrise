package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/bitrise-io/bitrise/exitcode"
	envmanModels "github.com/bitrise-io/envman/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
)

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

func (config *BitriseDataModel) getWorkflowIDs() []string {
	uniqueWorkflowIDs := map[string]bool{}

	for workflowID := range config.Workflows {
		uniqueWorkflowIDs[workflowID] = true
	}

	var workflowIDs []string
	for workflowID := range uniqueWorkflowIDs {
		workflowIDs = append(workflowIDs, workflowID)
	}

	return workflowIDs
}

func (config *BitriseDataModel) getPipelineIDs() []string {
	uniquePipelineIDs := map[string]bool{}

	for pipelineID := range config.Pipelines {
		uniquePipelineIDs[pipelineID] = true
	}

	var pipelineIDs []string
	for pipelineID := range uniquePipelineIDs {
		pipelineIDs = append(pipelineIDs, pipelineID)
	}

	return pipelineIDs
}

// ----------------------------
// --- Normalize

func (workflow *WorkflowModel) Normalize() error {
	for _, env := range workflow.Environments {
		if err := env.Normalize(); err != nil {
			return err
		}
	}

	for _, stepListItem := range workflow.Steps {
		key, step, _, err := stepListItem.GetStepListItemKeyAndValue()
		if err != nil {
			return err
		}
		if key != StepListItemWithKey {
			if err := step.Normalize(); err != nil {
				return err
			}
			stepListItem[key] = step
		}
	}

	return nil
}

func (app *AppModel) Normalize() error {
	for _, env := range app.Environments {
		if err := env.Normalize(); err != nil {
			return err
		}
	}
	return nil
}

func (config *BitriseDataModel) Normalize() error {
	if err := config.App.Normalize(); err != nil {
		return err
	}

	normalizedTriggerMap, err := config.TriggerMap.Normalized()
	if err != nil {
		return err
	}
	config.TriggerMap = normalizedTriggerMap

	for _, workflow := range config.Workflows {
		if err := workflow.Normalize(); err != nil {
			return err
		}
	}
	normalizedMeta, err := stepmanModels.JSONMarshallable(config.Meta)
	if err != nil {
		return err
	}
	config.Meta = normalizedMeta

	return nil
}

// ----------------------------
// --- Validate

func (with WithModel) Validate(workflowID string, containers, services map[string]Container) ([]string, error) {
	var warnings []string

	if with.ContainerID != "" {
		if _, ok := containers[with.ContainerID]; !ok {
			return warnings, fmt.Errorf("container (%s) referenced in workflow (%s), but this container is not defined", with.ContainerID, workflowID)
		}
	}

	serviceIDs := map[string]bool{}
	for _, serviceID := range with.ServiceIDs {
		if _, ok := services[serviceID]; !ok {
			return warnings, fmt.Errorf("service (%s) referenced in workflow (%s), but this service is not defined", serviceID, workflowID)
		}

		if _, ok := serviceIDs[serviceID]; ok {
			return warnings, fmt.Errorf("service (%s) specified multiple times for workflow (%s)", serviceID, workflowID)
		}
		serviceIDs[serviceID] = true
	}

	for _, stepListItem := range with.Steps {
		stepID, step, err := stepListItem.GetStepIDAndStep()
		if err != nil {
			return warnings, err
		}

		warns, err := validateStep(stepID, step)
		warnings = append(warnings, warns...)
		if err != nil {
			return warnings, err
		}
	}

	return warnings, nil

}

func (workflow *WorkflowModel) Validate() ([]string, error) {
	var warnings []string

	for _, env := range workflow.Environments {
		if err := env.Validate(); err != nil {
			return warnings, err
		}
	}

	for _, stepListItem := range workflow.Steps {
		key, step, _, err := stepListItem.GetStepListItemKeyAndValue()
		if err != nil {
			return warnings, err
		}

		if key != StepListItemWithKey {
			stepID := key
			warns, err := validateStep(stepID, step)
			warnings = append(warnings, warns...)
			if err != nil {
				return warnings, err
			}

			// TODO: Why is this assignment needed?
			stepListItem[stepID] = step
		}
	}

	return warnings, nil
}

func validateStep(stepID string, step stepmanModels.StepModel) ([]string, error) {
	var warnings []string

	if err := stepid.Validate(stepID); err != nil {
		return warnings, err
	}

	if err := step.ValidateInputAndOutputEnvs(false); err != nil {
		return warnings, err
	}

	stepInputMap := map[string]bool{}
	for _, input := range step.Inputs {
		key, _, err := input.GetKeyValuePair()
		if err != nil {
			return warnings, err
		}

		_, found := stepInputMap[key]
		if found {
			warnings = append(warnings, fmt.Sprintf("invalid step: duplicated input found: (%s)", key))
		}
		stepInputMap[key] = true
	}

	return warnings, nil
}

func (app *AppModel) Validate() error {
	for _, env := range app.Environments {
		if err := env.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (config *BitriseDataModel) Validate() ([]string, error) {
	var warnings []string

	if config.FormatVersion == "" {
		return warnings, fmt.Errorf("missing format_version")
	}

	// trigger map
	workflows := config.getWorkflowIDs()
	pipelines := config.getPipelineIDs()
	warns, err := config.TriggerMap.Validate(workflows, pipelines)
	warnings = append(warnings, warns...)
	if err != nil {
		return warnings, err
	}
	// ---

	// app
	if err := config.App.Validate(); err != nil {
		return warnings, err
	}
	// ---

	// containers
	for containerID, containerDef := range config.Containers {
		if containerID == "" {
			return nil, fmt.Errorf("service (image: %s) has empty ID defined", containerDef.Image)
		}
		if strings.TrimSpace(containerDef.Image) == "" {
			return warnings, fmt.Errorf("service (%s) has no image defined", containerID)
		}
	}

	for serviceID, serviceDef := range config.Services {
		if serviceID == "" {
			return nil, fmt.Errorf("service (image: %s) has empty ID defined", serviceDef.Image)
		}
		if strings.TrimSpace(serviceDef.Image) == "" {
			return warnings, fmt.Errorf("service (%s) has no image defined", serviceID)
		}
	}
	// ---

	// pipelines
	pipelineWarnings, err := validatePipelines(config)
	warnings = append(warnings, pipelineWarnings...)
	if err != nil {
		return warnings, err
	}
	// ---

	// stages
	stageWarnings, err := validateStages(config)
	warnings = append(warnings, stageWarnings...)
	if err != nil {
		return warnings, err
	}
	// ---

	// workflows
	workflowWarnings, err := validateWorkflows(config)
	warnings = append(warnings, workflowWarnings...)
	if err != nil {
		return warnings, err
	}

	for workflowID, workflow := range config.Workflows {
		for _, stepListItem := range workflow.Steps {
			key, _, with, err := stepListItem.GetStepListItemKeyAndValue()
			if err != nil {
				return warnings, err
			}
			if key == StepListItemWithKey {
				warns, err := with.Validate(workflowID, config.Containers, config.Services)
				warnings = append(warnings, warns...)
				if err != nil {
					return warnings, err
				}
			}
		}
	}
	// ---

	return warnings, nil
}

func validatePipelines(config *BitriseDataModel) ([]string, error) {
	pipelineWarnings := make([]string, 0)
	for ID, pipeline := range config.Pipelines {
		idWarning, err := validateID(ID, "pipeline")
		if idWarning != "" {
			pipelineWarnings = append(pipelineWarnings, idWarning)
		}
		if err != nil {
			return pipelineWarnings, err
		}

		if len(pipeline.Stages) == 0 {
			return pipelineWarnings, fmt.Errorf("pipeline (%s) should have at least 1 stage", ID)
		}

		for _, pipelineStage := range pipeline.Stages {
			pipelineStageID, err := getStageID(pipelineStage)
			if err != nil {
				return pipelineWarnings, err
			}
			found := false
			for stageID := range config.Stages {
				if stageID == pipelineStageID {
					found = true
					break
				}
			}
			if !found {
				return pipelineWarnings, fmt.Errorf("stage (%s) defined in pipeline (%s), but does not exist", pipelineStageID, ID)
			}
		}
	}

	return pipelineWarnings, nil
}

func validateStages(config *BitriseDataModel) ([]string, error) {
	stageWarnings := make([]string, 0)
	for ID, stage := range config.Stages {
		idWarning, err := validateID(ID, "stage")
		if idWarning != "" {
			stageWarnings = append(stageWarnings, idWarning)
		}
		if err != nil {
			return stageWarnings, err
		}

		if len(stage.Workflows) == 0 {
			return stageWarnings, fmt.Errorf("stage (%s) should have at least 1 workflow", ID)
		}

		for _, stageWorkflow := range stage.Workflows {
			found := false
			stageWorkflowID, err := getWorkflowID(stageWorkflow)

			if isUtilityWorkflow(stageWorkflowID) {
				return stageWarnings, fmt.Errorf("workflow (%s) defined in stage (%s), is a utility workflow", stageWorkflowID, ID)
			}

			if err != nil {
				return stageWarnings, err
			}
			for workflowID := range config.Workflows {
				if workflowID == stageWorkflowID {
					found = true
					break
				}
			}
			if !found {
				return stageWarnings, fmt.Errorf("workflow (%s) defined in stage (%s), but does not exist", stageWorkflowID, ID)
			}
		}
	}

	return stageWarnings, nil
}

func isUtilityWorkflow(workflowID string) bool {
	return strings.HasPrefix(workflowID, "_")
}

func validateWorkflows(config *BitriseDataModel) ([]string, error) {
	workflowWarnings := make([]string, 0)
	for ID, workflow := range config.Workflows {
		idWarning, err := validateID(ID, "workflow")
		if idWarning != "" {
			workflowWarnings = append(workflowWarnings, idWarning)
		}
		if err != nil {
			return workflowWarnings, err
		}

		warns, err := workflow.Validate()
		workflowWarnings = append(workflowWarnings, warns...)
		if err != nil {
			return workflowWarnings, fmt.Errorf("validation error in workflow: %s: %s", ID, err)
		}

		if err := checkWorkflowReferenceCycle(ID, workflow, *config, []string{}); err != nil {
			return workflowWarnings, err
		}
	}

	return workflowWarnings, nil
}

func validateID(id, modelType string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("invalid %s ID (%s): empty", modelType, id)
	}

	r := regexp.MustCompile(`[A-Za-z0-9-_.]+`)
	if find := r.FindString(id); find != id {
		return fmt.Sprintf("invalid %s ID (%s): doesn't conform to: [A-Za-z0-9-_.]", modelType, id), nil
	}

	return "", nil
}

// ----------------------------
// --- FillMissingDefaults

func (workflow *WorkflowModel) FillMissingDefaults(title string) error {
	// Don't call step.FillMissingDefaults()
	// StepLib versions of steps (which are the default versions),
	// contains different env defaults then normal envs
	// example: isExpand = true by default for normal envs,
	// but script step content input env isExpand = false by default

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

func (app *AppModel) FillMissingDefaults() error {
	for _, env := range app.Environments {
		if err := env.FillMissingDefaults(); err != nil {
			return err
		}
	}
	return nil
}

func (config *BitriseDataModel) FillMissingDefaults() error {
	if err := config.App.FillMissingDefaults(); err != nil {
		return err
	}

	for title, workflow := range config.Workflows {
		if err := workflow.FillMissingDefaults(title); err != nil {
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

	if options.IsSensitive != nil {
		if *options.IsSensitive == envmanModels.DefaultIsSensitive {
			options.IsSensitive = nil
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
	if options.SkipIfEmpty != nil {
		if *options.SkipIfEmpty == envmanModels.DefaultSkipIfEmpty {
			options.SkipIfEmpty = nil
		} else {
			hasOptions = true
		}
	}
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
	if options.Category != nil {
		if *options.Category == "" {
			options.Category = nil
		} else {
			hasOptions = true
		}
	}
	if options.ValueOptions != nil && len(options.ValueOptions) > 0 {
		hasOptions = true
	}
	if options.IsRequired != nil {
		if *options.IsRequired == envmanModels.DefaultIsRequired {
			options.IsRequired = nil
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
	if options.IsTemplate != nil {
		if *options.IsTemplate == envmanModels.DefaultIsTemplate {
			options.IsTemplate = nil
		} else {
			hasOptions = true
		}
	}
	if options.Meta != nil && len(options.Meta) > 0 {
		hasOptions = true
	}

	if hasOptions {
		(*env)[envmanModels.OptionsKey] = options
	} else {
		delete(*env, envmanModels.OptionsKey)
	}

	return nil
}

func (workflow *WorkflowModel) removeRedundantFields() error {
	// Don't call step.RemoveRedundantFields()
	// StepLib versions of steps (which are the default versions),
	// contains different env defaults then normal envs
	// example: isExpand = true by default for normal envs,
	// but script step content input env isExpand = false by default
	for _, env := range workflow.Environments {
		if err := removeEnvironmentRedundantFields(&env); err != nil {
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
// --- WorkflowIDData

func getWorkflowID(workflowListItem StageWorkflowListItemModel) (string, error) {
	if len(workflowListItem) > 1 {
		return "", errors.New("StageWorkflowListItemModel contains more than 1 key-value pair")
	}
	for key := range workflowListItem {
		return key, nil
	}
	return "", errors.New("StageWorkflowListItemModel does not contain a key-value pair")
}

// ----------------------------
// --- StageIDData

func getStageID(stageListItem StageListItemModel) (string, error) {
	if len(stageListItem) > 1 {
		return "", errors.New("StageListItemModel contains more than 1 key-value pair")
	}
	for key := range stageListItem {
		return key, nil
	}
	return "", errors.New("StageListItemModel does not contain a key-value pair")
}

// ----------------------------
// --- StepIDData

func (stepListItem *StepListItemModel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	var key string
	for k := range raw {
		key = k
		break
	}

	if key == StepListItemWithKey {
		var withItem StepListWithItemModel
		if err := unmarshal(&withItem); err != nil {
			return err
		}

		*stepListItem = map[string]interface{}{}
		for k, v := range withItem {
			(*stepListItem)[k] = v
		}
	} else {
		var stepItem StepListStepItemModel
		if err := unmarshal(&stepItem); err != nil {
			return err
		}

		*stepListItem = map[string]interface{}{}
		for k, v := range stepItem {
			(*stepListItem)[k] = v
		}
	}

	return nil
}

func (stepListStepItem *StepListStepItemModel) GetStepIDAndStep() (string, stepmanModels.StepModel, error) {
	if stepListStepItem == nil {
		return "", stepmanModels.StepModel{}, nil
	}

	if len(*stepListStepItem) == 0 {
		return "", stepmanModels.StepModel{}, errors.New("stepListStepItem does not contain a key-value pair")
	}

	if len(*stepListStepItem) > 1 {
		return "", stepmanModels.StepModel{}, errors.New("stepListStepItem contains more than 1 key-value pair")
	}

	var stepID string
	var step stepmanModels.StepModel
	for k, v := range *stepListStepItem {
		stepID = k
		step = v
		break
	}

	return stepID, step, nil
}

// GetStepListItemKeyAndValue returns the Step List Item key and value. The key is either a Step ID or 'with'.
// If the key is 'with' the returned WithModel is relevant otherwise the StepModel.
func (stepListItem *StepListItemModel) GetStepListItemKeyAndValue() (string, stepmanModels.StepModel, WithModel, error) {
	if stepListItem == nil {
		return "", stepmanModels.StepModel{}, WithModel{}, nil
	}

	if len(*stepListItem) == 0 {
		return "", stepmanModels.StepModel{}, WithModel{}, errors.New("StepListItem does not contain a key-value pair")
	}

	if len(*stepListItem) > 1 {
		return "", stepmanModels.StepModel{}, WithModel{}, errors.New("StepListItem contains more than 1 key-value pair")
	}

	for key, value := range *stepListItem {
		if key == StepListItemWithKey {
			with := value.(WithModel)
			return key, stepmanModels.StepModel{}, with, nil
		} else {
			step, ok := value.(stepmanModels.StepModel)
			if ok {
				return key, step, WithModel{}, nil
			}

			// StepListItemModel is a map[string]interface{}, when it comes from a JSON/YAML unmarshal
			// the StepModel has a pointer type.
			stepPtr, ok := value.(*stepmanModels.StepModel)
			if ok {
				return key, *stepPtr, WithModel{}, nil
			}

			return key, stepmanModels.StepModel{}, WithModel{}, nil
		}
	}
	return "", stepmanModels.StepModel{}, WithModel{}, nil
}

// ----------------------------
// --- BuildRunResults

func (buildRes BuildRunResultsModel) IsStepLibUpdated(stepLib string) bool {
	return (buildRes.StepmanUpdates[stepLib] > 0)
}

func (buildRes BuildRunResultsModel) IsBuildFailed() bool {
	return len(buildRes.FailedSteps) > 0
}

func (buildRes BuildRunResultsModel) ExitCode() int {
	if !buildRes.IsBuildFailed() {
		return 0
	}

	if buildRes.isBuildAbortedWithNoOutputTimeout() {
		return exitcode.CLIAbortedWithNoOutputTimeout
	}

	if buildRes.isBuildAbortedWithTimeout() {
		return exitcode.CLIAbortedWithCustomTimeout
	}

	return exitcode.CLIFailed
}

func (buildRes BuildRunResultsModel) HasFailedSkippableSteps() bool {
	return len(buildRes.FailedSkippableSteps) > 0
}

func (buildRes BuildRunResultsModel) ResultsCount() int {
	return len(buildRes.SuccessSteps) + len(buildRes.FailedSteps) + len(buildRes.FailedSkippableSteps) + len(buildRes.SkippedSteps)
}

func (buildRes BuildRunResultsModel) isBuildAbortedWithTimeout() bool {
	for _, stepResult := range buildRes.FailedSteps {
		if stepResult.Status == StepRunStatusAbortedWithCustomTimeout {
			return true
		}
	}

	return false
}

func (buildRes BuildRunResultsModel) isBuildAbortedWithNoOutputTimeout() bool {
	for _, stepResult := range buildRes.FailedSteps {
		if stepResult.Status == StepRunStatusAbortedWithNoOutputTimeout {
			return true
		}
	}

	return false
}

func (buildRes BuildRunResultsModel) unorderedResults() []StepRunResultsModel {
	results := append([]StepRunResultsModel{}, buildRes.SuccessSteps...)
	results = append(results, buildRes.FailedSteps...)
	results = append(results, buildRes.FailedSkippableSteps...)
	return append(results, buildRes.SkippedSteps...)
}

func (buildRes BuildRunResultsModel) OrderedResults() []StepRunResultsModel {
	results := make([]StepRunResultsModel, buildRes.ResultsCount())
	unorderedResults := buildRes.unorderedResults()
	for _, result := range unorderedResults {
		results[result.Idx] = result
	}
	return results
}
