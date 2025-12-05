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

// SetupOptions contains options for tool setup.
type SetupOptions struct {
	// VersionFiles is a list of version file paths to read tools from.
	VersionFiles []string

	// WorkingDir is the directory to search for version files if not explicitly provided.
	WorkingDir string

	// ProviderName is the tool provider to use (asdf, mise).
	ProviderName string

	// ExperimentalFastInstall enables experimental fast install.
	ExperimentalFastInstall bool

	// ExtraPlugins contains additional plugin sources.
	ExtraPlugins map[models.ToolID]string
}

// SetupFromVersionFiles installs tools from version files.
func SetupFromVersionFiles(opts SetupOptions, tracker analytics.Tracker, silent bool) ([]provider.EnvironmentActivation, error) {
	// If no version files specified, search in working directory.
	if len(opts.VersionFiles) == 0 {
		if opts.WorkingDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("get working directory: %w", err)
			}
			opts.WorkingDir = cwd
		}

		foundFiles, err := versionfile.FindVersionFiles(opts.WorkingDir)
		if err != nil {
			return nil, fmt.Errorf("find version files: %w", err)
		}

		if len(foundFiles) == 0 {
			if !silent {
				log.Warnf("No version files found in %s", opts.WorkingDir)
			}
			return nil, nil
		}

		opts.VersionFiles = foundFiles
		log.Debugf("Found version files: %v", foundFiles)
	}

	// Parse all version files.
	var allTools []versionfile.ToolVersion
	for _, versionFile := range opts.VersionFiles {
		absPath, err := filepath.Abs(versionFile)
		if err != nil {
			return nil, fmt.Errorf("resolve path %s: %w", versionFile, err)
		}

		log.Debugf("Reading version file: %s", absPath)
		tools, err := versionfile.ParseVersionFile(absPath)
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

		var pluginURL *string
		if opts.ExtraPlugins != nil {
			if url, ok := opts.ExtraPlugins[models.ToolID(tool.ToolName)]; ok {
				pluginURL = &url
			}
		}

		toolRequests = append(toolRequests, provider.ToolRequest{
			ToolName:           tool.ToolName,
			UnparsedVersion:    v,
			ResolutionStrategy: strategy,
			PluginURL:          pluginURL,
		})
	}

	toolConfig := models.ToolConfigModel{
		Provider:                opts.ProviderName,
		ExperimentalFastInstall: opts.ExperimentalFastInstall,
		ExtraPlugins:            opts.ExtraPlugins,
	}

	if toolConfig.Provider == "" {
		toolConfig.Provider = "mise"
	}

	return installTools(toolRequests, toolConfig, tracker, silent)
}
