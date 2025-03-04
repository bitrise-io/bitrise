package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/sliceutil"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/heimdalr/dag"
)

// TODO: can we replace these with slices package functions?
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

func checkStepBundleReferenceCycle(stepBundleID string, stepBundle StepBundleModel, bitriseConfig BitriseDataModel, stepBundleStack []string) error {
	if sliceutil.IsStringInSlice(stepBundleID, stepBundleStack) {
		stackStr := ""
		for _, aStepBundleID := range stepBundleStack {
			stackStr += aStepBundleID + " -> "
		}
		stackStr += stepBundleID
		return fmt.Errorf("step bundle reference cycle found: %s", stackStr)
	}
	stepBundleStack = append(stepBundleStack, stepBundleID)

	for _, stepListItem := range stepBundle.Steps {
		key, t, err := stepListItem.GetKeyAndType()
		if err != nil {
			return err
		}
		if t == StepListItemTypeBundle {
			definition := bitriseConfig.StepBundles[key]

			err := checkStepBundleReferenceCycle(key, definition, bitriseConfig, stepBundleStack)
			if err != nil {
				return err
			}
		}
	}

	return nil
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

func (bundle *StepBundleModel) Normalize() error {
	for idx, stepListItem := range bundle.Steps {
		key, t, err := stepListItem.GetKeyAndType()
		if err != nil {
			return err
		}

		if t == StepListItemTypeStep {
			step, err := stepListItem.GetStep()
			if err != nil {
				return err
			}

			if err := step.Normalize(); err != nil {
				return err
			}

			stepListItem[key] = *step
			bundle.Steps[idx] = stepListItem
		} else if t == StepListItemTypeBundle {
			b, err := stepListItem.GetBundle()
			if err != nil {
				return err
			}

			if err := b.Normalize(); err != nil {
				return err
			}

			stepListItem[StepListItemStepBundleKeyPrefix+key] = *b
			bundle.Steps[idx] = stepListItem
		}
	}

	for i, input := range bundle.Inputs {
		if err := input.Normalize(); err != nil {
			return err
		}
		bundle.Inputs[i] = input
	}

	for i, env := range bundle.Environments {
		if err := env.Normalize(); err != nil {
			return err
		}
		bundle.Environments[i] = env
	}

	return nil
}

func (bundle *StepBundleListItemModel) Normalize() error {
	for i, input := range bundle.Inputs {
		if err := input.Normalize(); err != nil {
			return err
		}
		bundle.Inputs[i] = input
	}

	for i, env := range bundle.Environments {
		if err := env.Normalize(); err != nil {
			return err
		}
		bundle.Environments[i] = env
	}

	return nil
}

func (with *WithModel) Normalize() error {
	for idx, stepListItem := range with.Steps {
		stepID, step, err := stepListItem.GetStepIDAndStep()
		if err != nil {
			return err
		}

		if err := step.Normalize(); err != nil {
			return err
		}
		stepListItem[stepID] = step
		with.Steps[idx] = stepListItem
	}
	return nil
}

func (container *Container) Normalize() error {
	for i, env := range container.Envs {
		if err := env.Normalize(); err != nil {
			return err
		}
		container.Envs[i] = env
	}
	return nil
}

func (workflow *WorkflowModel) Normalize() error {
	for _, env := range workflow.Environments {
		if err := env.Normalize(); err != nil {
			return err
		}
	}

	for idx, stepListItem := range workflow.Steps {
		key, t, err := stepListItem.GetKeyAndType()
		if err != nil {
			return err
		}
		if t == StepListItemTypeStep {
			step, err := stepListItem.GetStep()
			if err != nil {
				return err
			}

			if err := step.Normalize(); err != nil {
				return err
			}

			stepListItem[key] = *step
			workflow.Steps[idx] = stepListItem
		} else if t == StepListItemTypeBundle {
			bundle, err := stepListItem.GetBundle()
			if err != nil {
				return err
			}

			if err := bundle.Normalize(); err != nil {
				return err
			}

			stepListItem[StepListItemStepBundleKeyPrefix+key] = *bundle
			workflow.Steps[idx] = stepListItem
		} else if t == StepListItemTypeWith {
			with, err := stepListItem.GetWith()
			if err != nil {
				return err
			}

			if err := with.Normalize(); err != nil {
				return err
			}

			stepListItem[key] = *with
			workflow.Steps[idx] = stepListItem
		}
	}

	normalizedMeta, err := stepmanModels.JSONMarshallable(workflow.Meta)
	if err != nil {
		return fmt.Errorf("failed to normalize meta: %w", err)
	}
	workflow.Meta = normalizedMeta

	return nil
}

func (app *AppModel) Normalize() error {
	for idx, env := range app.Environments {
		if err := env.Normalize(); err != nil {
			return err
		}
		app.Environments[idx] = env
	}
	return nil
}

func (config *BitriseDataModel) Normalize() error {
	if err := config.App.Normalize(); err != nil {
		return fmt.Errorf("failed to normalize app: %w", err)
	}

	normalizedTriggerMap, err := config.TriggerMap.Normalized()
	if err != nil {
		return fmt.Errorf("failed to normalize trigger_map: %w", err)
	}
	config.TriggerMap = normalizedTriggerMap

	for containerID, container := range config.Containers {
		if err := container.Normalize(); err != nil {
			return fmt.Errorf("failed to normalize container: %w", err)
		}
		config.Containers[containerID] = container
	}

	for serviceID, service := range config.Services {
		if err := service.Normalize(); err != nil {
			return fmt.Errorf("failed to normalize service: %w", err)
		}
		config.Services[serviceID] = service
	}

	for stepBundleID, stepBundle := range config.StepBundles {
		if err := stepBundle.Normalize(); err != nil {
			return fmt.Errorf("failed to normalize step_bundle: %w", err)
		}
		config.StepBundles[stepBundleID] = stepBundle
	}

	for workflowID, workflow := range config.Workflows {
		if err := workflow.Normalize(); err != nil {
			return fmt.Errorf("failed to normalize workflow: %w", err)
		}
		config.Workflows[workflowID] = workflow
	}

	normalizedMeta, err := stepmanModels.JSONMarshallable(config.Meta)
	if err != nil {
		return fmt.Errorf("failed to normalize meta: %w", err)
	}
	config.Meta = normalizedMeta

	return nil
}

// ----------------------------
// --- Validate

func (bundle *StepBundleListItemModel) Validate(stepBundleDefinition StepBundleModel) error {
	stepBundleDefinitionInputKeys := map[string]bool{}
	for _, input := range stepBundleDefinition.Inputs {
		key, _, err := input.GetKeyValuePair()
		if err != nil {
			return err
		}
		stepBundleDefinitionInputKeys[key] = true
	}

	for _, input := range bundle.Inputs {
		if err := input.Validate(); err != nil {
			return err
		}
		key, _, err := input.GetKeyValuePair()
		if err != nil {
			return err
		}
		if _, ok := stepBundleDefinitionInputKeys[key]; !ok {
			return fmt.Errorf("input (%s) is not defined in the step bundle definition", key)
		}
	}

	for _, env := range bundle.Environments {
		if err := env.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (bundle *StepBundleModel) Validate(stepBundleDefinitions map[string]StepBundleModel) ([]string, error) {
	var warnings []string

	for _, stepListItem := range bundle.Steps {
		key, t, err := stepListItem.GetKeyAndType()
		if err != nil {
			return warnings, err
		}

		if t == StepListItemTypeWith {
			return warnings, errors.New("'with group' is not allowed in a step bundle")
		}

		if t == StepListItemTypeStep {
			step, err := stepListItem.GetStep()
			if err != nil {
				return warnings, err
			}

			warns, err := validateStep(key, *step)
			warnings = append(warnings, warns...)
			if err != nil {
				return warnings, err
			}
		} else if t == StepListItemTypeBundle {
			override, err := stepListItem.GetBundle()
			if err != nil {
				return warnings, err
			}

			definition, ok := stepBundleDefinitions[key]
			if !ok {
				return warnings, fmt.Errorf("referenced step bundle not defined: %s", key)
			}

			if err := override.Validate(definition); err != nil {
				return warnings, err
			}
		}
	}

	for _, input := range bundle.Inputs {
		if err := input.Validate(); err != nil {
			return warnings, err
		}
	}

	for _, env := range bundle.Environments {
		if err := env.Validate(); err != nil {
			return warnings, err
		}
	}

	return warnings, nil
}

func (container *Container) Validate() error {
	for _, env := range container.Envs {
		if err := env.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (with *WithModel) Validate(workflowID string, containers, services map[string]Container) ([]string, error) {
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

		if stepID == StepListItemWithKey {
			return warnings, fmt.Errorf("invalid 'with group' in workflow (%s): 'with group' is not allowed in a 'with group''s step list", workflowID)
		} else if strings.HasPrefix(stepID, StepListItemStepBundleKeyPrefix) {
			return warnings, fmt.Errorf("invalid 'with group' in workflow (%s): step bundle is not allowed in a 'with group''s step list", workflowID)
		}

		warns, err := validateStep(stepID, step)
		warnings = append(warnings, warns...)
		if err != nil {
			return warnings, err
		}
	}

	return warnings, nil
}

func (workflow *WorkflowModel) Validate() error {
	for _, env := range workflow.Environments {
		if err := env.Validate(); err != nil {
			return err
		}
	}

	return validateStatusReportName(workflow.StatusReportName)
}

const statusReportNameRegex = `^[a-zA-Z0-9,./():\-_ <>[\]|]*$`

func validateStatusReportName(statusReportName string) error {
	if len(statusReportName) > 100 {
		return fmt.Errorf("status_report_name (%s) is too long, max length is 100 characters", statusReportName)
	}

	re := regexp.MustCompile(statusReportNameRegex)
	if !re.MatchString(statusReportName) {
		return fmt.Errorf("status_report_name (%s) contains invalid characters, should match the '%s' regex", statusReportName, statusReportNameRegex)
	}
	return nil
}

func validatePriority(priority *int) error {
	if priority == nil {
		return nil
	}

	if *priority > 100 || *priority < -100 {
		return fmt.Errorf("priority (%d) should be between -100 and 100", *priority)
	}
	return nil
}

func (app *AppModel) Validate() error {
	return app.internalValidation(true)
}

func (app *AppModel) MinimalValidation() error {
	return app.internalValidation(false)
}

func (app *AppModel) internalValidation(full bool) error {
	for _, env := range app.Environments {
		if err := env.Validate(); err != nil {
			return err
		}
	}

	if !full {
		return nil
	}

	return validateStatusReportName(app.StatusReportName)
}

func (config *BitriseDataModel) Validate() ([]string, error) {
	return config.internalValidation(true)
}

func (config *BitriseDataModel) MinimalValidation() ([]string, error) {
	return config.internalValidation(false)
}

func (config *BitriseDataModel) internalValidation(full bool) ([]string, error) {
	var warnings []string

	if config.FormatVersion == "" {
		return warnings, fmt.Errorf("missing format_version")
	}

	// trigger map
	if full {
		workflows := config.getWorkflowIDs()
		pipelines := config.getPipelineIDs()
		triggerMapWarnings, err := config.TriggerMap.Validate(workflows, pipelines)
		warnings = append(warnings, triggerMapWarnings...)
		if err != nil {
			return warnings, err
		}
	}
	// ---

	// app
	var appValidationFunc func() error
	if full {
		appValidationFunc = config.App.Validate
	} else {
		appValidationFunc = config.App.MinimalValidation
	}
	if err := appValidationFunc(); err != nil {
		return warnings, err
	}
	// ---

	// containers
	if err := validateContainers(*config); err != nil {
		return warnings, err
	}
	// ---

	// step_bundles
	stepBundleWarnings, err := validateStepBundles(*config)
	warnings = append(warnings, stepBundleWarnings...)
	if err != nil {
		return warnings, err
	}
	// ---

	// pipelines
	if full {
		pipelineWarnings, err := validatePipelines(config)
		warnings = append(warnings, pipelineWarnings...)
		if err != nil {
			return warnings, err
		}
	}
	// ---

	// stages
	if full {
		stageWarnings, err := validateStages(config)
		warnings = append(warnings, stageWarnings...)
		if err != nil {
			return warnings, err
		}
	}
	// ---

	// workflows
	workflowWarnings, err := validateWorkflows(config)
	warnings = append(warnings, workflowWarnings...)
	if err != nil {
		return warnings, err
	}
	// ---

	return warnings, nil
}

func validateContainers(config BitriseDataModel) error {
	for containerID, containerDef := range config.Containers {
		if containerID == "" {
			return fmt.Errorf("container (image: %s) has empty ID defined", containerDef.Image)
		}
		if strings.TrimSpace(containerDef.Image) == "" {
			return fmt.Errorf("container (%s) has no image defined", containerID)
		}
		if err := containerDef.Validate(); err != nil {
			return fmt.Errorf("container (%s) has config issue: %w", containerID, err)
		}
	}

	for serviceID, serviceDef := range config.Services {
		if serviceID == "" {
			return fmt.Errorf("service (image: %s) has empty ID defined", serviceDef.Image)
		}
		if strings.TrimSpace(serviceDef.Image) == "" {
			return fmt.Errorf("service (%s) has no image defined", serviceID)
		}
		if err := serviceDef.Validate(); err != nil {
			return fmt.Errorf("container (%s) has config issue: %w", serviceID, err)
		}
	}

	return nil
}

func validateStepBundles(config BitriseDataModel) ([]string, error) {
	var warnings []string

	bundleIDs := make([]string, len(config.StepBundles))
	for bundleID := range config.StepBundles {
		if bundleID == "" {
			return warnings, fmt.Errorf("step bundle has empty ID defined")
		}

		bundleIDs = append(bundleIDs, bundleID)
	}
	sort.Strings(bundleIDs)

	for _, bundleID := range bundleIDs {
		bundle := config.StepBundles[bundleID]

		if err := checkStepBundleReferenceCycle(bundleID, bundle, config, []string{}); err != nil {
			return warnings, err
		}

		warns, err := bundle.Validate(config.StepBundles)
		warnings = append(warnings, warns...)
		if err != nil {
			return warnings, fmt.Errorf("step bundle (%s) has config issue: %w", bundleID, err)
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

func validatePipelines(config *BitriseDataModel) ([]string, error) {
	pipelineWarnings := make([]string, 0)
	for pipelineID, pipeline := range config.Pipelines {
		idWarning, err := validateID(pipelineID, "pipeline")
		if idWarning != "" {
			pipelineWarnings = append(pipelineWarnings, idWarning)
		}
		if err != nil {
			return pipelineWarnings, err
		}

		if err := validateStatusReportName(pipeline.StatusReportName); err != nil {
			return pipelineWarnings, fmt.Errorf("pipeline (%s) has invalid status_report_name: %w", pipelineID, err)
		}

		if err := validatePriority(pipeline.Priority); err != nil {
			return pipelineWarnings, fmt.Errorf("pipeline (%s) has invalid priority: %w", pipelineID, err)
		}

		hasStages := len(pipeline.Stages) > 0
		hasWorkflows := len(pipeline.Workflows) > 0

		if hasStages && hasWorkflows {
			return pipelineWarnings, fmt.Errorf("pipeline (%s) has both stages and workflows", pipelineID)
		}

		// A pipeline is considered valid if it has neither stages nor workflows.
		// This is useful for the WFE to be able to save a pipeline that is not yet fully defined.
		if !hasStages && !hasWorkflows {
			warning := fmt.Sprintf("pipeline (%s) should have at least 1 stage or workflow", pipelineID)
			pipelineWarnings = append(pipelineWarnings, warning)

			continue
		}

		if hasStages {
			if err := validateStagedPipeline(pipelineID, &pipeline, config); err != nil {
				return pipelineWarnings, err
			}
		} else {
			if err := validateDAGPipeline(pipelineID, &pipeline, config); err != nil {
				return pipelineWarnings, err
			}
		}
	}

	return pipelineWarnings, nil
}

func validateStagedPipeline(pipelineID string, pipeline *PipelineModel, config *BitriseDataModel) error {
	for _, pipelineStage := range pipeline.Stages {
		pipelineStageID, err := getStageID(pipelineStage)
		if err != nil {
			return err
		}

		if _, ok := config.Stages[pipelineStageID]; !ok {
			return fmt.Errorf("stage (%s) defined in pipeline (%s) does not exist", pipelineStageID, pipelineID)
		}
	}

	return nil
}

func validateDAGPipeline(pipelineID string, pipeline *PipelineModel, config *BitriseDataModel) error {
	for pipelineWorkflowID, pipelineWorkflow := range pipeline.Workflows {
		if isUtilityWorkflow(pipelineWorkflowID) {
			return fmt.Errorf("workflow (%s) defined in pipeline (%s) is a utility workflow", pipelineWorkflowID, pipelineID)
		}

		isParallelizedWorkflow := pipelineWorkflow.Parallel != ""
		if isParallelizedWorkflow {
			if err := validateParallelizedWorkflow(pipelineID, pipelineWorkflowID, pipelineWorkflow, config); err != nil {
				return err
			}
		}

		isWorkflowVariant := pipelineWorkflow.Uses != ""
		if isWorkflowVariant {
			if err := validateWorkflowVariant(pipelineID, pipelineWorkflowID, pipelineWorkflow, config); err != nil {
				return err
			}
		} else {
			if _, ok := config.Workflows[pipelineWorkflowID]; !ok {
				return fmt.Errorf("workflow (%s) defined in pipeline (%s) is not found in the workflow definitions", pipelineWorkflowID, pipelineID)
			}

			if 0 < len(pipelineWorkflow.Inputs) {
				return fmt.Errorf("workflow (%s) defined in pipeline (%s) has inputs but it is not a workflow variant", pipelineWorkflowID, pipelineID)
			}
		}

		uniqueItems := make(map[string]bool)

		for _, identifier := range pipelineWorkflow.DependsOn {
			if uniqueItems[identifier] {
				return fmt.Errorf("workflow (%s) is duplicated in the dependency list (%s)", identifier, pipelineWorkflowID)
			}

			uniqueItems[identifier] = true

			if _, ok := pipeline.Workflows[identifier]; !ok {
				return fmt.Errorf("workflow (%s) defined in dependencies (%s) is not part of pipeline (%s)", identifier, pipelineWorkflowID, pipelineID)
			}
		}
	}

	return validateGraph(pipeline)
}

func validateParallelizedWorkflow(pipelineID, workflowID string, workflow GraphPipelineWorkflowModel, config *BitriseDataModel) error {
	isDynamic := false
	parallel, err := strconv.Atoi(workflow.Parallel)
	if err != nil {
		isDynamic = strings.HasPrefix(workflow.Parallel, "$")

		if !isDynamic {
			return fmt.Errorf("workflow (%s) defined in pipeline (%s) has invalid parallel value (%s), should be an integer or a reference to an environment variable", workflowID, pipelineID, workflow.Parallel)
		}
	}

	if isDynamic {
		return validateDynamicParallelizedWorkflow(pipelineID, workflowID, config)
	}

	return validateStaticParallelizedWorkflow(pipelineID, workflowID, parallel, config)
}

func validateStaticParallelizedWorkflow(pipelineID, workflowID string, parallel int, config *BitriseDataModel) error {
	if parallel < 1 {
		return fmt.Errorf("workflow (%s) defined in pipeline (%s) has invalid parallel value (%d), should be at least 1", workflowID, pipelineID, parallel)
	}

	for i := 1; i <= parallel; i++ {
		generatedWorkflowID := fmt.Sprintf("%s_%d", workflowID, i)
		if _, ok := config.Workflows[generatedWorkflowID]; ok {
			return fmt.Errorf("parallel workflow variant (%s) would be generated by workflow (%s) defined in pipeline (%s), but it is also defined as a workflow", generatedWorkflowID, workflowID, pipelineID)
		}
	}

	return nil
}

func validateDynamicParallelizedWorkflow(pipelineID, workflowID string, config *BitriseDataModel) error {
	pattern := fmt.Sprintf("^%s_[0-9]+$", workflowID)

	for workflowName := range config.Pipelines[pipelineID].Workflows {
		matched, err := regexp.MatchString(pattern, workflowName)
		if err != nil {
			return err
		}
		if matched {
			return fmt.Errorf("dynamic parallel workflow (%s) could collide with pipeline workflow (%s) during runtime", workflowID, workflowName)
		}
	}

	for workflowName := range config.Workflows {
		matched, err := regexp.MatchString(pattern, workflowName)
		if err != nil {
			return err
		}
		if matched {
			return fmt.Errorf("dynamic parallel workflow (%s) could collide with workflow (%s) during runtime", workflowID, workflowName)
		}
	}

	return nil
}

func validateWorkflowVariant(pipelineID, workflowID string, workflow GraphPipelineWorkflowModel, config *BitriseDataModel) error {
	if _, ok := config.Workflows[workflow.Uses]; !ok {
		return fmt.Errorf("workflow (%s) referenced in pipeline (%s) in workflow variant (%s) is not found in the workflow definitions", workflow.Uses, pipelineID, workflowID)
	}

	if _, ok := config.Workflows[workflowID]; ok {
		return fmt.Errorf("workflow (%s) defined in pipeline (%s) is a variant of another workflow, but it is also defined as a workflow", workflowID, pipelineID)
	}

	return nil
}

func validateGraph(pipeline *PipelineModel) error {
	d := dag.NewDAG()
	for identifier := range pipeline.Workflows {
		// The second argument in AddVertexByID is the "value" which cannot be empty,
		// but we will rely on the first argument (ID) only
		err := d.AddVertexByID(identifier, identifier)
		if err != nil {
			return err
		}
	}

	for identifier, workflow := range pipeline.Workflows {
		for _, dependency := range workflow.DependsOn {
			err := d.AddEdge(dependency, identifier)
			if err != nil {
				if errors.As(err, &dag.EdgeLoopError{}) {
					// The workflows are in a map object, and the order of these two in the error message was non-deterministic.
					// We need to sort them, so they appear always in the same order.
					items := []string{identifier, dependency}
					sort.Strings(items)

					return fmt.Errorf("the dependency between workflow '%s' and workflow '%s' creates a cycle in the graph", items[0], items[1])
				}
				return err
			}
		}
	}

	return nil
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
	var warnings []string

	for workflowID, workflow := range config.Workflows {
		idWarning, err := validateID(workflowID, "workflow")
		if idWarning != "" {
			warnings = append(warnings, idWarning)
		}
		if err != nil {
			return warnings, err
		}

		if err := checkWorkflowReferenceCycle(workflowID, workflow, *config, []string{}); err != nil {
			return warnings, err
		}

		if err := workflow.Validate(); err != nil {
			return warnings, fmt.Errorf("workflow (%s) has config issue: %w", workflowID, err)
		}

		if err := validatePriority(workflow.Priority); err != nil {
			return warnings, fmt.Errorf("workflow (%s) has invalid priority: %w", workflowID, err)
		}

		for _, stepListItem := range workflow.Steps {
			key, t, err := stepListItem.GetKeyAndType()
			if err != nil {
				return warnings, err
			}

			if t == StepListItemTypeStep {
				step, err := stepListItem.GetStep()
				if err != nil {
					return warnings, err
				}
				stepID := key
				warns, err := validateStep(stepID, *step)
				warnings = append(warnings, warns...)
				if err != nil {
					return warnings, err
				}

				// TODO: Why is this assignment needed?
				stepListItem[stepID] = *step
			} else if t == StepListItemTypeWith {
				with, err := stepListItem.GetWith()
				if err != nil {
					return warnings, err
				}

				warns, err := with.Validate(workflowID, config.Containers, config.Services)
				warnings = append(warnings, warns...)
				if err != nil {
					return warnings, err
				}
			} else if t == StepListItemTypeBundle {
				bundleID := strings.TrimPrefix(key, StepListItemStepBundleKeyPrefix)
				bundleDefinition, ok := config.StepBundles[bundleID]
				if !ok {
					return warnings, fmt.Errorf("step bundle (%s) referenced in workflow (%s), but this step-bundle is not defined", bundleID, workflowID)
				}

				bundle, err := stepListItem.GetBundle()
				if err != nil {
					return warnings, err
				}

				if err := bundle.Validate(bundleDefinition); err != nil {
					return warnings, fmt.Errorf("step bundle (%s) referenced in workflow (%s) has config issue: %w", bundleID, workflowID, err)
				}
			}
		}
	}

	return warnings, nil
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
	if len(options.ValueOptions) > 0 {
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
	if len(options.Meta) > 0 {
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
// --- Merge

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

	// merge options
	options, err := env.GetOptions()
	if err != nil {
		return err
	}

	otherOptions, err := otherEnv.GetOptions()
	if err != nil {
		return err
	}

	if otherOptions.IsSensitive != nil {
		options.IsSensitive = pointers.NewBoolPtr(*otherOptions.IsSensitive)
	}
	if otherOptions.IsExpand != nil {
		options.IsExpand = pointers.NewBoolPtr(*otherOptions.IsExpand)
	}
	if otherOptions.SkipIfEmpty != nil {
		options.SkipIfEmpty = pointers.NewBoolPtr(*otherOptions.SkipIfEmpty)
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
	if otherOptions.Category != nil {
		options.Category = pointers.NewStringPtr(*otherOptions.Category)
	}
	if len(otherOptions.ValueOptions) > 0 {
		options.ValueOptions = otherOptions.ValueOptions
	}
	if otherOptions.IsRequired != nil {
		options.IsRequired = pointers.NewBoolPtr(*otherOptions.IsRequired)
	}
	if otherOptions.IsDontChangeValue != nil {
		options.IsDontChangeValue = pointers.NewBoolPtr(*otherOptions.IsDontChangeValue)
	}
	if otherOptions.IsTemplate != nil {
		options.IsTemplate = pointers.NewBoolPtr(*otherOptions.IsTemplate)
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

func MergeStepWith(step, otherStep stepmanModels.StepModel) (stepmanModels.StepModel, error) {
	if otherStep.Title != nil {
		step.Title = pointers.NewStringPtr(*otherStep.Title)
	}
	if otherStep.Summary != nil {
		step.Summary = pointers.NewStringPtr(*otherStep.Summary)
	}
	if otherStep.Description != nil {
		step.Description = pointers.NewStringPtr(*otherStep.Description)
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

	if otherStep.PublishedAt != nil {
		step.PublishedAt = pointers.NewTimePtr(*otherStep.PublishedAt)
	}
	if otherStep.Source != nil {
		step.Source = new(stepmanModels.StepSourceModel)

		if otherStep.Source.Git != "" {
			step.Source.Git = otherStep.Source.Git
		}
		if otherStep.Source.Commit != "" {
			step.Source.Commit = otherStep.Source.Commit
		}
	}
	if len(otherStep.AssetURLs) > 0 {
		step.AssetURLs = otherStep.AssetURLs
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
	if len(otherStep.Dependencies) > 0 {
		step.Dependencies = otherStep.Dependencies
	}
	if otherStep.Toolkit != nil {
		step.Toolkit = new(stepmanModels.StepToolkitModel)
		*step.Toolkit = *otherStep.Toolkit
	}
	if otherStep.Deps != nil && (len(otherStep.Deps.Brew) > 0 || len(otherStep.Deps.AptGet) > 0) {
		step.Deps = otherStep.Deps
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
	if otherStep.Timeout != nil {
		step.Timeout = pointers.NewIntPtr(*otherStep.Timeout)
	}
	if otherStep.NoOutputTimeout != nil {
		step.NoOutputTimeout = pointers.NewIntPtr(*otherStep.NoOutputTimeout)
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

func (stepListItem *StepListItemStepOrBundleModel) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	var key string
	for k := range raw {
		key = k
		break
	}

	if strings.HasPrefix(key, StepListItemStepBundleKeyPrefix) {
		var stepBundleItem StepListStepBundleItemModel
		if err := json.Unmarshal(b, &stepBundleItem); err != nil {
			return err
		}

		*stepListItem = map[string]interface{}{}
		for k, v := range stepBundleItem {
			(*stepListItem)[k] = v
		}
	} else {
		var stepItem StepListStepItemModel
		if err := json.Unmarshal(b, &stepItem); err != nil {
			return err
		}

		*stepListItem = map[string]interface{}{}
		for k, v := range stepItem {
			(*stepListItem)[k] = v
		}
	}

	return nil
}

func (stepListItem *StepListItemStepOrBundleModel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	var key string
	for k := range raw {
		key = k
		break
	}

	if strings.HasPrefix(key, StepListItemStepBundleKeyPrefix) {
		var stepBundleItem StepListStepBundleItemModel
		if err := unmarshal(&stepBundleItem); err != nil {
			return err
		}

		*stepListItem = map[string]interface{}{}
		for k, v := range stepBundleItem {
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

func (stepListItem *StepListItemStepOrBundleModel) GetKeyAndType() (string, StepListItemType, error) {
	if stepListItem == nil {
		return "", StepListItemTypeUnknown, nil
	}

	if len(*stepListItem) == 0 {
		return "", StepListItemTypeUnknown, errors.New("StepListItem does not contain a key-value pair")
	}

	if len(*stepListItem) > 1 {
		return "", StepListItemTypeUnknown, fmt.Errorf("StepListItem contains more than 1 key-value pair: %#v", *stepListItem)
	}

	for key := range *stepListItem {
		switch {
		case strings.HasPrefix(key, StepListItemStepBundleKeyPrefix):
			return strings.TrimPrefix(key, StepListItemStepBundleKeyPrefix), StepListItemTypeBundle, nil
		case key == StepListItemWithKey:
			return key, StepListItemTypeWith, fmt.Errorf("'with group' is not allowed in a step bundle's step list")
		default:
			return key, StepListItemTypeStep, nil
		}
	}

	return "", StepListItemTypeUnknown, nil
}

func (stepListItem *StepListItemStepOrBundleModel) GetBundle() (*StepBundleListItemModel, error) {
	if stepListItem == nil {
		return nil, fmt.Errorf("empty stepListItem")
	}

	for _, value := range *stepListItem {
		bundle, ok := value.(StepBundleListItemModel)
		if ok {
			return &bundle, nil
		}
		break
	}

	return nil, fmt.Errorf("stepListItem is not a StepBundle")
}

func (stepListItem *StepListItemStepOrBundleModel) GetStep() (*stepmanModels.StepModel, error) {
	if stepListItem == nil {
		return nil, fmt.Errorf("empty stepListItem")
	}

	for _, value := range *stepListItem {
		s, ok := value.(stepmanModels.StepModel)
		if ok {
			return &s, nil
		}
		break
	}

	return nil, fmt.Errorf("stepListItem is not a Step")
}

func (stepListItem *StepListItemModel) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	var key string
	for k := range raw {
		key = k
		break
	}

	if key == StepListItemWithKey {
		var withItem StepListWithItemModel
		if err := json.Unmarshal(b, &withItem); err != nil {
			return err
		}

		*stepListItem = map[string]interface{}{}
		for k, v := range withItem {
			(*stepListItem)[k] = v
		}
	} else if strings.HasPrefix(key, StepListItemStepBundleKeyPrefix) {
		var stepBundleItem StepListStepBundleItemModel
		if err := json.Unmarshal(b, &stepBundleItem); err != nil {
			return err
		}

		*stepListItem = map[string]interface{}{}
		for k, v := range stepBundleItem {
			(*stepListItem)[k] = v
		}
	} else {
		var stepItem StepListStepItemModel
		if err := json.Unmarshal(b, &stepItem); err != nil {
			return err
		}

		*stepListItem = map[string]interface{}{}
		for k, v := range stepItem {
			(*stepListItem)[k] = v
		}
	}

	return nil
}

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
	} else if strings.HasPrefix(key, StepListItemStepBundleKeyPrefix) {
		var stepBundleItem StepListStepBundleItemModel
		if err := unmarshal(&stepBundleItem); err != nil {
			return err
		}

		*stepListItem = map[string]interface{}{}
		for k, v := range stepBundleItem {
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

// ----------------------------
// --- StepIDData

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

type StepListItemType int

const (
	StepListItemTypeUnknown StepListItemType = iota
	StepListItemTypeStep
	StepListItemTypeWith
	StepListItemTypeBundle
)

func (stepListItem *StepListItemModel) GetKeyAndType() (string, StepListItemType, error) {
	if stepListItem == nil {
		return "", StepListItemTypeUnknown, nil
	}

	if len(*stepListItem) == 0 {
		return "", StepListItemTypeUnknown, errors.New("StepListItem does not contain a key-value pair")
	}

	if len(*stepListItem) > 1 {
		return "", StepListItemTypeUnknown, fmt.Errorf("StepListItem contains more than 1 key-value pair: %#v", *stepListItem)
	}

	for key := range *stepListItem {
		switch {
		case strings.HasPrefix(key, StepListItemStepBundleKeyPrefix):
			return strings.TrimPrefix(key, StepListItemStepBundleKeyPrefix), StepListItemTypeBundle, nil
		case key == StepListItemWithKey:
			return key, StepListItemTypeWith, nil
		default:
			return key, StepListItemTypeStep, nil
		}
	}

	return "", StepListItemTypeUnknown, nil
}

func (stepListItem *StepListItemModel) GetBundle() (*StepBundleListItemModel, error) {
	if stepListItem == nil {
		return nil, fmt.Errorf("empty stepListItem")
	}

	for _, value := range *stepListItem {
		bundle, ok := value.(StepBundleListItemModel)
		if ok {
			return &bundle, nil
		}
		break
	}

	return nil, fmt.Errorf("stepListItem is not a StepBundle")
}

func (stepListItem *StepListItemModel) GetWith() (*WithModel, error) {
	if stepListItem == nil {
		return nil, fmt.Errorf("empty stepListItem")
	}

	for _, value := range *stepListItem {
		with, ok := value.(WithModel)
		if ok {
			return &with, nil
		}
		break
	}

	return nil, fmt.Errorf("stepListItem is not a With")
}

func (stepListItem *StepListItemModel) GetStep() (*stepmanModels.StepModel, error) {
	if stepListItem == nil {
		return nil, fmt.Errorf("empty stepListItem")
	}

	for _, value := range *stepListItem {
		s, ok := value.(stepmanModels.StepModel)
		if ok {
			return &s, nil
		}
		break
	}

	return nil, fmt.Errorf("stepListItem is not a Step")
}
