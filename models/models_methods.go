package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"
	stepmanModels "github.com/bitrise-io/stepman/models"
	"github.com/ryanuber/go-glob"
)

func (triggerItem TriggerMapItemModel) String(printWorkflow bool) string {
	str := ""

	if triggerItem.PushBranch != "" {
		str = fmt.Sprintf("push_branch: %s", triggerItem.PushBranch)
	}

	if triggerItem.PullRequestSourceBranch != "" || triggerItem.PullRequestTargetBranch != "" {
		if str != "" {
			str += " "
		}

		if triggerItem.PullRequestSourceBranch != "" {
			str += fmt.Sprintf("pull_request_source_branch: %s", triggerItem.PullRequestSourceBranch)
		}
		if triggerItem.PullRequestTargetBranch != "" {
			if triggerItem.PullRequestSourceBranch != "" {
				str += " && "
			}

			str += fmt.Sprintf("pull_request_target_branch: %s", triggerItem.PullRequestTargetBranch)
		}
	}

	if triggerItem.Tag != "" {
		if str != "" {
			str += " "
		}

		str += fmt.Sprintf("tag: %s", triggerItem.Tag)
	}

	if triggerItem.Pattern != "" {
		if str != "" {
			str += " "
		}

		str += fmt.Sprintf("pattern: %s && is_pull_request_allowed: %v", triggerItem.Pattern, triggerItem.IsPullRequestAllowed)
	}

	if printWorkflow {
		str += fmt.Sprintf(" -> workflow: %s", triggerItem.WorkflowID)
	}

	return str
}

func triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag string) (TriggerEventType, error) {
	if pushBranch != "" {
		// Ensure not mixed with code-push event
		if prSourceBranch != "" {
			return TriggerEventTypeUnknown, fmt.Errorf("push_branch (%s) selects code-push trigger event, but pull_request_source_branch (%s) also provided", pushBranch, prSourceBranch)
		}
		if prTargetBranch != "" {
			return TriggerEventTypeUnknown, fmt.Errorf("push_branch (%s) selects code-push trigger event, but pull_request_target_branch (%s) also provided", pushBranch, prTargetBranch)
		}

		// Ensure not mixed with tag event
		if tag != "" {
			return TriggerEventTypeUnknown, fmt.Errorf("push_branch (%s) selects code-push trigger event, but tag (%s) also provided", pushBranch, tag)
		}

		return TriggerEventTypeCodePush, nil
	} else if prSourceBranch != "" || prTargetBranch != "" {
		// Ensure not mixed with tag event
		if tag != "" {
			return TriggerEventTypeUnknown, fmt.Errorf("pull_request_source_branch (%s) and pull_request_target_branch (%s) selects pull-request trigger event, but tag (%s) also provided", prSourceBranch, prTargetBranch, tag)
		}

		return TriggerEventTypePullRequest, nil
	} else if tag != "" {
		return TriggerEventTypeTag, nil
	}

	return TriggerEventTypeUnknown, fmt.Errorf("failed to determin trigger event from params: push-branch: %s, pr-source-branch: %s, pr-target-branch: %s, tag: %s", pushBranch, prSourceBranch, prTargetBranch, tag)
}

func migrateDeprecatedTriggerItem(triggerItem TriggerMapItemModel) []TriggerMapItemModel {
	migratedItems := []TriggerMapItemModel{
		TriggerMapItemModel{
			PushBranch: triggerItem.Pattern,
			WorkflowID: triggerItem.WorkflowID,
		},
	}
	if triggerItem.IsPullRequestAllowed {
		migratedItems = append(migratedItems, TriggerMapItemModel{
			PullRequestSourceBranch: triggerItem.Pattern,
			WorkflowID:              triggerItem.WorkflowID,
		})
	}
	return migratedItems
}

