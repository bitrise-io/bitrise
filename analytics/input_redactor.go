package analytics

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"

	"github.com/bitrise-io/bitrise/bitrise"
	bitriseModels "github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/bitrise/tools/filterwriter"
	"github.com/bitrise-io/envman/env"
	"github.com/bitrise-io/envman/models"
)

// InputRedactor ...
type InputRedactor interface {
	Redact(inputs []models.EnvironmentItemModel) (map[string]interface{}, error)
}

type inputRedactor struct {
	secrets           []string
	environment       []models.EnvironmentItemModel
	buildRunResults   bitriseModels.BuildRunResultsModel
	isCIMode          bool
	isPullRequestMode bool
	envSource         env.EnvironmentSource
}

// NewInputRedactor ...
func NewInputRedactor(secrets []string, environment []models.EnvironmentItemModel, buildRunResults bitriseModels.BuildRunResultsModel, isCIMode bool, isPullRequestMode bool, envSource env.EnvironmentSource) InputRedactor {
	return inputRedactor{secrets: secrets, environment: environment, buildRunResults: buildRunResults, isCIMode: isCIMode, isPullRequestMode: isPullRequestMode, envSource: envSource}
}

// Redact ...
func (i inputRedactor) Redact(inputs []models.EnvironmentItemModel) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for _, input := range inputs {
		options, key, value, err := getData(input)
		if err != nil {
			return nil, err
		}
		raw := map[string]interface{}{}
		expanded := map[string]interface{}{}
		if err := i.fillMap(options, raw, expanded, key, reflect.ValueOf(value)); err != nil {
			return nil, err
		}
		var finalValue interface{}
		if v, ok := expanded[key]; ok {
			finalValue = v
		} else {
			finalValue = raw[key]
		}
		valueMap := map[string]interface{}{"value": finalValue}
		result[key] = valueMap
		if !reflect.DeepEqual(raw[key], expanded[key]) {
			valueMap["original_value"] = raw[key]
		}
	}
	return result, nil
}

func (i inputRedactor) growSlice(options models.EnvironmentItemOptionsModel, container []interface{}, expandedContainer []interface{}, value reflect.Value) ([]interface{}, []interface{}, error) {
	for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		acc, expandedAcc, err := i.collectSliceItems(options, value)
		if err != nil {
			return nil, nil, err
		}
		return append(container, acc), append(expandedContainer, expandedAcc), nil
	case reflect.Map:
		acc, expandedAcc, err := i.collectMapItems(options, value)
		if err != nil {
			return nil, nil, err
		}
		return append(container, acc), append(expandedContainer, expandedAcc), nil
	case reflect.String:
		v, e, err := i.expandAndRedactString(options, value)
		if err != nil {
			return nil, nil, err
		}
		return append(container, v), append(expandedContainer, e), nil
	default:
		if value.IsValid() && value.CanInterface() {
			return append(container, value.Interface()), append(expandedContainer, value.Interface()), nil
		} else {
			return append(container, ""), append(expandedContainer, ""), nil
		}
	}
}

func (i inputRedactor) fillMap(options models.EnvironmentItemOptionsModel, container map[string]interface{}, expandedContainer map[string]interface{}, key string, value reflect.Value) error {
	for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		acc, expandedAcc, err := i.collectSliceItems(options, value)
		if err != nil {
			return err
		}
		container[key] = acc
		expandedContainer[key] = expandedAcc
	case reflect.Map:
		acc, expandedAcc, err := i.collectMapItems(options, value)
		if err != nil {
			return err
		}
		container[key] = acc
		expandedContainer[key] = expandedAcc
	case reflect.String:
		v, e, err := i.expandAndRedactString(options, value)
		if err != nil {
			return err
		}
		container[key] = v
		expandedContainer[key] = e
	default:
		if value.IsValid() && value.CanInterface() {
			container[key] = value.Interface()
			expandedContainer[key] = value.Interface()
		} else {
			container[key] = ""
			expandedContainer[key] = ""
		}
	}
	return nil
}

func (i inputRedactor) collectMapItems(options models.EnvironmentItemOptionsModel, value reflect.Value) (map[string]interface{}, map[string]interface{}, error) {
	acc := map[string]interface{}{}
	expandedAcc := map[string]interface{}{}
	iter := value.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		if err := i.fillMap(options, acc, expandedAcc, fmt.Sprintf("%s", k), v); err != nil {
			return nil, nil, err
		}
	}
	return acc, expandedAcc, nil
}

