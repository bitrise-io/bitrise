package models

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/ryanuber/go-glob"
)

type TriggerEventType string

const (
	TriggerEventTypeCodePush    TriggerEventType = "code-push"
	TriggerEventTypePullRequest TriggerEventType = "pull-request"
	TriggerEventTypeTag         TriggerEventType = "tag"
	TriggerEventTypeUnknown     TriggerEventType = "unknown"
)

type PullRequestReadyState string

const (
	PullRequestReadyStateDraft                     PullRequestReadyState = "draft"
	PullRequestReadyStateReadyForReview            PullRequestReadyState = "ready_for_review"
	PullRequestReadyStateConvertedToReadyForReview PullRequestReadyState = "converted_to_ready_for_review"
)

const defaultDraftPullRequestEnabled = true

type TriggerItemType string

const (
	CodePushType    TriggerItemType = "push"
	PullRequestType TriggerItemType = "pull_request"
	TagPushType     TriggerItemType = "tag"
)

type TriggerMapItemModel struct {
	// Trigger Item shared properties
	Type       TriggerItemType `json:"type,omitempty" yaml:"type,omitempty"`
	Enabled    *bool           `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	PipelineID string          `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
	WorkflowID string          `json:"workflow,omitempty" yaml:"workflow,omitempty"`

	// Code Push Item conditions
	PushBranch    interface{} `json:"push_branch,omitempty" yaml:"push_branch,omitempty"`
	CommitMessage interface{} `json:"commit_message,omitempty" yaml:"commit_message,omitempty"`
	ChangedFiles  interface{} `json:"changed_files,omitempty" yaml:"changed_files,omitempty"`

	// Tag Push Item conditions
	Tag interface{} `json:"tag,omitempty" yaml:"tag,omitempty"`

	// Pull Request Item conditions
	PullRequestSourceBranch interface{} `json:"pull_request_source_branch,omitempty" yaml:"pull_request_source_branch,omitempty"`
	PullRequestTargetBranch interface{} `json:"pull_request_target_branch,omitempty" yaml:"pull_request_target_branch,omitempty"`
	DraftPullRequestEnabled *bool       `json:"draft_pull_request_enabled,omitempty" yaml:"draft_pull_request_enabled,omitempty"`
	PullRequestLabel        interface{} `json:"pull_request_label,omitempty" yaml:"pull_request_label,omitempty"`

	// Deprecated properties
	Pattern              string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	IsPullRequestAllowed bool   `json:"is_pull_request_allowed,omitempty" yaml:"is_pull_request_allowed,omitempty"`
}

func (item TriggerMapItemModel) MatchWithParams(pushBranch, prSourceBranch, prTargetBranch string, prReadyState PullRequestReadyState, tag string) (bool, error) {
	paramsEventType, err := triggerEventType(pushBranch, prSourceBranch, prTargetBranch, tag)
	if err != nil {
		return false, err
	}

	migratedTriggerItems := []TriggerMapItemModel{item}
	if item.Pattern != "" {
		migratedTriggerItems = migrateDeprecatedTriggerItem(item)
	}

	for _, migratedTriggerItem := range migratedTriggerItems {
		itemEventType, err := triggerEventType(stringFromTriggerCondition(migratedTriggerItem.PushBranch),
			stringFromTriggerCondition(migratedTriggerItem.PullRequestSourceBranch),
			stringFromTriggerCondition(migratedTriggerItem.PullRequestTargetBranch),
			stringFromTriggerCondition(migratedTriggerItem.Tag))
		if err != nil {
			return false, err
		}

		if paramsEventType != itemEventType {
			continue
		}

		switch itemEventType {
		case TriggerEventTypeCodePush:
			match := glob.Glob(stringFromTriggerCondition(migratedTriggerItem.PushBranch), pushBranch)
			return match, nil
		case TriggerEventTypePullRequest:
			sourceMatch := false
			if stringFromTriggerCondition(migratedTriggerItem.PullRequestSourceBranch) == "" {
				sourceMatch = true
			} else {
				sourceMatch = glob.Glob(stringFromTriggerCondition(migratedTriggerItem.PullRequestSourceBranch), prSourceBranch)
			}

			targetMatch := false
			if stringFromTriggerCondition(migratedTriggerItem.PullRequestTargetBranch) == "" {
				targetMatch = true
			} else {
				targetMatch = glob.Glob(stringFromTriggerCondition(migratedTriggerItem.PullRequestTargetBranch), prTargetBranch)
			}

			// When a PR is converted to ready for review:
			// - if draft PR trigger is enabled, this event is just a status change on the PR
			// 	 and the given status of the code base already triggered a build.
			// - if draft PR trigger is disabled, the given status of the code base didn't trigger a build yet.
			stateMismatch := false
			if migratedTriggerItem.IsDraftPullRequestEnabled() {
				if prReadyState == PullRequestReadyStateConvertedToReadyForReview {
					stateMismatch = true
				}
			} else {
				if prReadyState == PullRequestReadyStateDraft {
					stateMismatch = true
				}
			}

			return sourceMatch && targetMatch && !stateMismatch, nil
		case TriggerEventTypeTag:
			match := glob.Glob(stringFromTriggerCondition(migratedTriggerItem.Tag), tag)
			return match, nil
		}
	}

	return false, nil
}

