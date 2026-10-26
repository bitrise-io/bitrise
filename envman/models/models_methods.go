package models

import (
	"encoding/json"
)

const (
	// OptionsKey ...
	OptionsKey = "opts"
)

const (
	// DefaultIsExpand ...
	DefaultIsExpand = true
	// DefaultIsSensitive ...
	DefaultIsSensitive = false
	// DefaultSkipIfEmpty ...
	DefaultSkipIfEmpty = false

	// DefaultIsRequired ...
	DefaultIsRequired = false
	// DefaultIsDontChangeValue ...
	DefaultIsDontChangeValue = false
	// DefaultIsTemplate ...
	DefaultIsTemplate = false
	// DefaultUnset ...
	DefaultUnset = false
)

func NewEnvJSONList(jsonStr string) (EnvsJSONListModel, error) {
	list := EnvsJSONListModel{}
	if err := json.Unmarshal([]byte(jsonStr), &list); err != nil {
		return EnvsJSONListModel{}, err
	}
	return list, nil
}
