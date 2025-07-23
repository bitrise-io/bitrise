package toolprovider

import (
	"fmt"
	"regexp"
	"strings"
)

const latestSyntaxPattern = `(.*):latest$`
const installedSyntaxPattern = `(.*):installed$`

// ParseVersionString takes a string like `3.12:latest` and parses it into a plain version string (3.12) and a ResolutionStrategy (latest released).
func ParseVersionString(versionString string) (string, ResolutionStrategy, error) {
	versionString = strings.TrimSpace(versionString)

	latestSyntaxPattern, err := regexp.Compile(latestSyntaxPattern)
	if err != nil {
		return "", 0, fmt.Errorf("compile regex pattern: %v", err)
	}
	preinstalledSyntaxPattern, err := regexp.Compile(installedSyntaxPattern)
	if err != nil {
		return "", 0, fmt.Errorf("compile regex pattern: %v", err)
	}

	var resolutionStrategy ResolutionStrategy
	var plainVersion string
	if latestSyntaxPattern.MatchString(versionString) {
		resolutionStrategy = ResolutionStrategyLatestReleased
		matches := latestSyntaxPattern.FindStringSubmatch(versionString)
		if len(matches) > 1 {
			plainVersion = matches[1]
		} else {
			return "", 0, fmt.Errorf("%s does not match :latest syntax", versionString)
		}
	} else if preinstalledSyntaxPattern.MatchString(versionString) {
		resolutionStrategy = ResolutionStrategyLatestInstalled
		matches := preinstalledSyntaxPattern.FindStringSubmatch(versionString)
		if len(matches) > 1 {
			plainVersion = matches[1]
		} else {
			return "", 0, fmt.Errorf("%s does not match :installed syntax", versionString)
		}
	} else {
		resolutionStrategy = ResolutionStrategyStrict
		plainVersion = versionString
	}

	return plainVersion, resolutionStrategy, nil
}
