package asdf

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
)

// Unlikely to conflict with any plugin name or URL, but clearly separates the plugin name and URL.
const PluginSourceSeparator = "::"

type PluginSource struct {
	PluginName  provider.ToolID
	GitCloneURL string
}

var pluginSourceMap = map[provider.ToolID]PluginSource{
	"flutter": {PluginName: "flutter", GitCloneURL: "https://github.com/asdf-community/asdf-flutter.git"},
	"golang":  {PluginName: "golang", GitCloneURL: "https://github.com/asdf-community/asdf-golang.git"},
	"nodejs":  {PluginName: "nodejs", GitCloneURL: "https://github.com/asdf-vm/asdf-nodejs.git"},
	"python":  {PluginName: "python", GitCloneURL: "https://github.com/danhper/asdf-python.git"},
	"ruby":    {PluginName: "ruby", GitCloneURL: "https://github.com/asdf-vm/asdf-ruby.git"},
	"tuist":   {PluginName: "tuist", GitCloneURL: "https://github.com/tuist/asdf-tuist.git"},
}

// InstallPlugin installs a plugin for the specified tool, if needed.
//
// It resolves the plugin source from the tool request or predefined map,
// checks if the plugin is already installed, and if not, installs it using asdf.
func (a AsdfToolProvider) InstallPlugin(tool provider.ToolRequest) error {
	plugin, err := fetchPluginSource(tool)
	if err != nil {
		// E.g. parse error while resolving plugin source.
		return provider.ToolInstallError{
			ToolName:         tool.ToolName,
			RequestedVersion: tool.UnparsedVersion,
			Cause:            fmt.Sprintf("Couldn't resolve plugin source: %s", err),
			Recommendation:   "Review the syntax of the `plugin` field",
		}
	}
	if plugin == nil {
		return provider.ToolInstallError{
			ToolName:         tool.ToolName,
			RequestedVersion: tool.UnparsedVersion,
			Cause:            fmt.Sprintf("This tool integration (%s) is not tested or vetted by Bitrise.", tool.ToolName),
			Recommendation:   fmt.Sprintf("If you want to use this tool anyway, look up its asdf plugin and provide it in the `plugin` field of the tool declaration. For example: `plugin: %s::https://github/url/to/asdf/plugin/repo.git`", tool.ToolName),
		}
	}
	if plugin.PluginName == "" {
		// Plugin name is required to install the plugin.
		return fmt.Errorf("plugin name for tool %s is not defined", tool.ToolName)
	}

	installed, err := a.isPluginInstalled(*plugin)
	if err != nil {
		log.Warnf("Failed to check if plugin is already installed: %v", err)
	}
	if installed {
		log.Debugf("Tool plugin %s is already installed, skipping installation.", tool.ToolName)
		return nil
	}

	pluginAddArgs := []string{"add", string(plugin.PluginName)}

	if plugin.GitCloneURL != "" {
		pluginAddArgs = append(pluginAddArgs, plugin.GitCloneURL)
	}

	_, err = a.ExecEnv.RunAsdfPlugin(pluginAddArgs...)
	if err != nil {
		return err
	}

	// Check if the plugin is found in the list of installed plugins after adding.
	installed, err = a.isPluginInstalled(*plugin)
	if err != nil {
		return fmt.Errorf("check if plugin was installed successfully: %w", err)
	}
	if !installed {
		return fmt.Errorf("%s plugin could not be installed", tool.ToolName)
	}

	return nil
}

func (a *AsdfToolProvider) isPluginInstalled(plugin PluginSource) (bool, error) {
	pluginListArgs := []string{"list", "--urls"}
	out, err := a.ExecEnv.RunAsdfPlugin(pluginListArgs...)
	if err != nil {
		return false, err
	}
	// If no plugins are installed, asdf returns exit code 0.

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, string(plugin.PluginName)) {
			if plugin.GitCloneURL != "" && !strings.Contains(line, string(plugin.GitCloneURL)) {
				log.Warnf("installed, but required URL does not match current:\n%s %s\n%s", plugin.PluginName, plugin.GitCloneURL, line)
			}
			return true, nil
		}
	}

	return false, nil
}

func fetchPluginSource(toolRequest provider.ToolRequest) (*PluginSource, error) {
	if toolRequest.PluginIdentifier != nil {
		pluginInput := strings.TrimSpace(*toolRequest.PluginIdentifier)
		if pluginInput != "" {
			// User provided a non empty plugin identifier, parse it.
			plugin, err := parsePluginSourceFromInput(pluginInput)
			if err != nil {
				return nil, fmt.Errorf("parse plugin identifier %s: %w", pluginInput, err)
			}
			return plugin, nil
		}
	}

	// Check if we have a predefined plugin source.
	if toolPlugin, exists := pluginSourceMap[toolRequest.ToolName]; exists {
		return &toolPlugin, nil
	}

	// No predefined plugin source found and no error in plugin identifier parsing,
	// return nil to indicate that no plugin source is defined for this tool.
	return nil, nil
}

// parsePluginSourceFromInput parses a plugin identifier string into a PluginSource struct.
// The expected format is "pluginName::[gitCloneURL]", where gitCloneURL is optional.
func parsePluginSourceFromInput(pluginIdentifier string) (*PluginSource, error) {
	parts := strings.Split(pluginIdentifier, PluginSourceSeparator)
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid plugin identifier format: %s, expected format is 'pluginName%s[gitCloneURL]'", pluginIdentifier, PluginSourceSeparator)
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid plugin identifier format: %s", pluginIdentifier)
	}
	pluginName := strings.TrimSpace(parts[0])
	if pluginName == "" {
		return nil, fmt.Errorf("plugin name cannot be empty in identifier: %s", pluginIdentifier)
	}
	if strings.HasPrefix(pluginName, "http://") || strings.HasPrefix(pluginName, "https://") {
		return nil, fmt.Errorf("plugin name should not contain URL: %s", pluginName)
	}

	pluginURL := ""
	if len(parts) > 1 {
		pluginURL = strings.TrimSpace(parts[1])
	}

	return &PluginSource{
		PluginName:  provider.ToolID(pluginName),
		GitCloneURL: pluginURL,
	}, nil
}
