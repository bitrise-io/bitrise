package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/bitrise-io/go-utils/parseutil"
	"github.com/bitrise-io/go-utils/pointers"
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

func (env EnvironmentItemModel) GetKeyValuePairWithType() (string, interface{}, error) {
	var keys []string
	var values []interface{}

	for key, value := range env {
		keys = append(keys, key)
		values = append(values, value)
	}

	if len(keys) == 0 {
		return "", "", errors.New("no environment key specified")
	} else if len(keys) > 2 {
		sort.Strings(keys)
		return "", "", fmt.Errorf("more than 2 keys specified: %v", keys)
	}

	// Collect env key and value
	key := ""
	var value interface{}
	optionsFound := false

	for i := 0; i < len(keys); i++ {
		k := keys[i]
		if k != OptionsKey {
			key = k
			value = values[i]
		} else {
			optionsFound = true
		}
	}

	if key == "" {
		sort.Strings(keys)
		return "", "", fmt.Errorf("no environment key found, keys: %v", keys)
	}
	if len(keys) > 1 && !optionsFound {
		sort.Strings(keys)
		return "", "", fmt.Errorf("more than 1 environment key specified: %v", keys)
	}

	return key, value, nil
}

func (env EnvironmentItemModel) GetKeyValuePair() (string, string, error) {
	key, value, err := env.GetKeyValuePairWithType()
	if err != nil {
		return "", "", err
	}

	// Cast env value to string
	valueStr := ""

	if value != nil {
		if str, ok := value.(string); ok {
			valueStr = str
		} else if str := parseutil.CastToString(value); str != "" {
			valueStr = str
		} else {
			return "", "", fmt.Errorf("value (%#v) is not a string for key (%s)", value, key)
		}
	}

	return key, valueStr, nil
}

func (env EnvironmentItemModel) GetOptions() (EnvironmentItemOptionsModel, error) {
	value, found := env[OptionsKey]
	if !found {
		return EnvironmentItemOptionsModel{}, nil
	}

	if opts, ok := value.(EnvironmentItemOptionsModel); ok {
		return opts, nil
	}

	// if it's read from a file (YAML/JSON) then it's most likely not the proper type
	//  so cast it from the generic interface-interface map
	optsMap := make(map[string]interface{})
	if m, ok := value.(map[string]interface{}); ok {
		optsMap = m
	} else if m, ok := value.(map[interface{}]interface{}); ok {
		for k, v := range m {
			kStr, ok := k.(string)
			if !ok {
				return EnvironmentItemOptionsModel{}, fmt.Errorf("failed to cast option key (%#v) to string", k)
			}
			optsMap[kStr] = v
		}
	} else {
		return EnvironmentItemOptionsModel{}, errors.New("opts is not a map")
	}

	options := EnvironmentItemOptionsModel{}
	if err := options.ParseFromInterfaceMap(optsMap); err != nil {
		return EnvironmentItemOptionsModel{}, err
	}

	return options, nil

}

func (env EnvironmentItemModel) NormalizeValidateFillDefaults() error {
	if err := env.Normalize(); err != nil {
		return err
	}

	if err := env.Validate(); err != nil {
		return err
	}

	return env.FillMissingDefaults()
}

func (env EnvironmentItemModel) Normalize() error {
	opts, err := env.GetOptions()
	if err != nil {
		return err
	}
	env[OptionsKey] = opts
	return nil
}

func (env EnvironmentItemModel) Validate() error {
	_, _, err := env.GetKeyValuePair()
	if err != nil {
		return err
	}
	_, err = env.GetOptions()
	if err != nil {
		return err
	}
	return nil
}

func (env EnvironmentItemModel) FillMissingDefaults() error {
	options, err := env.GetOptions()
	if err != nil {
		return err
	}
	if options.Title == nil {
		options.Title = pointers.NewStringPtr("")
	}
	if options.Description == nil {
		options.Description = pointers.NewStringPtr("")
	}
	if options.Summary == nil {
		options.Summary = pointers.NewStringPtr("")
	}
	if options.Category == nil {
		options.Category = pointers.NewStringPtr("")
	}
	if options.ValueOptions == nil {
		options.ValueOptions = []string{}
	}
	if options.IsRequired == nil {
		options.IsRequired = pointers.NewBoolPtr(DefaultIsRequired)
	}
	if options.IsExpand == nil {
		options.IsExpand = pointers.NewBoolPtr(DefaultIsExpand)
	}
	if options.IsSensitive == nil {
		options.IsSensitive = pointers.NewBoolPtr(DefaultIsSensitive)
	}
	if options.IsDontChangeValue == nil {
		options.IsDontChangeValue = pointers.NewBoolPtr(DefaultIsDontChangeValue)
	}
	if options.IsTemplate == nil {
		options.IsTemplate = pointers.NewBoolPtr(DefaultIsTemplate)
	}
	if options.SkipIfEmpty == nil {
		options.SkipIfEmpty = pointers.NewBoolPtr(DefaultSkipIfEmpty)
	}
	if options.Meta == nil {
		options.Meta = map[string]interface{}{}
	}
	if options.Unset == nil {
		options.Unset = pointers.NewBoolPtr(DefaultUnset)
	}

	env[OptionsKey] = options
	return nil
}

