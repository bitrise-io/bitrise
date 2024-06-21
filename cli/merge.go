package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"os"
	"reflect"
)

type YmlTreeModel struct {
	FileName string         `json:"fileName" yaml:"fileName"`
	Config   string         `json:"contents,omitempty" yaml:"contents,omitempty"`
	Includes []YmlTreeModel `json:"includes,omitempty" yaml:"includes,omitempty"`
}

type MergeResponseModel struct {
	Config *string `json:"config,omitempty" yaml:"config,omitempty"`
	Error  *string `json:"error,omitempty" yaml:"error,omitempty"`
}

type YamlMap = map[any]any

// JSON ...
func (m MergeResponseModel) JSON() string {
	bytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf(`"Failed to marshal merge result (%#v), err: %s"`, m, err)
	}
	return string(bytes)
}

func (m MergeResponseModel) String() string {
	if m.Error != nil {
		msg := fmt.Sprintf("%s: %s", colorstring.Red("Error"), *m.Error)
		return msg
	}

	if m.Config != nil {
		msg := *m.Config
		return msg
	}

	return ""
}

// NewMergeResponse ...
func NewMergeResponse(config string) MergeResponseModel {
	return MergeResponseModel{
		Config: &config,
	}
}

// NewMergeError ...
func NewMergeError(err string) MergeResponseModel {
	return MergeResponseModel{
		Error: &err,
	}
}

func merge(c *cli.Context) error {
	// Expand cli.Context
	ymlTreeBase64Data := c.String(ConfigBase64Key)
	// TODO also read from path? temp file needed?

	format := c.String(OuputFormatKey)
	if format == "" {
		format = output.FormatRaw
	}

	var log Logger
	log = NewDefaultRawLogger()
	if format == output.FormatRaw {
		log = NewDefaultRawLogger()
	} else if format == output.FormatJSON {
		log = NewDefaultJSONLogger()
	} else {
		log.Print(NewMergeError(fmt.Sprintf("Invalid format: %s", format)))
		os.Exit(1)
	}

	ymlTree, err := decodeYmlTree(ymlTreeBase64Data)
	if err != nil {
		log.Print(NewMergeError(err.Error()))
		os.Exit(1)
	}

	mergeResult, err := runMerge(ymlTree)
	if err != nil {
		log.Print(NewMergeError(err.Error()))
		os.Exit(1)
	}

	mergedYml, err := yaml.Marshal(mergeResult)
	if err != nil {
		log.Print(NewMergeError(err.Error()))
		os.Exit(1)
	}

	log.Print(NewMergeResponse(string(mergedYml)))

	return nil
}

func decodeYmlTree(ymlTreeBase64Data string) (*YmlTreeModel, error) {
	configBase64Bytes, err := base64.StdEncoding.DecodeString(ymlTreeBase64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 input, error: %s", err)
	}

	var ymlTree YmlTreeModel
	if err = json.Unmarshal(configBase64Bytes, &ymlTree); err != nil {
		return nil, fmt.Errorf("failed to parse YML files from JSON, error: %s", err)
	}
	return &ymlTree, nil
}

func runMerge(ymlTree *YmlTreeModel) (YamlMap, error) {
	// Initial state is an empty map (YAML root)
	initial := make(YamlMap)

	result, err := mergeTree(initial, ymlTree)
	if err != nil {
		return nil, fmt.Errorf("failed to merge YML files, error: %s", err)
	}

	// Remove include list from result
	delete(result, "include")

	return result, nil
}

func mergeTree(existingValue YamlMap, treeToMerge *YmlTreeModel) (YamlMap, error) {
	var err error

	// DFS: first the includes in the specified order, then the including file
	for _, includedTree := range treeToMerge.Includes {
		existingValue, err = mergeTree(existingValue, &includedTree)
		if err != nil {
			return nil, fmt.Errorf("failed to merge YML file %s, error: %s", includedTree.FileName, err)
		}
	}

	// We assume that each YML file has a map at root, it's invalid otherwise
	var config YamlMap
	err = yaml.Unmarshal([]byte(treeToMerge.Config), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YML file %s, error: %s", treeToMerge.FileName, err)
	}

	if config == nil {
		// File is empty
		config = make(YamlMap)
	}

	return mergeMap(existingValue, reflect.ValueOf(config)), nil
}

func mergeValue(existingValue any, valueToMerge reflect.Value) any {
	valueKind := valueToMerge.Kind()

	if valueKind == reflect.Interface {
		// Skip one level of recursion by unwrapping interfaces
		valueToMerge = valueToMerge.Elem()
		valueKind = valueToMerge.Kind()
	}

	switch valueKind {
	case reflect.Map:
		existingMap, ok := existingValue.(YamlMap)
		if !ok {
			// Existing value is not a map, replace with new value
			existingMap = make(YamlMap)
		}
		return mergeMap(existingMap, valueToMerge)
	case reflect.Slice:
		existingSlice, ok := existingValue.([]any)
		if !ok {
			// Existing value is not a slice, replace with new value
			existingSlice = nil
		}
		return mergeSlice(existingSlice, valueToMerge)
	default:
		// Simple types
		return valueToMerge.Interface()
	}
}

func mergeMap(existingMap YamlMap, mapToMerge reflect.Value) YamlMap {
	iterator := mapToMerge.MapRange()
	for iterator.Next() {
		key := iterator.Key().Interface()
		valueToMerge := iterator.Value()

		existingValue, exists := existingMap[key]
		if exists {
			// Key exists in result and merged maps
			existingMap[key] = mergeValue(existingValue, valueToMerge)
		} else {
			// Key doesn't exist in result yet
			existingMap[key] = valueToMerge.Interface()
		}
	}

	return existingMap
}

func mergeSlice(existingArray []any, arrayToAppend reflect.Value) []any {
	for i := 0; i < arrayToAppend.Len(); i++ {
		valueToAdd := arrayToAppend.Index(i)

		existingArray = append(existingArray, valueToAdd.Interface())
	}

	return existingArray
}
