package versionfile

import (
	"bufio"
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
}

func Parse(path string) ([]ToolVersion, error) {
	if filepath.Base(path) == ".tool-versions" {
		return parseToolVersionsFile(path)
	}

	tool, err := parseSingleToolVersion(path)
	if err != nil {
		return nil, err
	}

	return []ToolVersion{tool}, nil
}

// FindVersionFiles searches for version files in the given directory,
// returns paths to found version files.
func FindVersionFiles(dir string) ([]string, error) {
	var versionFiles []string

	commonVersionFiles := []string{
		".tool-versions",
		".ruby-version",
		".node-version",
		".python-version",
		".java-version",
		".go-version",
		".terraform-version",
		".kubectl-version",
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
