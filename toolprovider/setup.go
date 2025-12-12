package toolprovider

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise/v2/analytics"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/models"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/bitrise-io/bitrise/v2/toolprovider/versionfile"
)

// RunVersionFileSetup installs tools from version files.
func RunVersionFileSetup(versionFilePaths []string, tracker analytics.Tracker, silent bool) ([]provider.EnvironmentActivation, error) {
	toolRequests, err := makeToolRequests(versionFilePaths, silent)
	if err != nil {
		return nil, err
	}

	toolConfig := models.ToolConfigModel{
		Provider:                "mise",
		ExperimentalFastInstall: false,
	}

	return installTools(toolRequests, toolConfig, tracker, silent)
}

func makeToolRequests(versionFilePaths []string, silent bool) ([]provider.ToolRequest, error) {
	if len(versionFilePaths) == 0 {
		// If no version files specified, search in working directory.
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
		foundFiles, err := versionfile.FindVersionFiles(cwd)
		if err != nil {
			return nil, fmt.Errorf("find version files: %w", err)
		}

		if len(foundFiles) == 0 {
			if !silent {
				log.Warnf("No version files found in %s", cwd)
			}
			return nil, nil
		}

		versionFilePaths = foundFiles
		if !silent {
			log.Debugf("Found version files: %v", foundFiles)
		}
	}

	// Parse all version files.
	var allTools []versionfile.ToolVersion
	for _, versionFile := range versionFilePaths {
		absPath, err := filepath.Abs(versionFile)
		if err != nil {
			return nil, fmt.Errorf("resolve path %s: %w", versionFile, err)
		}

		if !silent {
			log.Debugf("Reading version file: %s", absPath)
		}
		tools, err := versionfile.Parse(absPath)
		if err != nil {
			return nil, fmt.Errorf("parse version file %s: %w", absPath, err)
		}

		allTools = append(allTools, tools...)
	}

	if len(allTools) == 0 {
		log.Warnf("No tools found in version files")
		return nil, nil
	}

	// Convert to tool requests.
	toolRequests := make([]provider.ToolRequest, 0, len(allTools))
	for _, tool := range allTools {
		v, strategy, err := ParseVersionString(tool.Version)
		if err != nil {
			return nil, fmt.Errorf("parse %s version %s: %w", tool.ToolName, tool.Version, err)
		}

		toolRequests = append(toolRequests, provider.ToolRequest{
			ToolName:           tool.ToolName,
			UnparsedVersion:    v,
			ResolutionStrategy: strategy,
			PluginURL:          nil,
		})
	}

	return toolRequests, nil
}
