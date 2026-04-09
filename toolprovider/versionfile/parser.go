package versionfile

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/v2/toolprovider/alias"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// ToolVersion represents a tool and its version from a version file, such as .tool-versions, .ruby-version, etc.
type ToolVersion struct {
	ToolName provider.ToolID
	Version  string

	// IsConstraint indicates that the version is a semver constraint (e.g., from package.json engines field)
	// rather than a plain version string.
	IsConstraint bool
}

func Parse(path string) ([]ToolVersion, error) {
	base := filepath.Base(path)
	switch base {
	case ".tool-versions":
		return parseToolVersionsFile(path)
	case ".fvmrc":
		return parseFVMRC(path)
	case ".nvmrc":
		return parseNVMRC(path)
	case "package.json":
		return parsePackageJSON(path)
	case "fvm_config.json":
		return parseFVMConfigJSON(path)
	default:
		tool, err := parseSingleToolVersion(path)
		if err != nil {
			return nil, err
		}
		return []ToolVersion{tool}, nil
	}
}

// FindVersionFiles searches for version files in the given directory,
// returns paths to found version files.
func FindVersionFiles(dir string) ([]string, error) {
	var versionFiles []string

	commonVersionFiles := []string{
		".tool-versions",
		".ruby-version",
		".node-version",
		".nvmrc",
		".python-version",
		".java-version",
		".go-version",
		".terraform-version",
		".kubectl-version",
		".fvmrc",
		".fvm/fvm_config.json",
		"package.json",
	}

	for _, filename := range commonVersionFiles {
		path := filepath.Join(dir, filename)
		if _, err := os.Stat(path); err == nil {
			versionFiles = append(versionFiles, path)
		}
	}

	return versionFiles, nil
}

// parseSingleToolVersion parses a version file that contains only a version string.
// Used for files like .ruby-version, .java-version, .node-version, etc.
// The tool name is inferred from the filename.
func parseSingleToolVersion(path string) (ToolVersion, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return ToolVersion{}, fmt.Errorf("read %s: %w", path, err)
	}

	version := strings.TrimSpace(string(content))
	if version == "" {
		return ToolVersion{}, fmt.Errorf("%s: empty version file", path)
	}

	if idx := strings.IndexAny(version, "\n\r"); idx != -1 {
		version = version[:idx]
		version = strings.TrimSpace(version)
		if version == "" {
			return ToolVersion{}, fmt.Errorf("%s: empty version on first line", path)
		}
	}

	filename := filepath.Base(path)
	toolName := inferToolID(filename)
	if toolName == "" {
		return ToolVersion{}, fmt.Errorf("%s: cannot infer tool name from filename", path)
	}

	return ToolVersion{
		ToolName: provider.ToolID(toolName),
		Version:  version,
	}, nil
}

// inferToolID extracts the tool name from a version filename.
func inferToolID(filename string) provider.ToolID {
	name := strings.TrimPrefix(filename, ".")
	name = strings.TrimSuffix(name, "-version")
	return alias.GetCanonicalToolID(provider.ToolID(name))
}

// flutterChannels are Flutter release channels that are not valid version specifiers
// for asdf/mise installation.
var flutterChannels = map[string]bool{
	"stable": true,
	"beta":   true,
	"dev":    true,
	"master": true,
	"main":   true,
}

// normalizeFlutterVersion converts FVM's "version@channel" format to "version-channel"
// and rejects channel-only values.
func normalizeFlutterVersion(version string) (string, error) {
	version = strings.Replace(version, "@", "-", 1)

	if flutterChannels[version] {
		return "", fmt.Errorf("channel-only value %q is not supported, specify a version to install the latest stable release", version)
	}

	return version, nil
}

// readJSONFile reads and unmarshals a JSON file into a map.
func readJSONFile(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var config map[string]any
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return config, nil
}

// extractStringValue extracts a non-empty string from a map by key.
func extractStringValue(config map[string]any, path string, key string) (string, error) {
	value, ok := config[key]
	if !ok {
		return "", fmt.Errorf("%s: missing '%s' key", path, key)
	}

	str, ok := value.(string)
	if !ok || str == "" {
		return "", fmt.Errorf("%s: '%s' key is not a non-empty string", path, key)
	}

	return str, nil
}

