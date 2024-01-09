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
	for i := 0; i < len(triggerMap); i++ {
		triggerItemI := triggerMap[i]

		for j := i + 1; j < len(triggerMap); j++ {
			triggerItemJ := triggerMap[j]

			if triggerItemI.String(false) == triggerItemJ.String(false) {
				return fmt.Errorf("duplicated trigger item found (%s)", triggerItemI.String(false))
			}

		}
	}

	return nil
}
