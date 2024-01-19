package models

import (
	"fmt"
	"strings"

	"github.com/ryanuber/go-glob"
)

type TriggerEventType string

const (
	TriggerEventTypeCodePush    TriggerEventType = "code-push"
	TriggerEventTypePullRequest TriggerEventType = "pull-request"
	TriggerEventTypeTag         TriggerEventType = "tag"
	TriggerEventTypeUnknown     TriggerEventType = "unknown"
)

const defaultDraftPullRequestEnabled = true

type TriggerMapItemModel struct {
	// Trigger target
	PipelineID string `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
	WorkflowID string `json:"workflow,omitempty" yaml:"workflow,omitempty"`
	// Commit push event criteria
	PushBranch string `json:"push_branch,omitempty" yaml:"push_branch,omitempty"`
	// Tag push event criteria
	Tag string `json:"tag,omitempty" yaml:"tag,omitempty"`
	// Pull Request event criteria
	PullRequestSourceBranch string `json:"pull_request_source_branch,omitempty" yaml:"pull_request_source_branch,omitempty"`
	PullRequestTargetBranch string `json:"pull_request_target_branch,omitempty" yaml:"pull_request_target_branch,omitempty"`
	DraftPullRequestEnabled *bool  `json:"draft_pull_request_enabled,omitempty" yaml:"draft_pull_request_enabled,omitempty"`

	// Deprecated
	Pattern              string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	IsPullRequestAllowed bool   `json:"is_pull_request_allowed,omitempty" yaml:"is_pull_request_allowed,omitempty"`
}

func (triggerItem TriggerMapItemModel) Validate(workflows, pipelines []string) ([]string, error) {
	var warnings []string

	// Validate target
	if triggerItem.PipelineID != "" && triggerItem.WorkflowID != "" {
		return warnings, fmt.Errorf("both pipeline and workflow are defined as trigger target: %s", triggerItem.String(false))
	}
	if triggerItem.PipelineID == "" && triggerItem.WorkflowID == "" {
		return warnings, fmt.Errorf("no pipeline nor workflow is defined as a trigger target: %s", triggerItem.String(false))
	}

	if strings.HasPrefix(triggerItem.WorkflowID, "_") {
		warnings = append(warnings, fmt.Sprintf("workflow (%s) defined in trigger item (%s), but utility workflows can't be triggered directly", triggerItem.WorkflowID, triggerItem.String(true)))
	}

	found := false
	if triggerItem.PipelineID != "" {
		for _, pipelineID := range pipelines {
			if pipelineID == triggerItem.PipelineID {
				found = true
				break
			}
		}

		if !found {
			return warnings, fmt.Errorf("pipeline (%s) defined in trigger item (%s), but does not exist", triggerItem.PipelineID, triggerItem.String(true))
		}
	} else {
		for _, workflowID := range workflows {
			if workflowID == triggerItem.WorkflowID {
				found = true
				break
			}
		}

		if !found {
			return warnings, fmt.Errorf("workflow (%s) defined in trigger item (%s), but does not exist", triggerItem.WorkflowID, triggerItem.String(true))
		}
	}

	// Validate match criteria
	if triggerItem.Pattern == "" {
		_, err := triggerEventType(triggerItem.PushBranch, triggerItem.PullRequestSourceBranch, triggerItem.PullRequestTargetBranch, triggerItem.Tag)
		if err != nil {
			return warnings, fmt.Errorf("trigger map item (%s) validate failed, error: %s", triggerItem.String(true), err)
		}
	} else if triggerItem.PushBranch != "" ||
		triggerItem.PullRequestSourceBranch != "" || triggerItem.PullRequestTargetBranch != "" || triggerItem.Tag != "" {
		return warnings, fmt.Errorf("deprecated trigger item (pattern defined), mixed with trigger params (push_branch: %s, pull_request_source_branch: %s, pull_request_target_branch: %s, tag: %s)", triggerItem.PushBranch, triggerItem.PullRequestSourceBranch, triggerItem.PullRequestTargetBranch, triggerItem.Tag)
	}

	return warnings, nil
}

func (triggerItem TriggerMapItemModel) MatchWithParams(pushBranch, prSourceBranch, prTargetBranch string, tag string) (bool, error) {
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

			return sourceMatch && targetMatch, nil
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

func (triggerItem TriggerMapItemModel) String(printTarget bool) string {
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

		str += fmt.Sprintf(" && draft_pull_request_enabled: %v", triggerItem.IsDraftPullRequestEnabled())
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

	if printTarget {
		if triggerItem.PipelineID != "" {
			str += fmt.Sprintf(" -> pipeline: %s", triggerItem.PipelineID)
		} else {
			str += fmt.Sprintf(" -> workflow: %s", triggerItem.WorkflowID)
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
