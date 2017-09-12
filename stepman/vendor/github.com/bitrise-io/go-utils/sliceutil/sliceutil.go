package sliceutil

// UniqueStringSlice - returns a cleaned up list,
// where every item is unique.
// Does NOT guarantee any ordering, the result can
// be in any order!
func UniqueStringSlice(strs []string) []string {
	lookupMap := map[string]interface{}{}
	for _, aStr := range strs {
		lookupMap[aStr] = 1
	}
	uniqueStrs := []string{}
	for k := range lookupMap {
		uniqueStrs = append(uniqueStrs, k)
	}
	return uniqueStrs
}

// IndexOfStringInSlice ...
func IndexOfStringInSlice(searchFor string, searchIn []string) int {
	for idx, anItm := range searchIn {
		if anItm == searchFor {
			return idx
		}
	}
	return -1
}

// IsStringInSlice ...
func IsStringInSlice(searchFor string, searchIn []string) bool {
	return IndexOfStringInSlice(searchFor, searchIn) >= 0
}
