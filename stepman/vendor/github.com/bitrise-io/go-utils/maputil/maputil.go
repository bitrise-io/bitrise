package maputil

// KeysOfStringInterfaceMap ...
func KeysOfStringInterfaceMap(m map[string]interface{}) []string {
	keys := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

// KeysOfStringStringMap ...
func KeysOfStringStringMap(m map[string]string) []string {
	keys := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

// CloneStringStringMap ...
func CloneStringStringMap(m map[string]string) map[string]string {
	cloneMap := map[string]string{}
	for k, v := range m {
		cloneMap[k] = v
	}
	return cloneMap
}

// MergeStringStringMap - returns a new map object,
//  DOES NOT modify either input map
func MergeStringStringMap(m1, m2 map[string]string) map[string]string {
	mergedMap := CloneStringStringMap(m1)
	for k, v := range m2 {
		mergedMap[k] = v
	}
	return mergedMap
}
