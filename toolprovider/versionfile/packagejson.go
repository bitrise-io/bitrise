package versionfile

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise/v2/toolprovider/alias"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// supportedEngines lists the engines fields that we support for tool installation (prio ordered).
var supportedEngines = []string{"node", "npm", "pnpm", "yarn"}

// parsePackageJSON parses a package.json file to extract tool version requirements.
// It reads the engines and the packageManager field.
func parsePackageJSON(path string) ([]ToolVersion, error) {
	config, err := readJSONFile(path)
	if err != nil {
		return nil, err
	}

	var result []ToolVersion
	indices := make(map[provider.ToolID]int) // track index of each tool in result

	// Parse engines field
	if enginesRaw, ok := config["engines"]; ok {
		engines, ok := enginesRaw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: 'engines' is not an object", path)
		}

		for _, name := range supportedEngines {
			versionStr, ok := engines[name]
			if !ok {
				continue
			}

			str, ok := versionStr.(string)
			if !ok || str == "" {
				continue
			}

			toolID := alias.GetCanonicalToolID(provider.ToolID(name))
			indices[toolID] = len(result)
			result = append(result, ToolVersion{
				ToolName:     toolID,
				Version:      str,
				IsConstraint: true,
			})
		}
	}

	// Parse packageManager field
	if pmRaw, ok := config["packageManager"]; ok {
		pmStr, ok := pmRaw.(string)
		if ok && pmStr != "" {
			name, ver, err := parsePackageManagerField(pmStr)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}

			toolID := alias.GetCanonicalToolID(provider.ToolID(name))
			// packageManager takes precedence over engines for the same tool
			if idx, ok := indices[toolID]; ok {
				result[idx].Version = ver
				result[idx].IsConstraint = false
			} else {
				indices[toolID] = len(result)
				result = append(result, ToolVersion{
					ToolName:     toolID,
					Version:      ver,
					IsConstraint: false,
				})
			}
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("%s: no tool version requirements found (no engines or packageManager field)", path)
	}

	return result, nil
}

// parsePackageManagerField parses a packageManager field value like "yarn@4.0.0" or "pnpm@8.0.0+sha256.abc".
func parsePackageManagerField(value string) (name, ver string, err error) {
	atIdx := strings.Index(value, "@")
	if atIdx < 1 {
		return "", "", fmt.Errorf("invalid packageManager value %q: expected 'name@version' format", value)
	}

	name = value[:atIdx]
	ver = value[atIdx+1:]

	// Strip hash suffix if present (e.g., "4.0.0+sha256.abc" → "4.0.0")
	if plusIdx := strings.Index(ver, "+"); plusIdx > 0 {
		ver = ver[:plusIdx]
	}

	if name == "" || ver == "" {
		return "", "", fmt.Errorf("invalid packageManager value %q: name and version must not be empty", value)
	}

	switch name {
	case "npm", "yarn", "pnpm":
		// valid
	default:
		return "", "", fmt.Errorf("unsupported package manager %q in packageManager field", name)
	}

	return name, ver, nil
}
