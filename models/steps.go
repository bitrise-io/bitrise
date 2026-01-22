package models

import (
	"fmt"

	stepmanModels "github.com/bitrise-io/stepman/models"
)

type StepListItem struct {
	key      string
	itemType StepListItemType
	step     *stepmanModels.StepModel
	with     *WithModel
	bundle   *StepBundleListItemModel
}

func NewStepListItemFromWorkflowStep(source StepListItemModel) (*StepListItem, error) {
	k, t, err := source.GetKeyAndType()
	if err != nil {
		return nil, err
	}

	item := &StepListItem{
		key:      k,
		itemType: t,
	}

	switch t {
	case StepListItemTypeStep:
		step, err := source.GetStep()
		if err != nil {
			return nil, err
		}
		item.step = step
	case StepListItemTypeWith:
		with, err := source.GetWith()
		if err != nil {
			return nil, err
		}
		item.with = with
	case StepListItemTypeBundle:
		bundle, err := source.GetBundle()
		if err != nil {
			return nil, err
		}
		item.bundle = bundle
	default:
		return nil, fmt.Errorf("unknown step list item type")
	}
	return item, nil
}

func NewStepListItemFromWithStep(source StepListStepItemModel) (*StepListItem, error) {
	stepID, step, err := source.GetStepIDAndStep()
	if err != nil {
		return nil, err
	}

	return &StepListItem{
		key:      stepID,
		itemType: StepListItemTypeStep,
		step:     &step,
	}, nil
}

func NewStepListItemFromBundleStep(source StepListItemStepOrBundleModel) (*StepListItem, error) {
	k, t, err := source.GetKeyAndType()
	if err != nil {
		return nil, err
	}

	item := &StepListItem{
		key:      k,
		itemType: t,
	}

	switch t {
	case StepListItemTypeStep:
		step, err := source.GetStep()
		if err != nil {
			return nil, err
		}
		item.step = step
	case StepListItemTypeBundle:
		bundle, err := source.GetBundle()
		if err != nil {
			return nil, err
		}
		item.bundle = bundle
	default:
		return nil, fmt.Errorf("unknown step list item type")
	}
	return item, nil
}

func (i *StepListItem) GetKeyAndType() (string, StepListItemType) {
	return i.key, i.itemType
}

func (i *StepListItem) GetStep() *stepmanModels.StepModel {
	return i.step
}

func (i *StepListItem) GetWithGroup() *WithModel {
	return i.with
}

func (i *StepListItem) GetBundle() *StepBundleListItemModel {
	return i.bundle
}

type Containerisable struct {
	Step       *stepmanModels.StepModel
	StepBundle *StepBundleListItemModel
}

func newContainerisableFromStep(step stepmanModels.StepModel) Containerisable {
	return Containerisable{
		Step: &step,
	}
}

func newContainerisableFromStepBundle(stepBundle StepBundleListItemModel) Containerisable {
	return Containerisable{
		StepBundle: &stepBundle,
	}
}

func (step Containerisable) GetExecutionContainerConfig() (*ContainerConfig, error) {
	if step.StepBundle != nil {
		return getContainerConfig(step.StepBundle.ExecutionContainer)
	} else if step.Step != nil {
		return getContainerConfig(step.Step.ExecutionContainer)
	}
	return nil, nil
}

func (step Containerisable) GetServiceContainerConfigs() ([]ContainerConfig, error) {
	var serviceContainers []any
	if step.StepBundle != nil {
		serviceContainers = step.StepBundle.ServiceContainers
	} else if step.Step != nil {
		serviceContainers = step.Step.ServiceContainers
	}
	if serviceContainers == nil {
		return nil, nil
	}

	var containerConfigs []ContainerConfig
	for _, containerDef := range serviceContainers {
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
