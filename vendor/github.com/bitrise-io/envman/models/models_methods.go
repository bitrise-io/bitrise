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
	//DefaultIsRequired ...
	DefaultIsRequired = false
	// DefaultIsDontChangeValue ...
	DefaultIsDontChangeValue = false
	// DefaultIsTemplate ...
	DefaultIsTemplate = false
	// DefaultSkipIfEmpty ...
	DefaultSkipIfEmpty = false
)

// NewEnvJSONList ...
func NewEnvJSONList(jsonStr string) (EnvsJSONListModel, error) {
	list := EnvsJSONListModel{}
	if err := json.Unmarshal([]byte(jsonStr), &list); err != nil {
		return EnvsJSONListModel{}, err
	}
	return list, nil
}

// GetKeyValuePair ...
func (env EnvironmentItemModel) GetKeyValuePair() (string, string, error) {
	// Collect keys and values
	keys := []string{}
	values := []interface{}{}

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

// ParseFromInterfaceMap ...
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
		default:
			return fmt.Errorf("not supported key found in options: %s", keyStr)
		}
	}
	return nil
}

// GetOptions ...
func (env EnvironmentItemModel) GetOptions() (EnvironmentItemOptionsModel, error) {
	value, found := env[OptionsKey]
	if !found {
		return EnvironmentItemOptionsModel{}, nil
	}

	envItmCasted, ok := value.(EnvironmentItemOptionsModel)
	if ok {
		return envItmCasted, nil
	}

	// if it's read from a file (YAML/JSON) then it's most likely not the proper type
	//  so cast it from the generic interface-interface map
	normalizedOptsInterfaceMap := make(map[string]interface{})
	isNormalizeOK := false
	if optionsInterfaceMap, ok := value.(map[interface{}]interface{}); ok {
		// Try to normalize every key to String
		for key, value := range optionsInterfaceMap {
			keyStr, ok := key.(string)
			if !ok {
				return EnvironmentItemOptionsModel{}, fmt.Errorf("failed to cask options key (%#v) to string", key)
			}
			normalizedOptsInterfaceMap[keyStr] = value
		}
		isNormalizeOK = true
	} else {
		if castedTmp, ok := value.(map[string]interface{}); ok {
			normalizedOptsInterfaceMap = castedTmp
			isNormalizeOK = true
		}
	}

	if isNormalizeOK {
		options := EnvironmentItemOptionsModel{}
		err := options.ParseFromInterfaceMap(normalizedOptsInterfaceMap)
		if err != nil {
			return EnvironmentItemOptionsModel{}, err
		}

		return options, nil
	}

	return EnvironmentItemOptionsModel{}, fmt.Errorf("failed to cast options value: (%#v)", value)
}

// Normalize ...
func (env *EnvironmentItemModel) Normalize() error {
	opts, err := env.GetOptions()
	if err != nil {
		return err
	}
	(*env)[OptionsKey] = opts
	return nil
}

// Normalize - if successful this makes the model JSON serializable.
// Without this, if the object was created with e.g. a YAML parser,
// the type of `opts` might be map[interface]interface, which is not JSON serializable.
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

// FillMissingDefaults ...
func (env *EnvironmentItemModel) FillMissingDefaults() error {
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
	if options.IsRequired == nil {
		options.IsRequired = pointers.NewBoolPtr(DefaultIsRequired)
	}
	if options.IsExpand == nil {
		options.IsExpand = pointers.NewBoolPtr(DefaultIsExpand)
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
	(*env)[OptionsKey] = options
	return nil
}

// Validate ...
func (env EnvironmentItemModel) Validate() error {
	key, _, err := env.GetKeyValuePair()
	if err != nil {
		return err
	}
	if key == "" {
		return errors.New("no environment key found")
	}
	_, err = env.GetOptions()
	if err != nil {
		return err
	}
	return nil
}

// NormalizeValidateFillDefaults ...
func (env EnvironmentItemModel) NormalizeValidateFillDefaults() error {
	if err := env.Normalize(); err != nil {
		return err
	}

	if err := env.Validate(); err != nil {
		return err
	}

	if err := env.FillMissingDefaults(); err != nil {
		return err
	}
	return nil
}
