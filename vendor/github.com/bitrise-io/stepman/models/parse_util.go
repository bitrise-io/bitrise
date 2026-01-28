package models

import (
	"fmt"
)

// JSONMarshallable replaces map[interface{}]interface{} with map[string]string recursively
// map[interface{}]interface{} is usually returned by parser go-yaml/v2
func JSONMarshallable(source map[string]interface{}) (map[string]interface{}, error) {
	target, err := RecursiveJSONMarshallable(source)
	if err != nil {
		return nil, err
	}
	castedTarget, ok := target.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("could not cast to map[string]interface{}")
	}
	return castedTarget, nil
}

// RecursiveJSONMarshallable recursively converts map[interface{}]interface{} to map[string]interface{}
func RecursiveJSONMarshallable(source interface{}) (interface{}, error) {
	if array, ok := source.([]interface{}); ok {
		var convertedArray []interface{}
		for _, element := range array {
			convertedValue, err := RecursiveJSONMarshallable(element)
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

			convertedValue, err := RecursiveJSONMarshallable(value)
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
			convertedValue, err := RecursiveJSONMarshallable(value)
			if err != nil {
				return nil, err
			}
			target[key] = convertedValue
		}
		return target, nil
	}

	return source, nil
}
