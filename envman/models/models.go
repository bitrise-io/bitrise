// Package models aliases the types of the published envman/v2 models.
//
// Stepman's StepModel uses envman/v2 models for step inputs and outputs, so bitrise and the in-repo envman
// must share the exact same types: EnvironmentItemModel stores an EnvironmentItemOptionsModel struct inside
// a map, and a struct of a different (even identical) type fails GetOptions at runtime.
// The methods of these types come from envman/v2 through the aliases.
// Once stepman is moved into this repository (STEP-2546), restore the type definitions here and the methods
// in models_methods.go (see commit 13e2aea9).
package models

import envmanModels "github.com/bitrise-io/envman/v2/models"

type (
	EnvironmentItemOptionsModel = envmanModels.EnvironmentItemOptionsModel
	EnvironmentItemModel        = envmanModels.EnvironmentItemModel
	EnvsSerializeModel          = envmanModels.EnvsSerializeModel
	EnvsJSONListModel           = envmanModels.EnvsJSONListModel
)
