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
	Config   string         `json:"contents,omitempty" yaml:"contents,omitempty"`
	Includes []YmlTreeModel `json:"includes,omitempty" yaml:"includes,omitempty"`
}

type MergeResponseModel struct {
	Config *string `json:"config,omitempty" yaml:"config,omitempty"`
	Error  *string `json:"error,omitempty" yaml:"error,omitempty"`
}

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

	config, err := runMerge(ymlTree)
	if err != nil {
		log.Print(NewMergeError(err.Error()))
		os.Exit(1)
	}

	configString, err := yaml.Marshal(config)
	if err != nil {
		log.Print(NewMergeError(err.Error()))
		os.Exit(1)
	}

	log.Print(NewMergeResponse(string(configString)))

	return nil
}

func decodeYmlTree(ymlTreeBase64Data string) (*YmlTreeModel, error) {
	configBase64Bytes, err := base64.StdEncoding.DecodeString(ymlTreeBase64Data)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode base 64 string, error: %s", err)
	}

	var ymlTree YmlTreeModel
	if err = yaml.Unmarshal(configBase64Bytes, &ymlTree); err != nil {
		return nil, fmt.Errorf("Failed to parse yml tree, error: %s", err)
	}
	return &ymlTree, nil
}

func runMerge(ymlTree *YmlTreeModel) (map[interface{}]interface{}, error) {
	// In general, YAML could have a simple value or a list as root element,
	// but we assume it's an object, otherwise it's not a valid bitrise.yml anyway.
	initial := make(map[interface{}]interface{})

	result, err := mergeTree(initial, ymlTree)
	if err != nil {
		return nil, err // TODO wrap? log?
	}

	resultMap := result.(map[interface{}]interface{})
	return resultMap, nil
}

func mergeTree(existingValue interface{}, treeToMerge *YmlTreeModel) (interface{}, error) {
	var err error
	// DFS: first the includes in the specified order, then the including file
	for _, include := range treeToMerge.Includes {
		existingValue, err = mergeTree(existingValue, &include)
		if err != nil {
			return nil, err // TODO wrap? log?
		}
	}

	var parsedConfig interface{}
	err = yaml.Unmarshal([]byte(treeToMerge.Config), &parsedConfig)
	if err != nil {
		return nil, err // TODO wrap? log?
	}

	config := reflect.ValueOf(parsedConfig)

	return mergeValue(existingValue, config)
}

func mergeValue(existingValue interface{}, valueToMerge reflect.Value) (interface{}, error) {
	switch valueToMerge.Kind() {
	case reflect.Interface:
		return mergeValue(existingValue, valueToMerge.Elem())
	case reflect.Map:
		existingMap := existingValue.(map[interface{}]interface{})
		return mergeMap(existingMap, valueToMerge)
	case reflect.Slice, reflect.Array:
		existingArray := existingValue.([]interface{})
		return mergeArray(existingArray, valueToMerge)
	default:
		// Simple types
		return valueToMerge.Interface(), nil
	}
}

func mergeMap(existingMap map[interface{}]interface{}, mapToMerge reflect.Value) (map[interface{}]interface{}, error) {
	var err error
	iterator := mapToMerge.MapRange()
	for iterator.Next() {
		key := iterator.Key().Interface()
		valueToMerge := iterator.Value()

		existingValue, exists := existingMap[key]
		if exists {
			// Key exists in result and merged maps
			existingValue, err = mergeValue(existingValue, valueToMerge)
			if err != nil {
				return nil, err // TODO wrap? log?
			}
			existingMap[key] = existingValue
		} else {
			// Key doesn't exist in result yet
			existingMap[key] = valueToMerge.Interface()
		}
	}

	return existingMap, nil
}

func mergeArray(existingArray []interface{}, valueToMerge reflect.Value) ([]interface{}, error) {
	for i := 0; i < valueToMerge.Len(); i++ {
		valueToAdd := valueToMerge.Index(i)

		existingArray = append(existingArray, valueToAdd)
	}

	return existingArray, nil
}
