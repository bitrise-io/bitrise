package models

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var ToolProviders = []string{"asdf", "mise"}

type ToolID string

// ToolsModel is a mapping of tool IDs to their versions (see package toolprovider about the version syntax)
type ToolsModel map[ToolID]string

type ToolConfigModel struct {
	Provider string `json:"provider,omitempty" yaml:"provider,omitempty"`

	// Extra tool-plugins on top of Bitrise-vetted integrations. This is very provider-specific, but the map value is a URL to the plugin source.
	ExtraPlugins map[ToolID]string `json:"extra_plugins,omitempty" yaml:"extra_plugins,omitempty"`
}

const ToolSyntaxPatternLatest = `(.*):latest$`
const ToolSyntaxPatternInstalled = `(.*):installed$`

func isProviderSupported(providerName string) bool {
	return slices.Contains(ToolProviders, providerName) || providerName == "" // Provider is optional, there is a default provider.
}

func validateToolConfig(toolConfig *ToolConfigModel) error {
	if toolConfig == nil {
		return nil
	}

	if !isProviderSupported(toolConfig.Provider) {
		return fmt.Errorf("invalid provider: %s, should be one of: %v", toolConfig.Provider, ToolProviders)
	}

	for id, url := range toolConfig.ExtraPlugins {
		if url == "" {
			return fmt.Errorf("URL of extra plugin %s is empty", id)
		}
	}

	return nil
}

func validateTools(config *BitriseDataModel) error {
	if config == nil {
		return nil
	}

	for toolID, versionString := range config.Tools {
		err := validateVersionString(versionString)
		if err != nil {
			return fmt.Errorf("%s: invalid version syntax %s: %w", toolID, versionString, err)
		}
	}

	for _, wf := range config.Workflows {
		for toolID, versionString := range wf.Tools {
			err := validateVersionString(versionString)
			if err != nil {
				return fmt.Errorf("%s: invalid version syntax %s: %w", toolID, versionString, err)
			}
		}
	}

	return nil
}

// validateVersionString takes a string like `3.12:latest` or `3.12.0` and validates it against the expected syntax.
func validateVersionString(versionString string) error {
	versionString = strings.TrimSpace(versionString)

	latestSyntaxPattern, err := regexp.Compile(ToolSyntaxPatternLatest)
	if err != nil {
		return fmt.Errorf("compile regex pattern: %v", err)
	}
	preinstalledSyntaxPattern, err := regexp.Compile(ToolSyntaxPatternInstalled)
	if err != nil {
		return fmt.Errorf("compile regex pattern: %v", err)
	}

	if latestSyntaxPattern.MatchString(versionString) {
		matches := latestSyntaxPattern.FindStringSubmatch(versionString)
		if len(matches) <= 1 {
			return fmt.Errorf("%s does not match version:latest syntax", versionString)
		}
	} else if preinstalledSyntaxPattern.MatchString(versionString) {
		matches := preinstalledSyntaxPattern.FindStringSubmatch(versionString)
		if len(matches) <= 1 {
			return fmt.Errorf("%s does not match version:installed syntax", versionString)
		}
	}

	// Input doesn't match any of the special patterns, this is treated as a plain version string. No further validation is done here.
	return nil
}