func (envSerModel *EnvironmentItemOptionsModel) ParseFromInterfaceMap(input map[string]interface{}) error {
	for keyStr, value := range input {
		switch keyStr {
		case "title":
			envSerModel.Title = parseutil.CastToStringPtr(value)
		case "description":
			envSerModel.Description = parseutil.CastToStringPtr(value)
		case "summary":
			envSerModel.Summary = parseutil.CastToStringPtr(value)
		case "category":
			envSerModel.Category = parseutil.CastToStringPtr(value)
		case "value_options":
			castedValue, ok := value.([]string)
			if !ok {
				// try with []interface{} instead and cast the
				//  items to string
				castedValue = []string{}
				interfArr, ok := value.([]interface{})
				if !ok {
					return fmt.Errorf("invalid value type (%#v) for key: %s", value, keyStr)
				}
				for _, interfItm := range interfArr {
					castedItm, ok := interfItm.(string)
					if !ok {
						castedItm = parseutil.CastToString(interfItm)
						if castedItm == "" {
							return fmt.Errorf("not a string value (%#v) in value_options", interfItm)
						}
					}
					castedValue = append(castedValue, castedItm)
				}
			}
			envSerModel.ValueOptions = castedValue
		case "is_required":
			castedBoolPtr, ok := parseutil.CastToBoolPtr(value)
			if !ok {
				return fmt.Errorf("failed to parse bool value (%#v) for key (%s)", value, keyStr)
			}
			envSerModel.IsRequired = castedBoolPtr
		case "is_expand":
			castedBoolPtr, ok := parseutil.CastToBoolPtr(value)
			if !ok {
				return fmt.Errorf("failed to parse bool value (%#v) for key (%s)", value, keyStr)
			}
			envSerModel.IsExpand = castedBoolPtr
		case "is_sensitive":
			castedBoolPtr, ok := parseutil.CastToBoolPtr(value)
			if !ok {
				return fmt.Errorf("failed to parse bool value (%#v) for key (%s)", value, keyStr)
			}
			envSerModel.IsSensitive = castedBoolPtr
		case "is_dont_change_value":
			castedBoolPtr, ok := parseutil.CastToBoolPtr(value)
			if !ok {
				return fmt.Errorf("failed to parse bool value (%#v) for key (%s)", value, keyStr)
			}
			envSerModel.IsDontChangeValue = castedBoolPtr
		case "is_template":
			castedBoolPtr, ok := parseutil.CastToBoolPtr(value)
			if !ok {
				return fmt.Errorf("failed to parse bool value (%#v) for key (%s)", value, keyStr)
			}
			envSerModel.IsTemplate = castedBoolPtr
		case "skip_if_empty":
			castedBoolPtr, ok := parseutil.CastToBoolPtr(value)
			if !ok {
				return fmt.Errorf("failed to parse bool value (%#v) for key (%s)", value, keyStr)
			}
			envSerModel.SkipIfEmpty = castedBoolPtr
		case "unset":
			castedBoolPtr, ok := parseutil.CastToBoolPtr(value)
			if !ok {
				return fmt.Errorf("failed to parse bool value (%#v) for key (%s)", value, keyStr)
			}
			envSerModel.Unset = castedBoolPtr
		case "meta":
			metaValue, err := convertMetaValue(value)
			if err != nil {
				return fmt.Errorf("failed to parse meta: %s", err)
			}
			envSerModel.Meta = metaValue
		default:
			// intentional no-op case -- we just ignore unrecognized fields
		}
	}
	return nil
}

// Normalize - if successful this makes the model JSON serializable.
// Without this, if the object was created with e.g. a YAML parser,
// the type of `opts` might be a map[interface]interface, which is not JSON serializable.
// After this call it's ensured that the type of objects is map[string]interface,
// which is JSON serializable.
func (envsSerializeObj *EnvsSerializeModel) Normalize() error {
	for _, envObj := range envsSerializeObj.Envs {
		if err := envObj.Normalize(); err != nil {
			return err
		}
	}
	return nil
}

func convertMetaValue(metaValue any) (map[string]any, error) {
	converted, err := recursiveConvertToStringKeyedMap(metaValue)
	if err != nil {
		return nil, err
	}
	convertedMetaValue, ok := converted.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("meta value is not a map")
	}
	return convertedMetaValue, nil
}

func recursiveConvertToStringKeyedMap(source interface{}) (interface{}, error) {
	if array, ok := source.([]interface{}); ok {
		var convertedArray []interface{}
		for _, element := range array {
			convertedValue, err := recursiveConvertToStringKeyedMap(element)
			if err != nil {
				return nil, err
			}
			convertedArray = append(convertedArray, convertedValue)
		}
		return convertedArray, nil
	}

	if interfaceToInterfaceMap, ok := source.(map[interface{}]interface{}); ok {
		target := map[string]interface{}{}
		for key, value := range interfaceToInterfaceMap {
			strKey, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("failed to convert map key from type interface{} to string")
			}

			convertedValue, err := recursiveConvertToStringKeyedMap(value)
			if err != nil {
				return nil, err
			}
			target[strKey] = convertedValue
		}
		return target, nil
	}

	if stringToInterfaceMap, ok := source.(map[string]interface{}); ok {
		target := map[string]interface{}{}
		for key, value := range stringToInterfaceMap {
			convertedValue, err := recursiveConvertToStringKeyedMap(value)
			if err != nil {
				return nil, err
			}
			target[key] = convertedValue
		}
		return target, nil
	}

	return source, nil
}
