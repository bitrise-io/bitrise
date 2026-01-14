package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	stepmanModels "github.com/bitrise-io/stepman/models"
)

type StepListItemType int

const (
	StepListItemTypeUnknown StepListItemType = iota
	StepListItemTypeStep
	StepListItemTypeWith
	StepListItemTypeBundle
)

const (
	StepListItemWithKey             = "with"
	StepListItemStepBundleKeyPrefix = "bundle::"
)

type StepListItem interface {
	GetKeyAndType() (string, StepListItemType, error)
	GetStep() (string, *stepmanModels.StepModel, error)
	GetBundle() (*StepBundleListItemModel, error)
	GetWith() (*WithModel, error)
}

type StepListItemRaw map[string]any

func (raw StepListItemRaw) GetKeyAndType() (string, StepListItemType, error) {
	if raw == nil {
		return "", StepListItemTypeUnknown, fmt.Errorf("step list item is nil")
	}

	if len(raw) == 0 {
		return "", StepListItemTypeUnknown, errors.New("empty step list item")
	}

	if len(raw) > 1 {
		return "", StepListItemTypeUnknown, fmt.Errorf("step list item has more than 1 key: %#v", raw)
	}

	var itemID string
	for key := range raw {
		itemID = key
		break
	}

	switch {
	case strings.HasPrefix(itemID, StepListItemStepBundleKeyPrefix):
		return strings.TrimPrefix(itemID, StepListItemStepBundleKeyPrefix), StepListItemTypeBundle, nil
	case itemID == StepListItemWithKey:
		return itemID, StepListItemTypeWith, nil
	default:
		return itemID, StepListItemTypeStep, nil
	}
}

func (raw StepListItemRaw) GetItem(target interface{}) (string, error) {
	if raw == nil {
		return "", fmt.Errorf("step list item is nil")
	}

	if len(raw) == 0 {
		return "", fmt.Errorf("step list item is empty")
	}

	if len(raw) > 1 {
		return "", fmt.Errorf("step list item has more than 1 key: %#v", raw)
	}

	var itemID string
	var value any
	for key, val := range raw {
		itemID = key
		value = val
		break
	}

	switch ptr := target.(type) {
	case *stepmanModels.StepModel:
		step, ok := value.(stepmanModels.StepModel)
		if !ok {
			return "", fmt.Errorf("step list item value is not a Step (got %T)", value)
		}
		*ptr = step
	case *StepBundleListItemModel:
		bundle, ok := value.(StepBundleListItemModel)
		if !ok {
			return "", fmt.Errorf("step list item value is not a Step Bundle")
		}
		*ptr = bundle
	case *WithModel:
		with, ok := value.(WithModel)
		if !ok {
			return "", fmt.Errorf("step list item value is not a With group")
		}
		*ptr = with
	default:
		return "", fmt.Errorf("unsupported target type: %T", target)
	}

	return itemID, nil
}

// StepListItemModel represents a step list items for a Workflow (can be a step, step bundle and with group)
type StepListItemModel StepListItemRaw

func (stepListItem *StepListItemModel) GetKeyAndType() (string, StepListItemType, error) {
	if stepListItem == nil {
		return "", StepListItemTypeUnknown, fmt.Errorf("step list item is nil")
	}

	return StepListItemRaw(*stepListItem).GetKeyAndType()
}

func (stepListItem *StepListItemModel) GetStep() (string, *stepmanModels.StepModel, error) {
	if stepListItem == nil {
		return "", nil, fmt.Errorf("step list item is nil")
	}

	var step stepmanModels.StepModel
	stepID, err := StepListItemRaw(*stepListItem).GetItem(&step)
	if err != nil {
		return "", nil, err
	}

	return stepID, &step, nil
}

func (stepListItem *StepListItemModel) GetBundle() (*StepBundleListItemModel, error) {
	if stepListItem == nil {
		return nil, fmt.Errorf("step list item is nil")
	}

	var stepBundle StepBundleListItemModel
	_, err := StepListItemRaw(*stepListItem).GetItem(&stepBundle)
	if err != nil {
		return nil, err
	}

	return &stepBundle, nil
}

func (stepListItem *StepListItemModel) GetWith() (*WithModel, error) {
	if stepListItem == nil {
		return nil, fmt.Errorf("step list item is nil")
	}

	var withGroup WithModel
	_, err := StepListItemRaw(*stepListItem).GetItem(&withGroup)
	if err != nil {
		return nil, err
	}

	return &withGroup, nil
}

func (stepListItem *StepListItemModel) UnmarshalJSON(b []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	var key string
	for k := range raw {
		key = k
		break
	}

	if key == StepListItemWithKey {
		var withItem StepListWithItemModel
		if err := json.Unmarshal(b, &withItem); err != nil {
			return err
		}

		*stepListItem = map[string]any{}
		for k, v := range withItem {
			(*stepListItem)[k] = v
		}
	} else if strings.HasPrefix(key, StepListItemStepBundleKeyPrefix) {
		var stepBundleItem StepListStepBundleItemModel
		if err := json.Unmarshal(b, &stepBundleItem); err != nil {
			return err
		}

		*stepListItem = map[string]any{}
		for k, v := range stepBundleItem {
			(*stepListItem)[k] = v
		}
	} else {
		var stepItem map[string]stepmanModels.StepModel
		if err := json.Unmarshal(b, &stepItem); err != nil {
			return err
		}

		*stepListItem = map[string]any{}
		for k, v := range stepItem {
			(*stepListItem)[k] = v
		}
	}

	return nil
}

