package models

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

const (
	// OptionsKey ...
	OptionsKey = "opts"
)

var (
	//DefaultIsRequired ...
	DefaultIsRequired = false
	// DefaultIsExpand ...
	DefaultIsExpand = true
	// DefaultIsDontChangeValue ...
	DefaultIsDontChangeValue = false
	// DefaultIsAlwaysRun ...
	DefaultIsAlwaysRun = false
	// DefaultIsRequiresAdminUser ...
	DefaultIsRequiresAdminUser = false
	// DefaultIsSkippable ...
	DefaultIsSkippable = false
)

// GetKeyValuePair ...
func (env EnvironmentItemModel) GetKeyValuePair() (string, string, error) {
	if len(env) > 2 {
		return "", "", errors.New("Invalid env: more then 2 fields ")
	}

	retKey := ""
	retValue := ""

	for key, value := range env {
		if key != OptionsKey {
			if retKey != "" {
				return "", "", errors.New("Invalid env: more then 1 key-value field found!")
			}

			valueStr, ok := value.(string)
			if !ok {
				if value == nil {
					valueStr = ""
				} else {
					return "", "", fmt.Errorf("Invalid value (key:%#v) (value:%#v)", key, value)
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
func (envSerModel *EnvironmentItemOptionsModel) ParseFromInterfaceMap(input map[interface{}]interface{}) error {
	for key, value := range input {
		keyStr, ok := key.(string)
		if !ok {
			return fmt.Errorf("Invalid key, should be a string: %#v", key)
		}
		log.Debugf("  ** processing (key:%#v) (value:%#v) (envSerModel:%#v)", key, value, envSerModel)
		switch keyStr {
		case "title":
			castedValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.Title = &castedValue
		case "description":
			castedValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.Description = &castedValue
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
			envSerModel.IsRequired = &castedValue
		case "is_expand":
			castedValue, ok := value.(bool)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.IsExpand = &castedValue
		case "is_dont_change_value":
			castedValue, ok := value.(bool)
			if !ok {
				return fmt.Errorf("Invalid value type (key:%s): %#v", keyStr, value)
			}
			envSerModel.IsDontChangeValue = &castedValue
		default:
			return fmt.Errorf("Not supported key found in options: %#v", key)
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
	optionsInterfaceMap, ok := value.(map[interface{}]interface{})
	if !ok {
		return EnvironmentItemOptionsModel{}, fmt.Errorf("Invalid options (value:%#v) - failed to map-interface cast", value)
	}

	options := EnvironmentItemOptionsModel{}
	err := options.ParseFromInterfaceMap(optionsInterfaceMap)
	if err != nil {
		return EnvironmentItemOptionsModel{}, err
	}

	log.Debugf("Parsed options: %#v\n", options)

	return options, nil
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
	defaultString := ""

	options, err := env.GetOptions()
	if err != nil {
		return err
	}
	if options.Description == nil {
		options.Description = &defaultString
	}
	if options.IsRequired == nil {
		options.IsRequired = &DefaultIsRequired
	}
	if options.IsExpand == nil {
		options.IsExpand = &DefaultIsExpand
	}
	if options.IsDontChangeValue == nil {
		options.IsDontChangeValue = &DefaultIsDontChangeValue
	}
	(*env)[OptionsKey] = options
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
