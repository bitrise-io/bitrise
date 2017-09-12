package builtinutil

import (
	"fmt"
	"reflect"
)

// CastInterfaceToInterfaceSlice ...
func CastInterfaceToInterfaceSlice(slice interface{}) ([]interface{}, error) {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return []interface{}{}, fmt.Errorf("Input is not a slice: %#v", slice)
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret, nil
}

// DeepEqualSlices ...
func DeepEqualSlices(expected, actual []interface{}) bool {
	expectedMap := map[interface{}]bool{}
	for _, itm := range expected {
		expectedMap[itm] = true
	}
	actualMap := map[interface{}]bool{}
	for _, itm := range actual {
		actualMap[itm] = true
	}
	return reflect.DeepEqual(expectedMap, actualMap)
}
