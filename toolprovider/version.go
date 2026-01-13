package toolprovider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bitrise-io/bitrise/v2/models/yml"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// ParseVersionString takes a string like `3.12:latest` and parses it into a plain version string (3.12) and a ResolutionStrategy (latest released).
func ParseVersionString(versionString string) (string, provider.ResolutionStrategy, error) {
	versionString = strings.TrimSpace(versionString)

	latestSyntaxPattern, err := regexp.Compile(yml.ToolSyntaxPatternLatest)
	if err != nil {
		return "", 0, fmt.Errorf("compile regex pattern: %v", err)
	}
	preinstalledSyntaxPattern, err := regexp.Compile(yml.ToolSyntaxPatternInstalled)
	if err != nil {
		return "", 0, fmt.Errorf("compile regex pattern: %v", err)
	}

	var resolutionStrategy provider.ResolutionStrategy
	var plainVersion string
	if latestSyntaxPattern.MatchString(versionString) {
		resolutionStrategy = provider.ResolutionStrategyLatestReleased
		matches := latestSyntaxPattern.FindStringSubmatch(versionString)
		if len(matches) > 1 {
			plainVersion = matches[1]
		} else {
			return "", 0, fmt.Errorf("%s does not match :latest syntax", versionString)
		}
	} else if preinstalledSyntaxPattern.MatchString(versionString) {
		resolutionStrategy = provider.ResolutionStrategyLatestInstalled
		matches := preinstalledSyntaxPattern.FindStringSubmatch(versionString)
		if len(matches) > 1 {
			plainVersion = matches[1]
		} else {
			return "", 0, fmt.Errorf("%s does not match :installed syntax", versionString)
		}
	} else {
		resolutionStrategy = provider.ResolutionStrategyStrict
		plainVersion = versionString
	}

	return plainVersion, resolutionStrategy, nil
}
