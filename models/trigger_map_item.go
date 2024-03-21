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

type TriggerItemConditionStringValue string

type TriggerItemConditionRegexValue struct {
	Regex string `json:"regex" yaml:"regex"`
}

type TriggerItemType string

const (
	CodePushType    TriggerItemType = "code-push"
	PullRequestType TriggerItemType = "pull-request"
	TagPushType     TriggerItemType = "tag-push"
)

type TriggerMapItemModel struct {
	// Trigger Item shared properties
	Type       TriggerItemType `json:"type" yaml:"type"`
	Enabled    bool            `json:"enabled" yaml:"enabled"`
	PipelineID string          `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
	WorkflowID string          `json:"workflow,omitempty" yaml:"workflow,omitempty"`

	// Code Push Item conditions
	// TODO: introduce regex values
	PushBranch    string `json:"push_branch,omitempty" yaml:"push_branch,omitempty"`
	CommitMessage string `json:"commit_message" yaml:"commit_message"`
	ChangedFiles  string `json:"changed_files" yaml:"changed_files"`

	// Tag Push Item conditions
	Tag string `json:"tag,omitempty" yaml:"tag,omitempty"`

	// Pull Request Item conditions
	PullRequestSourceBranch string `json:"pull_request_source_branch,omitempty" yaml:"pull_request_source_branch,omitempty"`
	PullRequestTargetBranch string `json:"pull_request_target_branch,omitempty" yaml:"pull_request_target_branch,omitempty"`
	DraftPullRequestEnabled *bool  `json:"draft_pull_request_enabled,omitempty" yaml:"draft_pull_request_enabled,omitempty"`
	PullRequestLabel        string `json:"pull_request_label" yaml:"pull_request_label"`

	// Deprecated properties
	Pattern              string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	IsPullRequestAllowed bool   `json:"is_pull_request_allowed,omitempty" yaml:"is_pull_request_allowed,omitempty"`
}

func (triggerItem TriggerMapItemModel) Validate(idx int, workflows, pipelines []string) ([]string, error) {
	warnings, err := triggerItem.validateTarget(idx, workflows, pipelines)
	if err != nil {
		return warnings, err
	}

	if triggerItem.Pattern != "" {
		if err := triggerItem.validateTypeOfLegacyItem(idx); err != nil {
			return warnings, err
		}
	} else if triggerItem.Type == "" {
		if err := triggerItem.validateTypeOfItem(idx); err != nil {
			return warnings, err
		}
	} else {
		if err := triggerItem.validateTypeOfItemWithExplicitType(idx); err != nil {
			return warnings, err
		}
	}

	// TODO: validate condition values (regex or string literal)

	return warnings, nil
}

func (triggerItem TriggerMapItemModel) validateTypeOfLegacyItem(idx int) error {
	if triggerItem.PushBranch != "" {
		return fmt.Errorf("both pattern and push_branch defined in the %d. trigger item", idx+1)
	}
	if triggerItem.PullRequestSourceBranch != "" {
		return fmt.Errorf("both pattern and pull_request_source_branch defined in the %d. trigger item", idx+1)
	}
	if triggerItem.PullRequestTargetBranch != "" {
		return fmt.Errorf("both pattern and pull_request_target_branch defined in the %d. trigger item", idx+1)
	}
	if triggerItem.Tag != "" {
		return fmt.Errorf("both pattern and tag defined in the %d. trigger item", idx+1)
	}
	// TODO: check other fields
	return nil
}

func (triggerItem TriggerMapItemModel) validateTypeOfItem(idx int) error {
	if triggerItem.PushBranch != "" {
		if triggerItem.PullRequestSourceBranch != "" {
			return fmt.Errorf("both push_branch and pull_request_source_branch defined in the %d. trigger item", idx+1)
		}
		if triggerItem.PullRequestTargetBranch != "" {
			return fmt.Errorf("both push_branch and pull_request_target_branch defined in the %d. trigger item", idx+1)
		}
		if triggerItem.Tag != "" {
			return fmt.Errorf("both push_branch and tag defined in the %d. trigger item", idx+1)
		}
	} else if triggerItem.PullRequestSourceBranch != "" {
		if triggerItem.Tag != "" {
			return fmt.Errorf("both pull_request_source_branch and tag defined in the %d. trigger item", idx+1)
		}
	} else if triggerItem.PullRequestTargetBranch != "" {
		if triggerItem.Tag != "" {
			return fmt.Errorf("both pull_request_target_branch and tag defined in the %d. trigger item", idx+1)
		}
	} else if triggerItem.Tag == "" {
		return fmt.Errorf("no trigger condition defined defined in the %d. trigger item", idx+1)
	}

	return nil
}

func (triggerItem TriggerMapItemModel) validateTypeOfItemWithExplicitType(idx int) error {
	switch triggerItem.Type {
	case CodePushType:
		if triggerItem.PullRequestSourceBranch != "" {
			return fmt.Errorf("pull_request_source_branch defined for a push type trigger item in the %d. trigger item", idx+1)
		}
		if triggerItem.PullRequestTargetBranch != "" {
			return fmt.Errorf("pull_request_target_branch defined for a push type trigger item in the %d. trigger item", idx+1)
		}
		if triggerItem.Tag != "" {
			return fmt.Errorf("tag defined for a push type trigger item in the %d. trigger item", idx+1)
		}

		// TODO: check other fields too (label)
	case PullRequestType:
		if triggerItem.PushBranch != "" {
			return fmt.Errorf("push_branch defined for a pull request type trigger item in the %d. trigger item", idx+1)
		}
		if triggerItem.Tag != "" {
			return fmt.Errorf("tag defined for a pull request type item in the %d. trigger item", idx+1)
		}

		// TODO: check other fields too (file_changes, commit_message)
	case TagPushType:
		if triggerItem.PullRequestSourceBranch != "" {
			return fmt.Errorf("pull_request_source_branch defined for a tag type trigger item in the %d. trigger item", idx+1)
		}
		if triggerItem.PullRequestTargetBranch != "" {
			return fmt.Errorf("pull_request_target_branch defined for a tag type trigger item in the %d. trigger item", idx+1)
		}
		if triggerItem.PushBranch != "" {
			return fmt.Errorf("push_branch defined for a tag type trigger item in the %d. trigger item", idx+1)
		}
		// TODO: check other fields too (file_changes, commit_message)
	}
	return nil
}

func (triggerItem TriggerMapItemModel) validateTarget(idx int, workflows, pipelines []string) ([]string, error) {
	var warnings []string

	// Validate target
	if triggerItem.PipelineID != "" && triggerItem.WorkflowID != "" {
		return warnings, fmt.Errorf("both pipeline and workflow are defined as trigger target for the %d. trigger item", idx+1)
	}
	if triggerItem.PipelineID == "" && triggerItem.WorkflowID == "" {
		return warnings, fmt.Errorf("no pipeline nor workflow is defined as a trigger target for the %d. trigger item", idx+1)
	}

	if strings.HasPrefix(triggerItem.WorkflowID, "_") {
		warnings = append(warnings, fmt.Sprintf("utility workflow (%s) defined as trigger target for the %d. trigger item, but utility workflows can't be triggered directly", triggerItem.WorkflowID, idx+1))
	}

	if triggerItem.PipelineID != "" {
		if !sliceutil.IsStringInSlice(triggerItem.PipelineID, pipelines) {
			return warnings, fmt.Errorf("pipeline (%s) defined in the %d. trigger item, but does not exist", triggerItem.PipelineID, idx+1)
		}
	} else {
		if !sliceutil.IsStringInSlice(triggerItem.WorkflowID, workflows) {
			return warnings, fmt.Errorf("workflow (%s) defined in the %d. trigger item, but does not exist", triggerItem.WorkflowID, idx+1)
		}
	}

	return warnings, nil
}

func (triggerItem TriggerMapItemModel) MatchWithParams(pushBranch, prSourceBranch, prTargetBranch string, prReadyState PullRequestReadyState, tag string) (bool, error) {
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
			match := glob.Glob(migratedTriggerItem.Tag, tag)
			return match, nil
		}
	}

	return false, nil
}

func (triggerItem TriggerMapItemModel) IsDraftPullRequestEnabled() bool {
	draftPullRequestEnabled := defaultDraftPullRequestEnabled
	if triggerItem.DraftPullRequestEnabled != nil {
		draftPullRequestEnabled = *triggerItem.DraftPullRequestEnabled
	}
	return draftPullRequestEnabled
}

func (triggerItem TriggerMapItemModel) String() string {
	str := ""

	rv := reflect.Indirect(reflect.ValueOf(&triggerItem))
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("yaml")
		tag = strings.TrimSuffix(tag, ",omitempty")
		if tag == "pipeline" || tag == "workflow" || tag == "type" || tag == "enabled" {
			continue
		}

		value := rv.FieldByName(field.Name).Interface()
		str += fmt.Sprintf("%s:%v", tag, value)

		if i < rt.NumField()-1 {
			str += "&"
		}
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
