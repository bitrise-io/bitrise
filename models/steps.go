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

func (stepListItem *StepListItemRaw) GetKeyAndType() (string, StepListItemType, error) {
	if stepListItem == nil {
		return "", StepListItemTypeUnknown, fmt.Errorf("step list item is nil")
	}

	if len(*stepListItem) == 0 {
		return "", StepListItemTypeUnknown, errors.New("empty step list item")
	}

	if len(*stepListItem) > 1 {
		return "", StepListItemTypeUnknown, fmt.Errorf("step list item has more than 1 key: %#v", stepListItem)
	}

	var itemID string
	for key := range *stepListItem {
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

func (stepListItem *StepListItemRaw) GetItem(target interface{}) (string, error) {
	key, t, err := stepListItem.GetKeyAndType()
	if err != nil {
		return "", err
	}

	var value any
	for _, v := range *stepListItem {
		value = v
		break
	}

	switch ptr := target.(type) {
	case *stepmanModels.StepModel:
		if t != StepListItemTypeStep {
			return "", fmt.Errorf("step list item (%s) is not a step", key)
		}

		step, ok := value.(stepmanModels.StepModel)
		if !ok {
			// TODO: why is this needed?
			stepPtr, ok := value.(*stepmanModels.StepModel)
			if !ok {
				return "", fmt.Errorf("step list item value is not a step")
			}
			step = *stepPtr
		}
		*ptr = step
	case *StepBundleListItemModel:
		if t != StepListItemTypeBundle {
			return "", fmt.Errorf("step list item (%s) is not a step bundle", key)
		}

		bundle, ok := value.(StepBundleListItemModel)
		if !ok {
			bundlePtr, ok := value.(*StepBundleListItemModel)
			if !ok {
				return "", fmt.Errorf("step list item value is not a Step Bundle")
			}
			bundle = *bundlePtr
		}
		*ptr = bundle
	case *WithModel:
		if t != StepListItemTypeWith {
			return "", fmt.Errorf("step list item (%s) is not a With group", key)
		}

		with, ok := value.(WithModel)
		if !ok {
			withPtr, ok := value.(*WithModel)
			if !ok {
				return "", fmt.Errorf("step list item value is not a With group")
			}
			with = *withPtr
		}
		*ptr = with
	default:
		return "", fmt.Errorf("unsupported target type: %T", target)
	}

	return key, nil
}

func (stepListItem *StepListItemRaw) UnmarshalJSON(b []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	var key string
	for k := range raw {
		key = k
		break
	}

	*stepListItem = map[string]any{}
	if key == StepListItemWithKey {
		var withItem map[string]WithModel
		if err := json.Unmarshal(b, &withItem); err != nil {
			return err
		}

		for k, v := range withItem {
			(*stepListItem)[k] = v
		}
	} else if strings.HasPrefix(key, StepListItemStepBundleKeyPrefix) {
		var stepBundleItem map[string]StepBundleListItemModel
		if err := json.Unmarshal(b, &stepBundleItem); err != nil {
			return err
		}

		for k, v := range stepBundleItem {
			(*stepListItem)[k] = v
		}
	} else {
		var stepItem map[string]stepmanModels.StepModel
		if err := json.Unmarshal(b, &stepItem); err != nil {
			return err
		}

		for k, v := range stepItem {
			(*stepListItem)[k] = v
		}
	}

	return nil
}

func (stepListItem *StepListItemRaw) UnmarshalYAML(unmarshal func(any) error) error {
	var raw map[string]any
	if err := unmarshal(&raw); err != nil {
		return err
	}

	var key string
	for k := range raw {
		key = k
		break
	}

	*stepListItem = map[string]any{}
	if key == StepListItemWithKey {
		var withItem map[string]WithModel
		if err := unmarshal(&withItem); err != nil {
			return err
		}

		for k, v := range withItem {
			(*stepListItem)[k] = v
		}
	} else if strings.HasPrefix(key, StepListItemStepBundleKeyPrefix) {
		var stepBundleItem map[string]StepBundleListItemModel
		if err := unmarshal(&stepBundleItem); err != nil {
			return err
		}

		for k, v := range stepBundleItem {
			(*stepListItem)[k] = v
		}
	} else {
		var stepItem map[string]stepmanModels.StepModel
		if err := unmarshal(&stepItem); err != nil {
			return err
		}

		for k, v := range stepItem {
			(*stepListItem)[k] = v
		}
	}

	return nil
}

// StepListItemModel represents a step list items for a Workflow (can be a step, step bundle and with group)
type StepListItemModel StepListItemRaw

func (stepListItem *StepListItemModel) GetKeyAndType() (string, StepListItemType, error) {
	if stepListItem == nil {
		return "", StepListItemTypeUnknown, fmt.Errorf("step list item is nil")
	}

	raw := StepListItemRaw(*stepListItem)
	return raw.GetKeyAndType()
}

func (stepListItem *StepListItemModel) GetStep() (string, *stepmanModels.StepModel, error) {
	if stepListItem == nil {
		return "", nil, fmt.Errorf("step list item is nil")
	}

	var step stepmanModels.StepModel
	raw := StepListItemRaw(*stepListItem)
	stepID, err := raw.GetItem(&step)
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
	raw := StepListItemRaw(*stepListItem)
	_, err := raw.GetItem(&stepBundle)
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
	raw := StepListItemRaw(*stepListItem)
	_, err := raw.GetItem(&withGroup)
	if err != nil {
		return nil, err
	}

	return &withGroup, nil
}

func (stepListItem *StepListItemModel) UnmarshalJSON(b []byte) error {
	var raw StepListItemRaw
	if err := raw.UnmarshalJSON(b); err != nil {
		return err
	}

	*stepListItem = StepListItemModel(raw)
	return nil
}

func (stepListItem *StepListItemModel) UnmarshalYAML(unmarshal func(any) error) error {
	var raw StepListItemRaw
	if err := raw.UnmarshalYAML(unmarshal); err != nil {
		return err
	}

	*stepListItem = StepListItemModel(raw)
	return nil
}

// StepListItemStepOrBundleModel represents a step list items for a Step Bundle (can be a step or step bundle)
type StepListItemStepOrBundleModel StepListItemRaw

func (stepListItem *StepListItemStepOrBundleModel) GetKeyAndType() (string, StepListItemType, error) {
	if stepListItem == nil {
		return "", StepListItemTypeUnknown, fmt.Errorf("step list item is nil")
	}

	raw := StepListItemRaw(*stepListItem)
	key, t, err := raw.GetKeyAndType()
	if err != nil {
		return "", StepListItemTypeUnknown, err
	}
	if t == StepListItemTypeWith {
		return "", StepListItemTypeUnknown, fmt.Errorf("step list item of step bundle cannot be a with group")
	}

	return key, t, nil
}

func (stepListItem *StepListItemStepOrBundleModel) GetStep() (string, *stepmanModels.StepModel, error) {
	if stepListItem == nil {
		return "", nil, fmt.Errorf("step list item is nil")
	}

	var step stepmanModels.StepModel
	raw := StepListItemRaw(*stepListItem)
	stepID, err := raw.GetItem(&step)
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
	raw := StepListItemRaw(*stepListItem)
	_, err := raw.GetItem(&stepBundle)
	if err != nil {
		return nil, err
	}

	return &stepBundle, nil
}

func (stepListItem *StepListItemStepOrBundleModel) GetWith() (*WithModel, error) {
	return nil, fmt.Errorf("step list item of step bundle cannot be a with group")
}

func (stepListItem *StepListItemStepOrBundleModel) UnmarshalJSON(b []byte) error {
	var raw StepListItemRaw
	if err := raw.UnmarshalJSON(b); err != nil {
		return err
	}

	*stepListItem = StepListItemStepOrBundleModel(raw)
	return nil
}

func (stepListItem *StepListItemStepOrBundleModel) UnmarshalYAML(unmarshal func(any) error) error {
	var raw StepListItemRaw
	if err := raw.UnmarshalYAML(unmarshal); err != nil {
		return err
	}

	*stepListItem = StepListItemStepOrBundleModel(raw)
	return nil
}

// StepListStepItemModel represents a step list items for a With group (can be a step)
type StepListStepItemModel StepListItemRaw

func (stepListItem *StepListStepItemModel) GetKeyAndType() (string, StepListItemType, error) {
	if stepListItem == nil {
		return "", StepListItemTypeUnknown, fmt.Errorf("step list item is nil")
	}

	raw := StepListItemRaw(*stepListItem)
	key, t, err := raw.GetKeyAndType()
	if err != nil {
		return "", StepListItemTypeUnknown, err
	}
	if t == StepListItemTypeWith {
		return "", StepListItemTypeUnknown, fmt.Errorf("step list item of a with group cannot be a with group")
	} else if t == StepListItemTypeBundle {
		return "", StepListItemTypeUnknown, fmt.Errorf("step list item of a with group cannot be a step bundle")
	}

	return key, t, nil
}

func (stepListItem *StepListStepItemModel) GetStep() (string, *stepmanModels.StepModel, error) {
	if stepListItem == nil {
		return "", nil, fmt.Errorf("step list item is nil")
	}

	var step stepmanModels.StepModel
	raw := StepListItemRaw(*stepListItem)
	stepID, err := raw.GetItem(&step)
	if err != nil {
		return "", nil, err
	}

	return stepID, &step, nil
}

func (stepListItem *StepListStepItemModel) GetBundle() (*StepBundleListItemModel, error) {
	return nil, fmt.Errorf("step list item of a with group cannot be a step bundle")
}

func (stepListItem *StepListStepItemModel) GetWith() (*WithModel, error) {
	return nil, fmt.Errorf("step list item of a with group cannot be a with group")
}

func (stepListItem *StepListStepItemModel) UnmarshalJSON(b []byte) error {
	var raw StepListItemRaw
	if err := raw.UnmarshalJSON(b); err != nil {
		return err
	}

	*stepListItem = StepListStepItemModel(raw)
	return nil
}

func (stepListItem *StepListStepItemModel) UnmarshalYAML(unmarshal func(any) error) error {
	var raw StepListItemRaw
	if err := raw.UnmarshalYAML(unmarshal); err != nil {
		return err
	}

	*stepListItem = StepListStepItemModel(raw)
	return nil
}

type Containerisable interface {
	GetExecutionContainerConfig() (*ContainerConfig, error)
	GetServiceContainerConfigs() ([]ContainerConfig, error)
}

type containerisableStep struct {
	Step stepmanModels.StepModel
}

func newContainerisableStep(step stepmanModels.StepModel) Containerisable {
	return containerisableStep{
		Step: step,
	}
}

func (step containerisableStep) GetExecutionContainerConfig() (*ContainerConfig, error) {
	if step.Step.ExecutionContainer == nil {
		return nil, nil
	}

	return getContainerConfig(step.Step.ExecutionContainer)
}

func (step containerisableStep) GetServiceContainerConfigs() ([]ContainerConfig, error) {
	if step.Step.ServiceContainers == nil {
		return nil, nil
	}

	var containerConfigs []ContainerConfig
	for _, containerDef := range step.Step.ServiceContainers {
		ctrConfig, err := getContainerConfig(containerDef)
		if err != nil {
			return nil, err
		}
		if ctrConfig != nil {
			containerConfigs = append(containerConfigs, *ctrConfig)
		}
	}
	return containerConfigs, nil
}

type containerisableStepBundle struct {
	StepBundle StepBundleListItemModel
}

func newContainerisableStepBundle(stepBundle StepBundleListItemModel) Containerisable {
	return containerisableStepBundle{
		StepBundle: stepBundle,
	}
}

func (stepBundle containerisableStepBundle) GetExecutionContainerConfig() (*ContainerConfig, error) {
	if stepBundle.StepBundle.ExecutionContainer == nil {
		return nil, nil
	}

	return getContainerConfig(stepBundle.StepBundle.ExecutionContainer)
}

func (stepBundle containerisableStepBundle) GetServiceContainerConfigs() ([]ContainerConfig, error) {
	if stepBundle.StepBundle.ServiceContainers == nil {
		return nil, nil
	}

	var containerConfigs []ContainerConfig
	for _, containerDef := range stepBundle.StepBundle.ServiceContainers {
		ctrConfig, err := getContainerConfig(containerDef)
		if err != nil {
			return nil, err
		}
		if ctrConfig != nil {
			containerConfigs = append(containerConfigs, *ctrConfig)
		}
	}
	return containerConfigs, nil
}

/*
Get ContainerConfig from container definition which can be either a string or a map.

Examples:
  - redis
  - postgres: {recreate: true}
*/
func getContainerConfig(container any) (*ContainerConfig, error) {
	if container == nil {
		return nil, nil
	}

	if ctrStr, ok := container.(string); ok {
		return &ContainerConfig{
			ContainerID: ctrStr,
			Recreate:    false,
		}, nil
	}

	var id string
	var recreate bool
	if ctrMap, ok := container.(map[any]any); ok {
		for k, v := range ctrMap {
			id, ok = k.(string)
			if !ok {
				return nil, fmt.Errorf("invalid container config ID type: %T", k)
			}

			if ctrCfg, ok := v.(map[any]any); ok {
				recreateVal, ok := ctrCfg["recreate"]
				if ok {
					recreate, ok = recreateVal.(bool)
					if !ok {
						return nil, fmt.Errorf("invalid recreate value type: %T", recreateVal)
					}
				}
			}

			break
		}

		return &ContainerConfig{
			ContainerID: id,
			Recreate:    recreate,
		}, nil
	}

	return nil, nil
}
