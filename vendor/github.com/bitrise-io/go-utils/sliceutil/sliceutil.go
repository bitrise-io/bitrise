package sliceutil

import "strings"

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

// CleanWhitespace removes leading and trailing white space from each element of the input slice.
// Elements that end up as empty strings are excluded from the result depending on the value of the omitEmpty flag.
func CleanWhitespace(list []string, omitEmpty bool) (items []string) {
	for _, e := range list {
		e = strings.TrimSpace(e)
		if !omitEmpty || len(e) > 0 {
			items = append(items, e)
		}
	}
	return
}
