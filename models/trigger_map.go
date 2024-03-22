package models

import (
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