func (item TriggerMapItemModel) IsDraftPullRequestEnabled() bool {
	draftPullRequestEnabled := defaultDraftPullRequestEnabled
	if item.DraftPullRequestEnabled != nil {
		draftPullRequestEnabled = *item.DraftPullRequestEnabled
	}
	return draftPullRequestEnabled
}

// Normalized casts trigger item values from map[interface{}]interface{} to map[string]interface{}
// to support JSON marshalling of the bitrise.yml.
func (item TriggerMapItemModel) Normalized(idx int) (TriggerMapItemModel, error) {
	mapInterface, ok := item.PushBranch.(map[interface{}]interface{})
	if ok {
		value, err := castInterfaceKeysToStringKeys(idx, "push_branch", mapInterface)
		if err != nil {
			return TriggerMapItemModel{}, err
		}
		item.PushBranch = value
	}

	mapInterface, ok = item.CommitMessage.(map[interface{}]interface{})
	if ok {
		value, err := castInterfaceKeysToStringKeys(idx, "commit_message", mapInterface)
		if err != nil {
			return TriggerMapItemModel{}, err
		}
		item.CommitMessage = value
	}

	mapInterface, ok = item.ChangedFiles.(map[interface{}]interface{})
	if ok {
		value, err := castInterfaceKeysToStringKeys(idx, "changed_files", mapInterface)
		if err != nil {
			return TriggerMapItemModel{}, err
		}
		item.ChangedFiles = value
	}

	mapInterface, ok = item.Tag.(map[interface{}]interface{})
	if ok {
		value, err := castInterfaceKeysToStringKeys(idx, "tag", mapInterface)
		if err != nil {
			return TriggerMapItemModel{}, err
		}
		item.Tag = value
	}

	mapInterface, ok = item.PullRequestSourceBranch.(map[interface{}]interface{})
	if ok {
		value, err := castInterfaceKeysToStringKeys(idx, "pull_request_source_branch", mapInterface)
		if err != nil {
			return TriggerMapItemModel{}, err
		}
		item.PullRequestSourceBranch = value
	}

	mapInterface, ok = item.PullRequestTargetBranch.(map[interface{}]interface{})
	if ok {
		value, err := castInterfaceKeysToStringKeys(idx, "pull_request_target_branch", mapInterface)
		if err != nil {
			return TriggerMapItemModel{}, err
		}
		item.PullRequestTargetBranch = value
	}

	mapInterface, ok = item.PullRequestLabel.(map[interface{}]interface{})
	if ok {
		value, err := castInterfaceKeysToStringKeys(idx, "pull_request_label", mapInterface)
		if err != nil {
			return TriggerMapItemModel{}, err
		}
		item.PullRequestLabel = value
	}

	return item, nil
}