func (stepListItem *StepListItemModel) UnmarshalYAML(unmarshal func(any) error) error {
	var raw map[string]any
	if err := unmarshal(&raw); err != nil {
		return err
	}

	var key string
	for k := range raw {
		key = k
		break
	}

	if key == StepListItemWithKey {
		var withItem StepListWithItemModel
		if err := unmarshal(&withItem); err != nil {
			return err
		}

		*stepListItem = map[string]any{}
		for k, v := range withItem {
			(*stepListItem)[k] = v
		}
	} else if strings.HasPrefix(key, StepListItemStepBundleKeyPrefix) {
		var stepBundleItem StepListStepBundleItemModel
		if err := unmarshal(&stepBundleItem); err != nil {
			return err
		}

		*stepListItem = map[string]any{}
		for k, v := range stepBundleItem {
			(*stepListItem)[k] = v
		}
	} else {
		var stepItem map[string]stepmanModels.StepModel
		if err := unmarshal(&stepItem); err != nil {
			return err
		}

		*stepListItem = map[string]any{}
		for k, v := range stepItem {
			(*stepListItem)[k] = v
		}
	}

	return nil
}

// StepListItemStepOrBundleModel represents a step list items for a Step Bundle (can be a step or step bundle)
type StepListItemStepOrBundleModel StepListItemRaw

func (stepListItem *StepListItemStepOrBundleModel) GetKeyAndType() (string, StepListItemType, error) {
	if stepListItem == nil {
		return "", StepListItemTypeUnknown, fmt.Errorf("step list item is nil")
	}

	return StepListItemRaw(*stepListItem).GetKeyAndType()
}

func (stepListItem *StepListItemStepOrBundleModel) GetStep() (string, *stepmanModels.StepModel, error) {
	if stepListItem == nil {
		return "", nil, fmt.Errorf("step list item is nil")
	}

	var step stepmanModels.StepModel
	stepID, err := StepListItemRaw(*stepListItem).GetItem(&step)
	if err != nil {
		return "", nil, err
	}

	return stepID, &step, nil
}

func (stepListItem *StepListItemStepOrBundleModel) GetBundle() (*StepBundleListItemModel, error) {
	if stepListItem == nil {
		return nil, fmt.Errorf("step list item is nil")
	}

	var stepBundle StepBundleListItemModel
	_, err := StepListItemRaw(*stepListItem).GetItem(&stepBundle)
	if err != nil {
		return nil, err
	}

	return &stepBundle, nil
}

func (stepListItem *StepListItemStepOrBundleModel) UnmarshalJSON(b []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	var key string
	for k := range raw {
		key = k
		break
	}

	if strings.HasPrefix(key, StepListItemStepBundleKeyPrefix) {
		var stepBundleItem StepListStepBundleItemModel
		if err := json.Unmarshal(b, &stepBundleItem); err != nil {
			return err
		}

		*stepListItem = map[string]any{}
		for k, v := range stepBundleItem {
			(*stepListItem)[k] = v
		}
	} else {
		var stepItem map[string]stepmanModels.StepModel
		if err := json.Unmarshal(b, &stepItem); err != nil {
			return err
		}

		*stepListItem = map[string]any{}
		for k, v := range stepItem {
			(*stepListItem)[k] = v
		}
	}

	return nil
}

func (stepListItem *StepListItemStepOrBundleModel) UnmarshalYAML(unmarshal func(any) error) error {
	var raw map[string]any
	if err := unmarshal(&raw); err != nil {
		return err
	}

	var key string
	for k := range raw {
		key = k
		break
	}

	if strings.HasPrefix(key, StepListItemStepBundleKeyPrefix) {
		var stepBundleItem StepListStepBundleItemModel
		if err := unmarshal(&stepBundleItem); err != nil {
			return err
		}

		*stepListItem = map[string]any{}
		for k, v := range stepBundleItem {
			(*stepListItem)[k] = v
		}
	} else {
		var stepItem StepListStepItemModel
		if err := unmarshal(&stepItem); err != nil {
			return err
		}

		*stepListItem = map[string]any{}
		for k, v := range stepItem {
			(*stepListItem)[k] = v
		}
	}

	return nil
}

// StepListStepItemModel represents a step list items for a With group (can be a step)
type StepListStepItemModel StepListItemRaw

func (stepListItem *StepListStepItemModel) GetStep() (string, *stepmanModels.StepModel, error) {
	if stepListItem == nil {
		return "", nil, fmt.Errorf("step list item is nil")
	}

	var step stepmanModels.StepModel
	stepID, err := StepListItemRaw(*stepListItem).GetItem(&step)
	if err != nil {
		return "", nil, err
	}

	return stepID, &step, nil
}

func (stepListItem *StepListStepItemModel) UnmarshalJSON(b []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	var stepItem map[string]stepmanModels.StepModel
	if err := json.Unmarshal(b, &stepItem); err != nil {
		return err
	}

	*stepListItem = map[string]any{}
	for k, v := range stepItem {
		(*stepListItem)[k] = v
	}

	return nil
}

func (stepListItem *StepListStepItemModel) UnmarshalYAML(unmarshal func(any) error) error {
	var raw map[string]any
	if err := unmarshal(&raw); err != nil {
		return err
	}

	var stepItem map[string]stepmanModels.StepModel
	if err := unmarshal(&stepItem); err != nil {
		return err
	}

	*stepListItem = map[string]any{}
	for k, v := range stepItem {
		(*stepListItem)[k] = v
	}

	return nil
}
