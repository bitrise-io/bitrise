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

// StepListItemModel is a map representing a step list item of a workflow, the value is either a Step, a With Group or Step Bundle.
type StepListItemModel map[string]interface{}

// StepListStepItemModel is a map representing a step list item of a With group, the value is a Step.
type StepListStepItemModel map[string]stepmanModels.StepModel

// StepListItemStepOrBundleModel is a map representing a step list item of a Step Bundle, the value is either a Step or a Step Bundle.
type StepListItemStepOrBundleModel map[string]any