func (item TriggerMapItemModel) Validate(idx int, workflows, pipelines []string) ([]string, error) {
	warnings, err := item.validateTarget(idx, workflows, pipelines)
	if err != nil {
		return warnings, err
	}

	if item.Pattern != "" {
		if err := item.validateLegacyItemType(idx); err != nil {
			return warnings, err
		}
	} else {
		if err := item.validateType(idx); err != nil {
			return warnings, err
		}
		if err := item.validateConditionValues(idx); err != nil {
			return warnings, err
		}
	}

	return warnings, nil
}

func (item TriggerMapItemModel) validateTarget(idx int, workflows, pipelines []string) ([]string, error) {
	var warnings []string

	// Validate target
	if item.PipelineID != "" && item.WorkflowID != "" {
		return warnings, fmt.Errorf("both pipeline and workflow are defined as trigger target for the %d. trigger item", idx+1)
	}
	if item.PipelineID == "" && item.WorkflowID == "" {
		return warnings, fmt.Errorf("no pipeline nor workflow is defined as a trigger target for the %d. trigger item", idx+1)
	}

	if strings.HasPrefix(item.WorkflowID, "_") {
		warnings = append(warnings, fmt.Sprintf("utility workflow (%s) defined as trigger target for the %d. trigger item, but utility workflows can't be triggered directly", item.WorkflowID, idx+1))
	}

	if item.PipelineID != "" {
		if !sliceutil.IsStringInSlice(item.PipelineID, pipelines) {
			return warnings, fmt.Errorf("pipeline (%s) defined in the %d. trigger item, but does not exist", item.PipelineID, idx+1)
		}
	} else {
		if !sliceutil.IsStringInSlice(item.WorkflowID, workflows) {
			return warnings, fmt.Errorf("workflow (%s) defined in the %d. trigger item, but does not exist", item.WorkflowID, idx+1)
		}
	}

	return warnings, nil
}

func (item TriggerMapItemModel) validateLegacyItemType(idx int) error {
	if err := item.validateNoCodePushConditionsSet(idx, "pattern"); err != nil {
		return err
	}
	if err := item.validateNoTagPushConditionsSet(idx, "pattern"); err != nil {
		return err
	}
	if err := item.validateNoPullRequestConditionsSet(idx, "pattern"); err != nil {
		return err
	}
	return nil
}

