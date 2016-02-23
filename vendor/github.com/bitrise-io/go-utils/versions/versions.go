package versions

import (
	"log"
	"strconv"
	"strings"
)

// CompareVersions ...
// semantic version (X.Y.Z)
// 1 if version 2 is greater then version 1, -1 if not
// -2 & error if can't compare (if not supported component found)
func CompareVersions(version1, version2 string) (int, error) {
	version1Slice := strings.Split(version1, ".")
	version2Slice := strings.Split(version2, ".")

	lenDiff := len(version1Slice) - len(version2Slice)
	if lenDiff != 0 {
		makeDefVerComps := func(compLen int) []string {
			comps := make([]string, compLen, compLen)
			for idx := len(comps) - 1; idx >= 0; idx-- {
				comps[idx] = "0"
			}
			return comps
		}
		if lenDiff > 0 {
			// v1 slice is longer
			version2Slice = append(version2Slice, makeDefVerComps(lenDiff)...)
		} else {
			// v2 slice is longer
			version1Slice = append(version1Slice, makeDefVerComps(-lenDiff)...)
		}
	}

	cnt := len(version1Slice)
	for i, num := range version1Slice {
		num1, err := strconv.ParseInt(num, 0, 64)
		if err != nil {
			log.Println("Failed to parse int:", err)
			return -2, err
		}

		num2, err2 := strconv.ParseInt(version2Slice[i], 0, 64)
		if err2 != nil {
			log.Println("Failed to parse int:", err2)
			return -2, err
		}

		if num2 > num1 {
			return 1, nil
		} else if num2 < num1 {
			return -1, nil
		}

		if i == cnt-1 {
			// last one
			if num2 == num1 {
				return 0, nil
			}
		}
	}
	return -1, nil
}

// IsVersionBetween ...
//  returns true if it's between the lower and upper limit
//  or in case it matches the lower or the upper limit
func IsVersionBetween(verBase, verLower, verUpper string) (bool, error) {
	r1, err := CompareVersions(verBase, verLower)
	if err != nil {
		return false, err
	}
	if r1 == 1 {
		return false, nil
	}

	r2, err := CompareVersions(verBase, verUpper)
	if err != nil {
		return false, err
	}
	if r2 == -1 {
		return false, nil
	}

	return true, nil
}

// IsVersionGreaterOrEqual ...
//  returns true if verBase is greater or equal to verLower
func IsVersionGreaterOrEqual(verBase, verLower string) (bool, error) {
	r1, err := CompareVersions(verBase, verLower)
	if err != nil {
		return false, err
	}
	if r1 == 1 {
		return false, nil
	}

	return true, nil
}
