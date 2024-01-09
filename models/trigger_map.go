package models

import "fmt"

// TriggerMapModel ...
type TriggerMapModel []TriggerMapItemModel

// Validate ...
func (triggerMap TriggerMapModel) Validate(workflows, pipelines []string) error {
	for _, item := range triggerMap {
		if err := item.Validate(workflows, pipelines); err != nil {
			return err
		}
	}

	if err := triggerMap.checkDuplicatedTriggerMapItems(); err != nil {
		return err
	}

	return nil
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