func (item TriggerMapItemModel) validateType(idx int) error {
	if item.Type != "" {
		if !sliceutil.IsStringInSlice(string(item.Type), []string{string(CodePushType), string(PullRequestType), string(TagPushType)}) {
			return fmt.Errorf("invalid type (%s) set in the %d. trigger item, valid types are: push, pull_request and tag", item.Type, idx+1)
		}
	}

	if isStringLiteralOrRegexSet(item.PushBranch) || item.Type == CodePushType {
		var field string
		if item.Type != "" {
			field = fmt.Sprintf("type: %s", item.Type)
		} else {
			field = "push_branch"
		}

		if err := item.validateNoTagPushConditionsSet(idx, field); err != nil {
			return err
		}
		if err := item.validateNoPullRequestConditionsSet(idx, field); err != nil {
			return err
		}

		return nil
	} else if isStringLiteralOrRegexSet(item.PullRequestSourceBranch) || isStringLiteralOrRegexSet(item.PullRequestTargetBranch) || item.Type == PullRequestType {
		var field string
		if item.Type != "" {
			field = fmt.Sprintf("type: %s", item.Type)
		} else {
			if isStringLiteralOrRegexSet(item.PullRequestSourceBranch) {
				field = "pull_request_source_branch"
			}
			if isStringLiteralOrRegexSet(item.PullRequestTargetBranch) {
				if field != "" {
					field += " and "
				}
				field += "pull_request_target_branch"
			}
		}

		if err := item.validateNoCodePushConditionsSet(idx, field); err != nil {
			return err
		}
		if err := item.validateNoTagPushConditionsSet(idx, field); err != nil {
			return err
		}

		return nil
	} else if isStringLiteralOrRegexSet(item.Tag) || item.Type == TagPushType {
		var field string
		if item.Type != "" {
			field = fmt.Sprintf("type: %s", item.Type)
		} else {
			field = "tag"
		}

		if err := item.validateNoCodePushConditionsSet(idx, field); err != nil {
			return err
		}
		if err := item.validateNoPullRequestConditionsSet(idx, field); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("no type or trigger condition defined in the %d. trigger item", idx+1)
}

func (item TriggerMapItemModel) validateConditionValues(idx int) error {
	if err := validateStringOrRegexType(idx, "push_branch", item.PushBranch); err != nil {
		return err
	}
	if err := validateStringOrRegexType(idx, "commit_message", item.CommitMessage); err != nil {
		return err
	}
	if err := validateStringOrRegexType(idx, "changed_files", item.ChangedFiles); err != nil {
		return err
	}

	if err := validateStringOrRegexType(idx, "tag", item.Tag); err != nil {
		return err
	}

	if err := validateStringOrRegexType(idx, "pull_request_source_branch", item.PullRequestSourceBranch); err != nil {
		return err
	}
	if err := validateStringOrRegexType(idx, "pull_request_target_branch", item.PullRequestTargetBranch); err != nil {
		return err
	}
	if err := validateStringOrRegexType(idx, "pull_request_label", item.PullRequestLabel); err != nil {
		return err
	}
	return nil
}

func (item TriggerMapItemModel) validateNoCodePushConditionsSet(idx int, field string) error {
	if isStringLiteralOrRegexSet(item.PushBranch) {
		return fmt.Errorf("both %s and push_branch defined in the %d. trigger item", field, idx+1)
	}
	if isStringLiteralOrRegexSet(item.CommitMessage) {
		return fmt.Errorf("both %s and commit_message defined in the %d. trigger item", field, idx+1)
	}
	if isStringLiteralOrRegexSet(item.ChangedFiles) {
		return fmt.Errorf("both %s and changed_files defined in the %d. trigger item", field, idx+1)
	}
	return nil
}

func (item TriggerMapItemModel) validateNoTagPushConditionsSet(idx int, field string) error {
	if isStringLiteralOrRegexSet(item.Tag) {
		return fmt.Errorf("both %s and tag defined in the %d. trigger item", field, idx+1)
	}
	return nil
}

func (item TriggerMapItemModel) validateNoPullRequestConditionsSet(idx int, field string) error {
	if isStringLiteralOrRegexSet(item.PullRequestSourceBranch) {
		return fmt.Errorf("both %s and pull_request_source_branch defined in the %d. trigger item", field, idx+1)
	}
	if isStringLiteralOrRegexSet(item.PullRequestTargetBranch) {
		return fmt.Errorf("both %s and pull_request_target_branch defined in the %d. trigger item", field, idx+1)
	}
	//nolint:gosimple
	if item.IsDraftPullRequestEnabled() != defaultDraftPullRequestEnabled {
		return fmt.Errorf("both %s and draft_pull_request_enabled defined in the %d. trigger item", field, idx+1)
	}
	if isStringLiteralOrRegexSet(item.PullRequestLabel) {
		return fmt.Errorf("both %s and pull_request_label defined in the %d. trigger item", field, idx+1)
	}
	return nil
}

func (item TriggerMapItemModel) conditionsString() string {
	str := ""

	rv := reflect.Indirect(reflect.ValueOf(&item))
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("yaml")
		tag = strings.TrimSuffix(tag, ",omitempty")
		if tag == "pipeline" || tag == "workflow" || tag == "type" || tag == "enabled" {
			continue
		}

		value := rv.FieldByName(field.Name).Interface()
		if value == nil {
			continue
		}

		if tag == "draft_pull_request_enabled" {
			if boolPtrValue, ok := value.(*bool); ok {
				//nolint:gosimple
				if boolPtrValue == nil || *boolPtrValue == defaultDraftPullRequestEnabled {
					continue
				}
				value = *boolPtrValue
			}
		}

		if strValue, ok := value.(string); ok {
			if strValue == "" {
				continue
			}
		}

		if tag == "is_pull_request_allowed" {
			if boolPtrValue, ok := value.(bool); ok {
				if !boolPtrValue {
					continue
				}
			}
		}

		if str != "" {
			str += " & "
		}
		str += fmt.Sprintf("%s: %v", tag, value)
	}

	if str == "" && item.Type != "" {
		// Trigger Item without any condition is valid,
		// this case we use the type to differentiate push, pull-request and tag items
		str = "type: " + string(item.Type)
	}

	return str
}