func (i inputRedactor) collectSliceItems(options models.EnvironmentItemOptionsModel, value reflect.Value) ([]interface{}, []interface{}, error) {
	var acc []interface{}
	var expandedAcc []interface{}
	var err error
	for j := 0; j < value.Len(); j++ {
		acc, expandedAcc, err = i.growSlice(options, acc, expandedAcc, value.Index(j))
		if err != nil {
			return nil, nil, err
		}
	}
	return acc, expandedAcc, nil
}

func (i inputRedactor) expandAndRedactString(options models.EnvironmentItemOptionsModel, value reflect.Value) (string, string, error) {
	v := value.String()
	e, err := i.expand(options, v)
	if err != nil {
		return "", "", err
	}
	v, err = RedactStringWithSecret(v, i.secrets)
	if err != nil {
		return "", "", err
	}
	e, err = RedactStringWithSecret(e, i.secrets)
	if err != nil {
		return "", "", err
	}
	return v, e, nil
}

func (i inputRedactor) expand(options models.EnvironmentItemOptionsModel, value string) (string, error) {
	result := value
	var envs map[string]string
	if options.IsTemplate != nil && *options.IsTemplate {
		if envs == nil {
			effects, err := env.GetDeclarationsSideEffects(i.environment, i.envSource)
			if err != nil {
				return "", err
			}
			envs = effects.ResultEnvironment
		}

		evaluatedValue, err := bitrise.EvaluateTemplateToString(result, i.isCIMode, i.isPullRequestMode, i.buildRunResults, envs)
		if err == nil {
			result = evaluatedValue
		}
	}
	if options.IsExpand != nil && *options.IsExpand {
		if envs == nil {
			effects, err := env.GetDeclarationsSideEffects(i.environment, i.envSource)
			if err != nil {
				return "", err
			}
			envs = effects.ResultEnvironment
		}
		result = os.Expand(result, func(s string) string {
			if _, ok := envs[s]; !ok {
				return ""
			}
			return envs[s]
		})
	}
	return result, nil
}

func getData(input models.EnvironmentItemModel) (models.EnvironmentItemOptionsModel, string, interface{}, error) {
	options, err := input.GetOptions()
	if err != nil {
		return models.EnvironmentItemOptionsModel{}, "", nil, err
	}
	var keys []string
	var values []interface{}

	for key, value := range input {
		keys = append(keys, key)
		values = append(values, value)
	}

	if len(keys) == 0 {
		return models.EnvironmentItemOptionsModel{}, "", nil, errors.New("no environment key specified")
	} else if len(keys) > 2 {
		sort.Strings(keys)
		return models.EnvironmentItemOptionsModel{}, "", nil, fmt.Errorf("more than 2 keys specified: %v", keys)
	}

	// Collect env key and value
	key := ""
	var value interface{}
	optionsFound := false

	for i := 0; i < len(keys); i++ {
		k := keys[i]
		if k != models.OptionsKey {
			key = k
			value = values[i]
		} else {
			optionsFound = true
		}
	}

	if key == "" {
		sort.Strings(keys)
		return models.EnvironmentItemOptionsModel{}, "", nil, fmt.Errorf("no environment key found, keys: %v", keys)
	}
	if len(keys) > 1 && !optionsFound {
		sort.Strings(keys)
		return models.EnvironmentItemOptionsModel{}, "", nil, fmt.Errorf("more than 1 environment key specified: %v", keys)
	}
	return options, key, value, nil
}

// RedactStringWithSecret ...
func RedactStringWithSecret(value string, secrets []string) (string, error) {
	src := bytes.NewReader([]byte(value))
	dstBuf := new(bytes.Buffer)
	secretFilterDst := filterwriter.New(secrets, dstBuf)

	if _, err := io.Copy(secretFilterDst, src); err != nil {
		return "", fmt.Errorf("failed to redact secrets, stream copy failed: %s", err)
	}
	if _, err := secretFilterDst.Flush(); err != nil {
		return "", fmt.Errorf("failed to redact secrets, stream flush failed: %s", err)
	}
	return dstBuf.String(), nil
}
