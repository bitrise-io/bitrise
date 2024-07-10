package models

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type ConfigFileTreeModel struct {
	Path     string                `json:"path" yaml:"path"`
	Contents string                `json:"contents,omitempty" yaml:"contents,omitempty"`
	Includes []ConfigFileTreeModel `json:"includes,omitempty" yaml:"includes,omitempty"`

	Depth int `json:"-" yaml:"-"`
}

func (configTree *ConfigFileTreeModel) Merge() (string, error) {
	result, err := merge(configTree)
	if err != nil {
		return "", err
	}

	mergedYml, err := yaml.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to create YML result, error: %s", err)
	}

	return string(mergedYml), nil
}

type yamlMap = map[any]any

func merge(ymlTree *ConfigFileTreeModel) (yamlMap, error) {
	// Initial state is an empty map (YAML root)
	initial := make(yamlMap)

	result, err := mergeTree(initial, ymlTree)
	if err != nil {
		return nil, fmt.Errorf("failed to merge YML files, error: %s", err)
	}

	// Remove include list from result
	delete(result, "include")

	return result, nil
}

func mergeTree(existingValue yamlMap, treeToMerge *ConfigFileTreeModel) (yamlMap, error) {
	var err error

	// DFS: first the includes in the specified order, then the including file
	for _, includedTree := range treeToMerge.Includes {
		existingValue, err = mergeTree(existingValue, &includedTree)
		if err != nil {
			return nil, fmt.Errorf("failed to merge YML file %s, error: %s", includedTree.Path, err)
		}
	}

	// We assume that each YML file has a map at root, it's invalid otherwise
	var config yamlMap
	err = yaml.Unmarshal([]byte(treeToMerge.Contents), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YML file %s, error: %s", treeToMerge.Path, err)
	}

	if config == nil {
		// File is empty
		config = make(yamlMap)
	}

	return mergeMap(existingValue, config), nil
}

func mergeValue(existingValue any, valueToMerge any) any {

	switch valueToMerge.(type) {
	case yamlMap:
		existingMap, ok := existingValue.(yamlMap)
		if !ok {
			// Existing value is not a map, replace with new value
			existingMap = make(yamlMap)
		}
		return mergeMap(existingMap, valueToMerge.(yamlMap))
	case []any:
		existingSlice, ok := existingValue.([]any)
		if !ok {
			// Existing value is not a slice, replace with new value
			existingSlice = nil
		}
		return mergeSlice(existingSlice, valueToMerge.([]any))
	default:
		// Simple types
		return valueToMerge
	}
}

func mergeMap(existingMap yamlMap, mapToMerge yamlMap) yamlMap {
	for key, valueToMerge := range mapToMerge {
		existingValue, exists := existingMap[key]
		if exists {
			// Key exists in result and merged maps
			existingMap[key] = mergeValue(existingValue, valueToMerge)
		} else {
			// Key doesn't exist in result yet
			existingMap[key] = valueToMerge
		}
	}

	return existingMap
}

func mergeSlice(existingArray []any, arrayToAppend []any) []any {
	existingArray = append(existingArray, arrayToAppend...)
	return existingArray
}