func validateStringOrRegexType(idx int, field string, value interface{}) error {
	if value == nil {
		return nil
	}
	_, ok := value.(string)
	if ok {
		return nil
	}

	valueMap, ok := value.(map[interface{}]interface{})
	if ok {
		if len(valueMap) != 1 {
			return fmt.Errorf("single 'regex' key is expected for regex condition in %s field of the %d. trigger item", field, idx+1)
		}

		_, ok := valueMap["regex"]
		if !ok {
			return fmt.Errorf("'regex' key is expected for regex condition in %s field of the %d. trigger item", field, idx+1)
		}

		return nil
	}

	valueInterfaceMap, ok := value.(map[string]interface{})
	if ok {
		if len(valueInterfaceMap) != 1 {
			return fmt.Errorf("single 'regex' key is expected for regex condition in %s field of the %d. trigger item", field, idx+1)
		}

		regex, ok := valueInterfaceMap["regex"]
		if !ok {
			return fmt.Errorf("'regex' key is expected for regex condition in %s field of the %d. trigger item", field, idx+1)
		}

		_, ok = regex.(string)
		if !ok {
			return fmt.Errorf("'regex' key is expected to have a string value in %s field of the %d. trigger item", field, idx+1)
		}

		return nil
	}

	valueStringMap, ok := value.(map[string]string)
	if ok {
		if len(valueStringMap) != 1 {
			return fmt.Errorf("single 'regex' key is expected for regex condition in %s field of the %d. trigger item", field, idx+1)
		}

		_, ok := valueStringMap["regex"]
		if !ok {
			return fmt.Errorf("'regex' key is expected for regex condition in %s field of the %d. trigger item", field, idx+1)
		}

		return nil
	}

	return fmt.Errorf("string literal or regex value is expected for %s in the %d. trigger item", field, idx+1)
}

func stringFromTriggerCondition(value interface{}) string {
	if value == nil {
		return ""
	}
	return value.(string)
}

func stringLiteralOrRegex(value interface{}) string {
	if value == nil {
		return ""
	}
	str, ok := value.(string)
	if ok {
		return string(str)
	}

	valueMap, ok := value.(map[interface{}]interface{})
	if ok {
		regex, ok := valueMap["regex"]
		if ok {
			return regex.(string)
		}
	}

	valueInterfaceMap, ok := value.(map[string]interface{})
	if ok {
		regex, ok := valueInterfaceMap["regex"]
		if ok {
			return regex.(string)
		}
	}

	valueStringMap, ok := value.(map[string]string)
	if ok {
		return valueStringMap["regex"]
	}

	return ""
}

func isStringLiteralOrRegexSet(value interface{}) bool {
	return stringLiteralOrRegex(value) != ""
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

func castInterfaceKeysToStringKeys(idx int, field string, value map[interface{}]interface{}) (map[string]interface{}, error) {
	mapString := map[string]interface{}{}
	for key, value := range value {
		keyStr, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("%s should be a string literal or a hash with a single 'regex' key in the %d. trigger item", field, idx+1)
		}
		mapString[keyStr] = value
	}
	return mapString, nil
}