// parseNVMRC parses an NVM .nvmrc file to extract the Node.js version.
func parseNVMRC(path string) ([]ToolVersion, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	content = bytes.TrimSpace(content)

	if len(content) == 0 {
		return nil, fmt.Errorf("%s: empty version file", path)
	}

	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.Contains(line, "=") {
			continue
		}
		version := strings.TrimPrefix(line, "v")
		if version == "" {
			return nil, fmt.Errorf("%s: invalid version (empty after removing 'v' prefix)", path)
		}
		return []ToolVersion{
			{ToolName: "node", Version: version},
		}, nil
	}

	return nil, fmt.Errorf("%s: no valid version found", path)
}

// parseFVMRC parses an FVM 3.x .fvmrc JSON file to extract the Flutter version(s).
// Supports the main "flutter" key and optional "flavors" map.
func parseFVMRC(path string) ([]ToolVersion, error) {
	config, err := readJSONFile(path)
	if err != nil {
		return nil, err
	}

	mainVersion, err := extractStringValue(config, path, "flutter")
	if err != nil {
		return nil, err
	}

	mainVersion, err = normalizeFlutterVersion(mainVersion)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	// Collect unique versions to avoid duplicate installs
	seen := map[string]bool{mainVersion: true}
	tools := []ToolVersion{
		{ToolName: "flutter", Version: mainVersion},
	}

	if flavorsValue, ok := config["flavors"]; ok {
		flavors, ok := flavorsValue.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: 'flavors' is not a map", path)
		}

		for name, v := range flavors {
			vStr, ok := v.(string)
			if !ok || vStr == "" {
				return nil, fmt.Errorf("%s: flavor %q is not a non-empty string", path, name)
			}

			normalized, err := normalizeFlutterVersion(vStr)
			if err != nil {
				return nil, fmt.Errorf("%s: flavor %q: %w", path, name, err)
			}

			if !seen[normalized] {
				seen[normalized] = true
				tools = append(tools, ToolVersion{ToolName: "flutter", Version: normalized})
			}
		}
	}

	return tools, nil
}

// parsePackageJSON parses a package.json file to extract the Node.js version from the engines field.
// Package manager versions (npm, yarn, pnpm) are intentionally ignored as corepack handles those.
func parsePackageJSON(path string) ([]ToolVersion, error) {
	config, err := readJSONFile(path)
	if err != nil {
		return nil, err
	}

	enginesRaw, ok := config["engines"]
	if !ok {
		// No engines field is common in package.json, silently skip.
		return nil, nil
	}

	engines, ok := enginesRaw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: 'engines' is not an object", path)
	}

	nodeVersionRaw, ok := engines["node"]
	if !ok {
		// engines exists but no node field, silently skip.
		return nil, nil
	}

	nodeVersion, ok := nodeVersionRaw.(string)
	if !ok || nodeVersion == "" {
		return nil, fmt.Errorf("%s: engines.node is empty or not a string", path)
	}

	toolID := alias.GetCanonicalToolID(provider.ToolID("node"))

	return []ToolVersion{
		{
			ToolName:     toolID,
			Version:      nodeVersion,
			IsConstraint: true,
		},
	}, nil
}

// parseFVMConfigJSON parses a legacy .fvm/fvm_config.json file to extract the Flutter version.
// The file format is: {"flutterSdkVersion": "3.19.0"} or {"flutterSdkVersion": "3.19.0@stable"}
func parseFVMConfigJSON(path string) ([]ToolVersion, error) {
	config, err := readJSONFile(path)
	if err != nil {
		return nil, err
	}

	versionStr, err := extractStringValue(config, path, "flutterSdkVersion")
	if err != nil {
		return nil, err
	}

	normalized, err := normalizeFlutterVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return []ToolVersion{
		{ToolName: "flutter", Version: normalized},
	}, nil
}

// parseToolVersionsFile parses a .tool-versions file (asdf/mise format).
func parseToolVersionsFile(path string) ([]ToolVersion, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var tools []ToolVersion
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			return nil, fmt.Errorf("%s:%d: invalid format, expected '<tool> <version>'", path, lineNum)
		}

		toolName := parts[0]
		version := parts[1]

		tools = append(tools, ToolVersion{
			ToolName: provider.ToolID(toolName),
			Version:  version,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	return tools, nil
}
