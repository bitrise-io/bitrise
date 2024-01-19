package models

import (
	"errors"
	"fmt"
)

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

func (triggerMap TriggerMapModel) FirstMatchingTarget(pushBranch, prSourceBranch, prTargetBranch string, prReadyState string, tag string) (string, string, error) {
	for _, item := range triggerMap {
		match, err := item.MatchWithParams(pushBranch, prSourceBranch, prTargetBranch, tag)
		if err != nil {
			return "", "", err
		}
		if match {
			if item.IsDraftPullRequestEnabled() == true && prReadyState == "converted_to_ready_for_review" {
				return "", "", errors.New("skipped")
			}
			if item.IsDraftPullRequestEnabled() == false && prReadyState == "draft" {
				return "", "", errors.New("skipped")
			}
		}

		if match {
			return item.PipelineID, item.WorkflowID, nil
		}
	}

	return "", "", fmt.Errorf("no matching pipeline & workflow found with trigger params: push-branch: %s, pr-source-branch: %s, pr-target-branch: %s, tag: %s", pushBranch, prSourceBranch, prTargetBranch, tag)
}

func (triggerMap TriggerMapModel) checkDuplicatedTriggerMapItems() error {
	items := make(map[string]struct{})

	for _, triggerItem := range triggerMap {
		content := triggerItem.String(false)

		_, ok := items[content]
		if ok {
			return fmt.Errorf("duplicated trigger item found (%s)", content)
		}

		items[content] = struct{}{}
	}

	return nil
}
