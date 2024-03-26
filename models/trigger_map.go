package models

import (
	"fmt"
)

type TriggerMapModel []TriggerMapItemModel

func (triggerMap TriggerMapModel) Normalised() ([]TriggerMapItemModel, error) {
	var items []TriggerMapItemModel
	for _, item := range triggerMap {
		normalizedItem, err := item.Normalized()
		if err != nil {
			return nil, err
		}
		items = append(items, normalizedItem)
	}
	return items, nil
}

func (triggerMap TriggerMapModel) Validate(workflows, pipelines []string) ([]string, error) {
	var warnings []string

	if err := triggerMap.checkDuplicatedTriggerMapItems(); err != nil {
		return warnings, err
	}

	for idx, item := range triggerMap {
		warns, err := item.Validate(idx, workflows, pipelines)
		warnings = append(warnings, warns...)
		if err != nil {
			return warnings, err
		}
	}

	return warnings, nil
}

func (triggerMap TriggerMapModel) FirstMatchingTarget(pushBranch, prSourceBranch, prTargetBranch string, prReadyState PullRequestReadyState, tag string) (string, string, error) {
	for _, item := range triggerMap {
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, prReadyState, tag)
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
	items := make(map[string]int)

	for idx, triggerItem := range triggerMap {
		itemStr := triggerItem.conditionsString()

		storedItemIdx, ok := items[itemStr]
		if ok {
			return fmt.Errorf("the %d. trigger item duplicates the %d. trigger item", idx+1, storedItemIdx+1)
		}

		items[itemStr] = idx
	}

	return nil
}
