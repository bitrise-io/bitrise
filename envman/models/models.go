// Package models re-exports the published envman/v2 models.
//
// Stepman's StepModel uses envman/v2 models for step inputs and outputs, so bitrise and the in-repo envman
// must share the exact same types
//
// TODO: Once stepman is moved into this repository, as a part of STEP-2545, restore the full model code
// here (see commit 4f6d3680).
package models

import envmanModels "github.com/bitrise-io/envman/v2/models"

type (
	EnvironmentItemOptionsModel = envmanModels.EnvironmentItemOptionsModel
	EnvironmentItemModel        = envmanModels.EnvironmentItemModel
	EnvsSerializeModel          = envmanModels.EnvsSerializeModel
	EnvsJSONListModel           = envmanModels.EnvsJSONListModel
)

const (
	OptionsKey = envmanModels.OptionsKey

	DefaultIsExpand          = envmanModels.DefaultIsExpand
	DefaultIsSensitive       = envmanModels.DefaultIsSensitive
	DefaultSkipIfEmpty       = envmanModels.DefaultSkipIfEmpty
	DefaultIsRequired        = envmanModels.DefaultIsRequired
	DefaultIsDontChangeValue = envmanModels.DefaultIsDontChangeValue
	DefaultIsTemplate        = envmanModels.DefaultIsTemplate
	DefaultUnset             = envmanModels.DefaultUnset
)

// NewEnvJSONList ...
func NewEnvJSONList(jsonStr string) (EnvsJSONListModel, error) {
	return envmanModels.NewEnvJSONList(jsonStr)
}
