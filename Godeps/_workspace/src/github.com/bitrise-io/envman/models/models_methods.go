package models

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/pointers"
)

const (
	// OptionsKey ...
	OptionsKey = "opts"
)

const (
	//DefaultIsRequired ...
	DefaultIsRequired = false
	// DefaultIsExpand ...
	DefaultIsExpand = true
	// DefaultIsDontChangeValue ...
	DefaultIsDontChangeValue = false
)

// GetKeyValuePair ...
func (env EnvironmentItemModel) GetKeyValuePair() (string, string, error) {
	if len(env) > 2 {
		return "", "", fmt.Errorf("Invalid env: more than 2 fields: %#v", env)
	}

	retKey := ""
	retValue := ""

	for key, value := range env {
		if key != OptionsKey {
			if retKey != "" {
				return "", "", fmt.Errorf("Invalid env: more than 1 key-value field found: %#v", env)
			}

			valueStr, ok := value.(string)
			if !ok {
				if value == nil {
					valueStr = ""
				} else {
					return "", "", fmt.Errorf("Invalid value, not a string (key:%#v) (value:%#v)", key, value)
				}
			}

			retKey = key
			retValue = valueStr
		}
	}

	if retKey == "" {
		return "", "", errors.New("Invalid env: no envKey specified!")
	}

	return retKey, retValue, nil
}

// ParseFromInterfaceMap ...
func (envSerModel *EnvironmentItemOptionsModel) ParseFromInterfaceMap(input map[string]interface{}) error {
	for keyStr, value := range input {
		log.Debugf("  ** processing (key:%#v) (value:%#v) (envSerModel:%#v)", keyStr, value, envSerModel)
		switch keyStr {
		case "title":
			castedValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.Title = pointers.NewStringPtr(castedValue)
		case "description":
			castedValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.Description = pointers.NewStringPtr(castedValue)
		case "summary":
			castedValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.Summary = pointers.NewStringPtr(castedValue)
		case "value_options":
			castedValue, ok := value.([]string)
			if !ok {
				// try with []interface{} instead and cast the
				//  items to string
				castedValue = []string{}
				interfArr, ok := value.([]interface{})
				if !ok {
					return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
				}
				for _, interfItm := range interfArr {
					castedItm, ok := interfItm.(string)
					if !ok {
						return fmt.Errorf("Invalid value in value_options (%#v), not a string: %#v", interfArr, interfItm)
					}
					castedValue = append(castedValue, castedItm)
				}
			}
			envSerModel.ValueOptions = castedValue
		case "is_required":
			castedValue, ok := value.(bool)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.IsRequired = pointers.NewBoolPtr(castedValue)
		case "is_expand":
			castedValue, ok := value.(bool)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.IsExpand = pointers.NewBoolPtr(castedValue)
		case "is_dont_change_value":
			castedValue, ok := value.(bool)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.IsDontChangeValue = pointers.NewBoolPtr(castedValue)
		default:
			return fmt.Errorf("Not supported key found in options: %#v", keyStr)
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

	log.Debugf(" * processing env:%#v", env)

	// if it's read from a file (YAML/JSON) then it's most likely not the proper type
	//  so cast it from the generic interface-interface map
	normalizedOptsInterfaceMap := make(map[string]interface{})
	isNormalizeOK := false
	if optionsInterfaceMap, ok := value.(map[interface{}]interface{}); ok {
		// Try to normalize every key to String
		for key, value := range optionsInterfaceMap {
			keyStr, ok := key.(string)
			if !ok {
				return EnvironmentItemOptionsModel{}, fmt.Errorf("Failed to cask Options key to String: %#v", key)
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

		log.Debugf("Parsed options: %#v\n", options)
		return options, nil
	}

	return EnvironmentItemOptionsModel{}, fmt.Errorf("Invalid options (value:%#v) - failed to cast", value)
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

// FillMissingDefaults ...
func (env *EnvironmentItemModel) FillMissingDefaults() error {
	options, err := env.GetOptions()
	if err != nil {
		return err
	}
	if options.Description == nil {
		options.Description = pointers.NewStringPtr("")
	}
	if options.Summary == nil {
		options.Summary = pointers.NewStringPtr("")
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
		return errors.New("Invalid environment: empty env_key")
	}
	_, err = env.GetOptions()
	if err != nil {
		return err
	}
	return nil
}

// NormalizeEnvironmentItemModel ...
func (env EnvironmentItemModel) NormalizeEnvironmentItemModel() error {
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
