package models

import "fmt"

type TriggerMapModel []TriggerMapItemModel

func (triggerMap TriggerMapModel) Validate(workflows, pipelines []string) ([]string, error) {
	var warnings []string
	for _, item := range triggerMap {
		warns, err := item.Validate(workflows, pipelines)
		warnings = append(warnings, warns...)
		if err != nil {
			return warnings, err
		}
	}

	if err := triggerMap.checkDuplicatedTriggerMapItems(); err != nil {
		return warnings, err
	}

	return warnings, nil
}

func (triggerMap TriggerMapModel) FirstMatchingTarget(pushBranch, prSourceBranch, prTargetBranch, tag string) (string, string, error) {
	for _, item := range triggerMap {
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		if err != nil {
			return "", "", err
		}
		if match {
			return item.PipelineID, item.WorkflowID, nil
		}
	}

	return "", "", fmt.Errorf("no matching pipeline & workflow found with trigger params: push-branch: %s, pr-source-branch: %s, pr-target-branch: %s, tag: %s", pushBranch, prSourceBranch, prTargetBranch, tag)
}

func (triggerMap TriggerMapModel) checkDuplicatedTriggerMapItems() error {
	triggerTypeItemMap := map[string][]TriggerMapItemModel{}

	for _, triggerItem := range triggerMap {
		if triggerItem.Pattern == "" {
			triggerType, err := triggerEventType(triggerItem.PushBranch, triggerItem.PullRequestSourceBranch, triggerItem.PullRequestTargetBranch, triggerItem.Tag)
			if err != nil {
				return fmt.Errorf("trigger map item (%v) validate failed, error: %s", triggerItem, err)
			}

			triggerItems := triggerTypeItemMap[string(triggerType)]

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
			triggerTypeItemMap[string(triggerType)] = triggerItems
		} else if triggerItem.Pattern != "" {
			triggerItems := triggerTypeItemMap["deprecated"]

			for _, item := range triggerItems {
				if triggerItem.Pattern == item.Pattern &&
					triggerItem.IsPullRequestAllowed == item.IsPullRequestAllowed {
					return fmt.Errorf("duplicated trigger item found (%s)", triggerItem.String(false))
				}
			}

			triggerItems = append(triggerItems, triggerItem)
			triggerTypeItemMap["deprecated"] = triggerItems
		}
	}

	return nil
}