// MatchWithParams ...
func (triggerItem TriggerMapItemModel) MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag string) (bool, error) {
	paramsEventType, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
	if err != nil {
		return false, err
	}

	migratedTriggerItems := []TriggerMapItemModel{triggerItem}
	if triggerItem.Pattern != "" {
		migratedTriggerItems = migrateDeprecatedTriggerItem(triggerItem)
	}

	for _, migratedTriggerItem := range migratedTriggerItems {
		itemEventType, err := triggerEventType(migratedTriggerItem.PushBranch, migratedTriggerItem.PullRequestSourceBranch, migratedTriggerItem.PullRequestTargetBranch, migratedTriggerItem.Tag)
		if err != nil {
			return false, err
		}

		if paramsEventType != itemEventType {
			continue
		}

		switch itemEventType {
		case TriggerEventTypeCodePush:
			match := glob.Glob(migratedTriggerItem.PushBranch, pushBranch)
			return match, nil
		case TriggerEventTypePullRequest:
			sourceMatch := false
			if migratedTriggerItem.PullRequestSourceBranch == "" {
				sourceMatch = true
			} else {
				sourceMatch = glob.Glob(migratedTriggerItem.PullRequestSourceBranch, prSourceBranch)
			}

			targetMatch := false
			if migratedTriggerItem.PullRequestTargetBranch == "" {
				targetMatch = true
			} else {
				targetMatch = glob.Glob(migratedTriggerItem.PullRequestTargetBranch, prTargetBranch)
			}

			return (sourceMatch && targetMatch), nil
		case TriggerEventTypeTag:
			match := glob.Glob(migratedTriggerItem.Tag, tag)
			return match, nil
		}
	}

	return false, nil
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

// ----------------------------
// --- Normalize

// Normalize ...
func (workflow *WorkflowModel) Normalize() error {
	for _, env := range workflow.Environments {
		if err := env.Normalize(); err != nil {
			return err
		}
	}

	for _, stepListItem := range workflow.Steps {
		stepID, step, err := GetStepIDStepDataPair(stepListItem)
		if err != nil {
			return err
		}
		if err := step.Normalize(); err != nil {
			return err
		}
		stepListItem[stepID] = step
	}

	return nil
}

// Normalize ...
func (app *AppModel) Normalize() error {
	for _, env := range app.Environments {
		if err := env.Normalize(); err != nil {
			return err
		}
	}
	return nil
}

// Normalize ...
func (config *BitriseDataModel) Normalize() error {
	if err := config.App.Normalize(); err != nil {
		return err
	}

	for _, workflow := range config.Workflows {
		if err := workflow.Normalize(); err != nil {
			return err
		}
	}

	return nil
}

// ----------------------------
// --- Validate

// Validate ...
func (workflow *WorkflowModel) Validate() ([]string, error) {
	for _, env := range workflow.Environments {
		if err := env.Validate(); err != nil {
			return []string{}, err
		}
	}

	warnings := []string{}
	for _, stepListItem := range workflow.Steps {
		stepID, step, err := GetStepIDStepDataPair(stepListItem)
		if err != nil {
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

		stepListItem[stepID] = step
	}

	return warnings, nil
}

// Validate ...
func (app *AppModel) Validate() error {
	for _, env := range app.Environments {
		if err := env.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate ...
func (triggerItem TriggerMapItemModel) Validate() error {
	if triggerItem.WorkflowID == "" {
		return fmt.Errorf("invalid trigger item: (%s) -> (%s), error: empty workflow id", triggerItem.Pattern, triggerItem.WorkflowID)
	}

	if triggerItem.Pattern == "" {
		_, err := triggerEventType(triggerItem.PushBranch, triggerItem.PullRequestSourceBranch, triggerItem.PullRequestTargetBranch, triggerItem.Tag)
		if err != nil {
			return fmt.Errorf("trigger map item (%s) validate failed, error: %s", triggerItem.String(true), err)
		}
	} else if triggerItem.PushBranch != "" ||
		triggerItem.PullRequestSourceBranch != "" || triggerItem.PullRequestTargetBranch != "" || triggerItem.Tag != "" {
		return fmt.Errorf("deprecated trigger item (pattern defined), mixed with trigger params (push_branch: %s, pull_request_source_branch: %s, pull_request_target_branch: %s, tag: %s)", triggerItem.PushBranch, triggerItem.PullRequestSourceBranch, triggerItem.PullRequestTargetBranch, triggerItem.Tag)
	}

	return nil
}

// Validate ...
func (triggerMap TriggerMapModel) Validate() error {
	for _, item := range triggerMap {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func checkDuplicatedTriggerMapItems(triggerMap TriggerMapModel) error {
	triggeTypeItemMap := map[string][]TriggerMapItemModel{}

	for _, triggerItem := range triggerMap {
		if triggerItem.Pattern == "" {
			triggerType, err := triggerEventType(triggerItem.PushBranch, triggerItem.PullRequestSourceBranch, triggerItem.PullRequestTargetBranch, triggerItem.Tag)
			if err != nil {
				return fmt.Errorf("trigger map item (%s) validate failed, error: %s", triggerItem, err)
			}

			triggerItems := triggeTypeItemMap[string(triggerType)]

			for _, item := range triggerItems {
				switch triggerType {
				case TriggerEventTypeCodePush:
					if triggerItem.PushBranch == item.PushBranch {
						return fmt.Errorf("duplicated trigger item found (%s)", triggerItem.String(false))
					}
				case TriggerEventTypePullRequest:
					if triggerItem.PullRequestSourceBranch == item.PullRequestSourceBranch &&
						triggerItem.PullRequestTargetBranch == item.PullRequestTargetBranch {
						return fmt.Errorf("duplicated trigger item found (%s)", triggerItem.String(false))
					}
				case TriggerEventTypeTag:
					if triggerItem.Tag == item.Tag {
						return fmt.Errorf("duplicated trigger item found (%s)", triggerItem.String(false))
					}
				}
			}

			triggerItems = append(triggerItems, triggerItem)
			triggeTypeItemMap[string(triggerType)] = triggerItems
		} else if triggerItem.Pattern != "" {
			triggerItems := triggeTypeItemMap["deprecated"]

			for _, item := range triggerItems {
				if triggerItem.Pattern == item.Pattern &&
					triggerItem.IsPullRequestAllowed == item.IsPullRequestAllowed {
					return fmt.Errorf("duplicated trigger item found (%s)", triggerItem.String(false))
				}
			}

			triggerItems = append(triggerItems, triggerItem)
			triggeTypeItemMap["deprecated"] = triggerItems
		}
	}

	return nil
}

// Validate ...
func (config *BitriseDataModel) Validate() ([]string, error) {
	warnings := []string{}

	if config.FormatVersion == "" {
		return warnings, fmt.Errorf("missing format_version")
	}

	if err := config.TriggerMap.Validate(); err != nil {
		return warnings, err
	}

	for _, triggerMapItem := range config.TriggerMap {
		if strings.HasPrefix(triggerMapItem.WorkflowID, "_") {
			warnings = append(warnings, fmt.Sprintf("workflow (%s) defined in trigger item (%s), but utility workflows can't be triggered directly", triggerMapItem.WorkflowID, triggerMapItem.String(true)))
		}

		found := false

		for workflowID := range config.Workflows {
			if workflowID == triggerMapItem.WorkflowID {
				found = true
			}
		}

		if !found {
			return warnings, fmt.Errorf("workflow (%s) defined in trigger item (%s), but does not exist", triggerMapItem.WorkflowID, triggerMapItem.String(true))
		}
	}

	if err := checkDuplicatedTriggerMapItems(config.TriggerMap); err != nil {
		return warnings, err
	}

	if err := config.App.Validate(); err != nil {
		return warnings, err
	}

	for ID, workflow := range config.Workflows {
		if ID == "" {
			warnings = append(warnings, fmt.Sprintf("invalid workflow ID (%s): empty", ID))
		}

		r := regexp.MustCompile(`[A-Za-z0-9-_.]+`)
		if find := r.FindString(ID); find != ID {
			warnings = append(warnings, fmt.Sprintf("invalid workflow ID (%s): doesn't conform to: [A-Za-z0-9-_.]", ID))
		}

		warns, err := workflow.Validate()
		warnings = append(warnings, warns...)
		if err != nil {
			return warnings, err
		}

		if err := checkWorkflowReferenceCycle(ID, workflow, *config, []string{}); err != nil {
			return warnings, err
		}
	}

	return warnings, nil
}

// ----------------------------
// --- FillMissingDefaults

// FillMissingDefaults ...
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

// FillMissingDefaults ...
func (app *AppModel) FillMissingDefaults() error {
	for _, env := range app.Environments {
		if err := env.FillMissingDefaults(); err != nil {
			return err
		}
	}
	return nil
}

// FillMissingDefaults ...
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
	if options.IsTemplate != nil {
		if *options.IsTemplate == envmanModels.DefaultIsTemplate {
			options.IsTemplate = nil
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

// MergeStepWith ...
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
	if otherStep.Deps != nil && (len(otherStep.Deps.Brew) > 0 || len(otherStep.Deps.AptGet) > 0 || len(otherStep.Deps.CheckOnly) > 0) {
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
		return "", stepmanModels.StepModel{}, errors.New("StepListItem contains more than 1 key-value pair")
	}
	for key, value := range stepListItem {
		return key, value, nil
	}
	return "", stepmanModels.StepModel{}, errors.New("StepListItem does not contain a key-value pair")
}

// CreateStepIDDataFromString ...
// compositeVersionStr examples:
//  * local path:
//    * path::~/path/to/step/dir
//  * direct git url and branch or tag:
//    * git::https://github.com/bitrise-io/steps-timestamp.git@master
//  * Steplib independent step:
//    * _::https://github.com/bitrise-io/steps-bash-script.git@2.0.0:
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

// IsUniqueResourceID : true if this ID is a unique resource ID, which is true
// if the ID refers to the exact same step code/data every time.
// Practically, this is only true for steps from StepLibrary collections,
// a local path or direct git step ID is never guaranteed to identify the
// same resource every time, the step's behaviour can change at every execution!
//
// __If the ID is a Unique Resource ID then the step can be cached (locally)__,
// as it won't change between subsequent step execution.
func (sIDData StepIDData) IsUniqueResourceID() bool {
	switch sIDData.SteplibSource {
	case "path":
		return false
	case "git":
		return false
	case "_":
		return false
	case "":
		return false
	}

	// in any other case, it's a StepLib URL
	// but it's only unique if StepID and Step Version are all defined!
	if len(sIDData.IDorURI) > 0 && len(sIDData.Version) > 0 {
		return true
	}

	// in every other case, it's not unique, not even if it's from a StepLib
	return false
}

// ----------------------------
// --- BuildRunResults

// IsStepLibUpdated ...
func (buildRes BuildRunResultsModel) IsStepLibUpdated(stepLib string) bool {
	return (buildRes.StepmanUpdates[stepLib] > 0)
}

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
